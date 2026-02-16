#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
COMPOSE_FILE="${SCRIPT_DIR}/docker-compose.yml"
SYNC_SCRIPT="${SCRIPT_DIR}/sync-world-datapack.sh"
FUNCTION_NAME="mcserver:world_settings"

if [[ ! -x "${SYNC_SCRIPT}" ]]; then
  echo "missing sync script: ${SYNC_SCRIPT}" >&2
  exit 1
fi

"${SYNC_SCRIPT}"
docker compose -f "${COMPOSE_FILE}" exec -T --user 1000 world mc-send-to-console "reload" </dev/null
docker compose -f "${COMPOSE_FILE}" exec -T --user 1000 world mc-send-to-console "function ${FUNCTION_NAME}" </dev/null

echo "applied world settings function ${FUNCTION_NAME}"
