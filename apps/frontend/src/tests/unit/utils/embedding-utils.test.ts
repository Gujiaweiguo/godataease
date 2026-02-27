import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { encodeOuterParams, decodeOuterParams, validateOuterParams } from '@/utils/embeddedTokenUtils'
import {
  isTokenExpiringSoon,
  extractTokenExpiryTime,
  needsTokenRefresh
} from '@/utils/embeddedTokenUtils'
import { isAllowedEmbeddedMessageOrigin } from '@/utils/embedded'

describe('Embedding Parameter Initialization and Callback Messaging', () => {
  beforeEach(() => {
    vi.spyOn(console, 'error').mockImplementation(() => undefined)
    vi.spyOn(console, 'warn').mockImplementation(() => undefined)
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  describe('outerParams utilities', () => {
    describe('encodeOuterParams', () => {
      it('should encode params to JSON string', () => {
        const params = {
          resourceId: '123',
          busiFlag: 'dark'
        }

        const encoded = encodeOuterParams(params, { format: 'json' })

        expect(encoded).toBe(JSON.stringify(params))
      })

      it('should encode params to Base64', () => {
        const params = {
          resourceId: '123',
          busiFlag: 'dark'
        }

        const encoded = encodeOuterParams(params, { format: 'base64' })

        const decoded = JSON.parse(atob(encoded))
        expect(decoded).toEqual(params)
      })

      it('should use JSON format by default', () => {
        const params = { resourceId: 'value' }

        const encoded = encodeOuterParams(params)

        expect(encoded).toBe(JSON.stringify(params))
      })
    })

    describe('decodeOuterParams', () => {
      it('should decode JSON params', () => {
        const params = { resourceId: '123' }
        const encoded = JSON.stringify(params)

        const decoded = decodeOuterParams(encoded, { format: 'json' })

        expect(decoded.params).toEqual(params)
        expect(decoded.isValid).toBe(true)
      })

      it('should decode Base64 params', () => {
        const params = { resourceId: '123' }
        const encoded = btoa(JSON.stringify(params))

        const decoded = decodeOuterParams(encoded, { format: 'base64' })

        expect(decoded.params).toEqual(params)
        expect(decoded.isValid).toBe(true)
      })

      it('should handle invalid JSON gracefully', () => {
        const decoded = decodeOuterParams('invalid-json')

        expect(decoded.isValid).toBe(false)
        expect(decoded.error).toBeDefined()
        expect(decoded.params).toEqual({})
      })

      it('should use JSON format by default', () => {
        const params = { resourceId: 'value' }
        const encoded = JSON.stringify(params)

        const decoded = decodeOuterParams(encoded)

        expect(decoded.params).toEqual(params)
      })
    })

    describe('validateOuterParams', () => {
      it('should validate correct params object', () => {
        const params = { resourceId: '123' }

        const result = validateOuterParams(params)

        expect(result.isValid).toBe(true)
        expect(result.error).toBeUndefined()
      })

      it('should reject non-object params', () => {
        const result = validateOuterParams(null)

        expect(result.isValid).toBe(false)
        expect(result.error).toBe('Params must be an object')
      })

      it('should reject string params', () => {
        const result = validateOuterParams('string' as unknown)

        expect(result.isValid).toBe(false)
        expect(result.error).toBe('Params must be an object')
      })

      it('should validate required fields', () => {
        const params = { resourceId: '123' }
        const result = validateOuterParams(params, ['resourceId'])

        expect(result.isValid).toBe(true)
      })

      it('should detect missing required fields', () => {
        const params = { theme: 'dark' }
        const result = validateOuterParams(params, ['resourceId'])

        expect(result.isValid).toBe(false)
        expect(result.error).toContain('resourceId')
      })
    })
  })

  describe('tokenExpiry utilities', () => {
    describe('isTokenExpiringSoon', () => {
      it('should return false if no expiry time', () => {
        const result = isTokenExpiringSoon(undefined, 5)

        expect(result).toBe(false)
      })

      it('should return false if token is already expired', () => {
        const expiredTime = Date.now() - 1000

        const result = isTokenExpiringSoon(expiredTime, 5)

        expect(result).toBe(false)
      })

      it('should return true if token expires within warning threshold', () => {
        const warningTime = Date.now() + 4 * 60 * 1000

        const result = isTokenExpiringSoon(warningTime, 5)

        expect(result).toBe(true)
      })

      it('should return false if token expires after warning threshold', () => {
        const futureTime = Date.now() + 6 * 60 * 1000

        const result = isTokenExpiringSoon(futureTime, 5)

        expect(result).toBe(false)
      })

      it('should use 5 minute default warning threshold', () => {
        const warningTime = Date.now() + 4 * 60 * 1000

        const result = isTokenExpiringSoon(warningTime)

        expect(result).toBe(true)
      })

      it('should support custom warning threshold', () => {
        const warningTime = Date.now() + 14 * 60 * 1000

        const result = isTokenExpiringSoon(warningTime, 15)

        expect(result).toBe(true)
      })
    })

    describe('extractTokenExpiryTime', () => {
      it('should extract expiry time from valid JWT', () => {
        const expTime = Math.floor((Date.now() + 3600000) / 1000)
        const header = Buffer.from('{"alg":"HS256"}').toString('base64')
        const payload = Buffer.from(`{"exp":${expTime}}`).toString('base64')
        const signature = 'signature'
        const jwtToken = `${header}.${payload}.${signature}`

        const result = extractTokenExpiryTime(jwtToken)

        expect(result).toBe(expTime * 1000)
      })

      it('should return undefined for invalid JWT format', () => {
        const result = extractTokenExpiryTime('invalid-token')

        expect(result).toBeUndefined()
      })

      it('should return undefined for token without expiry', () => {
        const header = Buffer.from('{"alg":"HS256"}').toString('base64')
        const payload = Buffer.from('{"sub":"user"}').toString('base64')
        const signature = 'signature'
        const jwtToken = `${header}.${payload}.${signature}`

        const result = extractTokenExpiryTime(jwtToken)

        expect(result).toBeUndefined()
      })

      it('should handle malformed JWT gracefully', () => {
        const result = extractTokenExpiryTime('invalid..token')

        expect(result).toBeUndefined()
      })
    })

    describe('needsTokenRefresh', () => {
      it('should return true if no expiry time', () => {
        const result = needsTokenRefresh(undefined)

        expect(result).toBe(true)
      })

      it('should return true if token is expired', () => {
        const expiredTime = Date.now() - 1000

        const result = needsTokenRefresh(expiredTime)

        expect(result).toBe(true)
      })

      it('should return false if token is still valid', () => {
        const futureTime = Date.now() + 120 * 60 * 1000

        const result = needsTokenRefresh(futureTime)

        expect(result).toBe(false)
      })

      it('should return true if within refresh threshold', () => {
        const nearExpiry = Date.now() + 30 * 60 * 1000

        const result = needsTokenRefresh(nearExpiry, 60)

        expect(result).toBe(true)
      })

      it('should use 60 minute default refresh threshold', () => {
        const nearExpiry = Date.now() + 30 * 60 * 1000

        const result = needsTokenRefresh(nearExpiry)

        expect(result).toBe(true)
      })

      it('should support custom refresh threshold', () => {
        const nearExpiry = Date.now() + 45 * 60 * 1000

        const result = needsTokenRefresh(nearExpiry, 90)

        expect(result).toBe(true)
      })
    })
  })

  describe('originValidation utilities', () => {
    it('allows exact origin in allowlist when enforced', () => {
      const result = isAllowedEmbeddedMessageOrigin(
        'https://example.com',
        ['https://example.com'],
        true
      )
      expect(result).toBe(true)
    })

    it('rejects non-matching origin when allowlist enforced', () => {
      const result = isAllowedEmbeddedMessageOrigin('https://other.com', ['https://example.com'], true)
      expect(result).toBe(false)
    })

    it('allows hostname-style allowlist entry', () => {
      const result = isAllowedEmbeddedMessageOrigin('https://example.com', ['example.com'], true)
      expect(result).toBe(true)
    })

    it('rejects empty origin', () => {
      const result = isAllowedEmbeddedMessageOrigin('', ['https://example.com'], true)
      expect(result).toBe(false)
    })
  })
})
