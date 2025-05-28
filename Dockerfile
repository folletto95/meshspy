# Dockerfile
# syntax=docker/dockerfile:1.4

#############################################
# 1) Builder: Go cross‐compile + plugin     #
#############################################
FROM golang:1.24-alpine AS builder

ARG GOOS=linux
ARG GOARCH=amd64
ARG GOARM=

WORKDIR /app

# Dipendenze base
COPY go.mod go.sum ./
RUN go mod download

# Copia tutto il sorgente (inclusi main.go, plugin.go, tutti i .pb.go, ecc)
COPY . .

# Compila il binario principale
RUN CGO_ENABLED=0 \
    GOOS=${GOOS} \
    GOARCH=${GOARCH} \
    GOARM=${GOARM} \
    go build -trimpath -o meshspy .

# Compila il plugin (serve cgo abilitato, quindi installa toolchain C)
RUN apk add --no-cache gcc musl-dev
RUN CGO_ENABLED=1 \
    GOOS=${GOOS} \
    GOARCH=${GOARCH} \
    GOARM=${GOARM} \
    go build -buildmode=plugin -trimpath -o ghdownloader.so plugin.go

#############################################
# 2) Runtime minimal “scratch”              #
#############################################
FROM scratch

COPY --from=builder /app/meshspy      /meshspy
COPY --from=builder /app/ghdownloader.so /ghdownloader.so

WORKDIR /

ENTRYPOINT ["/meshspy"]
