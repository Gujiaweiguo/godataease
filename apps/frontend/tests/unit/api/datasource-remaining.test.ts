import { describe, it, expect, vi, beforeEach } from 'vitest'

const { requestMock } = vi.hoisted(() => ({
  requestMock: {
    post: vi.fn().mockResolvedValue({ data: {} }),
    get: vi.fn().mockResolvedValue({ data: {} })
  }
}))

vi.mock('@/config/axios', () => ({ default: requestMock }))

import { getDatasetTree, getDsTree, getDeEngine } from '@/api/datasource'

describe('Datasource API wrappers (remaining uncovered)', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('getDatasetTree posts to dataset/tree with busiFlag dataset', async () => {
    requestMock.post.mockResolvedValueOnce({ data: [{ id: '1', name: 'root' }] })
    const result = await getDatasetTree({ keyWord: 'sales' })
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/dataset/tree',
      data: { keyWord: 'sales', busiFlag: 'dataset' }
    })
    expect(result).toEqual([{ id: '1', name: 'root' }])
  })

  it('getDatasetTree merges busiFlag into data even when data is empty', async () => {
    requestMock.post.mockResolvedValueOnce({ data: [] })
    const result = await getDatasetTree()
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/dataset/tree',
      data: { busiFlag: 'dataset' }
    })
    expect(result).toEqual([])
  })

  it('getDatasetTree preserves existing properties alongside busiFlag', async () => {
    requestMock.post.mockResolvedValueOnce({ data: [{ id: '2' }] })
    await getDatasetTree({ pid: '0', weight: 5 })
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/dataset/tree',
      data: { pid: '0', weight: 5, busiFlag: 'dataset' }
    })
  })

  it('getDsTree posts to ds/tree with busiFlag datasource', async () => {
    requestMock.post.mockResolvedValueOnce({ data: [{ id: '10', name: 'ds-root' }] })
    const result = await getDsTree({ keyWord: 'mysql' })
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/ds/tree',
      data: { keyWord: 'mysql', busiFlag: 'datasource' }
    })
    expect(result).toEqual([{ id: '10', name: 'ds-root' }])
  })

  it('getDsTree merges busiFlag into data when data is empty', async () => {
    requestMock.post.mockResolvedValueOnce({ data: [] })
    const result = await getDsTree()
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/ds/tree',
      data: { busiFlag: 'datasource' }
    })
    expect(result).toEqual([])
  })

  it('getDsTree preserves existing properties alongside busiFlag', async () => {
    requestMock.post.mockResolvedValueOnce({ data: [{ id: '3' }] })
    await getDsTree({ pid: '1', weight: 7 })
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/ds/tree',
      data: { pid: '1', weight: 7, busiFlag: 'datasource' }
    })
  })

  it('getDatasetTree returns null when response data is null', async () => {
    requestMock.post.mockResolvedValueOnce({ data: null })
    const result = await getDatasetTree()
    expect(result).toBeNull()
  })

  it('getDsTree returns null when response data is null', async () => {
    requestMock.post.mockResolvedValueOnce({ data: null })
    const result = await getDsTree()
    expect(result).toBeNull()
  })

  it('getDeEngine fetches engine configuration', () => {
    getDeEngine()
    expect(requestMock.get).toHaveBeenCalledWith({ url: '/engine/getEngine' })
  })

  it('getDeEngine returns the request promise', async () => {
    requestMock.get.mockResolvedValueOnce({ data: { engine: 'mysql' } })
    const result = await getDeEngine()
    expect(result).toEqual({ data: { engine: 'mysql' } })
  })
})
