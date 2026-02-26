import { expect, test } from '@playwright/test'

test.describe('Smoke Tests', () => {
  test('application should load', async ({ page }) => {
    await page.goto('/')

    // Basic check that the app loaded - check for login form instead of title
    await expect(page.locator('input[type="text"]').first()).toBeVisible({ timeout: 10000 })
  })

  test('should redirect to login when not authenticated', async ({ page }) => {
    // Clear any existing auth
    await page.context().clearCookies()

    // Try to access protected route
    await page.goto('/dashboard')

    // Should be redirected to login
    await page.waitForURL(/.*login.*/, { timeout: 10000 }).catch(() => {
      // Might already be on login or a different redirect
    })

    const url = page.url()
    expect(url).toMatch(/login|auth/i)
  })

  test('should have responsive layout', async ({ page }) => {
    await page.goto('/')

    // Test mobile viewport
    await page.setViewportSize({ width: 375, height: 667 })
    await page.waitForTimeout(500)

    // App should still be visible
    await expect(page.locator('body')).toBeVisible()

    // Test desktop viewport
    await page.setViewportSize({ width: 1280, height: 720 })
    await page.waitForTimeout(500)

    await expect(page.locator('body')).toBeVisible()
  })
})
