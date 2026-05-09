import { describe, it, expect, vi, beforeEach } from 'vitest'

const { requestMock } = vi.hoisted(() => ({
  requestMock: {
    post: vi.fn().mockResolvedValue({ data: {} }),
    get: vi.fn().mockResolvedValue({ data: {} })
  }
}))

vi.mock('@/config/axios', () => ({ default: requestMock }))
vi.mock('@/views/visualized/data/dataset/form/util.js', () => ({ guid: () => 'mock-guid' }))
vi.mock('element-plus-secondary', () => ({
  ElMessage: { error: vi.fn(), success: vi.fn(), warning: vi.fn() }
}))

import { uploadFile, findResourceAsBase64 } from '@/api/staticResource'

describe('api/staticResource', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('uploadFile posts to staticResource upload with fileId and multipart', () => {
    const param = new FormData()
    param.append('file', new Blob(['test']))
    uploadFile('file-123', param)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/staticResource/upload/file-123',
      headersType: 'multipart/form-data',
      loading: true,
      data: param
    })
  })

  it('uploadFile accepts numeric fileId', () => {
    const param = new FormData()
    uploadFile(42, param)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/staticResource/upload/42',
      headersType: 'multipart/form-data',
      loading: true,
      data: param
    })
  })

  it('findResourceAsBase64 posts to findResourceAsBase64', () => {
    const params = { ids: ['res-1', 'res-2'] }
    findResourceAsBase64(params)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/staticResource/findResourceAsBase64',
      data: params
    })
  })
})
