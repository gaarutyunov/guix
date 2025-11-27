#!/bin/bash
set -e

echo "Building Guix Calculator Example..."

# Generate Go code from .gx files
echo "Generating components..."
cd ../..
go run ./cmd/guix generate -p examples/calculator
cd examples/calculator

# Patch generated files to workaround codegen issues
echo "Patching generated code..."
# Fix stateChannel references in calculator_gen.go
sed -i 's/handleNumber(stateChannel,/handleNumber(c.StateChannel,/g' calculator_gen.go
sed -i 's/handleOperator(stateChannel,/handleOperator(c.StateChannel,/g' calculator_gen.go
sed -i 's/handleClear(stateChannel)/handleClear(c.StateChannel)/g' calculator_gen.go
sed -i 's/handleEquals(stateChannel,/handleEquals(c.StateChannel,/g' calculator_gen.go

# Initialize the state channel with initial value in app_gen.go
sed -i '/c.stateChannel = make(chan CalculatorState, 10)/a\	c.stateChannel <- CalculatorState{Display: "0", PreviousValue: 0, Operator: "", WaitingForOperand: false}' app_gen.go

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
