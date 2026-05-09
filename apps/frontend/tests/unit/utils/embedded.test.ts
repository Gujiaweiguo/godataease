import { describe, it, expect, vi, afterEach } from 'vitest'
import { resolveEmbeddedOrigin, isAllowedEmbeddedMessageOrigin } from '@/utils/embedded'

describe('embedded', () => {
  afterEach(() => {
    vi.restoreAllMocks()
  })

  describe('resolveEmbeddedOrigin', () => {
    it('should return window.location.origin when not in iframe', () => {
      const result = resolveEmbeddedOrigin()
      expect(result).toBe(window.location.origin)
    })

    it('should return referrer origin when in iframe with referrer', () => {
      const origSelf = window.self
      const origTop = window.top
      Object.defineProperty(window, 'self', { value: {} as any, writable: true, configurable: true })
      Object.defineProperty(window, 'top', { value: {} as any, writable: true, configurable: true })
      vi.spyOn(document, 'referrer', 'get').mockReturnValue('https://example.com/page/path')
      const result = resolveEmbeddedOrigin()
      expect(result).toBe('https://example.com')
      Object.defineProperty(window, 'self', { value: origSelf, writable: true, configurable: true })
      Object.defineProperty(window, 'top', { value: origTop, writable: true, configurable: true })
    })

    it('should return empty string for invalid referrer URL in iframe', () => {
      const origSelf = window.self
      const origTop = window.top
      Object.defineProperty(window, 'self', { value: {} as any, writable: true, configurable: true })
      Object.defineProperty(window, 'top', { value: {} as any, writable: true, configurable: true })
      vi.spyOn(document, 'referrer', 'get').mockReturnValue('not-a-valid-url')
      const result = resolveEmbeddedOrigin()
      expect(result).toBe('')
      Object.defineProperty(window, 'self', { value: origSelf, writable: true, configurable: true })
      Object.defineProperty(window, 'top', { value: origTop, writable: true, configurable: true })
    })

    it('should return window.location.origin when in iframe but no referrer', () => {
      const origSelf = window.self
      const origTop = window.top
      Object.defineProperty(window, 'self', { value: {} as any, writable: true, configurable: true })
      Object.defineProperty(window, 'top', { value: {} as any, writable: true, configurable: true })
      vi.spyOn(document, 'referrer', 'get').mockReturnValue('')
      const result = resolveEmbeddedOrigin()
      expect(result).toBe(window.location.origin)
      Object.defineProperty(window, 'self', { value: origSelf, writable: true, configurable: true })
      Object.defineProperty(window, 'top', { value: origTop, writable: true, configurable: true })
    })
  })

  describe('isAllowedEmbeddedMessageOrigin', () => {
    it('should return false for empty origin', () => {
      expect(isAllowedEmbeddedMessageOrigin('')).toBe(false)
    })

    it('should return false for whitespace-only origin', () => {
      expect(isAllowedEmbeddedMessageOrigin('   ')).toBe(false)
    })

    it('should return true when not enforcing allowlist and resolveEmbeddedOrigin returns empty', () => {
      const origSelf = window.self
      const origTop = window.top
      Object.defineProperty(window, 'self', { value: {} as any, writable: true, configurable: true })
      Object.defineProperty(window, 'top', { value: {} as any, writable: true, configurable: true })
      vi.spyOn(document, 'referrer', 'get').mockReturnValue('not-valid')
      expect(isAllowedEmbeddedMessageOrigin('https://example.com')).toBe(true)
      Object.defineProperty(window, 'self', { value: origSelf, writable: true, configurable: true })
      Object.defineProperty(window, 'top', { value: origTop, writable: true, configurable: true })
    })

    it('should return false when enforceAllowlist is true but allowlist is empty', () => {
      expect(isAllowedEmbeddedMessageOrigin('https://example.com', [], true)).toBe(false)
    })

    it('should match exact URL in allowlist when enforceAllowlist is true', () => {
      expect(isAllowedEmbeddedMessageOrigin('https://example.com', ['https://example.com'], true)).toBe(true)
    })

    it('should not match different origin in allowlist', () => {
      expect(isAllowedEmbeddedMessageOrigin('https://evil.com', ['https://example.com'], true)).toBe(false)
    })

    it('should match hostname-only allowlist entry', () => {
      expect(isAllowedEmbeddedMessageOrigin('https://example.com', ['example.com'], true)).toBe(true)
    })

    it('should handle origin with trailing slashes', () => {
      expect(isAllowedEmbeddedMessageOrigin('https://example.com///', ['https://example.com'], true)).toBe(true)
    })

    it('should skip empty allowlist entries', () => {
      expect(isAllowedEmbeddedMessageOrigin('https://example.com', ['', 'example.com'], true)).toBe(true)
    })

    it('should match across multiple allowlist entries', () => {
      expect(
        isAllowedEmbeddedMessageOrigin('https://other.com', ['example.com', 'other.com'], true)
      ).toBe(true)
    })

    it('should return false for non-matching hostname-only entry', () => {
      expect(isAllowedEmbeddedMessageOrigin('https://evil.com', ['example.com'], true)).toBe(false)
    })
  })
})
