import { test, expect } from '@playwright/test';

test.describe('Tapestry Smoke Tests', () => {

  test('homepage loads with monthly view', async ({ page }) => {
    await page.goto('/');
    await expect(page).toHaveTitle(/\w+ \d{4} — Tapestry/);
    await expect(page.locator('.stats-cards .card')).toHaveCount(4);
    await expect(page.locator('.month-nav h1')).toBeVisible();
  });

  test('homepage uses Dracula theme', async ({ page }) => {
    await page.goto('/');
    const bg = await page.evaluate(() =>
      getComputedStyle(document.documentElement).getPropertyValue('--bg').trim()
    );
    expect(bg).toBe('#282a36');
  });

  test('homepage hides empty rigs', async ({ page }) => {
    await page.goto('/');
    const sections = await page.locator('.db-section').all();
    for (const section of sections) {
      const totalText = await section.locator('.stat').first().textContent();
      const total = parseInt(totalText?.replace('Total: ', '') || '0');
      expect(total).toBeGreaterThan(0);
    }
  });

  test('homepage hides internal statuses', async ({ page }) => {
    await page.goto('/');
    const badges = await page.locator('.db-stats .badge').allTextContents();
    for (const badge of badges) {
      expect(badge).not.toMatch(/^hooked:/);
      expect(badge).not.toMatch(/^pinned:/);
    }
  });

  test('status page loads', async ({ page }) => {
    await page.goto('/status');
    await expect(page).toHaveTitle(/Executive Status — Tapestry/);
    await expect(page.locator('.status-header h1')).toHaveText('Executive Status');
    await expect(page.locator('.status-card')).toHaveCount(5);
  });

  test('status page has no double-slash URLs', async ({ page }) => {
    await page.goto('/status');
    const links = await page.locator('a[href*="/bead/"]').all();
    for (const link of links) {
      const href = await link.getAttribute('href');
      expect(href).not.toContain('//');
    }
  });

  test('briefing page loads', async ({ page }) => {
    await page.goto('/briefing');
    await expect(page).toHaveTitle(/Briefing — Tapestry/);
    await expect(page.locator('.briefing-card').first()).toBeVisible();
  });

  test('agents page loads', async ({ page }) => {
    await page.goto('/agents');
    await expect(page).toHaveTitle(/Agents — Tapestry/);
    await expect(page.locator('.agent-card').first()).toBeVisible();
  });

  test('agents page shows short names', async ({ page }) => {
    await page.goto('/agents');
    const names = await page.locator('.agent-name').allTextContents();
    for (const name of names) {
      expect(name).not.toContain('@');
    }
  });

  test('beads page loads', async ({ page }) => {
    await page.goto('/beads');
    await expect(page).toHaveTitle(/Beads — Tapestry/);
  });

  test('decisions page loads', async ({ page }) => {
    await page.goto('/decisions');
    await expect(page).toHaveTitle(/Decisions — Tapestry/);
    await expect(page.locator('.decisions-page h1')).toHaveText('Decisions');
    await expect(page.locator('.stats-grid .stat-card')).toHaveCount(4);
    await expect(page.locator('.filters a')).toHaveCount(4);
  });

  test('decisions filter links work', async ({ page }) => {
    await page.goto('/decisions?filter=pending');
    await expect(page.locator('.filters a.active')).toHaveText('Pending');
  });

  test('achievements page loads', async ({ page }) => {
    await page.goto('/achievements');
    await expect(page).toHaveTitle(/Achievements — Tapestry/);
    await expect(page.locator('.achievements-page h1')).toHaveText('Achievements');
    await expect(page.locator('.progress-count')).toBeVisible();
    await expect(page.locator('.achievement-card')).toHaveCount(25);
  });

  test('achievements shows unlocked and locked cards', async ({ page }) => {
    await page.goto('/achievements');
    const unlocked = await page.locator('.achievement-card.unlocked').count();
    const locked = await page.locator('.achievement-card.locked').count();
    expect(unlocked).toBe(16);
    expect(locked).toBe(9);
  });

  test('achievements category filter works', async ({ page }) => {
    await page.goto('/achievements?category=infrastructure');
    await expect(page.locator('.filters a.active')).toContainText('infrastructure');
    const cards = await page.locator('.achievement-card').count();
    expect(cards).toBeLessThan(25);
    expect(cards).toBeGreaterThan(0);
  });

  test('homelab page loads', async ({ page }) => {
    await page.goto('/homelab');
    await expect(page).toHaveTitle(/Homelab — Tapestry/);
    await expect(page.locator('.homelab-page h1')).toBeVisible();
    await expect(page.locator('.briefing-card')).toHaveCount(3);
  });

  test('homelab shows target status', async ({ page }) => {
    await page.goto('/homelab');
    const upBadges = await page.locator('.badge.status-closed').count();
    expect(upBadges).toBeGreaterThan(0);
  });

  test('designs page loads', async ({ page }) => {
    await page.goto('/designs');
    await expect(page).toHaveTitle(/Designs — Tapestry/);
    await expect(page.locator('.designs-page h1')).toBeVisible();
    const cards = await page.locator('.design-card').count();
    expect(cards).toBeGreaterThan(10);
  });

  test('design doc renders markdown', async ({ page }) => {
    await page.goto('/designs/clean-desk');
    await expect(page.locator('.design-content')).toBeVisible();
    await expect(page.locator('.design-content h2')).toHaveCount({ minimum: 1 });
  });

  test('search page loads', async ({ page }) => {
    await page.goto('/search');
    await expect(page).toHaveTitle(/Search — Tapestry/);
    await expect(page.locator('text=Search')).toBeVisible();
  });

  test('navigation links work', async ({ page }) => {
    // Start from a fast-loading page to avoid homepage Dolt timeout
    await page.goto('/status');
    await expect(page.locator('.status-header h1')).toBeVisible();

    await page.goto('/briefing');
    await expect(page.locator('.briefing-card').first()).toBeVisible();

    await page.goto('/agents');
    await expect(page.locator('.agent-card').first()).toBeVisible();

    await page.goto('/decisions');
    await expect(page.locator('.decisions-page h1')).toBeVisible();

    await page.goto('/achievements');
    await expect(page.locator('.achievements-page h1')).toBeVisible();

    await page.goto('/homelab');
    await expect(page.locator('.homelab-page h1')).toBeVisible();

    await page.goto('/designs');
    await expect(page.locator('.designs-page h1')).toBeVisible();

    await page.goto('/beads');
    await expect(page).toHaveTitle(/Beads — Tapestry/);
  });

  test('month navigation works', async ({ page }) => {
    await page.goto('/');
    // Click previous month link and wait for HTMX content swap
    await page.click('.month-nav a:first-child');
    await page.waitForURL(/\/\d{4}\/\d{2}/);
    const url = page.url();
    expect(url).toMatch(/\/\d{4}\/\d{2}/);
  });

});
