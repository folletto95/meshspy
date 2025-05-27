#!/usr/bin/env bash
set -euo pipefail

# carica .env
if [[ -f .env ]]; then source .env; fi

# fallback
MODULE_PATH=${MODULE_PATH:-github.com/nicbad/meshspy}
PROTO_VERSION=${PROTO_VERSION:-v2.0.14}
IMAGE=${IMAGE:-nicbad/meshspy}
TAG=${TAG:-latest}
ARCHS=(amd64 386 armv6 armv7 arm64)

# login Docker
if [[ -n "${DOCKER_USERNAME:-}" && -n "${DOCKER_PASSWORD:-}" ]]; then
  echo "$DOCKER_PASSWORD" | docker login --username "$DOCKER_USERNAME" --password-stdin
fi

echo "🔄 Fetching Meshtastic protobufs@$PROTO_VERSION and generating Go code…"
docker run --rm -v "${PWD}":/app -w /app golang:1.21-alpine sh -c "\
  apk add --no-cache git protobuf && \
  go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.30.0 && \
  rm -rf protobufs pb && \
  git clone --depth 1 --branch ${PROTO_VERSION} https://github.com/meshtastic/protobufs.git protobufs && \
  for f in protobufs/meshtastic/*.proto; do \
    sed -i 's|option go_package = .*;|option go_package = \"${MODULE_PATH}/pb/meshtastic\";|' \"\$f\"; \
  done && \
  mkdir -p pb/meshtastic && \
  protoc \
    --proto_path=protobufs \
    --go_out=pb/meshtastic --go_opt=paths=source_relative \
    protobufs/meshtastic/*.proto"

# genera go.mod/se non esiste
if [[ ! -f go.mod ]]; then
  echo "🛠 Generating go.mod and go.sum…"
  docker run --rm -v "${PWD}":/app -w /app golang:1.24-alpine sh -c "\
    go mod init ${MODULE_PATH} && \
    go get github.com/eclipse/paho.mqtt.golang@v1.5.0 \
           github.com/tarm/serial@latest \
           google.golang.org/protobuf@latest && \
    go mod tidy"
fi

# build & push multi-arch
declare -A GOARCH=( [amd64]=amd64 [386]=386 [armv6]=arm [armv7]=arm [arm64]=arm64 )
declare -A GOARM=(  [armv6]=6     [armv7]=7                )
declare -A MAN_OPTS=(
  [amd64]="--os linux --arch amd64"
  [386]="--os linux --arch 386"
  [armv6]="--os linux --arch arm --variant v6"
  [armv7]="--os linux --arch arm --variant v7"
  [arm64]="--os linux --arch arm64"
)

for arch in "${ARCHS[@]}"; do
  TAG_ARCH="${IMAGE}:${TAG}-${arch}"
  echo "🔨 Building ${TAG_ARCH}"
  args=( --no-cache -t "$TAG_ARCH" --build-arg "GOOS=linux" --build-arg "GOARCH=${GOARCH[$arch]}" )
  [[ -n "${GOARM[$arch]:-}" ]] && args+=( --build-arg "GOARM=${GOARM[$arch]}" )
  args+=( . )
  docker build "${args[@]}"
  docker push "$TAG_ARCH"
done

echo "📦 Creating manifest ${IMAGE}:${TAG}"
docker manifest rm "${IMAGE}:${TAG}" >/dev/null 2>&1 || true
margs=( manifest create "${IMAGE}:${TAG}" )
for arch in "${ARCHS[@]}"; do margs+=( "${IMAGE}:${TAG}-${arch}" ); done
docker "${margs[@]}"

for arch in "${ARCHS[@]}"; do
  docker manifest annotate "${IMAGE}:${TAG}" \
    "${IMAGE}:${TAG}-${arch}" ${MAN_OPTS[$arch]}
done

echo "🚀 Pushing multi-arch manifest ${IMAGE}:${TAG}"
docker manifest push "${IMAGE}:${TAG}"
echo "✅ Done!"
