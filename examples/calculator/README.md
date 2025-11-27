# Calculator Example

A fully functional calculator built entirely in Guix, demonstrating state management, event handling, and efficient UI updates.

## Features

- **Complete Calculator Implementation**: Basic arithmetic operations (+, -, ×, ÷)
- **Efficient State Management**: Display only updates when necessary
- **Running Calculations**: Operations are evaluated as you type
- **Channel-based Reactivity**: Uses Go channels for state updates
- **Beautiful UI**: Modern, responsive design with smooth animations

## Calculator Behavior

The calculator follows a running calculation model where operations are evaluated immediately:

1. **User presses 3** → Display shows `3`
2. **User presses +** → Display stays `3`, plus operator is active
3. **User presses 2** → Display shows `5` (3 + 2)
4. **User presses -** → Display stays `5`, minus operator is active
5. **User presses 3** → Display shows `2` (5 - 3)

This provides immediate feedback and natural calculation flow.

## Structure

- `calculator.gx` - Calculator component with state management and button handlers
- `app.gx` - App component that initializes the calculator state
- `calculator_gen.go` - Generated Go code (auto-generated, do not edit)
- `app_gen.go` - Generated Go code (auto-generated, do not edit)
- `main.go` - WASM entry point
- `index.html` - HTML page with styling
- `calculator.go` - Package documentation

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

## How it Works

### State Management

The calculator uses a `CalculatorState` struct to track:

```go
type CalculatorState struct {
    Display          string   // Current display value
    PreviousValue    float64  // Left operand
    Operator         string   // Active operator (+, -, *, /)
    WaitingForOperand bool    // Whether expecting new number
}
```

### Event Handlers

- **handleNumber**: Appends digits or starts new number
- **handleOperator**: Evaluates pending operation and sets new operator
- **handleEquals**: Completes current calculation
- **handleClear**: Resets calculator to initial state

### Efficient Updates

The display only changes when:
- A new digit is entered
- An operation result is calculated
- The calculator is cleared

Operator presses don't change the display, providing a smooth user experience.

### Code Generation

Guix generates:
1. Props struct for the Calculator component with `State` field
2. `WithState()` option function for passing channels
3. `Render()` method that creates virtual DOM nodes
4. Event handler bindings for button clicks

### Channel Reactivity

State updates flow through Go channels:
1. User clicks button → handler reads current state
2. Handler computes new state based on input
3. Handler sends new state to channel
4. Component re-renders with updated display

## Implementation Highlights

### Running Calculator Logic

Unlike simple calculators that wait for equals, this calculator evaluates operations immediately when the next operator is pressed, creating a natural flow.

### Number Formatting

Results are formatted to remove unnecessary decimal places:
- `5.0` displays as `5`
- `2.5` displays as `2.5`
- `0.333333333` displays with appropriate precision

### Error Handling

Division by zero returns `0` to prevent errors and maintain calculator state.

## Learn More

This example demonstrates:
- Complex state management in Guix
- Channel-based reactive updates
- Event handling with closures
- Component composition
- Efficient UI updates (only when necessary)
- WebAssembly compilation and browser integration
