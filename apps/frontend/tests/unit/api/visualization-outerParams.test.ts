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

import {
  queryWithVisualizationId,
  updateOuterParamsSet,
  getOuterParamsInfo
} from '@/api/visualization/outerParams'

describe('OuterParams API wrappers', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('queryWithVisualizationId gets outer params by dvId', () => {
    queryWithVisualizationId('dv1')
    expect(requestMock.get).toHaveBeenCalledWith({
      url: '/outerParams/queryWithVisualizationId/dv1'
    })
  })

  it('updateOuterParamsSet posts to update endpoint with loading true', () => {
    const requestInfo = { dvId: 'dv1', params: [] }
    updateOuterParamsSet(requestInfo)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/outerParams/updateOuterParamsSet',
      data: requestInfo,
      loading: true
    })
  })

  it('getOuterParamsInfo gets outer params info by dvId', () => {
    getOuterParamsInfo('dv1')
    expect(requestMock.get).toHaveBeenCalledWith({
      url: '/outerParams/getOuterParamsInfo/dv1',
      method: 'get',
      loading: false
    })
  })

  it('queryWithVisualizationId returns the request promise', async () => {
    requestMock.get.mockResolvedValueOnce({ data: { params: [] } })
    const result = await queryWithVisualizationId('dv1')
    expect(result).toEqual({ data: { params: [] } })
  })

  it('updateOuterParamsSet returns the request promise', async () => {
    requestMock.post.mockResolvedValueOnce({ data: { success: true } })
    const result = await updateOuterParamsSet({ dvId: 'dv1' })
    expect(result).toEqual({ data: { success: true } })
  })
})
