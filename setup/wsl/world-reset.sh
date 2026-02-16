#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=./worlds-common.sh
source "${SCRIPT_DIR}/worlds-common.sh"

TARGET_WORLD="${WORLD:-resource}"
CFG="${WORLD_ROOT}/${TARGET_WORLD}/world.env.yml"

if [[ ! -f "${CFG}" ]]; then
  echo "missing world config: ${CFG}" >&2
  exit 1
fi

if [[ "$(yaml_get "${CFG}" resettable)" != "true" ]]; then
  echo "world '${TARGET_WORLD}' is not resettable" >&2
  exit 1
fi

if world_registered_in_multiverse "${TARGET_WORLD}"; then
  send_console "mv unload ${TARGET_WORLD} --remove-players"
  send_console "mv remove ${TARGET_WORLD}"
fi

rm -rf \
  "${RUNTIME_ROOT}/${TARGET_WORLD}" \
  "${RUNTIME_ROOT}/${TARGET_WORLD}_nether" \
  "${RUNTIME_ROOT}/${TARGET_WORLD}_the_end"

ensure_world "${CFG}" true
send_console "reload"
apply_world_init_function "${CFG}"

echo "reset world '${TARGET_WORLD}'"
