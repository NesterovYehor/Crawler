package tests

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/NesterovYehor/Crawler/internal/config"
	"github.com/NesterovYehor/Crawler/internal/politeness"
	"github.com/NesterovYehor/Crawler/internal/storage"
	"github.com/NesterovYehor/Crawler/internal/storage/cache"
	"github.com/NesterovYehor/Crawler/tests/testutils"
	"github.com/NesterovYehor/Crawler/tests/testutils/mocks"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPolitenessManager(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cfg := &config.Config{
		Scripts: &config.Scripts{
			Access: "",
			Update: "",
		},
	}
	cleanUpFuncs := make([]func() error, 0)
	redisClient, cleanUp, err := testutils.RunRedis(ctx)
	require.NoError(t, err)
	cleanUpFuncs = append(cleanUpFuncs, cleanUp)
	metrics := mocks.NewNoopMetrics()
	cache, err := cache.NewCache(redisClient, metrics.Store.CacheMetrics())
	require.NoError(t, err)
	require.NoError(t, loadScripts(redisClient, cfg.Scripts))

	st := storage.Storage{
		Cache: cache,
	}
	cleanUpFuncs = append(cleanUpFuncs, cleanUp)

	assert.NoError(t, err)
	defer func() {
		for _, cleanUp := range cleanUpFuncs {
			if err := cleanUp(); err != nil {
				assert.NoError(t, err)
			}
		}
	}()

	pm := politeness.NewPM(&st, cfg.Scripts)
	testcases := []struct {
		host    string
		rules   string
		allowed bool
	}{
		{
			host: "google.com",
			rules: `User-agent: *
                Disallow: /lessons/
                Disallow: /course/
                Disallow: /project/
                Disallow: /certificate/
                Disallow: /lesson-difficulty`,
			allowed: true,
		},
	}

	for _, tc := range testcases {
		assert.NoError(t, pm.SaveRules(ctx, tc.host, tc.rules))
		r, err := pm.GetRules(tc.host, ctx)
		assert.NoError(t, err)
		assert.NotNil(t, r)
		assert.True(t, r.Allowed)
		assert.Equal(t, tc.rules, r.Rules)
		assert.NoError(t, pm.UpdateHostLimit(tc.host, ctx))
		r, err = pm.GetRules(tc.host, ctx)
		assert.NoError(t, err)
		assert.False(t, r.Allowed)
		time.Sleep(5 * time.Second)
		r, err = pm.GetRules(tc.host, ctx)
		assert.NoError(t, err)
		assert.NotNil(t, r)
		assert.True(t, r.Allowed)
		assert.Equal(t, tc.rules, r.Rules)
	}
}

func loadScripts(client *redis.Client, cfg *config.Scripts) error {
	scriptByts, err := os.ReadFile("../scripts/politeness_gate.lua")
	if err != nil {
		return err
	}
	if scriptByts == nil {
		return err
	}
	cfg.Access, err = client.ScriptLoad(context.Background(), string(scriptByts)).Result()
	if err != nil {
		return err
	}
	scriptByts, err = os.ReadFile("../scripts/update_token_limit.lua")
	if err != nil {
		return err
	}
	if scriptByts == nil {
		return err
	}

	cfg.Update, err = client.ScriptLoad(context.Background(), string(scriptByts)).Result()
	if err != nil {
		return err
	}
	return nil
}
