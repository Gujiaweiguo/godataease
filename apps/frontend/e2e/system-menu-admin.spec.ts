import type { BrowserContext, Page } from '@playwright/test'
import { expect, test } from '@playwright/test'
import { getLoginButton, getPasswordInput, getUsernameInput, hasLoginForm } from './utils/auth'

const loginAndOpenMenuPage = async (page: Page, context: BrowserContext) => {
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

  const menuQueryResp = page.waitForResponse(
    response => response.url().includes('/menu/query') && response.status() === 200
  )

  await page.goto('/#/system/menu')
  await menuQueryResp
  await expect(page.locator('.menu-management')).toBeVisible({ timeout: 10000 })
  return true
}

test('system menu page should load visible menu tree', async ({ page, context }) => {
  const opened = await loginAndOpenMenuPage(page, context)
  if (!opened) return

  await expect(page.locator('body')).toContainText(/菜单管理|新建根菜单/)
  await expect(page.locator('.el-table__body-wrapper .el-table__row').first()).toBeVisible()
})

test('system menu edit entry should load menu detail payload', async ({ page, context }) => {
  const opened = await loginAndOpenMenuPage(page, context)
  if (!opened) return

  const detailResp = page.waitForResponse(
    response => response.url().includes('/menu/detail/') && response.status() === 200
  )

  await page.getByRole('button', { name: '编辑' }).first().click()
  await detailResp

  await expect(page.locator('.el-dialog')).toContainText('编辑菜单')
  await expect(page.locator('.el-dialog input[placeholder="请输入菜单名称"]')).toBeVisible()
})

test('system menu page should support root menu create and delete entry path', async ({ page, context }) => {
  const opened = await loginAndOpenMenuPage(page, context)
  if (!opened) return

  const uniqueName = `E2E菜单${Date.now()}`
  const uniquePath = `e2e-menu-${Date.now()}`

  await page.getByRole('button', { name: '新建根菜单' }).click()
  await expect(page.locator('.el-dialog')).toContainText('新建根菜单')

  await page.locator('.el-dialog input[placeholder="请输入菜单名称"]').fill(uniqueName)
  await page.locator('.el-dialog input[placeholder*="例如：/system/menu 或 menu"]').fill(uniquePath)
  await page.locator('.el-dialog').getByRole('combobox').nth(1).click()
  await page.locator('.el-select-dropdown__item').filter({ hasText: '目录' }).first().click()

  const createResp = page.waitForResponse(
    response => response.url().includes('/menu/create') && response.status() === 200
  )
  await page.locator('.el-dialog').getByRole('button', { name: '确定' }).click()
  await createResp
  await expect(page.locator('.el-message--success')).toContainText('菜单创建成功', { timeout: 10000 })
  await expect(page.locator('.el-table__body-wrapper')).toContainText(uniqueName)

  const createdRow = page.locator('.el-table__body-wrapper .el-table__row').filter({ hasText: uniqueName }).first()
  const deleteResp = page.waitForResponse(
    response => response.url().includes('/menu/delete/') && response.status() === 200
  )
  await createdRow.getByRole('button', { name: '删除' }).click()
  await page.locator('.el-message-box').getByRole('button', { name: '确定' }).click()
  await deleteResp
  await expect(page.locator('.el-message--success')).toContainText('删除成功', { timeout: 10000 })
})
