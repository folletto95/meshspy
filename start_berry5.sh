#!/bin/bash

# Script per avviare MeshSpy su Raspberry Pi 5
# Opzioni:
#   --clean : rimuove e riscarica l'immagine Docker
#   --log   : mostra i log del container dopo l'avvio

CONTAINER_NAME="meshspy"
IMAGE_NAME="nicbad/meshspy:latest"

CLEAN=false
LOG=false

for arg in "$@"; do
  case "$arg" in
    --clean)
      CLEAN=true
      shift
      ;;
    --log)
      LOG=true
      shift
      ;;
    *)
      echo "Uso: $0 [--clean] [--log]"
      exit 1
      ;;
  esac
done

if [ "$CLEAN" = true ]; then
  echo "Pulizia immagine Docker: rimozione di $IMAGE_NAME..."
  docker image rm -f $IMAGE_NAME || true
  echo "Pull immagine Docker: scarico $IMAGE_NAME..."
  docker pull $IMAGE_NAME
fi

CONTAINER_ID="$(docker ps -a -q -f name=${CONTAINER_NAME})"
if [ -n "$CONTAINER_ID" ]; then
  echo "Rimuovo il container esistente '${CONTAINER_NAME}'..."
  docker rm -f ${CONTAINER_NAME}
fi

echo "Avvio del container '${CONTAINER_NAME}' per Raspberry Pi 5..."
docker run -d \
  --name ${CONTAINER_NAME} \
  --platform linux/arm64 \
  --device /dev/ttyACM0:/dev/ttyACM0 \
  --privileged \
  -p 8080:8080 \
  -v ~/meshspy_data:/app/data \
  --env-file .env.runtime \
  -e SERIAL_PORT=/dev/ttyACM0 \
  -e BAUD_RATE=115200 \
  -e MQTT_BROKER=tcp://smpisa.ddns.net:1883 \
  -e MQTT_TOPIC=meshspy \
  -e MQTT_CLIENT_ID=meshspy-berry5 \
  -e MQTT_USER="testmeshspy" \
  -e MQTT_PASS="test1" \
  -e SEND_ALIVE_ON_START=true \
  -e NODE_DB_PATH=/app/data/nodes.db \
  $IMAGE_NAME

echo "Container '${CONTAINER_NAME}' avviato."

if [ "$LOG" = true ]; then
  echo "Avvio visualizzazione log. Premi Ctrl+C per interrompere."
  docker logs -f ${CONTAINER_NAME}
fi
