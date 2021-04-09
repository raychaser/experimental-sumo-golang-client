#!/usr/bin/env bash
set -e

go env
echo $GOPATH
echo BUILD_GOOS:   $BUILD_GOOS
echo BUILD_GOARCH: $BUILD_GOARCH
export OUT=out/${BUILD_GOOS}_${BUILD_GOARCH}
rm -rf $OUT
mkdir -p $OUT
export OUT=`realpath $OUT`
echo OUT:          $OUT
cd sumo-cli
GOOS=$BUILD_GOOS GOARCH=$BUILD_GOARCH go build -o $OUT/sumo-cli
cd ..
