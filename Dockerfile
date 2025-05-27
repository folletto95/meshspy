# Dockerfile

############################################
# Stage 1: builder
############################################
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copia tutto il sorgente
COPY . .

# Se non esiste go.mod, inizializza il modulo
RUN go mod init meshspy || true

# Prende le versioni corrette di paho.mqtt e tarm/serial
RUN go get github.com/eclipse/paho.mqtt.golang@v1.5.0 \
           github.com/tarm/serial@latest

# Pulisce e genera go.sum
RUN go mod tidy

# Compila statico per Linux
RUN CGO_ENABLED=0 GOOS=linux go build -o meshspy .

############################################
# Stage 2: runtime
############################################
FROM alpine:latest

# Certificati per TLS, se servono
RUN apk add --no-cache ca-certificates

WORKDIR /root/

# Copia il binario dal builder
COPY --from=builder /app/meshspy .

# Crea utente non-root
RUN addgroup -S mesh && adduser -S -G mesh mesh

USER mesh

# ENTRYPOINT
ENTRYPOINT ["./meshspy"]
