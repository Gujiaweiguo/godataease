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
  getDatasourceRelationship,
  getDatasetRelationship,
  getPanelRelationship,
  resourceCheckPermission
} from '@/api/relation/index'

describe('Relation API wrappers', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('getDatasourceRelationship posts to relation/datasource/:id', () => {
    getDatasourceRelationship('ds1')
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/relation/datasource/ds1'
    })
  })

  it('getDatasetRelationship posts to relation/dataset/:id', () => {
    getDatasetRelationship('set1')
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/relation/dataset/set1'
    })
  })

  it('getPanelRelationship posts to relation/dv/:id', () => {
    getPanelRelationship('dv1')
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/relation/dv/dv1'
    })
  })

  it('resourceCheckPermission posts to resource/checkPermission/:id', () => {
    resourceCheckPermission('res1')
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/resource/checkPermission/res1'
    })
  })

  it('getDatasourceRelationship returns the request promise', async () => {
    requestMock.post.mockResolvedValueOnce({ data: { nodes: [] } })
    const result = await getDatasourceRelationship('ds1')
    expect(result).toEqual({ data: { nodes: [] } })
  })

  it('getDatasetRelationship returns the request promise', async () => {
    requestMock.post.mockResolvedValueOnce({ data: { relations: [] } })
    const result = await getDatasetRelationship('set1')
    expect(result).toEqual({ data: { relations: [] } })
  })

  it('getPanelRelationship returns the request promise', async () => {
    requestMock.post.mockResolvedValueOnce({ data: { panels: [] } })
    const result = await getPanelRelationship('dv1')
    expect(result).toEqual({ data: { panels: [] } })
  })

  it('resourceCheckPermission returns the request promise', async () => {
    requestMock.post.mockResolvedValueOnce({ data: { permitted: true } })
    const result = await resourceCheckPermission('res1')
    expect(result).toEqual({ data: { permitted: true } })
  })
})
