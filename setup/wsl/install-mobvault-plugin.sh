#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
PLUGIN_SRC_DIR="${REPO_ROOT}/plugins/mobvault/src"
RUNTIME_LOBBY_DIR="${SCRIPT_DIR}/runtime/lobby"
RUNTIME_SURVIVAL_DIR="${SCRIPT_DIR}/runtime/survival"
TARGET_LOBBY_JAR="${RUNTIME_LOBBY_DIR}/plugins/mobvault.jar"
TARGET_SURVIVAL_JAR="${RUNTIME_SURVIVAL_DIR}/plugins/mobvault.jar"

POSTGRES_JDBC_VERSION="42.7.7"
POSTGRES_JDBC_URL="https://repo1.maven.org/maven2/org/postgresql/postgresql/${POSTGRES_JDBC_VERSION}/postgresql-${POSTGRES_JDBC_VERSION}.jar"
JETBRAINS_ANNOTATIONS_VERSION="26.0.2"
JETBRAINS_ANNOTATIONS_URL="https://repo1.maven.org/maven2/org/jetbrains/annotations/${JETBRAINS_ANNOTATIONS_VERSION}/annotations-${JETBRAINS_ANNOTATIONS_VERSION}.jar"

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

mkdir -p "${TMP_DIR}/classes" "${TMP_DIR}/jdbc-classes" "$(dirname "${TARGET_LOBBY_JAR}")" "$(dirname "${TARGET_SURVIVAL_JAR}")"

JDBC_JAR="${TMP_DIR}/postgresql.jar"
ANNOTATIONS_JAR="${TMP_DIR}/jetbrains-annotations.jar"
curl -fsSL "${POSTGRES_JDBC_URL}" -o "${JDBC_JAR}"
curl -fsSL "${JETBRAINS_ANNOTATIONS_URL}" -o "${ANNOTATIONS_JAR}"

javac --release 21 \
  -cp "${JAVA_CP}:${JDBC_JAR}:${ANNOTATIONS_JAR}" \
  -d "${TMP_DIR}/classes" \
  "${PLUGIN_SRC_DIR}/dev/kyoh86/minecraft/mobvault/MobVaultPlugin.java"

(
  cd "${TMP_DIR}/jdbc-classes"
  jar xf "${JDBC_JAR}"
)

jar --create \
  --file "${TARGET_LOBBY_JAR}" \
  -C "${TMP_DIR}/classes" . \
  -C "${TMP_DIR}/jdbc-classes" . \
  -C "${PLUGIN_SRC_DIR}" plugin.yml \
  -C "${PLUGIN_SRC_DIR}" config.yml

cp "${TARGET_LOBBY_JAR}" "${TARGET_SURVIVAL_JAR}"

echo "installed: ${TARGET_LOBBY_JAR}"
echo "installed: ${TARGET_SURVIVAL_JAR}"

docker compose -f "${SCRIPT_DIR}/docker-compose.yml" up -d --force-recreate lobby survival
echo "restarted: lobby/survival"
