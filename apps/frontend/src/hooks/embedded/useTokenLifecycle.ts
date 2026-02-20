import { onBeforeUnmount, onMounted, type Ref } from 'vue'
import { useEmbedded } from '@/store/modules/embedded'
import { useTokenManager } from '@/services/TokenManager'
import type { TokenValidationResult } from '@/services/TokenManager'

export function useTokenLifecycle() {
  const embeddedStore = useEmbedded()
  const tokenManager = useTokenManager()
  const tokenValidationResult: Ref<TokenValidationResult | null> = ref(null)
  const isInitialized = ref(false)
  const lastRefreshTime = ref<number>(0)

  /**
   * Initialize token lifecycle on mount.
   */
  const initialize = async (
    token: string,
    options?: {
      refreshEnabled?: boolean
      tokenType?: 'iframe' | 'div' | 'module'
      resourceId?: string
    }
  ) => {
    const origin = window.location.origin
    const validation = await tokenManager.initializeToken(token, origin, options)

    tokenValidationResult.value = validation
    isInitialized.value = validation.isValid

    if (validation.isValid && options?.refreshEnabled) {
      lastRefreshTime.value = Date.now()
    }

    return validation
  }

  /**
   * Refresh token if needed.
   */
  const refresh = async (origin: string) => {
    const needsRefreshCheck = tokenManager.needsRefresh(origin)

    if (needsRefreshCheck) {
      lastRefreshTime.value = Date.now()
      const success = await tokenManager.refreshToken(origin)

      if (!success) {
        console.warn('Token refresh failed')
        tokenValidationResult.value = {
          isValid: false,
          error: 'Token refresh failed'
        }
      } else {
        console.info('Token refreshed successfully')
        tokenValidationResult.value = null
      }
    }
  }

  /**
   * Invalidate current token.
   */
  const invalidate = () => {
    tokenManager.invalidateToken()
    tokenValidationResult.value = null
    isInitialized.value = false
  }

  /**
   * Check if token needs refresh.
   */
  const needsRefresh = (origin: string): boolean => {
    return tokenManager.needsRefresh(origin)
  }

  /**
   * Get current token validation status.
   */
  const getValidationStatus = (): TokenValidationResult | null => {
    return tokenValidationResult.value
  }

  /**
   * Setup lifecycle hooks.
   */
  onMounted(() => {
    const refreshCheckInterval = setInterval(() => {
      const origin = window.location.origin
      const needsRefreshCheck = tokenManager.needsRefresh(origin)
      const tokenInfo = tokenManager.getCurrentTokenInfo()

      if (needsRefreshCheck && tokenInfo?.expiryTime) {
        const timeUntilExpiry = tokenInfo.expiryTime - Date.now()
        const warningThreshold = 5 * 60 * 1000
        if (timeUntilExpiry <= warningThreshold && timeUntilExpiry > 0) {
          console.warn(`Token expiring in ${Math.ceil(timeUntilExpiry / 60000)} minutes`)
        }
      }
    }, 60 * 1000)

    onBeforeUnmount(() => {
      clearInterval(refreshCheckInterval)
      tokenManager.cleanup()
    })
  })

  return {
    initialize,
    refresh,
    invalidate,
    needsRefresh,
    getValidationStatus,
    isInitialized,
    tokenValidationResult,
    lastRefreshTime
  }
}
