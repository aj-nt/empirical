import { defineConfig } from '@playwright/test';

export default defineConfig({
  testDir: './e2e',
  timeout: 30000,
  retries: 1,
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
