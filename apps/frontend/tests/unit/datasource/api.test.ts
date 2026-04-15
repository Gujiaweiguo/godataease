import { describe, it, expect, beforeEach, vi } from 'vitest'

const { requestMock, nameTrimMock } = vi.hoisted(() => ({
  requestMock: {
    post: vi.fn(),
    get: vi.fn()
  },
  nameTrimMock: vi.fn()
}))

vi.mock('@/config/axios', () => ({
  default: requestMock
}))

vi.mock('@/utils/utils', async () => {
  const { createUtilsModuleMock } = await import('../helpers')
  return createUtilsModuleMock({ nameTrim: nameTrimMock })
})

import {
  createFolder,
  listDatasourceTables,
  listDatasources,
  move,
  reName,
  validateById,
  checkRepeat,
  deleteById,
  getById,
  getHidePwById,
  getSchema,
  getTableField,
  getTableStatus,
  getSimpleDs,
  perDeleteDatasource,
  uploadFile,
  save,
  update,
  validate,
  syncApiTable,
  syncApiDs,
  checkApiItem
} from '@/api/datasource'
import { getDatasourceList, getTables } from '@/api/dataset'

describe('Datasource API wrappers', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('keeps string id for validate endpoint', () => {
    const id = 9851884
    validateById(id)
    expect(requestMock.get).toHaveBeenCalledWith({ url: `/ds/validate/${id}` })
  })

  it('uses canonical validation checking routes while keeping compatibility wrappers', async () => {
    requestMock.post.mockResolvedValueOnce({ data: false })
    requestMock.post.mockResolvedValueOnce({ data: { valid: true } })

    const repeatResult = await checkRepeat({ name: 'demo', type: 'mysql' })
    const apiResult = await checkApiItem({ url: 'http://example.com' })

    expect(requestMock.post).toHaveBeenNthCalledWith(1, {
      url: '/ds/checkRepeat',
      data: { name: 'demo', type: 'mysql' }
    })
    expect(requestMock.post).toHaveBeenNthCalledWith(2, {
      url: '/ds/checkApiDatasource',
      data: { url: 'http://example.com' }
    })
    expect(repeatResult).toBe(false)
    expect(apiResult).toEqual({ data: { valid: true } })
  })

  it('prefers post delete endpoint for datasource removal', async () => {
    const id = 9851884
    requestMock.post.mockResolvedValueOnce({ code: '000000' })

    await deleteById(id)

    expect(requestMock.post).toHaveBeenCalledWith({ url: `/ds/delete/${id}`, data: {} })
    expect(requestMock.get).not.toHaveBeenCalled()
  })

  it('keeps string id for get/hidePw/getSimpleDs endpoints', () => {
    const id = 9851884
    getById(id)
    getHidePwById(id)
    getSimpleDs(id)

    expect(requestMock.get).toHaveBeenCalledWith({ url: `/ds/${id}` })
    expect(requestMock.get).toHaveBeenCalledWith({ url: `/ds/hidePw/${id}` })
    expect(requestMock.get).toHaveBeenCalledWith({ url: `/ds/simple/${id}` })
  })

  it('uses canonical route for permanent datasource deletion', async () => {
    const id = 9851884
    requestMock.post.mockResolvedValueOnce({ data: true })

    await perDeleteDatasource(id)

    expect(requestMock.post).toHaveBeenCalledWith({ url: `/ds/perDelete/${id}`, data: {} })
  })

  it('injects datasource busiFlag when listing trees', async () => {
    requestMock.post.mockResolvedValueOnce({ data: { list: [] } })
    const result = await listDatasources({ keyWord: 'demo' })

    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/ds/tree',
      data: { keyWord: 'demo', busiFlag: 'datasource' }
    })
    expect(result).toEqual({ list: [] })
  })

  it('sends upload request with multipart headers and loading', async () => {
    requestMock.post.mockResolvedValueOnce({ code: '000000' })
    const payload = { file: 'mock-file' }
    await uploadFile(payload)

    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/ds/uploadFile',
      data: payload,
      loading: true,
      headersType: 'multipart/form-data;'
    })
  })

  it('trims payload before save', async () => {
    requestMock.post.mockResolvedValueOnce({ data: { id: '1' } })
    await save({ name: '  demo  ' })

    expect(nameTrimMock).toHaveBeenCalledTimes(1)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/ds/save',
      data: { name: '  demo  ' }
    })
  })

  it('trims payload before update and uses canonical route', async () => {
    requestMock.post.mockResolvedValueOnce({ data: { id: '1' } })
    await update({ id: '1', name: '  demo  ' })

    expect(nameTrimMock).toHaveBeenCalledTimes(1)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/ds/update',
      data: { id: '1', name: '  demo  ' }
    })
  })

  it('uses withDatasourceError decorator for validate method', async () => {
    requestMock.post.mockResolvedValueOnce({ data: { valid: true } })
    const result = await validate({ id: '123' })
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/ds/validate',
      data: { id: '123' }
    })
    expect(result).toEqual({ data: { valid: true } })
  })

  it('migrates tree and folder management wrappers to canonical datasource routes', async () => {
    requestMock.post.mockResolvedValueOnce({ data: { id: '1', pid: '0' } })
    requestMock.post.mockResolvedValueOnce({ data: { id: '1', name: 'rename' } })
    requestMock.post.mockResolvedValueOnce({ data: { id: '2', name: 'folder' } })

    const moved = await move({ id: '1', pid: '0' })
    const renamed = await reName({ id: '1', name: '  rename  ' })
    const folder = await createFolder({ pid: '0', name: '  folder  ' })

    expect(requestMock.post).toHaveBeenNthCalledWith(1, {
      url: '/ds/move',
      data: { id: '1', pid: '0' }
    })
    expect(requestMock.post).toHaveBeenNthCalledWith(2, {
      url: '/ds/reName',
      data: { id: '1', name: '  rename  ' }
    })
    expect(requestMock.post).toHaveBeenNthCalledWith(3, {
      url: '/ds/createFolder',
      data: { pid: '0', name: '  folder  ' }
    })
    expect(nameTrimMock).toHaveBeenCalledTimes(2)
    expect(moved).toEqual({ id: '1', pid: '0' })
    expect(renamed).toEqual({ id: '1', name: 'rename' })
    expect(folder).toEqual({ id: '2', name: 'folder' })
  })

  it('migrates dataset datasource selection wrappers to canonical datasource routes', async () => {
    requestMock.post.mockResolvedValueOnce({ data: [{ id: '1', name: 'folder' }] })
    requestMock.post.mockResolvedValueOnce({ data: [{ tableName: 'orders' }] })

    const tree = await getDatasourceList(3)
    const tables = await getTables({ datasourceId: '123' })

    expect(requestMock.post).toHaveBeenNthCalledWith(1, {
      url: '/ds/tree',
      data: { busiFlag: 'datasource', weight: 3 }
    })
    expect(requestMock.post).toHaveBeenNthCalledWith(2, {
      url: '/ds/tables',
      data: { datasourceId: '123' }
    })
    expect(tree).toEqual([{ id: '1', name: 'folder' }])
    expect(tables).toEqual([{ tableName: 'orders' }])
  })

  it('uses withDatasourceError decorator for syncApiTable method', async () => {
    requestMock.post.mockResolvedValueOnce({ data: { tables: [] } })
    const result = await syncApiTable({ datasourceId: '123' })
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/ds/syncApiTable',
      data: { datasourceId: '123' }
    })
    expect(result).toEqual({ data: { tables: [] } })
  })

  it('uses canonical route for syncApiDs method', async () => {
    requestMock.post.mockResolvedValueOnce({ data: { datasourceId: '123' } })
    const result = await syncApiDs({ datasourceId: '123' })
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/ds/syncApiDs',
      data: { datasourceId: '123' }
    })
    expect(result).toEqual({ data: { datasourceId: '123' } })
  })

  it('queries datasource table status through the stage2 compatibility endpoint', async () => {
    requestMock.post.mockResolvedValueOnce({
      data: [{ tableName: 'orders', status: 'Pending', lastUpdateTime: 1710000000000 }]
    })

    const result = await getTableStatus({ datasourceId: '123' })

    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/ds/tableStatus',
      data: { datasourceId: '123' }
    })
    expect(result).toEqual({
      data: [{ tableName: 'orders', status: 'Pending', lastUpdateTime: 1710000000000 }]
    })
  })

  it('queries datasource tables through the canonical datasource route', async () => {
    requestMock.post.mockResolvedValueOnce({ data: [{ tableName: 'orders' }] })

    const result = await listDatasourceTables({ datasourceId: '123' })

    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/ds/tables',
      data: { datasourceId: '123' }
    })
    expect(result).toEqual({ data: [{ tableName: 'orders' }] })
  })

  it('queries datasource table fields through the canonical datasource route', async () => {
    requestMock.post.mockResolvedValueOnce({ data: [{ fieldName: 'amount' }] })

    const result = await getTableField({ datasourceId: '123', tableName: 'orders' })

    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/ds/tableField',
      data: { datasourceId: '123', tableName: 'orders' }
    })
    expect(result).toEqual({ data: [{ fieldName: 'amount' }] })
  })

  it('queries datasource schema through the canonical datasource route', async () => {
    requestMock.post.mockResolvedValueOnce({ data: ['public'] })

    const result = await getSchema({ datasourceId: '123' })

    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/ds/schema',
      data: { datasourceId: '123' }
    })
    expect(result).toEqual({ data: ['public'] })
  })

  it('queries preview and sync datasource routes through canonical aliases', async () => {
    requestMock.post.mockResolvedValueOnce({ data: { rows: [] } })
    requestMock.post.mockResolvedValueOnce({ data: { tables: [] } })
    requestMock.post.mockResolvedValueOnce({ data: { datasourceId: '123' } })

    const { previewData } = await import('@/api/datasource')

    await previewData({ datasourceId: '123', tableName: 'orders' })
    await syncApiTable({ datasourceId: '123' })
    await syncApiDs({ datasourceId: '123' })

    expect(requestMock.post).toHaveBeenNthCalledWith(1, {
      url: '/ds/previewData',
      data: { datasourceId: '123', tableName: 'orders' }
    })
    expect(requestMock.post).toHaveBeenNthCalledWith(2, {
      url: '/ds/syncApiTable',
      data: { datasourceId: '123' }
    })
    expect(requestMock.post).toHaveBeenNthCalledWith(3, {
      url: '/ds/syncApiDs',
      data: { datasourceId: '123' }
    })
  })

  it('queries upload and remote datasource routes through canonical aliases', async () => {
    requestMock.post.mockResolvedValueOnce({ code: '000000' })
    requestMock.post.mockResolvedValueOnce({ data: { loaded: true } })

    await uploadFile({ file: 'mock-file' })
    const { loadRemoteFile } = await import('@/api/datasource')
    await loadRemoteFile({ url: 'http://example.com/demo.csv' })

    expect(requestMock.post).toHaveBeenNthCalledWith(1, {
      url: '/ds/uploadFile',
      data: { file: 'mock-file' },
      loading: true,
      headersType: 'multipart/form-data;'
    })
    expect(requestMock.post).toHaveBeenNthCalledWith(2, {
      url: '/ds/loadRemoteFile',
      data: { url: 'http://example.com/demo.csv' }
    })
  })

  it('keeps non-scoped datasource endpoints on compatibility aliases', async () => {
    requestMock.post.mockResolvedValueOnce({ data: { valid: true } })
    requestMock.get.mockResolvedValueOnce({ data: { supported: true } })

    await checkApiItem({ url: 'http://example.com' })
    const { supportSetKey } = await import('@/api/datasource')
    await supportSetKey()

    expect(requestMock.post).toHaveBeenNthCalledWith(1, {
      url: '/ds/checkApiDatasource',
      data: { url: 'http://example.com' }
    })
    expect(requestMock.get).toHaveBeenNthCalledWith(1, {
      url: '/engine/supportSetKey'
    })
  })

  it('uses withDatasourceError decorator for checkApiItem method', async () => {
    requestMock.post.mockResolvedValueOnce({ data: { valid: true } })
    const result = await checkApiItem({ url: 'http://example.com' })
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/ds/checkApiDatasource',
      data: { url: 'http://example.com' }
    })
    expect(result).toEqual({ data: { valid: true } })
  })

})
