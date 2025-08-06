package wp

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/NesterovYehor/Crawler/internal/crawler"
	"github.com/NesterovYehor/Crawler/internal/metrics"
	"github.com/NesterovYehor/Crawler/internal/models"
	"github.com/NesterovYehor/Crawler/internal/politeness"
	"github.com/NesterovYehor/Crawler/internal/queue"
	"github.com/NesterovYehor/Crawler/internal/utils"
	"github.com/redis/go-redis/v9"
	"github.com/temoto/robotstxt"
)

var start = time.Now()

type Worker struct {
	ID         string
	pool       *WorkerPool
	taskSource *queue.Source
	metrics    *metrics.Metrics
}

func newWorker(
	id string,
	pool *WorkerPool,
	taskSource *queue.Source,
	metrics *metrics.Metrics,
) *Worker {
	return &Worker{
		ID:         id,
		pool:       pool,
		taskSource: taskSource,
		metrics:    metrics,
	}
}

func (w *Worker) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			slog.Info(fmt.Sprintf("Stoping Worker: %v", w.ID))
			return
		default:
			task := w.pool.getNextTask(w, ctx)
			if task == nil {
				time.Sleep(time.Millisecond * 50)
				continue
			}
			err := w.dispatchTask(ctx, task)
			if err != nil {
				slog.Error("error processing task", "err", err, "url", task)
				continue
			}
		}
	}
}

func (w *Worker) dispatchTask(ctx context.Context, task *models.Task) error {
	defer func() {
		w.pool.operationChan.delChan <- task
	}()

	domain, err := utils.GetDomain(task.URL)
	if err != nil {
		return fmt.Errorf("failed to parse domain: %w", err)
	}

	switch task.Topic {
	case processRulesTask:
		exists, err := w.pool.st.ExistsInBF(ctx, "robots:"+domain)
		if err != nil {
			w.pool.operationChan.retryChan <- task
			return nil
		}
		if exists {
			task.Topic = processPageTask
			w.addToSuccessChannel(task)
			return nil
		}

		newTask, retry, err := w.processRobotsTask(ctx, task.URL)
		if err != nil {
			slog.Error("robots.txt error", "err", err)
			if retry {
				w.pool.operationChan.retryChan <- task
			} else {
				if err := w.pool.pm.SaveRules(ctx, domain, ""); err != nil {
					return err
				}
				task.Topic = processPageTask
				w.addToSuccessChannel(task)
			}
			return nil
		}

		w.addToSuccessChannel(newTask)

	case processPageTask:
		retry, err := w.processPageTask(ctx, task)
		if err != nil {
			slog.Error("page crawl error", "err", err, "url", task.URL)
			if retry {
				defer func() { w.pool.operationChan.retryChan <- task }()
				if err := w.pool.pm.UpdateHostLimit(domain, ctx); err != nil {
					return err
				}
			}
			return err
		}

		if err := w.pool.st.AddToBF(ctx, task.URL); err != nil {
			return err
		}
	case storeDataTask:
		if err := w.processStoreTask(ctx, *task); err != nil {
			return err
		}

	}

	return nil
}

func (w *Worker) processRobotsTask(ctx context.Context, rawURL string) (*models.Task, bool, error) {
	domain, err := utils.GetDomain(rawURL)
	if err != nil || domain == "" {
		return nil, false, fmt.Errorf("invalid domain: %w", err)
	}

	rules, retry, siteMap, err := w.pool.httpClient.FetchRules(domain)
	if err != nil {
		return nil, retry, err
	}

	if err := w.pool.pm.SaveRules(ctx, domain, string(rules)); err != nil {
		return nil, true, err
	}

	var tasks []*models.Task
	for _, url := range siteMap {
		if url != "" {
			tasks = append(tasks, models.NewTask(processPageTask, url, queue.HighPriorityQueue, ""))
		}
	}
	w.pool.operationChan.addChan <- tasks

	return models.NewTask(processPageTask, rawURL, queue.MediumPriorityQueue, ""), false, nil
}

func (w *Worker) processPageTask(ctx context.Context, task *models.Task) (bool, error) {
	timer := time.Now()
	domain, err := utils.GetDomain(task.URL)
	if err != nil {
		return false, fmt.Errorf("invalid URL: %w", err)
	}

	rules, err := w.getDomainRules(domain, task, ctx)
	if err != nil {
		return true, err
	}

	if !w.isAllowedByRobotsTxt(task.URL, rules) {
		return rules.Allowed, fmt.Errorf("access disallowed by robots.txt for domain: %v", domain)
	}

	crawlResult, err := crawler.CrawlPage(task.URL, domain, w.pool.httpClient)
	w.metrics.Crawler.Update(err != nil, time.Since(timer))
	if err != nil {
		return crawlResult.Retry, fmt.Errorf("crawl failed: %w", err)
	}

	if err := w.processCrawledData(ctx, crawlResult, task); err != nil {
		return false, fmt.Errorf("processing crawled data failed: %w", err)
	}

	return false, nil
}

func (w *Worker) processStoreTask(ctx context.Context, task models.Task) error {
	start := time.Now()
	data, err := w.pool.st.GetTempByUUID(ctx, task.DataID)
	if err != nil {
		return fmt.Errorf("Error getting temp data: %v", err)
	}
	if err := w.pool.st.SaveIfNew(ctx, data); err != nil {
		w.metrics.Store.Update(true, time.Since(start))
		return fmt.Errorf("Error saving data: %v", err)
	}
	w.metrics.Store.Update(false, time.Since(start))
	return nil
}


func (w *Worker) processCrawledData(ctx context.Context, result *crawler.CrawlResult, task *models.Task) error {
	dataID, err := w.pool.st.SaveTempWithUUID(ctx, result.PageData)
	if err != nil {
		return err
	}

	var tasks []*models.Task
	tasks = append(tasks, models.NewTask(storeDataTask, task.URL, queue.StoreQueue, dataID))

	for _, childURL := range result.Urls {
		if childURL != "" {
			t := models.NewTask(processPageTask, childURL, queue.MediumPriorityQueue, "")
			tasks = append(tasks, t)
		}
	}

	w.pool.operationChan.addChan <- tasks
	return nil
}

func (w *Worker) addToSuccessChannel(task *models.Task) {
	w.pool.operationChan.addChan <- []*models.Task{task}
}

func (w *Worker) isAllowedByRobotsTxt(url string, rules *politeness.RateLimitResult) bool {
	allowed, err := w.processRateLimiter(url, rules.Rules)
	if err != nil {
		return false
	}
	return allowed
}

func (w *Worker) processRateLimiter(url, rawRules string) (bool, error) {
	if rawRules != "" {
		robotsData, err := robotstxt.FromString(rawRules)
		if err != nil {
			return false, err
		}
		if robotsData != nil && !robotsData.TestAgent(url, "*") {
			return false, nil
		}
	}
	return true, nil
}

func (w *Worker) getDomainRules(domain string, task *models.Task, ctx context.Context) (*politeness.RateLimitResult, error) {
	rules, err := w.pool.pm.GetRules(domain, ctx)
	if err != nil {
		if err == redis.Nil {
			task.Topic = processRulesTask
			w.pool.operationChan.addChan <- []*models.Task{task}
		}
		return &politeness.RateLimitResult{Allowed: false}, fmt.Errorf("failed to get rules: %w", err)
	}

	return rules, nil
}
