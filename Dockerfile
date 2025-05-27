# syntax=docker/dockerfile:1.4

#############################################
# Stage 1: Builder                          #
#############################################
FROM golang:1.24-alpine AS builder
ARG GOOS=linux
ARG GOARCH=amd64
ARG GOARM=

WORKDIR /app

# 1) Copia moduli e scarica dipendenze
COPY go.mod go.sum ./
RUN go mod download

# 2) Copia sorgenti e binding
COPY . .

# 3) Scarica runtime Protobuf e serial/MQTT
RUN go get github.com/eclipse/paho.mqtt.golang@v1.5.0 \
           github.com/tarm/serial@latest \
           google.golang.org/protobuf@latest && \
    go mod tidy

# 4) Compila statico per target
RUN CGO_ENABLED=0 \
    GOOS=${GOOS} GOARCH=${GOARCH} \
    $( [ -n "${GOARM}" ] && echo "GOARM=${GOARM}" ) \
    go build -o meshspy .

#############################################
# Stage 2: Runtime                         #
#############################################
FROM alpine:latest
RUN apk add --no-cache ca-certificates

WORKDIR /root/
COPY --from=builder /app/meshspy .

RUN addgroup -S mesh && adduser -S -G mesh mesh
USER mesh

ENTRYPOINT ["./meshspy"]
