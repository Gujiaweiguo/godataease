import type { BrowserContext, Page } from '@playwright/test'
import { expect, test } from '@playwright/test'
import { getLoginButton, getPasswordInput, getUsernameInput, hasLoginForm } from './utils/auth'

const loginAndOpenRoleTab = async (page: Page, context: BrowserContext) => {
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

  const roleListResp = page.waitForResponse(
    response => response.url().includes('/role/byCurOrg') && response.status() === 200
  )

  await page.goto('/#/system/user')
  await page.getByRole('tab', { name: '角色' }).click()
  await roleListResp
  await expect(page.locator('body')).toContainText(/角色管理|菜单授权|权限设置/)
  return true
}

test('role tab should load role list after login', async ({ page, context }) => {
  const opened = await loginAndOpenRoleTab(page, context)
  if (!opened) return

  await expect(page).not.toHaveURL(/401|404/)
  await expect(page.locator('.role-management')).toBeVisible()
})

test('role tab menu auth dialog should load menu tree and current role auth', async ({ page, context }) => {
  const opened = await loginAndOpenRoleTab(page, context)
  if (!opened) return

  const menuTreeResp = page.waitForResponse(
    response => response.url().includes('/auth/menuResource') && response.status() === 200
  )
  const roleMenuResp = page.waitForResponse(
    response => response.url().includes('/roleMenu/auth/') && response.status() === 200
  )

  await page.getByRole('button', { name: '菜单授权' }).first().click()

  await Promise.all([menuTreeResp, roleMenuResp])
  await expect(page.locator('.el-dialog')).toContainText('菜单授权')
})

test('role tab permission dialog should load resource tree and current role permissions', async ({ page, context }) => {
  const opened = await loginAndOpenRoleTab(page, context)
  if (!opened) return

  const resourceTreeResp = page.waitForResponse(
    response => response.url().includes('/auth/busiResource/1') && response.status() === 200
  )
  const rolePermResp = page.waitForResponse(
    response => response.url().includes('/auth/busiPermission') && response.status() === 200
  )

  await page.getByRole('button', { name: '权限设置' }).first().click()

  const [, permissionResponse] = await Promise.all([resourceTreeResp, rolePermResp])
  const permissionBody = await permissionResponse.json()
  await expect(page.locator('.el-dialog')).toContainText('权限设置')

  const expectedPermCount = Array.isArray(permissionBody?.data?.permIds)
    ? permissionBody.data.permIds.length
    : 0
  const checkedCount = await page.locator('.el-dialog .el-tree .is-checked').count()

  if (expectedPermCount > 0) {
    expect(checkedCount).toBeGreaterThan(0)
  }
})
