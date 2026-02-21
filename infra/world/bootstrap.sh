#!/bin/sh
set -eu

secret="$(tr -d '\r\n' < /config/forwarding.secret)"
if [ -z "$secret" ]; then
  echo "forwarding.secret is empty" >&2
  exit 1
fi

mkdir -p \
  /data/config \
  /data/plugins/ClickMobs \
  /data/plugins/ClickMobsRegionGuard

cp /plugins/ClickMobs/config.yml /data/plugins/ClickMobs/config.yml
cp /plugins/ClickMobsRegionGuard.jar /data/plugins/ClickMobsRegionGuard.jar
cp /plugins/ClickMobsRegionGuard/config.yml /data/plugins/ClickMobsRegionGuard/config.yml

tmp="$(mktemp)"
{
  echo "proxies:"
  echo "  velocity:"
  echo "    enabled: true"
  echo "    online-mode: true"
  printf '    secret: "%s"\n' "$secret"
  if [ -f /data/config/paper-global.yml ]; then
    awk '
      BEGIN { skip = 0 }
      /^proxies:/ { skip = 1; next }
      skip == 1 {
        if ($0 ~ /^[^[:space:]]/) {
          skip = 0
          print
        }
        next
      }
      { print }
    ' /data/config/paper-global.yml
  fi
} > "$tmp"
mv "$tmp" /data/config/paper-global.yml

exec /image/scripts/start
