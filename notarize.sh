#!/usr/bin/env bash
set -e

# Historical... use goreleaser instead
gon -log-level=info -log-json ./notarize.hcl
