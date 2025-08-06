package mocks

import (
	"time"

	"github.com/NesterovYehor/Crawler/internal/metrics"
)

func NewNoopMetrics() *metrics.Metrics {
	return &metrics.Metrics{
		Crawler: &CrawlerNoopMetrics{},
		Queue:   &QueueNoopMetrics{},
		Store: &StoreNoopMetrics{
			DB: &DBNoopMetrics{},
			Cache: &CacheNoopMetrics{
				Redis:       &RedisNoopMetrics{},
				BloomFilter: &BloomFilterNoopMetrics{},
			},
		},
	}
}

// --- Store ---
type StoreNoopMetrics struct {
	DB    *DBNoopMetrics
	Cache *CacheNoopMetrics
}

func (m *StoreNoopMetrics) Update(_ bool, _ time.Duration) {}

func (m *StoreNoopMetrics) CacheMetrics() metrics.CacheMetrics { return m.Cache }
func (m *StoreNoopMetrics) DBMetrics() metrics.DBMetrics       { return m.DB }

// --- Crawler ---
type CrawlerNoopMetrics struct{}

func (m *CrawlerNoopMetrics) Update(_ bool, _ time.Duration) {}

// --- DB ---
type DBNoopMetrics struct{}

func (m *DBNoopMetrics) Update(_ bool, _ time.Duration) {}

// --- Queue ---
type QueueNoopMetrics struct{}

func (m *QueueNoopMetrics) ObserveAdd(_ string)   {}
func (m *QueueNoopMetrics) ObserveFailure()       {}
func (m *QueueNoopMetrics) ObserveFetch(_ string) {}

// --- Cache ---
type CacheNoopMetrics struct {
	Redis       *RedisNoopMetrics
	BloomFilter *BloomFilterNoopMetrics
}

func (m *CacheNoopMetrics) RedisMetrics() metrics.RedisMetrics {
	return m.Redis
}

func (m *CacheNoopMetrics) BloomFilterMetrics() metrics.BloomFilterMetrics {
	return m.BloomFilter
}

// --- BloomFilter ---
type BloomFilterNoopMetrics struct{}

func (m *BloomFilterNoopMetrics) ObserveAdd()         {}
func (m *BloomFilterNoopMetrics) ObserveFailure()     {}
func (m *BloomFilterNoopMetrics) ObserveFetch(_ bool) {}

// --- Redis ---
type RedisNoopMetrics struct{}

func (m *RedisNoopMetrics) ObserveAdd(_ time.Duration)   {}
func (m *RedisNoopMetrics) ObserveFetch(_ time.Duration) {}
func (m *RedisNoopMetrics) ObserveFailure()              {}
