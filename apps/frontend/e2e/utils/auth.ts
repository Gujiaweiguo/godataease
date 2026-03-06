import type { Locator, Page } from '@playwright/test'

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
