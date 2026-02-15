#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SRC_DIR="${SCRIPT_DIR}/datapacks/lobby-base"
DST_ROOT="${SCRIPT_DIR}/runtime/lobby/world/datapacks"
DST_DIR="${DST_ROOT}/lobby-base"

if [[ ! -d "${SRC_DIR}" ]]; then
  echo "missing datapack source: ${SRC_DIR}" >&2
  exit 1
fi

mkdir -p "${DST_ROOT}"
rm -rf "${DST_DIR}"
cp -a "${SRC_DIR}" "${DST_DIR}"

echo "synced datapack to ${DST_DIR}"
