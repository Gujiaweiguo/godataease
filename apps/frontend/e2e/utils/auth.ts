import { expect, type Locator, type Page } from '@playwright/test'

export const getE2ECredentials = () => ({
  username: process.env.E2E_USERNAME || 'admin',
  password: process.env.E2E_PASSWORD || 'DataEase123456'
})

export const getLoginButton = (page: Page): Locator => {
  return page.locator('button:has-text("Login")').or(page.locator('button:has-text("登录")'))
}

export const getUsernameInput = (page: Page): Locator => {
  return page
    .locator(
      [
        '.login-form input:not([type="password"])',
        'input[autocomplete="username"]',
        'input[name="username"]',
        'input[placeholder*="用户名"]',
        'input[placeholder*="用户"]',
        'input[placeholder*="账号"]',
        'input[placeholder*="Account"]',
        'input[placeholder*="account"]',
        'input[placeholder*="邮箱"]',
        'input[placeholder*="Email"]',
        'input[placeholder*="email"]',
        'input[placeholder*="ID"]',
        'input[type="text"]',
        'input:not([type="password"])',
      ].join(', ')
    )
    .first()
}

export const getPasswordInput = (page: Page): Locator => {
  return page.locator('input[type="password"]').first()
}

export const hasLoginForm = async (page: Page): Promise<boolean> => {
  const usernameCount = await getUsernameInput(page).count()
  const passwordCount = await getPasswordInput(page).count()
  const loginButtonCount = await getLoginButton(page).count()
  return usernameCount > 0 && passwordCount > 0 && loginButtonCount > 0
}

export const clearAuthState = async (page: Page): Promise<void> => {
  await page.context().clearCookies()
  await page.goto('/')
  await page.evaluate(() => {
    localStorage.clear()
    sessionStorage.clear()
  })
}

export const loginWithValidCredentials = async (page: Page): Promise<void> => {
  const { username, password } = getE2ECredentials()

  await getUsernameInput(page).fill(username)
  await getPasswordInput(page).fill(password)
  await getLoginButton(page).click()
}

export const loginAndVerify = async (page: Page, loginPath = '/#/login'): Promise<void> => {
  await clearAuthState(page)
  await page.goto(loginPath)
  await expect(page.locator('body')).toBeVisible({ timeout: 10000 })
  await expect.poll(() => hasLoginForm(page), { timeout: 20000 }).toBeTruthy()

  await loginWithValidCredentials(page)
  await page.waitForURL(/#\/workbranch|data\/datasource|module-datasource/, { timeout: 20000 })

  const hasToken = await page.evaluate(() => Object.keys(localStorage).includes('user.token'))
  if (!hasToken) {
    throw new Error('Login failed: user.token not found in localStorage')
  }
}
