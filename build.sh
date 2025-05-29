#!/usr/bin/env bash
set -euo pipefail

IMAGE="nicbad/meshspy"
TAG="latest-armv6"

echo "🔧 Building ARMv6 image..."

docker buildx build \
  --platform linux/arm/v6 \
  --push \
  -t ${IMAGE}:${TAG} \
  --build-arg GOARCH=arm \
  --build-arg GOARM=6 \
  --build-arg BASE_IMAGE=arm32v6/golang:1.21.0-alpine \
  .

echo "✅ Build e push completati: ${IMAGE}:${TAG}"
