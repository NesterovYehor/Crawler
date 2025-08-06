package testutils

import (
	"context"
	"log"
	"os"

	"github.com/NesterovYehor/Crawler/internal/config"
	"github.com/NesterovYehor/Crawler/internal/storage"
	"github.com/NesterovYehor/Crawler/internal/storage/blob"
	"github.com/NesterovYehor/Crawler/internal/storage/cache"
	"github.com/NesterovYehor/Crawler/internal/storage/metadata"
	"github.com/NesterovYehor/Crawler/tests/testutils/mocks"
	"github.com/redis/go-redis/v9"
)

func SetupStorage(redis *redis.Client, cfg *config.Config) (storage.Interface, func() error, error) {
	if cfg != nil {
		if err := loadScripts(redis, cfg.Scripts); err != nil {
			return nil, nil, err
		}
	}
	metrics := mocks.NewNoopMetrics()
	cache, err := cache.NewCache(redis, metrics.Store.CacheMetrics())
	if err != nil {
		return nil, nil, err
	}

	blob, err := blob.NewS3Client()
	if err != nil {
		return nil, nil, err
	}
	dbHost, cleanUp, err := RunCassandra(context.Background())
	if err != nil {
		return nil, cleanUp, err
	}
	cfg.DB.Addr = dbHost
	ms, err := metadata.NewCassandraStore(cfg.DB, metrics.Store.DBMetrics())
	if err != nil {
		return nil, cleanUp, err
	}
	st := storage.NewStorage(&storage.StorageOpts{
		Meta:    ms,
		Blob:    blob,
		Cache:   cache,
		Metrics: metrics.Store,
	})

	return st, cleanUp, nil
}

func loadScripts(client *redis.Client, cfg *config.Scripts) error {
	scriptByts, err := os.ReadFile("../scripts/politeness_gate.lua")
	if err != nil {
		return err
	}
	if scriptByts == nil {
		return err
	}
	cfg.Access, err = client.ScriptLoad(context.Background(), string(scriptByts)).Result()
	if err != nil {
		return err
	}
	scriptByts, err = os.ReadFile("../scripts/update_token_limit.lua")
	if err != nil {
		return err
	}
	if scriptByts == nil {
		return err
	}

	cfg.Update, err = client.ScriptLoad(context.Background(), string(scriptByts)).Result()
	if err != nil {
		return err
	}

	log.Printf("Acceess: %v || UPDATE: %v", cfg.Access, cfg.Update)
	return nil
}
