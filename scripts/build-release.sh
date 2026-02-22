#!/usr/bin/env sh
# Build standup binaries for macOS (for GitHub Releases).
# Upload the files from dist/ to a new release, then install.sh will work.
# Usage: ./scripts/build-release.sh

set -e

BINARY_NAME="standup"
DIST="dist"
mkdir -p "$DIST"

echo "Building ${BINARY_NAME} for darwin/arm64..."
GOOS=darwin GOARCH=arm64 go build -o "${DIST}/${BINARY_NAME}-darwin-arm64" .

echo "Building ${BINARY_NAME} for darwin/amd64..."
GOOS=darwin GOARCH=amd64 go build -o "${DIST}/${BINARY_NAME}-darwin-amd64" .

echo "Done. Binaries in ${DIST}/:"
ls -la "$DIST"
