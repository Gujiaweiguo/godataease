import { describe, it, expect, vi, beforeEach } from 'vitest'

const mockSetLocaleMessage = vi.fn()
const mockI18nGlobal = {
  t: vi.fn((key: string) => key),
  setLocaleMessage: mockSetLocaleMessage,
  locale: { value: 'zh-CN' }
}

vi.mock('@/plugins/vue-i18n', () => ({
  get i18n() {
    return {
      mode: 'composition',
      global: mockI18nGlobal
    }
  }
}))

const mockSetCurrentLocale = vi.fn()
const mockGetLocaleMap = vi.fn()
vi.mock('@/store/modules/locale', () => ({
  useLocaleStoreWithOut: () => ({
    setCurrentLocale: mockSetCurrentLocale,
    get getLocaleMap() {
      return Promise.resolve(mockGetLocaleMap())
    }
  })
}))

vi.mock('@/plugins/vue-i18n/helper', () => ({
  setHtmlPageLang: vi.fn()
}))

vi.mock('@/config/axios/service', () => ({
  PATH_URL: '/api'
}))

vi.mock('@/hooks/web/useEmitt', () => ({
  useEmitt: () => ({
    emitter: { emit: vi.fn() }
  })
}))

import { useLocale } from '@/hooks/web/useLocale'

describe('useLocale', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should return a changeLocale function', () => {
    const { changeLocale } = useLocale()
    expect(typeof changeLocale).toBe('function')
  })

  it('should load zh-CN locale module and set locale message', async () => {
    const { changeLocale } = useLocale()
    await changeLocale('zh-CN')
    expect(mockSetLocaleMessage).toHaveBeenCalledWith('zh-CN', expect.anything())
  })

  it('should load en locale module and set locale message', async () => {
    const { changeLocale } = useLocale()
    await changeLocale('en')
    expect(mockSetLocaleMessage).toHaveBeenCalledWith('en', expect.anything())
  })

  it('should load tw locale module and set locale message', async () => {
    const { changeLocale } = useLocale()
    await changeLocale('tw')
    expect(mockSetLocaleMessage).toHaveBeenCalledWith('tw', expect.anything())
  })

  it('should call setHtmlPageLang with the locale', async () => {
    const { setHtmlPageLang } = await import('@/plugins/vue-i18n/helper')
    const { changeLocale } = useLocale()
    await changeLocale('en')
    expect(setHtmlPageLang).toHaveBeenCalledWith('en')
  })

  it('should call setCurrentLocale with correct lang', async () => {
    const { changeLocale } = useLocale()
    await changeLocale('zh-CN')
    expect(mockSetCurrentLocale).toHaveBeenCalledWith({ lang: 'zh-CN' })
  })

  it('should throw error for unknown locale without matching localeMap entry', async () => {
    mockGetLocaleMap.mockReturnValue([])
    const { changeLocale } = useLocale()
    await expect(changeLocale('xx-XX' as any)).rejects.toThrow('missing locale option: xx-XX')
  })

  it('should handle locale not in built-in list but present in localeMap', async () => {
    mockGetLocaleMap.mockReturnValue([
      { lang: 'ja', name: 'front', custom: true }
    ])
    const { changeLocale } = useLocale()
    await expect(changeLocale('ja' as any)).rejects.toThrow()
  })

  it('should update i18n.global.locale.value for composition mode', async () => {
    const { changeLocale } = useLocale()
    await changeLocale('en')
    expect(mockI18nGlobal.locale.value).toBe('en')
  })
})
