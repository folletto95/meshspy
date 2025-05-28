# syntax=docker/dockerfile:1.4

#############################################
# 1) Builder: Go cross‐compile + plugin     #
#############################################
FROM golang:1.24-alpine AS builder

ARG GOOS=linux
ARG GOARCH=amd64
ARG GOARM=

WORKDIR /app

# Dipendenze Go
COPY go.mod go.sum ./
RUN go mod download

# Copia tutto il sorgente
COPY . .

# Installa protoc (protoc e plugin go)
RUN apk add --no-cache git protobuf protoc gcc musl-dev

# --- GENERA I .pb.go PER TUTTE LE VERSIONI DEI PROTOBUF --- #
# (aggiungi qui un blocco RUN per ogni versione supportata)
RUN mkdir -p pb/meshtastic-v2.0.14/meshtastic && \
    protoc --proto_path=proto/v2.0.14/meshtastic \
      --go_out=pb/meshtastic-v2.0.14/meshtastic \
      proto/v2.0.14/meshtastic/*.proto

RUN mkdir -p pb/meshtastic-v2.1.0/meshtastic && \
    protoc --proto_path=proto/v2.1.0/meshtastic \
      --go_out=pb/meshtastic-v2.1.0/meshtastic \
      proto/v2.1.0/meshtastic/*.proto

# --- (Aggiungi blocchi RUN per ogni nuova versione!) --- #

# Compila il binario principale
RUN CGO_ENABLED=0 \
    GOOS=${GOOS} \
    GOARCH=${GOARCH} \
    GOARM=${GOARM} \
    go build -trimpath -o meshspy .

# Compila il plugin (cgo abilitato, per plugin Go)
RUN CGO_ENABLED=1 \
    GOOS=${GOOS} \
    GOARCH=${GOARCH} \
    GOARM=${GOARM} \
    go build -buildmode=plugin -trimpath -o ghdownloader.so ./plugin.go

#############################################
# 2) Runtime minimal “scratch”              #
#############################################
FROM scratch

COPY --from=builder /app/meshspy         /meshspy
COPY --from=builder /app/ghdownloader.so /ghdownloader.so

WORKDIR /

ENTRYPOINT ["/meshspy"]
