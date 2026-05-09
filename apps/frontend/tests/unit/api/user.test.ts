import { describe, it, expect, vi, beforeEach } from 'vitest'

const { requestMock } = vi.hoisted(() => ({
  requestMock: {
    post: vi.fn().mockResolvedValue({ data: {} }),
    get: vi.fn().mockResolvedValue({ data: {} })
  }
}))

vi.mock('@/config/axios', () => ({ default: requestMock }))

import {
  mountedOrg,
  switchOrg,
  userInfo,
  userOptionForRoleApi,
  userSelectedForRoleApi,
  personInfoApi,
  ipInfoApi,
  beforeUnmountInfoApi,
  unMountUserApi,
  mountUserApi,
  searchExternalUserApi,
  mountExternalUserApi,
  switchLangApi,
  defaultPwdApi,
  resetPwdApi,
  switchEnableApi
} from '@/api/user'

describe('User API wrappers', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('mountedOrg posts keyword to mounted endpoint', () => {
    mountedOrg('test')
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/org/mounted',
      data: { keyword: 'test' }
    })
  })

  it('mountedOrg works without keyword argument', () => {
    mountedOrg()
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/org/mounted',
      data: { keyword: undefined }
    })
  })

  it('switchOrg posts to switch endpoint with id', () => {
    switchOrg(42)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/user/switch/42'
    })
  })

  it('switchOrg accepts string id', () => {
    switchOrg('abc-123')
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/user/switch/abc-123'
    })
  })

  it('userInfo fetches user info', () => {
    userInfo()
    expect(requestMock.get).toHaveBeenCalledWith({ url: '/user/info' })
  })

  it('userOptionForRoleApi posts data to role user option', () => {
    const data = { roleId: 1 }
    userOptionForRoleApi(data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/role/user/option',
      data
    })
  })

  it('userSelectedForRoleApi posts paginated data', () => {
    const data = { roleId: 1 }
    userSelectedForRoleApi(2, 10, data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/role/user/selected',
      data: { roleId: 1, page: 2, limit: 10 }
    })
  })

  it('personInfoApi fetches personInfo', () => {
    personInfoApi()
    expect(requestMock.get).toHaveBeenCalledWith({ url: '/user/personInfo' })
  })

  it('ipInfoApi fetches ipInfo', () => {
    ipInfoApi()
    expect(requestMock.get).toHaveBeenCalledWith({ url: '/user/ipInfo' })
  })

  it('beforeUnmountInfoApi posts data', () => {
    const data = { userId: '123' }
    beforeUnmountInfoApi(data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/role/beforeUnmountInfo',
      data
    })
  })

  it('unMountUserApi posts data to unmount endpoint', () => {
    const data = { userId: '123' }
    unMountUserApi(data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/role/unMountUser',
      data
    })
  })

  it('mountUserApi posts data to mount endpoint', () => {
    const data = { userId: '123' }
    mountUserApi(data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/role/mountUser',
      data
    })
  })

  it('searchExternalUserApi fetches with keyword suffix', () => {
    searchExternalUserApi('john')
    expect(requestMock.get).toHaveBeenCalledWith({
      url: '/role/searchExternalUser/john'
    })
  })

  it('mountExternalUserApi posts data', () => {
    const data = { userId: 'ext-123' }
    mountExternalUserApi(data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/role/mountExternalUser',
      data
    })
  })

  it('switchLangApi posts data to switchLanguage', () => {
    const data = { lang: 'en' }
    switchLangApi(data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/user/switchLanguage',
      data
    })
  })

  it('defaultPwdApi fetches default password endpoint', () => {
    defaultPwdApi()
    expect(requestMock.get).toHaveBeenCalledWith({ url: '/system/user/defaultPwd' })
  })

  it('resetPwdApi posts to resetPwd with uid', () => {
    resetPwdApi(42)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/system/user/resetPwd/42'
    })
  })

  it('switchEnableApi posts data to enable endpoint', () => {
    const data = { id: 1, enabled: true }
    switchEnableApi(data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/system/user/enable',
      data
    })
  })
})
