import { test, expect } from '@playwright/test';

test.describe('WebGPU Rotating Cube', () => {
  test('should show initialization logs', async ({ page }) => {
    // Collect all console messages - set up BEFORE navigation
    const consoleMessages: string[] = [];
    page.on('console', msg => {
      consoleMessages.push(`[${msg.type()}] ${msg.text()}`);
    });

    // Navigate and reload to ensure fresh WASM
    await page.goto('http://localhost:8080');
    await page.reload({ waitUntil: 'networkidle' });

    // Wait for canvas to be rendered (indicates WebGPU initialization)
    await page.waitForSelector('canvas', { timeout: 5000 });

    // Wait a bit more for messages
    await page.waitForTimeout(500);

    // Print all console messages for debugging
    console.log('=== Console Messages ===');
    consoleMessages.forEach(msg => console.log(msg));
    console.log('========================');

    // Check for HTML initialization
    const hasHTMLCheck = consoleMessages.some(msg =>
      msg.includes('[HTML] Checking WebGPU support')
    );
    expect(hasHTMLCheck).toBeTruthy();

    // Check for WASM start
    const hasGoStart = consoleMessages.some(msg =>
      msg.includes('[Go] WASM module started')
    );
    expect(hasGoStart).toBeTruthy();

    // Check for app mounted successfully
    const hasAppMounted = consoleMessages.some(msg =>
      msg.includes('[Go] App mounted successfully')
    );
    expect(hasAppMounted).toBeTruthy();
  });

  test('should load without errors', async ({ page }) => {
    // Collect all console messages for debugging
    const allMessages: string[] = [];
    const errors: string[] = [];
    page.on('console', msg => {
      const msgText = `[${msg.type()}] ${msg.text()}`;
      allMessages.push(msgText);
      if (msg.type() === 'error') {
        errors.push(msg.text());
      }
    });

    // Navigate to the page
    await page.goto('http://localhost:8080');

    // Wait for the page to load
    await page.waitForLoadState('networkidle');

    // Wait a bit for any errors to appear
    await page.waitForTimeout(1000);

    // Print all console messages for debugging
    console.log('\n=== Console Messages (should load without errors) ===');
    allMessages.forEach(msg => console.log(msg));
    console.log('====================================================\n');

    // Filter out known acceptable errors
    const criticalErrors = errors.filter(err =>
      !err.includes('DevTools') &&
      !err.includes('favicon')
    );

    expect(criticalErrors).toHaveLength(0);
  });

  test('should display WebGPU support message', async ({ page }) => {
    // Collect all console messages for debugging
    const allMessages: string[] = [];
    const initMessages: string[] = [];
    page.on('console', msg => {
      const msgText = `[${msg.type()}] ${msg.text()}`;
      allMessages.push(msgText);
      if (msg.type() === 'log') {
        initMessages.push(msg.text());
      }
    });

    // Navigate to the page
    await page.goto('http://localhost:8080');

    // Wait for initialization
    await page.waitForTimeout(1000);

    // Print all console messages for debugging
    console.log('\n=== Console Messages (should display WebGPU support message) ===');
    allMessages.forEach(msg => console.log(msg));
    console.log('==============================================================\n');

    // Check for key initialization messages
    const hasWebGPUMessage = initMessages.some(msg =>
      msg.includes('WebGPU Rotating Cube Example')
    );
    const hasInitMessage = initMessages.some(msg =>
      msg.includes('Initializing WebGPU') ||
      msg.includes('WebGPU initialized successfully')
    );

    expect(hasWebGPUMessage).toBeTruthy();
    expect(hasInitMessage).toBeTruthy();
  });

  test('should render canvas element', async ({ page }) => {
    // Collect all console messages for debugging
    const allMessages: string[] = [];
    page.on('console', msg => {
      allMessages.push(`[${msg.type()}] ${msg.text()}`);
    });

    // Navigate to the page
    await page.goto('http://localhost:8080');

    // Wait for canvas to be created
    await page.waitForSelector('canvas', { timeout: 5000 });

    // Print all console messages for debugging
    console.log('\n=== Console Messages (should render canvas element) ===');
    allMessages.forEach(msg => console.log(msg));
    console.log('======================================================\n');

    // Check canvas dimensions
    const canvas = await page.locator('canvas');
    await expect(canvas).toBeVisible();

    const boundingBox = await canvas.boundingBox();
    expect(boundingBox).not.toBeNull();
    expect(boundingBox?.width).toBeGreaterThan(0);
    expect(boundingBox?.height).toBeGreaterThan(0);
  });

  test('should display control buttons', async ({ page }) => {
    // Collect all console messages for debugging
    const allMessages: string[] = [];
    page.on('console', msg => {
      allMessages.push(`[${msg.type()}] ${msg.text()}`);
    });

    // Navigate to the page
    await page.goto('http://localhost:8080');

    // Wait for canvas to be rendered
    await page.waitForSelector('canvas', { timeout: 5000 });

    // Wait for either controls or error to appear
    await page.waitForSelector('#controls, div[style*="background: #ff4444"]', { timeout: 5000 });

    // Print all console messages for debugging
    console.log('\n=== Console Messages (should display control buttons) ===');
    allMessages.forEach(msg => console.log(msg));
    console.log('=========================================================\n');

    // Check that no error occurred
    const errorDivs = await page.locator('div[style*="background: #ff4444"]');
    const errorCount = await errorDivs.count();
    expect(errorCount).toBe(0);

    // Check for arrow buttons
    const upButton = await page.locator('#btn-up');
    const downButton = await page.locator('#btn-down');
    const leftButton = await page.locator('#btn-left');
    const rightButton = await page.locator('#btn-right');
    const toggleButton = await page.locator('#btn-toggle');

    await expect(upButton).toBeVisible();
    await expect(downButton).toBeVisible();
    await expect(leftButton).toBeVisible();
    await expect(rightButton).toBeVisible();
    await expect(toggleButton).toBeVisible();
  });

  test('should respond to button clicks', async ({ page }) => {
    // Collect all console messages for debugging
    const allMessages: string[] = [];
    page.on('console', msg => {
      allMessages.push(`[${msg.type()}] ${msg.text()}`);
    });

    // Navigate to the page
    await page.goto('http://localhost:8080');

    // Wait for canvas to be rendered
    await page.waitForSelector('canvas', { timeout: 5000 });

    // Wait for either controls or error to appear
    await page.waitForSelector('#btn-toggle, div[style*="background: #ff4444"]', { timeout: 5000 });

    // Print all console messages for debugging
    console.log('\n=== Console Messages (should respond to button clicks) ===');
    allMessages.forEach(msg => console.log(msg));
    console.log('=========================================================\n');

    // Check that no error occurred
    const errorDivs = await page.locator('div[style*="background: #ff4444"]');
    const errorCount = await errorDivs.count();
    expect(errorCount).toBe(0);

    // Get initial button text
    const toggleButton = await page.locator('#btn-toggle');
    const initialText = await toggleButton.textContent();
    expect(initialText).toBe('⏸'); // Should be pause icon initially

    // Click toggle button
    await toggleButton.click();
    await page.waitForTimeout(100);

    // Check that button text changed
    const newText = await toggleButton.textContent();
    expect(newText).toBe('▶'); // Should be play icon now

    // Click again to toggle back
    await toggleButton.click();
    await page.waitForTimeout(100);

    const finalText = await toggleButton.textContent();
    expect(finalText).toBe('⏸'); // Should be pause icon again
  });

  test('should show speed control when auto-rotate is enabled', async ({ page }) => {
    // Collect all console messages for debugging
    const allMessages: string[] = [];
    page.on('console', msg => {
      allMessages.push(`[${msg.type()}] ${msg.text()}`);
    });

    // Navigate to the page
    await page.goto('http://localhost:8080');

    // Wait for canvas to be rendered
    await page.waitForSelector('canvas', { timeout: 5000 });

    // Wait for either controls or error to appear
    await page.waitForSelector('#speed-control, div[style*="background: #ff4444"]', { timeout: 5000 });

    // Print all console messages for debugging
    console.log('\n=== Console Messages (should show speed control when auto-rotate is enabled) ===');
    allMessages.forEach(msg => console.log(msg));
    console.log('===============================================================================\n');

    // Check that no error occurred
    const errorDivs = await page.locator('div[style*="background: #ff4444"]');
    const errorCount = await errorDivs.count();
    expect(errorCount).toBe(0);

    // Speed control should be visible initially (auto-rotate is on by default)
    const speedControl = await page.locator('#speed-control');
    await expect(speedControl).toBeVisible();

    // Toggle auto-rotate off
    const toggleButton = await page.locator('#btn-toggle');
    await toggleButton.click();
    await page.waitForTimeout(200);

    // Speed control should not be rendered (not in DOM) when auto-rotate is off
    const speedControlCount = await page.locator('#speed-control').count();
    expect(speedControlCount).toBe(0);
  });

  test('should handle keyboard controls', async ({ page }) => {
    // Collect all console messages for debugging
    const allMessages: string[] = [];
    page.on('console', msg => {
      allMessages.push(`[${msg.type()}] ${msg.text()}`);
    });

    // Navigate to the page
    await page.goto('http://localhost:8080');

    // Wait for canvas to be rendered
    await page.waitForSelector('canvas', { timeout: 5000 });

    // Wait for either controls or error to appear
    await page.waitForSelector('#btn-toggle, div[style*="background: #ff4444"]', { timeout: 5000 });

    // Print all console messages for debugging
    console.log('\n=== Console Messages (should handle keyboard controls) ===');
    allMessages.forEach(msg => console.log(msg));
    console.log('==========================================================\n');

    // Check that no error occurred
    const errorDivs = await page.locator('div[style*="background: #ff4444"]');
    const errorCount = await errorDivs.count();
    expect(errorCount).toBe(0);

    // Press space to toggle auto-rotation
    await page.keyboard.press('Space');
    await page.waitForTimeout(100);

    // Check that toggle button changed
    const toggleButton = await page.locator('#btn-toggle');
    const text = await toggleButton.textContent();
    expect(text).toBe('▶'); // Should be play icon after toggling

    // Press arrow keys (just verify no errors)
    await page.keyboard.press('ArrowLeft');
    await page.keyboard.press('ArrowRight');
    await page.keyboard.press('ArrowUp');
    await page.keyboard.press('ArrowDown');

    // Wait briefly and verify page still works
    await page.waitForTimeout(500);
    await expect(toggleButton).toBeVisible();
  });

  test('should render scene and capture screenshot', async ({ page }) => {
    // Collect all console messages for debugging
    const allMessages: string[] = [];
    page.on('console', msg => {
      allMessages.push(`[${msg.type()}] ${msg.text()}`);
    });

    // Navigate to the page
    await page.goto('http://localhost:8080');

    // Wait for canvas to be rendered
    await page.waitForSelector('canvas', { timeout: 5000 });

    // Wait a bit more for animation
    await page.waitForTimeout(1000);

    // Print all console messages for debugging
    console.log('\n=== Console Messages (should render scene and capture screenshot) ===');
    allMessages.forEach(msg => console.log(msg));
    console.log('====================================================================\n');

    // Take full page screenshot for debugging
    await page.screenshot({ path: 'test-results/webgpu-rendering.png', fullPage: true });

    // Verify canvas exists and has content
    const canvas = await page.locator('canvas');
    await expect(canvas).toBeVisible();

    // Check canvas has non-zero dimensions
    const boundingBox = await canvas.boundingBox();
    expect(boundingBox).not.toBeNull();
    expect(boundingBox?.width).toBeGreaterThan(0);
    expect(boundingBox?.height).toBeGreaterThan(0);
  });

  test('should not show error messages', async ({ page }) => {
    // Collect all console messages for debugging
    const allMessages: string[] = [];
    page.on('console', msg => {
      allMessages.push(`[${msg.type()}] ${msg.text()}`);
    });

    // Navigate to the page
    await page.goto('http://localhost:8080');

    // Wait for page to load
    await page.waitForTimeout(1000);

    // Print all console messages for debugging
    console.log('\n=== Console Messages (should not show error messages) ===');
    allMessages.forEach(msg => console.log(msg));
    console.log('=========================================================\n');

    // Check that no error divs are present
    const errorDivs = await page.locator('div[style*="background: #ff4444"]');
    const count = await errorDivs.count();
    expect(count).toBe(0);
  });
});
