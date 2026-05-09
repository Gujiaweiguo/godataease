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

import { msgCountApi } from '@/api/msg'

describe('Msg API wrappers', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('msgCountApi posts to msg-center/count with empty data', () => {
    msgCountApi()
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/msg-center/count',
      data: {}
    })
  })

  it('msgCountApi returns the request promise', async () => {
    requestMock.post.mockResolvedValueOnce({ data: { count: 5 } })
    const result = await msgCountApi()
    expect(result).toEqual({ data: { count: 5 } })
  })
})
