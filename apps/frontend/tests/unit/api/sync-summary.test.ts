import { describe, it, expect, vi, beforeEach } from 'vitest'

const { requestMock } = vi.hoisted(() => ({
  requestMock: {
    post: vi.fn().mockResolvedValue({ data: {} }),
    get: vi.fn().mockResolvedValue({ data: {} }),
    delete: vi.fn().mockResolvedValue({ data: {} }),
    put: vi.fn().mockResolvedValue({ data: {} }),
    download: vi.fn().mockResolvedValue({ data: {} })
  }
}))

vi.mock('@/config/axios', () => ({ default: requestMock }))

import { getResourceCount, getJobLogLienChartInfo } from '@/api/sync/syncSummary'

describe('SyncSummary API wrappers', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('getResourceCount gets sync/summary/resourceCount', async () => {
    requestMock.get.mockResolvedValueOnce({
      data: { jobCount: 5, datasourceCount: 3, jobLogCount: 10 }
    })
    const result = await getResourceCount()
    expect(requestMock.get).toHaveBeenCalledWith({
      url: '/sync/summary/resourceCount',
      method: 'get'
    })
    expect(result).toEqual({ jobCount: 5, datasourceCount: 3, jobLogCount: 10 })
  })

  it('getResourceCount returns data as IResourceCount', async () => {
    requestMock.get.mockResolvedValueOnce({
      data: { jobCount: 0, datasourceCount: 0, jobLogCount: 0 }
    })
    const result = await getResourceCount()
    expect(result).toEqual({ jobCount: 0, datasourceCount: 0, jobLogCount: 0 })
  })

  it('getJobLogLienChartInfo posts to sync/summary/logChartData', () => {
    getJobLogLienChartInfo()
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/sync/summary/logChartData',
      method: 'post',
      data: ''
    })
  })

  it('getJobLogLienChartInfo returns the request promise', async () => {
    requestMock.post.mockResolvedValueOnce({ data: { chart: 'info' } })
    const result = await getJobLogLienChartInfo()
    expect(result).toEqual({ data: { chart: 'info' } })
  })
})
