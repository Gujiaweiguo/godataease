import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useAppStore } from '@/store/modules/app'

// Mock dependencies
vi.mock('@/hooks/web/useCache', async () => {
  const { createUseCacheModuleMock } = await import('../helpers')
  return createUseCacheModuleMock()
})

vi.mock('@/api/login', async () => {
  const { createResolvedApiModuleMock } = await import('../helpers')
  return createResolvedApiModuleMock({ modelApi: { data: false } })
})

describe('App Store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  describe('Initial State', () => {
    it('should have correct initial values', () => {
    const store = useAppStore()
    expect(store.getSize).toBe(true)
    expect(store.getPageLoading).toBe(false)
    expect(store.getTitle).toBe('')
    expect(store.getDekey).toBe('DataEaseKey')
    expect(store.getIsDataEaseBi).toBe(false)
    expect(store.getIsIframe).toBe(false)
    expect(store.getDesktop).toBe(false)
    expect(store.getArrowSide).toBe(false)
  })
  })

  describe('setSize', () => {
    it('should set size to false', () => {
      const store = useAppStore()
      store.setSize(false)
      expect(store.getSize).toBe(false)
    })

    it('should toggle size', () => {
      const store = useAppStore()
      expect(store.getSize).toBe(true)
      store.setSize(false)
      expect(store.getSize).toBe(false)
      store.setSize(true)
      expect(store.getSize).toBe(true)
    })
  })

  describe('setPageLoading', () => {
    it('should set page loading to true', () => {
      const store = useAppStore()
      store.setPageLoading(true)
      expect(store.getPageLoading).toBe(true)
    })

    it('should set page loading to false', () => {
      const store = useAppStore()
      store.setPageLoading(true)
      store.setPageLoading(false)
      expect(store.getPageLoading).toBe(false)
    })
  })

  describe('setTitle', () => {
    it('should set title', () => {
      const store = useAppStore()
      store.setTitle('Test Title')
      expect(store.getTitle).toBe('Test Title')
    })

    it('should update document title', () => {
      const store = useAppStore()
      store.setTitle('New Title')
      expect(document.title).toBe('New Title')
    })
  })

  describe('setDekey', () => {
    it('should set dekey', () => {
      const store = useAppStore()
      store.setDekey('CustomKey')
      expect(store.getDekey).toBe('CustomKey')
    })
  })

  describe('setIsDataEaseBi', () => {
    it('should set isDataEaseBi to true', () => {
      const store = useAppStore()
      store.setIsDataEaseBi(true)
      expect(store.getIsDataEaseBi).toBe(true)
    })
  })

  describe('setIsIframe', () => {
    it('should set isIframe to true', () => {
      const store = useAppStore()
      store.setIsIframe(true)
      expect(store.getIsIframe).toBe(true)
    })
  })

  describe('setDesktop', () => {
    it('should set desktop to true', () => {
      const store = useAppStore()
      store.setDesktop(true)
      expect(store.getDesktop).toBe(true)
    })
  })

  describe('setArrowSide', () => {
    it('should set arrowSide to true', () => {
      const store = useAppStore()
      store.setArrowSide(true)
      expect(store.getArrowSide).toBe(true)
    })
  })

  describe('setAppModel', () => {
    it('should fetch and set app model', async () => {
      const { modelApi } = await import('@/api/login')
      vi.mocked(modelApi).mockResolvedValueOnce({ data: true })

      const store = useAppStore()
      await store.setAppModel()
      
      expect(store.getDesktop).toBe(true)
    })
  })
})
