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

import { getRoleRouters, getDefaultSettings } from '@/api/common'

describe('Common API wrappers', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('getRoleRouters gets roleRouter/query', async () => {
    requestMock.get.mockResolvedValueOnce({ data: [{ path: '/home' }] })
    const result = await getRoleRouters()
    expect(requestMock.get).toHaveBeenCalledWith({ url: '/roleRouter/query' })
    expect(result).toEqual([{ path: '/home' }])
  })

  it('getRoleRouters returns undefined when res is null', async () => {
    requestMock.get.mockResolvedValueOnce(null)
    const result = await getRoleRouters()
    expect(result).toBeUndefined()
  })

  it('getDefaultSettings gets sysParameter/defaultSettings', async () => {
    requestMock.get.mockResolvedValueOnce({ data: { sort: 'name' } })
    const result = await getDefaultSettings()
    expect(requestMock.get).toHaveBeenCalledWith({ url: '/sysParameter/defaultSettings' })
    expect(result).toEqual({ sort: 'name' })
  })

  it('getDefaultSettings returns undefined when res is null', async () => {
    requestMock.get.mockResolvedValueOnce(null)
    const result = await getDefaultSettings()
    expect(result).toBeUndefined()
  })
})
