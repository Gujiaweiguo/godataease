import { describe, it, expect, vi, beforeEach } from 'vitest'

const { requestMock } = vi.hoisted(() => ({
  requestMock: {
    post: vi.fn().mockResolvedValue({ data: {} }),
    get: vi.fn().mockResolvedValue({ data: {} })
  }
}))

vi.mock('@/config/axios', () => ({ default: requestMock }))

import {
  variableCreateApi,
  variableEditApi,
  variableDetailApi,
  variableDeletelApi,
  searchVariableApi,
  valueSelectedForVariableApi,
  valueForVariable,
  variableValueCreateApi,
  variableValueDeletelApi,
  variableValueEditApi,
  batchDelApi
} from '@/api/variable'

describe('api/variable', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('variableCreateApi posts to create', () => {
    const data = { name: 'var1' }
    variableCreateApi(data)
    expect(requestMock.post).toHaveBeenCalledWith({ url: '/sysVariable/create', data })
  })

  it('variableEditApi posts to edit', () => {
    const data = { id: 1, name: 'updated' }
    variableEditApi(data)
    expect(requestMock.post).toHaveBeenCalledWith({ url: '/sysVariable/edit', data })
  })

  it('variableDetailApi gets detail by id', () => {
    variableDetailApi(5)
    expect(requestMock.get).toHaveBeenCalledWith({ url: '/sysVariable/detail/5' })
  })

  it('variableDeletelApi gets delete by id', () => {
    variableDeletelApi(3)
    expect(requestMock.get).toHaveBeenCalledWith({ url: '/sysVariable/delete/3' })
  })

  it('searchVariableApi posts to query', () => {
    const data = { keyword: '' }
    searchVariableApi(data)
    expect(requestMock.post).toHaveBeenCalledWith({ url: '/sysVariable/query', data })
  })

  it('valueSelectedForVariableApi posts to value selected with pagination', () => {
    const data = { variableId: 1 }
    valueSelectedForVariableApi(1, 20, data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/sysVariable/value/selected/1/20',
      data
    })
  })

  it('valueForVariable gets value selected by id', () => {
    valueForVariable(7)
    expect(requestMock.get).toHaveBeenCalledWith({ url: '/sysVariable/value/selected/7' })
  })

  it('variableValueCreateApi posts to value create', () => {
    const data = { variableId: 1, value: 'val1' }
    variableValueCreateApi(data)
    expect(requestMock.post).toHaveBeenCalledWith({ url: '/sysVariable/value/create', data })
  })

  it('variableValueDeletelApi gets value delete by id', () => {
    variableValueDeletelApi(9)
    expect(requestMock.get).toHaveBeenCalledWith({ url: '/sysVariable/value/delete/9' })
  })

  it('variableValueEditApi posts to value edit', () => {
    const data = { id: 1, value: 'updated' }
    variableValueEditApi(data)
    expect(requestMock.post).toHaveBeenCalledWith({ url: '/sysVariable/value/edit', data })
  })

  it('batchDelApi posts to value batchDel', () => {
    const data = { ids: [1, 2, 3] }
    batchDelApi(data)
    expect(requestMock.post).toHaveBeenCalledWith({ url: '/sysVariable/value/batchDel', data })
  })
})
