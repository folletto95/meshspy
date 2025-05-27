# Dockerfile
# syntax=docker/dockerfile:1.4

#############################################
# 1) Builder: Go cross‐compile              #
#############################################
FROM golang:1.21-alpine AS builder

# dichiaro gli ARG qui, con default
ARG GOOS=linux
ARG GOARCH=amd64
ARG GOARM=

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# ora GOARM è sempre parte della stessa linea di env, mai un “comando” a sé
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
