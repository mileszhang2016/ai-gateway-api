#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DATA_DIR="${SCRIPT_DIR}/../integration/data"

echo "Cleaning runtime data..."
if [ -d "${DATA_DIR}" ]; then
    rm -f "${DATA_DIR}"/ai-gateway-api-*.exe
    rm -f "${DATA_DIR}"/test_ai_gateway_*.db
    rm -f "${DATA_DIR}"/test_ai_gateway_*.db-wal
    rm -f "${DATA_DIR}"/test_ai_gateway_*.db-shm
    rm -rf "${DATA_DIR}"/tmp*
    echo "Cleaned."
else
    echo "Data directory not found: ${DATA_DIR}"
fi
