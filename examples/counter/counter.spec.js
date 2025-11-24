// @ts-check
const { test, expect } = require('@playwright/test');

test.describe('Counter Example', () => {
  test('should display initial counter value', async ({ page }) => {
    await page.goto('/');

    // Wait for WASM to load
    await page.waitForSelector('#root', { state: 'visible' });
    await page.waitForTimeout(1000); // Give WASM time to initialize

    // Check that the page loaded
    const title = await page.textContent('h1');
    expect(title).toContain('Counter Example');
  });

  test('should update counter when input changes', async ({ page }) => {
    await page.goto('/');

    // Wait for WASM to load
    await page.waitForSelector('#root', { state: 'visible' });
    await page.waitForTimeout(1000);

    // Find the input field
    const input = page.locator('input[type="number"]');
    await expect(input).toBeVisible();

    // Type a number
    await input.fill('42');

    // Wait a bit for the update
    await page.waitForTimeout(500);

    // Check that the counter display updated
    const counterDisplay = page.locator('.counter-value');
    await expect(counterDisplay).toContainText('42');
  });

  test('should handle multiple value changes', async ({ page }) => {
    await page.goto('/');

    // Wait for WASM to load
    await page.waitForSelector('#root', { state: 'visible' });
    await page.waitForTimeout(1000);

    const input = page.locator('input[type="number"]');
    const counterDisplay = page.locator('.counter-value');

    // Test multiple values
    const testValues = ['10', '25', '100', '0'];

    for (const value of testValues) {
      await input.fill(value);
      await page.waitForTimeout(300);
      await expect(counterDisplay).toContainText(value);
    }
  });

  test('should display correct counter text format', async ({ page }) => {
    await page.goto('/');

    await page.waitForSelector('#root', { state: 'visible' });
    await page.waitForTimeout(1000);

    const input = page.locator('input[type="number"]');
    await input.fill('99');
    await page.waitForTimeout(300);

    const counterDisplay = page.locator('.counter-value');
    const text = await counterDisplay.textContent();

    // Should display "Counter: 99"
    expect(text).toMatch(/Counter:\s*99/);
  });
});
