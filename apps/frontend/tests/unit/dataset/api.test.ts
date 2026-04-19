import { beforeEach, describe, expect, it, vi } from 'vitest'

const { requestMock } = vi.hoisted(() => ({
  requestMock: {
    post: vi.fn(),
    get: vi.fn()
  }
}))

vi.mock('@/config/axios', () => ({
  default: requestMock
}))

import {
  deleteField,
  deleteFieldByChartId,
  exportDatasetData,
  exportRetry,
  exportTasks,
  exportTasksRecords,
  getDatasetTree,
  getPreviewSql,
  getPreviewData,
  getTableField,
  saveDatasetTree,
  createDatasetTree,
  renameDatasetTree,
  moveDatasetTree,
  delDatasetTree,
  perDelete,
  enumValueObj,
  enumValueDs,
  barInfoApi,
  getDatasetPreview,
  getDatasetTotal,
  getDatasetDetails,
  getDsDetails,
  getSqlParams,
  getEnumValue
} from '@/api/dataset'

describe('Dataset API wrappers', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('injects dataset busiFlag and normalizes leaf state from nodeType', async () => {
    requestMock.post.mockResolvedValueOnce({
      data: [
        {
          id: 1,
          name: 'Dataset Folder',
          nodeType: 'folder',
          children: [{ id: 2, name: 'Sales Dataset', nodeType: 'dataset' }]
        }
      ]
    })

    const result = await getDatasetTree({})

    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/dataset/tree',
      data: { busiFlag: 'dataset' }
    })
    expect(result[0].leaf).toBe(false)
    expect(result[0].children[0].leaf).toBe(true)
  })

  it('requests dataset fields through the canonical dataset route', async () => {
    requestMock.post.mockResolvedValueOnce({ data: [{ id: 'field-1', originName: 'orders.amount' }] })

    const payload = { datasourceId: 7, tableName: 'orders' }
    const result = await getTableField(payload)

    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/dataset/fields',
      data: payload
    })
    expect(result).toEqual([{ id: 'field-1', originName: 'orders.amount' }])
  })

  it('requests dataset preview through the canonical dataset route and preserves field post-processing', async () => {
    const response = {
      data: {
        allFields: [{ originName: 'orders.amount', name: 'Amount' }],
        data: {
          fields: [{ originName: 'orders.amount', name: 'Amount' }],
          data: [{ amount: 12 }]
        }
      }
    }
    requestMock.post.mockResolvedValueOnce(response)

    const payload = { allFields: [{ originName: '[orders.amount]', name: 'Amount' }] }
    const result = await getPreviewData(payload)

    expect(requestMock.post).toHaveBeenCalledWith(
      expect.objectContaining({
        url: '/dataset/preview',
        data: expect.any(Object)
      })
    )
    expect(result).toEqual(response.data)
    expect(payload.allFields[0].originName).toBe('[orders.amount]')
  })

  it('passes sqlVariableDetails through the previewSql canonical route', async () => {
    requestMock.post.mockResolvedValueOnce({ data: { data: { fields: [], data: [] }, sql: 'U0VMRUNUIDE=' } })

    const payload = {
      sql: 'U0VMRUNUIDE=',
      datasourceId: 66,
      isCross: false,
      sqlVariableDetails: JSON.stringify([{ variableName: 'region', defaultValue: '华东' }])
    }

    const result = await getPreviewSql(payload)

    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/dataset/previewSql',
      data: payload
    })
    expect(result).toEqual({ data: { fields: [], data: [] }, sql: 'U0VMRUNUIDE=' })
  })

  it('requests export-center task counters through the records alias', async () => {
    requestMock.post.mockResolvedValueOnce({ code: '000000', data: { ALL: 3, FAILED: 1 } })

    const result = await exportTasksRecords()

    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/exportCenter/exportTasks/records',
      data: {}
    })
    expect(result).toEqual({ code: '000000', data: { ALL: 3, FAILED: 1 } })
  })

  it('requests export-center tasks for the given status and pagination', async () => {
    requestMock.post.mockResolvedValueOnce({ code: '000000', data: { total: 1, records: [] } })

    const result = await exportTasks(2, 20, 'FAILED')

    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/exportCenter/exportTasks/FAILED/2/20',
      data: {}
    })
    expect(result).toEqual({ code: '000000', data: { total: 1, records: [] } })
  })

  it('retries an export task through the retry alias', async () => {
    requestMock.post.mockResolvedValueOnce({ data: { code: '000000' } })

    const result = await exportRetry('task-9')

    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/exportCenter/retry/task-9',
      data: {}
    })
    expect(result).toEqual({ code: '000000' })
  })

  it('keeps non-blob dataset export requests on the export-center lifecycle contract', async () => {
    requestMock.post.mockResolvedValueOnce({
      code: '000000',
      data: { taskId: 'task-7', status: 'PENDING' }
    })

    const payload = { id: '12', filename: 'Orders' }
    const result = await exportDatasetData(payload)

    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/dataset/exportDataset',
      method: 'post',
      data: payload,
      loading: true
    })
    expect(result).toEqual({ code: '000000', data: { taskId: 'task-7', status: 'PENDING' } })
  })

  it('uses blob response for inline dataset export downloads', async () => {
    requestMock.post.mockResolvedValueOnce({ data: 'blob-data' })

    await exportDatasetData({ id: '12', dataEaseBi: true })

    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/dataset/exportDataset',
      method: 'post',
      data: { id: '12', dataEaseBi: true },
      loading: true,
      responseType: 'blob'
    })
  })

  it('posts dataset field deletion through the compatibility endpoint', async () => {
    requestMock.post.mockResolvedValueOnce({ data: null })

    await deleteField(705)

    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/datasetField/delete/705',
      data: {}
    })
  })

  it('posts dataset field bulk chart deletion through the compatibility endpoint', async () => {
    requestMock.post.mockResolvedValueOnce({ data: null })

    await deleteFieldByChartId(446)

    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/datasetField/deleteByChartId/446',
      data: {}
    })
  })

  it('migrates datasetTree CRUD wrappers to canonical dataset routes', async () => {
    requestMock.post.mockResolvedValueOnce({ data: { id: '1', name: 'test' } })
    requestMock.post.mockResolvedValueOnce({ data: { id: '2', name: 'new' } })
    requestMock.post.mockResolvedValueOnce({ data: { id: '1', name: 'renamed' } })
    requestMock.post.mockResolvedValueOnce({ data: { id: '1', pid: '0' } })
    requestMock.post.mockResolvedValueOnce({ data: null })
    requestMock.post.mockResolvedValueOnce({ data: true })

    await saveDatasetTree({ name: 'test', nodeType: 'dataset' } as any)
    await createDatasetTree({ name: 'new', nodeType: 'dataset' } as any)
    await renameDatasetTree({ id: 1, name: 'renamed' } as any)
    await moveDatasetTree({ id: 1, pid: 0 } as any)
    await delDatasetTree(1)
    await perDelete(1)

    expect(requestMock.post).toHaveBeenNthCalledWith(1, expect.objectContaining({ url: '/dataset/save' }))
    expect(requestMock.post).toHaveBeenNthCalledWith(2, expect.objectContaining({ url: '/dataset/create' }))
    expect(requestMock.post).toHaveBeenNthCalledWith(3, expect.objectContaining({ url: '/dataset/rename' }))
    expect(requestMock.post).toHaveBeenNthCalledWith(4, expect.objectContaining({ url: '/dataset/move' }))
    expect(requestMock.post).toHaveBeenNthCalledWith(5, expect.objectContaining({ url: '/dataset/delete/1' }))
    expect(requestMock.post).toHaveBeenNthCalledWith(6, expect.objectContaining({ url: '/dataset/perDelete/1' }))
  })

  it('migrates datasetTree and datasetData query wrappers to canonical dataset routes', async () => {
    requestMock.post.mockResolvedValueOnce({ data: {} })
    requestMock.get.mockResolvedValueOnce({ data: {} })
    requestMock.post.mockResolvedValueOnce({ data: { allFields: [] } })
    requestMock.post.mockResolvedValueOnce({ data: 0 })
    requestMock.post.mockResolvedValueOnce({ data: { id: '1', name: 'test' } })
    requestMock.post.mockResolvedValueOnce({ data: [] })
    requestMock.post.mockResolvedValueOnce({ data: [] })
    requestMock.post.mockResolvedValueOnce({ data: {} })
    requestMock.post.mockResolvedValueOnce({ data: [] })
    requestMock.post.mockResolvedValueOnce({ data: [] })

    await enumValueObj({ queryId: '1', searchText: '' })
    await barInfoApi(1)
    await getDatasetPreview(1)
    await getDatasetTotal(1)
    await getDatasetDetails(1)
    await getDsDetails({ ids: [1] })
    await getSqlParams({ ids: [1] })
    await getEnumValue({ fieldIds: [1] })
    await enumValueDs({ field: { id: 1 } })

    expect(requestMock.post).toHaveBeenNthCalledWith(1, expect.objectContaining({ url: '/dataset/enumValueObj' }))
    expect(requestMock.get).toHaveBeenCalledWith(expect.objectContaining({ url: '/dataset/barInfo/1' }))
    expect(requestMock.post).toHaveBeenNthCalledWith(2, expect.objectContaining({ url: '/dataset/get/1' }))
    expect(requestMock.post).toHaveBeenNthCalledWith(3, expect.objectContaining({ url: '/dataset/getDatasetTotal' }))
    expect(requestMock.post).toHaveBeenNthCalledWith(4, expect.objectContaining({ url: '/dataset/details/1' }))
    expect(requestMock.post).toHaveBeenNthCalledWith(5, expect.objectContaining({ url: '/dataset/dsDetails' }))
    expect(requestMock.post).toHaveBeenNthCalledWith(6, expect.objectContaining({ url: '/dataset/getSqlParams' }))
    expect(requestMock.post).toHaveBeenNthCalledWith(7, expect.objectContaining({ url: '/dataset/enumValue' }))
    expect(requestMock.post).toHaveBeenNthCalledWith(8, expect.objectContaining({ url: '/dataset/enumValueDs' }))
  })
})
