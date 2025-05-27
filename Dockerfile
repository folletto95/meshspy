# Dockerfile
# syntax=docker/dockerfile:1.4

#############################################
# 1) Builder: generate protobuf + compile  #
#############################################
FROM golang:1.24-alpine AS builder

ARG PROTO_VERSION
ARG GOOS=linux
ARG GOARCH=amd64
ARG GOARM=

WORKDIR /app

# install protoc, git, protoc-gen-go
RUN apk add --no-cache git protobuf protoc && \
    go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.30.0

# fetch Meshtastic .proto and patch go_package
RUN rm -rf protobufs pb && \
    git clone --depth 1 --branch "${PROTO_VERSION}" \
      https://github.com/meshtastic/protobufs.git protobufs && \
    mkdir -p pb/meshtastic && \
    for f in protobufs/meshtastic/*.proto; do \
      sed -i 's|option go_package = .*;|option go_package = "meshspy/pb/meshtastic";|' "$f"; \
    done && \
    protoc \
      --proto_path=protobufs \
      --go_out=pb/meshtastic --go_opt=paths=source_relative \
      protobufs/meshtastic/*.proto

# init module as local "meshspy"
RUN go mod init meshspy

# get deps (including pb/meshtastic locally) and tidy
RUN go get github.com/eclipse/paho.mqtt.golang@v1.5.0 \
           github.com/tarm/serial@latest \
           google.golang.org/protobuf@latest && \
    go mod tidy

# copy rest of sources and compile
COPY main.go ./
RUN CGO_ENABLED=0 \
    GOOS=${GOOS} GOARCH=${GOARCH} \
    $( [ -n "${GOARM}" ] && echo "GOARM=${GOARM}" ) \
    go build -o meshspy .

#############################################
# 2) Runtime: minimal image                #
#############################################
FROM alpine:latest
RUN apk add --no-cache ca-certificates

WORKDIR /root/
COPY --from=builder /app/meshspy .

RUN addgroup -S mesh && adduser -S -G mesh mesh
USER mesh

ENTRYPOINT ["./meshspy"]
