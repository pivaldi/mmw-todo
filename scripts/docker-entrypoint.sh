#!/bin/sh
set -e

echo "Waiting for PostgreSQL to be ready..."
until pg_isready -h postgres -p 5432 -U postgres; do
  echo "PostgreSQL is unavailable - sleeping"
  sleep 1
done

echo "PostgreSQL is up - checking migrations"

# Check if migrate tool is available (in production builds)
if command -v migrate >/dev/null 2>&1; then
  echo "Running database migrations..."
  migrate -path /app/migrations -database "${DATABASE_URL}" up
  echo "Migrations complete"
else
  echo "Warning: migrate tool not found, skipping migrations"
fi

echo "Starting application..."
exec "$@"
