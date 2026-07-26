#!/usr/bin/env bash
set -euo pipefail

MODULE="${1:-}"
if [ -z "${MODULE}" ]; then
    echo "Usage: $0 <module>"
    echo "Available modules:"
    ls "$(dirname "${BASH_SOURCE[0]}")/../integration/tests"
    exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "${SCRIPT_DIR}/../integration"

echo "Running integration tests for module: ${MODULE}"
go test -v -count=1 -timeout 120s "./tests/${MODULE}/..."
