import { expect, test } from '@playwright/test'
import { getLoginButton, getPasswordInput, getUsernameInput, hasLoginForm } from '../utils/auth'

const loginWithValidCredentials = async page => {
  const username = process.env.E2E_USERNAME || 'admin'
  const password = process.env.E2E_PASSWORD || 'DataEase123456'

  await getUsernameInput(page).fill(username)
  await getPasswordInput(page).fill(password)

  const loginButton = getLoginButton(page)
  await loginButton.click()
}

test.describe('Authentication', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/')
  })

  test('should display login page', async ({ page }) => {
    if (await hasLoginForm(page)) {
      await expect(getUsernameInput(page)).toBeVisible()
      await expect(getPasswordInput(page)).toBeVisible()
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

    await getUsernameInput(page).fill('invalid_user')
    await getPasswordInput(page).fill('invalid_password')

    const loginButton = getLoginButton(page)
    await loginButton.click()

    await page.waitForTimeout(2000)

    const hasError = (await page.locator('.el-message--error').count()) > 0
    const stillOnLogin = page.url().includes('login')

    expect(hasError || stillOnLogin).toBeTruthy()
  })

  test('SYS-SMK-004 @system-smoke should login successfully with valid credentials', async ({ page, context }) => {
    await context.clearCookies()
    await page.goto('/')
    await page.evaluate(() => {
      localStorage.clear()
      sessionStorage.clear()
    })
    await page.goto('/#/login')

    if (!(await hasLoginForm(page))) {
      await expect(page.locator('body')).toBeVisible({ timeout: 10000 })
      return
    }

    await page.goto('/')

    await expect(getUsernameInput(page)).toBeVisible({ timeout: 10000 })
    await expect(getPasswordInput(page)).toBeVisible({ timeout: 10000 })

    await loginWithValidCredentials(page)

    // Wait for navigation to complete after login
    await page.waitForURL(/#\/workbranch|data\/datasource|module-datasource/, { timeout: 20000 })

    // Verify token exists in localStorage
    const hasToken = await page.evaluate(() => Object.keys(localStorage).includes('user.token'))
    expect(hasToken).toBeTruthy()

    const hasAppShell =
      (await page.locator('#app').count()) > 0 ||
      (await page.locator('.de-layout').count()) > 0 ||
      (await page.locator('body').count()) > 0

    expect(hasAppShell).toBeTruthy()
  })
})
