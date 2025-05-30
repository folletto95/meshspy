# syntax=docker/dockerfile:1.4

###########################
# 🔨 STAGE: Builder
###########################

ARG BASE_IMAGE
FROM ${BASE_IMAGE:-golang:1.21-bullseye} AS builder

ARG GOOS=linux
ARG GOARCH=arm
ARG GOARM=6

ENV CGO_ENABLED=0 \
    GOOS=${GOOS} \
    GOARCH=${GOARCH} \
    GOARM=${GOARM}

WORKDIR /app

# 🔁 Installa git condizionalmente (Alpine vs Debian)
RUN echo "🔧 Installing git depending on base image: ${BASE_IMAGE}" && \
    ( \
      command -v apt-get >/dev/null 2>&1 && apt-get update && apt-get install -y git || \
      command -v apk >/dev/null 2>&1 && apk add --no-cache git \
    )

# Scarica i moduli Go del progetto principale
COPY go.mod ./
COPY go.sum ./
RUN go mod download

# Copia i sorgenti principali
COPY . .

# ✅ COMPILA meshspy
RUN go build -ldflags="-s -w" -o meshspy ./cmd/meshspy

# ✅ CLONA E COMPILA meshtastic-go
RUN git clone https://github.com/lmatte7/meshtastic-go.git /tmp/meshtastic-go \
    && cd /tmp/meshtastic-go \
    && go build -ldflags="-s -w" -o /usr/local/bin/meshtastic-go \
    && chmod +x /usr/local/bin/meshtastic-go

###########################
# 🏁 STAGE: Runtime finale
###########################

FROM alpine:3.18

WORKDIR /app

# Copia binario principale
COPY --from=builder /app/meshspy .

# Copia binario meshtastic-go
COPY --from=builder /usr/local/bin/meshtastic-go /usr/local/bin/meshtastic-go

# Avvio del servizio principale
ENTRYPOINT ["./meshspy"]
