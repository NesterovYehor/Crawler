package testutils

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/cassandra"
	"github.com/testcontainers/testcontainers-go/wait"
)

func RunCassandra(ctx context.Context) (string, func() error, error) {

	scriptPath := filepath.Join("testutils", "testdata", "init.cql")

	container, err := cassandra.Run(
		ctx,
		"cassandra:4.1.3",
		cassandra.WithInitScripts(scriptPath),
		testcontainers.WithWaitStrategy(
			wait.ForLog("Created default superuser role 'cassandra'").
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		return "", nil, fmt.Errorf("failed to run container: %w", err)
	}

	cleanUp := func() error {
		return testcontainers.TerminateContainer(container)
	}

	db_host, err := container.ConnectionHost(ctx)
	if err != nil {
		return "", nil, err
	}
	return db_host, cleanUp, nil
}

