# Dockerfile
FROM golang:1.21-bullseye AS builder

# Installazione tool di sistema e protoc
RUN apt-get update && apt-get install -y \
    curl \
    unzip \
    protobuf-compiler \
    zlib1g-dev \
    git \
    && rm -rf /var/lib/apt/lists/*

# Installazione plugin protoc-gen-go
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

WORKDIR /app

# Copia del codice sorgente
COPY . .

# Compilazione Go (se serve)
RUN go build -o meshspy .

# Costruzione immagine finale minimale
FROM debian:bullseye-slim

# Dipendenze runtime
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

WORKDIR /root/

# Copia del binario compilato
COPY --from=builder /app/meshspy .
COPY --from=builder /app/internal/proto /app/internal/proto

CMD ["./meshspy"]
