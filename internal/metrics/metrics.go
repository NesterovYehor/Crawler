package metrics

import (
	"time"

	"github.com/NesterovYehor/Crawler/internal/queue"
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	Crawler CrawlerMetrics
	Queue   QueueMetrics
	Store   StoreMetrics
}

type StorePrometheusMetrics struct {
	storeTotalRequests prometheus.Counter
	storeTotalFails    prometheus.Counter
	storeLatency       prometheus.Histogram
	DB                 DBMetrics
	Cache              CacheMetrics
}

func (m *StorePrometheusMetrics) Update(failed bool, dur time.Duration) {
	if !failed {
		m.storeTotalRequests.Inc()
		m.storeLatency.Observe(dur.Seconds())
		return
	}
	m.storeTotalFails.Inc()
}

func (m *StorePrometheusMetrics) CacheMetrics() CacheMetrics {
	return m.Cache
}

func (m *StorePrometheusMetrics) DBMetrics() DBMetrics {
	return m.DB
}

type CrawlerPrometheusMetrics struct {
	pagesCrawledTotal    prometheus.Counter
	pagesFailedTotal     prometheus.Counter
	crawlDurationSeconds prometheus.Histogram
}

func (m *CrawlerPrometheusMetrics) Update(failed bool, dur time.Duration) {
	if !failed {
		m.pagesCrawledTotal.Inc()
		m.crawlDurationSeconds.Observe(dur.Seconds())
		return
	}
	m.pagesFailedTotal.Inc()
}

type DBPrometheusMetrics struct {
	cassandraWritesTotal      prometheus.Counter
	cassandraWriteErrorsTotal prometheus.Counter
	cassandraWriteLatency     prometheus.Histogram
}

func (m *DBPrometheusMetrics) Update(failed bool, dur time.Duration) {
	m.cassandraWritesTotal.Inc()
	if !failed {
		m.cassandraWriteLatency.Observe(dur.Seconds())
		return
	}
	if failed {
		m.cassandraWriteErrorsTotal.Inc()
	}
}

type QueuePrometheusMetrics struct {
	totalFails       prometheus.Counter
	lengthHigh       prometheus.Gauge
	lengthMedium     prometheus.Gauge
	lengthStore      prometheus.Gauge
	lengthLow        prometheus.Gauge
	fetchTotalHigh   prometheus.Counter
	fetchTotalMedium prometheus.Counter
	fetchTotalLow    prometheus.Counter
	fetchTotalStore  prometheus.Counter
}

func (m *QueuePrometheusMetrics) ObserveAdd(source string) {
	switch source {
	case queue.HighPriorityQueue:
		m.lengthHigh.Inc()
	case queue.MediumPriorityQueue:
		m.lengthMedium.Inc()
	case queue.StoreQueue:
		m.lengthStore.Inc()
	case queue.RetryPriorityQueue:
		m.lengthLow.Inc()
	}
}

func (m *QueuePrometheusMetrics) ObserveFailure() {
	m.totalFails.Inc()
}

func (m *QueuePrometheusMetrics) ObserveFetch(source string) {
	switch source {
	case queue.HighPriorityQueue:
		m.fetchTotalHigh.Inc()
		m.lengthHigh.Dec()
	case queue.MediumPriorityQueue:
		m.fetchTotalMedium.Inc()
		m.lengthMedium.Dec()
	case queue.StoreQueue:
		m.fetchTotalStore.Inc()
		m.lengthStore.Dec()
	case queue.RetryPriorityQueue:
		m.fetchTotalLow.Inc()
		m.lengthLow.Dec()
	}
}

type CachePrometheusMetrics struct {
	Redis       RedisMetrics
	BloomFilter BloomFilterMetrics
}

func (m *CachePrometheusMetrics) RedisMetrics() RedisMetrics {
	return m.Redis
}

func (m *CachePrometheusMetrics) BloomFilterMetrics() BloomFilterMetrics {
	return m.BloomFilter
}

type BloomFilterPrometheusMetrics struct {
	totalRequests         prometheus.Counter
	totalValuesAdded      prometheus.Counter
	totalFailures         prometheus.Counter
	totalPositiveResponse prometheus.Counter
	totalNegativeResponse prometheus.Counter
}

func (m *BloomFilterPrometheusMetrics) ObserveAdd() {
	m.totalValuesAdded.Inc()
}

func (m *BloomFilterPrometheusMetrics) ObserveFailure() {
	m.totalFailures.Inc()
}

func (m *BloomFilterPrometheusMetrics) ObserveFetch(success bool) {
	m.totalRequests.Inc()
	if success {
		m.totalPositiveResponse.Inc()
	} else {
		m.totalNegativeResponse.Inc()
	}
}

type RedisPrometheusMetrics struct {
	redisValuesTotal   prometheus.Gauge
	redisRequestsTotal prometheus.Counter
	redisFailuresTotal prometheus.Counter
	fetchLatency       prometheus.Histogram
	setLatency         prometheus.Histogram
}

func (m *RedisPrometheusMetrics) ObserveAdd(dur time.Duration) {
	m.setLatency.Observe(dur.Seconds())
	m.redisValuesTotal.Inc()
}

func (m *RedisPrometheusMetrics) ObserveFetch(dur time.Duration) {
	m.redisRequestsTotal.Inc()
	m.fetchLatency.Observe(dur.Seconds())
	m.redisValuesTotal.Dec()
}

func (m *RedisPrometheusMetrics) ObserveFailure() {
	m.redisFailuresTotal.Inc()
}
