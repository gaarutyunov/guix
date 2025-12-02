# Claude Development Guidelines for Guix

This document contains guidelines and rules for Claude when working on the Guix codebase.

## Pre-Commit Requirements

Before committing any changes, Claude must ensure the following checks pass.

### Using Mage (Recommended)

The easiest way to run all pre-commit checks is using Mage:

```bash
# Run all pre-commit checks (format, vet, test, build)
go run github.com/magefile/mage@latest

# Or if mage is installed locally
mage
```

This will automatically run all the checks below in the correct order.

### Manual Pre-Commit Checks

If you prefer to run checks manually:

#### 1. Code Formatting

All Go files must be formatted with `gofmt`:

```bash
gofmt -w .
```

This ensures consistent code style across the entire codebase.

#### 2. Linting

Run `go vet` to catch common Go mistakes:

```bash
go vet ./...
```

#### 3. Testing

Run all tests to ensure no regressions:

```bash
go test ./...
```

All tests must pass before committing.

#### 4. Build Verification

Verify the code compiles for both native and WASM targets:

```bash
# Native build
go build ./...

# WASM build (for runtime package)
GOOS=js GOARCH=wasm go build ./pkg/runtime/...
```

## Available Mage Targets

The following Mage targets are available:

- `mage` or `mage preCommit` - Run all pre-commit checks (default)
- `mage format` - Format all Go files
- `mage vet` - Run go vet
- `mage test` - Run all tests
- `mage build` - Build all packages
- `mage buildWasm` - Build WASM runtime
- `mage generate` - Regenerate all example code
- `mage clean` - Remove build artifacts
- `mage ci` - Run all CI checks

## Complete Pre-Commit Checklist (Manual)

If not using Mage, run these commands before every commit:

```bash
# Format all Go files
gofmt -w .

# Check for common mistakes
go vet ./...

# Run tests
go test ./...

# Verify native build
go build ./...

# Verify WASM build
GOOS=js GOARCH=wasm go build ./pkg/runtime/...
```

## Git Workflow

### Commit Message Format

Commit messages should follow this format:

```
<type>: <subject>

<body>

<footer>
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, missing semicolons, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

**Example:**
```
feat: Add WebGPU support for 3D graphics

Implement comprehensive WebGPU runtime with scene graphs,
PBR materials, and lighting system.

- Add 8 new runtime modules for GPU operations
- Create rotating cube example
- Add detailed documentation in docs/WEBGPU.md
```

### Branch Naming

Feature branches should follow the pattern: `claude/<feature-name>-<session-id>`

Example: `claude/add-webgpu-support-01BWtA2VBNZfg7Y29aanCzJG`

### Before Pushing

Always verify the build passes on CI/CD before pushing:

1. Format code: `gofmt -w .`
2. Run linter: `go vet ./...`
3. Run tests: `go test ./...`
4. Build all targets: `go build ./...`
5. Commit changes
6. Push to remote

## Code Style Guidelines

### General Go Style

Follow the [Effective Go](https://go.dev/doc/effective_go) guidelines:

- Use `gofmt` for formatting (enforced pre-commit)
- Keep functions small and focused
- Use meaningful variable names
- Comment exported functions and types
- Avoid global mutable state

### WASM-Specific Code

For code in `pkg/runtime/`:

- Always use build tags: `//go:build js && wasm`
- Handle `js.Value` carefully - check `Truthy()` before use
- Clean up `js.Func` with `Release()` to prevent memory leaks
- Use lowercase function names for internal helpers (e.g., `logError`, not `LogError`)

### WebGPU Code

For WebGPU-related code:

- Validate GPU context before operations
- Handle async promises with proper error checking
- Log errors using `logError()` from `dom.go`
- Clean up GPU resources in cleanup/unmount methods
- Test in WebGPU-enabled browsers (Chrome 113+, Edge 113+)

## Testing Guidelines

### Unit Tests

- Place tests in `*_test.go` files
- Use table-driven tests where appropriate
- Mock external dependencies
- Test error cases, not just happy paths

### Integration Tests

- Test WASM builds with browser automation (when available)
- Verify WebGPU functionality in supported browsers
- Test example applications end-to-end

### End-to-End (E2E) Tests

End-to-end tests use Playwright to test examples in real browsers:

```bash
# Navigate to example directory
cd examples/webgpu-cube

# Install dependencies
npm install

# Install Playwright browsers
npx playwright install --with-deps chromium

# Run tests
npx playwright test

# Run tests in headed mode (visible browser)
npx playwright test --headed

# Run tests in debug mode
npx playwright test --debug
```

**WebGPU E2E Testing Requirements:**

For WebGPU tests to run in CI environments (GitHub Actions), Chrome must be launched with specific flags:

- `--enable-unsafe-webgpu` - Enable WebGPU API
- `--use-angle=swiftshader` - Use software rendering (no GPU required)
- `--disable-vulkan-surface` - Disable Vulkan for CI compatibility
- `--no-sandbox` - Required for containerized environments

These flags are configured in `playwright.config.ts` for each example.

**Writing E2E Tests:**

```typescript
import { test, expect } from '@playwright/test';

test('should render canvas', async ({ page }) => {
  await page.goto('http://localhost:8080');

  // Wait for canvas to be created
  await page.waitForSelector('canvas', { timeout: 5000 });

  // Check canvas is visible
  const canvas = await page.locator('canvas');
  await expect(canvas).toBeVisible();
});
```

**Test Coverage:**

E2E tests should verify:
- Page loads without errors
- WebGPU initialization succeeds
- Canvas element is created and visible
- Controls respond to user interaction
- Keyboard shortcuts work correctly
- No error messages are displayed

## Documentation

### Code Documentation

- Add godoc comments to all exported types and functions
- Include examples in documentation where helpful
- Keep comments up-to-date with code changes

### User Documentation

- Update README.md for user-facing changes
- Add detailed guides to `docs/` for major features
- Include working examples in `examples/`
- Document browser compatibility requirements

## Common Pitfalls to Avoid

### Go/WASM Issues

1. **Unused imports**: Remove all unused imports (build will fail)
2. **Name conflicts**: Can't have type and function with same name in package
3. **Case sensitivity**: Exported names start with uppercase, internal with lowercase
4. **js.Value lifecycle**: Always check `Truthy()` and release `js.Func`

### WebGPU Issues

1. **Async initialization**: WebGPU requires async setup via promises
2. **Context configuration**: Canvas context must be configured before use
3. **Resource cleanup**: Always destroy GPU resources (buffers, textures, etc.)
4. **Browser support**: Check WebGPU availability before use

## CI/CD Expectations

The continuous integration system will run:

1. `gofmt -d .` (check formatting)
2. `go vet ./...` (linting)
3. `go test ./...` (unit tests)
4. `go build ./...` (build verification)
5. `GOOS=js GOARCH=wasm go build ./pkg/runtime/...` (WASM build)
6. Playwright E2E tests for all examples (counter, calculator, webgpu-cube)

All checks must pass for the build to succeed.

**WebGPU E2E Tests in CI:**

The WebGPU cube example runs E2E tests with:
- Headless Chrome with software rendering (SwiftShader)
- WebGPU enabled via `--enable-unsafe-webgpu` flag
- Tests verify initialization, rendering, and user interactions
- Screenshots are captured for visual regression testing

## Error Recovery

If the build fails:

1. Read the error message carefully
2. Fix the specific issue (formatting, imports, syntax)
3. Re-run the pre-commit checklist
4. Commit the fix with a descriptive message
5. Push the corrected code

## Resources

- [Go Documentation](https://go.dev/doc/)
- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [WebGPU Specification](https://www.w3.org/TR/webgpu/)
- [syscall/js Package](https://pkg.go.dev/syscall/js)
- [Playwright Documentation](https://playwright.dev/docs/intro)
- [Playwright for E2E Testing](https://playwright.dev/docs/writing-tests)

---

**Last Updated**: 2025-11-28
**Version**: 1.0
