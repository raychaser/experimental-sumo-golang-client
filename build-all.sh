#!/usr/bin/env bash
set -e

rm -rf out
export XDG_CACHE_HOME=/tmp/.cache

# Remove comments to enable tests as part of build-all
#cd sumo-cli
#go test -v ./...
#cd ..

BUILD_GOOS=linux BUILD_GOARCH=amd64 sh build.sh
BUILD_GOOS=windows BUILD_GOARCH=amd64 sh build.sh
BUILD_GOOS=darwin BUILD_GOARCH=amd64 sh build.sh


