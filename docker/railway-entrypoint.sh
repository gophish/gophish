#!/bin/bash
#
# Railway entrypoint for GoPhish.
#
# Railway mounts a persistent Volume at /data owned by root, while GoPhish
# runs as the non-root "app" user. Without fixing ownership, SQLite cannot
# create the database file and fails with:
#   "unable to open database file: no such file or directory"
#
# This script runs as root only long enough to guarantee the DB directory
# exists and is writable by "app", then drops privileges via gosu and hands
# off to the upstream run.sh (which already honours DB_FILE_PATH).
#
# It never deletes, formats or overwrites an existing database.
set -e

# Single source of truth for the DB location: DB_FILE_PATH env.
DB_FILE_PATH="${DB_FILE_PATH:-/data/gophish.db}"
DB_DIR="$(dirname "${DB_FILE_PATH}")"

# Ensure the persistent directory exists and is owned by app.
# mkdir -p and chown are idempotent and do not touch existing data.
mkdir -p "${DB_DIR}"
chown -R app:app "${DB_DIR}"

# Drop to the non-root app user and start GoPhish. exec preserves signal
# handling and exit codes for graceful shutdown.
exec gosu app ./docker/run.sh
