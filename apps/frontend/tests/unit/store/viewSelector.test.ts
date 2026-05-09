import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { viewSelectorStore } from '@/store/modules/data-visualization/viewSelector'

describe('ViewSelector Store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  describe('Initial State', () => {
    it('should have enable as false', () => {
      const store = viewSelectorStore()
      expect(store.enable).toBe(false)
    })

    it('should have empty viewIdList', () => {
      const store = viewSelectorStore()
      expect(store.viewIdList).toEqual([])
    })

    it('should return false from getEnable getter', () => {
      const store = viewSelectorStore()
      expect(store.getEnable).toBe(false)
    })

    it('should return empty array from getViewIdList getter', () => {
      const store = viewSelectorStore()
      expect(store.getViewIdList).toEqual([])
    })
  })

  describe('setEnable', () => {
    it('should set enable to true', () => {
      const store = viewSelectorStore()
      store.setEnable(true)
      expect(store.getEnable).toBe(true)
    })

    it('should toggle enable', () => {
      const store = viewSelectorStore()
      store.setEnable(true)
      expect(store.getEnable).toBe(true)
      store.setEnable(false)
      expect(store.getEnable).toBe(false)
    })
  })

  describe('add', () => {
    it('should add an id to viewIdList', () => {
      const store = viewSelectorStore()
      store.add('view-1')
      expect(store.getViewIdList).toEqual(['view-1'])
    })

    it('should add multiple ids', () => {
      const store = viewSelectorStore()
      store.add('view-1')
      store.add('view-2')
      expect(store.getViewIdList).toEqual(['view-1', 'view-2'])
    })

    it('should prevent duplicate ids', () => {
      const store = viewSelectorStore()
      store.add('view-1')
      store.add('view-1')
      expect(store.getViewIdList).toEqual(['view-1'])
    })
  })

  describe('remove', () => {
    it('should remove all when called without id', () => {
      const store = viewSelectorStore()
      store.add('view-1')
      store.add('view-2')
      store.remove()
      expect(store.getViewIdList).toEqual([])
    })

    it('should remove specific id', () => {
      const store = viewSelectorStore()
      store.add('view-1')
      store.add('view-2')
      store.remove('view-1')
      expect(store.getViewIdList).toEqual(['view-2'])
    })

    it('should remove all occurrences of specific id', () => {
      const store = viewSelectorStore()
      store.add('view-1')
      store.viewIdList.push('view-1')
      store.add('view-2')
      store.remove('view-1')
      expect(store.getViewIdList).toEqual(['view-2'])
    })

    it('should do nothing when list is empty and remove called without id', () => {
      const store = viewSelectorStore()
      store.remove()
      expect(store.getViewIdList).toEqual([])
    })

    it('should do nothing when list is empty and remove called with id', () => {
      const store = viewSelectorStore()
      store.remove('nonexistent')
      expect(store.getViewIdList).toEqual([])
    })
  })

  describe('clear', () => {
    it('should reset enable to false and clear viewIdList', () => {
      const store = viewSelectorStore()
      store.setEnable(true)
      store.add('view-1')
      store.add('view-2')
      store.clear()
      expect(store.getEnable).toBe(false)
      expect(store.getViewIdList).toEqual([])
    })
  })
})
