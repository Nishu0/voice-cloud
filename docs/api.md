# Management API

Run from the repository root:

```sh
make api-run
```

Start the web app with `bun run dev` in `apps/web`, then open it to create the first
administrator with a name, email, and password. The setup route closes once any
account exists. Administrators can add members from Account settings. Public
signup is intentionally limited to this initial setup.

The API defaults to `127.0.0.1:8080`. Override `API_ADDR` and `DB_PATH` as needed.
The web server uses private `API_URL` (default `http://127.0.0.1:8080`); Compose
sets it to `http://backend:8080`. No API credential is built into frontend assets.

Authenticate API calls with a user session bearer token returned by login/signup:

```sh
curl http://127.0.0.1:8080/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"you@example.com","password":"your-password"}'
```

The console stores this token in an HttpOnly, SameSite=Lax cookie; Secure is set
on HTTPS. The Go API accepts bearer tokens, not browser cookies. SvelteKit checks
same-origin form submissions and forwards sessions on the server. Sessions last
seven days and are revoked on logout. Use HTTPS and configure `ORIGIN` at your
reverse proxy before external access.

`ADMIN_TOKEN` is optional legacy machine access; if set it must have at least 32
characters. It is not needed for setup, login, or the frontend. Router keys cannot
access the management API. Members can read shared data and update their own
name; configuration and member creation require an administrator.

## Endpoints

| Path | Methods | Behavior |
| --- | --- | --- |
| `/healthz` | GET | Public database availability check |
| `/v1/auth/status` | GET | Public `needs_setup` flag |
| `/v1/auth/signup` | POST | First account only; atomically creates administrator and session |
| `/v1/auth/login` | POST | Email/password sign-in; returns token, expiry, safe user fields |
| `/v1/auth/me` | GET, PUT | Current user; PUT accepts a name |
| `/v1/auth/logout` | POST | Revoke current session |
| `/v1/users` | POST | Administrator creates a member with name, email, password |
| `/v1/connections` | GET, POST | List/create provider connections |
| `/v1/connections/{id}` | GET, PUT, DELETE | Read/replace/delete a connection |
| `/v1/models` | GET, POST | List/create models |
| `/v1/models/{id}` | GET, PUT, DELETE | Read/replace/delete a model |
| `/v1/deployments` | GET, POST | List/create deployments |
| `/v1/deployments/{id}` | GET, PUT, DELETE | Read/replace/delete a deployment |
| `/v1/deployments/{id}/health` | GET | Read worker health; stale/changed/disabled results are unknown |
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

This is the configuration and observability API. The worker records deployment
health; see [the worker guide](worker.md) for probe configuration. Provider
inference, streaming, and activity recording are not implemented yet. There are no public
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
  -H "Authorization: Bearer $SESSION_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"name":"Local","adapter":"whisper","location_kind":"local","base_url":"http://localhost:9000"}'

curl http://127.0.0.1:8080/v1/models \
  -H "Authorization: Bearer $SESSION_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"name":"large-v3","publisher":"openai","task":"stt"}'

curl http://127.0.0.1:8080/v1/deployments \
  -H "Authorization: Bearer $SESSION_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"name":"Local Whisper","model_id":1,"connection_id":1,"upstream_model_id":"large-v3","capabilities_json":{"streaming":true}}'

curl http://127.0.0.1:8080/v1/policies \
  -H "Authorization: Bearer $SESSION_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"name":"Speech","task":"stt","objective":"latency"}'

curl -X PUT http://127.0.0.1:8080/v1/policies/1/targets/1 \
  -H "Authorization: Bearer $SESSION_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"priority":0}'

curl http://127.0.0.1:8080/v1/api-keys \
  -H "Authorization: Bearer $SESSION_TOKEN" \
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
