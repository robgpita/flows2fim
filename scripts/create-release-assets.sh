#!/bin/bash

# The script must be executed from the root of the repository
set -eo pipefail

./scripts/build-darwin-arm64.sh
./scripts/build-linux-amd64.sh
./scripts/build-windows-amd64.sh

echo "Creating release assets ..."
zip -j builds/flows2fim-windows-amd64.zip builds/windows-amd64/flows2fim.exe
tar -czvf builds/flows2fim-darwin-arm64.tar.gz -C builds/darwin-arm64 flows2fim
tar -czvf builds/flows2fim-linux-amd64.tar.gz -C builds/linux-amd64 flows2fim
echo "Release assets created"