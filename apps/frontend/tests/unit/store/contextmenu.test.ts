import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { contextmenuStore } from '@/store/modules/data-visualization/contextmenu'

describe('Contextmenu Store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  describe('Initial State', () => {
    it('should have correct initial menuTop', () => {
      const store = contextmenuStore()
      expect(store.menuTop).toBe(0)
    })

    it('should have correct initial menuLeft', () => {
      const store = contextmenuStore()
      expect(store.menuLeft).toBe(0)
    })

    it('should have correct initial menuShow', () => {
      const store = contextmenuStore()
      expect(store.menuShow).toBe(false)
    })

    it('should have correct initial position', () => {
      const store = contextmenuStore()
      expect(store.position).toBe('canvasCore')
    })
  })

  describe('showContextMenu', () => {
    it('should set menuShow to true', () => {
      const store = contextmenuStore()
      store.showContextMenu({ top: 100, left: 200, position: 'component' })
      expect(store.menuShow).toBe(true)
    })

    it('should set menuTop', () => {
      const store = contextmenuStore()
      store.showContextMenu({ top: 150, left: 250, position: 'component' })
      expect(store.menuTop).toBe(150)
    })

    it('should set menuLeft', () => {
      const store = contextmenuStore()
      store.showContextMenu({ top: 100, left: 300, position: 'component' })
      expect(store.menuLeft).toBe(300)
    })

    it('should set position', () => {
      const store = contextmenuStore()
      store.showContextMenu({ top: 100, left: 200, position: 'layer' })
      expect(store.position).toBe('layer')
    })

    it('should update all values correctly', () => {
      const store = contextmenuStore()
      store.showContextMenu({ top: 50, left: 100, position: 'custom' })
      
      expect(store.menuShow).toBe(true)
      expect(store.menuTop).toBe(50)
      expect(store.menuLeft).toBe(100)
      expect(store.position).toBe('custom')
    })
  })

  describe('hideContextMenu', () => {
    it('should set menuShow to false', () => {
      const store = contextmenuStore()
      store.showContextMenu({ top: 100, left: 200, position: 'component' })
      expect(store.menuShow).toBe(true)
      
      store.hideContextMenu()
      expect(store.menuShow).toBe(false)
    })

    it('should not affect other state values', () => {
      const store = contextmenuStore()
      store.showContextMenu({ top: 100, left: 200, position: 'component' })
      store.hideContextMenu()
      
      expect(store.menuTop).toBe(100)
      expect(store.menuLeft).toBe(200)
      expect(store.position).toBe('component')
    })
  })

  describe('Multiple operations', () => {
    it('should handle show and hide multiple times', () => {
      const store = contextmenuStore()
      
      store.showContextMenu({ top: 10, left: 20, position: 'pos1' })
      expect(store.menuShow).toBe(true)
      
      store.hideContextMenu()
      expect(store.menuShow).toBe(false)
      
      store.showContextMenu({ top: 30, left: 40, position: 'pos2' })
      expect(store.menuShow).toBe(true)
      expect(store.menuTop).toBe(30)
      expect(store.menuLeft).toBe(40)
      expect(store.position).toBe('pos2')
      
      store.hideContextMenu()
      expect(store.menuShow).toBe(false)
    })

    it('should handle zero values', () => {
      const store = contextmenuStore()
      store.showContextMenu({ top: 0, left: 0, position: '' })
      
      expect(store.menuTop).toBe(0)
      expect(store.menuLeft).toBe(0)
      expect(store.position).toBe('')
      expect(store.menuShow).toBe(true)
    })

    it('should handle negative values', () => {
      const store = contextmenuStore()
      store.showContextMenu({ top: -10, left: -20, position: 'negative' })
      
      expect(store.menuTop).toBe(-10)
      expect(store.menuLeft).toBe(-20)
      expect(store.position).toBe('negative')
    })

    it('should handle large values', () => {
      const store = contextmenuStore()
      store.showContextMenu({ top: 9999, left: 9999, position: 'large' })
      
      expect(store.menuTop).toBe(9999)
      expect(store.menuLeft).toBe(9999)
    })
  })
})
