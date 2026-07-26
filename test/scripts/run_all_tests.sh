#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "${SCRIPT_DIR}/../integration"

echo "Running all integration tests..."
go test -v -count=1 -timeout 300s ./tests/...
