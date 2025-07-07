# syntax=docker/dockerfile:1.4

###########################
# 🔨 STAGE: Builder
###########################

ARG BASE_IMAGE
FROM ${BASE_IMAGE:-golang:1.22-alpine} AS builder

ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT

ENV CGO_ENABLED=1 \
    GOOS=${TARGETOS:-linux} \
    GOARCH=${TARGETARCH:-amd64}

WORKDIR /app

# 🔁 Install git conditionally (Alpine vs Debian)
RUN echo "🔧 Installing build deps depending on base image: ${BASE_IMAGE}" && \
    if command -v apt-get >/dev/null 2>&1; then \
        apt-get update && apt-get install -y \
            git build-essential sqlite3 libsqlite3-dev \
            libgl1-mesa-dev libxi-dev libx11-dev libxcursor-dev \
            libxrandr-dev libxinerama-dev xorg-dev && \
        rm -rf /var/lib/apt/lists/*; \
    elif command -v apk >/dev/null 2>&1; then \
        apk add --no-cache \
            git build-base sqlite-dev mesa-dev \
            libx11-dev libxi-dev libxcursor-dev \
            libxrandr-dev libxinerama-dev; \
    else \
        echo "❌ Unsupported package manager" && exit 1; \
    fi

# Download the main project's Go modules
COPY go.mod ./
COPY go.sum ./
RUN go mod download

# Copy the main sources
COPY . .

# ✅ Build meshspy
RUN GOARM=$(echo ${TARGETVARIANT} | tr -d 'v') \
    go build -ldflags="-s -w" -o meshspy ./cmd/meshspy

# ✅ Build webapp
RUN GOARM=$(echo ${TARGETVARIANT} | tr -d 'v') \
    go build -ldflags="-s -w" -o webapp ./cmd/webapp

# ✅ Clone and build meshtastic-go
RUN git clone https://github.com/lmatte7/meshtastic-go.git /tmp/meshtastic-go \
    && cd /tmp/meshtastic-go \
       && GOARM=$(echo ${TARGETVARIANT} | tr -d 'v') \
       && go build -ldflags="-s -w" -o /usr/local/bin/meshtastic-go \
       && chmod +x /usr/local/bin/meshtastic-go

###########################
# 🏁 STAGE: Final runtime
###########################

ARG RUNTIME_IMAGE
FROM ${RUNTIME_IMAGE:-alpine:3.18}

WORKDIR /app

# Copy main binary
COPY --from=builder /app/meshspy .

# Copy the webapp binary and the HTML page
RUN apk add --no-cache sqlite-libs ca-certificates && mkdir -p /app/web
COPY --from=builder /app/webapp /usr/local/bin/webapp
COPY --from=builder /app/cmd/webapp/index.html /app/web/index.html

# Copy meshtastic-go binary
COPY --from=builder /usr/local/bin/meshtastic-go /usr/local/bin/meshtastic-go

# Copy the entrypoint script
COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
RUN chmod +x /usr/local/bin/docker-entrypoint.sh

###########################
# 🛠️ ENV: Runtime config
###########################

# Copy the .env.runtime file into the container (if present)
RUN echo "copio .env.runtime"
COPY .env.runtime /app/.env.runtime
RUN echo "copiato .env.runtime"
RUN echo "copio .env.example"
COPY .env.example /app/.env.example
RUN echo "copiato .env.example"

# Start the main service
ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]
