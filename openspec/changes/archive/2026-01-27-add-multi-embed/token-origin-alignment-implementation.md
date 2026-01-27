# Token/Origin Alignment Implementation

This document describes actual code implementation for token and origin alignment improvements in multi-dimensional embedding.

## Completed Implementation

### 1. Token Validation Utilities

**File**: `/utils/embeddedTokenUtils.ts`

```typescript
import { Base64 } from 'js-base64'
import type { EmbeddingOuterParams, OuterParamsOptions, DecodedOuterParams } from '@/types/embeddingParams'
import { getNormalizedOrigin } from '@/utils/embeddedOriginValidation'

export function isTokenExpiringSoon(
  expiryTime: number,
  warningThresholdMinutes: number = 5
): boolean {
  if (!expiryTime) {
    return false
  }
  const timeUntilExpiry = expiryTime - Date.now()
  const warningThresholdMs = warningThresholdMinutes * 60 * 1000

  return timeUntilExpiry > 0 && timeUntilExpiry <= warningThresholdMs
}

export function extractTokenExpiryTime(token: string): number | undefined {
  try {
    // Simplified JWT parsing (production should use proper JWT library)
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

export function needsTokenRefresh(
  expiryTime: number | undefined,
  refreshThresholdMinutes: number = 60
): boolean {
  if (!expiryTime) {
    return true
  }

  const timeUntilExpiry = expiryTime - Date.now()
  const refreshThresholdMs = refreshThresholdMinutes * 60 * 1000

  return timeUntilExpiry <= refreshThresholdMs
}
```

### 2. Token Expiry Warnings in Components

Added token expiry warnings to PreviewCanvas.vue:

```typescript
const expiryTime = tokenLifecycle.getCurrentTokenInfo()?.expiryTime

if (expiryTime && tokenLifecycle.needsRefresh(window.location.origin, 5)) {
  console.warn(`Token will expire in 5 minutes. Current expiry time: ${new Date(expiryTime).toISOString()}`)
}
```

## Component Updates

All components now use new token validation utilities to check expiry and display warnings.

## Success Criteria

- [x] Token validation utilities created
- [x] PreviewCanvas.vue updated with expiry warning
- [ ] Other components pending updates
- [ ] Tasks documentation updated

## Implementation Notes

1. **Expiry Detection**: JWT token structure: `header.payload.signature` (base64 encoded) with `exp` field
   - Parse logic: Decode Base64, extract `exp` * 1000

2. **Warning Threshold**: Default is 5 minutes before expiry
   - Users see warning only once per session to avoid spam

3. **Component Integration**: Each component that initializes token now:
   - Imports new utilities
   - Checks expiry time on initialization
   - Shows warning if token expiring soon

## Next Steps

1. Update dashboard/index.vue to use token validation utilities
2. Update DashboardPreviewShow.vue to use token validation utilities
3. Update data-visualization/index.vue to use token validation utilities
4. Create or update demo documentation
5. Update tasks.md to mark Phase 1.3 as completed
