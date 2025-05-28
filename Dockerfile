FROM golang:1.24-alpine AS builder

ARG PROTO_VERSION=v2.0.14
ARG GOOS=linux
ARG GOARCH=amd64
ARG GOARM

WORKDIR /src

RUN apk add --no-cache protobuf git protoc ca-certificates build-base curl && \
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

RUN git clone --depth 1 --branch "${PROTO_VERSION}" https://github.com/meshtastic/protobufs.git protobufs && \
    protoc \
      --proto_path=protobufs \
      --go_out=. \
      --go_opt=module=github.com/folletto95/meshspy \
      protobufs/meshtastic/*.proto

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} GOARM=${GOARM} \
    go build -ldflags="-s -w" -o meshspy .

FROM alpine:latest

RUN apk add --no-cache tzdata && \
    cp /usr/share/zoneinfo/Europe/Rome /etc/localtime && \
    echo "Europe/Rome" > /etc/timezone && apk del tzdata

WORKDIR /app
COPY --from=builder /src/meshspy .
ENTRYPOINT ["./meshspy"]
