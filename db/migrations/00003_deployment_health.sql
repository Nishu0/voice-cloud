-- +goose Up
CREATE TABLE deployment_health (
    deployment_id INTEGER PRIMARY KEY REFERENCES deployments(id) ON DELETE CASCADE,
    probe_kind TEXT NOT NULL CHECK (probe_kind = 'reachability'),
    status TEXT NOT NULL CHECK (status IN ('healthy', 'unhealthy', 'unknown')),
    checked_at INTEGER NOT NULL,
    expires_at INTEGER NOT NULL CHECK (expires_at > checked_at),
    latency_ms INTEGER CHECK (latency_ms IS NULL OR latency_ms >= 0),
    error_code TEXT,
    config_hash TEXT NOT NULL CHECK (length(config_hash) = 64)
);

-- +goose Down
DROP TABLE deployment_health;
