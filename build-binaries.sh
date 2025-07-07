#!/usr/bin/env bash
set -euo pipefail

# Build cross platform binaries for Windows, macOS and Linux.
# Output is placed in the dist/ directory.

PLATFORMS=(
  "linux/amd64"
  "windows/amd64"
  "darwin/amd64"
)

APP_NAME="meshspy"
DIST_DIR="dist"
PACKAGE="./cmd/meshspy"

rm -rf "$DIST_DIR"
mkdir -p "$DIST_DIR"

for PLATFORM in "${PLATFORMS[@]}"; do
  OS="${PLATFORM%%/*}"
  ARCH="${PLATFORM##*/}"
  EXT=""
  if [[ "$OS" == "windows" ]]; then
    EXT=".exe"
  fi

  echo "Building for $OS/$ARCH..."
  GOOS="$OS" GOARCH="$ARCH" CGO_ENABLED=0 go build -ldflags="-s -w" -o "$DIST_DIR/${APP_NAME}-${OS}-${ARCH}${EXT}" "$PACKAGE"

done

echo "Binaries written to $DIST_DIR/"