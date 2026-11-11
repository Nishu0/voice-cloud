<p align="center">
  <img src="apps/site/static/logo.svg" alt="Voice Cloud logo" width="220">
</p>

<h1 align="center">Voice Cloud</h1>

**An open router for voice.** Connect local and cloud voice models through one router, with policies for language, latency, and cost.

Voice Cloud is being built in Go for self-hosted deployments, with support for voice applications and agents.

> **Early development.** The features below describe the direction of the project. APIs and setup are still taking shape.

## What we're building

- **Local and cloud models** — connect models running on your hardware or hosted by a provider.
- **Policy-based routing** — choose models based on language, latency, and cost requirements.
- **Activity, logs, and analytics** — see which models handle requests and track usage, latency, and cost.
- **Agent support** — use the router as the voice layer for agents.
- **Self-hosting** — run the router on your own infrastructure.

## Repository layout

| Path | What it is |
| --- | --- |
| `apps/api` | Go management API |
| `apps/worker` | Go background worker: deployment health checks and cleanup |
| `apps/web` | The console (SvelteKit) |
| `apps/site` | The landing page (static SvelteKit) |
| `db` | SQLite migrations and queries |
| `docker` | Compose stack for the backend and console |
| `docs` | Guides for the API, worker, providers, users, and Docker |

## Run it

From the repository root, with Docker Compose v2.22+:

```sh
docker compose -f docker/compose.yml up --build -d
```

The console is at `http://localhost:3000` and the API at `http://localhost:8080`.
Open the console to create the first administrator. See [the Docker guide](docs/docker.md)
for development mode and [the API guide](docs/api.md) for running without Docker.

## Landing page

The marketing site lives in `apps/site` and builds to static files:

```sh
cd apps/site
bun install
bun run dev      # http://localhost:5174
bun run build    # output in apps/site/build
```

Its "Open app" and "Docs" links are set in `apps/site/src/lib/links.ts`, or with
`VITE_APP_URL` and `VITE_DOCS_URL` at build time. See [its README](apps/site/README.md).
