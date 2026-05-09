import { describe, it, expect, vi, beforeEach } from 'vitest'

const { requestMock } = vi.hoisted(() => ({
  requestMock: {
    post: vi.fn().mockResolvedValue({ data: {} }),
    get: vi.fn().mockResolvedValue({ data: {} })
  }
}))

vi.mock('@/config/axios', () => ({ default: requestMock }))

import {
  orgListApi,
  orgCreateApi,
  orgUpdateApi,
  orgDeleteApi,
  orgTreeApi,
  queryUserOptionsApi,
  permListApi,
  permCreateApi,
  permUpdateApi,
  permDeleteApi
} from '@/api/org'

describe('api/org', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('orgListApi gets organization list', () => {
    orgListApi()
    expect(requestMock.get).toHaveBeenCalledWith({ url: '/system/organization/list' })
  })

  it('orgCreateApi posts to organization create', () => {
    const data = { name: 'Org1' }
    orgCreateApi(data)
    expect(requestMock.post).toHaveBeenCalledWith({ url: '/system/organization/create', data })
  })

  it('orgUpdateApi posts to organization update', () => {
    const data = { id: 1, name: 'Updated' }
    orgUpdateApi(data)
    expect(requestMock.post).toHaveBeenCalledWith({ url: '/system/organization/update', data })
  })

  it('orgDeleteApi posts to organization delete with id in url', () => {
    orgDeleteApi(5)
    expect(requestMock.post).toHaveBeenCalledWith({ url: '/system/organization/delete/5' })
  })

  it('orgTreeApi gets organization tree', () => {
    orgTreeApi()
    expect(requestMock.get).toHaveBeenCalledWith({ url: '/system/organization/tree' })
  })

  it('queryUserOptionsApi gets user options', () => {
    queryUserOptionsApi()
    expect(requestMock.get).toHaveBeenCalledWith({ url: '/system/user/options' })
  })

  it('permListApi posts to permission list with default empty object', () => {
    permListApi()
    expect(requestMock.post).toHaveBeenCalledWith({ url: '/system/permission/list', data: {} })
  })

  it('permListApi posts to permission list with params', () => {
    const params = { type: 'menu' }
    permListApi(params)
    expect(requestMock.post).toHaveBeenCalledWith({ url: '/system/permission/list', data: params })
  })

  it('permCreateApi posts to permission create', () => {
    const data = { name: 'perm1' }
    permCreateApi(data)
    expect(requestMock.post).toHaveBeenCalledWith({ url: '/system/permission/create', data })
  })

  it('permUpdateApi posts to permission update', () => {
    const data = { id: 1, name: 'updated' }
    permUpdateApi(data)
    expect(requestMock.post).toHaveBeenCalledWith({ url: '/system/permission/update', data })
  })

  it('permDeleteApi posts to permission delete with id in url', () => {
    permDeleteApi(3)
    expect(requestMock.post).toHaveBeenCalledWith({ url: '/system/permission/delete/3' })
  })
})
