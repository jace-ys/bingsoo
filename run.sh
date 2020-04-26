#!/bin/sh

exec bingsoo \
  --port "$PORT" \
  --concurrency "$CONCURRENCY" \
  --signing-secret "$SIGNING_SECRET" \
  --access-token "$ACCESS_TOKEN" \
  --postgres-host "$POSTGRES_HOST" \
  --postgres-user "$POSTGRES_USER" \
  --postgres-password "$POSTGRES_PASSWORD" \
  --postgres-db "$POSTGRES_DB"
