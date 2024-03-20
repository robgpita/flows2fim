#!/bin/bash
set -eo pipefail

echo "Building for Linux..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -buildvcs=false -o builds/linux-amd64/flows2fim main.go
echo "Linux build completed."
