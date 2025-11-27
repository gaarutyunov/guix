// @ts-check
const { test, expect } = require('@playwright/test');

test.describe('Calculator Example', () => {
  test.beforeEach(async ({ page }) => {
    // Log console messages for debugging
    page.on('console', msg => console.log('BROWSER:', msg.text()));
    page.on('pageerror', err => console.error('PAGE ERROR:', err));
  });

  test('should display initial calculator state', async ({ page }) => {
    await page.goto('/');

    // Wait for WASM to load
    await page.waitForSelector('.calculator', { timeout: 10000 });

    // Check that the page loaded
    const title = await page.textContent('h1');
    expect(title).toContain('Calculator');

    // Check initial display shows 0
    const display = await page.textContent('.display');
    expect(display).toBe('0');
  });

  test('should perform addition', async ({ page }) => {
    await page.goto('/');
    await page.waitForSelector('.calculator', { timeout: 10000 });

    // Click 3 + 2 =
    await page.click('button:has-text("3")');
    await page.click('button:has-text("+")');
    await page.click('button:has-text("2")');
    await page.click('button:has-text("=")');

    // Wait for update
    await page.waitForTimeout(300);

    const display = await page.textContent('.display');
    expect(display).toBe('5');
  });

  test('should perform subtraction', async ({ page }) => {
    await page.goto('/');
    await page.waitForSelector('.calculator', { timeout: 10000 });

    // Click 9 − 4 =
    await page.click('button:has-text("9")');
    await page.click('button:has-text("−")');
    await page.click('button:has-text("4")');
    await page.click('button:has-text("=")');

    await page.waitForTimeout(300);

    const display = await page.textContent('.display');
    expect(display).toBe('5');
  });

  test('should perform multiplication', async ({ page }) => {
    await page.goto('/');
    await page.waitForSelector('.calculator', { timeout: 10000 });

    // Click 6 × 7 =
    await page.click('button:has-text("6")');
    await page.click('button:has-text("×")');
    await page.click('button:has-text("7")');
    await page.click('button:has-text("=")');

    await page.waitForTimeout(300);

    const display = await page.textContent('.display');
    expect(display).toBe('42');
  });

  test('should perform division', async ({ page }) => {
    await page.goto('/');
    await page.waitForSelector('.calculator', { timeout: 10000 });

    // Click 8 ÷ 2 =
    await page.click('button:has-text("8")');
    await page.click('button:has-text("÷")');
    await page.click('button:has-text("2")');
    await page.click('button:has-text("=")');

    await page.waitForTimeout(300);

    const display = await page.textContent('.display');
    expect(display).toBe('4');
  });

  test('should clear display', async ({ page }) => {
    await page.goto('/');
    await page.waitForSelector('.calculator', { timeout: 10000 });

    // Enter some numbers
    await page.click('button:has-text("3")');
    await page.click('button:has-text("6")');

    // Click clear
    await page.click('button:has-text("C")');
    await page.waitForTimeout(300);

    const display = await page.textContent('.display');
    expect(display).toBe('0');
  });

  test('should handle sequential operations', async ({ page }) => {
    await page.goto('/');
    await page.waitForSelector('.calculator', { timeout: 10000 });

    // Test 3+2-3 (should show 5 after +, then 2 after -)
    await page.click('button:has-text("3")');
    await page.waitForTimeout(100);

    await page.click('button:has-text("+")');
    await page.waitForTimeout(100);

    await page.click('button:has-text("2")');
    await page.waitForTimeout(300);

    // After pressing 2, should show 5 (immediate calculation)
    let display = await page.textContent('.display');
    expect(display).toBe('5');

    await page.click('button:has-text("−")');
    await page.waitForTimeout(100);

    await page.click('button:has-text("3")');
    await page.waitForTimeout(300);

    // After pressing 3, should show 2 (5-3=2)
    display = await page.textContent('.display');
    expect(display).toBe('2');
  });

  test('should handle division by zero', async ({ page }) => {
    await page.goto('/');
    await page.waitForSelector('.calculator', { timeout: 10000 });

    // Click 3 ÷ 0 =
    await page.click('button:has-text("3")');
    await page.click('button:has-text("÷")');
    await page.click('button:has-text("0")');
    await page.click('button:has-text("=")');

    await page.waitForTimeout(300);

    const display = await page.textContent('.display');
    // Should handle gracefully (0 or Error are acceptable)
    expect(['0', 'Error']).toContain(display);
  });

  test('should display built with Guix message', async ({ page }) => {
    await page.goto('/');
    await page.waitForSelector('.calculator', { timeout: 10000 });

    const info = await page.textContent('.info');
    expect(info).toContain('Built with Guix');
  });
});
