#!/bin/bash
set -e

echo "Building Guix Calculator Example..."

# Generate Go code from .gx files
echo "Generating components..."
cd ../..
go run ./cmd/guix generate -p examples/calculator
cd examples/calculator

# Copy wasm_exec.js
echo "Copying wasm_exec.js..."
cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" .

# Build WASM
echo "Building WASM..."
GOOS=js GOARCH=wasm go build -o main.wasm .

echo "Build complete!"
echo "To run: python3 -m http.server 8080"
echo "Then open http://localhost:8080 in your browser"
