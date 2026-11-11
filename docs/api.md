# Management API

Run from the repository root:

```sh
export ADMIN_TOKEN="$(openssl rand -hex 32)"
make api-run
```

`make api-run` applies migrations and starts the server at `127.0.0.1:8080`.
Override the address with `API_ADDR=127.0.0.1:9000` and the database with
`DB_PATH=/path/to/router.db`. Running the Go binary directly requires an already
migrated database. It refuses to create an empty database on startup.

All `/v1/` endpoints require `Authorization: Bearer $ADMIN_TOKEN`. Keep this
administration token separate from router keys. Configure HTTPS at your reverse
proxy before exposing the API outside a trusted local environment.

## Endpoints

| Path | Methods | Behavior |
| --- | --- | --- |
| `/healthz` | GET | Public database availability check |
| `/v1/connections` | GET, POST | List/create provider connections |
| `/v1/connections/{id}` | GET, PUT, DELETE | Read/replace/delete a connection |
| `/v1/models` | GET, POST | List/create models |
| `/v1/models/{id}` | GET, PUT, DELETE | Read/replace/delete a model |
| `/v1/deployments` | GET, POST | List/create deployments |
| `/v1/deployments/{id}` | GET, PUT, DELETE | Read/replace/delete a deployment |
| `/v1/deployments/{id}/languages` | GET | List supported language tags |
| `/v1/deployments/{id}/languages/{language}` | PUT, DELETE | Add/remove a language tag |
| `/v1/policies` | GET, POST | List/create routing policies |
| `/v1/policies/{id}` | GET, PUT, DELETE | Read/replace/delete a policy |
| `/v1/policies/{id}/targets` | GET | List target deployments by priority |
| `/v1/policies/{id}/targets/{deployment}` | PUT, DELETE | Set/remove a target; PUT accepts `{"priority":0}` |
| `/v1/api-keys` | GET, POST | List/mint router keys |
| `/v1/api-keys/{id}` | GET, DELETE | Read/revoke a key; revocation preserves history |
| `/v1/requests` | GET | Paginated request history |
| `/v1/requests/{request}` | GET | Read one request |
| `/v1/requests/{request}/attempts` | GET | List provider attempts in order |
| `/v1/requests/{request}/usage` | GET | List usage across all attempts |

This is the configuration and observability API. Provider inference, streaming,
and worker-written activity records are not implemented yet. There are no public
endpoints for fabricating or deleting activity records.

## Request and response conventions

POST and PUT bodies use `Content-Type: application/json`. The maximum body is
1 MiB; unknown fields and trailing JSON are rejected. Language PUT has no body.
Create returns 201, reads and replacements return 200, and delete/revoke returns
204. PUT replaces editable fields; it does not merge them.

Lists are JSON arrays, including `[]` when empty. Nullable values use JSON null,
boolean fields use true/false, and fields ending in `_json` contain JSON objects,
not serialized strings. Timestamps are Unix milliseconds. Request history accepts
`?limit=50&offset=0`, with a maximum limit of 100.

Errors have the shape `{"error":"message"}`. Invalid bodies/IDs return 400,
missing resources 404, database constraint conflicts 409, unsupported content
types 415, and oversized bodies 413. References prevent deletion of configuration
that is still in use. Disable a deployment/connection or revoke a key instead.

Model and policy tasks cannot be changed. A deployment's model, connection, and
upstream model ID cannot be changed; create another deployment for a different
route. This preserves policy compatibility and historical route identity.
Policy targets must match the policy task. Capabilities, pricing, settings, and
constraints currently accept arbitrary JSON objects; adapters will validate their
provider-specific contents when inference is implemented.

## Configure a route

```sh
curl http://127.0.0.1:8080/v1/connections \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"name":"Local","adapter":"whisper","location_kind":"local","base_url":"http://localhost:9000"}'

curl http://127.0.0.1:8080/v1/models \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"name":"large-v3","publisher":"openai","task":"stt"}'

curl http://127.0.0.1:8080/v1/deployments \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"name":"Local Whisper","model_id":1,"connection_id":1,"upstream_model_id":"large-v3","capabilities_json":{"streaming":true}}'

curl http://127.0.0.1:8080/v1/policies \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"name":"Speech","task":"stt","objective":"latency"}'

curl -X PUT http://127.0.0.1:8080/v1/policies/1/targets/1 \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"priority":0}'

curl http://127.0.0.1:8080/v1/api-keys \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"name":"My app","default_policy_id":1}'
```

The example IDs assume a fresh database; use the IDs returned by each response.
Connections/deployments default to enabled. Policies default to balanced routing,
fallback enabled, and two attempts. Supply `enabled:false` or
`fallback_enabled:false` to disable them. Missing JSON settings default to `{}`.
Provider secrets stay outside the database; `credential_ref` is an optional
external secret name, such as an environment variable name. It is not resolved yet.

Router key creation accepts a name, optional default policy ID, and optional
future `expires_at` timestamp. It returns `{"key":"vc_...","api_key":{...}}`.
Save the key then: it is shown once, stored as a SHA-256 hash, and never returned
by list/read endpoints. Router keys cannot call this administration API.

## Checks

```sh
make api-check
make db-check
```

API tests use the real Go SQLite driver and Goose migrations on temporary files,
without opening network listeners or contacting model providers.
