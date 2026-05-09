import { describe, expect, it } from 'vitest'

import { isOriginAllowed, validateOrigin } from '@/utils/embeddedOriginValidation'

describe('embeddedOriginValidation', () => {
  describe('validateOrigin', () => {
    it('validates http origin', () => {
      expect(validateOrigin('http://localhost:8080')).toEqual({ isValid: true })
    })

    it('validates https origin', () => {
      expect(validateOrigin('https://example.com')).toEqual({ isValid: true })
    })

    it('rejects ftp protocol', () => {
      expect(validateOrigin('ftp://example.com')).toEqual({
        isValid: false,
        error: 'Invalid protocol'
      })
    })

    it('rejects invalid URL', () => {
      expect(validateOrigin('not-a-url')).toEqual({
        isValid: false,
        error: 'Invalid origin URL'
      })
    })

    it('rejects empty string', () => {
      expect(validateOrigin('')).toEqual({
        isValid: false,
        error: 'Invalid origin URL'
      })
    })

    it('validates http origin with port', () => {
      expect(validateOrigin('http://192.168.1.1:3000')).toEqual({ isValid: true })
    })
  })

  describe('isOriginAllowed', () => {
    it('returns false when no allowedOrigins provided', () => {
      expect(isOriginAllowed('https://example.com')).toBe(false)
    })

    it('returns false when allowedOrigins is empty array', () => {
      expect(isOriginAllowed('https://example.com', [])).toBe(false)
    })

    it('returns true when allowWhenTokenMissing is true', () => {
      expect(isOriginAllowed('https://evil.com', [], true)).toBe(true)
    })

    it('matches exact origin', () => {
      expect(isOriginAllowed('https://example.com', ['https://example.com'])).toBe(true)
    })

    it('does not match different origin', () => {
      expect(isOriginAllowed('https://other.com', ['https://example.com'])).toBe(false)
    })

    it('matches wildcard pattern', () => {
      expect(isOriginAllowed('https://sub.example.com', ['https://*.example.com'])).toBe(true)
    })

    it('does not match wildcard for different domain', () => {
      expect(isOriginAllowed('https://sub.other.com', ['https://*.example.com'])).toBe(false)
    })

    it('matches any subdomain with wildcard', () => {
      expect(isOriginAllowed('https://a.b.example.com', ['https://*.example.com'])).toBe(true)
    })

    it('matches against multiple allowed origins', () => {
      expect(
        isOriginAllowed('https://sub.example.com', [
          'https://other.com',
          'https://*.example.com'
        ])
      ).toBe(true)
    })
  })
})
