import type { BrowserContext, Page } from '@playwright/test'
import { expect, test } from '@playwright/test'
import { getLoginButton, getPasswordInput, getUsernameInput, hasLoginForm } from './utils/auth'

const loginFromPage = async (page: Page) => {
  await expect(getUsernameInput(page)).toBeVisible({ timeout: 10000 })
  await expect(getPasswordInput(page)).toBeVisible({ timeout: 10000 })
  await getUsernameInput(page).fill(process.env.E2E_USERNAME || 'admin')
  await getPasswordInput(page).fill(process.env.E2E_PASSWORD || 'DataEase123456')
  await getLoginButton(page).click()
}

test('protected redirect should land on target route only after route readiness', async ({ page, context }) => {
  await context.clearCookies()
  await page.goto('/')
  await page.evaluate(() => {
    localStorage.clear()
    sessionStorage.clear()
  })

  await page.goto('/#/login?redirect=%2Fsystem%2Fmenu')
  if (!(await hasLoginForm(page))) return

  await loginFromPage(page)
  await page.waitForURL((url: URL) => url.toString().includes('/system/menu'), { timeout: 20000 })

  const finalUrl = page.url()
  expect(finalUrl.includes('/401')).toBeFalsy()
  expect(finalUrl.includes('/404')).toBeFalsy()
  await expect(page.locator('.menu-management')).toBeVisible({ timeout: 10000 })
})

test('protected route reload should keep current-user bootstrap and dynamic routes reliable', async ({ page, context }) => {
  await context.clearCookies()
  await page.goto('/')
  await page.evaluate(() => {
    localStorage.clear()
    sessionStorage.clear()
  })

  await page.goto('/#/login')
  if (!(await hasLoginForm(page))) return

  await loginFromPage(page)
  await page.waitForURL((url: URL) => !url.toString().includes('/login'), { timeout: 20000 })

  await page.goto('/#/system/permission')
  await expect(page.locator('.permission-config')).toBeVisible({ timeout: 10000 })

  await page.reload()
  await expect(page.locator('.permission-config')).toBeVisible({ timeout: 10000 })

  const finalUrl = page.url()
  expect(finalUrl.includes('/401')).toBeFalsy()
  expect(finalUrl.includes('/404')).toBeFalsy()
})
