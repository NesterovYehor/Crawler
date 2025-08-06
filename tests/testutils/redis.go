package testutils

import (
	"context"
	"net/url"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/redis"

	r "github.com/redis/go-redis/v9"
)

func RunRedis(ctx context.Context) (*r.Client, func() error, error) {
	rc, err := redis.Run(ctx, "redislabs/rebloom:latest")
	if err != nil {
		return nil, nil, err
	}

	addr, err := rc.ConnectionString(ctx)
	if err != nil {
		return nil, nil, err
	}

	u, err := url.Parse(addr)
	if err != nil {
		return nil, nil, err
	}

	client := r.NewClient(&r.Options{
		Addr: u.Host,
	})
	close := func() error {
		return testcontainers.TerminateContainer(rc)
	}
	return client, close, nil
}
