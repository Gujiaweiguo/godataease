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

import { queryVisualizationBackground } from '@/api/visualization/visualizationBackground'

describe('VisualizationBackground API wrappers', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('queryVisualizationBackground gets findAll endpoint', () => {
    queryVisualizationBackground()
    expect(requestMock.get).toHaveBeenCalledWith({
      url: '/visualizationBackground/findAll'
    })
  })

  it('queryVisualizationBackground returns the request promise', async () => {
    requestMock.get.mockResolvedValueOnce({ data: [{ id: 'bg1', url: '/bg.png' }] })
    const result = await queryVisualizationBackground()
    expect(result).toEqual({ data: [{ id: 'bg1', url: '/bg.png' }] })
  })
})
