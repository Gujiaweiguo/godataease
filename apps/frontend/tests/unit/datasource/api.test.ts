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
  save
} from '@/api/datasource'

describe('Datasource API wrappers', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('keeps string id for validate endpoint', () => {
    const id = '985188400292302848'
    validateById(id)
    expect(requestMock.get).toHaveBeenCalledWith({ url: `/datasource/validate/${id}` })
  })

  it('keeps string id for delete endpoint', () => {
    const id = '985188400292302848'
    deleteById(id)
    expect(requestMock.get).toHaveBeenCalledWith({ url: `/datasource/delete/${id}` })
  })

  it('keeps string id for get/hidePw/getSimpleDs endpoints', () => {
    const id = '985188400292302848'
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
})
