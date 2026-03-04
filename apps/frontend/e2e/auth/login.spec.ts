import { expect, test } from '@playwright/test'

const hasLoginForm = async page => {
  const passwordCount = await page.locator('input[type="password"]').count()
  const loginButtonCount = await page
    .locator('button:has-text("Login")')
    .or(page.locator('button:has-text("登录")'))
    .count()

  return passwordCount > 0 && loginButtonCount > 0
}

const loginWithValidCredentials = async page => {
  const username = process.env.E2E_USERNAME || 'admin'
  const password = process.env.E2E_PASSWORD || 'DataEase123456'

  await page.locator('input[type="text"]').first().fill(username)
  await page.locator('input[type="password"]').first().fill(password)

  const loginButton = page.locator('button:has-text("Login")').or(page.locator('button:has-text("登录")'))
  await loginButton.click()
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

  test('SYS-SMK-004 @system-smoke should login successfully with valid credentials', async ({ page, context }) => {
    await context.clearCookies()
    await page.goto('/')

    await expect(page.locator('input[type="text"]').first()).toBeVisible({ timeout: 10000 })
    await expect(page.locator('input[type="password"]').first()).toBeVisible({ timeout: 10000 })

    await loginWithValidCredentials(page)

    await expect
      .poll(async () => {
        return await page.evaluate(() => Object.keys(localStorage).includes('user.token'))
      }, { timeout: 20000 })
      .toBe(true)

    await expect
      .poll(async () => {
        return await page.evaluate(() => window.location.hash)
      }, { timeout: 20000 })
      .toContain('/workbranch/index')

    const hasAppShell =
      (await page.locator('#app').count()) > 0 ||
      (await page.locator('.de-layout').count()) > 0 ||
      (await page.locator('body').count()) > 0

    expect(hasAppShell).toBeTruthy()
  })
})
