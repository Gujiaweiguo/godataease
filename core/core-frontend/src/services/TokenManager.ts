import { embeddedInitIframeApi, embeddedGetTokenArgsApi } from '@/api/embedded'
import { useEmbedded } from '@/store/modules/embedded'

export interface TokenValidationResult {
  isValid: boolean
  error?: string
  expiryTime?: number
}

export interface TokenInfo {
  token: string
  expiryTime?: number
  type?: 'iframe' | 'div' | 'module'
  resourceId?: string
}

export class TokenManager {
  private static instance: TokenManager
  private refreshIntervalId: number | null = null
  private readonly REFRESH_INTERVAL_MS = 5 * 60 * 1000

  private constructor(private embeddedStore: ReturnType<typeof useEmbedded>) {}

  static getInstance(embeddedStore: ReturnType<typeof useEmbedded>): TokenManager {
    if (!TokenManager.instance) {
      TokenManager.instance = new TokenManager(embeddedStore)
    }
    return TokenManager.instance
  }

  async initializeToken(token: string, origin: string, options?: { refreshEnabled?: boolean; tokenType?: 'iframe' | 'div' | 'module'; resourceId?: string }): Promise<TokenValidationResult> {
    const { refreshEnabled = true, tokenType = 'iframe', resourceId } = options || {}

    this.embeddedStore.setToken(token)

    try {
      const initResult = await embeddedInitIframeApi({ token, origin })
      if (Array.isArray(initResult?.data)) {
        this.embeddedStore.setAllowedOrigins(initResult.data)
      }
    } catch (error) {
      console.error('Embedded iframe initialization failed:', error)
      return {
        isValid: false,
        error: 'Initialization failed'
      }
    }

    const validation = await this.validateToken(token, origin)
    if (!validation.isValid) {
      return validation
    }

    if (refreshEnabled) {
      this.setupAutoRefresh()
    }

    this.embeddedStore.setTokenInfo(new Map([['current', {
      token,
      expiryTime: validation.expiryTime,
      type: tokenType,
      resourceId
    }]))

    return validation
  }

  async validateToken(token: string, origin: string): Promise<TokenValidationResult> {
    try {
      if (!token || token.length === 0) {
        return {
          isValid: false,
          error: 'Token is empty'
        }
      }

      const expiryTime = this.extractExpiryTime(token)

      if (expiryTime && Date.now() >= expiryTime) {
        return {
          isValid: false,
          error: 'Token has expired',
          expiryTime
        }
      }

      return {
        isValid: true,
        expiryTime
      }
    } catch (error) {
      console.error('Token validation failed:', error)
      return {
        isValid: false,
        error: 'Token validation error'
      }
    }
  }

  private extractExpiryTime(token: string): number | undefined {
    try {
      const parts = token.split('.')
      if (parts.length !== 3) return undefined

      const payload = JSON.parse(atob(parts[1]))
      return payload.exp ? payload.exp * 1000 : undefined
    } catch (error) {
      console.warn('Failed to extract expiry time from token:', error)
      return undefined
    }
  }

  async refreshToken(origin: string): Promise<boolean> {
    try {
      const tokenArgs = await embeddedGetTokenArgsApi()
      if (tokenArgs?.data?.token) {
        const newToken = tokenArgs.data.token

        this.embeddedStore.setToken(newToken)
        if (Array.isArray(tokenArgs.data.allowedOrigins)) {
          this.embeddedStore.setAllowedOrigins(tokenArgs.data.allowedOrigins)
        }

        const validation = await this.validateToken(newToken, origin)
        this.embeddedStore.setTokenInfo(new Map([['current', {
          token: newToken,
          expiryTime: validation.expiryTime,
          type: 'iframe'
        }]))

        return true
      }

      return false
    } catch (error) {
      console.error('Token refresh failed:', error)
      return false
    }
  }

  private setupAutoRefresh(): void {
    this.stopAutoRefresh()

    this.refreshIntervalId = window.setInterval(() => {
      this.refreshToken(window.location.origin)
    }, this.REFRESH_INTERVAL_MS)
  }

  stopAutoRefresh(): void {
    if (this.refreshIntervalId !== null) {
      clearInterval(this.refreshIntervalId)
      this.refreshIntervalId = null
    }
  }

  invalidateToken(): void {
    this.embeddedStore.setToken('')
    this.embeddedStore.setAllowedOrigins([])
    this.embeddedStore.setTokenInfo(new Map())
    this.stopAutoRefresh()
  }

  getCurrentTokenInfo(): TokenInfo | undefined {
    const tokenInfoMap = this.embeddedStore.getTokenInfo()
    return tokenInfoMap?.get('current')
  }

  needsRefresh(origin: string): boolean {
    const tokenInfo = this.getCurrentTokenInfo()

    if (!tokenInfo) {
      return true
    }

    if (tokenInfo.expiryTime && Date.now() >= tokenInfo.expiryTime) {
      return true
    }

    return false
  }

  cleanup(): void {
    this.stopAutoRefresh()
  }
}

export function useTokenManager(): TokenManager {
  const embeddedStore = useEmbedded()
  return TokenManager.getInstance(embeddedStore)
}
