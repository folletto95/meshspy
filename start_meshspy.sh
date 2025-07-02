#!/bin/bash

# Script per avviare il container Docker MeshSpy
# Opzioni:
#   --clean : rimuove l'immagine e la riscarica
#   --log   : mostra i log del container ogni minuto

# Nome del container e dell'immagine
CONTAINER_NAME="meshspy"
IMAGE_NAME="nicbad/meshspy:latest"

# Valori di default per le opzioni
CLEAN=false
LOG=false

# Parsing degli argomenti
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

# Se richiesto, rimuovo l'immagine e la riscarico
if [ "$CLEAN" = true ]; then
  echo "Pulizia immagine Docker: rimozione di $IMAGE_NAME..."
  docker image rm -f $IMAGE_NAME || true
  echo "Pull immagine Docker: scarico $IMAGE_NAME..."
  docker pull $IMAGE_NAME
fi

# Se il container esiste già, lo rimuoviamo
if [ $(docker ps -a -q -f name=${CONTAINER_NAME}) ]; then
  echo "Rimuovo il container esistente '${CONTAINER_NAME}'..."
  docker rm -f ${CONTAINER_NAME}
fi

# Avvio del nuovo container MeshSpy
echo "Avvio del container '${CONTAINER_NAME}'..."
docker run -d \
  --name ${CONTAINER_NAME} \
  --device /dev/ttyACM0:/dev/ttyACM0 \
  --privileged \
  -p 8080:8080 \
  -v ~/meshspy_data:/app/data \
  -e SERIAL_PORT=/dev/ttyACM0 \
  -e BAUD_RATE=115200 \
  -e MQTT_BROKER=tcp://smpisa.ddns.net:1883 \
  -e MQTT_TOPIC=meshspy \
  -e MQTT_CLIENT_ID=meshspy-kali \
  -e MQTT_USER="testmeshspy" \
  -e MQTT_PASS="test1" \
  $IMAGE_NAME

echo "Container '${CONTAINER_NAME}' avviato con successo."

# Se richiesto, avvio il monitoraggio dei log ogni minuto
if [ "$LOG" = true ]; then
  echo "Avvio monitoraggio log del container ogni minuto. Premi Ctrl+C per interrompere."
  while true; do
    echo "---------- $(date +'%Y-%m-%d %H:%M:%S') ----------"
    docker logs --tail 50 ${CONTAINER_NAME}
    sleep 30
  done
fi
