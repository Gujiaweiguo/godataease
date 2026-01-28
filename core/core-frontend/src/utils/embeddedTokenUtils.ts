import { Base64 } from 'js-base64'
import type { EmbeddingOuterParams, OuterParamsOptions, DecodedOuterParams } from '@/types/embeddingParams'
import { getNormalizedOrigin } from '@/utils/embedded'
import type { OriginValidationOptions, OriginValidationResult } from '@/types/embeddingOrigin'

export function encodeOuterParams(params: EmbeddingOuterParams, options: OuterParamsOptions = {}): string {
  const format = options.format || 'json'
  const jsonStr = JSON.stringify(params)
  return format === 'base64' ? Base64.encode(jsonStr) : jsonStr
}

export function decodeOuterParams(encodedParams: string, options: OuterParamsOptions = {}): DecodedOuterParams | { params: Record<string, any> } {
  const format = options.format || 'json'
  const paramsStr = format === 'base64' ? Base64.decode(encodedParams) : encodedParams
  
  try {
    const paramsObj = JSON.parse(paramsStr) as DecodedOuterParams
    return {
      params: paramsObj,
      isValid: true
    }
  } catch (error) {
    console.error('Failed to decode outer params:', error)
    return {
      params: {},
      isValid: false,
      error: 'Invalid outer params format'
    }
  }
}

export function validateOuterParams(params: unknown, requiredFields: string[] = []): { isValid: boolean; error?: string } {
  if (!params || typeof params !== 'object') {
    return {
      isValid: false,
      error: 'Params must be an object'
    }
  }
  
  const missingFields = requiredFields.filter(field => !(field in params))
  if (missingFields.length > 0) {
    return {
      isValid: false,
      error: `Missing required fields: ${missingFields.join(', ')}`
    }
  }
  
  return {
    isValid: true,
    error: undefined
  }
}

export function isTokenExpiringSoon(expiryTime: number | undefined, warningThresholdMinutes: number = 5): boolean {
  if (!expiryTime) {
    return false
  }
  
  const timeUntilExpiry = expiryTime - Date.now()
  const warningThresholdMs = warningThresholdMinutes * 60 * 1000
  
  return timeUntilExpiry > 0 && timeUntilExpiry <= warningThresholdMs
}

export function extractTokenExpiryTime(token: string): number | undefined {
  try {
    const parts = token.split('.')
    if (parts.length !== 3) {
      return undefined
    }
    
    const payload = JSON.parse(atob(parts[1]))
    return payload.exp ? payload.exp * 1000 : undefined
  } catch (error) {
    console.warn('Failed to extract token expiry time:', error)
    return undefined
  }
}

export function needsTokenRefresh(expiryTime: number | undefined, refreshThresholdMinutes: number = 60): boolean {
  if (!expiryTime) {
    return true
  }
  
  const timeUntilExpiry = expiryTime - Date.now()
  const refreshThresholdMs = refreshThresholdMinutes * 60 * 1000
  
  return timeUntilExpiry <= refreshThresholdMs
}
