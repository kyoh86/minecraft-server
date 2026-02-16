#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

mkdir -p \
  "${SCRIPT_DIR}/runtime/lobby"

echo "Initialized: ${SCRIPT_DIR}/runtime"
