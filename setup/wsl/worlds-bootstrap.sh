#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=./worlds-common.sh
source "${SCRIPT_DIR}/worlds-common.sh"

if [[ ! -x "${SYNC_SCRIPT}" ]]; then
  echo "missing sync script: ${SYNC_SCRIPT}" >&2
  exit 1
fi

"${SYNC_SCRIPT}"
send_console "reload"

while IFS= read -r cfg; do
  ensure_world "${cfg}"
  apply_world_init_function "${cfg}"
done < <(list_world_configs)

echo "bootstrapped worlds from ${WORLD_ROOT}"
