#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

mkdir -p \
  "${SCRIPT_DIR}/runtime/velocity" \
  "${SCRIPT_DIR}/runtime/lobby" \
  "${SCRIPT_DIR}/runtime/survival"

cp -n "${SCRIPT_DIR}/templates/velocity.toml" "${SCRIPT_DIR}/runtime/velocity/velocity.toml"
cp -n "${SCRIPT_DIR}/templates/forwarding.secret" "${SCRIPT_DIR}/runtime/velocity/forwarding.secret"

echo "Initialized: ${SCRIPT_DIR}/runtime"
