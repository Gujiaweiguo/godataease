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

test.describe('Help Menu Tests', () => {
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

  test('should display help menu trigger', async ({ page }) => {
    const helpTrigger = page.locator('.more-menu-trigger')
    await expect(helpTrigger).toBeVisible({ timeout: 10000 })
  })

  test('should open help menu on hover', async ({ page }) => {
    const helpTrigger = page.locator('.more-menu-trigger')
    await helpTrigger.hover()
    await page.waitForTimeout(500)
    
    const popover = page.locator('.more-menu-popover')
    await expect(popover).toBeVisible({ timeout: 5000 })
  })

  test('should display export center option', async ({ page }) => {
    const helpTrigger = page.locator('.more-menu-trigger')
    await helpTrigger.hover()
    await page.waitForTimeout(500)
    
    const exportCenter = page.locator('.more-menu-content >> text=导出中心')
    await expect(exportCenter).toBeVisible({ timeout: 5000 })
  })

  test('should trigger export center event on click', async ({ page }) => {
    const helpTrigger = page.locator('.more-menu-trigger')
    await helpTrigger.hover()
    await page.waitForTimeout(500)
    
    const exportCenter = page.locator('.more-menu-item:has-text("导出中心")')
    await exportCenter.click()
  })
})
