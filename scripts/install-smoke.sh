#!/usr/bin/env sh
# CI smoke: install from a local fake release (run from repo root).
set -eu

ROOT=$(cd "$(dirname "$0")/.." && pwd)
cd "$ROOT"

go test ./internal/update/... -count=1

echo "install-smoke: update package tests passed"
