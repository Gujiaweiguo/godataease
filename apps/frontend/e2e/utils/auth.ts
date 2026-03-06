import type { Locator, Page } from '@playwright/test'

export const getLoginButton = (page: Page): Locator => {
  return page.locator('button:has-text("Login")').or(page.locator('button:has-text("登录")'))
}

export const getUsernameInput = (page: Page): Locator => {
  return page
    .locator(
      [
        'input[autocomplete="username"]',
        'input[placeholder*="账号"]',
        'input[placeholder*="Account"]',
        'input[placeholder*="邮箱"]',
        'input[placeholder*="Email"]',
        'input[placeholder*="ID"]',
        'input[type="text"]',
      ].join(', ')
    )
    .first()
}

export const getPasswordInput = (page: Page): Locator => {
  return page.locator('input[type="password"]').first()
}

export const hasLoginForm = async (page: Page): Promise<boolean> => {
  const passwordCount = await getPasswordInput(page).count()
  const loginButtonCount = await getLoginButton(page).count()
  return passwordCount > 0 && loginButtonCount > 0
}
