#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
COMPOSE_FILE="${SCRIPT_DIR}/docker-compose.yml"
RUNTIME_VELOCITY_DIR="${SCRIPT_DIR}/runtime/velocity"
PLUGINS_DIR="${RUNTIME_VELOCITY_DIR}/plugins"
TARGET_JAR="${PLUGINS_DIR}/rememberme.jar"

API_URL="https://hangar.papermc.io/api/v1/projects/ichiru/Rememberme/versions"

if ! command -v jq >/dev/null 2>&1; then
  echo "jq is required" >&2
  exit 1
fi

mkdir -p "${PLUGINS_DIR}"

version_json="$(curl -fsSL "${API_URL}")"
download_url="$(echo "${version_json}" | jq -r '.result | sort_by(.createdAt) | last | .downloads.VELOCITY.downloadUrl')"
sha256_hash="$(echo "${version_json}" | jq -r '.result | sort_by(.createdAt) | last | .downloads.VELOCITY.fileInfo.sha256Hash')"

if [[ -z "${download_url}" || "${download_url}" == "null" ]]; then
  echo "failed to resolve download URL from Hangar API" >&2
  exit 1
fi
if [[ -z "${sha256_hash}" || "${sha256_hash}" == "null" ]]; then
  echo "failed to resolve sha256 hash from Hangar API" >&2
  exit 1
fi

tmp_file="$(mktemp)"
trap 'rm -f "${tmp_file}"' EXIT

curl -fsSL "${download_url}" -o "${tmp_file}"
actual_hash="$(sha256sum "${tmp_file}" | awk '{print $1}')"
if [[ "${actual_hash}" != "${sha256_hash}" ]]; then
  echo "sha256 mismatch: expected=${sha256_hash} actual=${actual_hash}" >&2
  exit 1
fi

install -m 0644 "${tmp_file}" "${TARGET_JAR}"
echo "installed: ${TARGET_JAR}"

docker compose -f "${COMPOSE_FILE}" restart velocity
echo "restarted: velocity"
