import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { Base64 } from 'js-base64'
import {
  encodeOuterParams,
  decodeOuterParams,
  validateOuterParams,
  isTokenExpiringSoon,
  extractTokenExpiryTime,
  needsTokenRefresh
} from '@/utils/embeddedTokenUtils'

describe('embeddedTokenUtils', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.useFakeTimers()
    vi.setSystemTime(new Date('2025-01-01T00:00:00Z'))
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  // ─── encodeOuterParams ────────────────────────────────────────────────
  describe('encodeOuterParams', () => {
    it('should encode params as JSON string by default', () => {
      const params = { resourceId: 'abc123', dvId: 'dv001' }
      const result = encodeOuterParams(params)
      expect(result).toBe(JSON.stringify(params))
    })

    it('should encode params as JSON when format is json', () => {
      const params = { chartId: 'chart1' }
      const result = encodeOuterParams(params, { format: 'json' })
      expect(result).toBe(JSON.stringify(params))
    })

    it('should encode params as Base64 when format is base64', () => {
      const params = { resourceId: 'res1', busiFlag: 'dashboard' }
      const result = encodeOuterParams(params, { format: 'base64' })
      const decoded = Base64.decode(result)
      expect(decoded).toBe(JSON.stringify(params))
    })

    it('should handle empty object', () => {
      const result = encodeOuterParams({})
      expect(result).toBe('{}')
    })

    it('should handle params with nested objects', () => {
      const params = { outerParams: { key1: 'val1' }, callbackParams: { cb: true } }
      const result = encodeOuterParams(params)
      expect(JSON.parse(result)).toEqual(params)
    })
  })

  // ─── decodeOuterParams ────────────────────────────────────────────────
  describe('decodeOuterParams', () => {
    it('should decode valid JSON string', () => {
      const encoded = JSON.stringify({ resourceId: 'abc', dvId: 'dv1' })
      const result = decodeOuterParams(encoded) as any
      expect(result.isValid).toBe(true)
      expect(result.params).toEqual({ resourceId: 'abc', dvId: 'dv1' })
    })

    it('should decode Base64 encoded string', () => {
      const params = { chartId: 'c1', busiFlag: 'view' }
      const encoded = Base64.encode(JSON.stringify(params))
      const result = decodeOuterParams(encoded, { format: 'base64' }) as any
      expect(result.isValid).toBe(true)
      expect(result.params).toEqual(params)
    })

    it('should return error for invalid JSON', () => {
      const result = decodeOuterParams('not-valid-json') as any
      expect(result.isValid).toBe(false)
      expect(result.error).toBe('Invalid outer params format')
    })

    it('should return error for invalid Base64 decoded content', () => {
      const badBase64 = Base64.encode('not-json')
      const result = decodeOuterParams(badBase64, { format: 'base64' }) as any
      expect(result.isValid).toBe(false)
      expect(result.error).toBe('Invalid outer params format')
    })

    it('should decode empty object JSON', () => {
      const result = decodeOuterParams('{}') as any
      expect(result.isValid).toBe(true)
      expect(result.params).toEqual({})
    })
  })

  // ─── validateOuterParams ──────────────────────────────────────────────
  describe('validateOuterParams', () => {
    it('should return valid for a non-null object with no required fields', () => {
      const result = validateOuterParams({ foo: 'bar' })
      expect(result.isValid).toBe(true)
      expect(result.error).toBeUndefined()
    })

    it('should return invalid for null', () => {
      const result = validateOuterParams(null)
      expect(result.isValid).toBe(false)
      expect(result.error).toBe('Params must be an object')
    })

    it('should return invalid for undefined', () => {
      const result = validateOuterParams(undefined)
      expect(result.isValid).toBe(false)
      expect(result.error).toBe('Params must be an object')
    })

    it('should return invalid for a string', () => {
      const result = validateOuterParams('string')
      expect(result.isValid).toBe(false)
      expect(result.error).toBe('Params must be an object')
    })

    it('should return invalid for a number', () => {
      const result = validateOuterParams(42)
      expect(result.isValid).toBe(false)
    })

    it('should return valid when all required fields are present', () => {
      const result = validateOuterParams({ resourceId: 'r1', dvId: 'd1' }, [
        'resourceId',
        'dvId'
      ])
      expect(result.isValid).toBe(true)
    })

    it('should return invalid when required fields are missing', () => {
      const result = validateOuterParams({ resourceId: 'r1' }, ['resourceId', 'dvId'])
      expect(result.isValid).toBe(false)
      expect(result.error).toBe('Missing required fields: dvId')
    })

    it('should list all missing fields', () => {
      const result = validateOuterParams({}, ['a', 'b', 'c'])
      expect(result.isValid).toBe(false)
      expect(result.error).toContain('a')
      expect(result.error).toContain('b')
      expect(result.error).toContain('c')
    })
  })

  // ─── isTokenExpiringSoon ─────────────────────────────────────────────
  describe('isTokenExpiringSoon', () => {
    it('should return false for undefined expiry', () => {
      expect(isTokenExpiringSoon(undefined)).toBe(false)
    })

    it('should return false for already expired token', () => {
      const expiredTime = Date.now() - 10000
      expect(isTokenExpiringSoon(expiredTime)).toBe(false)
    })

    it('should return true when expiring within default threshold (5 min)', () => {
      const expiringTime = Date.now() + 3 * 60 * 1000 // 3 minutes
      expect(isTokenExpiringSoon(expiringTime)).toBe(true)
    })

    it('should return false when far from expiry', () => {
      const futureTime = Date.now() + 60 * 60 * 1000 // 1 hour
      expect(isTokenExpiringSoon(futureTime)).toBe(false)
    })

    it('should respect custom warning threshold', () => {
      const futureTime = Date.now() + 10 * 60 * 1000 // 10 minutes
      expect(isTokenExpiringSoon(futureTime, 15)).toBe(true)
      expect(isTokenExpiringSoon(futureTime, 5)).toBe(false)
    })

    it('should return false when exactly at threshold boundary (timeUntilExpiry equals threshold)', () => {
      const thresholdMs = 5 * 60 * 1000
      const futureTime = Date.now() + thresholdMs
      // timeUntilExpiry == threshold => condition is > 0 && <= threshold
      // exactly equal means it's included (<=)
      expect(isTokenExpiringSoon(futureTime)).toBe(true)
    })
  })

  // ─── extractTokenExpiryTime ──────────────────────────────────────────
  describe('extractTokenExpiryTime', () => {
    function makeJwt(payload: Record<string, any>): string {
      const header = btoa(JSON.stringify({ alg: 'HS256', typ: 'JWT' }))
      const body = btoa(JSON.stringify(payload))
      return `${header}.${body}.fakesignature`
    }

    it('should extract exp claim and convert to milliseconds', () => {
      const exp = Math.floor(Date.now() / 1000) + 3600
      const token = makeJwt({ exp, sub: 'test' })
      const result = extractTokenExpiryTime(token)
      expect(result).toBe(exp * 1000)
    })

    it('should return undefined for token without exp claim', () => {
      const token = makeJwt({ sub: 'test' })
      const result = extractTokenExpiryTime(token)
      expect(result).toBeUndefined()
    })

    it('should return undefined for invalid token format', () => {
      expect(extractTokenExpiryTime('not.a.valid.jwt.token')).toBeUndefined()
    })

    it('should return undefined for empty string', () => {
      expect(extractTokenExpiryTime('')).toBeUndefined()
    })

    it('should return undefined for token with only 2 parts', () => {
      expect(extractTokenExpiryTime('header.payload')).toBeUndefined()
    })
  })

  // ─── needsTokenRefresh ───────────────────────────────────────────────
  describe('needsTokenRefresh', () => {
    it('should return true for undefined expiry', () => {
      expect(needsTokenRefresh(undefined)).toBe(true)
    })

    it('should return true for already expired token', () => {
      const expiredTime = Date.now() - 10000
      expect(needsTokenRefresh(expiredTime)).toBe(true)
    })

    it('should return true when within refresh threshold (default 60 min)', () => {
      const futureTime = Date.now() + 30 * 60 * 1000 // 30 min
      expect(needsTokenRefresh(futureTime)).toBe(true)
    })

    it('should return false when far from expiry beyond threshold', () => {
      const futureTime = Date.now() + 120 * 60 * 1000 // 2 hours
      expect(needsTokenRefresh(futureTime)).toBe(false)
    })

    it('should respect custom refresh threshold', () => {
      const futureTime = Date.now() + 90 * 60 * 1000 // 90 minutes
      expect(needsTokenRefresh(futureTime, 120)).toBe(true)
      expect(needsTokenRefresh(futureTime, 60)).toBe(false)
    })

    it('should return true when exactly at threshold boundary', () => {
      const thresholdMs = 60 * 60 * 1000
      const futureTime = Date.now() + thresholdMs
      // timeUntilExpiry == threshold => condition is <= threshold, which is true
      expect(needsTokenRefresh(futureTime)).toBe(true)
    })
  })
})
