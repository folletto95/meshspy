# syntax=docker/dockerfile:1.4

#############################################
# Stage 1: builder                         #
#############################################
FROM golang:1.24-alpine AS builder

# Parametri passati da build.sh
ARG GOOS=linux
ARG GOARCH=amd64
ARG GOARM=
ARG MODULE_PATH
ARG PROTO_VERSION

WORKDIR /app

# 1) Installa protoc, git e protoc-gen-go
RUN apk add --no-cache git protobuf protoc \
    && go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.30.0

# 2) Clona i .proto e patch per il go_package corretto
RUN git clone --depth 1 --branch ${PROTO_VERSION} \
      https://github.com/meshtastic/protobufs.git protobufs \
 && for f in protobufs/meshtastic/*.proto; do \
      sed -i 's|option go_package =.*;|option go_package = "'${MODULE_PATH}'/pb/meshtastic";|' "$f"; \
    done

# 3) Genera i binding Go in pb/meshtastic
RUN mkdir -p pb/meshtastic \
 && protoc \
      --proto_path=protobufs \
      --go_out=pb/meshtastic --go_opt=paths=source_relative \
      protobufs/meshtastic/*.proto

# 4) Inizializza il modulo Go
RUN go mod init ${MODULE_PATH}

# 5) Copia solo main.go (e altri sorgenti) e tidy
COPY main.go . 
RUN go get github.com/eclipse/paho.mqtt.golang@v1.5.0 \
           github.com/tarm/serial@latest \
           google.golang.org/protobuf@latest \
 && go mod tidy

# 6) Compila il binario statico
RUN CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} \
    $( [ -n "${GOARM}" ] && echo "GOARM=${GOARM}" ) \
    go build -o meshspy .

#############################################
# Stage 2: runtime                         #
#############################################
FROM alpine:latest
RUN apk add --no-cache ca-certificates

WORKDIR /root/
COPY --from=builder /app/meshspy .

# Utente non-root
RUN addgroup -S mesh && adduser -S -G mesh mesh
USER mesh

ENTRYPOINT ["./meshspy"]
