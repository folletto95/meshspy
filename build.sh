#!/usr/bin/env bash
set -euo pipefail

# Carica variabili da .env se presente
if [[ -f .env ]]; then
  # shellcheck disable=SC1091
  source .env
fi

# Login automatico se configurato
if [[ -n "${DOCKER_USERNAME:-}" && -n "${DOCKER_PASSWORD:-}" ]]; then
  echo "$DOCKER_PASSWORD" | docker login docker.io \
    --username "$DOCKER_USERNAME" --password-stdin
fi

# Parametri (override in .env o CLI)
IMAGE="${IMAGE:-nicbad/meshspy}"
TAG="${TAG:-latest}"
GOOS="linux"
ARCHS=(amd64 386 armv6 armv7 arm64)

# === STEP: Scarica e compila proto Meshtastic ≥ v2.0.14 ===
PROTO_REPO="https://github.com/meshtastic/protobufs.git"
TMP_DIR=".proto_tmp"
PROTO_MAP_FILE=".proto_compile_map.sh"
rm -f "$PROTO_MAP_FILE"

echo "📥 Recupero tag disponibili da $PROTO_REPO"
git ls-remote --tags "$PROTO_REPO" | awk '{print $2}' |
  grep -E '^refs/tags/v[0-9]+\.[0-9]+\.[0-9]+$' | sed 's|refs/tags/||' | sort -V | while read -r PROTO_VERSION; do
  if [[ "$(printf '%s\n' "$PROTO_VERSION" v2.0.14 | sort -V | head -n1)" != "v2.0.14" ]]; then
    echo "⏩ Salto $PROTO_VERSION (proto non standard)"
    continue
  fi
  PROTO_DIR="internal/proto/${PROTO_VERSION}"
  if [[ -d "${PROTO_DIR}" ]]; then
    echo "✔️ Proto già presenti: $PROTO_DIR"
    continue
  fi

  echo "📥 Scaricando proto $PROTO_VERSION…"
  rm -rf "$TMP_DIR"
  git clone --depth 1 --branch "$PROTO_VERSION" "$PROTO_REPO" "$TMP_DIR"
  mkdir -p "/tmp/proto-${PROTO_VERSION}-copy"
  cp -r "$TMP_DIR/meshtastic" "/tmp/proto-${PROTO_VERSION}-copy/"
  # Scarica nanopb.proto nella directory temporanea
  curl -sSL https://raw.githubusercontent.com/nanopb/nanopb/master/generator/proto/nanopb.proto \
    -o "/tmp/proto-${PROTO_VERSION}-copy/nanopb.proto"

  echo "$PROTO_VERSION" >> "$PROTO_MAP_FILE"
  rm -rf "$TMP_DIR"
done

# Compila tutti i proto raccolti in un unico container
if [[ -s "$PROTO_MAP_FILE" ]]; then
  echo "📦 Compilazione .proto in un unico container…"
  docker run --rm \
    -v "$PWD":/app \
    -v /tmp:/tmp \
    -w /app \
    golang:1.21-bullseye bash -c '
      set -e
      apt-get update
      apt-get install -y unzip curl protobuf-compiler
      go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
      export PATH=$PATH:$(go env GOPATH)/bin
      while read -r version; do
        rm -rf internal/proto/$version
        mkdir -p internal/proto/$version
        protoc \
          --experimental_allow_proto3_optional \
          -I /tmp/proto-$version-copy \
          --go_out=internal/proto/$version \
          --go_opt=paths=source_relative \
          /tmp/proto-$version-copy/meshtastic/*.proto \
          /tmp/proto-$version-copy/nanopb.proto || true
      done < '"$PROTO_MAP_FILE"'
    '
  rm -f "$PROTO_MAP_FILE"
fi

# Se manca go.mod, lo generiamo con Go ≥1.21
if [[ ! -f go.mod ]]; then
  echo "🛠 Generating go.mod and go.sum…"
  docker run --rm \
    -v "${PWD}":/app -w /app \
    golang:1.21-alpine sh -c "\
      go mod init ${IMAGE#*/} && \
      go get github.com/eclipse/paho.mqtt.golang@v1.5.0 github.com/tarm/serial@latest && \
      go mod tidy"
fi

# Mappe per build-arg e manifest annotate
declare -A GOARCH=( [amd64]=amd64 [386]=386 [armv6]=arm [armv7]=arm [arm64]=arm64 )
declare -A GOARM=(  [armv6]=6     [armv7]=7                )
declare -A MAN_OPTS=(
  [amd64]="--os linux --arch amd64"
  [386]="--os linux --arch 386"
  [armv6]="--os linux --arch arm --variant v6"
  [armv7]="--os linux --arch arm --variant v7"
  [arm64]="--os linux --arch arm64"
)

# Assicura che buildx sia attivo
if ! docker buildx inspect meshspy-builder &>/dev/null; then
  docker buildx create --name meshspy-builder --use
fi
docker buildx use meshspy-builder
docker buildx inspect --bootstrap

echo "🛠 Building & pushing single-arch images for: ${ARCHS[*]}"
for arch in "${ARCHS[@]}"; do
  TAG_ARCH="${IMAGE}:${TAG}-${arch}"
  echo " • Building $TAG_ARCH"

  build_args=( --platform "linux/${GOARCH[$arch]}"
              --no-cache --push -t "$TAG_ARCH"
              --build-arg "GOOS=$GOOS"
              --build-arg "GOARCH=${GOARCH[$arch]}" )

  if [[ -n "${GOARM[$arch]:-}" ]]; then
    build_args+=( --platform "linux/arm/v${GOARM[$arch]}" )
    build_args+=( --build-arg "GOARM=${GOARM[$arch]}" )
  fi
  build_args+=( . )
  docker buildx build "${build_args[@]}"
done

echo "📦 Preparing manifest ${IMAGE}:${TAG}"
docker manifest rm "${IMAGE}:${TAG}" >/dev/null 2>&1 || true

manifest_args=( manifest create "${IMAGE}:${TAG}" )
for arch in "${ARCHS[@]}"; do
  manifest_args+=( "${IMAGE}:${TAG}-${arch}" )
done
docker "${manifest_args[@]}"

echo "⚙️ Annotating slices"
for arch in "${ARCHS[@]}"; do
  docker manifest annotate "${IMAGE}:${TAG}" \
    "${IMAGE}:${TAG}-${arch}" ${MAN_OPTS[$arch]}
done

echo "🚀 Pushing multi-arch manifest ${IMAGE}:${TAG}"
docker manifest push "${IMAGE}:${TAG}"

echo "✅ Done! Multi-arch image available: ${IMAGE}:${TAG}"
