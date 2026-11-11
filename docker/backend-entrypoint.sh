#!/usr/bin/env bash
set -euo pipefail

if (( $# )); then
    exec "$@"
fi

if [[ -n "${ADMIN_TOKEN:-}" ]] && (( ${#ADMIN_TOKEN} < 32 )); then
    echo 'ADMIN_TOKEN must contain at least 32 characters' >&2
    exit 1
fi

mkdir -p "$(dirname "$DB_PATH")"
goose -env=none -dir /app/db/migrations sqlite3 "$DB_PATH" up
python3 /app/scripts/seed.py --db "$DB_PATH"

api &
api_pid=$!
worker &
worker_pid=$!

shutdown() {
    trap '' TERM INT
    kill -TERM "$api_pid" "$worker_pid" 2>/dev/null || true
    wait "$api_pid" "$worker_pid" 2>/dev/null || true
}
trap 'shutdown; exit 0' TERM INT

status=0
wait -n "$api_pid" "$worker_pid" || status=$?
shutdown
# Even a clean child exit is unexpected while the backend is serving.
if (( status == 0 )); then status=1; fi
exit "$status"
