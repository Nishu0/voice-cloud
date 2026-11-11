# Docker deployment

The Compose stack has two services:

- `backend`: one image and container running the Go API and worker, despite their
  separate source directories. Startup applies Goose migrations and runs the
  provider seed before launching either process. If either process exits, the
  backend stops its sibling and exits so Compose can restart it.
- `web`: builds SvelteKit with adapter-node and runs the generated server using
  Bun. Both build and runtime stages use `oven/bun:1.3.13`.

The worker performs deployment health checks and expired auth-record cleanup;
see [the worker guide](worker.md). No extra worker service or database server is needed.

## Start

Run from the repository root with Docker Compose v2.22+:

```sh
export ADMIN_TOKEN="$(openssl rand -hex 32)"
docker compose -f docker/compose.yml up --build -d
docker compose -f docker/compose.yml logs -f backend
```

Web: `http://localhost:3000`. API: `http://localhost:8080`; public health check:
`http://localhost:8080/healthz`. All management endpoints require the admin bearer
header. The token is only passed to the backend, never the frontend.

Host ports bind to loopback. Use a reverse proxy with HTTPS for external access
and set `ORIGIN` to the frontend's public origin. Backend and frontend run as
unprivileged users. The frontend waits for the backend health check before starting.
The current frontend is still the SvelteKit starter page; this change deploys it
without adding application screens or browser authentication.

SQLite lives at `/data/voice-cloud.db` on the named `database` volume. Ordinary
container restarts, rebuilds, and `docker compose down` preserve it. `down -v`
deletes the volume and its data. Seeding inserts missing provider names without
replacing existing configuration. Seeded providers remain disabled.

For database maintenance, override the backend command, for example:

```sh
docker compose -f docker/compose.yml run --rm --no-deps backend \
  goose -env=none -dir /app/db/migrations sqlite3 /data/voice-cloud.db status
```

## Development

```sh
docker compose -p voice-cloud-dev \
  -f docker/compose.yml -f docker/dev.compose.yml up --build --watch
```

The frontend uses the Bun/Vite dev server with source syncing and hot reload.
Go code, migrations, and the seed script trigger backend image rebuilds. Manifest,
lockfile, and Vite configuration changes rebuild the frontend. Container dependencies
stay separate from host `node_modules`. The development project uses a separate
SQLite volume; stop the production stack first because both publish the same ports.

## Check backend supervision

After building the images:

```sh
docker compose -f docker/compose.yml build
docker run --rm --entrypoint python3 \
  -v "$PWD/docker/test_backend.py:/test_backend.py:ro" \
  voice-cloud-backend /test_backend.py
```

This checks migration/seeding, shutdown of both child processes, propagation of a
child failure, and token validation using a temporary database. Child executables
are substituted only inside this test; production runs the Go binaries.
