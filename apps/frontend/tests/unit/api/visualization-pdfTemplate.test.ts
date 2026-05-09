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

import { queryAll } from '@/api/visualization/pdfTemplate'

describe('PdfTemplate API wrappers', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('queryAll gets pdf-template/queryAll with loading false', () => {
    queryAll()
    expect(requestMock.get).toHaveBeenCalledWith({
      url: '/pdf-template/queryAll',
      loading: false
    })
  })

  it('queryAll returns the request promise', async () => {
    requestMock.get.mockResolvedValueOnce({ data: [{ id: 'tpl1' }] })
    const result = await queryAll()
    expect(result).toEqual({ data: [{ id: 'tpl1' }] })
  })
})
