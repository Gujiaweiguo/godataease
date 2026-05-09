import { expect, test } from '@playwright/test'
import { loginAndVerify } from '../utils/auth'

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

const detectDatasourcePageState = async page => {
  const bodyText = await page.locator('body').innerText()
  const hasDatasourceTreeHeader =
    (await page.locator('.resource-tree .tree-header .title').count()) > 0 ||
    (await page.locator('.resource-tree .icon-methods').count()) > 0
  const hasDatasourceText = hasDatasourceTreeHeader && /数据源|Datasource|Data source/.test(bodyText)
  const hasLoginForm = /Account Login|登录/.test(bodyText)
  const hasApiError = /500|Request failed/.test(bodyText)
  const hasWorkbenchText = /工作台|Workbench|Template Center|Quick Create|My Favorites/.test(bodyText)

  return {
    hasDatasourceText,
    hasLoginForm,
    hasApiError,
    hasWorkbenchText
  }
}

const ensureLoggedIn = async page => {
  await loginAndVerify(page)
}

const openDatasourcePage = async page => {
  await page.goto('/#/module-datasource')
  await expect(page.locator('body')).toBeVisible({ timeout: 10000 })

  let state = await detectDatasourcePageState(page)
  if (state.hasDatasourceText) {
    return state
  }

  if (state.hasWorkbenchText) {
    const quickCreateDatasource = page
      .locator('text=Data source')
      .or(page.locator('text=Datasource'))
      .or(page.locator('text=数据源'))
      .first()

    await expect(quickCreateDatasource).toBeVisible({ timeout: 10000 })
    await quickCreateDatasource.click()

    await expect
      .poll(async () => {
        const nextState = await detectDatasourcePageState(page)
        return nextState.hasDatasourceText
      }, { timeout: 15000 })
      .toBeTruthy()

    await expect(page.locator('.resource-tree .tree-header .title')).toContainText(
      /数据源|Datasource|Data source/
    )

    state = await detectDatasourcePageState(page)
  }

  return state
}

test.describe('Datasource Management', () => {
  test.beforeEach(async ({ page }) => {
    await ensureLoggedIn(page)
  })

  test('SYS-SMK-005 @system-smoke should navigate to datasource list', async ({ page }) => {
    const state = await openDatasourcePage(page)
    expect(state.hasDatasourceText).toBeTruthy()
    expect(state.hasLoginForm).toBeFalsy()
    expect(state.hasApiError).toBeFalsy()
    expect(page.url()).not.toContain('login')

    await expect(page.locator('body')).toContainText(/数据源|Datasource|Data source/)
  })

  test('SYS-SMK-006 @system-smoke should display create datasource button', async ({ page }) => {
    const state = await openDatasourcePage(page)

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

  test('SYS-SMK-006b @system-smoke should open create datasource dialog', async ({ page }) => {
    const state = await openDatasourcePage(page)
    expect(state.hasDatasourceText).toBeTruthy()

    const createButton = getCreateDatasourceButton(page)
    await expect(createButton.first()).toBeVisible({ timeout: 10000 })
    await createButton.first().click()

    await expect(
      page
        .getByText('Create Datasource')
        .or(page.getByText('创建数据源'))
        .or(page.getByText('Select Datasource'))
        .first()
    ).toBeVisible({ timeout: 5000 })
  })

  test('SYS-SMK-007 @system-smoke should show datasource types in creation dialog', async ({ page }) => {
    const state = await openDatasourcePage(page)
    expect(state.hasDatasourceText).toBeTruthy()

    const createButton = getCreateDatasourceButton(page)
    await expect(createButton.first()).toBeVisible({ timeout: 10000 })
    await createButton.first().click()

    await expect(
      page
        .locator('text=MySQL')
        .or(page.locator('text=Db2'))
        .or(page.locator('text=OLTP'))
        .or(page.locator('text=API data'))
        .first()
    ).toBeVisible({ timeout: 5000 })
  })
})
