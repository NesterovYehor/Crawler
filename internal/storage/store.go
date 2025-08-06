package storage

import (
	"context"
	"encoding/json"
	"time"

	"github.com/NesterovYehor/Crawler/internal/metrics"
	"github.com/NesterovYehor/Crawler/internal/models"
	"github.com/NesterovYehor/Crawler/internal/storage/blob"
	"github.com/NesterovYehor/Crawler/internal/storage/cache"
	"github.com/NesterovYehor/Crawler/internal/storage/metadata"
	"github.com/NesterovYehor/Crawler/internal/utils"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Storage struct {
	Metadata metadata.MetadataStore
	Blob     blob.BlobStore
	Cache    *cache.Cache
	metrics  metrics.StoreMetrics
}

type StorageOpts struct {
	Metrics metrics.StoreMetrics
	Meta    metadata.MetadataStore
	Blob    blob.BlobStore
	Cache   *cache.Cache
}

func NewStorage(opts *StorageOpts) Interface {
	return &Storage{
		Metadata: opts.Meta,
		Blob:     opts.Blob,
		Cache:    opts.Cache,
		metrics:  opts.Metrics,
	}
}

func (st *Storage) ExistsInBF(ctx context.Context, key string) (bool, error) {
	return st.Cache.CheckBF(ctx, key)
}

func (st *Storage) AddToBF(ctx context.Context, key string) error {
	return st.Cache.AddToBF(ctx, key)
}

func (st *Storage) SaveToCache(ctx context.Context, key string, values map[string]any) error {
	return st.Cache.Save(ctx, key, values)
}

func (st *Storage) SaveIfNew(ctx context.Context, data *models.PageDataModel) error {
	start := time.Now()
	if !data.IsValid() {
		return utils.ErrInValidPageData
	}
	exists, err := st.ExistsInBF(ctx, data.Metadata.HTMLHash)
	if err != nil {
		return err
	}
	if !exists {
		if err := st.Metadata.Save(ctx, data.Metadata); err != nil {
			return err
		}
		return st.AddToBF(ctx, data.Metadata.HTMLHash)
	}
	st.metrics.Update(false, time.Since(start))

	return nil
}

func (st *Storage) GetMemtadata(ctx context.Context) ([]models.Metadata, error) {
	return st.Metadata.Get(ctx)
}

func (st *Storage) RunScript(key, sriptHash string, ctx context.Context) (any, error) {
	return st.Cache.RunScript(key, sriptHash, ctx)
}

func (st *Storage) SaveTempWithUUID(ctx context.Context, data *models.PageDataModel) (string, error) {
	if !data.IsValid() {
		return "", utils.ErrInValidPageData
	}

	id := uuid.New().String()
	values := map[string]any{
		"metadata": data.Metadata,
		"content":  data.Content,
	}
	if err := st.Cache.SaveWithTTL(ctx, id, values, 24*time.Hour); err != nil {
		return "", err
	}
	return id, nil
}

func (st *Storage) GetTempByUUID(ctx context.Context, id string) (*models.PageDataModel, error) {
	values, err := st.Cache.Get(ctx, id)
	if err == redis.Nil {
		return nil, redis.Nil
	}
	if err != nil {
		return nil, err
	}
	var pageData models.PageDataModel
	if err := json.Unmarshal(values, &pageData); err != nil {
		return nil, err
	}

	if !pageData.IsValid() {
		return nil, utils.ErrInValidPageData
	}
	return &pageData, nil
}
