# syntax=docker/dockerfile:1.4

#############################################
# Stage 1: builder                         #
#############################################
FROM golang:1.24-alpine AS builder

ARG PROTO_VERSION
ARG GOOS=linux
ARG GOARCH=amd64
ARG GOARM=

WORKDIR /app

# 1) Installa gli strumenti necessari
RUN apk add --no-cache git protobuf protoc ca-certificates && \
    go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.30.0

# 2) Clona i .proto e genera i binding in pb/meshtastic
RUN git clone --depth 1 --branch "${PROTO_VERSION}" \
      https://github.com/meshtastic/protobufs.git protobufs && \
    mkdir -p pb/meshtastic && \
    for f in protobufs/meshtastic/*.proto; do \
      # patch go_package per pacchetto locale "meshspy/pb/meshtastic"
      sed -e 's|option go_package = .*;|option go_package = "meshspy/pb/meshtastic";|' \
          "$f" > pb/meshtastic/"$(basename "$f")"; \
    done && \
    protoc \
      --proto_path=protobufs \
      --go_out=pb/meshtastic --go_opt=paths=source_relative \
      protobufs/meshtastic/*.proto

# 3) Inizializza modulo e scarica dipendenze
RUN go mod init meshspy && \
    go get github.com/eclipse/paho.mqtt.golang@v1.5.0 \
           github.com/tarm/serial@latest \
           google.golang.org/protobuf@latest && \
    go mod tidy

# 4) Copia il tuo codice applicativo
COPY main.go ./

# 5) Compila il binario statico per il target
RUN CGO_ENABLED=0 \
    GOOS=${GOOS} GOARCH=${GOARCH} \
    $( [ -n "${GOARM}" ] && echo "GOARM=${GOARM}" ) \
    go build -o meshspy .

#############################################
# Stage 2: runtime                         #
#############################################
FROM alpine:latest
RUN apk add --no-cache ca-certificates

WORKDIR /root/
COPY --from=builder /app/meshspy .

RUN addgroup -S mesh && adduser -S -G mesh mesh
USER mesh

ENTRYPOINT ["./meshspy"]
