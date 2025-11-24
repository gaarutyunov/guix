# Counter Example

A simple interactive counter demonstrating Guix components, props, and template interpolation.

## Features

- Component definition with props
- Template string interpolation
- Input handling
- Component composition
- WebAssembly compilation

## Structure

- `counter.gx` - Guix source code defining Counter and App components
- `counter_gen.go` - Generated Go code (auto-generated, do not edit)
- `main.go` - Entry point for WASM application
- `index.html` - HTML page hosting the WASM app
- `counter.spec.js` - Playwright E2E tests

## Building

```bash
# Generate Go code from .gx files
../../guix generate -p .

# Copy wasm_exec.js from Go stdlib
cp $(go env GOROOT)/misc/wasm/wasm_exec.js .

# Build WebAssembly
GOOS=js GOARCH=wasm go build -o main.wasm .
```

Or use the build script:

```bash
./build.sh
```

## Running

Start a local server:

```bash
python3 -m http.server 8080
```

Open http://localhost:8080 in your browser.

## Testing

Install Playwright dependencies:

```bash
npm install
npx playwright install --with-deps chromium
```

Run tests:

```bash
npm test
```

Interactive test mode:

```bash
npm run test:ui
```

Debug mode:

```bash
npm run test:debug
```

## How it Works

1. **Component Definition** (`counter.gx`):
   ```go
   func Counter(count: int) {
       Div(Class("counter-display")) {
           Span(Class("counter-value")) {
               `Counter: {count}`
           }
       }
   }
   ```

2. **Code Generation**:
   - Generates `CounterProps` struct
   - Creates `WithCount()` option function
   - Builds `Render()` method with virtual DOM

3. **WASM Compilation**:
   - Compiles to WebAssembly for browser execution
   - Uses `syscall/js` for DOM interaction
   - Virtual DOM for efficient updates

4. **Browser Integration**:
   - Loads WASM module
   - Renders UI using generated components
   - Handles input events with JavaScript bridge
