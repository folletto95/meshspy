# Dockerfile
# syntax=docker/dockerfile:1.4

#############################################
# 1) Builder: Go cross-compile              #
#############################################
FROM golang:1.24-alpine AS builder

# Arg default
ARG GOOS=linux
ARG GOARCH=amd64
ARG GOARM=

WORKDIR /app

# Moduli: go.mod/go.sum generati dallo script
COPY go.mod go.sum ./
RUN go mod download

# Sorgenti + binding Protobuf generati
COPY . .

# Compila
RUN CGO_ENABLED=0 \
    GOOS=$GOOS \
    GOARCH=$GOARCH \
    GOARM=$GOARM \
    go build -o meshspy .

#############################################
# 2) Runtime lightweight “scratch”          #
#############################################
FROM scratch
COPY --from=builder /app/meshspy /meshspy
ENTRYPOINT ["/meshspy"]
