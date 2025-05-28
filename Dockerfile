# Dockerfile
# syntax=docker/dockerfile:1.4

#############################################
# 1) Builder: Go cross‐compile + plugin     #
#############################################
FROM golang:1.24-alpine AS builder

# ARG disponibili per go build (e plugin build)
ARG GOOS=linux
ARG GOARCH=amd64
ARG GOARM=

WORKDIR /app

# Copia i moduli e scarica dipendenze
COPY go.mod go.sum ./
RUN go mod download

# Copia tutto il sorgente
COPY . .

# Compila il binario principale
RUN CGO_ENABLED=0 \
    GOOS=${GOOS} \
    GOARCH=${GOARCH} \
    GOARM=${GOARM} \
    go build -trimpath -o meshspy .

# Compila il plugin (per lo stesso GOOS/GOARCH/GOARM)
RUN CGO_ENABLED=0 \
    GOOS=${GOOS} \
    GOARCH=${GOARCH} \
    GOARM=${GOARM} \
    go build -buildmode=plugin -trimpath -o ghdownloader.so .


#############################################
# 2) Runtime minimal “scratch”              #
#############################################
FROM scratch

# Copia il binario e il plugin
COPY --from=builder /app/meshspy      /meshspy
COPY --from=builder /app/ghdownloader.so /ghdownloader.so

# Imposta il working dir (facoltativo)
WORKDIR /

# Lancia direttamente il binario
ENTRYPOINT ["/meshspy"]
