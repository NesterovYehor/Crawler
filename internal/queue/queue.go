package queue

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type Queue struct {
	client *redis.Client
	stream string
}

func NewQueue(client *redis.Client, stream string) (*Queue, error) {
	if err := client.XGroupCreate(context.Background(), stream, "workers", "").Err(); err != nil {
		return nil, err
	}
	return &Queue{
		client: client,
		stream: stream,
	}, nil
}

func (q *Queue) AddToQueue(values map[string]any) error {
	return q.client.XAdd(context.Background(), &redis.XAddArgs{
		Stream: q.stream,
		Values: values,
	}).Err()
}

func (q *Queue) FetchFromQueue() ([]redis.XStream, error) {
	res, err := q.client.XReadGroup(context.Background(), &redis.XReadGroupArgs{
		Streams: []string{q.stream},
		Group:   "workers",
	}).Result()
	if err != nil {
		return nil, err
	}
	return res, nil
}
