import { describe, it, expect, vi, beforeEach } from 'vitest'

const { requestMock } = vi.hoisted(() => ({
  requestMock: {
    post: vi.fn().mockResolvedValue({ data: {} }),
    get: vi.fn().mockResolvedValue({ data: {} })
  }
}))

vi.mock('@/config/axios', () => ({ default: requestMock }))

import {
  searchMarket,
  searchMarketRecommend,
  searchMarketPreview,
  getCategories,
  getCategoriesObject
} from '@/api/templateMarket'

describe('api/templateMarket', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('searchMarket gets templateMarket search', () => {
    searchMarket()
    expect(requestMock.get).toHaveBeenCalledWith({ url: '/templateMarket/search' })
  })

  it('searchMarketRecommend gets templateMarket searchRecommend', () => {
    searchMarketRecommend()
    expect(requestMock.get).toHaveBeenCalledWith({ url: '/templateMarket/searchRecommend' })
  })

  it('searchMarketPreview gets templateMarket searchPreview', () => {
    searchMarketPreview()
    expect(requestMock.get).toHaveBeenCalledWith({ url: '/templateMarket/searchPreview' })
  })

  it('getCategories gets templateMarket categories', () => {
    getCategories()
    expect(requestMock.get).toHaveBeenCalledWith({ url: '/templateMarket/categories' })
  })

  it('getCategoriesObject gets templateMarket categoriesObject', () => {
    getCategoriesObject()
    expect(requestMock.get).toHaveBeenCalledWith({ url: '/templateMarket/categoriesObject' })
  })
})
