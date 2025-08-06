package models

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strconv"
	"time"

	"github.com/NesterovYehor/Crawler/internal/utils"
	"github.com/go-viper/mapstructure/v2"
)

const (
	baseDelay  = 2
	multiplier = 2
)

type Task struct {
	ID            string `mapstructure:"id"`
	NextAttemptAt int64  `mapstructure:"next_attempt_at"`
	Topic         string `mapstructure:"topic"`
	Retries       int    `mapstructure:"retries"`
	URL           string `mapstructure:"url"`
	DataID        string `mapstructure:"data_id"`
	SourceName    string
}

func NewTask(topic, url string, source string, dataID string) *Task {
	return &Task{
		Topic:         topic,
		Retries:       0,
		NextAttemptAt: 0,
		URL:           url,
		SourceName:    source,
		DataID:        dataID,
	}
}

func (m *Task) Decode(val map[string]any, source string) error {
	if retries, ok := val["retries"].(string); ok {
		r, err := strconv.Atoi(retries)
		if err != nil {
			return utils.ErrInvalidRetriesValue(err)
		}
		val["retries"] = r
	}
	if n, ok := val["next_attempt_at"].(string); ok {
		nextAttemptAt, err := strconv.Atoi(n)
		if err != nil {
			return utils.ErrInvalidRetriesValue(err)
		}
		val["next_attempt_at"] = int64(nextAttemptAt)
	}

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:  &m,
		TagName: "mapstructure",
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeHookFunc(time.RFC3339),
		),
	})
	if err != nil {
		return fmt.Errorf("failed to create decoder: %w", err)
	}

	if err := decoder.Decode(val); err != nil {
		return fmt.Errorf("Task parsing failed: %w", err)
	}
	m.SourceName = source
	return nil
}

func (t *Task) EncodeToStr() (string, error) {
	str, err := json.Marshal(t)
	if err != nil {
		return "", fmt.Errorf("failed to encode task: %v", err)
	}
	return string(str), nil
}

func (m *Task) Encode() (map[string]any, error) {
	if !m.IsValid() {
		return nil, utils.ErrInvalidTaskFormat(m)
	}
	return map[string]any{
		"topic":   m.Topic,
		"retries": m.Retries,
		"url":     m.URL,
		"data_id": m.DataID,
		"backoff": m.NextAttemptAt,
	}, nil
}

func (m *Task) IsValid() bool {
	if m == nil || m.Retries >= 5 || m.Retries < 0 || m.Topic == "" || m.URL == "" {
		return false
	}
	return true
}

func (t *Task) CountNextAttemptAt() {
	newBackoffDuration := time.Duration(int(float64(baseDelay)*math.Pow(float64(multiplier), float64(t.Retries-1)))) * time.Second

	t.NextAttemptAt = time.Now().Add(newBackoffDuration).Unix()
	log.Printf("Next attempt at: %v And new backoff:%v", time.Unix(t.NextAttemptAt, 0), newBackoffDuration)
}
