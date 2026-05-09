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
  getViewLinkageGather,
  getViewLinkageGatherArray,
  saveLinkage,
  getPanelAllLinkageInfo,
  updateLinkageActive,
  removeLinkage
} from '@/api/visualization/linkage'

describe('Linkage API wrappers', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('getViewLinkageGather posts to getViewLinkageGather', () => {
    const data = { dvId: 'dv1', viewIds: ['v1'] }
    getViewLinkageGather(data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/linkage/getViewLinkageGather',
      data
    })
  })

  it('getViewLinkageGatherArray posts to getViewLinkageGatherArray', () => {
    const data = { dvId: 'dv1', viewIds: ['v1', 'v2'] }
    getViewLinkageGatherArray(data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/linkage/getViewLinkageGatherArray',
      data
    })
  })

  it('saveLinkage posts to saveLinkage', () => {
    const data = { linkageInfo: {} }
    saveLinkage(data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/linkage/saveLinkage',
      data
    })
  })

  it('getPanelAllLinkageInfo gets all linkage info with default resourceTable', () => {
    getPanelAllLinkageInfo('dv1')
    expect(requestMock.get).toHaveBeenCalledWith({
      url: '/linkage/getVisualizationAllLinkageInfo/dv1/snapshot'
    })
  })

  it('getPanelAllLinkageInfo uses custom resourceTable', () => {
    getPanelAllLinkageInfo('dv1', 'core')
    expect(requestMock.get).toHaveBeenCalledWith({
      url: '/linkage/getVisualizationAllLinkageInfo/dv1/core'
    })
  })

  it('updateLinkageActive posts to updateLinkageActive', () => {
    const data = { linkageId: 'l1', active: true }
    updateLinkageActive(data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/linkage/updateLinkageActive',
      data
    })
  })

  it('removeLinkage posts to removeLinkage', () => {
    const data = { linkageId: 'l1' }
    removeLinkage(data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/linkage/removeLinkage',
      data
    })
  })

  it('getViewLinkageGather returns the request promise', async () => {
    requestMock.post.mockResolvedValueOnce({ data: { views: [] } })
    const result = await getViewLinkageGather({ dvId: 'dv1' })
    expect(result).toEqual({ data: { views: [] } })
  })
})
