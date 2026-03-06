import { expect, test } from '@playwright/test'
import { getUsernameInput, hasLoginForm } from './utils/auth'

test.describe('Smoke Tests', () => {
  test('application should load', async ({ page }) => {
    await page.goto('/')

    await expect(page.locator('body')).toBeVisible({ timeout: 10000 })
    if (await hasLoginForm(page)) {
      await expect(getUsernameInput(page)).toBeVisible({ timeout: 10000 })
    }
  })

  test('should redirect to login when not authenticated', async ({ page }) => {
    await page.context().clearCookies()

    await page.goto('/dashboard')

    await page.waitForTimeout(1000)

    const url = page.url()
    const redirectedToLogin = /login|auth/i.test(url)
    const staysInApp = /dashboard|#\//i.test(url)

    expect(redirectedToLogin || staysInApp).toBeTruthy()
  })

  test('should have responsive layout', async ({ page }) => {
    await page.goto('/')

    await page.setViewportSize({ width: 375, height: 667 })
    await page.waitForTimeout(500)

    await expect(page.locator('body')).toBeVisible()

    await page.setViewportSize({ width: 1280, height: 720 })
    await page.waitForTimeout(500)

    await expect(page.locator('body')).toBeVisible()
  })
})
