#!/bin/bash
set -eo pipefail

echo "Building for Windows..."
GOOS=windows GOARCH=amd64 go build -buildvcs=false -o ./builds/windows-amd64/flows2fim.exe
echo "Windows build completed."
