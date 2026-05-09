import { describe, it, expect, vi, beforeEach } from 'vitest'

const { requestMock } = vi.hoisted(() => ({
  requestMock: {
    post: vi.fn().mockResolvedValue({ data: {} }),
    get: vi.fn().mockResolvedValue({ data: {} })
  }
}))

vi.mock('@/config/axios', () => ({ default: requestMock }))

import { queryAuditLogsApi, exportAuditLogsApi } from '@/api/audit'

describe('api/audit', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('queryAuditLogsApi gets audit list with default empty params', () => {
    queryAuditLogsApi()
    expect(requestMock.get).toHaveBeenCalledWith({ url: '/audit/list', params: {} })
  })

  it('queryAuditLogsApi gets audit list with custom params', () => {
    const params = { page: 1, pageSize: 20 }
    queryAuditLogsApi(params)
    expect(requestMock.get).toHaveBeenCalledWith({ url: '/audit/list', params })
  })

  it('exportAuditLogsApi posts to audit export with default csv format', () => {
    const ids = [1, 2, 3]
    exportAuditLogsApi(ids)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/audit/export?format=csv',
      data: ids
    })
  })

  it('exportAuditLogsApi posts to audit export with custom format', () => {
    const ids = [4, 5]
    exportAuditLogsApi(ids, 'xlsx')
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/audit/export?format=xlsx',
      data: ids
    })
  })
})
