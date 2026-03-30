import type { BrowserContext, Page } from '@playwright/test'
import { expect, test } from '@playwright/test'
import { getLoginButton, getPasswordInput, getUsernameInput, hasLoginForm } from './utils/auth'

const loginAndWaitForShell = async (page: Page, context: BrowserContext) => {
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
  await expect(page.locator('body')).toBeVisible({ timeout: 10000 })
  return true
}

const navigateByMenuText = async (page: Page, label: string, expectedUrlPart: string) => {
  const target = page.getByText(label, { exact: false }).first()
  await expect(target).toBeVisible({ timeout: 10000 })
  await target.click()
  await page.waitForURL((url: URL) => url.toString().includes(expectedUrlPart), { timeout: 20000 })
  await expect(page).not.toHaveURL(/401|404/)
}

test('admin navigation should show new first-level groups', async ({ page, context }) => {
  const opened = await loginAndWaitForShell(page, context)
  if (!opened) return

  await expect(page.locator('body')).toContainText(/组织权限|系统设置|工具箱/)
})

test('system-management menu navigation should reach recovered admin pages', async ({ page, context }) => {
  const opened = await loginAndWaitForShell(page, context)
  if (!opened) return

  await navigateByMenuText(page, '用户管理', '/system/user')
  await expect(page.locator('.user-management')).toBeVisible()

  await navigateByMenuText(page, '组织管理', '/system/org')
  await expect(page.locator('.org-management')).toBeVisible()

  await navigateByMenuText(page, '角色管理', '/system/role')
  await expect(page.locator('.role-management')).toBeVisible()

  await navigateByMenuText(page, '菜单管理', '/system/menu')
  await expect(page.locator('.menu-management')).toBeVisible()

  await navigateByMenuText(page, '权限管理', '/system/permission')
  await expect(page.locator('.permission-config')).toBeVisible()
})
