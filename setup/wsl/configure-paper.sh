#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEMPLATE_FILE="${SCRIPT_DIR}/templates/paper-global.velocity.yml"
SECRET_FILE="${SCRIPT_DIR}/runtime/velocity/forwarding.secret"
MODE="${1:-all}"

if [[ "${MODE}" != "all" && "${MODE}" != "secret-only" ]]; then
  echo "Usage: $0 [all|secret-only]" >&2
  exit 1
fi

if [[ ! -f "${TEMPLATE_FILE}" ]]; then
  echo "missing template: ${TEMPLATE_FILE}" >&2
  exit 1
fi

if [[ ! -f "${SECRET_FILE}" ]]; then
  echo "missing secret: ${SECRET_FILE}" >&2
  echo "run: make init" >&2
  exit 1
fi

DESIRED_ENABLED="$(awk '/^    enabled:/{print $2; exit}' "${TEMPLATE_FILE}")"
DESIRED_ONLINE_MODE="$(awk '/^    online-mode:/{print $2; exit}' "${TEMPLATE_FILE}")"
SECRET="$(tr -d '\n' < "${SECRET_FILE}")"

if [[ -z "${SECRET}" ]]; then
  echo "empty secret file: ${SECRET_FILE}" >&2
  exit 1
fi

apply_file() {
  local file="$1"
  local tmp

  if [[ ! -f "${file}" ]]; then
    echo "missing file: ${file}" >&2
    return 1
  fi

  tmp="$(mktemp)"
  awk \
    -v mode="${MODE}" \
    -v desired_enabled="${DESIRED_ENABLED}" \
    -v desired_online_mode="${DESIRED_ONLINE_MODE}" \
    -v desired_secret="${SECRET}" '
    BEGIN {
      in_proxies = 0
      in_velocity = 0
      found_velocity = 0
      updated_enabled = 0
      updated_online_mode = 0
      updated_secret = 0
    }
    {
      if ($0 ~ /^proxies:$/) {
        in_proxies = 1
        in_velocity = 0
        print
        next
      }

      if (in_proxies && $0 ~ /^[^ ]/) {
        in_proxies = 0
        in_velocity = 0
      }

      if (in_proxies && $0 ~ /^  velocity:$/) {
        in_velocity = 1
        found_velocity = 1
        print
        next
      }

      if (in_velocity && $0 ~ /^  [^ ]/) {
        in_velocity = 0
      }

      if (in_velocity && $0 ~ /^    enabled:/ && mode == "all") {
        $0 = "    enabled: " desired_enabled
        updated_enabled = 1
      } else if (in_velocity && $0 ~ /^    online-mode:/ && mode == "all") {
        $0 = "    online-mode: " desired_online_mode
        updated_online_mode = 1
      } else if (in_velocity && $0 ~ /^    secret:/) {
        $0 = "    secret: " desired_secret
        updated_secret = 1
      }

      print
    }
    END {
      if (!found_velocity) {
        exit 2
      }
      if (!updated_secret) {
        exit 3
      }
      if (mode == "all" && (!updated_enabled || !updated_online_mode)) {
        exit 4
      }
    }
  ' "${file}" > "${tmp}" || {
    local code=$?
    rm -f "${tmp}"
    if [[ "${code}" -eq 2 ]]; then
      echo "velocity block not found: ${file}" >&2
    elif [[ "${code}" -eq 3 ]]; then
      echo "secret key not found in velocity block: ${file}" >&2
    else
      echo "failed to configure: ${file}" >&2
    fi
    return 1
  }

  mv "${tmp}" "${file}"
}

apply_file "${SCRIPT_DIR}/runtime/lobby/config/paper-global.yml"
apply_file "${SCRIPT_DIR}/runtime/survival/config/paper-global.yml"

echo "configured paper-global.yml mode=${MODE}"
