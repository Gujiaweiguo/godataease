import { describe, it, expect, vi, beforeEach } from 'vitest'

const { requestMock } = vi.hoisted(() => ({
  requestMock: {
    post: vi.fn().mockResolvedValue({ data: {} }),
    get: vi.fn().mockResolvedValue({ data: {} })
  }
}))

vi.mock('@/config/axios', () => ({ default: requestMock }))

import { watermarkSave, watermarkFind } from '@/api/watermark'

describe('api/watermark', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('watermarkSave posts to save with params', () => {
    const params = { enable: true, content: 'test' }
    watermarkSave(params)
    expect(requestMock.post).toHaveBeenCalledWith({ url: '/watermark/save', data: params })
  })

  it('watermarkFind gets from find', () => {
    watermarkFind()
    expect(requestMock.get).toHaveBeenCalledWith({ url: '/watermark/find' })
  })
})
