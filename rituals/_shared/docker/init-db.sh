#!/bin/sh
# Database initialization script
# This script is run by docker-compose on first startup

set -e

echo "🔧 Initializing database..."

# Wait for database to be ready
until pg_isready -U "${POSTGRES_USER}" -d "${POSTGRES_DB}"; do
  echo "⏳ Waiting for database to be ready..."
  sleep 2
done

echo "✅ Database is ready!"

# Run migrations if they exist
if [ -d "/docker-entrypoint-initdb.d" ]; then
  echo "📦 Running migrations from /docker-entrypoint-initdb.d..."
  for f in /docker-entrypoint-initdb.d/*.sql; do
    if [ -f "$f" ]; then
      echo "  Executing: $f"
      psql -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" -f "$f"
    fi
  done
fi

echo "🎉 Database initialization complete!"
