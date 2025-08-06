package models_test

import (
	"errors"
	"testing"

	"github.com/NesterovYehor/Crawler/internal/models"
	"github.com/NesterovYehor/Crawler/internal/queue"
	"github.com/NesterovYehor/Crawler/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestEncodeMessage(t *testing.T) {
	testcases := []struct {
		name     string
		msg      *models.Task
		err      error
		expected map[string]any
	}{
		{
			name: "successful test",
			msg:  models.NewTask("test-topic", "test.com", queue.HighPriorityQueue, "data-123"),
			err:  nil,
			expected: map[string]any{
				"topic":   "test-topic",
				"retries": 0,
				"backoff": int64(0),
				"url":     "test.com",
				"data_id": "data-123",
			},
		},
		{
			name:     "invalid Task - no topic",
			msg:      models.NewTask("", "test.com", queue.HighPriorityQueue, ""),
			err:      utils.ErrInvalidTaskFormat(models.NewTask("", "test.com", queue.HighPriorityQueue, "")),
			expected: nil,
		},
		{
			name:     "invalid Task - no url",
			msg:      models.NewTask("test-topic", "", queue.HighPriorityQueue, ""),
			err:      utils.ErrInvalidTaskFormat(models.NewTask("test-topic", "", queue.HighPriorityQueue, "")),
			expected: nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := tc.msg.Encode()
			assert.Equal(t, tc.err, err)
			if tc.expected != nil {
				assert.Equal(t, tc.expected, actual)
			}
		})
	}
}

func TestDecodeMessage(t *testing.T) {
	testcases := []struct {
		name     string
		input    map[string]any
		err      error
		expected *models.Task
	}{
		{
			name: "successful test",
			input: map[string]any{
				"topic":   "test-topic",
				"retries": float64(0), // Simulate the map decode which produces float64
				"backoff": float64(0),
				"url":     "test.com",
				"data_id": "data-123",
			},
			err:      nil,
			expected: models.NewTask("test-topic", "test.com", queue.HighPriorityQueue, "data-123"),
		},
		{
			name: "missing retries value",
			input: map[string]any{
				"topic":   "test-topic",
				"url":     "test.com",
				"data_id": "data-123",
			},
			err:      nil,
			expected: models.NewTask("test-topic", "test.com", queue.HighPriorityQueue, "data-123"),
		},
		{
			name: "invalid retries type",
			input: map[string]any{
				"topic":   "test-topic",
				"retries": "not-a-number",
				"url":     "test.com",
				"data_id": "data-123",
			},
			err:      errors.New("retries value is not a valid number"),
			expected: nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			msg := &models.Task{}
			err := msg.Decode(tc.input, queue.HighPriorityQueue)

			if tc.err != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tc.expected != nil {
				assert.Equal(t, tc.expected, msg)
			}
		})
	}
}
