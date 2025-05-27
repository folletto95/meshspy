#!/usr/bin/env bash
set -euo pipefail

# Carica .env (MODULE_PATH, IMAGE, TAG, PROTO_VERSION)
if [[ -f .env ]]; then source .env; fi

# Default
MODULE_PATH=${MODULE_PATH:-github.com/nicbad/meshspy}
PROTO_VERSION=${PROTO_VERSION:-v2.0.14}
IMAGE=${IMAGE:-nicbad/meshspy}
TAG=${TAG:-latest}
ARCHS=(amd64 386 armv6 armv7 arm64)

# Login Docker (opzionale)
if [[ -n "${DOCKER_USERNAME:-}" && -n "${DOCKER_PASSWORD:-}" ]]; then
  echo "$DOCKER_PASSWORD" | docker login --username "$DOCKER_USERNAME" --password-stdin
fi

# Costruisci e push multi‐arch
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
  echo "🔨 Building $TAG_ARCH"
  build_args=( --no-cache -t "$TAG_ARCH" )
  build_args+=( --build-arg GOOS=linux )
  build_args+=( --build-arg GOARCH=${GOARCH[$arch]} )
  [[ -n "${GOARM[$arch]:-}" ]] && build_args+=( --build-arg GOARM=${GOARM[$arch]} )
  build_args+=( --build-arg MODULE_PATH=$MODULE_PATH )
  build_args+=( --build-arg PROTO_VERSION=$PROTO_VERSION )
  build_args+=( . )
  docker build "${build_args[@]}"
  docker push "$TAG_ARCH"
done

echo "📦 Creating & pushing manifest ${IMAGE}:${TAG}"
docker manifest rm "${IMAGE}:${TAG}" 2>/dev/null || true
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

echo "✅ Done — multi‐arch image available: ${IMAGE}:${TAG}"
