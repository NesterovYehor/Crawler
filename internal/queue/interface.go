package queue

import (
	"context"

	"github.com/NesterovYehor/Crawler/internal/models"
)

type Interface interface {
	Add(messages []*models.Task) error
	GetTasks(ctx context.Context, count int, sourceName string) ([]*models.Task, error)
	getMainTasks(ctx context.Context, count int, sourceName string) ([]*models.Task, error)
	Del(messages []*models.Task) error
	IsEmpty(source string, ctx context.Context) (bool, error)
	Retry(ctx context.Context, taks models.Task) error
	Close(ctx context.Context) error
	getRetryTasks(ctx context.Context, count int) ([]*models.Task, error)
}
