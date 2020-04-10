#!/bin/sh

exec bingsoo \
  --port "$PORT" \
  --slack-access-token "$SLACK_ACCESS_TOKEN" \
  --slack-signing-secret "$SLACK_SIGNING_SECRET"
