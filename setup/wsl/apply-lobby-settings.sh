#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
COMPOSE_FILE="${SCRIPT_DIR}/docker-compose.yml"
SETTINGS_FILE="${SCRIPT_DIR}/lobby-settings.mcfunction"

if [[ ! -f "${SETTINGS_FILE}" ]]; then
  echo "missing settings file: ${SETTINGS_FILE}" >&2
  exit 1
fi

# Apply each command idempotently from the settings file.
while IFS= read -r line || [[ -n "${line}" ]]; do
  line="${line#"${line%%[![:space:]]*}"}"
  [[ -z "${line}" ]] && continue
  [[ "${line}" =~ ^# ]] && continue
  docker compose -f "${COMPOSE_FILE}" exec -T --user 1000 lobby mc-send-to-console "${line}"
done < "${SETTINGS_FILE}"

echo "applied lobby settings from ${SETTINGS_FILE}"
