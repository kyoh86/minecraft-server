#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

mkdir -p \
  "${SCRIPT_DIR}/runtime/velocity" \
  "${SCRIPT_DIR}/runtime/lobby" \
  "${SCRIPT_DIR}/runtime/survival" \
  "${SCRIPT_DIR}/runtime/postgres"

cp --update=none "${SCRIPT_DIR}/templates/velocity.toml" "${SCRIPT_DIR}/runtime/velocity/velocity.toml"

SECRET_FILE="${SCRIPT_DIR}/runtime/velocity/forwarding.secret"
if [[ ! -f "${SECRET_FILE}" ]]; then
  if command -v openssl >/dev/null 2>&1; then
    openssl rand -hex 48 > "${SECRET_FILE}"
  else
    tr -dc 'a-f0-9' < /dev/urandom | head -c 96 > "${SECRET_FILE}"
  fi
  chmod 600 "${SECRET_FILE}"
fi

echo "Initialized: ${SCRIPT_DIR}/runtime"
