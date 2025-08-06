package tests

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/NesterovYehor/Crawler/internal/config"
	"github.com/NesterovYehor/Crawler/internal/models"
	"github.com/NesterovYehor/Crawler/internal/storage/metadata"
	"github.com/NesterovYehor/Crawler/tests/testutils"
	"github.com/NesterovYehor/Crawler/tests/testutils/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCassandraStore(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	dbHost, cleanUp, err := testutils.RunCassandra(ctx)
	require.NoError(t, err)
	defer func() {
		if err := cleanUp(); err != nil {
			slog.Error(err.Error())
		}
	}()

	cfg := &config.DB{
		Addr:     dbHost,
		Keyspace: "test_keyspace",
	}
	metrics := mocks.NewNoopMetrics()

	ms, err := metadata.NewCassandraStore(cfg, metrics.Store.DBMetrics())
	require.NoError(t, err)
	defer ms.Close()

	testData := models.Metadata{
		URL:        "test_url",
		Host:       "test_host",
		HTMLHash:   "ajsfklasjfkladsjfkla",
		Latency:    19,
		Timestamp:  time.Now().Truncate(time.Millisecond),
		ContentLen: 3,
	}

	assert.NoError(t, ms.Save(ctx, testData))

	data, err := ms.Get(ctx)
	require.NoError(t, err)
	require.Len(t, data, 1)

	assert.Equal(t, testData.URL, data[0].URL)
	assert.Equal(t, testData.Host, data[0].Host)
	assert.Equal(t, testData.HTMLHash, data[0].HTMLHash)
	assert.Equal(t, testData.ContentLen, data[0].ContentLen)
	assert.WithinDuration(t, testData.Timestamp, data[0].Timestamp, time.Second)
}
