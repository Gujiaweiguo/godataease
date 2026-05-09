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
  getDatasourceListByTypeApi,
  getTaskInfoListApi,
  executeOneApi,
  startTaskApi,
  stopTaskApi,
  addApi,
  removeApi,
  batchRemoveApi,
  modifyApi,
  findTaskInfoByIdApi,
  getDatasourceTableListApi
} from '@/api/sync/syncTask'

describe('SyncTask API wrappers', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('getDatasourceListByTypeApi gets datasource list by type', () => {
    getDatasourceListByTypeApi('mysql')
    expect(requestMock.get).toHaveBeenCalledWith({
      url: '/sync/datasource/list/mysql'
    })
  })

  it('getDatasourceListByTypeApi uses different type', () => {
    getDatasourceListByTypeApi('doris')
    expect(requestMock.get).toHaveBeenCalledWith({
      url: '/sync/datasource/list/doris'
    })
  })

  it('getTaskInfoListApi posts to task pager endpoint', () => {
    const data = { keyword: 'task1' }
    getTaskInfoListApi(1, 10, data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/sync/task/pager/1/10',
      data
    })
  })

  it('executeOneApi gets execute endpoint with id', () => {
    executeOneApi('task1')
    expect(requestMock.get).toHaveBeenCalledWith({
      url: '/sync/task/execute/task1'
    })
  })

  it('startTaskApi gets start endpoint with id', () => {
    startTaskApi('task1')
    expect(requestMock.get).toHaveBeenCalledWith({
      url: '/sync/task/start/task1'
    })
  })

  it('stopTaskApi gets stop endpoint with id', () => {
    stopTaskApi('task1')
    expect(requestMock.get).toHaveBeenCalledWith({
      url: '/sync/task/stop/task1'
    })
  })

  it('addApi posts to add endpoint', () => {
    const data = { name: 'newTask', source: {}, target: {} }
    addApi(data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/sync/task/add',
      data
    })
  })

  it('removeApi posts to remove endpoint with taskId', () => {
    removeApi('task1')
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/sync/task/remove/task1'
    })
  })

  it('batchRemoveApi posts to batch/del endpoint', () => {
    const ids = ['task1', 'task2']
    batchRemoveApi(ids)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/sync/task/batch/del',
      data: ids
    })
  })

  it('modifyApi posts to update endpoint', () => {
    const data = { id: 'task1', name: 'updated' }
    modifyApi(data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/sync/task/update',
      data
    })
  })

  it('findTaskInfoByIdApi gets task by id', () => {
    findTaskInfoByIdApi('task1')
    expect(requestMock.get).toHaveBeenCalledWith({
      url: '/sync/task/get/task1'
    })
  })

  it('getDatasourceTableListApi gets table list by dsId', () => {
    getDatasourceTableListApi('ds1')
    expect(requestMock.get).toHaveBeenCalledWith({
      url: '/sync/datasource/table/list/ds1'
    })
  })

  it('getTaskInfoListApi returns the request promise', async () => {
    requestMock.post.mockResolvedValueOnce({ data: { total: 25 } })
    const result = await getTaskInfoListApi(1, 10, {})
    expect(result).toEqual({ data: { total: 25 } })
  })
})
