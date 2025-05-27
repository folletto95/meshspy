#!/usr/bin/env bash
set -euo pipefail

# carica .env
[[ -f .env ]] && source .env

# login Docker (opzionale, se vuoi pushare)
if [[ -n "${DOCKER_USERNAME:-}" && -n "${DOCKER_PASSWORD:-}" ]]; then
  echo "$DOCKER_PASSWORD" | docker login docker.io --username "$DOCKER_USERNAME" --password-stdin
fi

# parametri
IMAGE="${IMAGE:-nicbad/meshspy}"
TAG="${TAG:-latest}"
PROTO_VERSION="${PROTO_VERSION:-v2.0.14}"
GOOS="linux"
ARCHS=(amd64 386 armv6 armv7 arm64)

# 1) Genera i binding Protobuf in host
echo "🔄 Generating Protobuf bindings (v${PROTO_VERSION})…"
docker run --rm -v "$PWD":/app -w /app golang:1.24-alpine sh -euxc "
  apk add --no-cache git protobuf protoc && \
  go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.30.0 && \
  rm -rf protobufs pb && \
  git clone --depth 1 --branch ${PROTO_VERSION} https://github.com/meshtastic/protobufs.git protobufs && \
  mkdir -p pb/meshtastic && \
  protoc \
    --proto_path=protobufs/meshtastic \
    --go_out=pb/meshtastic --go_opt=paths=source_relative \
    protobufs/meshtastic/*.proto
"

# 2) Genera go.mod & go.sum in host con tutte le dipendenze
echo "🛠 Ensuring go.mod/go.sum…"
docker run --rm -v "$PWD":/app -w /app golang:1.24-alpine sh -euxc "
  if [ ! -f go.mod ]; then go mod init ${IMAGE#*/}; fi && \
  go get \
    google.golang.org/protobuf/proto@v1.30.0 \
    github.com/eclipse/paho.mqtt.golang@v1.5.0 \
    github.com/tarm/serial@latest && \
  go mod tidy
"

# 3) Build e push delle slice mono-arch + manifest
declare -A GOARCH=( [amd64]=amd64 [386]=386 [armv6]=arm [armv7]=arm [arm64]=arm64 )
declare -A GOARM=(  [armv6]=6    [armv7]=7                )
declare -A MAN_OPTS=(
  [amd64]="--os linux --arch amd64"
  [386]="--os linux --arch 386"
  [armv6]="--os linux --arch arm --variant v6"
  [armv7]="--os linux --arch arm --variant v7"
  [arm64]="--os linux --arch arm64"
)

echo "🛠 Building & pushing single-arch images: ${ARCHS[*]}"
for arch in "${ARCHS[@]}"; do
  TAG_ARCH="${IMAGE}:${TAG}-${arch}"
  echo " • $TAG_ARCH"
  docker build --no-cache -t "$TAG_ARCH" \
    --build-arg GOOS="$GOOS" \
    --build-arg GOARCH="${GOARCH[$arch]}" \
    $( [[ -n "${GOARM[$arch]:-}" ]] && echo --build-arg GOARM="${GOARM[$arch]}" ) \
    .
  docker push "$TAG_ARCH"
done

echo "📦 Creating manifest ${IMAGE}:${TAG}"
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

echo "✅ Multi-arch image ready: ${IMAGE}:${TAG}"
