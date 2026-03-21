import type { BrowserContext, Page } from '@playwright/test'
import { expect, test } from '@playwright/test'
import { getLoginButton, getPasswordInput, getUsernameInput, hasLoginForm } from './utils/auth'

const login = async (page: Page, context: BrowserContext) => {
  await context.clearCookies()
  await page.goto('/')
  await page.evaluate(() => {
    localStorage.clear()
    sessionStorage.clear()
  })
  await page.goto('/#/login')

  if (!(await hasLoginForm(page))) {
    await expect(page.locator('body')).toBeVisible({ timeout: 10000 })
    return false
  }

  await expect(getUsernameInput(page)).toBeVisible({ timeout: 10000 })
  await expect(getPasswordInput(page)).toBeVisible({ timeout: 10000 })
  await getUsernameInput(page).fill(process.env.E2E_USERNAME || 'admin')
  await getPasswordInput(page).fill(process.env.E2E_PASSWORD || 'DataEase123456')
  await getLoginButton(page).click()
  await page.waitForURL((url: URL) => !url.toString().includes('/login'), { timeout: 20000 })
  return true
}

test('missing system-management route should stay 404 instead of 401 after login', async ({ page, context }) => {
  const loggedIn = await login(page, context)
  if (!loggedIn) return

  await page.goto('/#/system/not-exists-for-recovery-check')
  await page.waitForTimeout(1500)

  const finalUrl = page.url()
  expect(finalUrl.includes('/404')).toBeTruthy()
  expect(finalUrl.includes('/401')).toBeFalsy()
})

test('protected system-management route should redirect to login instead of 404 without session', async ({ page, context }) => {
  const loggedIn = await login(page, context)
  if (!loggedIn) return

  await page.evaluate(() => {
    localStorage.clear()
    sessionStorage.clear()
  })
  await context.clearCookies()

  await page.goto('/#/system/user')
  await page.waitForTimeout(1500)

  const finalUrl = page.url()
  expect(finalUrl.includes('/login')).toBeTruthy()
  expect(finalUrl.includes('/404')).toBeFalsy()
})
