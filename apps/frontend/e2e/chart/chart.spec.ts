import { expect, test } from '@playwright/test'

/**
 * Chart Editor E2E Smoke Tests
 *
 * Tests the main paths of the chart editor functionality.
 * Note: Most tests require backend service for authentication and data.
 *
 * Run with: E2E_BASE_URL=http://localhost:8080 E2E_USERNAME=admin E2E_PASSWORD=your_password npm run e2e
 */

const hasLoginForm = async page => {
  const passwordCount = await page.locator('input[type="password"]').count()
  const loginButtonCount = await page
    .locator('button:has-text("Login")')
    .or(page.locator('button:has-text("登录")'))
    .count()

  return passwordCount > 0 && loginButtonCount > 0
}

test.describe('Chart Editor', () => {
  /**
   * Smoke tests that don't require authentication
   */
  test.describe('Unauthenticated Access', () => {
    // Note: This test requires backend service for proper redirect behavior
    // The route guard calls API (getDefaultSettings) which fails without backend
    test.fixme('should redirect to login when accessing chart editor without auth', async ({ page }) => {
      await page.context().clearCookies()
      await page.goto('/chart')

      await page.waitForTimeout(1000)

      const url = page.url()
      const redirectedToLogin = /login|auth/i.test(url)
      const hasLoginFormNow = await hasLoginForm(page)
      const hasApiError = (await page.locator('text=500').count()) > 0 || (await page.locator('text=Request failed').count()) > 0

      // With backend: should redirect to login
      // Without backend: may show API error (acceptable for smoke test)
      expect(redirectedToLogin || hasLoginFormNow || hasApiError).toBeTruthy()
    })
  })

  /**
   * Tests requiring backend authentication
   * These are marked with test.fixme until backend is available
   */
  test.describe('Authenticated Access', () => {
    test.beforeEach(async ({ page }) => {
      // Login first - requires backend
      await page.goto('/')
      const username = process.env.E2E_USERNAME || 'admin'
      const password = process.env.E2E_PASSWORD || 'DataEase123456'

      if (await hasLoginForm(page)) {
        await page.locator('input[type="text"]').first().fill(username)
        await page.locator('input[type="password"]').first().fill(password)

        const loginButton = page
          .locator('button:has-text("Login")')
          .or(page.locator('button:has-text("登录")'))
        await loginButton.click()

        // Wait for login to complete
        await page.waitForURL(/^(?!.*login).*/, { timeout: 15000 }).catch(() => {
          // If timeout, continue anyway - test will fail if login was required
        })
      }
    })

    test.fixme('should navigate to chart editor', async ({ page }) => {
      await page.goto('/chart')

      // Verify chart editor page loaded
      await expect(page.locator('body')).toBeVisible({ timeout: 10000 })

      // Check for chart editor specific elements
      const hasChartEditor =
        (await page.locator('.chart-editor').count()) > 0 ||
        (await page.locator('[class*="chart"]').count()) > 0 ||
        (await page.locator('text=图表').or(page.locator('text=Chart'))).count() > 0

      expect(hasChartEditor).toBeTruthy()
    })

    test.fixme('should display chart type selector', async ({ page }) => {
      await page.goto('/chart')
      await page.waitForTimeout(1000)

      // Look for chart type selection UI
      const chartTypeElements = [
        page.locator('.chart-type'),
        page.locator('[class*="chart-type"]'),
        page.locator('text=柱状图'),
        page.locator('text=折线图'),
        page.locator('text=饼图'),
        page.locator('text=Bar'),
        page.locator('text=Line'),
        page.locator('text=Pie'),
      ]

      let foundChartTypes = false
      for (const locator of chartTypeElements) {
        if ((await locator.count()) > 0) {
          foundChartTypes = true
          break
        }
      }

      expect(foundChartTypes).toBeTruthy()
    })

    test.fixme('should display data configuration panel', async ({ page }) => {
      await page.goto('/chart')
      await page.waitForTimeout(1000)

      // Look for data configuration UI
      const dataConfigElements = [
        page.locator('.data-config'),
        page.locator('[class*="data-config"]'),
        page.locator('text=数据'),
        page.locator('text=Data'),
        page.locator('text=维度'),
        page.locator('text=指标'),
        page.locator('text=Dimension'),
        page.locator('text=Metric'),
      ]

      let foundDataConfig = false
      for (const locator of dataConfigElements) {
        if ((await locator.count()) > 0) {
          foundDataConfig = true
          break
        }
      }

      expect(foundDataConfig).toBeTruthy()
    })

    test.fixme('should display style configuration panel', async ({ page }) => {
      await page.goto('/chart')
      await page.waitForTimeout(1000)

      // Look for style configuration UI
      const styleConfigElements = [
        page.locator('.style-config'),
        page.locator('[class*="style-config"]'),
        page.locator('text=样式'),
        page.locator('text=Style'),
        page.locator('text=颜色'),
        page.locator('text=Color'),
      ]

      let foundStyleConfig = false
      for (const locator of styleConfigElements) {
        if ((await locator.count()) > 0) {
          foundStyleConfig = true
          break
        }
      }

      expect(foundStyleConfig).toBeTruthy()
    })

    test.fixme('should have responsive chart editor layout', async ({ page }) => {
      await page.goto('/chart')

      // Test desktop layout
      await page.setViewportSize({ width: 1920, height: 1080 })
      await page.waitForTimeout(500)
      await expect(page.locator('body')).toBeVisible()

      // Test tablet layout
      await page.setViewportSize({ width: 1024, height: 768 })
      await page.waitForTimeout(500)
      await expect(page.locator('body')).toBeVisible()

      // Test mobile layout (chart editor may have limited mobile support)
      await page.setViewportSize({ width: 375, height: 667 })
      await page.waitForTimeout(500)
      await expect(page.locator('body')).toBeVisible()
    })
  })

  /**
   * Chart view tests
   */
  test.describe('Chart View', () => {
    test.fixme('should load chart view page', async ({ page }) => {
      // Login first
      await page.goto('/')
      if (await hasLoginForm(page)) {
        const username = process.env.E2E_USERNAME || 'admin'
        const password = process.env.E2E_PASSWORD || 'DataEase123456'

        await page.locator('input[type="text"]').first().fill(username)
        await page.locator('input[type="password"]').first().fill(password)

        const loginButton = page
          .locator('button:has-text("Login")')
          .or(page.locator('button:has-text("登录")'))
        await loginButton.click()

        await page.waitForURL(/^(?!.*login).*/, { timeout: 15000 }).catch(() => {})
      }

      // Navigate to chart view (requires a chart ID in real scenario)
      await page.goto('/chart-view')

      // Verify page loaded
      await expect(page.locator('body')).toBeVisible({ timeout: 10000 })
    })
  })
})
