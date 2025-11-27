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
GOROOT="$(go env GOROOT)"
# Try lib/wasm first (Go 1.24+), then misc/wasm (older versions)
if [ -f "$GOROOT/lib/wasm/wasm_exec.js" ]; then
  cp "$GOROOT/lib/wasm/wasm_exec.js" .
elif [ -f "$GOROOT/misc/wasm/wasm_exec.js" ]; then
  cp "$GOROOT/misc/wasm/wasm_exec.js" .
else
  echo "wasm_exec.js not found in GOROOT, downloading from GitHub"
  curl -o wasm_exec.js https://raw.githubusercontent.com/golang/go/go1.25.0/lib/wasm/wasm_exec.js
fi

# Build WASM
echo "Building WASM..."
GOOS=js GOARCH=wasm go build -o main.wasm .

echo "Build complete!"
echo "To run: python3 -m http.server 8080"
echo "Then open http://localhost:8080 in your browser"
