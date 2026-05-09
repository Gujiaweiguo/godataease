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

import { findBaseParams } from '@/api/aiComponent'

describe('AiComponent API wrappers', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('findBaseParams gets aiBase/findTargetUrl', () => {
    findBaseParams()
    expect(requestMock.get).toHaveBeenCalledWith({ url: '/aiBase/findTargetUrl' })
  })

  it('findBaseParams returns the request promise', async () => {
    requestMock.get.mockResolvedValueOnce({ data: { url: 'http://ai.example.com' } })
    const result = await findBaseParams()
    expect(result).toEqual({ data: { url: 'http://ai.example.com' } })
  })
})
