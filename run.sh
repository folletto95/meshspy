#!/usr/bin/env bash
# run.sh
set -euo pipefail
if [[ -f .env ]]; then source .env; fi

SERIAL_PORT=${SERIAL_PORT:-/dev/ttyACM0}
SERIAL_GROUP=${SERIAL_GROUP:-dialout}
GID=$(getent group "$SERIAL_GROUP" | cut -d: -f3)
if [[ -z "$GID" ]]; then echo "❌ Gruppo '$SERIAL_GROUP' non trovato"; exit 1; fi

docker run -d \
  --platform linux/arm/v6 \
  --name meshspy \
  --device "${SERIAL_PORT}:${SERIAL_PORT}" \
  --group-add "$GID" \
  -e SERIAL_PORT="$SERIAL_PORT" \
  -e BAUD_RATE="${BAUD_RATE:-115200}" \
  -e MQTT_BROKER="${MQTT_BROKER:-tcp://smpisa.ddns.net:1883}" \
  -e MQTT_TOPIC="${MQTT_TOPIC:-meshspy/nodo/connesso}" \
  -e MQTT_CLIENT_ID="${MQTT_CLIENT_ID:-meshspy-berry}" \
  -e MQTT_USER="${MQTT_USER:-}" \
  -e MQTT_PASS="${MQTT_PASS:-}" \
  "${IMAGE}:${TAG}"
