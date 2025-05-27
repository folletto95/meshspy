# syntax=docker/dockerfile:1.4

#############################################
# 1) Builder: cross‐compile static Go binary
#############################################
FROM golang:1.24-alpine AS builder

# Questi ARG vengono passati da build.sh per ogni arch
ARG GOOS=linux
ARG GOARCH=amd64
ARG GOARM=

WORKDIR /app

# 1.1) Copia solo go.mod/go.sum e scarica le dipendenze
COPY go.mod go.sum ./
RUN go mod download

# 1.2) Copia tutto il sorgente (inclusi pb/meshtastic/*.pb.go)
COPY . .

# 1.3) Compila statico per il target
RUN CGO_ENABLED=0 \
    GOOS=${GOOS} GOARCH=${GOARCH} \
    $( [ -n "${GOARM}" ] && echo "GOARM=${GOARM}" ) \
    go build -o meshspy .

#############################################
# 2) Runtime: immagine minimal
#############################################
FROM alpine:latest
RUN apk add --no-cache ca-certificates

WORKDIR /root/
COPY --from=builder /app/meshspy .

# usa utente non‐root per sicurezza
RUN addgroup -S mesh && adduser -S -G mesh mesh
USER mesh

ENTRYPOINT ["./meshspy"]
