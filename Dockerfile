# syntax=docker/dockerfile:1.4

###########################
# 🔨 STAGE: Builder
###########################

ARG BASE_IMAGE
FROM ${BASE_IMAGE:-golang:1.21-alpine} AS builder

ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT

ENV CGO_ENABLED=1 \
    GOOS=${TARGETOS:-linux} \
    GOARCH=${TARGETARCH:-amd64}

WORKDIR /app

# 🔁 Installa git condizionalmente (Alpine vs Debian)
RUN echo "🔧 Installing build deps depending on base image: ${BASE_IMAGE}" && \
    if command -v apt-get >/dev/null 2>&1; then \
        apt-get update && apt-get install -y git build-essential sqlite3 libsqlite3-dev && rm -rf /var/lib/apt/lists/*; \
    elif command -v apk >/dev/null 2>&1; then \
        apk add --no-cache git build-base sqlite-dev; \
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
    go build -tags gui -ldflags="-s -w" -o meshspy ./cmd/meshspy

# ✅ COMPILA webapp
RUN GOARM=$(echo ${TARGETVARIANT} | tr -d 'v') \
    go build -ldflags="-s -w" -o webapp ./cmd/webapp

# ✅ CLONA E COMPILA meshtastic-go
RUN git clone https://github.com/lmatte7/meshtastic-go.git /tmp/meshtastic-go \
    && cd /tmp/meshtastic-go \
       && GOARM=$(echo ${TARGETVARIANT} | tr -d 'v') \
       && go build -ldflags="-s -w" -o /usr/local/bin/meshtastic-go \
       && chmod +x /usr/local/bin/meshtastic-go

###########################
# 🏁 STAGE: Runtime finale
###########################

ARG RUNTIME_IMAGE
FROM ${RUNTIME_IMAGE:-alpine:3.18}

WORKDIR /app

# Copia binario principale
COPY --from=builder /app/meshspy .

# Copia il binario webapp e la pagina HTML
RUN apk add --no-cache sqlite-libs ca-certificates && mkdir -p /app/web
COPY --from=builder /app/webapp /usr/local/bin/webapp
COPY --from=builder /app/cmd/webapp/index.html /app/web/index.html

# Copia binario meshtastic-go
COPY --from=builder /usr/local/bin/meshtastic-go /usr/local/bin/meshtastic-go

# Copia lo script di entrypoint
COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
RUN chmod +x /usr/local/bin/docker-entrypoint.sh

###########################
# 🛠️ ENV: Runtime config
###########################

# Copia il file .env.runtime nel container (se presente)
#RUN echo "copio .env.runtime"
#COPY .env.runtime /app/.env.runtime
#RUN echo "copiato .env.runtime"
RUN echo "copio .env.example"
COPY .env.example /app/.env.example
RUN echo "copiato .env.example"

# Avvio del servizio principale
ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]
