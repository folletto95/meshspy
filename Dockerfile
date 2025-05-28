# syntax=docker/dockerfile:1.4

###########################
# 🔨 STAGE: Builder
###########################

ARG BASE_IMAGE
FROM ${BASE_IMAGE:-golang:1.21-bullseye} AS builder

ARG GOOS=linux
ARG GOARCH=amd64
ARG GOARM

ENV CGO_ENABLED=0 \
    GOOS=${GOOS} \
    GOARCH=${GOARCH} \
    GOARM=${GOARM}

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN go build -ldflags="-s -w" -o meshspy .

###########################
# 🏁 STAGE: Runtime finale
###########################

FROM alpine:3.18

WORKDIR /app
COPY --from=builder /app/meshspy .

ENTRYPOINT ["./meshspy"]
