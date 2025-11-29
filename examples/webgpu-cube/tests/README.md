# WebGPU Cube E2E Tests

End-to-end tests for the WebGPU rotating cube example using Playwright.

## Setup

```bash
# Install dependencies
npm install

# Install Playwright browsers
npx playwright install --with-deps chromium
```

## Running Tests

```bash
# Run all tests (headless)
npx playwright test

# Run tests in headed mode (visible browser)
npx playwright test --headed

# Run tests in debug mode (with Playwright Inspector)
npx playwright test --debug

# Run specific test file
npx playwright test webgpu-cube.spec.ts

# Run tests with UI mode
npx playwright test --ui
```

## Test Coverage

The test suite verifies:

1. **Initialization**
   - Page loads without critical errors
   - WebGPU support is detected
   - WebGPU context initializes successfully

2. **Rendering**
   - Canvas element is created
   - Canvas has correct dimensions
   - Scene renders without errors

3. **Controls**
   - All control buttons are visible
   - Toggle button changes state on click
   - Speed control visibility toggles correctly

4. **Keyboard Input**
   - Arrow keys work for rotation
   - Spacebar toggles auto-rotation

5. **Visual Regression**
   - Screenshot of rendered cube is captured
   - Visual differences are detected

## WebGPU in CI Environment

The tests are configured to run in GitHub Actions with headless Chrome using software rendering:

### Chrome Flags

The following flags are set in `playwright.config.ts`:

- `--enable-unsafe-webgpu` - Enable WebGPU API
- `--use-angle=swiftshader` - Use SwiftShader for software rendering
- `--disable-vulkan-surface` - Disable Vulkan (not available in CI)
- `--no-sandbox` - Required for containerized environments
- `--disable-gpu` - Disable hardware GPU acceleration

### Why Software Rendering?

GitHub Actions runners don't have physical GPUs, so we use SwiftShader (a CPU-based OpenGL/Vulkan implementation) to render WebGPU content in software. This is slower than hardware rendering but allows us to test WebGPU functionality in CI.

## Test Structure

```
tests/
├── README.md                    # This file
├── webgpu-cube.spec.ts         # Main test suite
└── webgpu-cube.spec.ts-snapshots/  # Screenshot baselines (auto-generated)
    └── webgpu-cube.png
```

## Writing Tests

Example test:

```typescript
import { test, expect } from '@playwright/test';

test('should display canvas', async ({ page }) => {
  await page.goto('http://localhost:8080');

  // Wait for canvas
  const canvas = await page.locator('canvas');
  await expect(canvas).toBeVisible();

  // Check dimensions
  const box = await canvas.boundingBox();
  expect(box?.width).toBeGreaterThan(0);
  expect(box?.height).toBeGreaterThan(0);
});
```

## Debugging

### Visual Debugging

```bash
# Run in headed mode to see the browser
npx playwright test --headed

# Open Playwright Inspector
npx playwright test --debug
```

### Capture Screenshots

```bash
# Update screenshot baselines
npx playwright test --update-snapshots
```

### Console Logs

The tests capture console messages. Check test output for:
- `console.log()` messages (initialization progress)
- `console.error()` messages (errors)

### Common Issues

**WebGPU not supported in test:**
- Ensure Chrome flags are set in `playwright.config.ts`
- Check that `--enable-unsafe-webgpu` is included
- Verify `--use-angle=swiftshader` for software rendering

**Tests timeout:**
- Increase timeout in test: `test.setTimeout(60000)`
- Check that HTTP server is running on port 8080

**Screenshot differences:**
- Software rendering may produce slightly different output
- Increase `maxDiffPixels` in screenshot comparison
- Update baselines if changes are intentional

## CI/CD Integration

Tests run automatically in GitHub Actions on:
- Push to `main` branch
- Pull requests to `main` branch

See `.github/workflows/e2e.yml` for the full CI configuration.

## Resources

- [Playwright Documentation](https://playwright.dev/docs/intro)
- [WebGPU Specification](https://www.w3.org/TR/webgpu/)
- [Chrome WebGPU Flags](https://peter.sh/experiments/chromium-command-line-switches/)
