import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

const hoisted = vi.hoisted(() => ({
  queryRoleApi: vi.fn(),
  menuTreeApi: vi.fn(),
  roleMenuAuthApi: vi.fn(),
  roleMenuAuthSaveApi: vi.fn(),
  messageSuccess: vi.fn(),
  messageError: vi.fn(),
  messageWarning: vi.fn()
}))

vi.mock('@/api/auth', () => ({
  queryRoleApi: hoisted.queryRoleApi,
  menuTreeApi: hoisted.menuTreeApi,
  roleMenuAuthApi: hoisted.roleMenuAuthApi,
  roleMenuAuthSaveApi: hoisted.roleMenuAuthSaveApi
}))

vi.mock('element-plus-secondary', () => ({
  ElMessage: {
    success: hoisted.messageSuccess,
    error: hoisted.messageError,
    warning: hoisted.messageWarning
  }
}))

import MenuPermission from '../../../../src/views/system/permission/MenuPermission.vue'

describe('MenuPermission canonical role-menu api rollout', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    hoisted.queryRoleApi.mockResolvedValue({
      code: '000000',
      data: { list: [{ roleId: 1, roleName: '管理员' }] }
    })
    hoisted.menuTreeApi.mockResolvedValue({
      code: '000000',
      data: [{ id: 10, name: '系统管理', children: [] }]
    })
    hoisted.roleMenuAuthApi.mockResolvedValue({
      code: '000000',
      data: { menuTree: [{ id: 10, name: '系统管理', children: [] }], menuIds: [10] }
    })
    hoisted.roleMenuAuthSaveApi.mockResolvedValue({ code: '000000' })
  })

  it('uses roleMenu auth apis for role load/save while preserving initial tree load', async () => {
    const wrapper = mount(MenuPermission, {
      global: {
        stubs: {
          'el-select': true,
          'el-option': true,
          'el-tree': true,
          'el-button': true,
          'el-empty': true
        }
      }
    })

    await flushPromises()

    expect(hoisted.menuTreeApi).toHaveBeenCalledTimes(1)
    expect(hoisted.roleMenuAuthApi).not.toHaveBeenCalled()

    const setupState = (wrapper.vm as any).$.setupState
    setupState.selectedRoleId = 1
    await setupState.handleRoleChange()

    expect(hoisted.roleMenuAuthApi).toHaveBeenCalledWith(1)

    setupState.selectedMenuIds = [10]
    await setupState.handleSave()

    expect(hoisted.roleMenuAuthSaveApi).toHaveBeenCalledWith({ roleId: 1, menuIds: [10] })
    expect(hoisted.messageSuccess).toHaveBeenCalled()
  })
})
