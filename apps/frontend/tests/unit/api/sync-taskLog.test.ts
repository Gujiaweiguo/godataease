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
  getTaskLogListApi,
  removeApi,
  getTaskLogDetailApi,
  clear,
  terminationTaskApi
} from '@/api/sync/syncTaskLog'

describe('SyncTaskLog API wrappers', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('getTaskLogListApi posts to task log pager endpoint', () => {
    const data = { keyword: 'error' }
    getTaskLogListApi(1, 10, data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/sync/task/log/pager/1/10',
      data
    })
  })

  it('getTaskLogListApi uses different page values', () => {
    getTaskLogListApi(3, 20, {})
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/sync/task/log/pager/3/20',
      data: {}
    })
  })

  it('removeApi posts to delete endpoint with logId', () => {
    removeApi('log1')
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/sync/task/log/delete/log1'
    })
  })

  it('getTaskLogDetailApi gets detail with logId and fromLineNum', () => {
    getTaskLogDetailApi('log1', 100)
    expect(requestMock.get).toHaveBeenCalledWith({
      url: '/sync/task/log/detail/log1/100'
    })
  })

  it('getTaskLogDetailApi uses different fromLineNum', () => {
    getTaskLogDetailApi('log2', 0)
    expect(requestMock.get).toHaveBeenCalledWith({
      url: '/sync/task/log/detail/log2/0'
    })
  })

  it('clear posts to clear endpoint with data', () => {
    const clearData = { taskId: 'task1' }
    clear(clearData)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/sync/task/log/clear',
      data: clearData
    })
  })

  it('clear works with empty object', () => {
    clear({})
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/sync/task/log/clear',
      data: {}
    })
  })

  it('terminationTaskApi posts to terminationTask with logId', () => {
    terminationTaskApi('log1')
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/sync/task/log/terminationTask/log1',
      data: {}
    })
  })

  it('getTaskLogListApi returns the request promise', async () => {
    requestMock.post.mockResolvedValueOnce({ data: { total: 50 } })
    const result = await getTaskLogListApi(1, 10, {})
    expect(result).toEqual({ data: { total: 50 } })
  })
})
