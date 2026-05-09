import { describe, it, expect, vi, beforeEach } from 'vitest'

const { requestMock } = vi.hoisted(() => ({
  requestMock: {
    post: vi.fn().mockResolvedValue({ data: {} }),
    get: vi.fn().mockResolvedValue({ data: {} })
  }
}))

vi.mock('@/config/axios', () => ({ default: requestMock }))

import { queryMapKeyApi, queryMapKeyApiByType, saveMapKeyApi } from '@/api/setting/sysParameter'

describe('api/setting/sysParameter', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('queryMapKeyApi gets queryOnlineMap', () => {
    queryMapKeyApi()
    expect(requestMock.get).toHaveBeenCalledWith({ url: '/sysParameter/queryOnlineMap' })
  })

  it('queryMapKeyApiByType gets queryOnlineMap with type', () => {
    queryMapKeyApiByType('gaode')
    expect(requestMock.get).toHaveBeenCalledWith({ url: '/sysParameter/queryOnlineMap/gaode' })
  })

  it('saveMapKeyApi posts to saveOnlineMap', () => {
    const data = { key: 'test-key' }
    saveMapKeyApi(data)
    expect(requestMock.post).toHaveBeenCalledWith({ url: '/sysParameter/saveOnlineMap', data })
  })
})
