# Background worker

The worker runs alongside the API in the backend container, using the same
`DB_PATH` and shared SQLite opener. It never applies migrations itself. The
backend entry point applies migrations and seeds providers before starting both
processes. For local development, `make worker-run` migrates first and then starts
the worker; `make worker-check` runs its tests with race detection.

Two jobs run immediately on startup:

| Job | Schedule | Behavior |
| --- | --- | --- |
| Auth cleanup | Every 15 minutes | Deletes expired user sessions and reset/verification tokens |
| Deployment health | Every 30 seconds | Checks enabled deployments on enabled connections, with at most four simultaneous probes |

Each job completes before its next run can begin. Cleanup uses one short
transaction and logs deletion counts. It never deletes users, router keys,
requests, usage, or other configuration. Transient job errors are logged and
retried on the next scheduled run. An unavailable or unmigrated database is a
startup error. SIGINT/SIGTERM cancels active work and waits for the jobs to stop.

## Configure an HTTP health probe

Set a deployment's `settings_json` through the management API:

```json
{
  "healthcheck": {
    "url": "http://127.0.0.1:9000/healthz"
  }
}
```

The connection must have `location_kind: "local"` or `adapter: "custom"`.
The URL must be absolute HTTP(S), with the same scheme, host, and explicit port
as the connection's base URL. User information, query strings, and fragments
are rejected. Use the deployment's actual URL, not a guessed vendor health path.

Checks are GET requests with a five-second deadline. Redirects are not followed;
only 2xx responses are healthy. Provider credentials are never resolved or added
to a generic probe. Response bodies and upstream error messages are not stored.
Latency measures time to response headers, not inference performance.

Inside Docker, `127.0.0.1` means the backend container. For a runtime on your host,
use an address reachable from that container in both the connection and probe
URL; for a networked runtime, use its container DNS name or network address.

Built-in cloud-provider probes are unsupported until their adapters implement
provider-specific checks. Missing configuration, unsupported providers, and
invalid probe settings produce `unknown` without sending a network request.

## Read health

`GET /v1/deployments/{id}/health` requires the admin bearer token. For example:

```json
{
  "deployment_id": 1,
  "probe_kind": "reachability",
  "status": "healthy",
  "checked_at": 1794400000000,
  "expires_at": 1794400060000,
  "latency_ms": 12,
  "error_code": null
}
```

Migration `00003_deployment_health.sql` adds one latest result per deployment.
Results expire after 60 seconds. The endpoint reports `unknown` when a result
is missing, stale, disabled, or belongs to a different configuration. Stale
results retain their measurement timestamps; changed/disabled results are hidden.
The worker discards in-flight results if the configuration changed or was deleted.
Shutdown cancellation does not overwrite health with a network failure.

Error codes include `not_checked`, `not_configured`, `unsupported_probe`,
`invalid_config`, `disabled`, `config_changed`, `stale`, `network_error`, `timeout`,
and `http_status_<code>`. A successful reachability check is not proof that a
model can complete inference. Routing does not consume these results yet.

This version assumes one worker per installation. It has no Redis, job queue,
distributed leases, async transcription, webhook delivery, or log-retention job.

## Checks

Run `make worker-check`, `make api-check`, and `make db-check` for the worker,
endpoint, migration, and generated-query checks. Worker tests use real temporary
SQLite databases with controlled HTTP transports for timeouts and concurrency.

After building the backend image, exercise the real binaries and HTTP transport:

```sh
docker compose -f docker/compose.yml build backend
docker run --rm --entrypoint python3 \
  -v "$PWD/docker/test_worker.py:/test_worker.py:ro" \
  voice-cloud-backend /test_worker.py
```

This uses a temporary database and a local HTTP server inside the test container;
it never contacts a paid provider or modifies the deployment's database volume.
