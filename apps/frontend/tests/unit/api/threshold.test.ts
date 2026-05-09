import { describe, it, expect, vi, beforeEach } from 'vitest'

const { requestMock } = vi.hoisted(() => ({
  requestMock: {
    post: vi.fn().mockResolvedValue({ data: {} }),
    get: vi.fn().mockResolvedValue({ data: {} })
  }
}))

vi.mock('@/config/axios', () => ({ default: requestMock }))

import {
  thresholdSave,
  thresholdEdit,
  thresholdDelete,
  thresholdFormInfo,
  thresholdSwitch,
  thresholdPager,
  thresholdPreview,
  thresholdBatchReci,
  thresholdInstancePager,
  thresholdAnyThreshold,
  thresholdDeleteWithChart
} from '@/api/threshold'

describe('api/threshold', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('thresholdSave posts to save', () => {
    const params = { name: 'alert1' }
    thresholdSave(params)
    expect(requestMock.post).toHaveBeenCalledWith({ url: '/threshold/save', data: params })
  })

  it('thresholdEdit posts to edit', () => {
    const params = { id: 1, name: 'updated' }
    thresholdEdit(params)
    expect(requestMock.post).toHaveBeenCalledWith({ url: '/threshold/edit', data: params })
  })

  it('thresholdDelete posts to delete with resourceTable and ids', () => {
    const ids = [1, 2, 3]
    thresholdDelete('chart', ids)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/threshold/delete/chart',
      data: ids
    })
  })

  it('thresholdFormInfo gets formInfo with id and resourceTable', () => {
    thresholdFormInfo('42', 'dataset')
    expect(requestMock.get).toHaveBeenCalledWith({ url: '/threshold/formInfo/42/dataset' })
  })

  it('thresholdSwitch posts to switch', () => {
    const params = { id: 1, enabled: true }
    thresholdSwitch(params)
    expect(requestMock.post).toHaveBeenCalledWith({ url: '/threshold/switch', data: params })
  })

  it('thresholdPager posts to pager with goPage and pageSize', () => {
    const params = { keyword: '' }
    thresholdPager(1, 20, params)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/threshold/pager/1/20',
      data: params
    })
  })

  it('thresholdPreview posts to preview', () => {
    const params = { chartId: 'abc' }
    thresholdPreview(params)
    expect(requestMock.post).toHaveBeenCalledWith({ url: '/threshold/preview', data: params })
  })

  it('thresholdBatchReci posts to batchReci', () => {
    const params = { ids: [1, 2] }
    thresholdBatchReci(params)
    expect(requestMock.post).toHaveBeenCalledWith({ url: '/threshold/batchReci', data: params })
  })

  it('thresholdInstancePager posts to instancePager with pagination', () => {
    const params = { thresholdId: 1 }
    thresholdInstancePager(2, 10, params)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/threshold/instancePager/2/10',
      data: params
    })
  })

  it('thresholdAnyThreshold gets anyThreshold with chartId and resourceTable', () => {
    thresholdAnyThreshold('chart-1', 'chart')
    expect(requestMock.get).toHaveBeenCalledWith({
      url: '/threshold/anyThreshold/chart-1/chart'
    })
  })

  it('thresholdDeleteWithChart gets deleteWithChart with chartId and resourceTable', () => {
    thresholdDeleteWithChart('chart-2', 'dataset')
    expect(requestMock.get).toHaveBeenCalledWith({
      url: '/threshold/deleteWithChart/chart-2/dataset'
    })
  })
})
