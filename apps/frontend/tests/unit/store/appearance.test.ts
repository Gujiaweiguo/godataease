import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useAppearanceStore } from '@/store/modules/appearance'

// Mock Less.js modules to avoid circular dependency issues in CI
vi.mock('less/lib/less/functions/color.js', () => ({
  default: {
    mix: vi.fn((_color1, _color2, options) => ({
      toRGB: () => `rgb(${options.value}, ${options.value}, ${options.value})`
    }))
  }
}))

vi.mock('less/lib/less/tree/color.js', () => ({
  default: vi.fn((color) => ({ color }))
}))

// Mock dependencies
vi.mock('@/hooks/web/useCache', async () => {
  const { createUseCacheModuleMock } = await import('../helpers')
  return createUseCacheModuleMock()
})

vi.mock('@/api/font', async () => {
  const { createResolvedApiModuleMock } = await import('../helpers')
  return createResolvedApiModuleMock({
    defaultFont: {},
    list: []
  })
})

vi.mock('@/api/login', async () => {
  const { createResolvedApiModuleMock } = await import('../helpers')
  return createResolvedApiModuleMock({ uiLoadApi: {} })
})

vi.mock('@/store/modules/embedded', async () => {
  const { createEmbeddedStoreMock, createEmbeddedModuleMock } = await import('../helpers')
  const mockStore = createEmbeddedStoreMock({ baseUrl: '' })
  return createEmbeddedModuleMock(mockStore)
})

vi.mock('@/utils/utils', async () => {
  const { createUtilsModuleMock } = await import('../helpers')
  return createUtilsModuleMock()
})

describe('Appearance Store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  describe('Initial State', () => {
    it('should have correct initial values', () => {
      const store = useAppearanceStore()
      expect(store.getThemeColor).toBe('')
      expect(store.getCustomColor).toBe('')
      expect(store.getNavigateBg).toBe('')
      expect(store.getHelp).toBe('')
      expect(store.getShowSlogan).toBe('true')
      expect(store.getSlogan).toBe('')
      expect(store.getName).toBe('')
      expect(store.getFoot).toBe('false')
      expect(store.getFootContent).toBe('')
      expect(store.getLoaded).toBe(false)
      expect(store.getShowDemoTips).toBe(false)
      expect(store.getDemoTipsContent).toBe('')
      expect(store.getCommunity).toBe(true)
      expect(store.fontList).toEqual([])
    })

    it('should have correct initial show flags', () => {
      const store = useAppearanceStore()
      expect(store.showAi).toBe('0')
      expect(store.showDoc).toBe('0')
      expect(store.showCopilot).toBe('0')
      expect(store.showAbout).toBe('0')
    })
  })

  describe('setNavigate', () => {
    it('should set navigate', () => {
      const store = useAppearanceStore()
      store.setNavigate('navigate-image.png')
      expect(store.navigate).toBe('navigate-image.png')
    })
  })

  describe('setMobileLogin', () => {
    it('should set mobileLogin', () => {
      const store = useAppearanceStore()
      store.setMobileLogin('mobile-login.png')
      expect(store.mobileLogin).toBe('mobile-login.png')
    })
  })

  describe('setMobileLoginBg', () => {
    it('should set mobileLoginBg', () => {
      const store = useAppearanceStore()
      store.setMobileLoginBg('mobile-bg.png')
      expect(store.mobileLoginBg).toBe('mobile-bg.png')
    })
  })

  describe('setHelp', () => {
    it('should set help', () => {
      const store = useAppearanceStore()
      store.setHelp('https://help.example.com')
      expect(store.getHelp).toBe('https://help.example.com')
    })
  })

  describe('setNavigateBg', () => {
    it('should set navigateBg', () => {
      const store = useAppearanceStore()
      store.setNavigateBg('#ffffff')
      expect(store.getNavigateBg).toBe('#ffffff')
    })
  })

  describe('setThemeColor', () => {
    it('should set themeColor', () => {
      const store = useAppearanceStore()
      store.setThemeColor('#1890ff')
      expect(store.getThemeColor).toBe('#1890ff')
    })
  })

  describe('setCustomColor', () => {
    it('should set customColor', () => {
      const store = useAppearanceStore()
      store.setCustomColor('#ff0000')
      expect(store.getCustomColor).toBe('#ff0000')
    })
  })

  describe('setLoaded', () => {
    it('should set loaded to true', () => {
      const store = useAppearanceStore()
      store.setLoaded(true)
      expect(store.getLoaded).toBe(true)
    })

    it('should toggle loaded', () => {
      const store = useAppearanceStore()
      expect(store.getLoaded).toBe(false)
      store.setLoaded(true)
      expect(store.getLoaded).toBe(true)
      store.setLoaded(false)
      expect(store.getLoaded).toBe(false)
    })
  })

  describe('Getters with null return', () => {
    it('should return null for getNavigate when not set', () => {
      const store = useAppearanceStore()
      expect(store.getNavigate).toBe(null)
    })

    it('should return null for getMobileLogin when not set', () => {
      const store = useAppearanceStore()
      expect(store.getMobileLogin).toBe(null)
    })

    it('should return null for getMobileLoginBg when not set', () => {
      const store = useAppearanceStore()
      expect(store.getMobileLoginBg).toBe(null)
    })

    it('should return null for getBg when not set', () => {
      const store = useAppearanceStore()
      expect(store.getBg).toBe(null)
    })

    it('should return null for getLogin when not set', () => {
      const store = useAppearanceStore()
      expect(store.getLogin).toBe(null)
    })

    it('should return null for getWeb when not set', () => {
      const store = useAppearanceStore()
      expect(store.getWeb).toBe(null)
    })
  })

  describe('showAi/showDoc/showAbout getters', () => {
    it('should return false for showAi when value is 0', () => {
      const store = useAppearanceStore()
      store.showAi = '0'
      expect(store.getShowAi).toBe(false)
    })

    it('should return true for showAi when value is 1', () => {
      const store = useAppearanceStore()
      store.showAi = '1'
      expect(store.getShowAi).toBe(true)
    })

    it('should return false for showDoc when value is 0', () => {
      const store = useAppearanceStore()
      store.showDoc = '0'
      expect(store.getShowDoc).toBe(false)
    })

    it('should return true for showDoc when value is 1', () => {
      const store = useAppearanceStore()
      store.showDoc = '1'
      expect(store.getShowDoc).toBe(true)
    })

    it('should return false for showAbout when value is 0', () => {
      const store = useAppearanceStore()
      store.showAbout = '0'
      expect(store.getShowAbout).toBe(false)
    })

    it('should return true for showAbout when value is 1', () => {
      const store = useAppearanceStore()
      store.showAbout = '1'
      expect(store.getShowAbout).toBe(true)
    })
  })

  describe('Multiple setters', () => {
    it('should set multiple values correctly', () => {
      const store = useAppearanceStore()
      store.setThemeColor('#1890ff')
      store.setNavigateBg('#f5f5f5')
      store.setCustomColor('#ff0000')
      store.setHelp('https://help.example.com')

      expect(store.getThemeColor).toBe('#1890ff')
      expect(store.getNavigateBg).toBe('#f5f5f5')
      expect(store.getCustomColor).toBe('#ff0000')
      expect(store.getHelp).toBe('https://help.example.com')
    })
  })

  describe('Direct state modification', () => {
    it('should allow direct modification of slogan', () => {
      const store = useAppearanceStore()
      store.slogan = 'Welcome to DataEase'
      expect(store.getSlogan).toBe('Welcome to DataEase')
    })

    it('should allow direct modification of name', () => {
      const store = useAppearanceStore()
      store.name = 'DataEase BI'
      expect(store.getName).toBe('DataEase BI')
    })

    it('should allow direct modification of foot', () => {
      const store = useAppearanceStore()
      store.foot = 'true'
      expect(store.getFoot).toBe('true')
    })

    it('should allow direct modification of footContent', () => {
      const store = useAppearanceStore()
      store.footContent = 'Copyright 2024'
      expect(store.getFootContent).toBe('Copyright 2024')
    })

    it('should allow direct modification of showDemoTips', () => {
      const store = useAppearanceStore()
      store.showDemoTips = true
      expect(store.getShowDemoTips).toBe(true)
    })

    it('should allow direct modification of demoTipsContent', () => {
      const store = useAppearanceStore()
      store.demoTipsContent = 'This is a demo'
      expect(store.getDemoTipsContent).toBe('This is a demo')
    })

    it('should allow direct modification of community', () => {
      const store = useAppearanceStore()
      store.community = false
      expect(store.getCommunity).toBe(false)
    })
  })
})
