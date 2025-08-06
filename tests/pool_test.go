package tests

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/NesterovYehor/Crawler/internal/config"
	"github.com/NesterovYehor/Crawler/internal/models"
	"github.com/NesterovYehor/Crawler/internal/politeness"
	"github.com/NesterovYehor/Crawler/internal/queue"
	"github.com/NesterovYehor/Crawler/internal/storage"
	"github.com/NesterovYehor/Crawler/internal/utils"
	"github.com/NesterovYehor/Crawler/internal/wp"
	"github.com/NesterovYehor/Crawler/tests/testutils"
	"github.com/NesterovYehor/Crawler/tests/testutils/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkerPoolProcesses(t *testing.T) {
	t.Run("Worker_Pool_Processes_Test", func(t *testing.T) {
		retrySignal := make(chan struct{}, 1)

		tesdMessage := []*models.Task{models.NewTask("fetch_rules", "https://example.com/high", queue.HighPriorityQueue, "")}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*20) // Reduced timeout for faster tests
		defer cancel()
		cfg, st, q, cleanUps := setupTestEnv(ctx, t)
		defer func() {
			for _, cleanUp := range cleanUps {
				require.NoError(t, cleanUp())
			}
		}()
		require.NoError(t, q.Add(tesdMessage))

		httpMock := mocks.MockHTTPClient{
			FetchBodyFn: func(rawURL string) (io.ReadCloser, error) {
				switch rawURL {
				case "https://example.com/high":
					return io.NopCloser(strings.NewReader(`
                        <html>
                            <body>
                                <a href="https://example.com/med">Medium Link</a>
                                <a href="https://example.com/low">Low Link</a>
                            </body>
                        </html>
                    `)), nil
				case "https://example.com/med":
					return io.NopCloser(strings.NewReader("<html><body>Page 2 content</body></html>")), nil
				case "https://example.com/low":
					select {
					case <-retrySignal:
						return io.NopCloser(strings.NewReader("<html><body>Page 3 content</body></html>")), nil
					default:
						retrySignal <- struct{}{}
						return nil, utils.ErrRetryLater
					}
				default:
					return nil, fmt.Errorf("unknown URL: %s", rawURL)
				}
			},
			FetchRulesFn: func(baseURL string) ([]byte, bool, []string, error) {
				switch baseURL {
				case "example.com":
					rules := []byte(`User-agent: * Allow: /`)
					return rules, false, nil, nil
				default:
					return nil, false, nil, fmt.Errorf("unknown base URL: %s", baseURL)
				}
			},
		}

		wp, err := wp.NewWorkerPool(&wp.WorkerPoolOpts{
			Config:     &cfg.Workers,
			Queue:      q,
			ST:         st,
			PM:         politeness.NewPM(st, cfg.Scripts),
			HttpClient: &httpMock,
			Metrics:    mocks.NewNoopMetrics(),
		})

		require.NoError(t, err)

		newCtx, cancel := context.WithTimeout(ctx, time.Second*5)
		defer cancel()
		go wp.Run(newCtx)

		<-newCtx.Done()

		for _, source := range queue.FallbackOrder {
			isEmpty, err := q.IsEmpty(source, context.Background()) // Use a new context for final checks
			assert.NoError(t, err)
			assert.True(t, isEmpty, "queue "+source+" should be empty")
		}

		// 2. Check that the correct data was saved to storage.
		data, err := st.GetMemtadata(context.Background())
		assert.NoError(t, err)
		require.NotNil(t, data)
		urls := extractURLs(data)
		require.Len(t, data, 3, "Should have stored 3 pages")
		assert.Contains(t, urls, "https://example.com/high")
		assert.Contains(t, urls, "https://example.com/med")
		assert.Contains(t, urls, "https://example.com/low")
	})
}

func setupTestEnv(ctx context.Context, t *testing.T) (*config.Config, storage.Interface, queue.Interface, []func() error) {
	redisClient, cleanUp, err := testutils.RunRedis(ctx)
	cleanUps := make([]func() error, 0)
	cleanUps = append(cleanUps, cleanUp)
	require.NoError(t, err)
	require.NotNil(t, redisClient)
	cfg := &config.Config{
		Queue: &config.Queue{
			Stream:     "test_stream",
			GroupName:  "test_goroup_name",
			ConsumerID: "test",
		},
		Scripts: &config.Scripts{
			Access: "",
			Update: "",
		},
		Workers: config.Workers{
			Fetch: config.Fetch{
				HighPrioretyCount: 2, // Increased workers to handle tasks faster
				MedPrioretyCount:  2,
				LowPrioretyCount:  2,
			},
			Upload: config.Upload{
				Count: 2,
			},
		},
		DB: &config.DB{
			Addr:     "",
			Keyspace: "test_keyspace",
		},
	}
	st, cleanUp, err := testutils.SetupStorage(redisClient, cfg)
	cleanUps = append(cleanUps, cleanUp)
	require.NoError(t, err)

	q, err := queue.NewQueue(ctx, redisClient, cfg.Queue)

	require.NoError(t, err)
	return cfg, st, q, cleanUps
}

func extractURLs(metadata []models.Metadata) []string {
	urls := make([]string, 0, len(metadata))
	for _, m := range metadata {
		urls = append(urls, m.URL)
	}
	return urls
}
