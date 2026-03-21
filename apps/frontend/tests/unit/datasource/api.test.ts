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
  listDatasources,
  validateById,
  deleteById,
  getById,
  getHidePwById,
  getSimpleDs,
  uploadFile,
  save,
  validate,
  syncApiTable,
  checkApiItem
} from '@/api/datasource'

describe('Datasource API wrappers', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('keeps string id for validate endpoint', () => {
    const id = 9851884
    validateById(id)
    expect(requestMock.get).toHaveBeenCalledWith({ url: `/datasource/validate/${id}` })
  })

  it('keeps string id for delete endpoint', () => {
    const id = 9851884
    deleteById(id)
    expect(requestMock.get).toHaveBeenCalledWith({ url: `/datasource/delete/${id}` })
  })

  it('keeps string id for get/hidePw/getSimpleDs endpoints', () => {
    const id = 9851884
    getById(id)
    getHidePwById(id)
    getSimpleDs(id)

    expect(requestMock.get).toHaveBeenCalledWith({ url: `/datasource/get/${id}` })
    expect(requestMock.get).toHaveBeenCalledWith({ url: `/datasource/hidePw/${id}` })
    expect(requestMock.get).toHaveBeenCalledWith({ url: `/datasource/getSimpleDs/${id}` })
  })

  it('injects datasource busiFlag when listing trees', async () => {
    requestMock.post.mockResolvedValueOnce({ data: { list: [] } })
    const result = await listDatasources({ keyWord: 'demo' })

    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/datasource/tree',
      data: { keyWord: 'demo', busiFlag: 'datasource' }
    })
    expect(result).toEqual({ list: [] })
  })

  it('sends upload request with multipart headers and loading', async () => {
    requestMock.post.mockResolvedValueOnce({ code: '000000' })
    const payload = { file: 'mock-file' }
    await uploadFile(payload)

    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/datasource/uploadFile',
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
      url: '/datasource/save',
      data: { name: '  demo  ' }
    })
  })

  it('uses withDatasourceError decorator for validate method', async () => {
    requestMock.post.mockResolvedValueOnce({ data: { valid: true } })
    const result = await validate({ id: '123' })
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/datasource/validate',
      data: { id: '123' }
    })
    expect(result).toEqual({ data: { valid: true } })
  })

  it('uses withDatasourceError decorator for syncApiTable method', async () => {
    requestMock.post.mockResolvedValueOnce({ data: { tables: [] } })
    const result = await syncApiTable({ datasourceId: '123' })
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/datasource/syncApiTable',
      data: { datasourceId: '123' }
    })
    expect(result).toEqual({ data: { tables: [] } })
  })

  it('uses withDatasourceError decorator for checkApiItem method', async () => {
    requestMock.post.mockResolvedValueOnce({ data: { valid: true } })
    const result = await checkApiItem({ url: 'http://example.com' })
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/datasource/checkApiDatasource',
      data: { url: 'http://example.com' }
    })
    expect(result).toEqual({ data: { valid: true } })
  })

})
