# syntax=docker/dockerfile:1.4

###########################
# 🔨 STAGE: Builder
###########################

# L'immagine base viene passata da build.sh tramite BASE_IMAGE
ARG BASE_IMAGE
FROM ${BASE_IMAGE:-golang:1.21-bullseye} AS builder

# Parametri di compilazione Go (impostati da build.sh)
ARG GOOS=linux
ARG GOARCH=amd64
ARG GOARM

# Costruzione statica del binario (CGO disabilitato)
ENV CGO_ENABLED=0 \
    GOOS=${GOOS} \
    GOARCH=${GOARCH} \
    GOARM=${GOARM}

WORKDIR /app

# Scarica i moduli Go
COPY go.mod ./
COPY go.sum ./
RUN go mod download

# Copia tutti i sorgenti
COPY . .

# Compilazione binario con ottimizzazioni
RUN go build -ldflags="-s -w" -o meshspy .

###########################
# 🏁 STAGE: Runtime finale
###########################

# Immagine runtime minima
FROM alpine:3.18

WORKDIR /app

# Copia solo il binario compilato
COPY --from=builder /app/meshspy .

# Avvio del binario
ENTRYPOINT ["./meshspy"]
