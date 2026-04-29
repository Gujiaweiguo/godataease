import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

const hoisted = vi.hoisted(() => ({
  queryRoleApi: vi.fn(),
  menuTargetPerApi: vi.fn(),
  menuTargetPerSaveApi: vi.fn(),
  messageSuccess: vi.fn(),
  messageError: vi.fn(),
  messageWarning: vi.fn()
}))

vi.mock('@/api/auth', () => ({
  queryRoleApi: hoisted.queryRoleApi,
  menuTargetPerApi: hoisted.menuTargetPerApi,
  menuTargetPerSaveApi: hoisted.menuTargetPerSaveApi
}))

vi.mock('element-plus-secondary', () => ({
  ElMessage: {
    success: hoisted.messageSuccess,
    error: hoisted.messageError,
    warning: hoisted.messageWarning
  }
}))

import MenuPermission from '../../../../src/views/system/permission/MenuPermission.vue'

describe('MenuPermission unified permission compat API', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    hoisted.queryRoleApi.mockResolvedValue({
      code: '000000',
      data: { list: [{ roleId: 1, roleName: '管理员' }] }
    })
    hoisted.menuTargetPerApi.mockResolvedValue({
      code: '000000',
      data: { menuTree: [{ id: 10, name: '系统管理', children: [] }], menuIds: [] }
    })
    hoisted.menuTargetPerSaveApi.mockResolvedValue({ code: '000000' })
  })

  it('loads menu tree on mount via menuTargetPerApi with roleId 0', async () => {
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

    expect(hoisted.menuTargetPerApi).toHaveBeenCalledWith({ roleId: 0 })
  })

  it('loads role menu permissions on role selection and saves via compat API', async () => {
    hoisted.menuTargetPerApi.mockResolvedValue({
      code: '000000',
      data: { menuTree: [{ id: 10, name: '系统管理', children: [] }], menuIds: [10] }
    })

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

    const setupState = (wrapper.vm as any).$.setupState
    setupState.selectedRoleId = 1
    await setupState.handleRoleChange()

    expect(hoisted.menuTargetPerApi).toHaveBeenCalledWith({ roleId: 1 })

    setupState.selectedMenuIds = [10]
    await setupState.handleSave()

    expect(hoisted.menuTargetPerSaveApi).toHaveBeenCalledWith({ roleId: 1, menuIds: [10] })
    expect(hoisted.messageSuccess).toHaveBeenCalled()
  })
})
