# syntax=docker/dockerfile:1.4

#############################################
# 1) Builder: sfrutta go.mod+go.sum + pb/   #
#############################################
FROM golang:1.24-alpine AS builder

# default build args, overridabili da CLI
ARG GOOS=linux
ARG GOARCH=amd64
ARG GOARM=

WORKDIR /app

# copia moduli e dipendenze già generate in host
COPY go.mod go.sum ./
RUN go mod download

# copia tutto il contesto (incl. main.go, pb/, ...)
COPY . .

# compila statico per il target
RUN CGO_ENABLED=0 \
    GOOS=$GOOS \
    GOARCH=$GOARCH \
    GOARM=$GOARM \
    go build -o meshspy .

#############################################
# 2) Runtime ultra-light                    #
#############################################
FROM scratch
COPY --from=builder /app/meshspy /meshspy
ENTRYPOINT ["/meshspy"]
