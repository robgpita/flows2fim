#!/bin/bash
set -eo pipefail

echo "Building for Linux..."
GOOS=linux GOARCH=amd64 go build -buildvcs=false -o ./builds/linux-amd64/go-cli-app
echo "Linux build completed."
