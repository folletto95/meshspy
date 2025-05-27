# syntax=docker/dockerfile:1.4

#############################################
# Stage 1: builder                         #
#############################################
FROM golang:1.24-alpine AS builder

ARG PROTO_VERSION
ARG GOOS=linux
ARG GOARCH=amd64
ARG GOARM=

WORKDIR /app

# 1) Installa protoc, git, protoc-gen-go
RUN apk add --no-cache git protobuf protoc ca-certificates && \
    go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.30.0

# 2) Clona e genera pb/meshtastic
RUN git clone --depth 1 --branch "${PROTO_VERSION}" \
      https://github.com/meshtastic/protobufs.git protobufs && \
    mkdir -p pb/meshtastic && \
    for f in protobufs/meshtastic/*.proto; do \
      sed -e 's|option go_package = .*;|option go_package = "meshspy/pb/meshtastic";|' \
          "$f" > pb/meshtastic/"$(basename "$f")"; \
    done && \
    protoc \
      --proto_path=protobufs \
      --go_out=pb/meshtastic --go_opt=paths=source_relative \
      protobufs/meshtastic/*.proto

# 3) Copia il tuo main.go
COPY main.go ./

# 4) Inizializza il modulo e lascia che go mod tidy rilevi
#    sia gli import esterni sia il pkg locale in pb/meshtastic
RUN go mod init meshspy && \
    go mod tidy

# 5) Compila statico per il target
RUN CGO_ENABLED=0 \
    GOOS=${GOOS} GOARCH=${GOARCH} \
    $( [ -n "${GOARM}" ] && echo "GOARM=${GOARM}" ) \
    go build -o meshspy .

#############################################
# Stage 2: runtime                         #
#############################################
FROM alpine:latest
RUN apk add --no-cache ca-certificates

WORKDIR /root/
COPY --from=builder /app/meshspy .

RUN addgroup -S mesh && adduser -S -G mesh mesh
USER mesh

ENTRYPOINT ["./meshspy"]
