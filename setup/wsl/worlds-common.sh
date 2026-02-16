#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
COMPOSE_FILE="${SCRIPT_DIR}/docker-compose.yml"
WORLD_ROOT="${SCRIPT_DIR}/worlds"
RUNTIME_ROOT="${SCRIPT_DIR}/runtime/world"
MULTIVERSE_WORLDS_FILE="${RUNTIME_ROOT}/plugins/Multiverse-Core/worlds.yml"
SYNC_SCRIPT="${SCRIPT_DIR}/sync-world-datapack.sh"

yaml_get() {
  local file="$1"
  local key="$2"
  sed -nE "s/^${key}:[[:space:]]*(.*)$/\\1/p" "$file" | head -n1 | sed -E 's/^"(.*)"$/\1/'
}

list_world_configs() {
  find "${WORLD_ROOT}" -mindepth 2 -maxdepth 2 -type f -name 'world.env.yml' | sort
}

send_console() {
  local command="$1"
  docker compose -f "${COMPOSE_FILE}" exec -T --user 1000 world mc-send-to-console "${command}" </dev/null
}

world_exists_on_disk() {
  local world_name="$1"
  [[ -d "${RUNTIME_ROOT}/${world_name}" ]]
}

world_registered_in_multiverse() {
  local world_name="$1"
  [[ -f "${MULTIVERSE_WORLDS_FILE}" ]] || return 1
  grep -qE "^${world_name}:" "${MULTIVERSE_WORLDS_FILE}"
}

ensure_world() {
  local cfg="$1"
  local force_create="${2:-false}"
  local name environment seed world_type create_cmd

  name="$(yaml_get "${cfg}" name)"
  environment="$(yaml_get "${cfg}" environment)"
  seed="$(yaml_get "${cfg}" seed)"
  world_type="$(yaml_get "${cfg}" world_type)"

  if [[ -z "${name}" || -z "${environment}" ]]; then
    echo "invalid world config: ${cfg}" >&2
    exit 1
  fi

  if [[ "${force_create}" != "true" ]] && world_registered_in_multiverse "${name}"; then
    return
  fi

  if [[ "${force_create}" != "true" ]] && world_exists_on_disk "${name}"; then
    send_console "mv import ${name} ${environment}"
    return
  fi

  create_cmd="mv create ${name} ${environment}"
  if [[ -n "${seed}" ]]; then
    create_cmd+=" -s ${seed}"
  fi
  if [[ -n "${world_type}" ]]; then
    create_cmd+=" -t ${world_type^^}"
  fi

  send_console "${create_cmd}"
}

apply_world_init_function() {
  local cfg="$1"
  local name function_name

  name="$(yaml_get "${cfg}" name)"
  function_name="$(yaml_get "${cfg}" function)"
  if [[ -z "${function_name}" ]]; then
    function_name="mcserver:worlds/${name}/init"
  fi

  send_console "function ${function_name}"
}
