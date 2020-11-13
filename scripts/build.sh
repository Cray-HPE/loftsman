#!/bin/bash

set -e

this_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
cd $this_dir/../

mkdir -p ./.build
echo "Building darwin-amd64 version..."
CGO_ENABLED=0 GO111MODULE="on" GOOS=darwin ARCH=amd64 go build -o ./.build/loftsman-darwin-amd64
echo "Building linux-amd64 version..."
CGO_ENABLED=0 GO111MODULE="on" GOOS=linux ARCH=amd64 go build -o ./.build/loftsman-linux-amd64
chmod +x ./.build/loftsman-darwin-amd64
chmod +x ./.build/loftsman-linux-amd64
