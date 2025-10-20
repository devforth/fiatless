#!/bin/sh

# Ensure PostgreSQL data directory has correct permissions
chown -R postgres:postgres /var/lib/postgresql

# Generate or load secrets persisted under the data volume
ENV_FILE=/var/secrets/fiatless.env
# Ensure secrets directory exists with restrictive permissions
SECRETS_DIR=$(dirname "$ENV_FILE")
if [ ! -d "$SECRETS_DIR" ]; then
    mkdir -p "$SECRETS_DIR"
    chown root:root "$SECRETS_DIR"
    chmod 700 "$SECRETS_DIR"
fi
if [ ! -f "$ENV_FILE" ]; then
    echo "Generating initial secrets in $ENV_FILE ..."
    DB_USER=fiatless
    DB_NAME=fiatless
    # generate strong random passwords and secrets
    DB_PASSWORD=$(head -c 24 /dev/urandom | base64 | tr -d '\n' | tr -d '=')
    ADMINFORTH_SECRET=$(head -c 24 /dev/urandom | base64 | tr -d '\n' | tr -d '=')
    IJSON_ENDPOINT=${IJSON_ENDPOINT:-http://localhost:8001}
    PORT=${PORT:-8080}
    # compose DSNs
    DATABASE_DSN="host=localhost user=$DB_USER password=$DB_PASSWORD dbname=$DB_NAME port=5432 sslmode=disable TimeZone=UTC"
    DATABASE_URL="postgres://$DB_USER:$DB_PASSWORD@localhost:5432/$DB_NAME?sslmode=disable"
    cat > "$ENV_FILE" <<EOF
PORT=$PORT
IJSON_ENDPOINT=$IJSON_ENDPOINT
DATABASE_DSN=$DATABASE_DSN
DATABASE_URL=$DATABASE_URL
ADMINFORTH_SECRET=$ADMINFORTH_SECRET
EOF
    chmod 600 "$ENV_FILE"
fi

# Export env for current process tree
. "$ENV_FILE"

# Run database initialization before starting supervisord
/init-db.sh

# Ensure PostgreSQL is stopped after initialization
if ps aux | grep postgres | grep -v grep > /dev/null; then
    echo "Ensuring PostgreSQL is stopped before supervisord starts it..."
    su postgres -c "pg_ctl -D /var/lib/postgresql/data stop -m fast"
    sleep 2
fi

# Start supervisord
exec supervisord -c /etc/supervisor/supervisord.conf
