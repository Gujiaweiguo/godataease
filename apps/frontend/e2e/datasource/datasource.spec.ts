import { expect, test } from '@playwright/test'

test.describe('Datasource Management', () => {
  test.beforeEach(async ({ page }) => {
    // Login first
    await page.goto('/')
    const username = process.env.E2E_USERNAME || 'admin'
    const password = process.env.E2E_PASSWORD || 'DataEase123456'

    await page.locator('input[type="text"]').first().fill(username)
    await page.locator('input[type="password"]').first().fill(password)

    const loginButton = page.locator('button:has-text("登录")').or(page.locator('button[type="submit"]'))
    await loginButton.click()

    // Wait for login to complete
    await page.waitForURL(/^(?!.*login).*/, { timeout: 10000 })
  })

  test('should navigate to datasource list', async ({ page }) => {
    // Navigate to datasource page
    await page.goto('/datasource')

    // Verify datasource page elements
    await expect(page.locator('text=数据源').or(page.locator('text=Datasource'))).toBeVisible({ timeout: 5000 })
  })

  test('should display create datasource button', async ({ page }) => {
    await page.goto('/datasource')

    // Look for create/new button
    const createButton = page.locator('button:has-text("新建")').or(page.locator('button:has-text("创建")')).or(page.locator('.el-button:has-text("新")'))

    await expect(createButton.first()).toBeVisible({ timeout: 5000 })
  })

  test('should open create datasource dialog', async ({ page }) => {
    await page.goto('/datasource')

    // Click create button
    const createButton = page.locator('button:has-text("新建")').or(page.locator('button:has-text("创建")')).or(page.locator('.el-button:has-text("新")'))

    await createButton.first().click()

    // Verify dialog or form appears
    await page.waitForTimeout(500)

    const dialogVisible = (await page.locator('.el-dialog').count()) > 0 || (await page.locator('.el-drawer').count()) > 0 || (await page.locator('form').count()) > 0

    expect(dialogVisible).toBeTruthy()
  })

  test('should show datasource types in creation dialog', async ({ page }) => {
    await page.goto('/datasource')

    const createButton = page.locator('button:has-text("新建")').or(page.locator('button:has-text("创建")')).or(page.locator('.el-button:has-text("新")'))

    await createButton.first().click()
    await page.waitForTimeout(500)

    // Check for common datasource type options
    const mysqlOption = page.locator('text=MySQL').or(page.locator('text=mysql'))
    const hasDatasourceTypes = (await mysqlOption.count()) > 0

    // At minimum, should have some datasource type selection
    expect(hasDatasourceTypes || (await page.locator('.datasource-type').count()) > 0).toBeTruthy()
  })
})
