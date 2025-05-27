# syntax=docker/dockerfile:1.4
#############################################
# 1) Builder: Go cross-compile              #
#############################################
FROM golang:1.24-alpine AS builder

ARG GOOS=linux
ARG GOARCH=amd64
ARG GOARM=

WORKDIR /app

# 1.1 Copia moduli e download
COPY go.mod go.sum ./
RUN go mod download

# 1.2 Copia tutto il codice (incl. pb/)
COPY . .

# 1.3 Scarica runtime Protobuf e serial/MQTT
RUN go get github.com/eclipse/paho.mqtt.golang@v1.5.0 \
           github.com/tarm/serial@latest \
           google.golang.org/protobuf@latest && \
    go mod tidy

# 1.4 Cross-compile statico
RUN CGO_ENABLED=0 \
    GOOS=${GOOS} GOARCH=${GOARCH} \
    $( [ -n "${GOARM}" ] && echo "GOARM=${GOARM}" ) \
    go build -o meshspy .

#############################################
# 2) Runtime minimal                        #
#############################################
FROM alpine:latest
RUN apk add --no-cache ca-certificates

WORKDIR /root/
COPY --from=builder /app/meshspy .

RUN addgroup -S mesh && adduser -S -G mesh mesh
USER mesh

ENTRYPOINT ["./meshspy"]
