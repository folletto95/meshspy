#!/usr/bin/env bash
set -euo pipefail

# ————————————————————————————
# 1. Caricamento .env (se esiste)
# ————————————————————————————
if [[ -f .env ]]; then
  set -o allexport
  source .env
  set +o allexport
fi

# ————————————————————————————
# 2. Controllo credenziali Docker
# ————————————————————————————
: "${DOCKER_USERNAME:?Devi impostare DOCKER_USERNAME (usa .env o segreti CI)}"
: "${DOCKER_PASSWORD:?Devi impostare DOCKER_PASSWORD (usa .env o segreti CI)}"
DOCKER_REGISTRY="${DOCKER_REGISTRY:-docker.io}"

echo "$DOCKER_PASSWORD" | docker login "$DOCKER_REGISTRY" \
  --username "$DOCKER_USERNAME" --password-stdin

# ————————————————————————————
# 3. Parametri configurabili
# ————————————————————————————
IMAGE="${IMAGE:-nicbad/meshspy}"
TAG="${TAG:-latest}"
GOOS="linux"
ARCHS=(amd64 386 armv6 armv7 arm64)

# ————————————————————————————
# 4. Bootstrap go.mod/go.sum (solo se mancanti)
# ————————————————————————————
if [[ ! -f go.mod ]]; then
  echo "🛠 Generating go.mod and go.sum…"
  docker run --rm \
    -v "${PWD}":/app -w /app \
    golang:1.24-alpine sh -c "\
      go mod init ${IMAGE#*/} && \
      go get github.com/eclipse/paho.mqtt.golang@v1.5.0 github.com/tarm/serial@latest && \
      go mod tidy"
fi

# ————————————————————————————
# 5. Mappe per build-arg e manifest annotate
# ————————————————————————————
declare -A GOARCH=( [amd64]=amd64 [386]=386 [armv6]=arm [armv7]=arm [arm64]=arm64 )
declare -A GOARM=(  [armv6]=6     [armv7]=7                )
declare -A MAN_OPTS=(
  [amd64]="--os linux --arch amd64"
  [386]="--os linux --arch 386"
  [armv6]="--os linux --arch arm --variant v6"
  [armv7]="--os linux --arch arm --variant v7"
  [arm64]="--os linux --arch arm64"
)

# ————————————————————————————
# 6. Build & Push mono-arch (incluso plugin)
# ————————————————————————————
echo "🛠 Building & pushing single-arch images for: ${ARCHS[*]}"
for arch in "${ARCHS[@]}"; do
  TAG_ARCH="${IMAGE}:${TAG}-${arch}"
  echo " • Building $TAG_ARCH"

  build_args=( --no-cache -t "$TAG_ARCH" )
  build_args+=( --build-arg "GOOS=$GOOS" )
  build_args+=( --build-arg "GOARCH=${GOARCH[$arch]}" )
  if [[ -n "${GOARM[$arch]:-}" ]]; then
    build_args+=( --build-arg "GOARM=${GOARM[$arch]}" )
  fi
  build_args+=( . )

  docker build "${build_args[@]}"
  echo " → Pushing $TAG_ARCH"
  docker push "$TAG_ARCH"
done

# ————————————————————————————
# 7. Manifest multi-arch
# ————————————————————————————
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
