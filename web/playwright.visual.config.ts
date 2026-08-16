import { defineConfig } from '@playwright/test';

/**
 * Playwright config for visual snapshot tests.
 * Separate from the main playwright.config.ts so visual tests
 * can have different settings (viewport, timeout, snapshot dir).
 */
export default defineConfig({
  testDir: './e2e/visual',
  timeout: 30000,
  retries: 0,
  snapshotDir: './e2e/visual/screenshots',
  expect: {
    toHaveScreenshot: {
      maxDiffPixels: 100,
      threshold: 0.2,
    },
  },
  use: {
    baseURL: 'http://localhost:5000',
    headless: true,
    viewport: { width: 1280, height: 800 },
  },
  webServer: {
    command: '/tmp/empirical serve 5000',
    url: 'http://localhost:5000',
    reuseExistingServer: true,
    timeout: 10000,
  },
});
