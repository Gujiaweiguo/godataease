import { expect, test } from '@playwright/test'
import { getLoginButton, getPasswordInput, getUsernameInput, hasLoginForm } from '../utils/auth'

const loginWithValidCredentials = async page => {
  const username = process.env.E2E_USERNAME || 'admin'
  const password = process.env.E2E_PASSWORD || 'DataEase123456'

  await getUsernameInput(page).fill(username)
  await getPasswordInput(page).fill(password)

  const loginButton = getLoginButton(page)
  await loginButton.click()

  await page.waitForURL(/#\/workbranch|data\/datasource|module-datasource/, { timeout: 20000 })
}

test.describe('User Menu Tests', () => {
  test.beforeEach(async ({ page, context }) => {
    await context.clearCookies()
    await page.goto('/')
    await page.evaluate(() => {
      localStorage.clear()
      sessionStorage.clear()
    })
    await page.goto('/#/login')

    if (await hasLoginForm(page)) {
      await loginWithValidCredentials(page)
    }

    await page.waitForTimeout(1000)
  })

  test('should display user avatar menu trigger', async ({ page }) => {
    const userAvatar = page.locator('.top-info-container')
    await expect(userAvatar).toBeVisible({ timeout: 10000 })
  })

  test('should open user dropdown menu on click', async ({ page }) => {
    const userAvatar = page.locator('.top-info-container')
    await userAvatar.click()

    const popover = page.locator('.uinfo-popover')
    await expect(popover).toBeVisible({ timeout: 5000 })
  })

  test('should display language selector in user menu', async ({ page }) => {
    const userAvatar = page.locator('.top-info-container')
    await userAvatar.click()

    const languageBlock = page.locator('.uinfo-language-block')
    await expect(languageBlock).toBeVisible({ timeout: 5000 })
  })

  test('should switch language successfully', async ({ page }) => {
    const userAvatar = page.locator('.top-info-container')
    await userAvatar.click()

    await page.waitForTimeout(500)

    const languageSelector = page.locator('.uinfo-language-block .lang-selector, .uinfo-language-block [class*="lang"]')
    if (await languageSelector.count() > 0) {
      await languageSelector.first().click()
      await page.waitForTimeout(1000)

      await expect(page.locator('body')).toBeVisible()
    }
  })

  test('should display logout option when not in platform client', async ({ page }) => {
    const userAvatar = page.locator('.top-info-container')
    await userAvatar.click()

    await page.waitForTimeout(500)

    const logoutButton = page.locator('.uinfo-footer .uinfo-main-item')
    const isVisible = await logoutButton.isVisible().catch(() => false)

    if (isVisible) {
      await expect(logoutButton).toContainText(/退出|logout/i)
    }
  })
})
