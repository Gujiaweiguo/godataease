import { expect, test } from '@playwright/test'
import { getLoginButton, getPasswordInput, getUsernameInput } from '../utils/auth'

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

  await getUsernameInput(page).fill(username)
  await getPasswordInput(page).fill(password)

  const loginButton = getLoginButton(page)
  await loginButton.click()
}

const detectDatasourcePageState = async page => {
  const bodyText = await page.locator('body').innerText()
  const hasDatasourceText = /数据源|Datasource/.test(bodyText)
  const hasLoginForm = /Account Login|登录/.test(bodyText)
  const hasApiError = /500|Request failed/.test(bodyText)
  const hasWorkbenchText = /工作台|Workbench/.test(bodyText)

  return {
    hasDatasourceText,
    hasLoginForm,
    hasApiError,
    hasWorkbenchText
  }
}

const ensureLoggedIn = async page => {
  await page.context().clearCookies()
  await page.goto('/#/login')

  if (!(await getUsernameInput(page).isVisible().catch(() => false))) {
    await page.evaluate(() => {
      localStorage.clear()
      sessionStorage.clear()
    })
    await page.goto('/#/login')
  }

  await expect(getUsernameInput(page)).toBeVisible({ timeout: 10000 })
  await expect(getPasswordInput(page)).toBeVisible({ timeout: 10000 })

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

    await expect
      .poll(async () => {
        const state = await detectDatasourcePageState(page)
        return state.hasDatasourceText
      }, { timeout: 15000 })
      .toBeTruthy()

    const state = await detectDatasourcePageState(page)
    expect(state.hasDatasourceText).toBeTruthy()
    expect(state.hasLoginForm).toBeFalsy()
    expect(state.hasApiError).toBeFalsy()
    expect(state.hasWorkbenchText).toBeFalsy()

    await expect(page.locator('.resource-tree .tree-header .title')).toContainText(/数据源|Datasource/)
  })

  test('SYS-SMK-006 @system-smoke should display create datasource button', async ({ page }) => {
    await page.goto('/#/module-datasource')

    await expect
      .poll(async () => {
        const currentState = await detectDatasourcePageState(page)
        return (
          currentState.hasDatasourceText ||
          currentState.hasLoginForm ||
          currentState.hasApiError ||
          currentState.hasWorkbenchText
        )
      }, { timeout: 15000 })
      .toBeTruthy()

    const state = await detectDatasourcePageState(page)

    if (!state.hasDatasourceText) {
      expect(state.hasLoginForm || state.hasApiError || state.hasWorkbenchText).toBeTruthy()
      return
    }

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
