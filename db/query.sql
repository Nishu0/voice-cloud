-- name: CreateConnection :one
INSERT INTO connections (name, adapter, location_kind, base_url, credential_ref, enabled)
VALUES (?, ?, ?, ?, ?, ?) RETURNING *;

-- name: GetConnection :one
SELECT * FROM connections WHERE id = ?;

-- name: ListConnections :many
SELECT * FROM connections ORDER BY name, id;

-- name: UpdateConnection :one
UPDATE connections SET name = ?, adapter = ?, location_kind = ?, base_url = ?,
    credential_ref = ?, enabled = ?, updated_at = ?
WHERE id = ? RETURNING *;

-- name: DeleteConnection :execrows
DELETE FROM connections WHERE id = ?;

-- name: CreateModel :one
INSERT INTO models (name, publisher, task, version) VALUES (?, ?, ?, ?) RETURNING *;

-- name: GetModel :one
SELECT * FROM models WHERE id = ?;

-- name: ListModels :many
SELECT * FROM models ORDER BY publisher, name, task, version, id;

-- name: UpdateModel :one
UPDATE models SET name = ?, publisher = ?, task = ?, version = ?, updated_at = ?
WHERE id = ? RETURNING *;

-- name: DeleteModel :execrows
DELETE FROM models WHERE id = ?;

-- name: CreateDeployment :one
INSERT INTO deployments (name, model_id, connection_id, upstream_model_id, enabled,
    capabilities_json, pricing_json, settings_json)
VALUES (?, ?, ?, ?, ?, ?, ?, ?) RETURNING *;

-- name: GetDeployment :one
SELECT * FROM deployments WHERE id = ?;

-- name: ListDeployments :many
SELECT * FROM deployments ORDER BY name, id;

-- name: ListModelDeployments :many
SELECT * FROM deployments WHERE model_id = ? ORDER BY name, id;

-- name: UpdateDeployment :one
UPDATE deployments SET name = ?, model_id = ?, connection_id = ?, upstream_model_id = ?,
    enabled = ?, capabilities_json = ?, pricing_json = ?, settings_json = ?, updated_at = ?
WHERE id = ? RETURNING *;

-- name: DeleteDeployment :execrows
DELETE FROM deployments WHERE id = ?;

-- name: AddDeploymentLanguage :exec
INSERT INTO deployment_languages (deployment_id, language_tag) VALUES (?, ?)
ON CONFLICT (deployment_id, language_tag) DO NOTHING;

-- name: ListDeploymentLanguages :many
SELECT * FROM deployment_languages WHERE deployment_id = ? ORDER BY language_tag;

-- name: RemoveDeploymentLanguage :execrows
DELETE FROM deployment_languages WHERE deployment_id = ? AND language_tag = ?;

-- name: CreateRoutingPolicy :one
INSERT INTO routing_policies (name, task, objective, fallback_enabled, max_attempts, constraints_json)
VALUES (?, ?, ?, ?, ?, ?) RETURNING *;

-- name: GetRoutingPolicy :one
SELECT * FROM routing_policies WHERE id = ?;

-- name: ListRoutingPolicies :many
SELECT * FROM routing_policies ORDER BY name, id;

-- name: UpdateRoutingPolicy :one
UPDATE routing_policies SET name = ?, task = ?, objective = ?, fallback_enabled = ?,
    max_attempts = ?, constraints_json = ?, updated_at = ?
WHERE id = ? RETURNING *;

-- name: DeleteRoutingPolicy :execrows
DELETE FROM routing_policies WHERE id = ?;

-- name: SetPolicyTarget :exec
INSERT INTO policy_targets (policy_id, deployment_id, priority) VALUES (?, ?, ?)
ON CONFLICT (policy_id, deployment_id) DO UPDATE SET priority = excluded.priority;

-- name: ListPolicyTargets :many
SELECT * FROM policy_targets WHERE policy_id = ? ORDER BY priority, deployment_id;

-- name: ListPolicyDeployments :many
SELECT d.* FROM policy_targets t
JOIN routing_policies p ON p.id = t.policy_id
JOIN deployments d ON d.id = t.deployment_id
JOIN connections c ON c.id = d.connection_id
JOIN models m ON m.id = d.model_id
WHERE t.policy_id = ? AND d.enabled = 1 AND c.enabled = 1 AND m.task = p.task
ORDER BY t.priority, d.id;

-- name: RemovePolicyTarget :execrows
DELETE FROM policy_targets WHERE policy_id = ? AND deployment_id = ?;

-- name: CreateAPIKey :one
INSERT INTO api_keys (name, key_hash, key_prefix, default_policy_id, expires_at, created_by_user_id)
VALUES (?, ?, ?, ?, ?, ?) RETURNING *;

-- name: GetAPIKey :one
SELECT * FROM api_keys WHERE id = ?;

-- name: GetActiveAPIKeyByHash :one
SELECT * FROM api_keys WHERE key_hash = sqlc.arg(key_hash) AND revoked_at IS NULL
    AND (expires_at IS NULL OR expires_at > sqlc.arg(now_ms));

-- name: ListAPIKeys :many
SELECT * FROM api_keys ORDER BY created_at DESC, id DESC;

-- name: UpdateAPIKey :one
UPDATE api_keys SET name = ?, default_policy_id = ?, expires_at = ?, updated_at = ?
WHERE id = ? RETURNING *;

-- name: RevokeAPIKey :execrows
UPDATE api_keys SET revoked_at = sqlc.arg(now_ms), updated_at = sqlc.arg(now_ms)
WHERE id = sqlc.arg(id) AND revoked_at IS NULL;

-- name: DeleteAPIKey :execrows
DELETE FROM api_keys WHERE id = ?;

-- name: CreateRequest :one
INSERT INTO requests (id, api_key_id, task, transport, language_tag, policy_id,
    routing_snapshot_json, started_at, trace_id, idempotency_key, request_fingerprint)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING *;

-- name: GetRequest :one
SELECT * FROM requests WHERE id = ?;

-- name: GetRequestByIdempotencyKey :one
SELECT * FROM requests WHERE api_key_id = ? AND idempotency_key = ?;

-- name: ListRequests :many
SELECT * FROM requests ORDER BY started_at DESC, id DESC LIMIT ? OFFSET ?;

-- name: ListAPIKeyRequests :many
SELECT * FROM requests WHERE api_key_id = ?
ORDER BY started_at DESC, id DESC LIMIT ? OFFSET ?;

-- name: ListTraceRequests :many
SELECT * FROM requests WHERE trace_id = ? ORDER BY started_at, id;

-- name: StartRequest :execrows
UPDATE requests SET status = 'running' WHERE id = ? AND status = 'pending';

-- name: FinishRequest :execrows
UPDATE requests SET status = sqlc.arg(status), finished_at = sqlc.arg(finished_at), error_code = sqlc.narg(error_code)
WHERE id = sqlc.arg(id) AND status IN ('pending', 'running')
    AND sqlc.arg(status) <> 'pending' AND sqlc.arg(status) <> 'running';

-- name: DeleteRequest :execrows
DELETE FROM requests WHERE id = ?;

-- name: CreateAttempt :one
INSERT INTO attempts (id, request_id, deployment_id, attempt_number, started_at, selection_reason)
VALUES (?, ?, ?, ?, ?, ?) RETURNING *;

-- name: GetAttempt :one
SELECT * FROM attempts WHERE id = ?;

-- name: ListRequestAttempts :many
SELECT * FROM attempts WHERE request_id = ? ORDER BY attempt_number;

-- name: SetAttemptProviderRequestID :execrows
UPDATE attempts SET provider_request_id = ? WHERE id = ? AND status = 'running';

-- name: SetAttemptFirstOutput :execrows
UPDATE attempts SET first_output_at = ?
WHERE id = ? AND status = 'running' AND first_output_at IS NULL;

-- name: FinishAttempt :execrows
UPDATE attempts SET status = sqlc.arg(status), finished_at = sqlc.arg(finished_at), error_code = sqlc.narg(error_code)
WHERE id = sqlc.arg(id) AND status = 'running'
    AND sqlc.arg(status) <> 'running';

-- name: DeleteAttempt :execrows
DELETE FROM attempts WHERE id = ?;

-- name: CreateAttemptUsage :one
INSERT INTO attempt_usage (attempt_id, metric, quantity, unit, rate_snapshot_json,
    cost_nanos, currency, cost_status)
VALUES (?, ?, ?, ?, ?, ?, ?, ?) RETURNING *;

-- name: GetAttemptUsage :one
SELECT * FROM attempt_usage WHERE attempt_id = ? AND metric = ?;

-- name: ListAttemptUsage :many
SELECT * FROM attempt_usage WHERE attempt_id = ? ORDER BY metric;

-- name: ListRequestUsage :many
SELECT u.* FROM attempt_usage u JOIN attempts a ON a.id = u.attempt_id
WHERE a.request_id = ? ORDER BY a.attempt_number, u.metric;

-- name: UpdateAttemptUsage :one
UPDATE attempt_usage SET quantity = ?, cost_nanos = ?, cost_status = ?, updated_at = ?
WHERE attempt_id = ? AND metric = ? RETURNING *;

-- name: DeleteAttemptUsage :execrows
DELETE FROM attempt_usage WHERE attempt_id = ? AND metric = ?;

-- name: CreateUser :one
INSERT INTO users (email, name, password_hash, role)
VALUES (?, ?, ?, ?) RETURNING *;

-- name: GetUser :one
SELECT * FROM users WHERE id = ?;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = ?;

-- name: ListUsers :many
SELECT id, email, name, role, status, email_verified_at, created_at, updated_at
FROM users ORDER BY created_at DESC, id DESC LIMIT ? OFFSET ?;

-- name: UpdateUser :one
UPDATE users SET name = ?, role = ?, updated_at = ? WHERE id = ? RETURNING *;

-- name: SetUserStatus :execrows
UPDATE users SET status = ?, updated_at = ? WHERE id = ?;

-- name: UpdateUserPassword :execrows
UPDATE users SET password_hash = ?, updated_at = ? WHERE id = ?;

-- name: VerifyUserEmail :execrows
UPDATE users SET email_verified_at = sqlc.arg(now_ms), updated_at = sqlc.arg(now_ms)
WHERE id = sqlc.arg(id) AND email_verified_at IS NULL;

-- name: DeleteUser :execrows
DELETE FROM users WHERE id = ?;

-- name: CreateUserSession :exec
INSERT INTO user_sessions (token_hash, user_id, created_at, expires_at)
VALUES (?, ?, ?, ?);

-- name: GetActiveUserSession :one
SELECT s.* FROM user_sessions s JOIN users u ON u.id = s.user_id
WHERE s.token_hash = sqlc.arg(token_hash) AND s.revoked_at IS NULL
    AND s.created_at <= sqlc.arg(now_ms) AND s.expires_at > sqlc.arg(now_ms)
    AND u.status = 'active';

-- name: RevokeUserSession :execrows
UPDATE user_sessions SET revoked_at = sqlc.arg(now_ms)
WHERE token_hash = sqlc.arg(token_hash) AND revoked_at IS NULL;

-- name: DeleteUserSessions :execrows
DELETE FROM user_sessions WHERE user_id = ?;

-- name: DeleteExpiredUserSessions :execrows
DELETE FROM user_sessions WHERE expires_at <= ?;

-- name: CreateUserToken :exec
INSERT INTO user_tokens (token_hash, user_id, purpose, created_at, expires_at)
VALUES (?, ?, ?, ?, ?);

-- name: ConsumeUserToken :one
UPDATE user_tokens SET used_at = sqlc.arg(now_ms)
WHERE token_hash = sqlc.arg(token_hash) AND purpose = sqlc.arg(purpose)
    AND used_at IS NULL AND created_at <= sqlc.arg(now_ms)
    AND expires_at > sqlc.arg(now_ms)
    AND user_id IN (SELECT id FROM users WHERE status = 'active')
RETURNING *;

-- name: DeleteUserTokens :execrows
DELETE FROM user_tokens WHERE user_id = ? AND purpose = ?;

-- name: DeleteExpiredUserTokens :execrows
DELETE FROM user_tokens WHERE expires_at <= ?;

-- name: ListHealthDeploymentIDs :many
SELECT d.id FROM deployments d JOIN connections c ON c.id = d.connection_id
WHERE d.enabled = 1 AND c.enabled = 1 ORDER BY d.id;

-- name: GetDeploymentHealth :one
SELECT * FROM deployment_health WHERE deployment_id = ?;

-- name: SetDeploymentHealth :execrows
INSERT INTO deployment_health
    (deployment_id, probe_kind, status, checked_at, expires_at, latency_ms, error_code, config_hash)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT (deployment_id) DO UPDATE SET
    probe_kind = excluded.probe_kind, status = excluded.status,
    checked_at = excluded.checked_at, expires_at = excluded.expires_at,
    latency_ms = excluded.latency_ms, error_code = excluded.error_code,
    config_hash = excluded.config_hash
WHERE excluded.checked_at >= deployment_health.checked_at;
