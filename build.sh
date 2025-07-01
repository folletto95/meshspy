#!/usr/bin/env bash
set -euo pipefail

# --- pulizia automatica a fine esecuzione/interruzione ---
cleanup() {
  rm -rf "$TMP_DIR" "$TMP_COPY_DIR" "$PROTO_MAP_FILE"
}
trap cleanup EXIT

# Carica variabili da .env se presente  
if [[ -f .env.build ]]; then
  source .env.build
fi

# Login automatico se configurato  
if [[ -n "${DOCKER_USERNAME:-}" && -n "${DOCKER_PASSWORD:-}" ]]; then
  echo "$DOCKER_PASSWORD" | docker login docker.io \
    --username "$DOCKER_USERNAME" --password-stdin
fi

# Parametri  
IMAGE="${IMAGE:-nicbad/meshspy}"
TAG="${TAG:-latest}"
MIN_PROTO="v2.0.14"

# directory temporanee
PROTO_REPO="https://github.com/meshtastic/protobufs.git"
TMP_DIR=$(mktemp -d)
TMP_COPY_DIR=$(mktemp -d)
PROTO_MAP_FILE=$(mktemp)

echo "📥 Recupero tag disponibili da $PROTO_REPO"
git ls-remote --tags "$PROTO_REPO" | awk '{print $2}' \
  | grep -E '^refs/tags/v[0-9]+\.[0-9]+\.[0-9]+$' \
  | sed 's|refs/tags/||' | sort -V \
  | while read -r PROTO_VERSION; do
      # salto versioni minori di MIN_PROTO
      if [[ "$(printf '%s\n' "$PROTO_VERSION" "$MIN_PROTO" | sort -V | head -1)" != "$MIN_PROTO" ]]; then
        echo "⏩ Salto $PROTO_VERSION (< $MIN_PROTO)"
        continue
      fi

      PROTO_DEST="proto/$PROTO_VERSION"
      if [[ -d "$PROTO_DEST" ]]; then
        echo "✔️ Proto già presenti: $PROTO_DEST"
        continue
      fi

      echo "📥 Scaricando proto $PROTO_VERSION…"
      rm -rf "$TMP_DIR"/*
      git clone --depth 1 --branch "$PROTO_VERSION" "$PROTO_REPO" "$TMP_DIR"
      mkdir -p "$TMP_COPY_DIR/$PROTO_VERSION"
      cp -r "$TMP_DIR/meshtastic" "$TMP_COPY_DIR/$PROTO_VERSION/"
      curl -sSL https://raw.githubusercontent.com/nanopb/nanopb/master/generator/proto/nanopb.proto \
           -o "$TMP_COPY_DIR/$PROTO_VERSION/nanopb.proto"

      echo "$PROTO_VERSION" >> "$PROTO_MAP_FILE"
done

# Compilazione dei proto  
if [[ -s "$PROTO_MAP_FILE" ]]; then
  echo "📦 Compilazione .proto in un unico container…"
  docker run --rm \
    -v "$PWD":/app \
    -v "$PWD/$TMP_COPY_DIR":/proto_copy \
    -w /app golang:1.21-bullseye bash -c '
      set -e
      apt-get update
      apt-get install -y unzip curl git protobuf-compiler
      go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.30.0
      export PATH=$PATH:$(go env GOPATH)/bin
      while read -r version; do
        echo "🔨 Compilo proto/$version"
        rm -rf proto/$version && mkdir -p proto/$version
        for f in /proto_copy/$version/*.proto; do
          [[ -f "$f" ]] || continue
          protoc \
            --experimental_allow_proto3_optional \
            -I /proto_copy/$version \
            --go_out=proto/$version \
            --go_opt=paths=source_relative \
            --go_opt=Mnanopb.proto=proto/$version/nanopb.proto \
            "$f"
        done
      done < '"$PROTO_MAP_FILE"'
    '
  rm -f "$PROTO_MAP_FILE"
fi

# ─── Download automatico di meshtastic-go ─────────────────────────────────────
MESHTASTIC_GO_VERSION="${MESHTASTIC_GO_VERSION:-v0.2.4}"
MESHTASTIC_GO_REPO="lmatte7/meshtastic-go"   # fork corretto
echo "📥 Scaricando meshtastic-go ${MESHTASTIC_GO_VERSION}…"
rm -rf meshtastic-go-bin && mkdir -p meshtastic-go-bin

RETRY_FLAGS="--tries=3 --retry-connrefused --timeout=30"
URL="https://github.com/${MESHTASTIC_GO_REPO}/releases/download/${MESHTASTIC_GO_VERSION}/meshtastic_go_linux_amd64"

if wget $RETRY_FLAGS -qO meshtastic-go-bin/meshtastic-go.tar.gz "$URL"; then
  tar -xzf meshtastic-go-bin/meshtastic-go.tar.gz -C meshtastic-go-bin \
    && chmod +x meshtastic-go-bin/meshtastic-go \
    && rm meshtastic-go-bin/meshtastic-go.tar.gz
  echo "✔️ meshtastic-go scaricato in meshtastic-go-bin/meshtastic-go"
else
  echo "❌ Errore: impossibile scaricare meshtastic-go da $URL" >&2
  exit 1
fi
# ───────────────────────────────────────────────────────────────────────────────

# Verifica o rigenera go.mod  
if [[ ! -f go.mod ]]; then
  echo "🛠 Inizializzo go.mod…"
  go mod init "${IMAGE#*/}"
fi
echo "🛠 Generating or fixing go.mod and go.sum…"
rm -f go.sum
docker run --rm \
  -v "${PWD}":/app -w /app golang:1.21-alpine sh -c "\
    go mod tidy"

# Setup buildx  
if ! docker buildx inspect meshspy-builder &>/dev/null; then
  docker buildx create --name meshspy-builder --use --driver docker-container --bootstrap
fi
docker buildx use meshspy-builder

# 🔨 Build ARMv6 separata  
echo "🐹 Build ARMv6 senza buildx (QEMU emulazione inclusa)"
docker buildx build \
  --platform linux/arm/v6 \
  --push \
  -t "${IMAGE}:${TAG}-armv6" \
  --build-arg GOARCH=arm \
  --build-arg GOARM=6 \
  --build-arg BASE_IMAGE=arm32v6/golang:1.21.0-alpine \
  .

# 🚀 Build multipiattaforma  
echo "🚀 Build & push multipiattaforma per: linux/arm/v7,linux/amd64,linux/386,linux/arm64"
docker buildx build \
  --platform linux/arm/v7,linux/amd64,linux/386,linux/arm64 \
  --push \
  -t "${IMAGE}:${TAG}" \
  --build-arg BASE_IMAGE=golang:1.21-bullseye \
  .

# 🔗 Unione ARMv6 nel manifest principale  
echo "🔗 Creazione manifest multipiattaforma completo"
docker manifest create "${IMAGE}:${TAG}" \
  "${IMAGE}:${TAG}-armv6" \
  "${IMAGE}:${TAG}"
docker manifest push "${IMAGE}:${TAG}"

echo "✅ Done! Multiarch image ready: ${IMAGE}:${TAG}"
