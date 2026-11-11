.DEFAULT_GOAL := help

DB_PATH ?= data/voice-cloud.db
MIGRATIONS_DIR := db/migrations
GOOSE := goose -env=none -dir "$(MIGRATIONS_DIR)" sqlite3 "$(DB_PATH)"

.PHONY: help db-generate db-new db-up db-down db-status db-version db-shell db-backup db-check

help:
	@printf '%s\n' \
	  'make db-generate                Generate Go code from SQL queries (sqlc)' \
	  'make db-new NAME=add_feature    Create a SQL migration (Goose)' \
	  'make db-up                      Apply pending migrations' \
	  'make db-down                    Roll back one migration (may delete data)' \
	  'make db-status                  Show migration status' \
	  'make db-version                 Show the current migration version' \
	  'make db-shell                   Open SQLite with foreign keys enabled' \
	  'make db-backup BACKUP=copy.db    Back up the database to a new file' \
	  'make db-check                   Test migrations, queries, and Go generation' \
	  'Override the database with DB_PATH=/path/to/database.db'

db-generate:
	sqlc generate

db-new:
	@test -n "$(NAME)" || { echo 'Usage: make db-new NAME=add_feature'; exit 1; }
	goose -env=none -dir "$(MIGRATIONS_DIR)" create "$(NAME)" sql

db-up:
	@mkdir -p "$$(dirname "$(DB_PATH)")"
	$(GOOSE) up

db-down:
	@test -f "$(DB_PATH)" || { echo 'Database does not exist; run make db-up first'; exit 1; }
	$(GOOSE) down

db-status:
	@mkdir -p "$$(dirname "$(DB_PATH)")"
	$(GOOSE) status

db-version:
	@test -f "$(DB_PATH)" || { echo 'Database does not exist; run make db-up first'; exit 1; }
	$(GOOSE) version

db-shell:
	@test -f "$(DB_PATH)" || { echo 'Database does not exist; run make db-up first'; exit 1; }
	sqlite3 -init /dev/null -cmd 'PRAGMA foreign_keys = ON;' "$(DB_PATH)"

db-backup:
	@test -f "$(DB_PATH)" || { echo 'Database does not exist; run make db-up first'; exit 1; }
	@test -n "$(BACKUP)" || { echo 'Usage: make db-backup BACKUP=copy.db'; exit 1; }
	python3 -c 'import sqlite3, sys; from pathlib import Path; source, target = sys.argv[1:]; Path(target).touch(exist_ok=False); src = sqlite3.connect(Path(source).resolve().as_uri() + "?mode=ro", uri=True); dst = sqlite3.connect(target); src.backup(dst); dst.close(); src.close()' "$(DB_PATH)" "$(BACKUP)"

db-check:
	goose -env=none -dir "$(MIGRATIONS_DIR)" validate
	python3 -B -m unittest discover -s db -p 'test_*.py' -v
