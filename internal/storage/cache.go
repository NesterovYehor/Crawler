package storage

import (
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/net/context"
)

type Cache struct {
	client *redis.Client
}

func NewCache(client *redis.Client) *Cache {
	return &Cache{
		client: client,
	}
}

func (c *Cache) Get(domain string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	val, err := c.client.Get(ctx, "robots:"+domain).Result()
	if err != err {
		return "", nil
	}
	return val, nil
}

func (c *Cache) Save(key string, value any) error {
	ctx, canel := context.WithTimeout(context.Background(), time.Second*5)
	defer canel()
	return c.client.Set(ctx, key, value, time.Hour*24*30).Err()
}

func (c *Cache) CheckBloomFilter(url string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	added, err := c.client.Do(ctx, "BF.ADD", "url_filter", url).Bool()
	if err != nil {
		return false, err
	}

	return !added, nil
}
