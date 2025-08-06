package wp

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/NesterovYehor/Crawler/internal/config"
	httpclient "github.com/NesterovYehor/Crawler/internal/http_client"
	"github.com/NesterovYehor/Crawler/internal/metrics"
	"github.com/NesterovYehor/Crawler/internal/models"
	"github.com/NesterovYehor/Crawler/internal/politeness"
	"github.com/NesterovYehor/Crawler/internal/queue"
	"github.com/NesterovYehor/Crawler/internal/storage"
	"github.com/NesterovYehor/Crawler/internal/utils"
)

const (
	processRulesTask = "fetch_rules"
	processPageTask  = "crawl_page"
	storeDataTask    = "store_data"
)

type OperationChan struct {
	addChan   chan []*models.Task
	delChan   chan *models.Task
	retryChan chan *models.Task
}

type WorkerPool struct {
	queue          queue.Interface
	cfg            *config.Workers
	st             storage.Interface
	pm             *politeness.PolitenessManager
	operationChan  *OperationChan
	buffers        map[string]chan *models.Task
	wg             *sync.WaitGroup
	httpClient     httpclient.Interface
	fillInProgress map[string]bool
	mu             sync.Mutex
	metrics        *metrics.Metrics
}

type WorkerPoolOpts struct {
	Config     *config.Workers
	Queue      queue.Interface
	ST         storage.Interface
	PM         *politeness.PolitenessManager
	HttpClient httpclient.Interface
	Metrics    *metrics.Metrics
}

func NewWorkerPool(opts *WorkerPoolOpts) (*WorkerPool, error) {
	return &WorkerPool{
		st:         opts.ST,
		queue:      opts.Queue,
		pm:         opts.PM,
		httpClient: opts.HttpClient,
		metrics:    opts.Metrics,
		operationChan: &OperationChan{
			addChan:   make(chan []*models.Task, 25000),
			delChan:   make(chan *models.Task, 7000),
			retryChan: make(chan *models.Task, 5000),
		},
		buffers: map[string]chan *models.Task{
			"queue:fetch:high":   make(chan *models.Task, opts.Config.Fetch.HighPrioretyCount*2),
			"queue:fetch:medium": make(chan *models.Task, opts.Config.Fetch.MedPrioretyCount*2),
			"queue:fetch:retry":  make(chan *models.Task, opts.Config.Fetch.LowPrioretyCount*2),
			"queue:store":        make(chan *models.Task, opts.Config.Upload.Count*2),
		},
		cfg: opts.Config,
		fillInProgress: map[string]bool{
			"queue:fetch:high":   false,
			"queue:fetch:medium": false,
			"queue:fetch:low":    false,
			"queue:store":        false,
		},
		wg: &sync.WaitGroup{},
	}, nil
}

func (wp *WorkerPool) Run(ctx context.Context) {
	wp.spawnWorkers(wp.cfg.Fetch.HighPrioretyCount, ctx, queue.HighPriorityQueue)
	wp.spawnWorkers(wp.cfg.Fetch.MedPrioretyCount, ctx, queue.MediumPriorityQueue)
	wp.spawnWorkers(wp.cfg.Fetch.LowPrioretyCount, ctx, queue.RetryPriorityQueue)
	wp.spawnWorkers(wp.cfg.Upload.Count, ctx, queue.StoreQueue)

	wp.wg.Add(3)
	go func() {
		defer wp.wg.Done()
		wp.handleAdd(ctx)
	}()
	go func() {
		defer wp.wg.Done()
		wp.handleDel(ctx)
	}()
	go func() {
		defer wp.wg.Done()
		wp.handleRetry(ctx)
	}()
	<-ctx.Done()
}

func (wp *WorkerPool) spawnWorkers(count int, ctx context.Context, sourceName string) {
	for i := range count {
		go func(idx int) {
			state := queue.NewSource(sourceName)
			id := fmt.Sprintf("%v:%v", sourceName, idx)
			w := newWorker(id, wp, state, wp.metrics)
			w.Start(ctx)
		}(i)
	}
}

func (wp *WorkerPool) getNextTask(worker *Worker, ctx context.Context) *models.Task {
	source := worker.taskSource.GetCurrentSource()
	if len(wp.buffers[source]) < max(1, cap(wp.buffers[source])/4) {
		wp.tryRefillQueue(ctx, worker.taskSource)
	}

	select {
	case task := <-wp.buffers[source]:
		return task
	default:
		return nil
	}
}
func (wp *WorkerPool) tryRefillQueue(ctx context.Context, source *queue.Source) {
	wp.mu.Lock()
	if wp.fillInProgress[source.Curr] {
		wp.mu.Unlock()
		return
	}
	wp.fillInProgress[source.Curr] = true
	wp.mu.Unlock()

	// This goroutine now performs one single refill operation and then exits.
	// This prevents the goroutine leak.
	go func() {
		defer func() {
			wp.mu.Lock()
			wp.fillInProgress[source.Curr] = false
			wp.mu.Unlock()
		}()

		// Check if the context is already done before starting.
		if ctx.Err() != nil {
			return
		}

		queueToTry := source.GetCurrentSource()
		err := wp.fillChannel(wp.buffers[queueToTry], ctx, queueToTry)
		if err != nil {
			if errors.Is(err, utils.ErrNoTasks) {
				source.MarkQueueFailed()
			} else if !errors.Is(err, context.Canceled) {
				slog.Error("WorkerPool: fillChannel background error", "error", err, "source", queueToTry)
				source.MarkQueueFailed()
			}
		}
	}()
}


func (wp *WorkerPool) fillChannel(ch chan *models.Task, ctx context.Context, sourceName string) error {
	messages, err := wp.queue.GetTasks(ctx, max(cap(ch)-len(ch), 1), sourceName)
	if err != nil {
		return err
	}
	for _, m := range messages {
		if m != nil {
			select {
			case ch <- m:
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Second):
				return fmt.Errorf("buffer for source %s full, failed to push task within timeout", sourceName)
			}
		} else {
			return nil
		}
	}
	return nil
}

func (wp *WorkerPool) handleAdd(ctx context.Context) {
	buffer := make([]*models.Task, 0, 10000)
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case msg, _ := <-wp.operationChan.addChan:
			for _, m := range msg {
				if m.Topic == processPageTask {
					crawled, err := wp.st.ExistsInBF(ctx, m.URL)
					if err != nil {
						slog.Error(err.Error())
					}
					if !crawled {
						buffer = append(buffer, m)
					}
				} else {
					buffer = append(buffer, m)

				}
			}
		if len(buffer) >= 1000 {
				wp.flushAdd(&buffer)
			}
		case <-ticker.C:
			if len(buffer) > 0 {
				wp.flushAdd(&buffer)
			}
		}
	}
}

func (wp *WorkerPool) flushAdd(buffer *[]*models.Task) {
	if err := wp.queue.Add(*buffer); err != nil {
		slog.Warn("Error while adding new messages to a queue: error", "error", err)
	}
	*buffer = (*buffer)[:0]
}

func (wp *WorkerPool) handleDel(ctx context.Context) {
	buffer := make([]*models.Task, 0, 500)
	ticker := time.NewTicker(30 * time.Millisecond)
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-wp.operationChan.delChan:
			if !ok {
				if len(buffer) > 100 {
					wp.flushDel(buffer)
					buffer = buffer[:0]
				}
			}
			buffer = append(buffer, msg)
		case <-ticker.C:
			if len(buffer) > 0 {
				wp.flushDel(buffer)
				buffer = buffer[:0]
			}

		}
	}
}

func (wp *WorkerPool) flushDel(buffer []*models.Task) {
	if err := wp.queue.Del(buffer); err != nil {
		slog.Warn(fmt.Sprintf("Error while deleating messages from a queue:%v", err))
	}
}

func (wp *WorkerPool) handleRetry(ctx context.Context) {
	for {
		select {

		case <-ctx.Done():
			return
		default:
			msg, ok := <-wp.operationChan.retryChan
			if !ok {
				return
			}
			if err := wp.queue.Retry(ctx, *msg); err != nil {
				slog.Warn(fmt.Sprintf("Error while retrying models.Task: %v", err))
			}
		}
	}
}
