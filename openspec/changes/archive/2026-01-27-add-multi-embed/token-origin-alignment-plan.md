# Token/Origin Alignment Improvement Plan

## Current State Analysis

### Token Initialization Points
Identified locations where `embeddedInitIframeApi()` is called:

1. **Chart Embedding**
   - `/views/chart/ChartView.vue`
   - Uses: `useEmbedded()` + `embeddedInitIframeApi()`
   - Entry: `componentMap` → `ViewWrapper` with `initIframe="initIframe"`

2. **Dashboard Embedding**
   - `/views/dashboard/index.vue`
   - Uses: `useEmbedded()` + `embeddedInitIframeApi()`
   - Entry: `componentMap` → `DashboardEditor` with `initIframe="initIframe"`

3. **Screen Embedding**
   - `/views/data-visualization/index.vue`
   - Uses: `useEmbedded()` + `embeddedInitIrameApi()`
   - Entry: `componentMap` → `VisualizationEditor` with `initIframe="initIframe"`

4. **Dataset/Datasource Embedding**
   - `/views/visualized/data/dataset/index.vue`
   - Uses: `useEmbedded()` + `embeddedInitIframeApi()`
   - Entry: `componentMap` → `DatasetEditor` with `initIframe="initIframe"`

5. **Mobile Panel Embedding**
   - `/views/mobile/panel/index.vue`
   - Uses: `useEmbedded()` + `embeddedInitIframeApi()`

6. **Iframe Entry Points (Dashboard Preview)**
   - `/views/mobile/panel/index.vue`
   - Uses: `init-iframe="DataFilling"` (hardcoded paths)

### Current Token/Origin Validation

**Origin Resolution:**
- **Implementation**: `resolveEmbeddedOrigin()` in `/utils/embedded.ts`
- **Logic**: Extract origin from `document.referrer`
- **Normalization**: Remove trailing slashes
- **Return Fallback**: Empty string on error

**Origin Validation:**
- **Implementation**: `isAllowedEmbeddedMessageOrigin()` in `/utils/embedded.ts`
- **Logic**: Compare message origin against allowed origins list
- **Parameters**: `origin`, `allowlist`, `enforceAllowlist`
- **Return**: Boolean indicating if origin is allowed

**Allowlist Sources:**
- `/embedded/domainListApi()` - Fetches backend allowlist
- `embeddedInitIframeApi()` - Stores result in `embeddedStore.setAllowedOrigins()`

### Business Type Flags (busiFlag)
Current mapping:
- `dashboard`: Dashboard resources
- `dataV`: Data visualization resources
- `dataset`: Dataset management
- `datasource`: Datasource management
- Used in `isAllowedEmbeddedMessageOrigin()` and embedded store

### Outer Parameters Handling
- **Encoding**: `outerParams` stored as JSON string (Base64 encoded)
- **Decoding**: `JSON.parse(Base64.decode())`
- **Usage**: Passed via URL params or `inject('embeddedParams')`

## Identified Issues

### Issue 1: No Unified Token Lifecycle
**Problem**:
- Each component calls `embeddedInitIframeApi()` independently
- Token is set in `embeddedStore.token` but no unified lifecycle
- No token refresh/renewal mechanism
- No token invalidation handling
- **Impact**: Components may use stale tokens after page is reloaded

**Evidence**:
```vue
// From /views/chart/ChartView.vue
const initIframe = async (name: string) => {
  if (embeddedStore.getToken) {
    try {
      const initResult = await embeddedInitIframeApi({
        token: embeddedStore.getToken,
        origin: resolveEmbeddedOrigin()
      })
      if (Array.isArray(initResult?.data)) {
        embeddedStore.setAllowedOrigins(initResult.data)
      }
    } catch (error) {
      console.error('Embedded iframe initialization failed', error)
    }
  }
}
```

**Recommendation**:
- Create unified token initialization manager service
- Centralize token lifecycle management
- Add token refresh interval (if needed)
- Implement token validation and auto-renewal

### Issue 2: No Standardized Outer Params Format
**Problem**:
- `outerParams` is a JSON string with no schema
- Multiple components parse `outerParams` differently
- No TypeScript interfaces for outer params
- Different key names across use cases

**Evidence**:
```typescript
// From embedded store
interface AppState {
  outerParams: string  // Plain string, no structure
}
```

**Recommendation**:
- Define TypeScript interfaces for outer params structure
- Create utility functions for safe parsing/encoding/decoding
- Document all supported outer param keys

### Issue 3: Duplicate Origin Validation Logic
**Problem**:
- `isAllowedEmbeddedMessageOrigin()` is implemented in multiple places
- Each component validates independently
- `enforceAllowlist` flag usage unclear

**Evidence**:
```typescript
// From /views/data-visualization/PreviewCanvas.vue
const handleMessage = e => {
  if (
    !isAllowedEmbeddedMessageOrigin(
      event.origin,
      embeddedStore.getAllowedOrigins,
      Boolean(embeddedStore.getToken)
    )
  ) {
    return
  }
}
```

**Recommendation**:
- Centralize origin validation in utility functions
- Add consistent error handling
- Document `enforceAllowlist` flag usage

### Issue 4: Scattered Token Storage
**Problem**:
- `token` is only in `embeddedStore.token`
- `tokenInfo` is a separate Map
- No unified token store across embed types

**Evidence**:
```typescript
// From embedded store
state: (): AppState => {
  return {
    token: '',
    tokenInfo: new Map()  // Stores separate token info
    allowedOrigins: [],
  }
}
```

**Recommendation**:
- Consider unified token storage architecture
- Define clear token usage patterns
- Add token type indicators (iframe vs DIV vs module)

### Issue 5: Inconsistent Message Patterns
**Problem**:
- Multiple `postMessage` patterns in use
- Event names are ad-hoc: `de_inner_params`, `canvas_init_ready`, `dataease-embedded-interactive`
- No standardized payload structures
- No clear event types definition

**Evidence**:
```typescript
// Parent to child (from ViewWrapper.vue)
window.parent.postMessage(targetPm, '*')  // Type: 'dataease-embedded-interactive'
// Component to parent (from PreviewCanvas.vue)
eventBus.emit('doCanvasInit-canvas-main')  // Event name with context

// Child to parent (from multiple components)
window['dataease-embedded-host']  // Global handler
```

**Recommendation**:
- Create event type registry
- Standardize event payload structures
- Document all supported events with examples
- Add event listener management utilities

## Proposed Improvements

### 1. Create Unified Token Initialization Manager

**Purpose**: Centralize all token initialization logic

**Components**:
1. `TokenManager` class/service
   - Methods: `initializeToken()`, `refreshToken()`, `invalidateToken()`, `validateToken()`
   - Handles both iframe and DIV tokens
   - Manages token lifecycle

2. `TokenStore` extension
   - Add `tokenType`, `tokenExpiry`, `lastRefreshTime` to embedded store
   - Add `tokenValidationStatus` for token errors

3. `TokenLifecycleHook` composable
   - Auto-inject token validation into Vue components
   - Handle token expiry and auto-refresh

### 2. Create Outer Params Standardization

**Purpose**: Standardize outer params format and handling

**Components**:
1. TypeScript interfaces:
   ```typescript
   interface EmbeddingOuterParams {
     resourceId?: string
     dvId?: string
     chartId?: string
     busiFlag?: string
     outerParams?: Record<string, any>
     callbackParams?: Record<string, any>
   }
   ```

2. Utility functions:
   - `encodeOuterParams(params)`: Encode to Base64 string
   - `decodeOuterParams(encoded)`: Decode from Base64 string
   - `getOuterParams(urlParams)`: Parse from URL params
   - `validateOuterParams(params)`: Validate structure and required fields

### 3. Create Origin Validation Standardization

**Purpose**: Centralize and simplify origin validation

**Functions**:
1. `validateOrigin(origin: string, allowedOrigins: string[], enforceAllowlist: boolean): ValidationResult`
2. `isOriginInAllowlist(origin: string, allowlist: string[]): boolean`
3. `getNormalizedOrigin(origin: string): string`

**Implementation**:
```typescript
import { resolveEmbeddedOrigin, isAllowedEmbeddedMessageOrigin } from '@/utils/embedded'

interface ValidationResult {
  isValid: boolean
  normalizedOrigin: string
  error?: string
}

export function validateOrigin(
  origin: string,
  allowedOrigins: string[],
  enforceAllowlist: boolean = false
): ValidationResult {
  const normalizedOrigin = normalizeOrigin(origin)
  const isInAllowlist = isOriginInAllowlist(normalizedOrigin, allowedOrigins)
  const isValid = isInAllowlist || (origin in window.location.origin)
  const error = !isValid ? `Origin ${origin} is not in allowlist` : undefined
  
  return { isValid, normalizedOrigin, error }
}
```

### 4. Create Message Protocol Standardization

**Purpose**: Standardize postMessage event types and payloads

**Event Types Registry**:
```typescript
enum EmbeddingEventType {
  INIT_READY = 'init_ready',
  PARAM_UPDATE = 'param_update',
  INTERACTION = 'user_interaction',
  ERROR = 'error',
  READY = 'ready',
  ATTACH_PARAMS = 'attach_params',
  JUMP_TO_TARGET = 'jump_to_target'
}

// Event Payload Interfaces
interface InitReadyPayload {
  resourceId: string
  dvId?: string
  chartId?: string
}

interface ParamUpdatePayload {
  [key: string]: any
}

interface UserInteractionPayload {
  param: any
  value: any
}

interface ErrorPayload {
  message: string
  error: Error
}
```

**Standardized Emitter**:
```typescript
import { useEmbedded } from '@/store/modules/embedded'

export function emitEmbeddingEvent(
  type: EmbeddingEventType,
  payload: InitReadyPayload | ParamUpdatePayload | UserInteractionPayload | ErrorPayload
) {
  window.parent.postMessage({ type, payload }, '*')
}
```

### 5. Create Token Storage Architecture

**Purpose**: Unified token storage across all embed types

**Enhanced Store Schema**:
```typescript
interface AppState {
  // Existing fields (keep for compatibility)
  type: string
  busiFlag: string
  outerParams: string
  suffixId: string
  baseUrl: string
  dvId: string
  pid: string
  chartId: string
  resourceId: string
  dfId: string
  opt: string
  createType: string
  templateParams: string
  jumpInfoParam: string
  outerUrl: string
  datasourceId: string
  tableName: string
  datasetId: string
  datasetCopyId: string
  datasetPid: string
  allowedOrigins: string[]
  
  // New fields
  tokenType: 'iframe' | 'div' | 'module'
  tokenExpiry: number | null
  lastRefreshTime: number | null
  tokenValidationStatus: 'valid' | 'expired' | 'invalid'
  currentToken: string
  currentTokenId: string | null
}
```

**Token Types**:
- `iframe`: For iframe-based embedding (ChartView, Dashboard, etc.)
- `div`: For DIV-based embedding (ViewWrapper, DashboardPreview)
- `module`: For module-level page embedding
- `external`: For external app integration

### 6. BusiFlag Type Registry

**Current Types**:
- `dashboard`: Dashboard embedding
- `dataV`: Data visualization embedding
- `dataset`: Dataset management embedding
- `datasource`: Datasource management embedding
- `module`: Module-level page embedding (NEW)

**Validation Rules**:
```typescript
const BUSI_FLAG_TYPES = {
  dashboard: 'dashboard',
  dataV: 'dataV',
  dataset: 'dataset',
  datasource: 'datasource',
  module: 'module' // NEW
} as const

function getBusiFlagForResource(resourceId: string): string {
  // Map resource IDs to busiFlag types
  const mapping = {
    'dashboard': 'dashboard',
    'dataV': 'dataV',
    'dataset': 'dataset',
    'datasource': 'datasource'
  }
  return mapping[resourceId] || 'dashboard' // Default
}
```

### 7. Enhanced Token Initialization Flow

**Current Flow**:
1. Component checks if token exists in `embeddedStore.token`
2. If yes, skip initialization
3. Call `embeddedInitIframeApi({ token, origin })`
4. Parse allowlist response
5. Set `embeddedStore.setAllowedOrigins(allowlist)`

**Improved Flow**:
```typescript
// In component
import { useEmbedded } from '@/store/modules/embedded'
import { initializeToken } from '@/utils/tokenManager'

const embeddedStore = useEmbedded()

async function initEmbeddedIframe(embedType: 'iframe' | 'div') {
  if (embeddedStore.tokenType === 'module' || embeddedStore.currentTokenId) {
    const newToken = await initializeToken(embedType)
    embeddedStore.setCurrentToken(newToken.id)
    embeddedStore.setToken(newToken.token)
    embeddedStore.setTokenType(embedType)
  }
  
  const token = embeddedStore.currentToken
  const origin = resolveEmbeddedOrigin()
  const allowlist = await fetchAllowlist()
  
  const initResult = await embeddedInitIframeApi({
    token,
    origin,
    embedType: embeddedStore.tokenType // NEW: specify embed type
  })
  
  if (Array.isArray(initResult?.data)) {
    embeddedStore.setAllowedOrigins(initResult.data)
  }
}
```

### 8. Compatibility Strategy

**Goal**: Maintain backward compatibility while improving token management

**Approach**:
1. Keep existing `embeddedStore.token` and `embeddedStore.getAllowedOrigins()` working
2. Add new token fields gradually to store
3. Use feature flags for rollout
4. Deprecate old patterns gradually

**Migration Path**:
- Phase 1: Add new fields to store (non-breaking)
- Phase 2: Update initialization logic to use new fields
- Phase 3: Update components gradually
- Phase 4: Remove deprecated code

### 9. Module-Level Page Embedding Support

**Purpose**: Support `/module-*` routes with tree navigation

**Proposed Routes**:
```typescript
// Module-level page routes
{
  path: '/module/market',
  name: 'module-market',
  redirect: '/module-market/index',
  component: () => import('@/components/module/ModulePageWithTree.vue'),
  meta: { title: '应用市场' }
}
```

**Components**:
1. `ModulePageWithTree.vue` - Main component with tree navigation
2. `ModulePageContent.vue` - Content area (dynamic loading based on selection)
3. `ModulePageRoute.vue` - Route wrapper for permissions/context

### 10. Implementation Priorities

**High Priority**:
1. Create unified token initialization manager
2. Standardize outer params format
3. Standardize message protocol
4. Enhance origin validation utilities

**Medium Priority**:
5. Add module-level page routing support
6. Create module page tree navigation component

**Low Priority**:
7. Comprehensive testing across all embed types
8. Documentation and examples

## Success Criteria

- [ ] All components use unified token initialization
- [ ] Token validation is consistent across iframe/DIV/module embedding
- [ ] Message protocols are standardized
- [ ] Outer params are type-safe and well-documented
- [ ] Token lifecycle is managed (refresh, expiry, validation)
- [ ] Module-level pages can be embedded with tree navigation
- [ ] All existing embeddings continue to work
- [ ] Documentation covers new unified approach

## Next Steps

1. Implement TokenManager service and utilities
2. Extend embedded store with new token fields
3. Update `embeddedInitIframeApi` to accept `embedType` parameter
4. Create TypeScript interfaces for outer params
5. Refactor origin validation utilities
6. Implement message protocol registry
7. Add module-level page routing infrastructure
8. Update existing components to use new token manager
9. Add comprehensive tests
10. Update documentation

## Open Questions

1. Should token validation be strict (deny by default) or permissive (allow by default)?
2. Should module-level pages use a common `/module-*` prefix or custom prefixes?
3. Do we need to support backward compatibility with custom embed integrations?
4. What is the target date for this feature completion?
