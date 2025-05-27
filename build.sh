#!/usr/bin/env bash
# build.sh
set -euo pipefail

if [[ -f .env ]]; then source .env; fi

ARCHS=(amd64 386 armv6 armv7 arm64)
declare -A GOARCH=( [amd64]=amd64 [386]=386 [armv6]=arm [armv7]=arm [arm64]=arm64 )
declare -A GOARM=(  [armv6]=6     [armv7]=7                )
declare -A MAN_OPTS=(
  [amd64]="--os linux --arch amd64"
  [386]="--os linux --arch 386"
  [armv6]="--os linux --arch arm --variant v6"
  [armv7]="--os linux --arch arm --variant v7"
  [arm64]="--os linux --arch arm64"
)

# login (opzionale)
if [[ -n "${DOCKER_USERNAME:-}" && -n "${DOCKER_PASSWORD:-}" ]]; then
  echo "$DOCKER_PASSWORD" \
    | docker login --username "$DOCKER_USERNAME" --password-stdin
fi

# build & push single‐arch
for arch in "${ARCHS[@]}"; do
  TAG_ARCH="${IMAGE}:${TAG}-${arch}"
  echo "🔨 Building ${TAG_ARCH}"
  args=( --no-cache -t "$TAG_ARCH" )
  args+=( --build-arg PROTO_VERSION="$PROTO_VERSION" )
  args+=( --build-arg GOOS=linux )
  args+=( --build-arg GOARCH="${GOARCH[$arch]}" )
  [[ -n "${GOARM[$arch]:-}" ]] && args+=( --build-arg GOARM="${GOARM[$arch]}" )
  args+=( . )
  docker build "${args[@]}"
  docker push "$TAG_ARCH"
done

# crea e push manifest
echo "📦 Creating manifest ${IMAGE}:${TAG}"
docker manifest rm "${IMAGE}:${TAG}" 2>/dev/null || true
margs=( manifest create "${IMAGE}:${TAG}" )
for arch in "${ARCHS[@]}"; do
  margs+=( "${IMAGE}:${TAG}-${arch}" )
done
docker "${margs[@]}"
for arch in "${ARCHS[@]}"; do
  docker manifest annotate "${IMAGE}:${TAG}" \
    "${IMAGE}:${TAG}-${arch}" ${MAN_OPTS[$arch]}
done
docker manifest push "${IMAGE}:${TAG}"

echo "✅ Multi‐arch image available: ${IMAGE}:${TAG}"
