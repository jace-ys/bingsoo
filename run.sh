#!/bin/sh

exec bingsoo \
  --port "$PORT" \
  --concurrency "$CONCURRENCY" \
  --postgres-host "$POSTGRES_HOST" \
  --postgres-user "$POSTGRES_USER" \
  --postgres-password "$POSTGRES_PASSWORD" \
  --postgres-db "$POSTGRES_DB" \
  --slack-access-token "$SLACK_ACCESS_TOKEN" \
  --slack-signing-secret "$SLACK_SIGNING_SECRET"
