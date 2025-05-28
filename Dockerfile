# Dockerfile
# syntax=docker/dockerfile:1.4

#############################################
# 1) Builder: Go cross‐compile              #
#############################################
FROM golang:1.24-alpine AS builder

# Arg default, disponibili in questo stage
ARG GOOS=linux
ARG GOARCH=amd64
ARG GOARM=

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Cross-compile statico per target
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
