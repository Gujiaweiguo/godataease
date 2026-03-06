import { expect, test } from '@playwright/test'

// Note: These tests require backend service for authentication
// Run with: E2E_BASE_URL=http://localhost:8080 E2E_USERNAME=admin E2E_PASSWORD=your_password npm run e2e

const getCreateDatasourceButton = page => {
  const textButton = page
    .locator('button:has-text("新建数据源")')
    .or(page.locator('button:has-text("新建")'))
    .or(page.locator('button:has-text("创建")'))
    .or(page.locator('button:has-text("Create")'))
    .or(page.locator('button:has-text("New")'))

  const iconButton = page.locator('.tree-header .icon-methods .custom-icon.btn').nth(1)

  return textButton.or(iconButton)
}

const loginWithValidCredentials = async page => {
  const username = process.env.E2E_USERNAME || 'admin'
  const password = process.env.E2E_PASSWORD || 'DataEase123456'

  await page.locator('input[type="text"]').first().fill(username)
  await page.locator('input[type="password"]').first().fill(password)

  const loginButton = page.locator('button:has-text("Login")').or(page.locator('button:has-text("登录")'))
  await loginButton.click()
}

const ensureLoggedIn = async page => {
  await page.goto('/#/login')

  await expect(page.locator('input[type="text"]').first()).toBeVisible({ timeout: 10000 })
  await expect(page.locator('input[type="password"]').first()).toBeVisible({ timeout: 10000 })

  await loginWithValidCredentials(page)

  // Wait for navigation to complete after login
  await page.waitForURL(/#\/workbranch|data\/datasource|module-datasource/, { timeout: 20000 })

  // Verify token exists in localStorage
  const hasToken = await page.evaluate(() => Object.keys(localStorage).includes('user.token'))
  if (!hasToken) {
    throw new Error('Login failed: user.token not found in localStorage')
  }
}

test.describe('Datasource Management', () => {
  test.beforeEach(async ({ page }) => {
    await ensureLoggedIn(page)
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

    // Wait for page to load
    await expect(page.locator('text=数据源').first()).toBeVisible({ timeout: 10000 })

    // Check for create button - "新建数据源" text button appears when tree is empty
    // When tree has data, the button is in tree header as icon button
    const textButton = page.locator('button:has-text("新建数据源")')
    const iconButton = page.locator('.tree-header .icon-methods .custom-icon.btn').nth(1)

    // Wait for either button to be visible
    let buttonFound = false
    try {
      await expect(textButton).toBeVisible({ timeout: 5000 })
      buttonFound = true
    } catch {
      try {
        await expect(iconButton).toBeVisible({ timeout: 5000 })
        buttonFound = true
      } catch {
        // Neither button found
      }
    }

    expect(buttonFound).toBeTruthy()
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
