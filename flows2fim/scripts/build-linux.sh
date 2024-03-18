#!/bin/bash
set -eo pipefail

echo "Building for Linux..."
GOOS=linux GOARCH=amd64 go build -buildvcs=false -o ./builds/linux-amd64/flows2fim
echo "Linux build completed."
