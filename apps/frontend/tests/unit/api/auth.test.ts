import { describe, it, expect, vi, beforeEach } from 'vitest'

const { requestMock } = vi.hoisted(() => ({
  requestMock: {
    post: vi.fn().mockResolvedValue({ data: {} }),
    get: vi.fn().mockResolvedValue({ data: {} })
  }
}))

vi.mock('@/config/axios', () => ({ default: requestMock }))

import {
  queryUserApi,
  queryRoleApi,
  userCreateApi,
  userUpdateApi,
  userDeleteApi,
  roleCreateApi,
  roleUpdateApi,
  roleDeleteApi,
  resourceTreeApi,
  resourcePerApi,
  busiPerSaveApi,
  resourceTargetPerApi,
  userPerspectiveApi,
  menuTargetPerApi,
  busiTargetPerSaveApi,
  menuTargetPerSaveApi
} from '@/api/auth'

describe('api/auth', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('queryUserApi posts to user list', () => {
    const data = { keyword: 'admin' }
    queryUserApi(data)
    expect(requestMock.post).toHaveBeenCalledWith({ url: '/system/user/list', data })
  })

  it('queryRoleApi posts to role byCurOrg', () => {
    const data = { orgId: 1 }
    queryRoleApi(data)
    expect(requestMock.post).toHaveBeenCalledWith({ url: '/role/byCurOrg', data })
  })

  it('userCreateApi posts to user create', () => {
    const data = { name: 'newuser' }
    userCreateApi(data)
    expect(requestMock.post).toHaveBeenCalledWith({ url: '/system/user/create', data })
  })

  it('userUpdateApi posts to user update', () => {
    const data = { id: 1, name: 'updated' }
    userUpdateApi(data)
    expect(requestMock.post).toHaveBeenCalledWith({ url: '/system/user/update', data })
  })

  it('userDeleteApi posts to user delete with id in url', () => {
    userDeleteApi(42)
    expect(requestMock.post).toHaveBeenCalledWith({ url: '/system/user/delete/42' })
  })

  it('roleCreateApi posts to role create', () => {
    const data = { name: 'admin' }
    roleCreateApi(data)
    expect(requestMock.post).toHaveBeenCalledWith({ url: '/system/role/create', data })
  })

  it('roleUpdateApi posts to role update', () => {
    const data = { id: 1, name: 'viewer' }
    roleUpdateApi(data)
    expect(requestMock.post).toHaveBeenCalledWith({ url: '/system/role/update', data })
  })

  it('roleDeleteApi posts to role delete with roleId in url', () => {
    roleDeleteApi(7)
    expect(requestMock.post).toHaveBeenCalledWith({ url: '/system/role/delete/7' })
  })

  it('resourceTreeApi gets busiResource with flag', () => {
    resourceTreeApi('panel')
    expect(requestMock.get).toHaveBeenCalledWith({ url: '/system/permission/busiResource/panel' })
  })

  it('resourcePerApi posts to busiPermission', () => {
    const data = { roleId: 1 }
    resourcePerApi(data)
    expect(requestMock.post).toHaveBeenCalledWith({ url: '/system/permission/busiPermission', data })
  })

  it('busiPerSaveApi posts to saveBusiPer', () => {
    const data = { permissions: [] }
    busiPerSaveApi(data)
    expect(requestMock.post).toHaveBeenCalledWith({ url: '/system/permission/saveBusiPer', data })
  })

  it('resourceTargetPerApi posts to busiTargetPermission', () => {
    const data = { targetId: 1 }
    resourceTargetPerApi(data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/system/permission/busiTargetPermission',
      data
    })
  })

  it('userPerspectiveApi posts to userPerspective', () => {
    const data = { userId: 1 }
    userPerspectiveApi(data)
    expect(requestMock.post).toHaveBeenCalledWith({ url: '/system/permission/userPerspective', data })
  })

  it('menuTargetPerApi posts to menuTargetPermission', () => {
    const data = { menuId: 1 }
    menuTargetPerApi(data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/system/permission/menuTargetPermission',
      data
    })
  })

  it('busiTargetPerSaveApi posts to saveBusiTargetPer', () => {
    const data = { perms: [] }
    busiTargetPerSaveApi(data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/system/permission/saveBusiTargetPer',
      data
    })
  })

  it('menuTargetPerSaveApi posts to saveMenuTargetPer', () => {
    const data = { menuPerms: [] }
    menuTargetPerSaveApi(data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/system/permission/saveMenuTargetPer',
      data
    })
  })
})
