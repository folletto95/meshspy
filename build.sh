#!/usr/bin/env bash
set -euo pipefail

# 0) Carica variabili da .env
if [[ -f .env ]]; then
  # shellcheck disable=SC1091
  source .env
fi

# 1) Login automatico se configurato
if [[ -n "${DOCKER_USERNAME:-}" && -n "${DOCKER_PASSWORD:-}" ]]; then
  echo "$DOCKER_PASSWORD" | docker login docker.io \
    --username "$DOCKER_USERNAME" --password-stdin
fi

# 2) Parametri (override in .env o CLI)
IMAGE="${IMAGE:-nicbad/meshspy}"
TAG="${TAG:-latest}"
GOOS="linux"
ARCHS=(amd64 386 armv6 armv7 arm64)
PROTO_VERSION="${PROTO_VERSION:-v2.0.14}"

# 3) Genera i binding Protobuf sul host
echo "🔄 Generating Protobuf Go bindings (proto v${PROTO_VERSION})…"
docker run --rm \
  -v "${PWD}":/app -w /app \
  golang:1.24-alpine sh -euxc "\
    apk add --no-cache git protobuf && \
    go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.30.0 && \
    rm -rf protobufs pb && \
    git clone --depth 1 --branch ${PROTO_VERSION} https://github.com/meshtastic/protobufs.git protobufs && \
    mkdir -p pb/meshtastic && \
    protoc \
      --proto_path=protobufs/meshtastic \
      --go_out=pb/meshtastic --go_opt=paths=source_relative \
      protobufs/meshtastic/*.proto"

# verifica pb
if [[ ! -d pb/meshtastic ]]; then
  echo "❌ pb/meshtastic non trovata. Protobuf generation failed."
  exit 1
fi

# 4) (Ri)genera go.mod/go.sum includendo il proto package
echo "🛠 Ensuring go.mod has all deps (including protobuf/proto)…"
docker run --rm \
  -v "${PWD}":/app -w /app \
  golang:1.24-alpine sh -euxc "\
    if [ ! -f go.mod ]; then go mod init ${IMAGE#*/}; fi && \
    go get google.golang.org/protobuf/proto@v1.30.0 \
           github.com/eclipse/paho.mqtt.golang@v1.5.0 \
           github.com/tarm/serial@latest && \
    go mod tidy"

# 5) Build & push mono-arch & manifest
declare -A GOARCH=( [amd64]=amd64 [386]=386 [armv6]=arm [armv7]=arm [arm64]=arm64 )
declare -A GOARM=(  [armv6]=6     [armv7]=7                )
declare -A MAN_OPTS=(
  [amd64]="--os linux --arch amd64"
  [386]="--os linux --arch 386"
  [armv6]="--os linux --arch arm --variant v6"
  [armv7]="--os linux --arch arm --variant v7"
  [arm64]="--os linux --arch arm64"
)

echo "🛠 Building & pushing single-arch slices: ${ARCHS[*]}"
for arch in "${ARCHS[@]}"; do
  TAG_ARCH="${IMAGE}:${TAG}-${arch}"
  echo " • $TAG_ARCH"
  build_args=( --no-cache -t "$TAG_ARCH" )
  build_args+=( --build-arg "GOOS=$GOOS" )
  build_args+=( --build-arg "GOARCH=${GOARCH[$arch]}" )
  if [[ -n "${GOARM[$arch]:-}" ]]; then
    build_args+=( --build-arg "GOARM=${GOARM[$arch]}" )
  fi
  build_args+=( . )
  docker build "${build_args[@]}"
  docker push "$TAG_ARCH"
done

echo "📦 Creating & annotating manifest ${IMAGE}:${TAG}"
docker manifest rm "${IMAGE}:${TAG}" >/dev/null 2>&1 || true
manifest_args=( manifest create "${IMAGE}:${TAG}" )
for arch in "${ARCHS[@]}"; do
  manifest_args+=( "${IMAGE}:${TAG}-${arch}" )
done
docker "${manifest_args[@]}"
for arch in "${ARCHS[@]}"; do
  docker manifest annotate "${IMAGE}:${TAG}" \
    "${IMAGE}:${TAG}-${arch}" ${MAN_OPTS[$arch]}
done
docker manifest push "${IMAGE}:${TAG}"

echo "✅ All done! Multi-arch image: ${IMAGE}:${TAG}"
