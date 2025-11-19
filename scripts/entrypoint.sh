#!/usr/bin/env sh
set -euo pipefail

echo "Running goose migrations..."
goose -dir /app/migrations postgres "${DATABASE_URL}" up

echo "Starting QA API service..."
exec "$@"


