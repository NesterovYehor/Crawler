package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/NesterovYehor/Crawler/internal/config"
	"github.com/NesterovYehor/Crawler/internal/models"
	"github.com/NesterovYehor/Crawler/internal/utils"
	"github.com/redis/go-redis/v9"
)

type queue struct {
	client     *redis.Client
	consumerID string
	groupName  string
}

func NewQueue(ctx context.Context, client *redis.Client, cfg *config.Queue) (Interface, error) {
	for _, source := range FallbackOrder {
		if source == RetryPriorityQueue {
			continue
		}
		err := client.XGroupCreateMkStream(ctx, source, "workers", "0").Err()
		if err != nil {
			if strings.Contains(err.Error(), "BUSYGROUP") {
				continue
			}
			return nil, fmt.Errorf("failed to create group for %s: %v", source, err)
		}
	}

	return &queue{
		client:     client,
		consumerID: cfg.ConsumerID,
		groupName:  "workers",
	}, nil
}

func (q *queue) Add(messages []*models.Task) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	pipe := q.client.Pipeline()
	defer func() {
		_, err := pipe.Exec(ctx)
		if err != nil {
			return
		}
	}()
	for _, m := range messages {
		if m == nil {
			continue
		}
		val, err := m.Encode()
		if err != nil {
			continue
		}

		if err := q.client.XAdd(ctx, &redis.XAddArgs{
			Stream: string(m.SourceName),
			Values: val,
		}).Err(); err != nil {
			continue
		}
	}
	return nil
}

func (q *queue) GetTasks(ctx context.Context, count int, source string) ([]*models.Task, error) {
	if count <= 0 {
		return nil, nil
	}
	if source == RetryPriorityQueue {
		return q.getRetryTasks(ctx, count)
	} else {
		return q.getMainTasks(ctx, count, source)
	}
}

func (q *queue) getMainTasks(ctx context.Context, count int, source string) ([]*models.Task, error) {
	res, err := q.client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    q.groupName,
		Consumer: q.consumerID,
		Streams:  []string{source, ">"},
		Block:    time.Millisecond * 100,
		Count:    int64(count),
		NoAck:    true,
	}).Result()

	if err == redis.Nil {
		return nil, utils.ErrNoTasks
	} else if err != nil {
		return nil, err
	}
	if len(res) >= 1 {
		messages := make([]*models.Task, count)
		for i, msg := range res[0].Messages {
			m := &models.Task{}
			err := m.Decode(msg.Values, source)
			if err != nil {
				return nil, err
			}
			m.ID = msg.ID
			messages[i] = m
		}
		return messages, nil
	}
	return nil, nil
}

func (q *queue) getRetryTasks(ctx context.Context, count int) ([]*models.Task, error) {
	members, err := q.fetchRetryMembers(ctx, count)
	if err != nil {
		return nil, err
	}
	if len(members) == 0 {
		return nil, utils.ErrNoTasks
	}

	tasks := make([]*models.Task, len(members))
	for i, memberJSON := range members {
		task := &models.Task{}
		if err := json.Unmarshal([]byte(memberJSON), task); err != nil {
			slog.Error(err.Error())
			continue
		}
		tasks[i] = task
	}

	return tasks, nil
}

func (q *queue) fetchRetryMembers(ctx context.Context, count int) ([]string, error) {
	currentUnixTime := time.Now().Unix()

	zmembers, err := q.client.ZPopMin(ctx, RetryPriorityQueue, int64(count)).Result()
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to atomically pop retry tasks from queue (ZPopMin): %w", err)
	}

	if len(zmembers) == 0 {
		return nil, utils.ErrNoTasks
	}
	var dueTasksJSON []string

	readdPipe := q.client.Pipeline()
	readdedCount := 0

	for _, zmember := range zmembers {
		memberJSON, ok := zmember.Member.(string)
		if !ok {
			continue
		}
		if int64(zmember.Score) <= currentUnixTime {
			dueTasksJSON = append(dueTasksJSON, memberJSON)
		} else {
			readdPipe.ZAdd(ctx, RetryPriorityQueue, redis.Z{
				Score:  zmember.Score,
				Member: zmember.Member,
			})
			readdedCount++
		}
	}

	if readdedCount > 0 {
		_, err := readdPipe.Exec(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to re-add %d non-due tasks to retry queue: %w", readdedCount, err)
		}
	}

	return dueTasksJSON, nil
}

func (q *queue) Del(messages []*models.Task) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	pipe := q.client.Pipeline()
	defer func() {
		_, err := pipe.Exec(ctx)
		if err != nil {
			slog.Error(err.Error())
			return
		}
	}()
	for _, m := range messages {
		pipe.XDel(ctx, string(m.SourceName), m.ID)
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (q *queue) Retry(ctx context.Context, taks models.Task) error {
	taks.SourceName = RetryPriorityQueue
	taks.Retries++
	taks.CountNextAttemptAt()
	data, err := taks.EncodeToStr()
	if err != nil {
		return err
	}
	z := redis.Z{
		Score:  float64(taks.NextAttemptAt),
		Member: data,
	}
	err = q.client.ZAdd(ctx, RetryPriorityQueue, z).Err()
	if err != nil {
		return fmt.Errorf("failed to store retry task: %v", err)
	}
	return nil
}

func (q *queue) Close(ctx context.Context) error {
	for _, source := range FallbackOrder {

		if err := q.client.XGroupDestroy(ctx, source, q.groupName).Err(); err != nil {
			return fmt.Errorf("failed to destroy consumer group: %w", err)
		}

		if err := q.client.Del(ctx, source).Err(); err != nil {
			return fmt.Errorf("failed to delete stream: %w", err)
		}
	}

	return nil
}

// Needs only for tests
func (q *queue) IsEmpty(source string, ctx context.Context) (bool, error) {
	if source == RetryPriorityQueue {
		return true, nil
	}
	n, err := q.client.XLen(ctx, source).Result()
	if err != nil && err != redis.Nil {
		return false, err
	}
	if n != 0 {
		res, err := q.client.XRead(ctx, &redis.XReadArgs{
			Streams: []string{source, "0"},
			Block:   time.Millisecond * 100,
			Count:   int64(n),
		}).Result()
		if err != nil {
			return false, err
		}
		slog.Info(fmt.Sprintf("[Is Queue Empty]Tasks that was not proceesed:%v", res))
	}
	return n == 0, nil
}
