import type { BrowserContext, Page } from '@playwright/test'
import { expect, test } from '@playwright/test'
import { getLoginButton, getPasswordInput, getUsernameInput, hasLoginForm } from './utils/auth'

type RoleSummary = {
  roleId: number
  roleName: string
}

type MenuNode = {
  id: number
  children?: MenuNode[]
}

type PermissionFixture = {
  roles: RoleSummary[]
  menuTree: MenuNode[]
}

const loginAndOpenPermissionPage = async (page: Page, context: BrowserContext) => {
  await context.clearCookies()
  await page.goto('/')
  await page.evaluate(() => {
    localStorage.clear()
    sessionStorage.clear()
  })
  await page.goto('/#/login')

  if (!(await hasLoginForm(page))) {
    await expect(page.locator('body')).toBeVisible({ timeout: 10000 })
    return false
  }

  await expect(getUsernameInput(page)).toBeVisible({ timeout: 10000 })
  await expect(getPasswordInput(page)).toBeVisible({ timeout: 10000 })

  await getUsernameInput(page).fill(process.env.E2E_USERNAME || 'admin')
  await getPasswordInput(page).fill(process.env.E2E_PASSWORD || 'DataEase123456')
  await getLoginButton(page).click()
  await page.waitForURL((url: URL) => !url.toString().includes('/login'), { timeout: 20000 })

  await page.goto('/#/system/permission')
  await expect(page.locator('.permission-config')).toBeVisible({ timeout: 10000 })
  return true
}

const fetchPermissionFixture = async (page: Page): Promise<PermissionFixture> => {
  return page.evaluate(async () => {
    const token = localStorage.getItem('user.token')
    const headers = {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${token}`
    }

    const [roleRes, menuRes] = await Promise.all([
      fetch('/de2api/role/byCurOrg', {
        method: 'POST',
        headers,
        body: JSON.stringify({ current: 1, size: 100 })
      }),
      fetch('/de2api/auth/menuResource', {
        method: 'GET',
        headers
      })
    ])

    const roleData = await roleRes.json()
    const menuData = await menuRes.json()

    return {
      roles: Array.isArray(roleData?.data?.list) ? roleData.data.list : [],
      menuTree: Array.isArray(menuData?.data) ? menuData.data : []
    }
  })
}

const fetchRoleMenuIds = async (page: Page, roleId: number): Promise<number[]> => {
  return page.evaluate(async currentRoleId => {
    const token = localStorage.getItem('user.token')
    const rsp = await fetch(`/de2api/roleMenu/auth/${currentRoleId}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token}`
      }
    })
    const data = await rsp.json()
    return Array.isArray(data?.data?.menuIds) ? data.data.menuIds : []
  }, roleId)
}

const restoreRoleMenuIds = async (page: Page, roleId: number, menuIds: number[]) => {
  await page.evaluate(
    async ({ currentRoleId, currentMenuIds }) => {
      const token = localStorage.getItem('user.token')
      await fetch('/de2api/roleMenu/auth', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`
        },
        body: JSON.stringify({ roleId: currentRoleId, menuIds: currentMenuIds })
      })
    },
    { currentRoleId: roleId, currentMenuIds: menuIds }
  )
}

const collectLeafMenuIds = (nodes: MenuNode[]): number[] => {
  const leafIds: number[] = []

  const walk = (items: MenuNode[]) => {
    items.forEach(item => {
      if (Array.isArray(item.children) && item.children.length > 0) {
        walk(item.children)
      } else {
        leafIds.push(item.id)
      }
    })
  }

  walk(nodes)
  return leafIds
}

const chooseTargetRole = (roles: RoleSummary[]): RoleSummary | undefined => {
  return roles.find(role => !/admin|管理员/i.test(role.roleName)) || roles[0]
}

const selectRoleInUi = async (page: Page, roleName: string) => {
  await page.locator('.menu-permission .el-select').click()
  await page.locator('.el-select-dropdown__item').filter({ hasText: roleName }).first().click()
}

const menuCheckbox = (page: Page, menuId: number) =>
  page.locator(`.menu-permission .el-tree [data-key="${menuId}"] .el-checkbox`).first()

const isCheckboxChecked = async (page: Page, menuId: number) => {
  const classes = (await menuCheckbox(page, menuId).getAttribute('class')) || ''
  return classes.includes('is-checked')
}

test('permission menu config should save and echo after reload', async ({ page, context }) => {
  const loggedIn = await loginAndOpenPermissionPage(page, context)
  if (!loggedIn) return

  const fixture = await fetchPermissionFixture(page)
  const targetRole = chooseTargetRole(fixture.roles)
  expect(targetRole).toBeTruthy()

  const roleId = targetRole!.roleId
  const roleName = targetRole!.roleName
  const originalMenuIds = await fetchRoleMenuIds(page, roleId)
  const leafMenuIds = collectLeafMenuIds(fixture.menuTree)
  const addCandidate = leafMenuIds.find(id => !originalMenuIds.includes(id))
  const removeCandidate = leafMenuIds.find(id => originalMenuIds.includes(id))
  const targetMenuId = addCandidate || removeCandidate

  expect(targetMenuId).toBeTruthy()

  const expectedChecked = Boolean(addCandidate)

  try {
    await selectRoleInUi(page, roleName)
    await expect(menuCheckbox(page, targetMenuId!)).toBeVisible({ timeout: 10000 })

    const beforeChecked = await isCheckboxChecked(page, targetMenuId!)
    if (beforeChecked !== expectedChecked) {
      await menuCheckbox(page, targetMenuId!).click()
    }

    await page.locator('.menu-permission .footer').getByRole('button', { name: '保存' }).click()
    await expect(page.locator('.el-message--success')).toContainText('菜单授权成功', { timeout: 10000 })

    await page.reload()
    await expect(page.locator('.permission-config')).toBeVisible({ timeout: 10000 })
    await selectRoleInUi(page, roleName)
    await expect(menuCheckbox(page, targetMenuId!)).toBeVisible({ timeout: 10000 })
    await expect.poll(async () => isCheckboxChecked(page, targetMenuId!)).toBe(expectedChecked)
  } finally {
    await restoreRoleMenuIds(page, roleId, originalMenuIds)
  }
})
