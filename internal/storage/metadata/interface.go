package metadata

import (
	"context"

	"github.com/NesterovYehor/Crawler/internal/models"
)

type MetadataStore interface {
	Save(ctx context.Context, data models.Metadata) error
	Close()
	Get(ctx context.Context) ([]models.Metadata, error)
}
