CREATE KEYSPACE IF NOT EXISTS metadata
WITH REPLICATION = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 };

CREATE TABLE IF NOT EXISTS metadata.metadata (
    url text PRIMARY KEY,
    host text,
    html_hash text,
    latency_ms bigint,
    time timestamp,
    content_length int
);


