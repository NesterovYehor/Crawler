package storage

import (
	"context"

	"github.com/NesterovYehor/Crawler/internal/models"
)

type Interface interface {
	ExistsInBF(ctx context.Context, key string) (bool, error)
	AddToBF(ctx context.Context, key string) error
	SaveToCache(ctx context.Context, key string, values map[string]any) error
	RunScript(key, sriptHash string, ctx context.Context) (any, error)
	SaveIfNew(ctx context.Context, data *models.PageDataModel) error
	GetMemtadata(ctx context.Context) ([]models.Metadata, error)
	SaveTempWithUUID(ctx context.Context, data *models.PageDataModel) (string, error)
	GetTempByUUID(ctx context.Context, id string) (*models.PageDataModel, error)
}
