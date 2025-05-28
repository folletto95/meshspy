# syntax=docker/dockerfile:1.4

# Set the target platform variable
ARG TARGETPLATFORM

# Use different base images depending on the platform
FROM --platform=$TARGETPLATFORM arm32v6/golang:1.22.9-alpine AS builder-armv6
FROM --platform=$TARGETPLATFORM golang:1.21-bullseye AS builder-default

# Choose builder depending on the platform
FROM builder-${TARGETPLATFORM//\//-} AS builder

# Installazione tool di sistema e protoc
RUN apt-get update && apt-get install -y \
    curl \
    unzip \
    protobuf-compiler \
    libprotobuf-dev \
    zlib1g-dev \
    git \
    && rm -rf /var/lib/apt/lists/*

# Installazione plugin protoc per Go
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.30.0

# Imposta directory di lavoro
WORKDIR /app

# Copia codice sorgente
COPY . .

# Verifica coerenza go.mod / go.sum e ricrea se necessario
RUN test -f go.mod && test -f go.sum || go mod init meshspy
RUN go mod tidy

# Compilazione Go
RUN go build -o meshspy .

# Costruzione immagine finale minimale
FROM debian:bullseye-slim
WORKDIR /root/
COPY --from=builder /app/meshspy .
ENTRYPOINT ["./meshspy"]
