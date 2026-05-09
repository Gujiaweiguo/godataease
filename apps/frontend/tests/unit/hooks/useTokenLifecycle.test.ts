import { describe, it, expect, vi, beforeEach } from 'vitest'

const mockTokenManager = {
  initializeToken: vi.fn(),
  refreshToken: vi.fn(),
  invalidateToken: vi.fn(),
  needsRefresh: vi.fn(),
  getCurrentTokenInfo: vi.fn(),
  cleanup: vi.fn()
}

let onMountedCallback: (() => void) | null = null
let onBeforeUnmountCallback: (() => void) | null = null

vi.mock('vue', async (importOriginal) => {
  const actual = await importOriginal()
  return {
    ...(actual as any),
    ref: (actual as any).ref,
    onMounted: vi.fn((cb: () => void) => {
      onMountedCallback = cb
    }),
    onBeforeUnmount: vi.fn((cb: () => void) => {
      onBeforeUnmountCallback = cb
    })
  }
})

vi.mock('@/services/TokenManager', () => ({
  useTokenManager: () => mockTokenManager
}))

import { useTokenLifecycle } from '@/hooks/embedded/useTokenLifecycle'

describe('useTokenLifecycle', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    onMountedCallback = null
    onBeforeUnmountCallback = null
  })

  it('should return lifecycle methods and reactive refs', () => {
    const lifecycle = useTokenLifecycle()
    expect(typeof lifecycle.initialize).toBe('function')
    expect(typeof lifecycle.refresh).toBe('function')
    expect(typeof lifecycle.invalidate).toBe('function')
    expect(typeof lifecycle.needsRefresh).toBe('function')
    expect(typeof lifecycle.getCurrentTokenInfo).toBe('function')
    expect(typeof lifecycle.getValidationStatus).toBe('function')
    expect(lifecycle.isInitialized).toBeDefined()
    expect(lifecycle.tokenValidationResult).toBeDefined()
    expect(lifecycle.lastRefreshTime).toBeDefined()
  })

  it('should initialize token and set isInitialized on valid result', async () => {
    mockTokenManager.initializeToken.mockResolvedValue({
      isValid: true,
      expiryTime: Date.now() + 60000
    })

    const lifecycle = useTokenLifecycle()
    const result = await lifecycle.initialize('test-token', {
      refreshEnabled: true,
      tokenType: 'iframe'
    })

    expect(result.isValid).toBe(true)
    expect(lifecycle.isInitialized.value).toBe(true)
    expect(mockTokenManager.initializeToken).toHaveBeenCalledWith(
      'test-token',
      expect.any(String),
      { refreshEnabled: true, tokenType: 'iframe' }
    )
  })

  it('should set isInitialized to false on invalid token', async () => {
    mockTokenManager.initializeToken.mockResolvedValue({
      isValid: false,
      error: 'Token is empty'
    })

    const lifecycle = useTokenLifecycle()
    const result = await lifecycle.initialize('')

    expect(result.isValid).toBe(false)
    expect(lifecycle.isInitialized.value).toBe(false)
  })

  it('should not update lastRefreshTime when refresh is disabled', async () => {
    mockTokenManager.initializeToken.mockResolvedValue({
      isValid: true,
      expiryTime: Date.now() + 60000
    })

    const lifecycle = useTokenLifecycle()
    await lifecycle.initialize('test-token', { refreshEnabled: false })

    expect(lifecycle.lastRefreshTime.value).toBe(0)
  })

  it('should refresh token when needsRefresh returns true', async () => {
    mockTokenManager.needsRefresh.mockReturnValue(true)
    mockTokenManager.refreshToken.mockResolvedValue(true)

    const lifecycle = useTokenLifecycle()
    await lifecycle.refresh('http://example.com')

    expect(mockTokenManager.refreshToken).toHaveBeenCalled()
    expect(lifecycle.tokenValidationResult.value).toBeNull()
  })

  it('should set validation error when refresh fails', async () => {
    mockTokenManager.needsRefresh.mockReturnValue(true)
    mockTokenManager.refreshToken.mockResolvedValue(false)

    const lifecycle = useTokenLifecycle()
    await lifecycle.refresh('http://example.com')

    expect(lifecycle.tokenValidationResult.value).toEqual({
      isValid: false,
      error: 'Token refresh failed'
    })
  })

  it('should not refresh when needsRefresh returns false', async () => {
    mockTokenManager.needsRefresh.mockReturnValue(false)

    const lifecycle = useTokenLifecycle()
    await lifecycle.refresh('http://example.com')

    expect(mockTokenManager.refreshToken).not.toHaveBeenCalled()
  })

  it('should invalidate token and reset state', () => {
    const lifecycle = useTokenLifecycle()
    lifecycle.isInitialized.value = true
    lifecycle.tokenValidationResult.value = { isValid: true }

    lifecycle.invalidate()

    expect(mockTokenManager.invalidateToken).toHaveBeenCalled()
    expect(lifecycle.isInitialized.value).toBe(false)
    expect(lifecycle.tokenValidationResult.value).toBeNull()
  })

  it('should delegate needsRefresh to tokenManager', () => {
    mockTokenManager.needsRefresh.mockReturnValue(true)
    const lifecycle = useTokenLifecycle()
    expect(lifecycle.needsRefresh('http://example.com')).toBe(true)
    expect(mockTokenManager.needsRefresh).toHaveBeenCalled()
  })

  it('should delegate getCurrentTokenInfo to tokenManager', () => {
    const tokenInfo = { token: 'abc', expiryTime: 12345, type: 'iframe' as const }
    mockTokenManager.getCurrentTokenInfo.mockReturnValue(tokenInfo)
    const lifecycle = useTokenLifecycle()
    expect(lifecycle.getCurrentTokenInfo()).toEqual(tokenInfo)
  })

  it('should return validation status from ref', async () => {
    mockTokenManager.initializeToken.mockResolvedValue({
      isValid: true,
      expiryTime: Date.now() + 60000
    })

    const lifecycle = useTokenLifecycle()
    await lifecycle.initialize('test-token')

    const status = lifecycle.getValidationStatus()
    expect(status).toEqual({ isValid: true, expiryTime: expect.any(Number) })
  })

  it('should setup interval on mount and cleanup on unmount', () => {
    vi.useFakeTimers()
    mockTokenManager.needsRefresh.mockReturnValue(false)
    mockTokenManager.getCurrentTokenInfo.mockReturnValue({
      token: 'test',
      expiryTime: Date.now() + 600000
    })

    useTokenLifecycle()

    expect(onMountedCallback).not.toBeNull()
    onMountedCallback!()

    expect(onBeforeUnmountCallback).not.toBeNull()
    onBeforeUnmountCallback!()

    expect(mockTokenManager.cleanup).toHaveBeenCalled()
    vi.useRealTimers()
  })
})
