package config

import (
	"errors"

	"github.com/NesterovYehor/Crawler/internal/queue"
	"github.com/NesterovYehor/Crawler/internal/storage"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

type Config struct {
	Queue          *queue.Queue
	Cache          *storage.Cache
	MaxConcurrency int
}

func NewConfig() *Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		if errors.Is(err, viper.ConfigFileNotFoundError{}) {
			setDefault()
		} else {
			panic(err)
		}
	}
	redisClient := redis.NewClient(&redis.Options{
		Addr: viper.GetString("redis_addr"),
	})

	queue, err := queue.NewQueue(redisClient, viper.GetString("stream_name"))
	if err != nil {
		panic(err)
	}
	cache := storage.NewCache(redisClient)

	cfg := &Config{
		Queue:          queue,
		Cache:          cache,
		MaxConcurrency: viper.GetInt("max_concurrency"),
	}

	return cfg
}

func setDefault() {
	viper.SetDefault("max_concurrency", 5)
	viper.SetDefault("stream_name", "domains")
	viper.SetDefault("redis_addr", "localhost:6379") // Should be a string
}
