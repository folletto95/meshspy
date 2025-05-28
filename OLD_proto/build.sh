#!/usr/bin/env bash
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

# optional Docker login
if [[ -n "${DOCKER_USERNAME:-}" && -n "${DOCKER_PASSWORD:-}" ]]; then
  echo "$DOCKER_PASSWORD" | docker login --username "$DOCKER_USERNAME" --password-stdin
fi

for arch in "${ARCHS[@]}"; do
  TAG_ARCH="${IMAGE}:${TAG}-${arch}"
  echo "🔨 Building $TAG_ARCH"
  docker build --no-cache \
    --build-arg PROTO_VERSION="$PROTO_VERSION" \
    --build-arg GOOS=linux \
    --build-arg GOARCH="${GOARCH[$arch]}" \
    $( [[ -n "${GOARM[$arch]:-}" ]] && echo "--build-arg GOARM=${GOARM[$arch]}" ) \
    -t "$TAG_ARCH" .
  docker push "$TAG_ARCH"
done

echo "📦 Creating & pushing manifest ${IMAGE}:${TAG}"
docker manifest rm "${IMAGE}:${TAG}" 2>/dev/null || true
margs=( manifest create "${IMAGE}:${TAG}" )
for arch in "${ARCHS[@]}"; do margs+=( "${IMAGE}:${TAG}-${arch}" ); done
docker "${margs[@]}"
for arch in "${ARCHS[@]}"; do
  docker manifest annotate "${IMAGE}:${TAG}" \
    "${IMAGE}:${TAG}-${arch}" ${MAN_OPTS[$arch]}
done
docker manifest push "${IMAGE}:${TAG}"

echo "✅ Done — image available: ${IMAGE}:${TAG}"
