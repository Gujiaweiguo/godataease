import { describe, it, expect, vi, beforeEach } from 'vitest'

const mockAddLoading = vi.hoisted(() => vi.fn())
const mockReduceLoading = vi.hoisted(() => vi.fn())
const mockLoadingMap: Record<string, number> = vi.hoisted(() => ({}))

vi.mock('@/store/modules/request', () => ({
  useRequestStoreWithOut: () => ({
    addLoading: mockAddLoading,
    reduceLoading: mockReduceLoading,
    loadingMap: mockLoadingMap
  })
}))

import { tryShowLoading, tryHideLoading } from '@/utils/loading'

describe('loading utils', () => {
  beforeEach(() => {
    mockAddLoading.mockReset()
    mockReduceLoading.mockReset()
    for (const key in mockLoadingMap) {
      delete mockLoadingMap[key]
    }
  })

  describe('tryShowLoading', () => {
    it('should call addLoading with the identification', () => {
      tryShowLoading('page-a')
      expect(mockAddLoading).toHaveBeenCalledWith('page-a')
    })

    it('should not call addLoading when identification is falsy (empty string)', () => {
      tryShowLoading('')
      expect(mockAddLoading).not.toHaveBeenCalled()
    })

    it('should not call addLoading when identification is null', () => {
      tryShowLoading(null)
      expect(mockAddLoading).not.toHaveBeenCalled()
    })

    it('should not call addLoading when identification is undefined', () => {
      tryShowLoading(undefined)
      expect(mockAddLoading).not.toHaveBeenCalled()
    })

    it('should work with string identification', () => {
      tryShowLoading('my-view')
      expect(mockAddLoading).toHaveBeenCalledTimes(1)
    })
  })

  describe('tryHideLoading', () => {
    it('should call reduceLoading when count is positive', () => {
      mockLoadingMap['page-a'] = 2
      tryHideLoading('page-a')
      expect(mockReduceLoading).toHaveBeenCalledWith('page-a')
    })

    it('should not call reduceLoading when count is 0', () => {
      mockLoadingMap['page-b'] = 0
      tryHideLoading('page-b')
      expect(mockReduceLoading).not.toHaveBeenCalled()
    })

    it('should not call reduceLoading when identification is falsy', () => {
      tryHideLoading('')
      tryHideLoading(null)
      tryHideLoading(undefined)
      expect(mockReduceLoading).not.toHaveBeenCalled()
    })

    it('should not call reduceLoading when identification is not in loadingMap', () => {
      tryHideLoading('nonexistent')
      expect(mockReduceLoading).not.toHaveBeenCalled()
    })

    it('should call reduceLoading when count is exactly 1', () => {
      mockLoadingMap['page-c'] = 1
      tryHideLoading('page-c')
      expect(mockReduceLoading).toHaveBeenCalledWith('page-c')
    })
  })
})
