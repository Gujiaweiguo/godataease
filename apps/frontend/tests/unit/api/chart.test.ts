import { describe, it, expect, vi, beforeEach } from 'vitest'

const { requestMock } = vi.hoisted(() => ({
  requestMock: {
    post: vi.fn().mockResolvedValue({ data: {} }),
    get: vi.fn().mockResolvedValue({ data: {} })
  }
}))

vi.mock('@/config/axios', () => ({ default: requestMock }))

import {
  getFieldByDQ,
  copyChartField,
  deleteChartField,
  deleteChartFieldByChartId,
  getData,
  innerExportDetails,
  innerExportDataSetDetails,
  getChart,
  saveChart,
  getFieldData,
  getDrillFieldData,
  getChartDetail,
  checkSameDataSet
} from '@/api/chart'

describe('api/chart', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    requestMock.post.mockResolvedValue({ data: {} })
    requestMock.get.mockResolvedValue({ data: {} })
  })

  it('getFieldByDQ posts to listByDQ with id and chartId', async () => {
    requestMock.post.mockResolvedValueOnce({ data: { dimensionList: [], quotaList: [] } })
    const data = { fieldIds: [1] }
    await getFieldByDQ('ds-1', 'chart-1', data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/chart/listByDQ/ds-1/chart-1',
      data
    })
  })

  it('copyChartField posts to copyField with id and chartId', async () => {
    requestMock.post.mockResolvedValueOnce({ data: { id: 1 } })
    await copyChartField(5, 10)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/chart/copyField/5/10',
      data: {}
    })
  })

  it('deleteChartField posts to deleteField with id', async () => {
    requestMock.post.mockResolvedValueOnce({ data: null })
    await deleteChartField(7)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/chart/deleteField/7',
      data: {}
    })
  })

  it('deleteChartFieldByChartId posts to deleteFieldByChart with chartId', async () => {
    requestMock.post.mockResolvedValueOnce({ data: null })
    await deleteChartFieldByChartId('chart-3')
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/chart/deleteFieldByChart/chart-3',
      data: {}
    })
  })

  it('getData posts to chartData/getData', async () => {
    requestMock.post.mockResolvedValueOnce({ code: 0, data: { data: {} } })
    const data = { id: 'chart-1', xAxis: [], yAxis: [] }
    await getData(data)
    expect(requestMock.post).toHaveBeenCalledWith(
      expect.objectContaining({
        url: '/chartData/getData'
      })
    )
  })

  it('innerExportDetails posts to chartData/innerExportDetails with blob', () => {
    const data = { viewId: 'chart-1' }
    innerExportDetails(data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/chartData/innerExportDetails',
      method: 'post',
      data,
      loading: true,
      responseType: 'blob'
    })
  })

  it('innerExportDataSetDetails posts to chartData/innerExportDataSetDetails with blob', () => {
    const data = { datasetId: 'ds-1' }
    innerExportDataSetDetails(data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/chartData/innerExportDataSetDetails',
      method: 'post',
      data,
      loading: true,
      responseType: 'blob'
    })
  })

  it('getChart posts to chart/getChart with id', async () => {
    requestMock.post.mockResolvedValueOnce({ data: { id: 'chart-1', name: 'Test' } })
    await getChart('chart-1')
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/chart/getChart/chart-1',
      data: {}
    })
  })

  it('saveChart posts to chart/save', async () => {
    requestMock.post.mockResolvedValueOnce({ data: { id: 'chart-1' } })
    const data = { id: 'chart-1', title: 'My Chart', xAxis: [], yAxis: [] }
    await saveChart(data)
    expect(requestMock.post).toHaveBeenCalledWith({ url: '/chart/save', data })
  })

  it('getFieldData posts to chartData/getFieldData with fieldId and fieldType', async () => {
    requestMock.post.mockResolvedValueOnce({ data: [] })
    const data = { datasetId: 'ds-1' }
    await getFieldData({ fieldId: 'f-1', fieldType: 'xAxis', data })
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/chartData/getFieldData/f-1/xAxis',
      data
    })
  })

  it('getDrillFieldData posts to chartData/getDrillFieldData with fieldId', async () => {
    requestMock.post.mockResolvedValueOnce({ data: [] })
    const data = { datasetId: 'ds-1' }
    await getDrillFieldData({ fieldId: 'f-2', data })
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/chartData/getDrillFieldData/f-2',
      data
    })
  })

  it('getChartDetail posts to chart/getDetail with id', async () => {
    requestMock.post.mockResolvedValueOnce({ data: { id: 'chart-1' } })
    await getChartDetail('chart-1')
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/chart/getDetail/chart-1',
      data: {}
    })
  })

  it('checkSameDataSet gets chart/checkSameDataSet with source and target', () => {
    checkSameDataSet('view-1', 'view-2')
    expect(requestMock.get).toHaveBeenCalledWith({
      url: '/chart/checkSameDataSet/view-1/view-2'
    })
  })
})
