import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { TokenManager } from '@/services/TokenManager'
import { embeddedGetTokenArgsApi, embeddedInitIframeApi } from '@/api/embedded'

vi.mock('@/api/embedded', async () => {
  const { createEmbeddedApiModuleMock } = await import('../../../../tests/unit/helpers')
  return createEmbeddedApiModuleMock()
})

const mockStore = {
  token: '',
  allowedOrigins: [] as string[],
  setToken: vi.fn(),
  setAllowedOrigins: vi.fn(),
  setTokenInfo: vi.fn(),
  getTokenInfo: new Map<string, any>()
}

describe('TokenManager', () => {
  let tokenManager: TokenManager

  beforeEach(() => {
    vi.clearAllMocks()
    ;(TokenManager as any).instance = undefined
    mockStore.getTokenInfo = new Map()
    vi.spyOn(console, 'error').mockImplementation(() => undefined)
    vi.spyOn(console, 'warn').mockImplementation(() => undefined)
    tokenManager = TokenManager.getInstance(mockStore as any)
  })

  afterEach(() => {
    tokenManager.cleanup()
    vi.restoreAllMocks()
  })

  it('initializes token successfully', async () => {
    vi.mocked(embeddedInitIframeApi).mockResolvedValue({ data: ['https://a.com'] } as any)
    const result = await tokenManager.initializeToken('token-1', 'https://host.com')

    expect(result.isValid).toBe(true)
    expect(mockStore.setToken).toHaveBeenCalledWith('token-1')
    expect(mockStore.setAllowedOrigins).toHaveBeenCalledWith(['https://a.com'])
    expect(mockStore.setTokenInfo).toHaveBeenCalled()
  })

  it('returns invalid when init api fails', async () => {
    vi.mocked(embeddedInitIframeApi).mockRejectedValue(new Error('fail'))
    const result = await tokenManager.initializeToken('token-1', 'https://host.com')

    expect(result.isValid).toBe(false)
    expect(result.error).toBe('Initialization failed')
  })

  it('rejects empty token', async () => {
    const result = await tokenManager.validateToken('')
    expect(result.isValid).toBe(false)
    expect(result.error).toBe('Token is empty')
  })

  it('parses jwt expiry time', async () => {
    const exp = Math.floor((Date.now() + 3600_000) / 1000)
    const token = `${btoa('{"alg":"HS256"}')}.${btoa(`{"exp":${exp}}`)}.sig`
    const result = await tokenManager.validateToken(token)

    expect(result.isValid).toBe(true)
    expect(result.expiryTime).toBe(exp * 1000)
  })

  it('refreshes token from api', async () => {
    vi.mocked(embeddedGetTokenArgsApi).mockResolvedValue({
      data: { token: 'new-token', allowedOrigins: ['https://new.com'] }
    } as any)
    const ok = await tokenManager.refreshToken()

    expect(ok).toBe(true)
    expect(mockStore.setToken).toHaveBeenCalledWith('new-token')
    expect(mockStore.setAllowedOrigins).toHaveBeenCalledWith(['https://new.com'])
  })

  it('returns false when refresh response has no token', async () => {
    vi.mocked(embeddedGetTokenArgsApi).mockResolvedValue({ data: {} } as any)
    const ok = await tokenManager.refreshToken()
    expect(ok).toBe(false)
  })

  it('invalidates token state', () => {
    tokenManager.invalidateToken()
    expect(mockStore.setToken).toHaveBeenCalledWith('')
    expect(mockStore.setAllowedOrigins).toHaveBeenCalledWith([])
    expect(mockStore.setTokenInfo).toHaveBeenCalledWith(new Map())
  })

  it('returns current token info from map', () => {
    mockStore.getTokenInfo = new Map([['current', { token: 'x', expiryTime: 123 }]])
    const info = tokenManager.getCurrentTokenInfo()
    expect(info).toEqual({ token: 'x', expiryTime: 123 })
  })

  it('needs refresh when token missing', () => {
    mockStore.getTokenInfo = new Map()
    expect(tokenManager.needsRefresh()).toBe(true)
  })
})
