#!/bin/sh

ENV_FILE=/var/secrets/fiatless.env
if [ -f "$ENV_FILE" ]; then
  . "$ENV_FILE"
fi

cd /code/

export DATABASE_URL="${DATABASE_URL:-postgres://fiatless:fiatless@localhost:5432/fiatless?sslmode=disable}"
export NODE_ENV="${NODE_ENV:-production}"
export ADMINFORTH_SECRET="${ADMINFORTH_SECRET:-123}"

npm run prod