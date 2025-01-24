#!/bin/bash

# The script must be executed from the root of the repository & within the docker container
set -eo pipefail

echo "Building for Darwin (MacOS) ARM64..."
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -buildvcs=false \
    -ldflags="-X main.GitTag=$(git describe --tags --always --dirty) -X main.GitCommit=$(git rev-parse --short HEAD) -X main.BuildDate=$(date +%Y-%m-%d)" \
    -o builds/darwin-arm64/flows2fim main.go
echo "Mac build completed."

chmod +x builds/darwin-arm64/flows2fim
