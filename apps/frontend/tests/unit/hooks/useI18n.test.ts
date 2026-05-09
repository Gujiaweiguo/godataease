import { describe, it, expect, vi, beforeEach } from 'vitest'

const mockT = vi.fn((key: string) => `translated:${key}`)
const mockI18n = {
  global: {
    t: mockT,
    locale: 'zh-CN',
    getLocaleMessage: vi.fn(),
    setLocaleMessage: vi.fn()
  },
  mode: 'composition'
}

vi.mock('@/plugins/vue-i18n', () => ({
  get i18n() {
    return mockI18n
  }
}))

import { useI18n, t } from '@/hooks/web/useI18n'

describe('useI18n', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should return a t function', () => {
    const { t } = useI18n()
    expect(typeof t).toBe('function')
  })

  it('should call i18n.global.t when key contains dot', () => {
    const { t } = useI18n()
    t('common.ok')
    expect(mockT).toHaveBeenCalledWith('common.ok')
  })

  it('should prepend namespace to key when namespace is provided', () => {
    const { t } = useI18n('login')
    t('title')
    expect(mockT).toHaveBeenCalledWith('login.title')
  })

  it('should not prepend namespace when key already starts with namespace', () => {
    const { t } = useI18n('login')
    t('login.title')
    expect(mockT).toHaveBeenCalledWith('login.title')
  })

  it('should return key as-is when key has no dot and no namespace', () => {
    const { t } = useI18n()
    const result = t('simple')
    expect(result).toBe('simple')
  })

  it('should return empty string for empty key', () => {
    const { t } = useI18n()
    const result = t('')
    expect(result).toBe('')
  })

  it('should pass additional arguments to i18n.global.t', () => {
    const { t } = useI18n()
    t('common.greeting', ['arg1', 'arg2'])
    expect(mockT).toHaveBeenCalledWith('common.greeting', ['arg1', 'arg2'])
  })

  it('should pass named record arguments to i18n.global.t', () => {
    const { t } = useI18n()
    t('common.hello', { name: 'world' })
    expect(mockT).toHaveBeenCalledWith('common.hello', { name: 'world' })
  })

  it('should handle namespace with dotted key', () => {
    const { t } = useI18n('ns')
    t('deep.key')
    expect(mockT).toHaveBeenCalledWith('ns.deep.key')
  })
})

describe('t (standalone)', () => {
  it('should return the key as-is', () => {
    expect(t('any.key')).toBe('any.key')
  })

  it('should return empty string for empty key', () => {
    expect(t('')).toBe('')
  })

  it('should return arbitrary string unchanged', () => {
    expect(t('hello.world')).toBe('hello.world')
  })
})
