import { expect, test } from '@playwright/test'

test.describe('Authentication', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/')
  })

  test('should display login page', async ({ page }) => {
    await expect(page).toHaveURL(/.*login.*/)
    await expect(page.locator('input[type="text"]').first()).toBeVisible()
    await expect(page.locator('input[type="password"]').first()).toBeVisible()
  })

  test('should show error with invalid credentials', async ({ page }) => {
    await page.locator('input[type="text"]').first().fill('invalid_user')
    await page.locator('input[type="password"]').first().fill('invalid_password')

    // Support both Chinese and English UI
    const loginButton = page.locator('button:has-text("Login")').or(page.locator('button:has-text("登录")'))
    await loginButton.click()

    // Wait for error message or URL change
    await page.waitForTimeout(2000)

    // Should still be on login page or show error
    const hasError = (await page.locator('.el-message--error').count()) > 0
    const stillOnLogin = page.url().includes('login')

    expect(hasError || stillOnLogin).toBeTruthy()
  })

  // Note: This test requires backend service to be running
  test.fixme('should login successfully with valid credentials', async ({ page }) => {
    const username = process.env.E2E_USERNAME || 'admin'
    const password = process.env.E2E_PASSWORD || 'DataEase123456'

    await page.locator('input[type="text"]').first().fill(username)
    await page.locator('input[type="password"]').first().fill(password)

    // Support both Chinese and English UI
    const loginButton = page.locator('button:has-text("Login")').or(page.locator('button:has-text("登录")'))
    await loginButton.click()

    // Wait for navigation away from login page
    await page.waitForURL(/^(?!.*login).*/, { timeout: 15000 }).catch(() => {
      // If timeout, check if we're still on login due to invalid credentials
    })

    // Verify we're not on login page anymore (successful login)
    const currentUrl = page.url()
    expect(currentUrl).not.toContain('login')
  })
})
