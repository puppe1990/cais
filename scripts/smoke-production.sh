#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
TMP="$(mktemp -d)"
PORT_NUM=$((18000 + RANDOM % 8000))
trap 'kill "${PID:-}" 2>/dev/null || true; rm -rf "$TMP"' EXIT

cd "$ROOT"
go build -o "$TMP/server" ./cmd/server

export ENV=production
export APP_URL=https://example.com
export ADMIN_TOKEN=ci-smoke-secret
export PORT=":${PORT_NUM}"
export DB_PATH="$TMP/app.db"

"$TMP/server" &
PID=$!

READY=0
for _ in $(seq 1 60); do
  if curl -sf "http://127.0.0.1:${PORT_NUM}/health" >/dev/null 2>&1; then
    READY=1
    break
  fi
  sleep 0.2
done
if [ "$READY" -ne 1 ]; then
  echo "server did not become ready on port ${PORT_NUM}" >&2
  exit 1
fi

BODY=$(curl -sf "http://127.0.0.1:${PORT_NUM}/health")
echo "$BODY" | grep -q '"status":"ok"'

CODE=$(curl -s -o /dev/null -w "%{http_code}" "http://127.0.0.1:${PORT_NUM}/login")
if [ "$CODE" != "200" ]; then
  echo "GET /login status = $CODE, want 200" >&2
  exit 1
fi

echo "smoke production: ok"
