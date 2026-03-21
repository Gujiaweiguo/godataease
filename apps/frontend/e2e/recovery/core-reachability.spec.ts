import { expect, test } from '@playwright/test'
import { getLoginButton, getPasswordInput, getUsernameInput, hasLoginForm } from '../utils/auth'

const loginWithValidCredentials = async page => {
  const username = process.env.E2E_USERNAME || 'admin'
  const password = process.env.E2E_PASSWORD || 'DataEase123456'

  await getUsernameInput(page).fill(username)
  await getPasswordInput(page).fill(password)
  await getLoginButton(page).click()
}

const loginAndWaitForAppShell = async (page, context) => {
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
  await loginWithValidCredentials(page)
  await page.waitForURL(url => !url.toString().includes('/login'), { timeout: 20000 })
  return true
}

const assertHealthyProtectedRoute = async (page, route: string) => {
  await page.goto(`/#${route}`)
  await page.waitForTimeout(1500)
  const finalUrl = page.url()
  expect(finalUrl.includes('/401')).toBeFalsy()
  expect(finalUrl.includes('/404')).toBeFalsy()
  await expect(page.locator('body')).toBeVisible()
}

const fetchLeafIdFromTree = async (page, path: string, payload: Record<string, unknown>) => {
  return page.evaluate(
    async ({ path, payload }) => {
      const token = localStorage.getItem('user.token')
      const rsp = await fetch(path, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`
        },
        body: JSON.stringify(payload)
      })
      const data = await rsp.json()
      const walk = nodes => {
        for (const node of nodes || []) {
          if (node.leaf) return node.id
          const hit = walk(node.children)
          if (hit) return hit
        }
        return null
      }
      const list = Array.isArray(data?.data) ? data.data : data?.data?.children || []
      return walk(list)
    },
    { path, payload }
  )
}

test.describe('Core feature reachability', () => {
  test('should keep RBAC admin routes reachable after login', async ({ page, context }) => {
    const loggedIn = await loginAndWaitForAppShell(page, context)
    if (!loggedIn) return

    for (const route of ['/system/user', '/system/role', '/system/org', '/system/menu', '/system/permission']) {
      await assertHealthyProtectedRoute(page, route)
    }
  })

  test('should keep BI entry routes reachable after login', async ({ page, context }) => {
    const loggedIn = await loginAndWaitForAppShell(page, context)
    if (!loggedIn) return

    for (const route of ['/module-datasource', '/module-dataset', '/dashboard', '/dvCanvas']) {
      await assertHealthyProtectedRoute(page, route)
    }
  })

  test('should keep modify password route reachable after login', async ({ page, context }) => {
    const loggedIn = await loginAndWaitForAppShell(page, context)
    if (!loggedIn) return

    await assertHealthyProtectedRoute(page, '/mine/modify-pwd')
  })

  test('should not misclassify help and about routes as 401 or 404 after login', async ({ page, context }) => {
    const loggedIn = await loginAndWaitForAppShell(page, context)
    if (!loggedIn) return

    for (const route of ['/mine/about', '/help/doc']) {
      await assertHealthyProtectedRoute(page, route)
    }
  })

  test('should keep workbranch quick create enabled for admin bootstrap', async ({ page, context }) => {
    const loggedIn = await loginAndWaitForAppShell(page, context)
    if (!loggedIn) return

    await page.goto('/#/workbranch')
    await page.waitForTimeout(2000)

    const quickCreateItems = page.locator('.quick-creation .item').filter({ hasNot: page.locator('.template-create') })
    await expect(quickCreateItems.first()).toBeVisible()
    const disabledCount = await page.locator('.quick-creation .item.quick-create-disabled').count()
    expect(disabledCount).toBeLessThan(4)
  })

  test('should not redirect logout flow to 401 when exiting from mine routes', async ({ page, context }) => {
    const loggedIn = await loginAndWaitForAppShell(page, context)
    if (!loggedIn) return

    await page.goto('/#/mine/about')
    await page.waitForTimeout(1000)

    const accountOperator = page.locator('.top-info-container').first()
    await accountOperator.click()
    await page.locator('.uinfo-footer .uinfo-main-item').click()

    await page.waitForURL(url => url.toString().includes('/login'), { timeout: 20000 })
    const finalUrl = page.url()
    expect(finalUrl.includes('/401')).toBeFalsy()
    expect(finalUrl.includes('/login?redirect=/workbranch')).toBeTruthy()
  })

  test('should open datasource edit view for an existing datasource', async ({ page, context }) => {
    const loggedIn = await loginAndWaitForAppShell(page, context)
    if (!loggedIn) return

    const datasourceId = await fetchLeafIdFromTree(page, '/api/datasource/tree', {})
    expect(datasourceId).toBeTruthy()

    await page.goto(`/#/data/datasource?id=${datasourceId}`)
    await page.waitForTimeout(2000)

    const finalUrl = page.url()
    expect(finalUrl.includes('/401')).toBeFalsy()
    expect(finalUrl.includes('/404')).toBeFalsy()
    await expect(page.locator('body')).toBeVisible()
  })

  test('should open dataset edit view for an existing dataset', async ({ page, context }) => {
    const loggedIn = await loginAndWaitForAppShell(page, context)
    if (!loggedIn) return

    const datasetId = await fetchLeafIdFromTree(page, '/api/datasetTree/tree', { busiFlag: 'dataset' })
    expect(datasetId).toBeTruthy()

    const pageErrors: string[] = []
    page.on('pageerror', err => pageErrors.push(String(err)))

    await page.goto(`/#/module-dataset?id=${datasetId}`)
    await page.waitForTimeout(2500)

    const finalUrl = page.url()
    expect(finalUrl.includes('/401')).toBeFalsy()
    expect(finalUrl.includes('/404')).toBeFalsy()
    expect(pageErrors.length).toBe(0)
    await expect(page.locator('.dataset-manage .resource-tree .tree-header .title')).toContainText(/数据集|Dataset/)
    await expect(page.locator('.dataset-content .dataset-info')).toBeVisible()
  })

  test('should render dataset preview content for an existing dataset', async ({ page, context }) => {
    const loggedIn = await loginAndWaitForAppShell(page, context)
    if (!loggedIn) return

    const datasetId = await fetchLeafIdFromTree(page, '/api/datasetTree/tree', { busiFlag: 'dataset' })
    expect(datasetId).toBeTruthy()

    const pageErrors: string[] = []
    page.on('pageerror', err => pageErrors.push(String(err)))

    await page.goto(`/#/module-dataset?id=${datasetId}`)
    await page.waitForTimeout(3000)

    const finalUrl = page.url()
    expect(finalUrl.includes('/401')).toBeFalsy()
    expect(finalUrl.includes('/404')).toBeFalsy()
    expect(pageErrors.length).toBe(0)

    await expect(page.locator('.dataset-content .dataset-info')).toBeVisible()
    await expect(page.locator('.dataset-preview_table')).toBeVisible()

    const previewHeaderCount = await page.locator('.dataset-preview_table .el-table__header th').count()
    const previewRowCount = await page.locator('.dataset-preview_table .el-table__body tbody tr').count()
    const emptyStateVisible = await page.locator('.dataset-preview_table .empty-background').isVisible().catch(() => false)

    expect(previewHeaderCount).toBeGreaterThan(0)
    expect(previewRowCount > 0 || emptyStateVisible).toBeTruthy()
  })

  test('should open dashboard edit view without invalid tree payload error', async ({ page, context }) => {
    const loggedIn = await loginAndWaitForAppShell(page, context)
    if (!loggedIn) return

    const dashboardId = await fetchLeafIdFromTree(page, '/api/dataVisualization/tree', { busiFlag: 'dashboard' })
    expect(dashboardId).toBeTruthy()

    const pageErrors: string[] = []
    const detailSuccessResponses: string[] = []
    page.on('pageerror', err => pageErrors.push(String(err)))
    page.on('response', async response => {
      if (!response.url().includes('/dataVisualization/findById') || response.status() >= 400) return
      const json = await response.json().catch(() => null)
      if (json?.code === '000000') {
        detailSuccessResponses.push(response.url())
      }
    })

    await page.goto(`/#/dashboard?resourceId=${dashboardId}`)
    await page.waitForTimeout(2500)

    const finalUrl = page.url()
    expect(finalUrl.includes('/401')).toBeFalsy()
    expect(finalUrl.includes('/404')).toBeFalsy()
    expect(pageErrors.some(err => err.includes('Invalid tree payload'))).toBeFalsy()
    expect(detailSuccessResponses.length).toBeGreaterThan(0)
    await expect(page.locator('body')).toBeVisible()
  })

  test('should open dashboard preview view with consumable discovery layout', async ({ page, context }) => {
    const loggedIn = await loginAndWaitForAppShell(page, context)
    if (!loggedIn) return

    const pageErrors: string[] = []
    page.on('pageerror', err => pageErrors.push(String(err)))

    await page.goto('/#/dashboardPreview')
    await page.waitForTimeout(2500)

    const finalUrl = page.url()
    expect(finalUrl.includes('/401')).toBeFalsy()
    expect(finalUrl.includes('/404')).toBeFalsy()
    expect(pageErrors.some(err => err.includes('Invalid tree payload'))).toBeFalsy()

    await expect(page.locator('.dv-preview')).toBeVisible()
    await expect(page.locator('.resource-area')).toBeVisible()
    await expect(page.locator('.preview-area')).toBeVisible()
    await expect(page.locator('.preview-area')).not.toHaveClass(/no-data/)
  })

  test('should open big-screen preview view with consumable discovery layout', async ({ page, context }) => {
    const loggedIn = await loginAndWaitForAppShell(page, context)
    if (!loggedIn) return

    const pageErrors: string[] = []
    const detailSuccessResponses: string[] = []
    page.on('pageerror', err => pageErrors.push(String(err)))
    page.on('response', async response => {
      if (!response.url().includes('/dataVisualization/findById') || response.status() >= 400) return
      const json = await response.json().catch(() => null)
      if (json?.code === '000000') {
        detailSuccessResponses.push(response.url())
      }
    })

    await page.goto('/#/previewShow')
    await page.waitForTimeout(2500)

    const finalUrl = page.url()
    expect(finalUrl.includes('/401')).toBeFalsy()
    expect(finalUrl.includes('/404')).toBeFalsy()
    expect(pageErrors.length).toBe(0)
    expect(detailSuccessResponses.length).toBeGreaterThan(0)

    await expect(page.locator('.dv-preview')).toBeVisible()
    await expect(page.locator('.resource-area')).toBeVisible()
    await expect(page.locator('.preview-area')).toBeVisible()
    await expect(page.locator('.preview-area')).not.toHaveClass(/no-data/)
  })

  test('should open big-screen edit view with consumable detail payload', async ({ page, context }) => {
    const loggedIn = await loginAndWaitForAppShell(page, context)
    if (!loggedIn) return

    const screenId = await fetchLeafIdFromTree(page, '/api/dataVisualization/tree', { busiFlag: 'screen' })
    expect(screenId).toBeTruthy()

    const pageErrors: string[] = []
    const detailSuccessResponses: string[] = []
    page.on('pageerror', err => pageErrors.push(String(err)))
    page.on('response', async response => {
      if (!response.url().includes('/dataVisualization/findById') || response.status() >= 400) return
      const json = await response.json().catch(() => null)
      if (json?.code === '000000') {
        detailSuccessResponses.push(response.url())
      }
    })

    await page.goto(`/#/dvCanvas?dvId=${screenId}`)
    await page.waitForTimeout(3000)

    const finalUrl = page.url()
    expect(finalUrl.includes('/401')).toBeFalsy()
    expect(finalUrl.includes('/404')).toBeFalsy()
    expect(pageErrors.some(err => err.includes('Invalid tree payload'))).toBeFalsy()
    expect(detailSuccessResponses.length).toBeGreaterThan(0)

    await expect(page.locator('#canvas-dv-outer')).toBeVisible()
  })

  test('should keep audit page entry and detail flow reachable without auth or not-found fallback', async ({ page, context }) => {
    const loggedIn = await loginAndWaitForAppShell(page, context)
    if (!loggedIn) return

    const pageErrors: string[] = []
    page.on('pageerror', err => pageErrors.push(String(err)))

    await page.goto('/#/audit')
    await page.waitForTimeout(2500)

    const finalUrl = page.url()
    expect(finalUrl.includes('/401')).toBeFalsy()
    expect(finalUrl.includes('/404')).toBeFalsy()
    expect(pageErrors.length).toBe(0)

    await expect(page.locator('.audit-log-management')).toBeVisible()
    await expect(page.locator('.audit-log-management .el-table')).toBeVisible()

    const detailButtons = page.getByRole('button', { name: '详情' })
    const detailCount = await detailButtons.count()
    expect(detailCount).toBeGreaterThan(0)

    await detailButtons.first().click()
    await page.waitForTimeout(500)

    const postClickUrl = page.url()
    expect(postClickUrl.includes('/401')).toBeFalsy()
    expect(postClickUrl.includes('/404')).toBeFalsy()
    expect(pageErrors.length).toBe(0)
  })

  test('should switch language without 404 and persist selected locale', async ({ page, context }) => {
    const loggedIn = await loginAndWaitForAppShell(page, context)
    if (!loggedIn) return

    const errors: string[] = []
    page.on('response', async response => {
      if (response.url().includes('switchLanguage') && response.status() >= 400) {
        errors.push(`switchLanguage:${response.status()}`)
      }
    })

    await page.locator('.top-info-container').first().click()
    await page.waitForTimeout(500)
    await page.locator('.uinfo-main-item .about-parent').first().hover()
    await page.waitForTimeout(500)

    const englishOption = page.locator('.language-item').filter({ hasText: 'English' }).first()
    await englishOption.click()
    await page.waitForTimeout(2000)

    expect(errors).toEqual([])
    const storedLanguage = await page.evaluate(() => localStorage.getItem('user.language'))
    expect(storedLanguage).toBe('en')
  })

  test('should switch language from desktop setting entry without interaction failure', async ({ page, context }) => {
    await context.clearCookies()
    await page.goto('/')
    await page.evaluate(() => {
      localStorage.clear()
      sessionStorage.clear()
      localStorage.setItem('app.desktop', '1')
    })
    await page.goto('/#/login')

    if (!(await hasLoginForm(page))) return

    await expect(getUsernameInput(page)).toBeVisible({ timeout: 10000 })
    await expect(getPasswordInput(page)).toBeVisible({ timeout: 10000 })
    await loginWithValidCredentials(page)
    await page.waitForURL(url => !url.toString().includes('/login'), { timeout: 20000 })

    await page.locator('.preview-download_icon').last().click()
    await page.waitForTimeout(500)

    const englishOption = page.locator('.language-item').filter({ hasText: 'English' }).first()
    await englishOption.click()
    await page.waitForTimeout(2000)

    const storedLanguage = await page.evaluate(() => localStorage.getItem('user.language'))
    expect(storedLanguage).toBe('en')
  })

  test('should expose language switching on system header in desktop mode', async ({ page, context }) => {
    await context.clearCookies()
    await page.goto('/')
    await page.evaluate(() => {
      localStorage.clear()
      sessionStorage.clear()
      localStorage.setItem('app.desktop', '1')
    })
    await page.goto('/#/login')

    if (!(await hasLoginForm(page))) return

    await expect(getUsernameInput(page)).toBeVisible({ timeout: 10000 })
    await expect(getPasswordInput(page)).toBeVisible({ timeout: 10000 })
    await loginWithValidCredentials(page)
    await page.waitForURL(url => !url.toString().includes('/login'), { timeout: 20000 })

    await page.goto('/#/sys-setting/parameter')
    await page.waitForTimeout(1500)

    await page.locator('.preview-download_icon').last().click()
    await page.waitForTimeout(500)
    const englishOption = page.locator('.language-item').filter({ hasText: 'English' }).first()
    await expect(englishOption).toBeVisible()
    await englishOption.click()
    await page.waitForTimeout(2000)

    const storedLanguage = await page.evaluate(() => localStorage.getItem('user.language'))
    expect(storedLanguage).toBe('en')
  })

  test('should still switch language after logout and re-login', async ({ page, context }) => {
    const loggedIn = await loginAndWaitForAppShell(page, context)
    if (!loggedIn) return

    await page.locator('.top-info-container').first().click()
    await page.waitForTimeout(500)
    await page.locator('.uinfo-main-item .about-parent').first().hover()
    await page.waitForTimeout(500)
    await page.locator('.language-item').filter({ hasText: 'English' }).first().click()
    await page.waitForTimeout(2000)

    let storedLanguage = await page.evaluate(() => localStorage.getItem('user.language'))
    expect(storedLanguage).toBe('en')

    await page.locator('.top-info-container').first().click()
    await page.waitForTimeout(500)
    await page.locator('.uinfo-footer .uinfo-main-item').click()
    await page.waitForURL(url => url.toString().includes('/login'), { timeout: 20000 })

    await expect(getUsernameInput(page)).toBeVisible({ timeout: 10000 })
    await expect(getPasswordInput(page)).toBeVisible({ timeout: 10000 })
    await loginWithValidCredentials(page)
    await page.waitForURL(url => !url.toString().includes('/login'), { timeout: 20000 })

    await page.locator('.top-info-container').first().click()
    await page.waitForTimeout(500)
    await page.locator('.uinfo-main-item .about-parent').first().hover()
    await page.waitForTimeout(500)
    await page.locator('.language-item').filter({ hasText: /简体中文|Chinese/i }).first().click()
    await page.waitForTimeout(2000)

    storedLanguage = await page.evaluate(() => localStorage.getItem('user.language'))
    expect(storedLanguage).toBe('zh-CN')
  })
})
