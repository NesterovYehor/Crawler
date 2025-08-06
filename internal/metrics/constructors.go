package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	once     sync.Once
	instance *Metrics
)

func NewMetrics() *Metrics {
	once.Do(func() {
		instance = &Metrics{
			Crawler: newCrawlerMetrics(),
			Queue:   newQueueMetrics(),
			Store:   newStoreMetrics(),
		}
	})
	return instance
}

func newStoreMetrics() StoreMetrics {
	return &StorePrometheusMetrics{
		storeTotalRequests: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: "store",
			Name:      "requests_total",
			Help:      "Total requests to the store",
		}),
		storeTotalFails: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: "store",
			Name:      "failures_total",
			Help:      "Total failed requests to the store",
		}),
		storeLatency: promauto.NewHistogram(prometheus.HistogramOpts{
			Namespace: "store",
			Name:      "latency_seconds",
			Help:      "Latency of store operations in seconds",
			Buckets:   prometheus.DefBuckets, // or define custom buckets
		}),
		DB:    newDBMetrics(),
		Cache: newCacheMetrics(),
	}
}

func newCrawlerMetrics() CrawlerMetrics {
	return &CrawlerPrometheusMetrics{
		pagesCrawledTotal: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: "crawler",
			Name:      "pages_crawled_total",
			Help:      "Total number of pages successfully crawled.",
		}),
		pagesFailedTotal: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: "crawler",
			Name:      "pages_failed_total",
			Help:      "Total number of failed page crawl attempts.",
		}),
		crawlDurationSeconds: promauto.NewHistogram(prometheus.HistogramOpts{
			Namespace: "crawler",
			Name:      "crawl_duration_seconds",
			Help:      "Histogram of durations for crawling pages.",
		}),
	}
}

func newDBMetrics() DBMetrics {
	return &DBPrometheusMetrics{
		cassandraWritesTotal: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: "store",
			Subsystem: "db",
			Name:      "writes_total",
			Help:      "Total number of write operations to the database.",
		}),
		cassandraWriteErrorsTotal: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: "store",
			Subsystem: "db",
			Name:      "write_errors_total",
			Help:      "Total number of failed write operations to the database.",
		}),
		cassandraWriteLatency: promauto.NewHistogram(prometheus.HistogramOpts{
			Namespace: "store",
			Subsystem: "db",
			Name:      "write_latency_seconds",
			Help:      "Latency distribution of write operations to the database.",
		}),
	}
}

func newCacheMetrics() CacheMetrics {
	return &CachePrometheusMetrics{
		Redis:       newRedisMetrics(),
		BloomFilter: newBloomFilterMetrics(),
	}
}

func newRedisMetrics() RedisMetrics {
	return &RedisPrometheusMetrics{
		redisValuesTotal: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: "store",
			Subsystem: "cache",
			Name:      "values_current",
			Help:      "Current number of values stored in Redis.",
		}),
		redisRequestsTotal: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: "store",
			Subsystem: "cache",
			Name:      "requests_total",
			Help:      "Total number of requests made to Redis.",
		}),
		redisFailuresTotal: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: "store",
			Subsystem: "cache",
			Name:      "failures_total",
			Help:      "Total number of failed requests to Redis.",
		}),
		setLatency: promauto.NewHistogram(prometheus.HistogramOpts{
			Namespace: "store",
			Subsystem: "cache",
			Name:      "set_latency_seconds",
			Help:      "Latency distribution of set requests made to Redis.",
		}),
		fetchLatency: promauto.NewHistogram(prometheus.HistogramOpts{
			Namespace: "store",
			Subsystem: "cache",
			Name:      "fetch_latency_seconds",
			Help:      "Latency distribution of fetch requests made to Redis.",
		}),
	}
}

func newBloomFilterMetrics() BloomFilterMetrics {
	return &BloomFilterPrometheusMetrics{
		totalRequests: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: "store",
			Subsystem: "bloom_filter",
			Name:      "requests_total",
			Help:      "Total number of requests made to the bloom filter.",
		}),
		totalValuesAdded: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: "store",
			Subsystem: "bloom_filter",
			Name:      "values_added_total",
			Help:      "Total number of values added to the bloom filter.",
		}),
		totalFailures: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: "store",
			Subsystem: "bloom_filter",
			Name:      "failures_total",
			Help:      "Total number of bloom filter operation failures.",
		}),
		totalPositiveResponse: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: "store",
			Subsystem: "bloom_filter",
			Name:      "positive_responses_total",
			Help:      "Total number of positive (hit) responses from the bloom filter.",
		}),
		totalNegativeResponse: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: "store",
			Subsystem: "bloom_filter",
			Name:      "negative_responses_total",
			Help:      "Total number of negative (miss) responses from the bloom filter.",
		}),
	}
}

func newQueueMetrics() QueueMetrics {
	return &QueuePrometheusMetrics{
		lengthHigh: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: "queue",
			Name:      "high_priority_length_current",
			Help:      "Current number of tasks in the high priority queue.",
		}),
		lengthMedium: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: "queue",
			Name:      "medium_priority_length_current",
			Help:      "Current number of tasks in the medium priority queue.",
		}),
		lengthLow: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: "queue",
			Name:      "low_priority_length_current",
			Help:      "Current number of tasks in the low priority queue.",
		}),
		lengthStore: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: "queue",
			Name:      "store_queue_length_current",
			Help:      "Current number of tasks in the store queue.",
		}),
		fetchTotalHigh: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: "queue",
			Name:      "fetch_high_priority_total",
			Help:      "Total number of tasks fetched from the high priority queue.",
		}),
		fetchTotalMedium: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: "queue",
			Name:      "fetch_medium_priority_total",
			Help:      "Total number of tasks fetched from the medium priority queue.",
		}),
		fetchTotalLow: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: "queue",
			Name:      "fetch_low_priority_total",
			Help:      "Total number of tasks fetched from the low priority queue.",
		}),
		fetchTotalStore: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: "queue",
			Name:      "fetch_store_queue_total",
			Help:      "Total number of tasks fetched from the store queue.",
		}),
		totalFails: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: "queue",
			Name:      "failures_total",
			Help:      "Total number of queue request failures.",
		}),
	}
}
