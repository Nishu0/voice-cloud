-- +goose Up
-- Enable foreign_keys on every runtime connection; set WAL when opening the database.
-- Timestamps are Unix milliseconds. Supply precise request/attempt times at runtime.
-- Provider credentials stay outside SQLite; credential_ref names an external secret.

CREATE TABLE connections (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE CHECK (length(trim(name)) > 0),
    adapter TEXT NOT NULL CHECK (length(trim(adapter)) > 0),
    location_kind TEXT NOT NULL CHECK (location_kind IN ('local', 'cloud')),
    base_url TEXT NOT NULL CHECK (length(trim(base_url)) > 0),
    credential_ref TEXT,
    enabled INTEGER NOT NULL DEFAULT 1 CHECK (enabled IN (0, 1)),
    created_at INTEGER NOT NULL DEFAULT (unixepoch() * 1000),
    updated_at INTEGER NOT NULL DEFAULT (unixepoch() * 1000)
);

CREATE TABLE models (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL CHECK (length(trim(name)) > 0),
    publisher TEXT NOT NULL CHECK (length(trim(publisher)) > 0),
    task TEXT NOT NULL CHECK (task IN ('stt', 'tts', 'llm')),
    version TEXT NOT NULL DEFAULT '',
    created_at INTEGER NOT NULL DEFAULT (unixepoch() * 1000),
    updated_at INTEGER NOT NULL DEFAULT (unixepoch() * 1000),
    UNIQUE (publisher, name, task, version)
);

CREATE TABLE deployments (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE CHECK (length(trim(name)) > 0),
    model_id INTEGER NOT NULL REFERENCES models(id),
    connection_id INTEGER NOT NULL REFERENCES connections(id),
    upstream_model_id TEXT NOT NULL CHECK (length(trim(upstream_model_id)) > 0),
    enabled INTEGER NOT NULL DEFAULT 1 CHECK (enabled IN (0, 1)),
    capabilities_json TEXT NOT NULL DEFAULT '{}'
        CHECK (json_valid(capabilities_json) AND json_type(capabilities_json) = 'object'),
    -- Pricing is adapter-specific; the applied rate is copied into attempt_usage.
    pricing_json TEXT NOT NULL DEFAULT '{}'
        CHECK (json_valid(pricing_json) AND json_type(pricing_json) = 'object'),
    settings_json TEXT NOT NULL DEFAULT '{}'
        CHECK (json_valid(settings_json) AND json_type(settings_json) = 'object'),
    created_at INTEGER NOT NULL DEFAULT (unixepoch() * 1000),
    updated_at INTEGER NOT NULL DEFAULT (unixepoch() * 1000)
);

CREATE INDEX deployments_model ON deployments(model_id);
CREATE INDEX deployments_connection ON deployments(connection_id);

CREATE TABLE deployment_languages (
    deployment_id INTEGER NOT NULL REFERENCES deployments(id) ON DELETE CASCADE,
    language_tag TEXT NOT NULL CHECK (length(trim(language_tag)) > 0),
    PRIMARY KEY (deployment_id, language_tag)
);

CREATE TABLE routing_policies (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE CHECK (length(trim(name)) > 0),
    task TEXT NOT NULL CHECK (task IN ('stt', 'tts', 'llm')),
    objective TEXT NOT NULL DEFAULT 'balanced'
        CHECK (objective IN ('balanced', 'latency', 'cost', 'quality')),
    fallback_enabled INTEGER NOT NULL DEFAULT 1 CHECK (fallback_enabled IN (0, 1)),
    max_attempts INTEGER NOT NULL DEFAULT 2 CHECK (max_attempts >= 1),
    constraints_json TEXT NOT NULL DEFAULT '{}'
        CHECK (json_valid(constraints_json) AND json_type(constraints_json) = 'object'),
    created_at INTEGER NOT NULL DEFAULT (unixepoch() * 1000),
    updated_at INTEGER NOT NULL DEFAULT (unixepoch() * 1000)
);

CREATE TABLE policy_targets (
    policy_id INTEGER NOT NULL REFERENCES routing_policies(id) ON DELETE CASCADE,
    deployment_id INTEGER NOT NULL REFERENCES deployments(id),
    priority INTEGER NOT NULL DEFAULT 0 CHECK (priority >= 0),
    PRIMARY KEY (policy_id, deployment_id)
);

CREATE INDEX policy_targets_order ON policy_targets(policy_id, priority, deployment_id);
CREATE INDEX policy_targets_deployment ON policy_targets(deployment_id);

CREATE TABLE api_keys (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL CHECK (length(trim(name)) > 0),
    key_hash TEXT NOT NULL UNIQUE CHECK (length(trim(key_hash)) > 0),
    key_prefix TEXT NOT NULL CHECK (length(trim(key_prefix)) > 0),
    default_policy_id INTEGER REFERENCES routing_policies(id),
    expires_at INTEGER,
    revoked_at INTEGER,
    created_at INTEGER NOT NULL DEFAULT (unixepoch() * 1000),
    updated_at INTEGER NOT NULL DEFAULT (unixepoch() * 1000)
);

CREATE INDEX api_keys_default_policy ON api_keys(default_policy_id);

CREATE TABLE requests (
    id TEXT PRIMARY KEY NOT NULL CHECK (length(trim(id)) > 0),
    api_key_id INTEGER NOT NULL REFERENCES api_keys(id),
    task TEXT NOT NULL CHECK (task IN ('stt', 'tts', 'llm')),
    transport TEXT NOT NULL CHECK (transport IN ('http', 'sse', 'websocket')),
    language_tag TEXT,
    policy_id INTEGER REFERENCES routing_policies(id),
    -- Includes explicit routing, client overrides, and the effective policy settings.
    routing_snapshot_json TEXT NOT NULL
        CHECK (json_valid(routing_snapshot_json) AND json_type(routing_snapshot_json) = 'object'),
    status TEXT NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'running', 'succeeded', 'failed', 'cancelled')),
    started_at INTEGER NOT NULL,
    finished_at INTEGER,
    error_code TEXT,
    trace_id TEXT,
    -- Claim this key before dispatch; duplicate streams are not replayed.
    idempotency_key TEXT CHECK (idempotency_key IS NULL OR length(trim(idempotency_key)) > 0),
    request_fingerprint TEXT,
    CHECK (finished_at IS NULL OR finished_at >= started_at),
    CHECK (idempotency_key IS NULL OR
        (request_fingerprint IS NOT NULL AND length(trim(request_fingerprint)) > 0)),
    UNIQUE (api_key_id, idempotency_key)
);

CREATE INDEX requests_started ON requests(started_at);
CREATE INDEX requests_key_started ON requests(api_key_id, started_at);
CREATE INDEX requests_policy ON requests(policy_id);
CREATE INDEX requests_trace ON requests(trace_id) WHERE trace_id IS NOT NULL;

CREATE TABLE attempts (
    id TEXT PRIMARY KEY NOT NULL CHECK (length(trim(id)) > 0),
    request_id TEXT NOT NULL REFERENCES requests(id),
    deployment_id INTEGER NOT NULL REFERENCES deployments(id),
    attempt_number INTEGER NOT NULL CHECK (attempt_number >= 1),
    provider_request_id TEXT,
    status TEXT NOT NULL DEFAULT 'running'
        CHECK (status IN ('running', 'succeeded', 'failed', 'cancelled')),
    started_at INTEGER NOT NULL,
    first_output_at INTEGER,
    finished_at INTEGER,
    error_code TEXT,
    selection_reason TEXT NOT NULL CHECK (length(trim(selection_reason)) > 0),
    CHECK (first_output_at IS NULL OR first_output_at >= started_at),
    CHECK (finished_at IS NULL OR finished_at >= started_at),
    CHECK (first_output_at IS NULL OR finished_at IS NULL OR first_output_at <= finished_at),
    UNIQUE (request_id, attempt_number)
);

CREATE INDEX attempts_deployment_started ON attempts(deployment_id, started_at);

CREATE TABLE attempt_usage (
    attempt_id TEXT NOT NULL REFERENCES attempts(id),
    metric TEXT NOT NULL CHECK (length(trim(metric)) > 0),
    quantity INTEGER NOT NULL CHECK (typeof(quantity) = 'integer' AND quantity >= 0),
    unit TEXT NOT NULL CHECK (unit IN ('audio_ms', 'characters', 'tokens', 'requests')),
    rate_snapshot_json TEXT NOT NULL DEFAULT '{}'
        CHECK (json_valid(rate_snapshot_json) AND json_type(rate_snapshot_json) = 'object'),
    -- One nanodollar is 10^-9 USD; NULL means unknown, not free.
    cost_nanos INTEGER CHECK (cost_nanos IS NULL OR
        (typeof(cost_nanos) = 'integer' AND cost_nanos >= 0)),
    currency TEXT NOT NULL DEFAULT 'USD' CHECK (currency = 'USD'),
    cost_status TEXT NOT NULL DEFAULT 'unknown'
        CHECK (cost_status IN ('unknown', 'estimated', 'final')),
    created_at INTEGER NOT NULL DEFAULT (unixepoch() * 1000),
    updated_at INTEGER NOT NULL DEFAULT (unixepoch() * 1000),
    CHECK ((cost_status = 'unknown' AND cost_nanos IS NULL) OR
        (cost_status IN ('estimated', 'final') AND cost_nanos IS NOT NULL)),
    PRIMARY KEY (attempt_id, metric)
);

-- +goose Down
DROP TABLE attempt_usage;
DROP TABLE attempts;
DROP TABLE requests;
DROP TABLE api_keys;
DROP TABLE policy_targets;
DROP TABLE routing_policies;
DROP TABLE deployment_languages;
DROP TABLE deployments;
DROP TABLE models;
DROP TABLE connections;
