package storage

import (
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/net/context"
)

type RedisClient struct {
	client *redis.Client
}

func NewRedisClient(addr string) *RedisClient {
	return &RedisClient{
		client: redis.NewClient(&redis.Options{Addr: addr}),
	}
}

func (c *RedisClient) Get(domain string, ctx context.Context) (string, error) {
	val, err := c.client.Get(ctx, "robots:"+domain).Result()
	if err != err {
		return "", nil
	}
	return val, nil
}

func (c *RedisClient) Save(domain, content string, ctx context.Context) error {
	return c.client.Set(ctx, "robots:"+domain, content, time.Hour*24).Err()
}
