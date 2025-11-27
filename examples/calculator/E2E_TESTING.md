# Calculator E2E Tests

End-to-end tests for the Guix calculator example using headless Chrome.

## Prerequisites

1. **Chrome/Chromium** - The tests use chromedp which requires Chrome or Chromium to be installed
2. **Go 1.21+** - Required for running the tests
3. **Built WASM files** - Run the build script before testing

## Running E2E Tests

### Step 1: Build the Calculator

```bash
cd examples/calculator
bash build.sh
```

### Step 2: Run E2E Tests

```bash
# Run all e2e tests
go test -v -tags=e2e -timeout=60s

# Run a specific test
go test -v -tags=e2e -run TestCalculatorE2E/Basic

# Run with more detailed output
go test -v -tags=e2e -timeout=60s -count=1
```

## Test Coverage

The e2e tests cover:

1. **Basic Arithmetic Operations**
   - Addition (3+2=5)
   - Subtraction (9-4=5)
   - Multiplication (6×7=42)
   - Division (8÷2=4)

2. **Clear Functionality**
   - Verifies the Clear button resets display to 0

3. **Sequential Operations**
   - Tests the running calculator behavior:
     - User presses 3 → display shows 3
     - User presses + → display stays 3
     - User presses 2 → display shows 5 (immediate calculation)
     - User presses - → display stays 5
     - User presses 3 → display shows 2 (5-3=2)

4. **Division by Zero**
   - Tests edge case handling

## Test Architecture

The tests use:
- **chromedp**: Chrome DevTools Protocol for browser automation
- **Build tag `e2e`**: Separates e2e tests from unit tests
- **Local HTTP server**: Serves the calculator on port 8888 during tests
- **Headless Chrome**: Tests run without visible browser window

## Troubleshooting

### Chrome not found
If you get an error about Chrome not being found:
- **Linux**: Install `chromium-browser` or `google-chrome`
- **macOS**: Install Google Chrome
- **Windows**: Install Google Chrome

### Tests timeout
If tests timeout:
- Increase timeout: `go test -v -tags=e2e -timeout=120s`
- Check that WASM files are built correctly
- Ensure port 8888 is not in use

### WASM not loading
- Run `bash build.sh` to rebuild WASM files
- Check that `wasm_exec.js` exists in the calculator directory
- Verify `calculator.wasm` was generated

## CI/CD Integration

To run e2e tests in CI:

```yaml
# Example GitHub Actions workflow
- name: Install Chrome
  run: |
    wget -q -O - https://dl-ssl.google.com/linux/linux_signing_key.pub | sudo apt-key add -
    sudo sh -c 'echo "deb [arch=amd64] http://dl.google.com/linux/chrome/deb/ stable main" >> /etc/apt/sources.list.d/google.list'
    sudo apt-get update
    sudo apt-get install -y google-chrome-stable

- name: Run E2E Tests
  run: |
    cd examples/calculator
    bash build.sh
    go test -v -tags=e2e -timeout=120s
```

## Writing New E2E Tests

To add new test cases:

1. Add a new test function following the pattern:
```go
func testMyFeature(t *testing.T, ctx context.Context) {
    var display string

    err := chromedp.Run(ctx,
        chromedp.Navigate(testURL),
        waitForWASM(),
        // Your test actions...
        getDisplayValue(&display),
    )

    // Assertions...
}
```

2. Add the test to `TestCalculatorE2E`:
```go
t.Run("My feature", func(t *testing.T) {
    testMyFeature(t, ctx)
})
```

## Button Selectors

The tests use CSS selectors to click buttons. Button layout:
- Row 1: 7, 8, 9, ÷
- Row 2: 4, 5, 6, ×
- Row 3: 1, 2, 3, -
- Row 4: 0, ., C, =, +

Selectors are in the format `button:nth-of-type(N)` where N is the button number.
