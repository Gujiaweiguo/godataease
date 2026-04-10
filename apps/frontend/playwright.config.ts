import { defineConfig, devices } from '@playwright/test'

const defaultLocalBaseURL = 'http://localhost:5173'

export default defineConfig({
  testDir: './e2e',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: [
    ['html', { outputFolder: 'playwright-report' }],
    ['json', { outputFile: 'playwright-report/results.json' }],
  ],
  use: {
    baseURL: process.env.E2E_BASE_URL || defaultLocalBaseURL,
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'], channel: 'chrome' },
    },
  ],
  webServer: process.env.CI || process.env.E2E_BASE_URL
    ? undefined
    : {
        command: 'npm run dev',
        url: defaultLocalBaseURL,
        reuseExistingServer: !process.env.CI,
        timeout: 120000,
      },
})
