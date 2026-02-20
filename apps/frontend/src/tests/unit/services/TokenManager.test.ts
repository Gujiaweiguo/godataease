import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { TokenManager } from '@/services/TokenManager'
import { useEmbedded } from '@/store/modules/embedded'

vi.mock('@/store/modules/embedded')

vi.mock('@/api/embedded', () => ({
  embeddedInitIframeApi: vi.fn(),
  embeddedGetTokenArgsApi: vi.fn()
}))

describe('TokenManager', () => {
  let tokenManager: TokenManager
  let mockEmbeddedStore: ReturnType<typeof useEmbedded>

  beforeEach(() => {
    mockEmbeddedStore = {
      setToken: vi.fn(),
      setAllowedOrigins: vi.fn(),
      setTokenInfo: vi.fn(),
      getToken: vi.fn(() => 'mock-token'),
      getTokenInfo: vi.fn(() => new Map()),
      getAllowedOrigins: vi.fn(() => []),
      setEmbedReady: vi.fn(),
      getEmbedReady: vi.fn(() => false)
    }

    tokenManager = TokenManager.getInstance(mockEmbeddedStore)
  })

  afterEach(() => {
    tokenManager.cleanup()
    vi.clearAllMocks()
  })

  describe('getInstance', () => {
    it('should return singleton instance', () => {
      const instance1 = TokenManager.getInstance(mockEmbeddedStore)
      const instance2 = TokenManager.getInstance(mockEmbeddedStore)

      expect(instance1).toBe(instance2)
    })
  })

  describe('initializeToken', () => {
    const mockToken = 'test-token-123'
    const mockOrigin = 'https://example.com'

    it('should initialize token and set in store', async () => {
      const result = await tokenManager.initializeToken(mockToken, mockOrigin)

      expect(mockEmbeddedStore.setToken).toHaveBeenCalledWith(mockToken)
      expect(result.isValid).toBe(true)
    })

    it('should validate token on initialization', async () => {
      await tokenManager.initializeToken(mockToken, mockOrigin)

      const tokenInfo = tokenManager.getCurrentTokenInfo()
      expect(tokenInfo).toBeDefined()
      expect(tokenInfo?.token).toBe(mockToken)
    })

    it('should setup auto-refresh when enabled', async () => {
      const spy = vi.spyOn(tokenManager as any, 'setupAutoRefresh')

      await tokenManager.initializeToken(mockToken, mockOrigin, { refreshEnabled: true })

      expect(spy).toHaveBeenCalled()
    })

    it('should not setup auto-refresh when disabled', async () => {
      const spy = vi.spyOn(tokenManager as any, 'setupAutoRefresh')

      await tokenManager.initializeToken(mockToken, mockOrigin, { refreshEnabled: false })

      expect(spy).not.toHaveBeenCalled()
    })

    it('should set token type', async () => {
      await tokenManager.initializeToken(mockToken, mockOrigin, { tokenType: 'div' })

      const tokenInfo = tokenManager.getCurrentTokenInfo()
      expect(tokenInfo?.type).toBe('div')
    })

    it('should set resource id', async () => {
      const resourceId = 'resource-123'
      await tokenManager.initializeToken(mockToken, mockOrigin, { resourceId })

      const tokenInfo = tokenManager.getCurrentTokenInfo()
      expect(tokenInfo?.resourceId).toBe(resourceId)
    })

    it('should handle initialization errors gracefully', async () => {
      mockEmbeddedStore.setToken = vi.fn(() => {
        throw new Error('Init failed')
      })

      const result = await tokenManager.initializeToken(mockToken, mockOrigin)

      expect(result.isValid).toBe(false)
      expect(result.error).toBeDefined()
    })
  })

  describe('validateToken', () => {
    it('should validate non-empty token', async () => {
      const result = await tokenManager.validateToken('valid-token', 'https://example.com')

      expect(result.isValid).toBe(true)
    })

    it('should reject empty token', async () => {
      const result = await tokenManager.validateToken('', 'https://example.com')

      expect(result.isValid).toBe(false)
      expect(result.error).toBe('Token is empty')
    })

    it('should reject null token', async () => {
      const result = await tokenManager.validateToken(null as any, 'https://example.com')

      expect(result.isValid).toBe(false)
      expect(result.error).toBe('Token is empty')
    })

    it('should extract expiry time from JWT token', async () => {
      const expTime = Date.now() + 3600000
      const header = Buffer.from('{"alg":"HS256"}').toString('base64')
      const payload = Buffer.from(`{"exp":${Math.floor(expTime / 1000)}}`).toString('base64')
      const signature = 'signature'
      const jwtToken = `${header}.${payload}.${signature}`

      const result = await tokenManager.validateToken(jwtToken, 'https://example.com')

      expect(result.isValid).toBe(true)
      expect(result.expiryTime).toBe(expTime)
    })

    it('should reject expired token', async () => {
      const expiredTime = Date.now() - 3600000
      const header = Buffer.from('{"alg":"HS256"}').toString('base64')
      const payload = Buffer.from(`{"exp":${Math.floor(expiredTime / 1000)}}`).toString('base64')
      const signature = 'signature'
      const jwtToken = `${header}.${payload}.${signature}`

      const result = await tokenManager.validateToken(jwtToken, 'https://example.com')

      expect(result.isValid).toBe(false)
      expect(result.error).toBe('Token has expired')
    })

    it('should handle non-JWT tokens gracefully', async () => {
      const result = await tokenManager.validateToken('simple-token', 'https://example.com')

      expect(result.isValid).toBe(true)
      expect(result.expiryTime).toBeUndefined()
    })

    it('should handle malformed tokens gracefully', async () => {
      const result = await tokenManager.validateToken('invalid..token', 'https://example.com')

      expect(result.isValid).toBe(true)
      expect(result.expiryTime).toBeUndefined()
    })
  })

  describe('refreshToken', () => {
    const mockOrigin = 'https://example.com'

    beforeEach(() => {
      vi.mocked(embeddedGetTokenArgsApi).mockResolvedValue({
        data: {
          token: 'new-refreshed-token',
          allowedOrigins: ['https://new-origin.com']
        }
      })
    })

    it('should refresh token successfully', async () => {
      const success = await tokenManager.refreshToken(mockOrigin)

      expect(success).toBe(true)
      expect(mockEmbeddedStore.setToken).toHaveBeenCalledWith('new-refreshed-token')
      expect(mockEmbeddedStore.setAllowedOrigins).toHaveBeenCalledWith(['https://new-origin.com'])
    })

    it('should update token info after refresh', async () => {
      await tokenManager.refreshToken(mockOrigin)

      const tokenInfo = tokenManager.getCurrentTokenInfo()
      expect(tokenInfo?.token).toBe('new-refreshed-token')
    })

    it('should handle refresh failure gracefully', async () => {
      vi.mocked(embeddedGetTokenArgsApi).mockRejectedValue(new Error('Refresh failed'))

      const success = await tokenManager.refreshToken(mockOrigin)

      expect(success).toBe(false)
    })

    it('should handle missing token in response', async () => {
      vi.mocked(embeddedGetTokenArgsApi).mockResolvedValue({
        data: {}
      })

      const success = await tokenManager.refreshToken(mockOrigin)

      expect(success).toBe(false)
    })
  })

  describe('invalidateToken', () => {
    it('should clear token from store', () => {
      tokenManager.invalidateToken()

      expect(mockEmbeddedStore.setToken).toHaveBeenCalledWith('')
    })

    it('should clear allowed origins from store', () => {
      tokenManager.invalidateToken()

      expect(mockEmbeddedStore.setAllowedOrigins).toHaveBeenCalledWith([])
    })

    it('should clear token info from store', () => {
      tokenManager.invalidateToken()

      expect(mockEmbeddedStore.setTokenInfo).toHaveBeenCalledWith(new Map())
    })

    it('should stop auto-refresh', () => {
      const spy = vi.spyOn(tokenManager, 'stopAutoRefresh')

      tokenManager.invalidateToken()

      expect(spy).toHaveBeenCalled()
    })
  })

  describe('getCurrentTokenInfo', () => {
    it('should return current token info', () => {
      mockEmbeddedStore.getTokenInfo = vi.fn(() => new Map([['current', { token: 'test-token' }]]))

      const tokenInfo = tokenManager.getCurrentTokenInfo()

      expect(tokenInfo).toEqual({ token: 'test-token' })
    })

    it('should return undefined if no token info exists', () => {
      mockEmbeddedStore.getTokenInfo = vi.fn(() => new Map())

      const tokenInfo = tokenManager.getCurrentTokenInfo()

      expect(tokenInfo).toBeUndefined()
    })
  })

  describe('needsRefresh', () => {
    it('should return true if no token info', () => {
      mockEmbeddedStore.getTokenInfo = vi.fn(() => new Map())

      const needsRefresh = tokenManager.needsRefresh('https://example.com')

      expect(needsRefresh).toBe(true)
    })

    it('should return true if token is expired', () => {
      const expiredTime = Date.now() - 1000
      mockEmbeddedStore.getTokenInfo = vi.fn(
        () => new Map([['current', { token: 'test-token', expiryTime: expiredTime }]])
      )

      const needsRefresh = tokenManager.needsRefresh('https://example.com')

      expect(needsRefresh).toBe(true)
    })

    it('should return false if token is valid and not expired', () => {
      const futureTime = Date.now() + 3600000
      mockEmbeddedStore.getTokenInfo = vi.fn(
        () => new Map([['current', { token: 'test-token', expiryTime: futureTime }]])
      )

      const needsRefresh = tokenManager.needsRefresh('https://example.com')

      expect(needsRefresh).toBe(false)
    })

    it('should return true if expiry time is undefined', () => {
      mockEmbeddedStore.getTokenInfo = vi.fn(() => new Map([['current', { token: 'test-token' }]]))

      const needsRefresh = tokenManager.needsRefresh('https://example.com')

      expect(needsRefresh).toBe(true)
    })
  })

  describe('cleanup', () => {
    it('should stop auto-refresh', () => {
      const spy = vi.spyOn(tokenManager, 'stopAutoRefresh')

      tokenManager.cleanup()

      expect(spy).toHaveBeenCalled()
    })
  })
})
