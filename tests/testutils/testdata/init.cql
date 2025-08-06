CREATE KEYSPACE IF NOT EXISTS test_keyspace
WITH REPLICATION = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 };

CREATE TABLE IF NOT EXISTS test_keyspace.metadata (
    url text PRIMARY KEY,
    host text,
    html_hash text,
    latency_ms bigint,
    time timestamp,
    content_length int
);

