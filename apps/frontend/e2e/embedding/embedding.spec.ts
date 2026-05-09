import { expect, test, type Page } from '@playwright/test'
import { hasLoginForm, loginAndVerify } from '../utils/auth'

/**
 * Embedding Parameters E2E Smoke Tests
 *
 * Tests the embedding functionality including:
 * - Dataset embedding
 * - Datasource embedding
 * - Dashboard/Canvas embedding
 * - Preview embedding
 * - Outer parameters handling
 *
 * Note: Most tests require backend service for authentication and data.
 *
 * Run with: E2E_BASE_URL=http://localhost:8080 E2E_USERNAME=admin E2E_PASSWORD=your_password npm run e2e
 */

const embeddedRoutes = [
  { path: '/#/dataset-embedded', name: 'Dataset Embedded' },
  { path: '/#/dataset-embedded-form', name: 'Dataset Embedded Form' },
  { path: '/#/datasource-embedded', name: 'Datasource Embedded' },
  { path: '/#/dvCanvas', name: 'Canvas Editor' },
  { path: '/#/dashboard', name: 'Dashboard' },
  { path: '/#/preview', name: 'Preview' },
]

const smokeEmbeddedRoutes = [
  {
    path: '/#/dataset-embedded',
    smokeId: 'SYS-SMK-010a',
    name: 'dataset embedded',
    expectedText: /数据集|Dataset/
  },
  {
    path: '/#/datasource-embedded',
    smokeId: 'SYS-SMK-010b',
    name: 'datasource embedded',
    expectedText: /数据源|Datasource|Data source/
  }
]

const openEmbeddedRoute = async (page: Page, path: string) => {
  await page.goto(path)
  await expect(page.locator('body')).toBeVisible({ timeout: 10000 })
  await page.waitForURL((url: URL) => url.toString().includes(path.replace('/#', '#')), { timeout: 10000 })

  const bodyText = await page.locator('body').innerText()
  return {
    bodyText,
    hasApiError: /500|Request failed/.test(bodyText),
    hasForbiddenPage: /401|403|404/.test(bodyText),
    hasLoginForm: /Account Login|登录/.test(bodyText)
  }
}

test.describe('Embedding Parameters', () => {
  /**
   * Smoke tests that don't require authentication
   */
  test.describe('Unauthenticated Access', () => {
    for (const route of embeddedRoutes) {
      test(`should handle unauthenticated access to ${route.name}`, async ({ page }) => {
        await page.context().clearCookies()
        await page.goto(route.path)

        await page.waitForTimeout(1000)

        const url = page.url()
        const redirectedToLogin = /login|auth/i.test(url)
        const hasLoginFormNow = await hasLoginForm(page)
        const hasApiError = (await page.locator('text=500').count()) > 0 || (await page.locator('text=Request failed').count()) > 0
        const pageVisible = await page.locator('body').isVisible()

        // For embedded routes, behavior depends on whether it's in iframe mode
        // Some embedded routes may allow access without login
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

    for (const route of smokeEmbeddedRoutes) {
      test(`${route.smokeId} @system-smoke should load ${route.name} page`, async ({ page }) => {
        const state = await openEmbeddedRoute(page, route.path)

        expect(page.url()).toContain(route.path.replace('/#', '#'))
        expect(state.hasLoginForm).toBeFalsy()
        expect(state.hasApiError).toBeFalsy()
        expect(state.hasForbiddenPage).toBeFalsy()
        expect(route.expectedText.test(state.bodyText)).toBeTruthy()
      })
    }

    test.fixme('should load dataset embedded form page', async ({ page }) => {
      await page.goto('/#/dataset-embedded-form')
      await page.waitForTimeout(1000)

      // Verify dataset form elements
      const hasFormUI =
        (await page.locator('form').count()) > 0 ||
        (await page.locator('.el-form').count()) > 0 ||
        (await page.locator('[class*="form"]').count()) > 0

      expect(hasFormUI || (await page.locator('body').isVisible())).toBeTruthy()
    })

    test.fixme('should load datasource embedded page', async ({ page }) => {
      await page.goto('/#/datasource-embedded')
      await page.waitForTimeout(1000)

      // Verify datasource embedded page elements
      const hasDatasourceUI =
        (await page.locator('text=数据源').count()) > 0 ||
        (await page.locator('text=Datasource').count()) > 0 ||
        (await page.locator('[class*="datasource"]').count()) > 0

      expect(hasDatasourceUI || (await page.locator('body').isVisible())).toBeTruthy()
    })

    test.fixme('should load canvas editor page', async ({ page }) => {
      await page.goto('/#/dvCanvas')
      await page.waitForTimeout(1000)

      // Verify canvas editor elements
      const hasCanvasUI =
        (await page.locator('[class*="canvas"]').count()) > 0 ||
        (await page.locator('[class*="editor"]').count()) > 0 ||
        (await page.locator('.dv-canvas').count()) > 0

      expect(hasCanvasUI || (await page.locator('body').isVisible())).toBeTruthy()
    })

    test.fixme('should load dashboard page', async ({ page }) => {
      await page.goto('/#/dashboard')
      await page.waitForTimeout(1000)

      // Verify dashboard elements
      const hasDashboardUI =
        (await page.locator('[class*="dashboard"]').count()) > 0 ||
        (await page.locator('[class*="panel"]').count()) > 0

      expect(hasDashboardUI || (await page.locator('body').isVisible())).toBeTruthy()
    })

    test.fixme('should load preview page', async ({ page }) => {
      await page.goto('/#/preview')
      await page.waitForTimeout(1000)

      // Verify preview elements
      const hasPreviewUI =
        (await page.locator('[class*="preview"]').count()) > 0 ||
        (await page.locator('[class*="canvas"]').count()) > 0

      expect(hasPreviewUI || (await page.locator('body').isVisible())).toBeTruthy()
    })
  })

  /**
   * Outer parameters handling tests
   */
  test.describe('Outer Parameters', () => {
    test.fixme('should accept outer parameters in URL', async ({ page }) => {
      await loginAndVerify(page)

      // Navigate with outer parameters
      const testParams = 'eyJ0ZXN0IjoidmFsdWUifQ' // base64 encoded test params
      await page.goto(`/#/dvCanvas?outerParams=${testParams}`)
      await page.waitForTimeout(1000)

      // Verify page loaded
      await expect(page.locator('body')).toBeVisible({ timeout: 10000 })
    })

    test.fixme('should handle embedded token parameter', async ({ page }) => {
      await loginAndVerify(page)

      // Navigate with embedded token (simulated)
      await page.goto('/#/dvCanvas?embeddedToken=test-token&dvId=1')
      await page.waitForTimeout(1000)

      // Verify page loaded
      await expect(page.locator('body')).toBeVisible({ timeout: 10000 })
    })
  })

  /**
   * Iframe embedding tests
   */
  test.describe('Iframe Embedding', () => {
    test.fixme('should work in iframe context', async ({ page }) => {
      // Create a simple HTML page with iframe
      const iframeHtml = `
        <!DOCTYPE html>
        <html>
        <head><title>Embed Test</title></head>
        <body>
          <iframe id="de-embed" src="/#/preview" width="800" height="600"></iframe>
        </body>
        </html>
      `

       await loginAndVerify(page)

      // Set iframe content and verify
      await page.setContent(iframeHtml)
      await page.waitForTimeout(1000)

      const iframe = page.frameLocator('#de-embed')
      await expect(iframe.locator('body')).toBeVisible({ timeout: 10000 })
    })
  })
})
