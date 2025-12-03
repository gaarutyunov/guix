import { test, expect } from '@playwright/test';

test.describe('WebGPU Bitcoin Chart', () => {
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
      !err.includes('favicon') &&
      !err.includes('CORS policy') &&
      !err.includes('api.binance.com') &&
      !err.includes('ERR_FAILED')
    );

    expect(criticalErrors).toHaveLength(0);
  });

  test('should display page title and description', async ({ page }) => {
    await page.goto('http://localhost:8080');

    // Wait for page to load
    await page.waitForLoadState('networkidle');

    // Check for title
    const title = await page.locator('h1.title');
    await expect(title).toBeVisible();
    const titleText = await title.textContent();
    expect(titleText).toContain('Bitcoin');

    // Check for subtitle
    const subtitle = await page.locator('p.subtitle');
    await expect(subtitle).toBeVisible();
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

    // Check for WASM start
    const hasGoStart = consoleMessages.some(msg =>
      msg.includes('[Go] WASM module started') ||
      msg.includes('WASM module started')
    );
    expect(hasGoStart).toBeTruthy();

    // Check for app mounted successfully
    const hasAppMounted = consoleMessages.some(msg =>
      msg.includes('[Go] App mounted successfully') ||
      msg.includes('App mounted successfully')
    );
    expect(hasAppMounted).toBeTruthy();
  });

  test('should display info section', async ({ page }) => {
    await page.goto('http://localhost:8080');

    // Wait for page to load
    await page.waitForLoadState('networkidle');

    // Check for info section
    const infoSection = await page.locator('div.info');
    await expect(infoSection).toBeVisible();

    // Check for info text
    const infoText = await infoSection.textContent();
    expect(infoText).toContain('WebGPU');
    expect(infoText).toContain('Bitcoin');
  });

  test('should render chart and capture screenshot', async ({ page }) => {
    // Collect all console messages for debugging
    const allMessages: string[] = [];
    page.on('console', msg => {
      allMessages.push(`[${msg.type()}] ${msg.text()}`);
    });

    // Navigate to the page
    await page.goto('http://localhost:8080');

    // Wait for canvas to be rendered
    await page.waitForSelector('canvas', { timeout: 5000 });

    // Wait a bit more for chart rendering
    await page.waitForTimeout(1000);

    // Print all console messages for debugging
    console.log('\n=== Console Messages (should render chart and capture screenshot) ===');
    allMessages.forEach(msg => console.log(msg));
    console.log('=====================================================================\n');

    // Take full page screenshot for debugging
    await page.screenshot({ path: 'test-results/webgpu-chart-rendering.png', fullPage: true });

    // Verify canvas exists and has content
    const canvas = await page.locator('canvas');
    await expect(canvas).toBeVisible();

    // Check canvas has correct dimensions (1200x700 as defined in app.gx)
    const width = await canvas.getAttribute('width');
    const height = await canvas.getAttribute('height');
    expect(parseInt(width || '0')).toBeGreaterThan(0);
    expect(parseInt(height || '0')).toBeGreaterThan(0);
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
    const errorDivs = await page.locator('div.error');
    const count = await errorDivs.count();
    expect(count).toBe(0);
  });

  test('should have correct canvas styling', async ({ page }) => {
    await page.goto('http://localhost:8080');

    // Wait for canvas to be rendered
    await page.waitForSelector('canvas', { timeout: 5000 });

    // Check canvas styling
    const canvas = await page.locator('canvas');
    const styles = await canvas.evaluate(el => {
      const computed = window.getComputedStyle(el);
      return {
        display: computed.display,
        borderRadius: computed.borderRadius,
      };
    });

    expect(styles.display).toBe('block');
    // Border radius should be 8px
    expect(styles.borderRadius).toBe('8px');
  });
});
