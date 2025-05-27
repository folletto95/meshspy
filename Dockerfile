# Dockerfile

############################################
# Stage 1: builder
############################################
FROM golang:1.21-alpine AS builder
WORKDIR /app

# Copia tutto il sorgente, inclusa la cartella pb/ con i .pb.go generati
COPY . .

# Se non esiste go.mod, inizializza il modulo
RUN go mod init meshspy || true

# Scarica le dipendenze (incluso protobuf runtime)
RUN go get github.com/eclipse/paho.mqtt.golang@v1.5.0 \
           github.com/tarm/serial@latest \
           google.golang.org/protobuf@latest && \
    go mod tidy

# Compila statico per Linux
ARG GOOS
ARG GOARCH
ARG GOARM
RUN CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} \
    $( [ -n "${GOARM:-}" ] && echo "GOARM=${GOARM}" ) \
    go build -o meshspy .

############################################
# Stage 2: runtime
############################################
FROM alpine:latest
RUN apk add --no-cache ca-certificates

WORKDIR /root/
COPY --from=builder /app/meshspy .

# Utente non-root
RUN addgroup -S mesh && adduser -S -G mesh mesh
USER mesh

ENTRYPOINT ["./meshspy"]
