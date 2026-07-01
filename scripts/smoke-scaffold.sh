#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

cd "$ROOT"
go build -o "$TMP/cais" ./cmd/cais

export CAIS_REPLACE="$ROOT"
export CAIS_SKIP_TIDY=1

APP="$TMP/smokeapp"
"$TMP/cais" new smokeapp "$APP"
cd "$APP"
go mod tidy

go test ./... -count=1
go build -o "$TMP/server" ./cmd/server

echo "smoke scaffold: ok"
