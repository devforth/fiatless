#!/bin/sh
set -e

# Load persisted env if exists
ENV_FILE=/var/secrets/fiatless.env
if [ -f "$ENV_FILE" ]; then
  . "$ENV_FILE"
fi

# Check if PostgreSQL is initialized
if [ ! -s /var/lib/postgresql/data/PG_VERSION ]; then
  echo "Initializing PostgreSQL database..."
  su postgres -c "initdb -D /var/lib/postgresql/data"
fi

# Start PostgreSQL
echo "Starting PostgreSQL temporarily for initialization..."
su postgres -c "pg_ctl -D /var/lib/postgresql/data -o '-c listen_addresses=*' start"

# Wait for PostgreSQL to start
echo "Waiting for PostgreSQL to start..."
sleep 3

# Check if database already exists
if ! su postgres -c "psql -lqt" | grep -q " ${DB_NAME:-fiatless} "; then
  echo "Creating fiatless user and database..."
  # Create user and database
  su postgres -c "psql -c \"CREATE USER ${DB_USER:-fiatless} WITH PASSWORD '${DB_PASSWORD:-fiatless}';\""
  su postgres -c "psql -c \"CREATE DATABASE ${DB_NAME:-fiatless} OWNER ${DB_USER:-fiatless};\""
  su postgres -c "psql -c \"GRANT ALL PRIVILEGES ON DATABASE ${DB_NAME:-fiatless} TO ${DB_USER:-fiatless};\""
  echo "Database initialized successfully"
else
  echo "Database already exists, skipping initialization"
fi

echo "Running migrations..."
migrate -path=/app/migrations -database "${DATABASE_URL:-postgres://fiatless:fiatless@localhost:5432/fiatless?sslmode=disable}" up

# Stop PostgreSQL (supervisord will start it)
echo "Stopping PostgreSQL after initialization..."
su postgres -c "pg_ctl -D /var/lib/postgresql/data stop -m fast"
sleep 2