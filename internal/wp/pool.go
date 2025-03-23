package wp

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"sort"
	"sync"

	"github.com/NesterovYehor/Crawler/internal/config"
	"github.com/NesterovYehor/Crawler/internal/crawler"
	"github.com/NesterovYehor/Crawler/internal/queue"
	"github.com/NesterovYehor/Crawler/internal/storage"
	"github.com/redis/go-redis/v9"
	"github.com/temoto/robotstxt"
)

const (
	processRulesTask = "fetch_robots"
	processPageTask  = "crawl_page"
)

type WorkerPool struct {
	mu             *sync.Mutex
	queue          *queue.Queue
	cache          *storage.Cache
	pages          map[string]int
	maxConcurrency int
	wg             *sync.WaitGroup
}

func NewWorkerPool(pages map[string]int, cfg *config.Config) (*WorkerPool, error) {
	return &WorkerPool{
		mu:             &sync.Mutex{},
		pages:          pages,
		cache:          cfg.Cache,
		queue:          cfg.Queue,
		maxConcurrency: cfg.MaxConcurrency,
		wg:             &sync.WaitGroup{},
	}, nil
}

func (wp *WorkerPool) Run() {
	for i := 0; i < wp.maxConcurrency; i++ {
		wp.wg.Add(1)
		go wp.worker(context.Background())
	}

	wp.wg.Wait()

	wp.printReport()
}

func (wp *WorkerPool) worker(ctx context.Context) {
	defer wp.wg.Done()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Worker shutting down")
			return
		default:
			streams, err := wp.queue.FetchFromQueue()
			if err != nil {
				fmt.Println("Error fetching from queue:", err)
				return
			}

			for _, stream := range streams {
				for _, message := range stream.Messages {
					url := message.Values["url"].(string)

					switch message.Values["topic"] {
					case processRulesTask:
						err = wp.ProcessRobotsTask(url)
						if err != nil {
							if err := wp.retryTask(processRulesTask, message.Values, url); err != nil {
								fmt.Println(err)
								continue
							}
						}

					case processPageTask:
						exist, err := wp.cache.CheckBloomFilter(url)
						if err != nil {
							fmt.Println("Error checking Bloom filter:", err)
							continue
						}
						if exist {
							continue
						}

						err = wp.ProcessPageTask(url)
						if err != nil {
							if err := wp.retryTask(processPageTask, message.Values, url); err != nil {
								fmt.Println(err)
								continue
							}
						}
					}
				}
			}
		}
	}
}

func (wp *WorkerPool) retryTask(topic string, message map[string]any, url string) error {
	newTask := map[string]any{
		"topic":   topic,
		"retries": message["retries"].(int) + 1,
		"url":     url,
	}
	return wp.queue.AddToQueue(newTask)
}

func (wp *WorkerPool) ProcessRobotsTask(rawURL string) error {
	rawRules, err := wp.cache.Get(rawURL)
	if err == redis.Nil { // Not cached, need to fetch robots.txt
		robotsURL := fmt.Sprintf("http://%s/robots.txt", rawURL)
		parsedURL, err := url.Parse(robotsURL)
		if err != nil {
			return fmt.Errorf("invalid domain format: %w", err)
		}
		resp, err := crawler.Fetch(parsedURL)
		if err != nil {
			return err
		}
		rules, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		err = wp.cache.Save(rawRules, rules)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}

func (wp *WorkerPool) ProcessPageTask(url string) error {
	rulesStr, err := wp.cache.Get(url)
	if err != nil {
		return err
	}
	rules, err := robotstxt.FromString(rulesStr)
	if err != nil {
		return err
	}

	if rules != nil && !rules.TestAgent(url, "*") {
		return nil
	}

	urls, err := crawler.CrawlPage(url)
	if err != nil {
		return err
	}

	if err := wp.cache.Save(processRulesTask, urls); err != nil {
		return err
	}
	return nil
}

func (wp *WorkerPool) printReport() {
	fmt.Println("\n=============================")
	fmt.Println("CRAWLING REPORT")
	fmt.Println("=============================")

	keys := make([]string, 0, len(wp.pages))
	for key := range wp.pages {
		keys = append(keys, key)
	}

	sort.Strings(keys)
	total := 0
	for _, key := range keys {
		fmt.Printf("Found %v internal links to %v\n", wp.pages[key], key)
		total += wp.pages[key]
	}
	fmt.Printf("Total amount of pages %v\n", total)
}
