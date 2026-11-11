# Voice Cloud

<img src="assets/logo.png" alt="Voice Cloud logo" width="160">

**An open router for voice.** Connect local and cloud voice models through one router, with policies for language, latency, and cost.

Voice Cloud is being built in Go for self-hosted deployments, with support for voice applications and agents.

> **Early development.** The features below describe the direction of the project. APIs and setup are still taking shape.

## What we're building

- **Local and cloud models** — connect models running on your hardware or hosted by a provider.
- **Policy-based routing** — choose models based on language, latency, and cost requirements.
- **Activity, logs, and analytics** — see which models handle requests and track usage, latency, and cost.
- **Agent support** — use the router as the voice layer for agents.
- **Self-hosting** — run the router on your own infrastructure.

## Database development

Install the [Goose CLI](https://pressly.github.io/goose/installation/),
[sqlc](https://docs.sqlc.dev/en/stable/overview/install.html), SQLite CLI, and Python 3.
Run `make` to list the database commands. The default database is
`data/voice-cloud.db`; override it with `DB_PATH=/path/to/database.db`.

```sh
make db-up
make db-status
make db-new NAME=add_feature
make db-check
make db-generate
make db-backup BACKUP=/tmp/voice-cloud-backup.db
```

Fill in both the `-- +goose Up` and `-- +goose Down` sections of new migrations
in `db/migrations/`, and update `db/schema.sql` to reflect the resulting schema.
Never edit a migration that has already been applied; add another migration.
`make db-check` checks the migrated schema against `db/schema.sql` and exercises
rollback and reapplication using a temporary database.

Add named queries to `db/queries.sql` before running `make db-generate`;
sqlc cannot generate code from the currently empty query file.
`make db-down` rolls back one migration and can delete data, including all router
tables when rolling back the initial migration. Back up first when needed.
Runtime connections must enable foreign keys and configure WAL separately.
