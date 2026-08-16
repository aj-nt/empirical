import { test, expect } from '@playwright/test';

/**
 * Visual snapshot tests for chart components.
 * 
 * These tests create a chart with known birth data, navigate to each
 * component, and take a screenshot. Playwright compares against checked-in
 * baselines in e2e/visual/screenshots/.
 * 
 * To generate baselines: npx playwright test --config=playwright.visual.config.ts --update-snapshots
 * To run: npx playwright test --config=playwright.visual.config.ts
 */

const AJ_BIRTH = {
  name: 'AJ',
  year: '1969',
  month: '2',
  day: '15',
  hour: '23',
  minute: '10',
  tz: '-8',
  lat: '47.038',
  lng: '-122.901',
};

async function createChart(page, data: typeof AJ_BIRTH) {
  await page.goto('/');
  await page.click('button:has-text("+ New Chart")');
  
  const numberInputs = page.locator('input[type="number"]');
  await numberInputs.nth(0).fill(data.year);
  await numberInputs.nth(1).fill(data.month);
  await numberInputs.nth(2).fill(data.day);
  await numberInputs.nth(3).fill(data.hour);
  await numberInputs.nth(4).fill(data.minute);
  await numberInputs.nth(5).fill(data.tz);
  
  // Fill lat/lng if present
  const allInputs = page.locator('input[type="number"]');
  const count = await allInputs.count();
  if (count >= 8) {
    await allInputs.nth(6).fill(data.lat);
    await allInputs.nth(7).fill(data.lng);
  }
  
  await page.fill('input[placeholder="e.g. AJ"]', data.name);
  await page.click('button:has-text("Save Chart")');
  
  // Wait for the wheel to render
  await expect(page.locator('svg')).toBeVisible({ timeout: 15000 });
}

test.describe('Visual: Chart Wheel', () => {
  test('natal wheel renders correctly', async ({ page }) => {
    await createChart(page, AJ_BIRTH);
    
    // The wheel tab should be active by default
    await expect(page.getByRole('button', { name: 'Wheel', exact: true })).toBeVisible();
    
    // Wait for SVG to be fully rendered
    await page.waitForTimeout(500);
    
    await expect(page).toHaveScreenshot('chart-wheel-natal.png', {
      fullPage: false,
      maxDiffPixels: 100,
    });
  });

  test('bi-wheel renders correctly', async ({ page }) => {
    await createChart(page, AJ_BIRTH);
    
    // Navigate to Bi-Wheel tab
    await page.click('button:has-text("Bi-Wheel")');
    
    // Wait for bi-wheel SVG to render
    await expect(page.locator('svg')).toBeVisible({ timeout: 15000 });
    await page.waitForTimeout(500);
    
    await expect(page).toHaveScreenshot('chart-bi-wheel.png', {
      fullPage: false,
      maxDiffPixels: 100,
    });
  });

  test('tri-wheel renders correctly', async ({ page }) => {
    await createChart(page, AJ_BIRTH);
    
    // Navigate to Tri-Wheel tab
    await page.click('button:has-text("Tri-Wheel")');
    
    // Wait for tri-wheel SVG to render
    await expect(page.locator('svg')).toBeVisible({ timeout: 15000 });
    await page.waitForTimeout(500);
    
    await expect(page).toHaveScreenshot('chart-tri-wheel.png', {
      fullPage: false,
      maxDiffPixels: 100,
    });
  });
});

test.describe('Visual: Natal Components', () => {
  test('planet table renders correctly', async ({ page }) => {
    await createChart(page, AJ_BIRTH);
    
    // Navigate to Natal tab
    await page.click('button:has-text("Natal")');
    await page.waitForTimeout(500);
    
    await expect(page).toHaveScreenshot('natal-planet-table.png', {
      fullPage: false,
      maxDiffPixels: 100,
    });
  });

  test('aspect grid renders correctly', async ({ page }) => {
    await createChart(page, AJ_BIRTH);
    
    // Navigate to Natal tab
    await page.click('button:has-text("Natal")');
    await page.waitForTimeout(500);
    
    // Scroll to aspect grid if needed
    const aspectGrid = page.locator('text=Aspects');
    if (await aspectGrid.isVisible()) {
      await aspectGrid.scrollIntoViewIfNeeded();
      await page.waitForTimeout(300);
    }
    
    await expect(page).toHaveScreenshot('natal-aspect-grid.png', {
      fullPage: false,
      maxDiffPixels: 100,
    });
  });

  test('dashboard renders correctly', async ({ page }) => {
    await createChart(page, AJ_BIRTH);
    
    // Navigate to Natal tab
    await page.click('button:has-text("Natal")');
    await page.waitForTimeout(500);
    
    // Scroll to dashboard/patterns section
    const dashboard = page.locator('text=Patterns');
    if (await dashboard.isVisible()) {
      await dashboard.scrollIntoViewIfNeeded();
      await page.waitForTimeout(300);
    }
    
    await expect(page).toHaveScreenshot('natal-dashboard.png', {
      fullPage: false,
      maxDiffPixels: 100,
    });
  });
});

test.describe('Visual: Reports', () => {
  test('page designer renders correctly', async ({ page }) => {
    await createChart(page, AJ_BIRTH);
    
    // Navigate to Reports → Page Designer
    await page.click('button:has-text("Reports")');
    await page.waitForTimeout(300);
    
    const designerBtn = page.locator('button:has-text("Page Designer")');
    if (await designerBtn.isVisible()) {
      await designerBtn.click();
      await page.waitForTimeout(500);
      
      await expect(page).toHaveScreenshot('reports-page-designer.png', {
        fullPage: false,
        maxDiffPixels: 100,
      });
    }
  });
});

test.describe('Visual: Theme', () => {
  test('dark theme renders correctly', async ({ page }) => {
    await createChart(page, AJ_BIRTH);
    await page.waitForTimeout(500);
    
    await expect(page).toHaveScreenshot('theme-dark.png', {
      fullPage: false,
      maxDiffPixels: 100,
    });
  });

  test('light theme renders correctly', async ({ page }) => {
    await createChart(page, AJ_BIRTH);
    
    // Switch to light theme
    await page.click('button[title="Light"]');
    await page.waitForTimeout(300);
    
    await expect(page).toHaveScreenshot('theme-light.png', {
      fullPage: false,
      maxDiffPixels: 100,
    });
  });

  test('high contrast theme renders correctly', async ({ page }) => {
    await createChart(page, AJ_BIRTH);
    
    // Switch to high contrast theme
    await page.click('button[title="High Contrast"]');
    await page.waitForTimeout(300);
    
    await expect(page).toHaveScreenshot('theme-high-contrast.png', {
      fullPage: false,
      maxDiffPixels: 100,
    });
  });
});
