#!/usr/bin/env bash
set -e

rm -rf out
export XDG_CACHE_HOME=/tmp/.cache

cd sumo-cli
go test -v ./...
cd ..
