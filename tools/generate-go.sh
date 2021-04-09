#!/usr/bin/env bash
set -e

rm -rf out

mkdir -p out/public
curl https://prod-api.sumologic.com/docs/sumologic-api.yaml -o out/public/public.yaml
java -jar lib/openapi-generator-cli-4.3.0.jar generate -i out/public/public.yaml -g go -o out/public
