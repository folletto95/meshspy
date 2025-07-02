# syntax=docker/dockerfile:1.4

###########################
# 🔨 STAGE: Builder
###########################

ARG BASE_IMAGE
FROM ${BASE_IMAGE:-golang:1.21-bullseye} AS builder

ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT

ENV CGO_ENABLED=0 \
    GOOS=${TARGETOS:-linux} \
    GOARCH=${TARGETARCH:-amd64} \
    GOARM=${TARGETVARIANT#v}

WORKDIR /app

# 🔁 Installa git condizionalmente (Alpine vs Debian)
RUN echo "🔧 Installing git depending on base image: ${BASE_IMAGE}" && \
    if command -v apt-get >/dev/null 2>&1; then \
        apt-get update && apt-get install -y git; \
    elif command -v apk >/dev/null 2>&1; then \
        apk add --no-cache git; \
    else \
        echo "❌ Unsupported package manager" && exit 1; \
    fi

# Scarica i moduli Go del progetto principale
COPY go.mod ./
COPY go.sum ./
RUN go mod download

# Copia i sorgenti principali
COPY . .

# ✅ COMPILA meshspy
RUN GOARM=$(echo ${TARGETVARIANT} | tr -d 'v') \
    go build -ldflags="-s -w" -o meshspy ./cmd/meshspyù

# ✅ CLONA E COMPILA meshtastic-go
RUN git clone https://github.com/lmatte7/meshtastic-go.git /tmp/meshtastic-go \
    && cd /tmp/meshtastic-go \
       && GOARM=$(echo ${TARGETVARIANT} | tr -d 'v') \
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

###########################
# 🛠️ ENV: Runtime config
###########################

# Copia il file .env.runtime nel container (se presente)
RUN echo "copio .env.runtime"
COPY .env.runtime /app/.env.runtime
RUN echo "copiato .env.runtime"
RUN echo "copio .env.example"
COPY .env.example /app/.env.example
RUN echo "copiato .env.example"

# Avvio del servizio principale
ENTRYPOINT ["./meshspy"]
