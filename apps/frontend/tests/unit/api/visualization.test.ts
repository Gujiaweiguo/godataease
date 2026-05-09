import { describe, it, expect, vi, beforeEach } from 'vitest'

const { requestMock } = vi.hoisted(() => ({
  requestMock: {
    post: vi.fn().mockResolvedValue({ data: {} }),
    get: vi.fn().mockResolvedValue({ data: {} })
  }
}))

vi.mock('@/config/axios', () => ({ default: requestMock }))

const { cloneDeepMock, originNameHandleWithArrMock } = vi.hoisted(() => ({
  cloneDeepMock: vi.fn((data: unknown) => ({ ...(data as object), _cloned: true })),
  originNameHandleWithArrMock: vi.fn()
}))

vi.mock('lodash-es', () => ({
  cloneDeep: cloneDeepMock
}))

vi.mock('@/utils/CalculateFields', () => ({
  originNameHandleWithArr: originNameHandleWithArrMock
}))

import {
  findCopyResource,
  updateCheckVersion,
  queryTreeApi,
  queryBusiTreeApi,
  findDvType,
  save,
  checkCanvasChange,
  saveCanvas,
  updatePublishStatus,
  recoverToPublished,
  appCanvasNameCheck,
  updateBase,
  moveResource,
  copyResource,
  deleteLogic,
  dvNameCheck,
  storeApi,
  storeStatusApi,
  exportLogApp,
  exportLogTemplate,
  exportLogPDF,
  exportLogImg
} from '@/api/visualization/dataVisualization'

describe('DataVisualization API wrappers', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('findCopyResource fetches copy resource by dvId and busiFlag', () => {
    findCopyResource('123', 'dataV')
    expect(requestMock.get).toHaveBeenCalledWith({
      url: '/dataVisualization/findCopyResource/123/dataV'
    })
  })

  it('updateCheckVersion fetches check version by dvId', () => {
    updateCheckVersion('456')
    expect(requestMock.get).toHaveBeenCalledWith({
      url: '/dataVisualization/updateCheckVersion/456'
    })
  })

  it('queryTreeApi posts to tree endpoint and extracts data', async () => {
    requestMock.post.mockResolvedValueOnce({ data: [{ id: '1', name: 'root' }] })
    const result = await queryTreeApi({ busiFlag: 'dataV' })
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/dataVisualization/tree',
      data: { busiFlag: 'dataV' }
    })
    expect(result).toEqual([{ id: '1', name: 'root' }])
  })

  it('queryBusiTreeApi posts to interactiveTree endpoint and extracts data', async () => {
    requestMock.post.mockResolvedValueOnce({ data: [{ id: '2', name: 'interactive' }] })
    const result = await queryBusiTreeApi({ keyword: 'test' })
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/dataVisualization/interactiveTree',
      data: { keyword: 'test' }
    })
    expect(result).toEqual([{ id: '2', name: 'interactive' }])
  })

  it('findDvType fetches dv type by dvId', () => {
    findDvType('789')
    expect(requestMock.get).toHaveBeenCalledWith({
      url: '/dataVisualization/findDvType/789'
    })
  })

  it('save posts data to save endpoint', () => {
    const data = { name: 'dashboard1', type: 'dataV' }
    save(data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/dataVisualization/save',
      data
    })
  })

  it('checkCanvasChange posts with loading true', () => {
    const data = { id: '123', componentData: [] }
    checkCanvasChange(data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/dataVisualization/checkCanvasChange',
      data,
      loading: true
    })
  })

  it('saveCanvas posts with loading true', () => {
    const data = { id: '123', canvasStyleData: {} }
    saveCanvas(data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/dataVisualization/saveCanvas',
      data,
      loading: true
    })
  })

  it('updatePublishStatus posts with loading false', () => {
    const data = { id: '123', status: 1 }
    updatePublishStatus(data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/dataVisualization/updatePublishStatus',
      data,
      loading: false
    })
  })

  it('recoverToPublished posts with loading true', () => {
    const data = { id: '123' }
    recoverToPublished(data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/dataVisualization/recoverToPublished',
      data,
      loading: true
    })
  })

  it('appCanvasNameCheck posts with loading false', () => {
    const data = { name: 'test-app' }
    appCanvasNameCheck(data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/dataVisualization/appCanvasNameCheck',
      data,
      loading: false
    })
  })

  it('updateBase posts data to updateBase endpoint', () => {
    const data = { id: '123', name: 'updated' }
    updateBase(data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/dataVisualization/updateBase',
      data
    })
  })

  it('moveResource posts data to move endpoint', () => {
    const data = { id: '123', pid: '456' }
    moveResource(data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/dataVisualization/move',
      data
    })
  })

  it('copyResource posts data to copy endpoint', () => {
    const data = { id: '123', name: 'copy' }
    copyResource(data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/dataVisualization/copy',
      data
    })
  })

  it('deleteLogic posts with dvId and busiFlag in URL', () => {
    deleteLogic('123', 'dataV')
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/dataVisualization/deleteLogic/123/dataV'
    })
  })

  it('dvNameCheck posts data to nameCheck endpoint', () => {
    const data = { name: 'new-dashboard' }
    dvNameCheck(data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/dataVisualization/nameCheck',
      data
    })
  })

  it('storeApi posts to store execute endpoint', () => {
    const data = { id: '123' }
    storeApi(data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/store/execute',
      data
    })
  })

  it('storeStatusApi fetches favorited status by id', () => {
    storeStatusApi('123')
    expect(requestMock.get).toHaveBeenCalledWith({
      url: '/store/favorited/123'
    })
  })

  it('exportLogApp posts data', () => {
    const data = { id: '123' }
    exportLogApp(data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/dataVisualization/exportLogApp',
      data
    })
  })

  it('exportLogTemplate posts data', () => {
    const data = { id: '456' }
    exportLogTemplate(data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/dataVisualization/exportLogTemplate',
      data
    })
  })

  it('exportLogPDF posts data', () => {
    const data = { id: '789' }
    exportLogPDF(data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/dataVisualization/exportLogPDF',
      data
    })
  })

  it('exportLogImg posts data', () => {
    const data = { id: '012' }
    exportLogImg(data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/dataVisualization/exportLogImg',
      data
    })
  })
})
