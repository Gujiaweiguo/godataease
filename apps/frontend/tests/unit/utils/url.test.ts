import { describe, it, expect, vi, beforeEach } from 'vitest'

const { mockStore } = vi.hoisted(() => ({
  mockStore: { baseUrl: '' }
}))

vi.mock('@/store/modules/embedded', () => ({
  useEmbedded: () => mockStore
}))

import { formatDataEaseBi } from '@/utils/url'

describe('url', () => {
  beforeEach(() => {
    mockStore.baseUrl = ''
  })

  it('should return url unchanged when baseUrl is empty string', () => {
    expect(formatDataEaseBi('/api/data')).toBe('/api/data')
  })

  it('should prepend baseUrl when it is set', () => {
    mockStore.baseUrl = '/embedded'
    expect(formatDataEaseBi('/api/data')).toBe('/embedded/api/data')
  })

  it('should return url unchanged when baseUrl is empty after reset', () => {
    mockStore.baseUrl = '/dash'
    mockStore.baseUrl = ''
    expect(formatDataEaseBi('/chart/view')).toBe('/chart/view')
  })

  it('should handle full URL prefix', () => {
    mockStore.baseUrl = 'https://example.com'
    expect(formatDataEaseBi('/resource/1')).toBe('https://example.com/resource/1')
  })

  it('should handle url without leading slash', () => {
    mockStore.baseUrl = '/base'
    expect(formatDataEaseBi('panel/edit')).toBe('/basepanel/edit')
  })

  it('should handle empty url string with no baseUrl', () => {
    expect(formatDataEaseBi('')).toBe('')
  })

  it('should handle empty url string with baseUrl set', () => {
    mockStore.baseUrl = '/prefix'
    expect(formatDataEaseBi('')).toBe('/prefix')
  })

  it('should preserve query parameters in url', () => {
    mockStore.baseUrl = '/base'
    expect(formatDataEaseBi('/api/data?id=1&name=test')).toBe('/base/api/data?id=1&name=test')
  })

  it('should handle hash in url', () => {
    mockStore.baseUrl = '/base'
    expect(formatDataEaseBi('/page#section')).toBe('/base/page#section')
  })
})
