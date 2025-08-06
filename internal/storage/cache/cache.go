package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/NesterovYehor/Crawler/internal/metrics"
	"github.com/redis/go-redis/v9"
)

type Cache struct {
	client  *redis.Client
	metrics metrics.CacheMetrics
}

func NewCache(client *redis.Client, metrics metrics.CacheMetrics) (*Cache, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	err := client.BFReserve(ctx, "url_filter", 0.01, 100000).Err()
	if err != nil && !strings.Contains(err.Error(), "item exists") {
		return nil, fmt.Errorf("failed to create bloom filter: %w", err)
	}
	return &Cache{
		client:  client,
		metrics: metrics,
	}, nil
}

func (c *Cache) RunScript(key, sriptHash string, ctx context.Context) (any, error) {
	return c.client.EvalSha(ctx, sriptHash, []string{key}).Result()
}

func (c *Cache) Get(ctx context.Context, key string) ([]byte, error) {
	start := time.Now()
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		c.metrics.RedisMetrics().ObserveFailure()
		return nil, err
	}
	c.metrics.RedisMetrics().ObserveFetch(time.Since(start))
	return data, nil
}

func (c *Cache) Save(ctx context.Context, key string, values map[string]any) error {
	start := time.Now()
	err := c.client.HSet(ctx, key, values).Err()
	if err != nil {
		c.metrics.RedisMetrics().ObserveFailure()
		return err
	}
	c.metrics.RedisMetrics().ObserveAdd(time.Since(start))
	return nil
}

func (c *Cache) SaveWithTTL(ctx context.Context, key string, values map[string]any, ttl time.Duration) error {
	start := time.Now()
	data, err := json.Marshal(values)
	if err != nil {
		return err
	}
	err = c.client.SetEx(ctx, key, data, ttl).Err()
	if err != nil {
		c.metrics.RedisMetrics().ObserveFailure()
		return err
	}
	c.metrics.RedisMetrics().ObserveAdd(time.Since(start))
	return nil
}

func (c *Cache) AddToBF(ctx context.Context, key string) error {
	err := c.client.BFAdd(ctx, key, key).Err()
	if err != nil {
		c.metrics.BloomFilterMetrics().ObserveFailure()
		return fmt.Errorf("failed to add URL to bloom filter: %w", err)
	}
	c.metrics.BloomFilterMetrics().ObserveAdd()
	return nil
}

func (c *Cache) CheckBF(ctx context.Context, key string) (bool, error) {
	exist, err := c.client.BFExists(ctx, key, key).Result()
	if err != nil {
		c.metrics.BloomFilterMetrics().ObserveFailure()
		return false, err
	}
	c.metrics.BloomFilterMetrics().ObserveFetch(exist)
	return exist, nil
}
