#!/usr/bin/env bash
set -euo pipefail

MODULE_NAME=${MODULE_NAME:-meshspy}
GO_IMAGE=golang:1.21-alpine

echo "🛠  Generazione go.mod e go.sum in container $GO_IMAGE…"
docker run --rm \
  -v "$(pwd)":/app -w /app \
  $GO_IMAGE sh -c "\
    if [ ! -f go.mod ]; then go mod init $MODULE_NAME; fi && \
    go get github.com/eclipse/paho.mqtt.golang@v1.5.0 github.com/tarm/serial@latest && \
    go mod tidy"

echo "🐳  Build dell’immagine Docker meshspy:latest…"
docker build -t meshspy:latest .
