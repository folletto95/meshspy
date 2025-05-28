# syntax=docker/dockerfile:1.4

FROM golang:1.24-alpine

ARG PROTO_VERSION=v2.0.14
ARG GOOS=linux
ARG GOARCH=amd64
ARG GOARM=

# 1) Installa protoc, git e protoc-gen-go
RUN apk add --no-cache git protobuf protoc ca-certificates && \
    go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.30.0

WORKDIR /src

# 2) Clona i .proto di Meshtastic e patcha il go_package
RUN git clone --depth 1 --branch "${PROTO_VERSION}" https://github.com/meshtastic/protobufs.git protobufs && \
    mkdir -p pb/meshtastic && \
    for f in protobufs/meshtastic/*.proto; do \
      sed -e 's|option go_package = .*;|option go_package = "meshspy/pb/meshtastic";|' \
          "$f" > pb/meshtastic/"$(basename "$f")"; \
    done

# 3) Genera i binding Go
RUN protoc \
      --proto_path=protobufs \
      --go_out=. \
      --go_opt=module=github.com/folletto95/meshspy \
      protobufs/meshtastic/*.proto

# 4) Copia il tuo file principale
COPY main.go .

# 5) Inizializza il modulo e risolvi tutte le dipendenze
RUN go mod init github.com/folletto95/meshspy && \
    go get github.com/eclipse/paho.mqtt.golang@v1.5.0 \
           github.com/tarm/serial@latest \
           google.golang.org/protobuf@latest && \
    go mod tidy

# 6) Compila il binario statico per la piattaforma target
RUN CGO_ENABLED=0 \
    GOOS=${GOOS} GOARCH=${GOARCH} \
    $( [ -n "${GOARM}" ] && echo "GOARM=${GOARM}" ) \
    go build -o meshspy .

# 7) Immagine runtime minimal
FROM alpine:latest
RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=0 /src/meshspy .

RUN addgroup -S mesh && adduser -S -G mesh mesh
USER mesh

ENTRYPOINT ["./meshspy"]
