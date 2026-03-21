import type { BrowserContext, Page } from '@playwright/test'
import { expect, test } from '@playwright/test'
import { getLoginButton, getPasswordInput, getUsernameInput, hasLoginForm } from './utils/auth'

const loginAndOpenUserPage = async (page: Page, context: BrowserContext) => {
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

  const userListResp = page.waitForResponse(
    response => response.url().includes('/user/byCurOrg') && response.status() === 200
  )
  const orgOptionResp = page.waitForResponse(
    response => response.url().includes('/user/org/option') && response.status() === 200
  )
  const roleListResp = page.waitForResponse(
    response => response.url().includes('/role/byCurOrg') && response.status() === 200
  )

  await page.goto('/#/system/user')

  await Promise.all([userListResp, orgOptionResp, roleListResp])
  await expect(page.locator('body')).toBeVisible({ timeout: 10000 })
  return true
}

test('system user page should initialize core APIs after login', async ({ page, context }) => {
  const opened = await loginAndOpenUserPage(page, context)
  if (!opened) return

  await expect(page).not.toHaveURL(/401|404/)
  await expect(page.locator('body')).toContainText(/用户|角色管理|新建用户/)
})

test('system user create dialog should load organization options instead of user options', async ({ page, context }) => {
  const opened = await loginAndOpenUserPage(page, context)
  if (!opened) return

  await page.getByRole('button', { name: '新建用户' }).click()
  await expect(page.locator('.el-dialog')).toContainText('新建用户')

  const orgSelect = page.locator('.el-dialog .el-select').nth(0)
  await orgSelect.click()

  const optionTexts = await page.locator('.el-select-dropdown__item').allTextContents()
  expect(optionTexts.some(text => text.includes('Default Organization'))).toBeTruthy()
  expect(optionTexts.some(text => text.includes('Administrator'))).toBeFalsy()
})

test('system user page should support create and delete flow', async ({ page, context }) => {
  const opened = await loginAndOpenUserPage(page, context)
  if (!opened) return

  const stamp = Date.now()
  const username = `e2e_user_${stamp}`
  const realName = `E2E User ${stamp}`
  const email = `e2e_${stamp}@example.com`

  await page.getByRole('button', { name: '新建用户' }).click()
  await expect(page.locator('.el-dialog')).toContainText('新建用户')

  await page.locator('.el-dialog input[placeholder="请输入用户名"]').fill(username)
  await page.locator('.el-dialog input[placeholder="请输入姓名"]').fill(realName)
  await page.locator('.el-dialog input[placeholder="请输入邮箱"]').fill(email)
  await page.locator('.el-dialog input[placeholder="请输入密码"]').fill('DataEase123456')

  const orgSelect = page.locator('.el-dialog .el-select').nth(0)
  await orgSelect.click()
  await page.locator('.el-select-dropdown__item').first().click()

  const createResp = page.waitForResponse(
    response => response.url().includes('/system/user/create') && response.status() === 200
  )
  await page.locator('.el-dialog').getByRole('button', { name: '确定' }).click()
  await createResp

  await expect(page.locator('.el-message--success')).toContainText('创建成功', { timeout: 10000 })
  const createdRow = page.locator('.el-table__body-wrapper .el-table__row').filter({ hasText: username }).first()
  await expect(createdRow).toBeVisible({ timeout: 10000 })

  const deleteResp = page.waitForResponse(
    response => response.url().includes('/system/user/delete/') && response.status() === 200
  )
  await createdRow.getByRole('button', { name: '删除' }).click()
  await page.locator('.el-message-box').getByRole('button', { name: '确定' }).click()
  await deleteResp
  await expect(page.locator('.el-message--success')).toContainText('删除成功', { timeout: 10000 })
})
