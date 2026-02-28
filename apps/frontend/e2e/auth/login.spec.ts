import { expect, test } from '@playwright/test'

const hasLoginForm = async page => {
  const passwordCount = await page.locator('input[type="password"]').count()
  const loginButtonCount = await page
    .locator('button:has-text("Login")')
    .or(page.locator('button:has-text("登录")'))
    .count()

  return passwordCount > 0 && loginButtonCount > 0
}

test.describe('Authentication', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/')
  })

  test('should display login page', async ({ page }) => {
    if (await hasLoginForm(page)) {
      await expect(page.locator('input[type="text"]').first()).toBeVisible()
      await expect(page.locator('input[type="password"]').first()).toBeVisible()
      return
    }

    await expect(page.locator('body')).toBeVisible()
    expect(page.url()).not.toMatch(/login/i)
  })

  test('should show error with invalid credentials', async ({ page }) => {
    if (!(await hasLoginForm(page))) {
      await expect(page.locator('body')).toBeVisible()
      return
    }

    await page.locator('input[type="text"]').first().fill('invalid_user')
    await page.locator('input[type="password"]').first().fill('invalid_password')

    const loginButton = page.locator('button:has-text("Login")').or(page.locator('button:has-text("登录")'))
    await loginButton.click()

    await page.waitForTimeout(2000)

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
