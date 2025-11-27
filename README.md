# Guix

**Guix is Go for the UI** - A complete Go-based UI language that transpiles to WebAssembly, featuring reactive components, virtual DOM, and type-safe event handling.

## Features

- 🚀 **Pure Go** - Write UI components in Go-like syntax
- ⚡ **Virtual DOM** - Efficient diffing and patching with keyed reconciliation
- 🔄 **Reactive** - Channel-based state management for real-time updates
- 📦 **Type-Safe** - Full Go type system with compile-time checks
- 🎯 **WebAssembly** - Compiles to WASM for native browser performance
- 🛠️ **Developer Tools** - Watch mode, incremental compilation, and hot reload
- 🎨 **Template Interpolation** - Backtick strings with expression interpolation

## Quick Start

### Installation

```bash
go install github.com/gaarutyunov/guix/cmd/guix@latest
```

### Create Your First Component

Create a file `hello.gx`:

```go
package main

func Hello(name: string) {
    Div {
        H1 {
            `Hello, {name}!`
        }
    }
}
```

### Generate Go Code

```bash
guix generate -p .
```

This creates `hello_gen.go` with the compiled component.

### Build for WebAssembly

```bash
# Generate components
guix generate -p .

# Copy wasm_exec.js
cp $(go env GOROOT)/misc/wasm/wasm_exec.js .

# Build WASM
GOOS=js GOARCH=wasm go build -o main.wasm .
```

## Language Syntax

### Components

Components are the building blocks of Guix applications:

```go
func Button(label: string, onClick: func(Event)) {
    Button(OnClick(onClick)) {
        Text(label)
    }
}
```

### Props and Options

Components automatically generate props structs and option functions:

```go
// Generated code:
type ButtonProps struct {
    Label   string
    OnClick func(Event)
}

type ButtonOption func(*Button)

func WithLabel(v string) ButtonOption { ... }
func WithOnClick(v func(Event)) ButtonOption { ... }

// Usage:
btn := NewButton(
    WithLabel("Click Me"),
    WithOnClick(handleClick),
)
```

### Template Interpolation

Use backticks for template strings with embedded expressions:

```go
func Counter(count: int) {
    Div {
        `Count: {count}`
    }
}
```

### Channel-Based State

Channels enable reactive, real-time updates:

```go
import "strconv"

func Counter(counterChannel: chan int) {
    Div(Class("counter-display")) {
        Span(Class("counter-value")) {
            `Counter: {<-counterChannel}`
        }
    }
}

func App() {
    counter := make(chan int, 10)

    Div(Class("app-container")) {
        H1 {
            "Counter Example"
        }
        Counter(WithCounterChannel(counter))
        Div(Class("input-group")) {
            Input(
                Type("number"),
                Placeholder("Enter a number"),
                OnInput(func(e: Event) {
                    value := e.Target.Value
                    n, _ := strconv.Atoi(value)
                    counter <- n
                })
            )
        }
    }
}
```

### Event Handlers

Type-safe event handling with Go functions:

```go
func Form() {
    Div {
        Input(
            OnInput(func(e: Event) {
                value := e.Target.Value
                // Handle input
            }),
            Placeholder("Enter text...")
        )

        Button(
            OnClick(func(e: Event) {
                e.Native.Call("preventDefault")
                // Handle click
            })
        ) {
            Text("Submit")
        }
    }
}
```

### Element Builders

Common HTML elements with type-safe APIs:

```go
Div(Class("container")) {
    H1 {
        Text("Title")
    }

    P(Style("color: blue;")) {
        Text("Paragraph text")
    }

    Button(
        ID("submit-btn"),
        OnClick(handleSubmit)
    ) {
        Text("Submit")
    }

    Img(
        Src("/image.png"),
        Attr{Key: "alt", Value: "Description"}
    )
}
```

## CLI Commands

### Generate

Generate Go code from `.gx` files:

```bash
# Generate all .gx files in current directory
guix generate

# Specify path
guix generate -p ./components

# Watch mode - auto-regenerate on changes
guix generate -w

# Lazy mode - only regenerate changed files
guix generate --lazy

# Verbose output
guix generate --verbose
```

### Clean

Remove generated files and cache:

```bash
guix clean
guix clean -p ./components
```

## Architecture

### Runtime Library

The runtime (`pkg/runtime`) provides:

- **VNode**: Virtual DOM node structure
- **Diffing**: Efficient tree comparison with keyed reconciliation
- **DOM Manipulation**: syscall/js wrappers for DOM operations
- **Scheduler**: requestAnimationFrame batching for performance
- **Event System**: Memory-safe event handler management

### Code Generator

The code generator (`pkg/codegen`) uses Go's AST package to:

1. Parse `.gx` files with participle
2. Generate Props structs for type safety
3. Create option functions for ergonomic APIs
4. Build Render methods with virtual DOM
5. Output formatted Go source code

### Parser

The parser (`pkg/parser`) implements:

- Stateful lexer for template interpolation
- LL(k) recursive descent parsing with participle
- Support for Go-like expressions and statements
- Channel type syntax (`<-chan T`, `chan T`)

## Tech Stack

Guix is built with the following core dependencies:

### Parser & Code Generation
- **[participle/v2](https://github.com/alecthomas/participle)** - LL(k) parser for the `.gx` syntax
- **go/ast** - Go's AST package for code generation
- **go/format** - Automatic formatting of generated Go code

### CLI & Tooling
- **[urfave/cli/v2](https://github.com/urfave/cli)** - Command-line interface framework
- **[fsnotify](https://github.com/fsnotify/fsnotify)** - File system notifications for watch mode

### Runtime
- **syscall/js** - WebAssembly JavaScript interop
- **sync** - Goroutine synchronization primitives
- **context** - Cancellation and timeout handling

### Development Tools
- Standard Go toolchain (go 1.24+)
- WebAssembly target support (`GOOS=js GOARCH=wasm`)
- Optional: TinyGo for smaller binary sizes

## Examples

### Counter

See `examples/counter/` for a complete counter application demonstrating:

- Component composition with the `Counter` component
- Reactive channel-based state management
- Event handling with `OnInput`
- Template interpolation with backtick strings
- Type-safe props and option functions

The example consists of two components in `.gx` files:

**`counter.gx`** - Display component that renders channel values:
```go
func Counter(counterChannel: chan int) {
    Div(Class("counter-display")) {
        Span(Class("counter-value")) {
            `Counter: {<-counterChannel}`
        }
    }
}
```

**`app.gx`** - Main app component with input handling:
```go
import "strconv"

func App() {
    counter := make(chan int, 10)

    Div(Class("app-container")) {
        H1 {
            "Counter Example"
        }
        Counter(WithCounterChannel(counter))
        Div(Class("input-group")) {
            Input(
                Type("number"),
                Placeholder("Enter a number"),
                ID("counter-input"),
                OnInput(func(e: Event) {
                    value := e.Target.Value
                    n, _ := strconv.Atoi(value)
                    counter <- n
                })
            )
        }
    }
}
```

To run:

```bash
cd examples/counter
./build.sh
python3 -m http.server 8080
# Open http://localhost:8080
```

## Binary Size Optimization

### Standard Go (~2-10MB)

```bash
GOOS=js GOARCH=wasm go build -o main.wasm
```

### TinyGo (~50-400KB)

```bash
tinygo build -o main.wasm -target wasm .
```

**Note**: TinyGo has some limitations (no `recover()`, cooperative goroutines).

## Performance

- **Virtual DOM diffing**: O(n) time complexity
- **Keyed reconciliation**: Efficient list updates
- **Batched updates**: requestAnimationFrame scheduling
- **Memory management**: Automatic cleanup of event handlers
- **Zero allocations**: Reusable patch buffers (coming soon)

## Development

### Building Guix

```bash
# Install dependencies
go mod download

# Run tests
go test ./...

# Build CLI
go build -o guix ./cmd/guix

# Install locally
go install ./cmd/guix
```

### Project Structure

```
guix/
├── cmd/guix/          # CLI tool
├── pkg/
│   ├── ast/           # Guix AST definitions
│   ├── parser/        # Participle-based parser
│   ├── codegen/       # Go AST code generator
│   └── runtime/       # Virtual DOM runtime
├── internal/cache/    # Incremental compilation
└── examples/          # Example applications
```

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

MIT License - see [LICENSE](LICENSE) for details.

## Inspiration

Guix draws inspiration from:

- **Vugu** - Go virtual DOM and instruction buffer patterns
- **Vecty** - Go WebAssembly component architecture
- **templ** - Template parsing and code generation techniques
- **React** - Component model and virtual DOM concepts
- **Elm** - Functional architecture and type safety
- **Yew** - Rust WebAssembly framework with `html!` macro for declarative UI
- **Dioxus** - Rust UI framework with `rsx!` macro for component composition

## Roadmap

- [ ] Component lifecycle hooks (onMount, onUnmount)
- [ ] CSS-in-Go styling system
- [ ] Server-side rendering (SSR)
- [ ] Dev server with hot reload
- [ ] Component testing utilities
- [ ] Browser DevTools integration
- [ ] Keyed fragments optimization
- [ ] Async component loading
- [ ] Web Components compatibility
- [ ] Progressive Web App (PWA) support

## FAQ

### Why Go for UI?

Go provides strong typing, excellent tooling, and native WebAssembly support. Guix lets you use Go's ecosystem and type safety for building web UIs.

### How does it compare to Vugu?

Guix uses a custom DSL (`.gx` files) with template interpolation, while Vugu uses Go files with HTML comments. Guix also provides automatic option function generation and a more React-like API.

### Can I use existing Go packages?

Yes! Guix generates standard Go code, so you can import and use any Go package that works with WebAssembly.

### What about browser compatibility?

Guix requires WebAssembly support, which is available in all modern browsers (Chrome 57+, Firefox 52+, Safari 11+, Edge 16+).

## Resources

- [WebAssembly Documentation](https://webassembly.org/)
- [Go WebAssembly Wiki](https://github.com/golang/go/wiki/WebAssembly)
- [syscall/js Package](https://pkg.go.dev/syscall/js)
- [Participle Parser](https://github.com/alecthomas/participle)

## Support

- 🐛 [Report Issues](https://github.com/gaarutyunov/guix/issues)
- 💬 [Discussions](https://github.com/gaarutyunov/guix/discussions)
- 📧 Email: support@guix.dev

---

**Guix** - Making Go a first-class language for web UI development.
