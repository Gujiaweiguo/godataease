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
  sourceDsPageApi,
  targetDsPageApi,
  latestUseApi,
  validateApi,
  getSchemaApi,
  saveApi,
  getByIdApi,
  updateApi,
  deleteByIdApi,
  batchDelApi,
  getFieldListApi,
  validateByIdApi
} from '@/api/sync/syncDatasource'

describe('SyncDatasource API wrappers', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('sourceDsPageApi posts to source pager endpoint', () => {
    const data = { keyword: 'mysql' }
    sourceDsPageApi(1, 10, data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/sync/datasource/source/pager/1/10',
      data
    })
  })

  it('targetDsPageApi posts to target pager endpoint', () => {
    const data = { keyword: 'doris' }
    targetDsPageApi(2, 20, data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/sync/datasource/target/pager/2/20',
      data
    })
  })

  it('latestUseApi posts to latestUse with sourceType', () => {
    latestUseApi('mysql')
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/sync/datasource/latestUse/mysql',
      data: {}
    })
  })

  it('validateApi posts to validate endpoint', () => {
    const data = { config: { host: 'localhost' } }
    validateApi(data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/sync/datasource/validate',
      data
    })
  })

  it('getSchemaApi posts to getSchema endpoint', () => {
    const data = { datasourceId: 'ds1' }
    getSchemaApi(data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/sync/datasource/getSchema',
      data
    })
  })

  it('saveApi posts to save endpoint', () => {
    const data = { name: 'newDs' }
    saveApi(data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/sync/datasource/save',
      data
    })
  })

  it('getByIdApi gets datasource by id', () => {
    getByIdApi('ds123')
    expect(requestMock.get).toHaveBeenCalledWith({
      url: '/sync/datasource/get/ds123'
    })
  })

  it('updateApi posts to update endpoint', () => {
    const data = { id: 'ds1', name: 'updated' }
    updateApi(data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/sync/datasource/update',
      data
    })
  })

  it('deleteByIdApi posts to delete endpoint with id', () => {
    deleteByIdApi('ds1')
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/sync/datasource/delete/ds1'
    })
  })

  it('batchDelApi posts to batchDel endpoint', () => {
    const ids = ['1', '2']
    batchDelApi(ids)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/sync/datasource/batchDel',
      data: ids
    })
  })

  it('getFieldListApi posts to fields endpoint', () => {
    const data = { datasourceId: 'ds1', table: 't1' }
    getFieldListApi(data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/sync/datasource/fields',
      data
    })
  })

  it('validateByIdApi gets validate by id', () => {
    validateByIdApi('ds1')
    expect(requestMock.get).toHaveBeenCalledWith({
      url: '/sync/datasource/validate/ds1'
    })
  })

  it('sourceDsPageApi returns the request promise', async () => {
    requestMock.post.mockResolvedValueOnce({ data: { total: 100 } })
    const result = await sourceDsPageApi(1, 10, {})
    expect(result).toEqual({ data: { total: 100 } })
  })
})
