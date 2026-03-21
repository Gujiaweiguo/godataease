import type { BrowserContext, Page } from '@playwright/test'
import { expect, test } from '@playwright/test'
import { getLoginButton, getPasswordInput, getUsernameInput, hasLoginForm } from './utils/auth'

const loginAndOpenOrgPage = async (page: Page, context: BrowserContext) => {
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

  await page.goto('/#/system/org')
  await expect(page.locator('body')).toBeVisible({ timeout: 10000 })
  return true
}

test('organization page should render returned rows instead of empty state', async ({ page, context }) => {
  const loggedIn = await loginAndOpenOrgPage(page, context)
  if (!loggedIn) return

  await page.waitForTimeout(2000)

  const rowCount = await page.locator('.el-table__body-wrapper .el-table__row').count()
  const emptyState = await page.locator('.el-table__empty-text').count()

  expect(emptyState).toBe(0)
  expect(rowCount).toBeGreaterThan(0)
})

test('organization page should support add-child then delete cleanup flow', async ({ page, context }) => {
  const loggedIn = await loginAndOpenOrgPage(page, context)
  if (!loggedIn) return

  const uniqueName = `E2E-ORG-${Date.now()}`
  const parentRow = page.locator('.el-table__body-wrapper .el-table__row').first()

  await expect(parentRow).toBeVisible({ timeout: 10000 })
  await parentRow.getByRole('button', { name: '添加子组织' }).click()

  await page.locator('input[placeholder="请输入组织名称"]').fill(uniqueName)
  await page.locator('textarea[placeholder="请输入组织描述"]').fill('org crud verification')
  await page.locator('.el-dialog__footer').getByRole('button', { name: '确定' }).click()

  await expect(page.locator('.el-message--success')).toContainText('创建成功', { timeout: 10000 })

  const expandIcon = page.locator('.el-table__body-wrapper .el-table__row').first().locator('.el-table__expand-icon')
  if ((await expandIcon.count()) > 0) {
    const expanded = await expandIcon.first().evaluate(node => node.classList.contains('el-table__expand-icon--expanded'))
    if (!expanded) {
      await expandIcon.first().click()
    }
  }

  const createdRow = page.locator('.el-table__body-wrapper .el-table__row').filter({ hasText: uniqueName })
  await expect(createdRow).toBeVisible({ timeout: 10000 })

  await createdRow.getByRole('button', { name: '删除' }).click()
  await page.locator('.el-message-box').getByRole('button', { name: '确定' }).click()
  await expect(page.locator('.el-message--success')).toContainText('删除成功', { timeout: 10000 })
  await expect(createdRow).toHaveCount(0, { timeout: 10000 })
})
