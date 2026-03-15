import { expect, test } from '@playwright/test'
import type { Page } from '@playwright/test'
import { getLoginButton, getPasswordInput, getUsernameInput } from '../utils/auth'

const loginIfNeeded = async (page: Page) => {
  const username = process.env.E2E_USERNAME || 'admin'
  const password = process.env.E2E_PASSWORD || 'DataEase123456'

  await page.goto('/#/login')

  if (/#\/login|\/login/.test(page.url())) {
    await expect(getUsernameInput(page)).toBeVisible({ timeout: 10000 })
    await expect(getPasswordInput(page)).toBeVisible({ timeout: 10000 })
    await getUsernameInput(page).fill(username)
    await getPasswordInput(page).fill(password)
    await getLoginButton(page).click()
    await page.waitForURL(/#\/workbranch|panel\/index|screen\/index|module-datasource/, {
      timeout: 20000
    })
  }
}

const clickTreeAndWaitFindById = async (page: Page, itemName: RegExp) => {
  const node = page.getByRole('treeitem', { name: itemName }).first()
  await expect(node).toBeVisible({ timeout: 20000 })
  const responsePromise = page.waitForResponse(
    res => res.url().includes('/api/dataVisualization/findById') && res.request().method() === 'POST',
    { timeout: 20000 }
  )
  await node.click()
  const response = await responsePromise
  expect(response.status()).toBe(200)
}

test.describe('Official Example Open/Edit E2E', () => {
  test.beforeEach(async ({ page }) => {
    await loginIfNeeded(page)
  })

  test('dashboard official example click can load preview', async ({ page }) => {
    await page.goto('/#/panel/index')
    await expect(page.locator('body')).toBeVisible({ timeout: 10000 })
    await clickTreeAndWaitFindById(page, /连锁茶饮销售看板/)
    await expect(page.getByText('请在左侧选择仪表板')).toHaveCount(0, { timeout: 20000 })
    await expect(page.getByRole('button', { name: /编辑|Edit/ }).first()).toBeVisible({ timeout: 10000 })
  })

  test('screen official example click can load preview', async ({ page }) => {
    const pageErrors: string[] = []
    page.on('pageerror', error => {
      pageErrors.push(error.message)
    })

    await page.goto('/#/screen/index')
    await expect(page.locator('body')).toBeVisible({ timeout: 10000 })
    await clickTreeAndWaitFindById(page, /官方示例数据大屏/)
    await expect(page.getByText('请在左侧选择数据大屏')).toHaveCount(0, { timeout: 20000 })
    await expect(page.getByRole('button', { name: /编辑|Edit/ }).first()).toBeVisible({ timeout: 10000 })

    const runtimeErrors = pageErrors.filter(msg => /seniorStyleSetting|Cannot read properties/.test(msg))
    expect(runtimeErrors).toEqual([])
  })

  test('dataV create page drag-drop should not throw runtime error', async ({ page }) => {
    const pageErrors: string[] = []
    page.on('pageerror', error => {
      pageErrors.push(error.message)
    })

    await page.goto('/#/dvCanvas?opt=create')
    await expect(page.locator('#canvas-dv-outer')).toBeVisible({ timeout: 20000 })

    await page.evaluate(() => {
      const dropTarget = document.querySelector('#canvas-dv-outer') as HTMLElement | null
      if (!dropTarget) {
        return
      }
      const dataTransfer = new DataTransfer()
      dataTransfer.setData('id', 'VQuery&VQuery')
      dropTarget.dispatchEvent(
        new DragEvent('dragover', {
          bubbles: true,
          cancelable: true,
          dataTransfer
        })
      )
      dropTarget.dispatchEvent(
        new DragEvent('drop', {
          bubbles: true,
          cancelable: true,
          dataTransfer,
          clientX: 400,
          clientY: 260
        })
      )
    })
    await page.waitForTimeout(1000)

    const runtimeErrors = pageErrors.filter(msg => /Cannot read properties|undefined|null/.test(msg))
    expect(runtimeErrors).toEqual([])
  })
})
