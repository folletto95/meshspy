# syntax=docker/dockerfile:1.4

########################
# 🛠️ STAGE: Builder
########################
ARG BASE_IMAGE
FROM ${BASE_IMAGE:-arm32v6/golang:1.21.0-alpine} AS builder

ARG GOOS=linux
ARG GOARCH=arm
ARG GOARM=6

ENV CGO_ENABLED=0 \
    GOOS=${GOOS} \
    GOARCH=${GOARCH} \
    GOARM=${GOARM}

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN GOOS=$GOOS GOARCH=$GOARCH GOARM=$GOARM CGO_ENABLED=$CGO_ENABLED \
    go build -ldflags="-s -w" -o meshspy ./cmd/meshspy

########################
# 🏁 STAGE: Runtime
########################
FROM alpine:3.18
WORKDIR /app
COPY --from=builder /app/meshspy .
ENTRYPOINT ["./meshspy"]
