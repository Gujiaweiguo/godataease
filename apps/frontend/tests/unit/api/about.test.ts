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

import { validateApi, buildVersionApi, updateInfoApi, revertApi } from '@/api/about'

describe('About API wrappers', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('validateApi posts to license/validate', () => {
    const data = { license: 'abc' }
    validateApi(data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/license/validate',
      data
    })
  })

  it('buildVersionApi gets license/version', () => {
    buildVersionApi()
    expect(requestMock.get).toHaveBeenCalledWith({ url: '/license/version' })
  })

  it('updateInfoApi posts to license/update', () => {
    const data = { license: 'new' }
    updateInfoApi(data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/license/update',
      data
    })
  })

  it('revertApi posts to license/revert', () => {
    revertApi()
    expect(requestMock.post).toHaveBeenCalledWith({ url: '/license/revert' })
  })

  it('validateApi returns the request promise', async () => {
    requestMock.post.mockResolvedValueOnce({ data: { valid: true } })
    const result = await validateApi({ license: 'abc' })
    expect(result).toEqual({ data: { valid: true } })
  })

  it('buildVersionApi returns the request promise', async () => {
    requestMock.get.mockResolvedValueOnce({ data: { version: '1.0' } })
    const result = await buildVersionApi()
    expect(result).toEqual({ data: { version: '1.0' } })
  })
})
