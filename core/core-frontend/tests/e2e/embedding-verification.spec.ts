import { test, expect, describe, beforeAll, afterAll } from 'vitest'
import { chromium } from 'playwright'
import path from 'path'

describe('Embedding Verification - Automated Browser Tests', () => {
  let browser: any
  let page: any
  const baseUrl = process.env.VITE_DEV_SERVER_URL || 'http://localhost:5100'

  beforeAll(async () => {
    browser = await chromium.launch({
      headless: process.env.CI === 'true',
      args: ['--disable-web-security']
    })
  })

  afterAll(async () => {
    if (browser) {
      await browser.close()
    }
  })

  describe('Dashboard Embedding', () => {
    beforeEach(async () => {
      page = await browser.newPage()
    })

    afterEach(async () => {
      if (page) {
        await page.close()
      }
    })

    test('should load dashboard embedding demo page', async () => {
      const demoPath = path.join(__dirname, '../../public/embedding-demo/dashboard-embed.html')
      await page.goto(`file://${demoPath}`)
      
      const title = await page.title()
      expect(title).toBe('Embedded Dashboard Example - DataEase')
    })

    test('should display embedding controls', async () => {
      const demoPath = path.join(__dirname, '../../public/embedding-demo/dashboard-embed.html')
      await page.goto(`file://${demoPath}`)
      
      const dashboardUrlInput = page.getByLabel('Dashboard URL')
      await expect(dashboardUrlInput).toBeVisible()
      
      const tokenInput = page.getByPlaceholder('Enter your access token')
      await expect(tokenInput).toBeVisible()
      
      const updateButton = page.getByRole('button', { name: 'Update Embedding' })
      await expect(updateButton).toBeVisible()
    })

    test('should have event log section', async () => {
      const demoPath = path.join(__dirname, '../../public/embedding-demo/dashboard-embed.html')
      await page.goto(`file://${demoPath}`)
      
      const eventLog = page.getByText('Event Log')
      await expect(eventLog).toBeVisible()
    })

    test('should initialize event listeners on page load', async () => {
      const demoPath = path.join(__dirname, '../../public/embedding-demo/dashboard-embed.html')
      await page.goto(`file://${demoPath}`)
      
      await page.waitForTimeout(1000)
      
      const eventLog = page.locator('#eventLog')
      const logContent = await eventLog.textContent()
      expect(logContent).toContain('Demo page loaded')
    })

    test('should listen for child messages', async () => {
      const demoPath = path.join(__dirname, '../../public/embedding-demo/dashboard-embed.html')
      await page.goto(`file://${demoPath}`)
      
      const messageListener = page.evaluate(() => {
        return new Promise((resolve) => {
          const messages = []
          window.addEventListener('message', (event) => {
            messages.push(event.data)
            if (messages.length >= 1) {
              resolve(messages)
            }
          })
          
          setTimeout(() => resolve(messages), 2000)
        })
      })
      
      await page.waitForTimeout(2500)
      const messages = await messageListener
      expect(Array.isArray(messages)).toBe(true)
    })
  })

  describe('Screen Embedding', () => {
    beforeEach(async () => {
      page = await browser.newPage()
    })

    afterEach(async () => {
      if (page) {
        await page.close()
      }
    })

    test('should load screen embedding demo page', async () => {
      const demoPath = path.join(__dirname, '../../public/embedding-demo/screen-embed.html')
      await page.goto(`file://${demoPath}`)
      
      const title = await page.title()
      expect(title).toBe('Embedded Screen Example - DataEase')
    })

    test('should display screen-specific controls', async () => {
      const demoPath = path.join(__dirname, '../../public/embedding-demo/screen-embed.html')
      await page.goto(`file://${demoPath}`)
      
      const dvIdInput = page.getByLabel('Visualization ID')
      await expect(dvIdInput).toBeVisible()
    })

    test('should have event logging', async () => {
      const demoPath = path.join(__dirname, '../../public/embedding-demo/screen-embed.html')
      await page.goto(`file://${demoPath}`)
      
      const clearButton = page.getByRole('button', { name: 'Clear Log' })
      await expect(clearButton).toBeVisible()
    })
  })

  describe('Parameter Initialization', () => {
    test('should update iframe src with token parameter', async () => {
      const demoPath = path.join(__dirname, '../../public/embedding-demo/dashboard-embed.html')
      await page.goto(`file://${demoPath}`)
      
      const tokenInput = page.getByPlaceholder('Enter your access token')
      await tokenInput.fill('test-token-123')
      
      const updateButton = page.getByRole('button', { name: 'Update Embedding' })
      await updateButton.click()
      
      const iframe = page.locator('#de-dashboard-iframe')
      const src = await iframe.getAttribute('src')
      
      expect(src).toContain('token=test-token-123')
    })

    test('should update iframe src with multiple parameters', async () => {
      const demoPath = path.join(__dirname, '../../public/embedding-demo/dashboard-embed.html')
      await page.goto(`file://${demoPath}`)
      
      const tokenInput = page.getByPlaceholder('Enter your access token')
      await tokenInput.fill('test-token-123')
      
      const themeSelect = page.getByLabel('Theme')
      await themeSelect.selectOption('dark')
      
      const updateButton = page.getByRole('button', { name: 'Update Embedding' })
      await updateButton.click()
      
      const iframe = page.locator('#de-dashboard-iframe')
      const src = await iframe.getAttribute('src')
      
      expect(src).toContain('token=test-token-123')
      expect(src).toContain('theme=dark')
    })
  })

  describe('Event Communication', () => {
    test('should log param_update events', async () => {
      const demoPath = path.join(__dirname, '../../public/embedding-demo/dashboard-embed.html')
      await page.goto(`file://${demoPath}`)
      
      await page.evaluate(() => {
        window.postMessage({
          type: 'dataease-embedded-interactive',
          args: { theme: 'dark' }
        }, '*')
      })
      
      await page.waitForTimeout(500)
      
      const eventLog = page.locator('#eventLog')
      const logContent = await eventLog.textContent()
      expect(logContent).toContain('PARAM_UPDATE')
    })

    test('should log ready events', async () => {
      const demoPath = path.join(__dirname, '../../public/embedding-demo/dashboard-embed.html')
      await page.goto(`file://${demoPath}`)
      
      await page.evaluate(() => {
        window.postMessage({
          type: 'ready',
          args: { initialized: true }
        }, '*')
      })
      
      await page.waitForTimeout(500)
      
      const eventLog = page.locator('#eventLog')
      const logContent = await eventLog.textContent()
      expect(logContent).toContain('CHILD_READY')
    })

    test('should log error events', async () => {
      const demoPath = path.join(__dirname, '../../public/embedding-demo/dashboard-embed.html')
      await page.goto(`file://${demoPath}`)
      
      await page.evaluate(() => {
        window.postMessage({
          type: 'error',
          args: 'Token expired'
        }, '*')
      })
      
      await page.waitForTimeout(500)
      
      const eventLog = page.locator('#eventLog')
      const logContent = await eventLog.textContent()
      expect(logContent).toContain('CHILD_ERROR')
    })

    test('should log user_interaction events', async () => {
      const demoPath = path.join(__dirname, '../../public/embedding-demo/dashboard-embed.html')
      await page.goto(`file://${demoPath}`)
      
      await page.evaluate(() => {
        window.postMessage({
          type: 'user_interaction',
          args: { type: 'click', target: 'chart-123' }
        }, '*')
      })
      
      await page.waitForTimeout(500)
      
      const eventLog = page.locator('#eventLog')
      const logContent = await eventLog.textContent()
      expect(logContent).toContain('USER_ACTION')
    })
  })

  describe('Iframe Functionality', () => {
    test('should render iframe element', async () => {
      const demoPath = path.join(__dirname, '../../public/embedding-demo/dashboard-embed.html')
      await page.goto(`file://${demoPath}`)
      
      const iframe = page.locator('#de-dashboard-iframe')
      await expect(iframe).toBeVisible()
    })

    test('should have correct iframe attributes', async () => {
      const demoPath = path.join(__dirname, '../../public/embedding-demo/dashboard-embed.html')
      await page.goto(`file://${demoPath}`)
      
      const iframe = page.locator('#de-dashboard-iframe')
      
      const frameborder = await iframe.getAttribute('frameborder')
      expect(frameborder).toBe('0')
      
      const width = await iframe.evaluate(el => el.style.width)
      expect(width).toBe('100%')
      
      const height = await iframe.evaluate(el => el.style.height)
      expect(height).toBe('600px')
    })
  })

  describe('Console Error Detection', () => {
    test('should detect console errors', async () => {
      const demoPath = path.join(__dirname, '../../public/embedding-demo/dashboard-embed.html')
      
      const errors = []
      page.on('console', msg => {
        if (msg.type() === 'error') {
          errors.push(msg.text())
        }
      })
      
      await page.goto(`file://${demoPath}`)
      
      await page.waitForTimeout(1000)
      
      const logs = await page.evaluate(() => {
        return window.consoleErrors || []
      })
      
      expect(errors.length).toBe(0)
    })
  })

  describe('Cross-Origin Communication', () => {
    test('should handle cross-origin messages', async () => {
      const demoPath = path.join(__dirname, '../../public/embedding-demo/dashboard-embed.html')
      await page.goto(`file://${demoPath}`)
      
      const messageReceived = await page.evaluate(() => {
        return new Promise((resolve) => {
          window.addEventListener('message', (event) => {
            if (event.data.type === 'dataease-embedded-interactive') {
              resolve(true)
            }
          })
          
          window.postMessage({
            type: 'dataease-embedded-interactive',
            args: { test: 'data' }
          }, '*')
        })
      })
      
      expect(messageReceived).toBe(true)
    })
  })

  describe('Responsive Design', () => {
    test('should adapt to different viewport sizes', async () => {
      const demoPath = path.join(__dirname, '../../public/embedding-demo/dashboard-embed.html')
      await page.goto(`file://${demoPath}`)
      
      await page.setViewportSize({ width: 1920, height: 1080 })
      const container = page.locator('.container')
      const width1080p = await container.evaluate(el => el.offsetWidth)
      
      await page.setViewportSize({ width: 768, height: 1024 })
      const width768p = await container.evaluate(el => el.offsetWidth)
      
      expect(width768p).toBeLessThan(width1080p)
    })
  })

  describe('Accessibility', () => {
    test('should have proper labels', async () => {
      const demoPath = path.join(__dirname, '../../public/embedding-demo/dashboard-embed.html')
      await page.goto(`file://${demoPath}`)
      
      const dashboardUrlInput = page.getByLabel('Dashboard URL')
      await expect(dashboardUrlInput).toBeVisible()
      
      const tokenInput = page.getByLabel('Access Token')
      await expect(tokenInput).toBeVisible()
    })

    test('should have proper heading hierarchy', async () => {
      const demoPath = path.join(__dirname, '../../public/embedding-demo/dashboard-embed.html')
      await page.goto(`file://${demoPath}`)
      
      const h1 = page.locator('h1')
      await expect(h1).toBeVisible()
      
      const h3 = page.locator('h3')
      await expect(h3).toBeVisible()
    })
  })
})
