package tests

import (
	"context"
	"testing"

	"github.com/NesterovYehor/Crawler/internal/config"
	"github.com/NesterovYehor/Crawler/internal/models"
	"github.com/NesterovYehor/Crawler/internal/queue"
	"github.com/NesterovYehor/Crawler/tests/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueueLifecycle(t *testing.T) {
	testCases := []struct {
		name    string
		message *models.Task
		err     error
	}{
		{
			name:    "success test",
			message: models.NewTask("test-topic", "https://example.com", queue.HighPriorityQueue, ""),
		},
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	client, cleanUp, err := testutils.RunRedis(ctx)
	require.NoError(t, err)

	qCfg := &config.Queue{
		ConsumerID: "test-consumer",
	}

	q, err := queue.NewQueue(ctx, client, qCfg)
	assert.NoError(t, err)

	defer func() {
		if err := cleanUp(); err != nil {
			t.Fatalf(err.Error())
		}
	}()
	for _, tc := range testCases {
		msg := tc.message
		err = q.Add([]*models.Task{msg})
		assert.NoError(t, err)

		got, err := q.GetTasks(ctx, 1, queue.HighPriorityQueue)
		assert.NoError(t, err)
		assert.Len(t, got, 1)
		assert.Equal(t, "test-topic", got[0].Topic)
		assert.Equal(t, "https://example.com", got[0].URL)

		err = q.Del(got)
		assert.NoError(t, err)

	}
}
