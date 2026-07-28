import { test, expect } from '@playwright/test';

test.describe('Empirical GUI smoke test', () => {
  test('loads the app', async ({ page }) => {
    await page.goto('/');
    await expect(page.locator('h1')).toContainText('Empirical');
  });

  test('creates a chart and views the wheel', async ({ page }) => {
    await page.goto('/');

    // Click "+ New Chart"
    await page.click('button:has-text("+ New Chart")');

    // Fill birth data
    await page.fill('input[placeholder="e.g. AJ"]', 'E2E Test');
    await page.fill('input[type="number"]', '1990'); // year — first number input
    // Month, day, hour, minute, tz — use nth selectors
    const numberInputs = page.locator('input[type="number"]');
    await numberInputs.nth(0).fill('1990'); // year
    await numberInputs.nth(1).fill('6');    // month
    await numberInputs.nth(2).fill('15');   // day
    await numberInputs.nth(3).fill('8');    // hour
    await numberInputs.nth(4).fill('30');   // minute
    await numberInputs.nth(5).fill('-4');   // tz_offset

    // Save
    await page.click('button:has-text("Save Chart")');

    // Should see the wheel tab active
    await expect(page.locator('button:has-text("Wheel")')).toBeVisible();
    // SVG should be rendered
    await expect(page.locator('svg')).toBeVisible({ timeout: 10000 });
  });

  test('navigates through all tabs', async ({ page }) => {
    await page.goto('/');

    // Create a chart first
    await page.click('button:has-text("+ New Chart")');
    const numberInputs = page.locator('input[type="number"]');
    await numberInputs.nth(0).fill('1990');
    await numberInputs.nth(1).fill('6');
    await numberInputs.nth(2).fill('15');
    await numberInputs.nth(3).fill('8');
    await numberInputs.nth(4).fill('30');
    await numberInputs.nth(5).fill('-4');
    await page.fill('input[placeholder="e.g. AJ"]', 'Tab Test');
    await page.click('button:has-text("Save Chart")');

    // Wait for wheel
    await expect(page.locator('svg')).toBeVisible({ timeout: 10000 });

    // Click each tab and verify the button stays visible (no crash)
    const tabs = ['Natal', 'Transits', 'Synastry', 'Maps', 'Reports', 'Ephemeris', 'Calendar', 'Research'];
    for (const tab of tabs) {
      await page.click(`button:has-text("${tab}")`);
      await page.waitForTimeout(300);
      // The tab button should still be visible (app didn't crash)
      await expect(page.locator(`button:has-text("${tab}")`)).toBeVisible();
    }
  });

  test('theme switcher works', async ({ page }) => {
    await page.goto('/');

    // Theme buttons should be visible in sidebar
    await expect(page.locator('button[title="Dark"]')).toBeVisible();
    await expect(page.locator('button[title="Light"]')).toBeVisible();
    await expect(page.locator('button[title="Sepia"]')).toBeVisible();
    await expect(page.locator('button[title="High Contrast"]')).toBeVisible();

    // Click light theme
    await page.click('button[title="Light"]');
    // Verify data-theme attribute changed
    const theme = await page.evaluate(() =>
      document.documentElement.getAttribute('data-theme'),
    );
    expect(theme).toBe('light');
  });

  test('export buttons appear on wheel tab', async ({ page }) => {
    await page.goto('/');

    // Create chart
    await page.click('button:has-text("+ New Chart")');
    const numberInputs = page.locator('input[type="number"]');
    await numberInputs.nth(0).fill('1990');
    await numberInputs.nth(1).fill('6');
    await numberInputs.nth(2).fill('15');
    await numberInputs.nth(3).fill('8');
    await numberInputs.nth(4).fill('30');
    await numberInputs.nth(5).fill('-4');
    await page.fill('input[placeholder="e.g. AJ"]', 'Export Test');
    await page.click('button:has-text("Save Chart")');

    // Wait for wheel
    await expect(page.locator('svg')).toBeVisible({ timeout: 10000 });

    // Export buttons should be visible
    await expect(page.locator('button:has-text("Export SVG")')).toBeVisible();
    await expect(page.locator('button:has-text("Export PNG")')).toBeVisible();
  });

  test('page designer loads and shows blocks', async ({ page }) => {
    await page.goto('/');

    // Create chart
    await page.click('button:has-text("+ New Chart")');
    const numberInputs = page.locator('input[type="number"]');
    await numberInputs.nth(0).fill('1990');
    await numberInputs.nth(1).fill('6');
    await numberInputs.nth(2).fill('15');
    await numberInputs.nth(3).fill('8');
    await numberInputs.nth(4).fill('30');
    await numberInputs.nth(5).fill('-4');
    await page.fill('input[placeholder="e.g. AJ"]', 'Designer Test');
    await page.click('button:has-text("Save Chart")');

    // Wait for wheel
    await expect(page.locator('svg')).toBeVisible({ timeout: 10000 });

    // Navigate to Reports → Page Designer
    await page.click('button:has-text("Reports")');
    await page.click('button:has-text("Page Designer")');

    // Should show templates and blocks
    await expect(page.locator('text=Templates')).toBeVisible({ timeout: 10000 });
    await expect(page.locator('text=Add Block')).toBeVisible();
    await expect(page.locator('text=Layout')).toBeVisible();
  });
});
