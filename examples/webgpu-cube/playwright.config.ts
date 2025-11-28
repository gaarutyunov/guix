import { defineConfig, devices } from '@playwright/test';

/**
 * See https://playwright.dev/docs/test-configuration.
 */
export default defineConfig({
  testDir: './tests',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: 'html',
  use: {
    baseURL: 'http://localhost:8080',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
  },

  projects: [
    {
      name: 'chromium-webgpu',
      use: {
        ...devices['Desktop Chrome'],
        launchOptions: {
          args: [
            // Enable WebGPU
            '--enable-unsafe-webgpu',

            // Use software rendering (for CI without real GPU)
            '--use-angle=swiftshader',

            // Additional WebGPU flags for CI environment
            '--disable-vulkan-surface',
            '--disable-vulkan-fallback-to-gl-for-testing',
            '--disable-features=Vulkan',

            // Performance and stability flags
            '--no-sandbox',
            '--disable-setuid-sandbox',
            '--disable-dev-shm-usage',
            '--disable-accelerated-2d-canvas',
            '--disable-gpu',

            // Disable Chrome features that might interfere
            '--disable-extensions',
            '--disable-background-networking',
            '--disable-sync',
            '--disable-translate',
            '--disable-default-apps',

            // Memory and performance
            '--js-flags=--expose-gc',
            '--enable-precise-memory-info',

            // WebGPU-specific
            '--enable-features=Vulkan,UseSkiaRenderer',
          ],
        },
      },
    },
  ],

  webServer: {
    command: 'python3 -m http.server 8080',
    url: 'http://localhost:8080',
    reuseExistingServer: !process.env.CI,
    timeout: 120000,
  },
});
