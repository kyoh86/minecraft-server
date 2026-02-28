#!/bin/sh
set -eu

PING_URL="$(tr -d '\r\n' < /run/healthchecks_heartbeat_url.txt)"
if [ -z "$PING_URL" ] || [ "$PING_URL" = "REPLACE_WITH_HEALTHCHECKS_PING_URL" ]; then
  echo "healthchecks ping url is missing. set secrets/healthchecks_heartbeat_url.txt and restart health-heartbeat."
  sleep infinity
fi

INTERVAL="${HEARTBEAT_INTERVAL_SECONDS:-60}"
TARGET_CONTAINERS="${HEARTBEAT_TARGET_CONTAINERS:-mc-world mc-velocity mc-ngrok}"
STARTUP_CHECK_INTERVAL="${HEARTBEAT_STARTUP_CHECK_INTERVAL_SECONDS:-5}"
LAST_STATE="unknown"

check_one() {
  container="$1"
  status="$(docker inspect -f '{{.State.Status}}' "$container" 2>/dev/null || echo missing)"
  if [ "$status" != "running" ]; then
    echo "$container:status=$status"
    return 1
  fi
  health="$(docker inspect -f '{{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}}' "$container" 2>/dev/null || echo missing)"
  if [ "$health" != "none" ] && [ "$health" != "healthy" ]; then
    echo "$container:health=$health"
    return 1
  fi
}

while true; do
  # Avoid false down alert on startup: start monitoring only after all targets are ready once.
  READY="yes"
  for container in $TARGET_CONTAINERS; do
    if ! reason="$(check_one "$container")"; then
      READY="no"
      echo "waiting for startup readiness: $reason"
      break
    fi
  done
  if [ "$READY" = "yes" ]; then
    break
  fi
  sleep "$STARTUP_CHECK_INTERVAL"
done

while true; do
  STATE="ok"
  REASON=""
  for container in $TARGET_CONTAINERS; do
    if ! reason="$(check_one "$container")"; then
      STATE="fail"
      REASON="$reason"
      break
    fi
  done

  if [ "$STATE" = "ok" ]; then
    wget -qO- "$PING_URL" >/dev/null 2>&1 || true
  else
    if [ "$LAST_STATE" != "fail" ]; then
      echo "heartbeat fail: $REASON"
      wget -qO- "$PING_URL/fail" >/dev/null 2>&1 || true
    fi
  fi
  LAST_STATE="$STATE"
  sleep "$INTERVAL"
done
