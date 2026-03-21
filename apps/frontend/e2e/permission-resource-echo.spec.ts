import type { BrowserContext, Page } from '@playwright/test'
import { expect, test } from '@playwright/test'
import { getLoginButton, getPasswordInput, getUsernameInput, hasLoginForm } from './utils/auth'

type RoleSummary = {
  roleId: number
  roleName: string
}

type PermissionNode = {
  permId: number
  parentId?: number | null
}

type PermissionFixture = {
  permissions: PermissionNode[]
  roles: RoleSummary[]
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

const openResourcePermissionTab = async (page: Page) => {
  await page.locator('.el-tabs__item').filter({ hasText: '资源权限' }).first().click()
  await expect(page.locator('.resource-permission')).toBeVisible({ timeout: 10000 })
}

const fetchPermissionFixture = async (page: Page): Promise<PermissionFixture> => {
  return page.evaluate(async () => {
    const token = localStorage.getItem('user.token')
    const headers = {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${token}`
    }

    const [roleRes, permissionRes] = await Promise.all([
      fetch('/de2api/role/byCurOrg', {
        method: 'POST',
        headers,
        body: JSON.stringify({ current: 1, size: 100 })
      }),
      fetch('/de2api/auth/busiResource/1', {
        method: 'GET',
        headers
      })
    ])

    const roleData = await roleRes.json()
    const permissionData = await permissionRes.json()

    return {
      permissions: Array.isArray(permissionData?.data) ? permissionData.data : [],
      roles: Array.isArray(roleData?.data?.list) ? roleData.data.list : []
    }
  })
}

const fetchRolePermissionIds = async (page: Page, roleId: number): Promise<number[]> => {
  return page.evaluate(async currentRoleId => {
    const token = localStorage.getItem('user.token')
    const rsp = await fetch('/de2api/auth/busiPermission', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token}`
      },
      body: JSON.stringify({ roleId: currentRoleId })
    })
    const data = await rsp.json()
    return Array.isArray(data?.data?.permIds) ? data.data.permIds : []
  }, roleId)
}

const restoreRolePermissionIds = async (page: Page, roleId: number, permIds: number[]) => {
  await page.evaluate(
    async ({ currentRoleId, currentPermIds }) => {
      const token = localStorage.getItem('user.token')
      await fetch('/de2api/auth/saveBusiPer', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`
        },
        body: JSON.stringify({ roleId: currentRoleId, permIds: currentPermIds })
      })
    },
    { currentRoleId: roleId, currentPermIds: permIds }
  )
}

const collectLeafPermissionIds = (permissions: PermissionNode[]): number[] => {
  const parentIds = new Set<number>()

  permissions.forEach(permission => {
    if (permission.parentId) {
      parentIds.add(permission.parentId)
    }
  })

  return permissions
    .filter(permission => !parentIds.has(permission.permId))
    .map(permission => permission.permId)
}

const chooseTargetRole = (roles: RoleSummary[]): RoleSummary | undefined => {
  return roles.find(role => !/admin|管理员/i.test(role.roleName)) || roles[0]
}

const selectRoleInUi = async (page: Page, roleName: string) => {
  await page.locator('.resource-permission .el-select').click()
  await page.locator('.el-select-dropdown__item').filter({ hasText: roleName }).first().click()
}

const permissionCheckbox = (page: Page, permId: number) =>
  page.locator(`.resource-permission .el-tree [data-key="${permId}"] .el-checkbox`).first()

const isCheckboxChecked = async (page: Page, permId: number) => {
  const classes = (await permissionCheckbox(page, permId).getAttribute('class')) || ''
  return classes.includes('is-checked')
}

test('resource permission should save and echo after reload', async ({ page, context }) => {
  const loggedIn = await loginAndOpenPermissionPage(page, context)
  if (!loggedIn) return

  const fixture = await fetchPermissionFixture(page)
  const targetRole = chooseTargetRole(fixture.roles)
  expect(targetRole).toBeTruthy()

  const roleId = targetRole!.roleId
  const roleName = targetRole!.roleName
  const originalPermIds = await fetchRolePermissionIds(page, roleId)
  const leafPermissionIds = collectLeafPermissionIds(fixture.permissions)
  const addCandidate = leafPermissionIds.find(id => !originalPermIds.includes(id))
  const removeCandidate = leafPermissionIds.find(id => originalPermIds.includes(id))
  const targetPermId = addCandidate || removeCandidate

  expect(targetPermId).toBeTruthy()

  const expectedChecked = Boolean(addCandidate)

  try {
    await openResourcePermissionTab(page)
    await selectRoleInUi(page, roleName)
    await expect(permissionCheckbox(page, targetPermId!)).toBeVisible({ timeout: 10000 })

    const beforeChecked = await isCheckboxChecked(page, targetPermId!)
    if (beforeChecked !== expectedChecked) {
      await permissionCheckbox(page, targetPermId!).click()
    }

    await page.locator('.resource-permission .footer').getByRole('button', { name: '保存' }).click()
    await expect(page.locator('.el-message--success')).toContainText('资源权限保存成功', {
      timeout: 10000
    })

    await page.reload()
    await expect(page.locator('.permission-config')).toBeVisible({ timeout: 10000 })
    await openResourcePermissionTab(page)
    await selectRoleInUi(page, roleName)
    await expect(permissionCheckbox(page, targetPermId!)).toBeVisible({ timeout: 10000 })
    await expect.poll(async () => isCheckboxChecked(page, targetPermId!)).toBe(expectedChecked)
  } finally {
    await restoreRolePermissionIds(page, roleId, originalPermIds)
  }
})
