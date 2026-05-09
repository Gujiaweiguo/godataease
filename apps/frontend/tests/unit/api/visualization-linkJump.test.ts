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
  getTableFieldWithViewId,
  queryWithViewId,
  updateJumpSet,
  queryTargetVisualizationJumpInfo,
  queryVisualizationJumpInfo,
  viewTableDetailList,
  updateJumpSetActive,
  removeJumpSet
} from '@/api/visualization/linkJump'

describe('LinkJump API wrappers', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('getTableFieldWithViewId gets field info by viewId', () => {
    getTableFieldWithViewId('view1')
    expect(requestMock.get).toHaveBeenCalledWith({
      url: '/linkJump/getTableFieldWithViewId/view1'
    })
  })

  it('queryWithViewId gets jump info by dvId and viewId', () => {
    queryWithViewId('dv1', 'view1')
    expect(requestMock.get).toHaveBeenCalledWith({
      url: '/linkJump/queryWithViewId/dv1/view1'
    })
  })

  it('updateJumpSet posts to updateJumpSet with loading true', () => {
    const requestInfo = { dvId: 'dv1', viewId: 'view1' }
    updateJumpSet(requestInfo)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/linkJump/updateJumpSet',
      data: requestInfo,
      loading: true
    })
  })

  it('queryTargetVisualizationJumpInfo posts with loading true', () => {
    const requestInfo = { dvId: 'dv1' }
    queryTargetVisualizationJumpInfo(requestInfo)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/linkJump/queryTargetVisualizationJumpInfo',
      data: requestInfo,
      loading: true
    })
  })

  it('queryVisualizationJumpInfo gets by dvId with default resourceTable', () => {
    queryVisualizationJumpInfo('dv1')
    expect(requestMock.get).toHaveBeenCalledWith({
      url: '/linkJump/queryVisualizationJumpInfo/dv1/snapshot',
      loading: false
    })
  })

  it('queryVisualizationJumpInfo uses custom resourceTable', () => {
    queryVisualizationJumpInfo('dv1', 'core')
    expect(requestMock.get).toHaveBeenCalledWith({
      url: '/linkJump/queryVisualizationJumpInfo/dv1/core',
      loading: false
    })
  })

  it('viewTableDetailList gets detail list by dvId', () => {
    viewTableDetailList('dv1')
    expect(requestMock.get).toHaveBeenCalledWith({
      url: '/linkJump/viewTableDetailList/dv1',
      loading: false
    })
  })

  it('updateJumpSetActive posts with loading true', () => {
    const requestInfo = { active: true }
    updateJumpSetActive(requestInfo)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/linkJump/updateJumpSetActive',
      data: requestInfo,
      loading: true
    })
  })

  it('removeJumpSet posts with loading true', () => {
    const requestInfo = { linkJumpId: 'jump1' }
    removeJumpSet(requestInfo)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/linkJump/removeJumpSet',
      data: requestInfo,
      loading: true
    })
  })

  it('getTableFieldWithViewId returns the request promise', async () => {
    requestMock.get.mockResolvedValueOnce({ data: { fields: [] } })
    const result = await getTableFieldWithViewId('view1')
    expect(result).toEqual({ data: { fields: [] } })
  })
})
