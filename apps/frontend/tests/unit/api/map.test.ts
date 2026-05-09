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
  getWorldTree,
  getGeoJson,
  listCustomGeoArea,
  getCustomGeoArea,
  deleteCustomGeoArea,
  saveCustomGeoArea,
  deleteCustomGeoSubArea,
  saveCustomGeoSubArea,
  listSubAreaOptions
} from '@/api/map'

describe('Map API wrappers', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('getWorldTree gets worldTree', () => {
    getWorldTree()
    expect(requestMock.get).toHaveBeenCalledWith({ url: '/map/worldTree' })
  })

  it('getGeoJson builds url for standard areaId', () => {
    getGeoJson('156110000')
    expect(requestMock.get).toHaveBeenCalledWith({ url: '/map/156/156110000.json' })
  })

  it('getGeoJson builds url for custom geo areaId (geo_ prefix)', () => {
    getGeoJson('geo_156110000')
    expect(requestMock.get).toHaveBeenCalledWith({ url: '/geo/156/156110000.json' })
  })

  it('getGeoJson extracts busi geo code after geo_ prefix', () => {
    getGeoJson('geo_999000000')
    expect(requestMock.get).toHaveBeenCalledWith({ url: '/geo/999/999000000.json' })
  })

  it('listCustomGeoArea gets custom geo area list', () => {
    listCustomGeoArea()
    expect(requestMock.get).toHaveBeenCalledWith({ url: '/customGeo/geoArea/list' })
  })

  it('getCustomGeoArea gets specific geo area by id', () => {
    getCustomGeoArea('area1')
    expect(requestMock.get).toHaveBeenCalledWith({ url: '/customGeo/geoArea/area1' })
  })

  it('deleteCustomGeoArea deletes geo area by id', () => {
    deleteCustomGeoArea('area1')
    expect(requestMock.delete).toHaveBeenCalledWith({ url: '/customGeo/geoArea/area1' })
  })

  it('saveCustomGeoArea posts to save endpoint', () => {
    const area = { id: '1', name: 'test' } as any
    saveCustomGeoArea(area)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/customGeo/geoArea/save',
      data: area
    })
  })

  it('deleteCustomGeoSubArea deletes sub area by id', () => {
    deleteCustomGeoSubArea('sub1')
    expect(requestMock.delete).toHaveBeenCalledWith({ url: '/customGeo/geoSubArea/sub1' })
  })

  it('saveCustomGeoSubArea posts to save endpoint', () => {
    const area = { id: '1', name: 'sub' } as any
    saveCustomGeoSubArea(area)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/customGeo/geoSubArea/save',
      data: area
    })
  })

  it('listSubAreaOptions gets sub area options', () => {
    listSubAreaOptions()
    expect(requestMock.get).toHaveBeenCalledWith({ url: '/customGeo/geoSubArea/options' })
  })

  it('getWorldTree returns the request promise', async () => {
    requestMock.get.mockResolvedValueOnce({ data: { children: [] } })
    const result = await getWorldTree()
    expect(result).toEqual({ data: { children: [] } })
  })

  it('deleteCustomGeoArea returns the request promise', async () => {
    requestMock.delete.mockResolvedValueOnce({ data: { success: true } })
    const result = await deleteCustomGeoArea('area1')
    expect(result).toEqual({ data: { success: true } })
  })
})
