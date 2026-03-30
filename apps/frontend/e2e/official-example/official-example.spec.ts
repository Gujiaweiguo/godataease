import { expect, test } from '@playwright/test'
import type { Page } from '@playwright/test'
import { getLoginButton, getPasswordInput, getUsernameInput } from '../utils/auth'

type OfficialModuleCase = {
  name: string
  path: string
  urlContains: string
  moduleTexts: string[]
  officialTexts: string[]
  requireOfficial: boolean
}

const officialCases: OfficialModuleCase[] = [
  {
    name: 'workbench',
    path: '/#/workbranch/index',
    urlContains: '/#/workbranch/index',
    moduleTexts: ['工作台', 'Workbench'],
    officialTexts: ['官方示例'],
    requireOfficial: false,
  },
  {
    name: 'datasource',
    path: '/#/module-datasource',
    urlContains: '/#/module-datasource',
    moduleTexts: ['数据源', 'Datasource'],
    officialTexts: ['官方示例-数据源', 'Demo MySQL'],
    requireOfficial: true,
  },
  {
    name: 'dataset',
    path: '/#/module-dataset',
    urlContains: '/#/module-dataset',
    moduleTexts: ['数据集', 'Dataset'],
    officialTexts: ['官方示例-数据集', '官方示例数据集'],
    requireOfficial: false,
  },
  {
    name: 'dashboard',
    path: '/#/panel/index',
    urlContains: '/#/panel/index',
    moduleTexts: ['仪表板', 'Dashboard'],
    officialTexts: ['连锁茶饮销售看板', '官方示例仪表板'],
    requireOfficial: false,
  },
  {
    name: 'screen',
    path: '/#/screen/index',
    urlContains: '/#/screen/index',
    moduleTexts: ['数据大屏', 'Screen', 'DataV'],
    officialTexts: ['官方示例数据大屏'],
    requireOfficial: false,
  },
]

const containsAnyText = async (page: Page, texts: string[]) => {
  for (const text of texts) {
    if ((await page.getByText(text, { exact: false }).count()) > 0) {
      return true
    }
  }
  return false
}

const loginIfNeeded = async (page: Page) => {
  const username = process.env.E2E_USERNAME || 'admin'
  const password = process.env.E2E_PASSWORD || 'DataEase123456'

  await page.context().clearCookies()
  await page.goto('/')
  await page.evaluate(() => {
    localStorage.clear()
    sessionStorage.clear()
  })
  await page.goto('/#/login')

  const shouldLogin = /#\/login|\/login/.test(page.url())

  if (shouldLogin) {
    await expect(getUsernameInput(page)).toBeVisible({ timeout: 10000 })
    await expect(getPasswordInput(page)).toBeVisible({ timeout: 10000 })

    await getUsernameInput(page).fill(username)
    await getPasswordInput(page).fill(password)
    await getLoginButton(page).click()
    await page.waitForURL(/#\/workbranch|module-datasource|module-dataset|panel\/index|screen\/index/, {
      timeout: 20000,
    })
  }

  await expect(page.locator('body')).toBeVisible({ timeout: 10000 })
}

test.describe('Official Example E2E @system-smoke', () => {
  test.beforeEach(async ({ page }) => {
    await loginIfNeeded(page)
  })

  for (const moduleCase of officialCases) {
    test(`should open ${moduleCase.name} and verify official example resources`, async ({ page }) => {
      await page.goto(moduleCase.path)
      await expect(page.locator('body')).toBeVisible({ timeout: 10000 })

      await expect
        .poll(() => page.url().includes(moduleCase.urlContains), { timeout: 20000 })
        .toBeTruthy()

      await expect
        .poll(async () => containsAnyText(page, moduleCase.moduleTexts), { timeout: 20000 })
        .toBeTruthy()

      if (moduleCase.requireOfficial) {
        await expect
          .poll(async () => containsAnyText(page, moduleCase.officialTexts), { timeout: 20000 })
          .toBeTruthy()
      }
    })
  }
})
