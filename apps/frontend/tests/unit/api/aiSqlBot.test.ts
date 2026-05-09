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

import { findDvSqlBotDataset } from '@/api/aiSqlBot'

describe('AiSqlBot API wrappers', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('findDvSqlBotDataset gets dataset by dvInfo', () => {
    findDvSqlBotDataset('dv123')
    expect(requestMock.get).toHaveBeenCalledWith({
      url: '/sqlbot/dataset/dv123'
    })
  })

  it('findDvSqlBotDataset uses different dvInfo values', () => {
    findDvSqlBotDataset('abc456')
    expect(requestMock.get).toHaveBeenCalledWith({
      url: '/sqlbot/dataset/abc456'
    })
  })

  it('findDvSqlBotDataset returns the request promise', async () => {
    requestMock.get.mockResolvedValueOnce({ data: { id: 'ds1' } })
    const result = await findDvSqlBotDataset('dv1')
    expect(result).toEqual({ data: { id: 'ds1' } })
  })
})
