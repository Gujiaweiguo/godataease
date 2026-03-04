import { expect, test } from '@playwright/test'

// Note: These tests require backend service for authentication
// Run with: E2E_BASE_URL=http://localhost:8080 E2E_USERNAME=admin E2E_PASSWORD=your_password npm run e2e

const getCreateDatasourceButton = page => {
  return page
    .locator('button:has-text("新建数据源")')
    .or(page.locator('button:has-text("新建")'))
    .or(page.locator('button:has-text("创建")'))
    .or(page.locator('button:has-text("Create")'))
    .or(page.locator('button:has-text("New")'))
}

const loginByApiAndInjectToken = async (page, request) => {
  const username = process.env.E2E_USERNAME || 'admin'
  const password = process.env.E2E_PASSWORD || 'DataEase123456'

  const response = await request.post('/login/localLogin', {
    data: {
      name: username,
      pwd: password
    }
  })

  expect(response.ok()).toBeTruthy()
  const payload = await response.json()
  expect(payload.code).toBe('000000')
  expect(payload.data?.token).toBeTruthy()

  const now = Date.now()
  await page.goto('/')
  await page.evaluate(
    ({ token, currentTime }) => {
      const wrapCacheValue = value =>
        JSON.stringify({ c: currentTime, e: 253402300799000, v: JSON.stringify(value) })

      localStorage.setItem('user.token', wrapCacheValue(token))
      localStorage.setItem('user.exp', wrapCacheValue(0))
      localStorage.setItem('user.time', wrapCacheValue(currentTime))
      localStorage.setItem('app.desktop', wrapCacheValue(false))
    },
    { token: payload.data.token, currentTime: now }
  )
}

test.describe('Datasource Management', () => {
  test.beforeEach(async ({ page, request }) => {
    await loginByApiAndInjectToken(page, request)

    await expect
      .poll(async () => {
        return await page.evaluate(() => Object.keys(localStorage).includes('user.token'))
      }, { timeout: 10000 })
      .toBe(true)
  })

  test('SYS-SMK-005 @system-smoke should navigate to datasource list', async ({ page }) => {
    await page.goto('/#/module-datasource')

    await expect(page.locator('body')).toBeVisible({ timeout: 10000 })
    await expect(page.locator('text=数据源').or(page.locator('text=Datasource')).first()).toBeVisible({
      timeout: 10000
    })
  })

  test('SYS-SMK-006 @system-smoke should display create datasource button', async ({ page }) => {
    await page.goto('/#/module-datasource')

    const createButton = getCreateDatasourceButton(page)
    const hasCreatePermission = await createButton
      .first()
      .isVisible({ timeout: 3000 })
      .catch(() => false)

    test.skip(!hasCreatePermission, 'Current test account has no datasource manage permission')

    await expect(createButton.first()).toBeVisible({ timeout: 10000 })
  })

  test.fixme('should open create datasource dialog', async ({ page }) => {
    await page.goto('/#/module-datasource')

    // Click create button
    const createButton = getCreateDatasourceButton(page)

    await createButton.first().click()

    // Verify dialog or form appears
    await page.waitForTimeout(500)

    const dialogVisible = (await page.locator('.el-dialog').count()) > 0 || (await page.locator('.el-drawer').count()) > 0 || (await page.locator('form').count()) > 0

    expect(dialogVisible).toBeTruthy()
  })

  test.fixme('should show datasource types in creation dialog', async ({ page }) => {
    await page.goto('/#/module-datasource')

    const createButton = getCreateDatasourceButton(page)

    await createButton.first().click()
    await page.waitForTimeout(500)

    // Check for common datasource type options
    const mysqlOption = page.locator('text=MySQL').or(page.locator('text=mysql'))
    const hasDatasourceTypes = (await mysqlOption.count()) > 0

    // At minimum, should have some datasource type selection
    expect(hasDatasourceTypes || (await page.locator('.datasource-type').count()) > 0).toBeTruthy()
  })
})
