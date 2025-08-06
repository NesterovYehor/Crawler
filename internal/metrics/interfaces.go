package metrics

import "time"

type StoreMetrics interface {
	Update(failed bool, dur time.Duration)
	CacheMetrics() CacheMetrics
	DBMetrics() DBMetrics
}

type DBMetrics interface {
	Update(failed bool, dur time.Duration)
}

type CacheMetrics interface {
	RedisMetrics() RedisMetrics
	BloomFilterMetrics() BloomFilterMetrics
}

type RedisMetrics interface {
	ObserveAdd(dur time.Duration)
	ObserveFetch(dur time.Duration)
	ObserveFailure()
}

type BloomFilterMetrics interface {
	ObserveAdd()
	ObserveFailure()
	ObserveFetch(success bool)
}

// === Crawler ===

type CrawlerMetrics interface {
	Update(failed bool, dur time.Duration)
}

// === Queue ===

type QueueMetrics interface {
	ObserveAdd(source string)
	ObserveFailure()
	ObserveFetch(source string)
}
