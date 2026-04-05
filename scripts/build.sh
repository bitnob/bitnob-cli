#!/bin/bash

# Build script for bitnob-cli with version information

VERSION=${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")}
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE=$(date -u '+%Y-%m-%dT%H:%M:%SZ')

LDFLAGS="-X main.buildVersion=${VERSION} -X main.buildCommit=${COMMIT} -X main.buildDate=${DATE}"

echo "Building bitnob-cli..."
echo "Version: ${VERSION}"
echo "Commit: ${COMMIT}"
echo "Date: ${DATE}"

go build -ldflags "${LDFLAGS}" -o bitnob ./cmd/bitnob

if [ $? -eq 0 ]; then
    echo "Build successful! Binary created: ./bitnob"
else
    echo "Build failed!"
    exit 1
fi