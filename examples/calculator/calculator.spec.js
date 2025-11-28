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

  test('should display expression as typed - simple addition', async ({ page }) => {
    await page.goto('/');
    await page.waitForSelector('.calculator', { timeout: 10000 });

    // Click 2
    await page.click('button:has-text("2")');
    await page.waitForTimeout(100);
    let display = await page.textContent('.display');
    expect(display).toBe('2');

    // Click +
    await page.click('button:has-text("+")');
    await page.waitForTimeout(100);
    display = await page.textContent('.display');
    expect(display).toBe('2 +');

    // Click 3
    await page.click('button:has-text("3")');
    await page.waitForTimeout(100);
    display = await page.textContent('.display');
    expect(display).toBe('2 + 3');

    // Click =
    await page.click('button:has-text("=")');
    await page.waitForTimeout(300);
    display = await page.textContent('.display');
    expect(display).toBe('5');
  });

  test('should perform simple subtraction', async ({ page }) => {
    await page.goto('/');
    await page.waitForSelector('.calculator', { timeout: 10000 });

    await page.click('button:has-text("9")');
    await page.click('button:has-text("−")');
    await page.click('button:has-text("4")');
    await page.waitForTimeout(100);

    // Verify expression is displayed
    let display = await page.textContent('.display');
    expect(display).toBe('9 - 4');

    await page.click('button:has-text("=")');
    await page.waitForTimeout(300);

    display = await page.textContent('.display');
    expect(display).toBe('5');
  });

  test('should perform simple multiplication', async ({ page }) => {
    await page.goto('/');
    await page.waitForSelector('.calculator', { timeout: 10000 });

    await page.click('button:has-text("6")');
    await page.click('button:has-text("×")');
    await page.click('button:has-text("7")');

    // Verify expression is displayed
    await page.waitForTimeout(100);
    let display = await page.textContent('.display');
    expect(display).toBe('6 * 7');

    await page.click('button:has-text("=")');
    await page.waitForTimeout(300);

    display = await page.textContent('.display');
    expect(display).toBe('42');
  });

  test('should perform simple division', async ({ page }) => {
    await page.goto('/');
    await page.waitForSelector('.calculator', { timeout: 10000 });

    await page.click('button:has-text("8")');
    await page.click('button:has-text("÷")');
    await page.click('button:has-text("2")');

    // Verify expression is displayed
    await page.waitForTimeout(100);
    let display = await page.textContent('.display');
    expect(display).toBe('8 / 2');

    await page.click('button:has-text("=")');
    await page.waitForTimeout(300);

    display = await page.textContent('.display');
    expect(display).toBe('4');
  });

  test('should handle multi-digit numbers', async ({ page }) => {
    await page.goto('/');
    await page.waitForSelector('.calculator', { timeout: 10000 });

    // Enter 123
    await page.click('button:has-text("1")');
    await page.click('button:has-text("2")');
    await page.click('button:has-text("3")');
    await page.waitForTimeout(100);

    let display = await page.textContent('.display');
    expect(display).toBe('123');

    await page.click('button:has-text("+")');

    // Enter 456
    await page.click('button:has-text("4")');
    await page.click('button:has-text("5")');
    await page.click('button:has-text("6")');
    await page.waitForTimeout(100);

    display = await page.textContent('.display');
    expect(display).toBe('123 + 456');

    await page.click('button:has-text("=")');
    await page.waitForTimeout(300);

    display = await page.textContent('.display');
    expect(display).toBe('579');
  });

  test('should handle complex expression with multiple operations', async ({ page }) => {
    await page.goto('/');
    await page.waitForSelector('.calculator', { timeout: 10000 });

    // Calculate 10 + 5 - 3 + 2 (left-to-right = 14)
    await page.click('button:has-text("1")');
    await page.click('button:has-text("0")');
    await page.click('button:has-text("+")');
    await page.click('button:has-text("5")');
    await page.click('button:has-text("−")');
    await page.click('button:has-text("3")');
    await page.click('button:has-text("+")');
    await page.click('button:has-text("2")');

    await page.waitForTimeout(100);
    let display = await page.textContent('.display');
    expect(display).toBe('10 + 5 - 3 + 2');

    await page.click('button:has-text("=")');
    await page.waitForTimeout(300);

    display = await page.textContent('.display');
    expect(display).toBe('14');
  });

  test('should continue calculation after equals', async ({ page }) => {
    await page.goto('/');
    await page.waitForSelector('.calculator', { timeout: 10000 });

    // Calculate 5 + 3 = 8
    await page.click('button:has-text("5")');
    await page.click('button:has-text("+")');
    await page.click('button:has-text("3")');
    await page.click('button:has-text("=")');
    await page.waitForTimeout(300);

    let display = await page.textContent('.display');
    expect(display).toBe('8');

    // Continue: × 2 = 16
    await page.click('button:has-text("×")');
    await page.waitForTimeout(100);
    display = await page.textContent('.display');
    expect(display).toBe('8 *');

    await page.click('button:has-text("2")');
    await page.waitForTimeout(100);
    display = await page.textContent('.display');
    expect(display).toBe('8 * 2');

    await page.click('button:has-text("=")');
    await page.waitForTimeout(300);

    display = await page.textContent('.display');
    expect(display).toBe('16');
  });

  test('should start fresh after equals when pressing number', async ({ page }) => {
    await page.goto('/');
    await page.waitForSelector('.calculator', { timeout: 10000 });

    // Calculate 5 + 3 = 8
    await page.click('button:has-text("5")');
    await page.click('button:has-text("+")');
    await page.click('button:has-text("3")');
    await page.click('button:has-text("=")');
    await page.waitForTimeout(300);

    let display = await page.textContent('.display');
    expect(display).toBe('8');

    // Press a number - should start new calculation
    await page.click('button:has-text("9")');
    await page.waitForTimeout(100);
    display = await page.textContent('.display');
    expect(display).toBe('9');
  });

  test('should clear display', async ({ page }) => {
    await page.goto('/');
    await page.waitForSelector('.calculator', { timeout: 10000 });

    // Enter some expression
    await page.click('button:has-text("5")');
    await page.click('button:has-text("+")');
    await page.click('button:has-text("3")');
    await page.waitForTimeout(100);

    // Clear
    await page.click('button:has-text("C")');
    await page.waitForTimeout(300);

    const display = await page.textContent('.display');
    expect(display).toBe('0');
  });

  test('should handle division by zero', async ({ page }) => {
    await page.goto('/');
    await page.waitForSelector('.calculator', { timeout: 10000 });

    await page.click('button:has-text("5")');
    await page.click('button:has-text("÷")');
    await page.click('button:has-text("0")');
    await page.click('button:has-text("=")');

    await page.waitForTimeout(300);

    const display = await page.textContent('.display');
    expect(display).toBe('0');
  });

  test('should display built with Guix message', async ({ page }) => {
    await page.goto('/');
    await page.waitForSelector('.calculator', { timeout: 10000 });

    const info = await page.textContent('.info');
    expect(info).toContain('Built with Guix');
  });
});
