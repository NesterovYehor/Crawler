package metadata

import (
	"context"
	"fmt"
	"time"

	"github.com/NesterovYehor/Crawler/internal/config"
	"github.com/NesterovYehor/Crawler/internal/metrics"
	"github.com/NesterovYehor/Crawler/internal/models"
	"github.com/gocql/gocql"
)

type cassandraStore struct {
	session *gocql.Session
	metrics metrics.DBMetrics
}
func NewCassandraStore(cfg *config.DB, metrics metrics.DBMetrics) (MetadataStore, error) {
	cluster := gocql.NewCluster(cfg.Addr)
	cluster.Consistency = gocql.Quorum
	cluster.ProtoVersion = 4

	tempSess, err := cluster.CreateSession() 
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary session for keyspace DDL: %w", err)
	}
	defer tempSess.Close() 

	err = tempSess.Query(`CREATE KEYSPACE IF NOT EXISTS metadata WITH REPLICATION = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 };`).Exec()
	if err != nil {
		return nil, fmt.Errorf("failed to create new keyspace: %w", err)
	}

	cluster.Keyspace = "metadata" 
	
	sess, err := cluster.CreateSession() 
	if err != nil {
		return nil, fmt.Errorf("failed to create session for metadata keyspace: %w", err)
	}

	err = sess.Query(`
		CREATE TABLE IF NOT EXISTS metadata.metadata (
			url text PRIMARY KEY,
			host text,
			html_hash text,
			latency_ms bigint,
			time timestamp,
			content_length int
		);
	`).Exec()
	if err != nil {
		return nil, fmt.Errorf("failed to create new table: %w", err)
	}

	return &cassandraStore{
		session: sess,
		metrics: metrics,
	}, nil
}
func (c *cassandraStore) Close() {
	c.session.Close()
}

func (c *cassandraStore) Save(ctx context.Context, data models.Metadata) error {
	start := time.Now()
	queue := `
        insert into metadata (url, host, html_hash, latency_ms, time, content_length) values (?,?,?,?,?, ?)
    `

	if err := c.session.Query(queue, data.URL, data.Host, data.HTMLHash, int64(data.Latency), data.Timestamp, data.ContentLen).Exec(); err != nil {
		c.metrics.Update(true, time.Since(start))
		return err
	}

	c.metrics.Update(false, time.Since(start))

	return nil
}

func (c *cassandraStore) Get(ctx context.Context) ([]models.Metadata, error) {
	var results []models.Metadata

	iter := c.session.Query(`SELECT url, host, html_hash, latency_ms, time, content_length FROM metadata`).Iter()

	var m models.Metadata
	var latencyMs int64

	for iter.Scan(&m.URL, &m.Host, &m.HTMLHash, &latencyMs, &m.Timestamp, &m.ContentLen) {
		m.Latency = models.Latency(time.Duration(latencyMs) * time.Millisecond)
		results = append(results, m)
	}

	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("failed to query: %v", err)
	}
	return results, nil
}
