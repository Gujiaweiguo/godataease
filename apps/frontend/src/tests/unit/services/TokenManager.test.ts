import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { TokenManager } from '@/services/TokenManager'
import { useEmbedded } from '@/store/modules/embedded'
import { embeddedGetTokenArgsApi, embeddedInitIframeApi } from '@/api/embedded'

vi.mock('@/api/embedded', () => ({
  embeddedInitIframeApi: vi.fn(),
  embeddedGetTokenArgsApi: vi.fn()
}))

describe('TokenManager', () => {
  type TokenManagerWithSetup = {
    setupAutoRefresh: () => void
  }

  let tokenManager: TokenManager
  let mockEmbeddedStore: ReturnType<typeof useEmbedded>
  let setTokenSpy: ReturnType<typeof vi.spyOn>
  let setAllowedOriginsSpy: ReturnType<typeof vi.spyOn>
  let setTokenInfoSpy: ReturnType<typeof vi.spyOn>

  beforeEach(() => {
    vi.spyOn(console, 'error').mockImplementation(() => undefined)
    vi.spyOn(console, 'warn').mockImplementation(() => undefined)

    mockEmbeddedStore = useEmbedded()
    mockEmbeddedStore.clearState()

    setTokenSpy = vi.spyOn(mockEmbeddedStore, 'setToken')
    setAllowedOriginsSpy = vi.spyOn(mockEmbeddedStore, 'setAllowedOrigins')
    setTokenInfoSpy = vi.spyOn(mockEmbeddedStore, 'setTokenInfo')

    tokenManager = TokenManager.getInstance(mockEmbeddedStore)
  })

  afterEach(() => {
    tokenManager.cleanup()
    vi.restoreAllMocks()
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

      expect(setTokenSpy).toHaveBeenCalledWith(mockToken)
      expect(result.isValid).toBe(true)
    })

    it('should validate token on initialization', async () => {
      await tokenManager.initializeToken(mockToken, mockOrigin)

      const tokenInfo = tokenManager.getCurrentTokenInfo()
      expect(tokenInfo).toBeDefined()
      expect(tokenInfo?.token).toBe(mockToken)
    })

    it('should setup auto-refresh when enabled', async () => {
      const managerWithSetup = Object.getPrototypeOf(tokenManager) as TokenManagerWithSetup
      const spy = vi.spyOn(managerWithSetup, 'setupAutoRefresh')

      await tokenManager.initializeToken(mockToken, mockOrigin, { refreshEnabled: true })

      expect(spy).toHaveBeenCalled()
    })

    it('should not setup auto-refresh when disabled', async () => {
      const managerWithSetup = Object.getPrototypeOf(tokenManager) as TokenManagerWithSetup
      const spy = vi.spyOn(managerWithSetup, 'setupAutoRefresh')

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
      vi.mocked(embeddedInitIframeApi).mockRejectedValue(new Error('Init failed'))

      const result = await tokenManager.initializeToken(mockToken, mockOrigin)

      expect(result.isValid).toBe(false)
      expect(result.error).toBeDefined()
    })
  })

  describe('validateToken', () => {
    it('should validate non-empty token', async () => {
      const result = await tokenManager.validateToken('valid-token')

      expect(result.isValid).toBe(true)
    })

    it('should reject empty token', async () => {
      const result = await tokenManager.validateToken('')

      expect(result.isValid).toBe(false)
      expect(result.error).toBe('Token is empty')
    })

    it('should reject null token', async () => {
      const result = await tokenManager.validateToken(null as unknown as string)

      expect(result.isValid).toBe(false)
      expect(result.error).toBe('Token is empty')
    })

    it('should extract expiry time from JWT token', async () => {
      const expTime = Date.now() + 3600000
      const header = Buffer.from('{"alg":"HS256"}').toString('base64')
      const payload = Buffer.from(`{"exp":${Math.floor(expTime / 1000)}}`).toString('base64')
      const signature = 'signature'
      const jwtToken = `${header}.${payload}.${signature}`

      const result = await tokenManager.validateToken(jwtToken)

      expect(result.isValid).toBe(true)
      expect(result.expiryTime).toBe(Math.floor(expTime / 1000) * 1000)
    })

    it('should reject expired token', async () => {
      const expiredTime = Date.now() - 3600000
      const header = Buffer.from('{"alg":"HS256"}').toString('base64')
      const payload = Buffer.from(`{"exp":${Math.floor(expiredTime / 1000)}}`).toString('base64')
      const signature = 'signature'
      const jwtToken = `${header}.${payload}.${signature}`

      const result = await tokenManager.validateToken(jwtToken)

      expect(result.isValid).toBe(false)
      expect(result.error).toBe('Token has expired')
    })

    it('should handle non-JWT tokens gracefully', async () => {
      const result = await tokenManager.validateToken('simple-token')

      expect(result.isValid).toBe(true)
      expect(result.expiryTime).toBeUndefined()
    })

    it('should handle malformed tokens gracefully', async () => {
      const result = await tokenManager.validateToken('invalid..token')

      expect(result.isValid).toBe(true)
      expect(result.expiryTime).toBeUndefined()
    })
  })

  describe('refreshToken', () => {
    beforeEach(() => {
      vi.mocked(embeddedGetTokenArgsApi).mockResolvedValue({
        data: {
          token: 'new-refreshed-token',
          allowedOrigins: ['https://new-origin.com']
        }
      })
    })

    it('should refresh token successfully', async () => {
      const success = await tokenManager.refreshToken()

      expect(success).toBe(true)
      expect(setTokenSpy).toHaveBeenCalledWith('new-refreshed-token')
      expect(setAllowedOriginsSpy).toHaveBeenCalledWith(['https://new-origin.com'])
    })

    it('should update token info after refresh', async () => {
      await tokenManager.refreshToken()

      const tokenInfo = tokenManager.getCurrentTokenInfo()
      expect(tokenInfo?.token).toBe('new-refreshed-token')
    })

    it('should handle refresh failure gracefully', async () => {
      vi.mocked(embeddedGetTokenArgsApi).mockRejectedValue(new Error('Refresh failed'))

      const success = await tokenManager.refreshToken()

      expect(success).toBe(false)
    })

    it('should handle missing token in response', async () => {
      vi.mocked(embeddedGetTokenArgsApi).mockResolvedValue({
        data: {}
      })

      const success = await tokenManager.refreshToken()

      expect(success).toBe(false)
    })
  })

  describe('invalidateToken', () => {
    it('should clear token from store', () => {
      tokenManager.invalidateToken()

      expect(setTokenSpy).toHaveBeenCalledWith('')
    })

    it('should clear allowed origins from store', () => {
      tokenManager.invalidateToken()

      expect(setAllowedOriginsSpy).toHaveBeenCalledWith([])
    })

    it('should clear token info from store', () => {
      tokenManager.invalidateToken()

      expect(setTokenInfoSpy).toHaveBeenCalledWith(new Map())
    })

    it('should stop auto-refresh', () => {
      const spy = vi.spyOn(tokenManager, 'stopAutoRefresh')

      tokenManager.invalidateToken()

      expect(spy).toHaveBeenCalled()
    })
  })

  describe('getCurrentTokenInfo', () => {
    it('should return current token info', () => {
      mockEmbeddedStore.setTokenInfo(new Map([['current', { token: 'test-token' }]]))

      const tokenInfo = tokenManager.getCurrentTokenInfo()

      expect(tokenInfo).toEqual({ token: 'test-token' })
    })

    it('should return undefined if no token info exists', () => {
      mockEmbeddedStore.setTokenInfo(new Map())

      const tokenInfo = tokenManager.getCurrentTokenInfo()

      expect(tokenInfo).toBeUndefined()
    })
  })

  describe('needsRefresh', () => {
    it('should return true if no token info', () => {
      mockEmbeddedStore.setTokenInfo(new Map())

      const needsRefresh = tokenManager.needsRefresh()

      expect(needsRefresh).toBe(true)
    })

    it('should return true if token is expired', () => {
      const expiredTime = Date.now() - 1000
      mockEmbeddedStore.setTokenInfo(new Map([['current', { token: 'test-token', expiryTime: expiredTime }]]))

      const needsRefresh = tokenManager.needsRefresh()

      expect(needsRefresh).toBe(true)
    })

    it('should return false if token is valid and not expired', () => {
      const futureTime = Date.now() + 3600000
      mockEmbeddedStore.setTokenInfo(new Map([['current', { token: 'test-token', expiryTime: futureTime }]]))

      const needsRefresh = tokenManager.needsRefresh()

      expect(needsRefresh).toBe(false)
    })

    it('should return false if expiry time is undefined', () => {
      mockEmbeddedStore.setTokenInfo(new Map([['current', { token: 'test-token' }]]))

      const needsRefresh = tokenManager.needsRefresh()

      expect(needsRefresh).toBe(false)
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
