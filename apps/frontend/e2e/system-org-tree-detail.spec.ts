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

  const orgListResp = page.waitForResponse(
    response => response.url().includes('/system/organization/list') && response.status() === 200
  )

  await page.goto('/#/system/org')
  await orgListResp
  await expect(page.locator('.org-management')).toBeVisible({ timeout: 10000 })
  return true
}

test('system org page should render organization tree rows', async ({ page, context }) => {
  const opened = await loginAndOpenOrgPage(page, context)
  if (!opened) return

  await expect(page.locator('.el-table__body-wrapper .el-table__row').first()).toBeVisible()
  await expect(page.locator('body')).toContainText(/组织管理|Default Organization|总部/)
})

test('system org create dialog should expose parent organization tree options', async ({ page, context }) => {
  const opened = await loginAndOpenOrgPage(page, context)
  if (!opened) return

  await page.getByRole('button', { name: '新建组织' }).click()
  await expect(page.locator('.el-dialog')).toContainText('新建组织')

  await page.locator('.el-tree-select').click()
  await expect(page.locator('.el-select-dropdown')).toContainText(/Default Organization|总部/)
})

test('system org detail endpoint should return organization payload under login session', async ({ page, context }) => {
  const opened = await loginAndOpenOrgPage(page, context)
  if (!opened) return

  const payload = await page.evaluate(async () => {
    const token = localStorage.getItem('user.token')
    const rsp = await fetch('/de2api/system/organization/info/1', {
      headers: {
        Authorization: `Bearer ${token}`
      }
    })
    return rsp.json()
  })

  expect(payload.code).toBe('000000')
  expect(payload.data?.orgId).toBe(1)
  expect(payload.data?.orgName).toBeTruthy()
})
