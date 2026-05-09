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

import {
  embeddedQueryGridApi,
  embeddedCreateApi,
  embeddedEditApi,
  embeddedDeleteApi,
  embeddedBatchDeleteApi,
  embeddedResetApi,
  embeddedDomainListApi,
  embeddedInitIframeApi,
  embeddedGetTokenArgsApi
} from '@/api/embedded'

describe('Embedded API wrappers', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('embeddedQueryGridApi posts to pager endpoint', () => {
    const data = { keyword: 'test' }
    embeddedQueryGridApi(1, 10, data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/embedded/pager/1/10',
      data
    })
  })

  it('embeddedQueryGridApi uses different page/pageSize values', () => {
    embeddedQueryGridApi(3, 20, {})
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/embedded/pager/3/20',
      data: {}
    })
  })

  it('embeddedCreateApi posts to create endpoint', () => {
    const data = { name: 'app' }
    embeddedCreateApi(data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/embedded/create',
      data
    })
  })

  it('embeddedEditApi posts to edit endpoint', () => {
    const data = { id: '1', name: 'updated' }
    embeddedEditApi(data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/embedded/edit',
      data
    })
  })

  it('embeddedDeleteApi posts to delete endpoint with id', () => {
    embeddedDeleteApi('123')
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/embedded/delete/123'
    })
  })

  it('embeddedBatchDeleteApi posts to batchDelete endpoint', () => {
    const ids = ['1', '2', '3']
    embeddedBatchDeleteApi(ids)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/embedded/batchDelete',
      data: ids
    })
  })

  it('embeddedResetApi posts to reset endpoint', () => {
    const data = { id: '1' }
    embeddedResetApi(data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/embedded/reset',
      data
    })
  })

  it('embeddedDomainListApi gets domainList', () => {
    embeddedDomainListApi()
    expect(requestMock.get).toHaveBeenCalledWith({
      url: '/embedded/domainList'
    })
  })

  it('embeddedInitIframeApi posts to initIframe endpoint', () => {
    const data = { url: 'http://example.com' }
    embeddedInitIframeApi(data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/embedded/initIframe',
      data
    })
  })

  it('embeddedGetTokenArgsApi gets getTokenArgs', () => {
    embeddedGetTokenArgsApi()
    expect(requestMock.get).toHaveBeenCalledWith({
      url: '/embedded/getTokenArgs'
    })
  })

  it('embeddedCreateApi returns the request promise', async () => {
    requestMock.post.mockResolvedValueOnce({ data: { id: 'new' } })
    const result = await embeddedCreateApi({ name: 'app' })
    expect(result).toEqual({ data: { id: 'new' } })
  })

  it('embeddedDomainListApi returns the request promise', async () => {
    requestMock.get.mockResolvedValueOnce({ data: ['a.com'] })
    const result = await embeddedDomainListApi()
    expect(result).toEqual({ data: ['a.com'] })
  })
})
