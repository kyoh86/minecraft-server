#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
PLUGIN_SRC_DIR="${REPO_ROOT}/plugins/gatebridge/src"
RUNTIME_LOBBY_DIR="${SCRIPT_DIR}/runtime/lobby"
TARGET_DIR="${RUNTIME_LOBBY_DIR}/plugins"
TARGET_JAR="${TARGET_DIR}/gatebridge.jar"

if [[ ! -d "${PLUGIN_SRC_DIR}" ]]; then
  echo "missing source dir: ${PLUGIN_SRC_DIR}" >&2
  exit 1
fi

LIBRARIES_DIR="${RUNTIME_LOBBY_DIR}/libraries"
if [[ ! -d "${LIBRARIES_DIR}" ]]; then
  echo "missing libraries dir under ${RUNTIME_LOBBY_DIR}" >&2
  echo "run: make up" >&2
  exit 1
fi

JAVA_CP="$(find "${LIBRARIES_DIR}" -type f -name '*.jar' | paste -sd ':' -)"
if [[ -z "${JAVA_CP}" ]]; then
  echo "no jars found under ${LIBRARIES_DIR}" >&2
  exit 1
fi

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "${TMP_DIR}"' EXIT

mkdir -p "${TMP_DIR}/classes" "${TARGET_DIR}"

javac --release 21 \
  -cp "${JAVA_CP}" \
  -d "${TMP_DIR}/classes" \
  "${PLUGIN_SRC_DIR}/dev/kyoh86/minecraft/gatebridge/GateBridgePlugin.java"

jar --create \
  --file "${TARGET_JAR}" \
  -C "${TMP_DIR}/classes" . \
  -C "${PLUGIN_SRC_DIR}" plugin.yml \
  -C "${PLUGIN_SRC_DIR}" config.yml

# Cleanup legacy artifacts.
rm -f "${TARGET_DIR}/123783.jar" "${TARGET_DIR}/.123783-version.json"
rm -f "${TARGET_DIR}/lobby-gate-switcher.jar"
rm -rf "${TARGET_DIR}/LobbyGateSwitcher"

echo "installed: ${TARGET_JAR}"
