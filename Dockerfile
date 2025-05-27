# Dockerfile
# syntax=docker/dockerfile:1.4

#############################################
# 1) Builder: Go cross-compile              #
#############################################
FROM golang:1.24-alpine AS builder

# default build args
ARG GOOS=linux
ARG GOARCH=amd64
ARG GOARM=

WORKDIR /app

# copia moduli e scarica dipendenze
COPY go.mod go.sum ./
RUN go mod download

# copia sorgenti + binding Protobuf generati
COPY . .

# compila in statico per il target
RUN CGO_ENABLED=0 \
    GOOS=$GOOS \
    GOARCH=$GOARCH \
    GOARM=$GOARM \
    go build -o meshspy .

#############################################
# 2) Runtime minimal “scratch”              #
#############################################
FROM scratch
COPY --from=builder /app/meshspy /meshspy
ENTRYPOINT ["/meshspy"]
