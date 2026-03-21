import type { BrowserContext, Page } from '@playwright/test'
import { expect, test } from '@playwright/test'
import { getLoginButton, getPasswordInput, getUsernameInput, hasLoginForm } from './utils/auth'

type DatasetNode = {
  id: number
  name: string
  leaf?: boolean
  children?: DatasetNode[]
}

const loginAndOpenPermissionPage = async (page: Page, context: BrowserContext) => {
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

  await page.goto('/#/system/permission')
  await expect(page.locator('.permission-config')).toBeVisible({ timeout: 10000 })
  return true
}

const openDataPermissionTab = async (page: Page) => {
  await page.locator('.el-tabs__item').filter({ hasText: '行列权限' }).first().click()
  await expect(page.locator('.data-permission')).toBeVisible({ timeout: 10000 })
}

const fetchFirstDatasetLeaf = async (page: Page): Promise<DatasetNode | null> => {
  return page.evaluate(async () => {
    const token = localStorage.getItem('user.token')
    const rsp = await fetch('/de2api/datasetTree/tree', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token}`
      },
      body: JSON.stringify({ busiFlag: 'dataset', leaf: false, weight: 0 })
    })
    const data = await rsp.json()

    const walk = (nodes: DatasetNode[]): DatasetNode | null => {
      for (const node of nodes || []) {
        if (node.leaf) {
          return node
        }
        const found = walk(node.children || [])
        if (found) {
          return found
        }
      }
      return null
    }

    return walk(Array.isArray(data?.data) ? data.data : [])
  })
}

test('data permission tab should load row and column pager endpoints for a dataset', async ({ page, context }) => {
  const loggedIn = await loginAndOpenPermissionPage(page, context)
  if (!loggedIn) return

  await openDataPermissionTab(page)
  const dataset = await fetchFirstDatasetLeaf(page)
  if (!dataset) {
    await expect(page.locator('.data-permission')).toBeVisible()
    return
  }

  const rowPagerResponse = page.waitForResponse(
    response => response.url().includes(`/dataset/rowPermissions/pager/${dataset.id}/1/100`) && response.status() === 200
  )
  const columnPagerResponse = page.waitForResponse(
    response => response.url().includes(`/dataset/columnPermissions/pager/${dataset.id}/1/100`) && response.status() === 200
  )

  await page.locator('.data-permission .el-select').click()
  await page.locator('.el-select-dropdown__item').filter({ hasText: dataset.name }).first().click()

  await rowPagerResponse
  await columnPagerResponse

  await expect(page.locator('.data-permission .permission-tabs')).toBeVisible({ timeout: 10000 })
})
