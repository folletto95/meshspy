# Stage 1: Build
FROM golang:1.24-alpine AS builder

# Imposta variabili di ambiente per Go build
ARG GOOS
ARG GOARCH
ARG GOARM
ENV GOOS=${GOOS}
ENV GOARCH=${GOARCH}
ENV GOARM=${GOARM}
ENV CGO_ENABLED=0

WORKDIR /app

# Copia file di configurazione mod
COPY go.mod go.sum ./
RUN go mod download

# Copia il codice sorgente
COPY . .

# Build binario (con supporto a GOARM se definito)
RUN if [ -n "${GOARM}" ]; then \
      go build -o meshspy; \
    else \
      go build -o meshspy; \
    fi

# Stage 2: Runtime
FROM alpine:3.19
COPY --from=builder /app/meshspy /meshspy
ENTRYPOINT ["/meshspy"]
