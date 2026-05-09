import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'

// Mock getLocale to return a fixed value since it reads from localStorage
vi.mock('@/utils/utils', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@/utils/utils')>()
  return {
    ...actual,
    getLocale: () => 'zh-CN'
  }
})

// Mock axios for getLocaleMap API call
vi.mock('@/config/axios', () => ({
  default: {
    get: vi.fn()
  }
}))

import { useLocaleStore } from '@/store/modules/locale'
import request from '@/config/axios'

describe('Locale Store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  describe('Initial State', () => {
    it('should have customLoaded as false', () => {
      const store = useLocaleStore()
      expect(store.customLoaded).toBe(false)
    })

    it('should have correct currentLocale with lang from getLocale', () => {
      const store = useLocaleStore()
      expect(store.currentLocale.lang).toBe('zh-CN')
    })

    it('should have elLocale set for currentLocale', () => {
      const store = useLocaleStore()
      expect(store.currentLocale.elLocale).toBeDefined()
    })

    it('should have 3 items in localeMap', () => {
      const store = useLocaleStore()
      expect(store.localeMap).toHaveLength(3)
    })

    it('should have correct localeMap entries', () => {
      const store = useLocaleStore()
      const langs = store.localeMap.map(item => item.lang)
      expect(langs).toContain('zh-CN')
      expect(langs).toContain('en')
      expect(langs).toContain('tw')
    })
  })

  describe('getCurrentLocale getter', () => {
    it('should return the currentLocale object', () => {
      const store = useLocaleStore()
      const locale = store.getCurrentLocale
      expect(locale.lang).toBe('zh-CN')
      expect(locale.elLocale).toBeDefined()
    })
  })

  describe('setCurrentLocale', () => {
    it('should change lang and elLocale', () => {
      const store = useLocaleStore()
      store.setCurrentLocale({ lang: 'en', elLocale: undefined as any })
      expect(store.currentLocale.lang).toBe('en')
      expect(store.currentLocale.elLocale).toBeDefined()
    })

    it('should resolve elLocale from lang', () => {
      const store = useLocaleStore()
      store.setCurrentLocale({ lang: 'tw', elLocale: undefined as any })
      expect(store.currentLocale.lang).toBe('tw')
      expect(store.currentLocale.elLocale).toBeDefined()
    })

    it('should fallback to zh-CN elLocale for unknown lang', () => {
      const store = useLocaleStore()
      store.setCurrentLocale({ lang: 'unknown' as any, elLocale: undefined as any })
      expect(store.currentLocale.lang).toBe('unknown')
      expect(store.currentLocale.elLocale).toBeDefined()
    })
  })

  describe('setLang', () => {
    it('should change both lang and elLocale', () => {
      const store = useLocaleStore()
      store.setLang('en')
      expect(store.currentLocale.lang).toBe('en')
      expect(store.currentLocale.elLocale).toBeDefined()
    })

    it('should handle zh-CN', () => {
      const store = useLocaleStore()
      store.setLang('en')
      store.setLang('zh-CN')
      expect(store.currentLocale.lang).toBe('zh-CN')
      expect(store.currentLocale.elLocale).toBeDefined()
    })

    it('should handle tw', () => {
      const store = useLocaleStore()
      store.setLang('tw')
      expect(store.currentLocale.lang).toBe('tw')
      expect(store.currentLocale.elLocale).toBeDefined()
    })
  })

  describe('getLocaleMap', () => {
    it('should return localeMap immediately when customLoaded is true', async () => {
      const store = useLocaleStore()
      store.customLoaded = true
      const result = await store.getLocaleMap
      expect(result).toHaveLength(3)
      expect(vi.mocked(request.get)).not.toHaveBeenCalled()
    })

    it('should fetch custom locales and append to localeMap', async () => {
      vi.mocked(request.get).mockResolvedValueOnce({
        data: { 'ja-JP': '日本語', 'ko-KR': '한국어' }
      })
      const store = useLocaleStore()
      const result = await store.getLocaleMap
      expect(result).toHaveLength(5)
      expect(store.customLoaded).toBe(true)
      expect(vi.mocked(request.get)).toHaveBeenCalledWith({ url: '/sysParameter/i18nOptions' })
    })

    it('should mark custom entries with custom: true', async () => {
      vi.mocked(request.get).mockResolvedValueOnce({
        data: { 'ja-JP': '日本語' }
      })
      const store = useLocaleStore()
      await store.getLocaleMap
      const customItem = store.localeMap.find(item => (item as any).lang === 'ja-JP')
      expect(customItem).toBeDefined()
      expect((customItem as any).custom).toBe(true)
    })

    it('should handle API error gracefully and return existing localeMap', async () => {
      vi.mocked(request.get).mockRejectedValueOnce(new Error('Network error'))
      const store = useLocaleStore()
      const result = await store.getLocaleMap
      expect(result).toHaveLength(3)
      expect(store.customLoaded).toBe(true)
    })

    it('should only fetch once (subsequent calls return cached)', async () => {
      vi.mocked(request.get).mockResolvedValueOnce({
        data: { 'ja-JP': '日本語' }
      })
      const store = useLocaleStore()
      await store.getLocaleMap
      await store.getLocaleMap
      expect(vi.mocked(request.get)).toHaveBeenCalledTimes(1)
    })
  })
})
