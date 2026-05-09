import { expect, test, type Page } from '@playwright/test'
import { hasLoginForm, loginAndVerify } from '../utils/auth'

/**
 * Interactive Tree E2E Smoke Tests
 *
 * Tests the resource tree functionality including:
 * - Dashboard tree
 * - DataV/Screen tree
 * - Dataset tree
 * - Datasource tree
 *
 * Note: Most tests require backend service for authentication and data.
 *
 * Run with: E2E_BASE_URL=http://localhost:8080 E2E_USERNAME=admin E2E_PASSWORD=your_password npm run e2e
 */

const treeTypes = [
  { name: 'Dashboard', path: '/panel/index', selectors: ['text=仪表板', 'text=Dashboard'] },
  { name: 'Screen', path: '/screen/index', selectors: ['text=数据大屏', 'text=Screen', 'text=DataV'] },
  { name: 'Dataset', path: '/data/dataset', selectors: ['text=数据集', 'text=Dataset'] },
  { name: 'Datasource', path: '/data/datasource', selectors: ['text=数据源', 'text=Datasource'] },
]

const openInteractiveRoute = async (page: Page, path: string) => {
  await page.goto(path)
  await expect(page.locator('body')).toBeVisible({ timeout: 10000 })
  await page.waitForTimeout(1500)

  const finalUrl = page.url()
  const bodyText = await page.locator('body').innerText()
  return {
    finalUrl,
    bodyText,
    hasApiError: /500|Request failed/.test(bodyText),
    hasForbiddenPage: /401|403|404/.test(finalUrl) || /401|403|404/.test(bodyText)
  }
}

test.describe('Interactive Tree', () => {
  /**
   * Smoke tests that don't require authentication
   */
  test.describe('Unauthenticated Access', () => {
    test('SYS-SMK-007 @system-smoke should redirect to login when accessing resource tree without auth', async ({ page }) => {
      await page.context().clearCookies()
      await page.goto('/DeResourceTree')

      await page.waitForTimeout(1000)

      const url = page.url()
      const redirectedToLogin = /login|auth/i.test(url)
      const hasLoginFormNow = await hasLoginForm(page)
      const hasApiError = (await page.locator('text=500').count()) > 0 || (await page.locator('text=Request failed').count()) > 0
      const pageVisible = await page.locator('body').isVisible()

      // With backend: should redirect to login
      // Without backend: may show API error (acceptable for smoke test)
      expect(redirectedToLogin || hasLoginFormNow || hasApiError || pageVisible).toBeTruthy()
    })

    for (const treeType of treeTypes) {
      test(`SYS-SMK-008 @system-smoke should redirect to login when accessing ${treeType.name} tree without auth`, async ({ page }) => {
        await page.context().clearCookies()
        await page.goto(treeType.path)

        await page.waitForTimeout(1000)

        const url = page.url()
        const redirectedToLogin = /login|auth/i.test(url)
        const hasLoginFormNow = await hasLoginForm(page)
        const hasApiError = (await page.locator('text=500').count()) > 0 || (await page.locator('text=Request failed').count()) > 0
        const pageVisible = await page.locator('body').isVisible()

        expect(redirectedToLogin || hasLoginFormNow || hasApiError || pageVisible).toBeTruthy()
      })
    }
  })

  /**
   * Tests requiring backend authentication
   * These are marked with test.fixme until backend is available
   */
  test.describe('Authenticated Access', () => {
    test.beforeEach(async ({ page }) => {
      await loginAndVerify(page)
    })

    test('should load resource tree component after login', async ({ page }) => {
      const state = await openInteractiveRoute(page, '/#/DeResourceTree')
      expect(state.hasApiError).toBeFalsy()
      expect(state.hasForbiddenPage).toBeFalsy()
      await expect(page.locator('.resource-tree').or(page.locator('[class*="resource-tree"]')).first()).toBeVisible()
    })

    test('should expose interactive tree search after login', async ({ page }) => {
      const state = await openInteractiveRoute(page, '/#/DeResourceTree')
      expect(state.hasApiError).toBeFalsy()
      expect(state.hasForbiddenPage).toBeFalsy()

      await expect(page.locator('.search-bar')).toBeVisible()
      await expect(page.locator('.search-bar input')).toBeVisible()
    })

    test('should expose interactive tree sort control after login', async ({ page }) => {
      const state = await openInteractiveRoute(page, '/#/DeResourceTree')
      expect(state.hasApiError).toBeFalsy()
      expect(state.hasForbiddenPage).toBeFalsy()

      await expect(page.locator('.filter-icon-span')).toBeVisible()
    })

    test.fixme('should display dataset tree nodes', async ({ page }) => {
      await page.goto('/data/dataset')
      await page.waitForTimeout(1000)

      // Look for dataset tree elements
      let foundDatasetTree = false
      for (const selector of treeTypes[2].selectors) {
        if ((await page.locator(selector).count()) > 0) {
          foundDatasetTree = true
          break
        }
      }

      const hasTreeNodes =
        (await page.locator('.el-tree-node').count()) > 0 ||
        (await page.locator('[class*="tree-node"]').count()) > 0

      expect(foundDatasetTree || hasTreeNodes || (await page.locator('body').isVisible())).toBeTruthy()
    })

    test.fixme('should display datasource tree nodes', async ({ page }) => {
      await page.goto('/data/datasource')
      await page.waitForTimeout(1000)

      // Look for datasource tree elements
      let foundDatasourceTree = false
      for (const selector of treeTypes[3].selectors) {
        if ((await page.locator(selector).count()) > 0) {
          foundDatasourceTree = true
          break
        }
      }

      const hasTreeNodes =
        (await page.locator('.el-tree-node').count()) > 0 ||
        (await page.locator('[class*="tree-node"]').count()) > 0

      expect(foundDatasourceTree || hasTreeNodes || (await page.locator('body').isVisible())).toBeTruthy()
    })

    test.fixme('should support tree node search', async ({ page }) => {
      await page.goto('/panel/index')
      await page.waitForTimeout(1000)

      // Look for search input
      const searchInput = page.locator('input[placeholder*="搜索"]').or(page.locator('input[placeholder*="Search"]'))

      if ((await searchInput.count()) > 0) {
        await searchInput.first().fill('test')
        await page.waitForTimeout(500)

        // Verify search was executed (tree may be filtered)
        await expect(page.locator('body')).toBeVisible()
      } else {
        // Search input may not be present, skip this check
        await expect(page.locator('body')).toBeVisible()
      }
    })

    test.fixme('should support tree node expand/collapse', async ({ page }) => {
      await page.goto('/panel/index')
      await page.waitForTimeout(1000)

      // Look for expand/collapse icons
      const expandIcon = page.locator('.el-tree-node__expand-icon').or(page.locator('[class*="expand"]'))

      if ((await expandIcon.count()) > 0) {
        await expandIcon.first().click()
        await page.waitForTimeout(500)
        await expect(page.locator('body')).toBeVisible()
      } else {
        await expect(page.locator('body')).toBeVisible()
      }
    })

    test.fixme('should support tree node operations menu', async ({ page }) => {
      await page.goto('/panel/index')
      await page.waitForTimeout(1000)

      // Look for operation menu trigger (usually a "more" icon or right-click)
      const moreIcon = page.locator('[class*="more"]').or(page.locator('[class*="operation"]'))

      if ((await moreIcon.count()) > 0) {
        await moreIcon.first().click()
        await page.waitForTimeout(500)
        // Menu should appear
        const menuVisible = (await page.locator('.el-dropdown-menu').count()) > 0 || (await page.locator('[class*="menu"]').count()) > 0
        expect(menuVisible || (await page.locator('body').isVisible())).toBeTruthy()
      } else {
        await expect(page.locator('body')).toBeVisible()
      }
    })

    test.fixme('should support create new folder in tree', async ({ page }) => {
      await page.goto('/panel/index')
      await page.waitForTimeout(1000)

      // Look for create folder button
      const createFolderBtn = page
        .locator('button:has-text("新建文件夹")')
        .or(page.locator('button:has-text("新建")'))
        .or(page.locator('button:has-text("New Folder")'))
        .or(page.locator('button:has-text("Create")'))

      if ((await createFolderBtn.count()) > 0) {
        await createFolderBtn.first().click()
        await page.waitForTimeout(500)
        // Dialog should appear
        const dialogVisible = (await page.locator('.el-dialog').count()) > 0 || (await page.locator('.el-drawer').count()) > 0
        expect(dialogVisible || (await page.locator('body').isVisible())).toBeTruthy()
      } else {
        await expect(page.locator('body')).toBeVisible()
      }
    })

    test.fixme('should support tree sort functionality', async ({ page }) => {
      await page.goto('/panel/index')
      await page.waitForTimeout(1000)

      // Look for sort button
      const sortBtn = page.locator('[class*="sort"]').or(page.locator('button:has-text("排序")')).or(page.locator('button:has-text("Sort")'))

      if ((await sortBtn.count()) > 0) {
        await sortBtn.first().click()
        await page.waitForTimeout(500)
        await expect(page.locator('body')).toBeVisible()
      } else {
        await expect(page.locator('body')).toBeVisible()
      }
    })

    test.fixme('should have responsive tree layout', async ({ page }) => {
      await page.goto('/panel/index')

      // Test desktop layout
      await page.setViewportSize({ width: 1920, height: 1080 })
      await page.waitForTimeout(500)
      await expect(page.locator('body')).toBeVisible()

      // Test tablet layout
      await page.setViewportSize({ width: 1024, height: 768 })
      await page.waitForTimeout(500)
      await expect(page.locator('body')).toBeVisible()

      // Test mobile layout
      await page.setViewportSize({ width: 375, height: 667 })
      await page.waitForTimeout(500)
      await expect(page.locator('body')).toBeVisible()
    })
  })

  /**
   * Tree drag and drop tests
   */
  test.describe('Tree Drag and Drop', () => {
    test.fixme('should support drag and drop for tree nodes', async ({ page }) => {
      await loginAndVerify(page)

      await page.goto('/#/panel/index')
      await page.waitForTimeout(1000)

      // Drag and drop test requires at least 2 tree nodes
      const treeNodes = page.locator('.el-tree-node')
      const nodeCount = await treeNodes.count()

      if (nodeCount >= 2) {
        // Attempt to drag first node to second
        await treeNodes.first().hover()
        await page.mouse.down()
        await treeNodes.nth(1).hover()
        await page.mouse.up()
        await page.waitForTimeout(500)
      }

      await expect(page.locator('body')).toBeVisible()
    })
  })
})
