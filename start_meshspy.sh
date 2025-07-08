#!/bin/bash

# Script to start the MeshSpy Docker container
# Options:
#   --clean : remove the image and pull it again
#   --log   : show container logs every minute

# Container and image names
CONTAINER_NAME="meshspy"
IMAGE_NAME="nicbad/meshspy:latest"

# Default values for the options
CLEAN=false
LOG=false

# Parse command line arguments
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

# Remove the image and pull it again when requested
if [ "$CLEAN" = true ]; then
  echo "Pulizia immagine Docker: rimozione di $IMAGE_NAME..."
  docker image rm -f $IMAGE_NAME || true
  echo "Pull immagine Docker: scarico $IMAGE_NAME..."
  docker pull $IMAGE_NAME
fi

# Remove the container if it already exists
CONTAINER_ID="$(docker ps -a -q -f name=${CONTAINER_NAME})"
if [ -n "$CONTAINER_ID" ]; then
  echo "Rimuovo il container esistente '${CONTAINER_NAME}'..."
  docker rm -f ${CONTAINER_NAME}
fi

# Start a new MeshSpy container
echo "Avvio del container '${CONTAINER_NAME}'..."
docker run -d \
  --name ${CONTAINER_NAME} \
  --device /dev/ttyACM0:/dev/ttyACM0 \
  --privileged \
  -p 8080:8080 \
  -v ~/meshspy_data:/app/data \
  --env-file .env.runtime \
  -e SERIAL_PORT=/dev/ttyACM0 \
  -e BAUD_RATE=115200 \
  -e MQTT_BROKER=tcp://smpisa.ddns.net:1883 \
  -e MQTT_TOPIC=meshspy \
  -e MQTT_CLIENT_ID=meshspy-kali \
  -e SEND_ALIVE_ON_START=true \
  $IMAGE_NAME

echo "Container '${CONTAINER_NAME}' avviato con successo."

# Start continuous log output if requested
if [ "$LOG" = true ]; then
  echo "Avvio visualizzazione log. Premi Ctrl+C per interrompere."
  docker logs -f ${CONTAINER_NAME}
fi
