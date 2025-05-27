# syntax=docker/dockerfile:1.4

#############################################
# 1) Builder: fetch proto, generate pb/, deps, build
#############################################
FROM golang:1.24-alpine AS builder

ARG PROTO_VERSION
ARG GOOS=linux
ARG GOARCH=amd64
ARG GOARM=

WORKDIR /app

# 1.1 Installa strumenti
RUN apk add --no-cache git protobuf protoc ca-certificates && \
    go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.30.0

# 1.2 Clona i proto e genera i binding in pb/meshtastic
RUN rm -rf protobufs pb && \
    git clone --depth 1 --branch "${PROTO_VERSION}" https://github.com/meshtastic/protobufs.git protobufs && \
    mkdir -p pb/meshtastic && \
    for f in protobufs/meshtastic/*.proto; do \
      sed -e 's|option go_package = .*;|option go_package = "meshspy/pb/meshtastic";|' \
          "$f" > pb/meshtastic/"$(basename "$f")"; \
    done && \
    protoc \
      --proto_path=protobufs \
      --go_out=pb/meshtastic --go_opt=paths=source_relative \
      protobufs/meshtastic/*.proto

# 1.3 Copia il codice applicativo (main.go) e il binding generato
COPY main.go ./

# 1.4 Inizializza modulo e scarica tutte le dipendenze:
RUN go mod init meshspy && \
    go get github.com/eclipse/paho.mqtt.golang@v1.5.0 \
           github.com/tarm/serial@latest \
           google.golang.org/protobuf@latest && \
    go mod tidy

# 1.5 Compila statico per la piattaforma target
RUN CGO_ENABLED=0 \
    GOOS=${GOOS} GOARCH=${GOARCH} \
    $( [ -n "${GOARM}" ] && echo "GOARM=${GOARM}" ) \
    go build -o meshspy .

#############################################
# 2) Runtime: immagine leggera
#############################################
FROM alpine:latest
RUN apk add --no-cache ca-certificates

WORKDIR /root/
COPY --from=builder /app/meshspy .

RUN addgroup -S mesh && adduser -S -G mesh mesh
USER mesh

ENTRYPOINT ["./meshspy"]
