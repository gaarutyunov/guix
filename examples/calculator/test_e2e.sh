#!/bin/bash

set -e

echo "Building calculator WASM..."
bash build.sh

echo ""
echo "Running E2E tests..."
go test -v -tags=e2e -timeout=60s "$@"
