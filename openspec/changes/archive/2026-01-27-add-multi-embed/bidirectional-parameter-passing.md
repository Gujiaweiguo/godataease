# Bidirectional Parameter Passing - Implementation

## Overview

This document provides implementation details for standardized bidirectional parameter passing between host and embedded systems (iframe/DIV/module-level pages) in DataEase.

## Phase 1: Define Standardized Interfaces

### 1.1 Parent → Child (Host → Embedded)
```typescript
// /store/modules/embedded/types/bidirectional.ts
export interface HostToChildMessage {
  messageType: 'param_update' | 'user_interaction' | 'canvas_init_ready' | 'error' | 'ready'
  type: 'param_update' | 'user_interaction' | 'canvas_init_ready' | 'error' | 'ready'
  payload: {
    param1?: string | number | boolean
    param2?: string | number | boolean
    resourceId?: string
    dvId?: string
    chartId?: string
    busiFlag?: 'dashboard' | 'dataV' | 'dataset' | 'datasource' | 'module'  // NEW
  }
}

export interface ChildToHostMessage {
  messageType: 'de_inner_params' | 'canvas_init_ready' | 'ready' | 'dataease-embedded-interactive'
  type: 'de_inner_params' | 'canvas_init_ready' | 'ready' | 'dataease-embedded-interactive'
  payload: {
    innerParams: Record<string, any>
    resourceId?: string
    dvId?: string
    chartId?: string
    busiFlag?: string
    outerParams?: string
    suffixId?: string
    chartId?: string
    pid?: string
    params?: string
    components?: Array<{
      id: string
      type: 'view' | 'editor' | 'template' | 'screen'
      options?: any
    mode?: 'iframe' | 'div' | 'module'
    }
}
```

### 1.2 Child → Parent (Embedded → Host)
```typescript
// /store/modules/embedded/composables/useParentCommunication.ts
export function useParentCommunication() {
  return {
    emitInit: (resourceId: string, busiFlag: string) => void,
    emitUpdate: (params: Record<string, any>) => void,
    emitInteraction: (param: any, value: any) => void,
    emitReady: () => void,
    emitError: (error: string) => void
  }
}
```

### 1.3 Parent → Child (DIV + IFrame)
```typescript
// /store/modules/embedded/composables/useIframeCommunication.ts
export function useIframeCommunication() {
  return {
    initIframe: (token: string, origin: string) => Promise<string[]>,
    postMessage: (message: string, payload?: object) => void,
    onMessage: (handler: (message: string, payload?: object) => void,
    ready: () => boolean
  }
}
```

## Phase 2: Implement Parent Communication Hook

### 2.1 Create Parent Communication Composable

**File**: `/store/modules/embedded/composables/useParentCommunication.ts`

```typescript
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { HostToChildMessage, ChildToHostMessage } from './types/bidirectional'

export function useParentCommunication() {
  const iframeMode = ref(false)
  const divMode = ref(false)

  const emitInit = (resourceId: string, busiFlag: string) => void => {
    window.parent.postMessage({
      type: 'init_ready',
      payload: { resourceId }
    }, '*')
  }

  const emitUpdate = (params: Record<string, any>) => void => {
    window.parent.postMessage({
      type: 'param_update',
      payload
    }, '*')
  }

  const emitInteraction = (param: any, value: any) => void => {
    window.parent.postMessage({
      type: 'user_interaction',
      payload: { param, value }
    }, '*')
  }

  const emitReady = () => void => {
    window.parent.postMessage({ type: 'ready' }, '*')
  }

  const emitError = (error: string) => void => {
    window.parent.postMessage({ type: 'error', message: error }, '*')
  }

  const emitInitReady = () => void => {
    window.parent.postMessage({
      type: 'canvas_init_ready',
      payload: { }
    }, '*')
  }

  const onMessage = (handler: (message: string, payload?: object) => void => {
    try {
      const parsed = JSON.parse(message)
      if (parsed.type && parsed.payload) {
        if (parsed.type === 'de_inner_params') {
          // Handle inner params update
          const { resourceId, chartId, busiFlag, ... } = parsed.payload
          // Update local state based on busiFlag and resourceId
        } else if (parsed.type === 'canvas_init_ready') {
          // Canvas ready event
          const { resourceId, dvId } = parsed.payload
          if (parsed.payload === '*') {
            if (resourceId && dvId) {
              embeddedStore.setResourceId(resourceId)
            }
            embeddedStore.setDvId(dvId)
          }
        }
      } else if (parsed.type === 'user_interaction') {
        // User interaction from embedded page
        const { param, value } = parsed.payload
        // Handle user actions
      } else if (parsed.type === 'ready') {
        // Embedded component ready
        embeddedStore.setEmbeddedReady(true)
      } else if (parsed.type === 'error') {
        console.error('Received error message:', message)
      }
    } catch (error) {
      console.error('Parent message parsing failed:', error)
    }
  }

  const ready = ref(false)

  // Cleanup on unmount
  onBeforeUnmount(() => {
    ready.value = false
  })
}
```

### 2.2 Create IFrame Communication Composable

**File**: `/store/modules/embedded/composables/useIframeCommunication.ts`

```typescript
import { computed } from 'vue'
import { useEmbedded } from '../store/modules/embedded'
import { resolveEmbeddedOrigin } from '@/utils/embedded'

export function useIframeCommunication() {
  const embeddedStore = useEmbedded()
  const ready = computed(() => {
    return embeddedStore.getType() === 'iframe' && embeddedStore.getToken()
  })

  const postMessage = (message: string, payload?: object) => void => {
    if (ready.value) {
      window.parent.postMessage({ type: 'ready' }, '*')
    }
  }

  const onMessage = (message: string, payload?: object) => void => {
    try {
      const parsed = JSON.parse(message)
      if (parsed.type && parsed.payload) {
        switch (parsed.type) {
          case 'de_inner_params':
            handleInnerParamsUpdate(parsed.payload)
            break
          case 'canvas_init_ready':
            handleCanvasInitReady(parsed.payload)
            break
          case 'user_interaction':
            handleUserInteraction(parsed.payload)
            break
          case 'ready':
            handleComponentReady(parsed.payload)
            break
          case 'error':
            handleError(parsed.payload)
            break
        }
      }
    } catch (error) {
      console.error('IFrame message parsing failed:', error)
    }
  }

  function handleInnerParamsUpdate(payload: any) {
    const { resourceId, chartId, busiFlag } = payload
    // Update local state based on busiFlag
    if (busiFlag === 'dashboard') {
      embeddedStore.setResourceId(resourceId)
    } else if (busiFlag === 'dataV') {
      embeddedStore.setDvId(dvId)
    }
  }

  function handleCanvasInitReady(payload: any) {
    const { resourceId, dvId } = payload
    // Canvas initialization
    if (resourceId) {
      embeddedStore.setResourceId(resourceId)
    }
    if (dvId) {
      embeddedStore.setDvId(dvId)
    }
  }

  function handleUserInteraction(payload: any) {
    const { param, value } = payload
    // Handle user actions from embedded page
    console.log('User interaction from embedded:', { param, value })
  }

  function handleComponentReady(payload: any) {
    const { resourceId, dvId } = payload
    // Component ready event
    if (resourceId) {
      embeddedStore.setResourceId(resourceId)
    }
    if (dvId) {
      embeddedStore.setDvId(dvId)
    }
  }

  function handleError(payload: any) {
    const { message } = payload
    console.error('Received error from embedded:', message)
  }

  // Ready check
  onMounted(() => {
    if (ready.value) {
      window.parent.postMessage({ type: 'ready' }, '*')
      ready.value = false
    }
  })
}
```

### 2.3 Create DIV + Module Parent Communication Composable

**File**: `/store/modules/embedded/composables/useDivAndModuleParentCommunication.ts`

```typescript
import { ref, computed } from 'vue'
import { useEmbedded } from '../store/modules/embedded'
import { resolveEmbeddedOrigin, isAllowedEmbeddedMessageOrigin } from '@/utils/embedded'

export function useDivAndModuleParentCommunication() {
  const embeddedStore = useEmbedded()
  const divMode = ref(false)
  const moduleMode = ref('module') // NEW for module-level page embedding

  const emitInit = (resourceId: string, busiFlag: string) => void => {
    window.parent.postMessage({
      type: 'module_init',
      payload: { mode: 'module', resourceId, busiFlag }
    }, '*')
  }

  const emitUpdate = (params: Record<string, any>) => void => {
    window.parent.postMessage({
      type: 'module_update',
      payload
    }, '*')
  }

  const emitInteraction = (param: any, value: any) => void => {
    window.parent.postMessage({
      type: 'module_interaction',
      payload: { mode: 'module', param, value }
    }, '*')
  }

  const onMessage = (message: string, payload?: object) => void => {
    try {
      const parsed = JSON.parse(message)
      if (parsed.type && parsed.payload) {
        if (parsed.type === 'de_inner_params') {
          handleModuleInnerParamsUpdate(parsed.payload)
        } else if (parsed.type === 'ready') {
          handleModuleReady(parsed.payload)
        } else if (parsed.type === 'module_interaction') {
          handleModuleInteraction(parsed.payload)
        } else if (parsed.type === 'error') {
          handleError(parsed.payload)
        }
      }
    } catch (error) {
      console.error('Module parent message parsing failed:', error)
    }
  }

  function handleModuleInnerParamsUpdate(payload: any) {
    const { params, resourceId, dvId, mode } = payload
    // Update module-level inner params
    embeddedStore.setResourceId(resourceId)
    if (dvId) {
      embeddedStore.setDvId(dvId)
    }
  }

  function handleModuleReady(payload: any) {
    const { resourceId, dvId, mode } = payload
    // Module ready event
    if (mode === 'dashboard' || mode === 'dataV') {
      embeddedStore.setResourceId(resourceId)
    }
    embeddedStore.setModuleReady(true)
  }

  function handleModuleInteraction(payload: any) {
    const { param, value } = payload
    // Module user interaction from embedded page
    console.log('Module interaction from embedded:', { param, value })
  }

  function handleError(payload: any) {
    const { message } = payload
    console.error('Received error from module:', message)
  }

  // Mode switching
  const setDivMode = () => { divMode.value = true; moduleMode.value = false }
  const setModuleMode = () => { divMode.value = false; moduleMode.value = true }

  // Ready check
  const ready = computed(() => {
    return embeddedStore.getType() === 'div' && embeddedStore.getToken()
  })

  onMounted(() => {
    if (ready.value) {
      window.parent.postMessage({ type: 'ready' }, '*')
      ready.value = false
    }
  })
}
```

## Phase 3: Integration in Key Components

### 3.1 Update `/views/chart/ChartView.vue`

Add bidirectional parameter support to chart embedding:

```vue
<template>
  <XpackComponent
    jsname="L21lbnUvZGF0YS9kYXRhLWZpbGxpbmcvbWFuYWdlL2Zvcm0vaW5kZXg='"
    @init-iframe="initIframe"
    @bind:event-check="eventCheck"
    @bind:event-check="eventCheck"
    @bind:event-check="eventCheck"
    @bind:event-check="eventCheck"
    @bind:event-check="eventCheck"
    @bind:event-check="eventCheck"
  />

  <script lang="ts" setup>
  import { shallowRef, ref, onMounted, onBeforeUnmount, computed } from 'vue'
  import { debounce } from 'lodash-es'
  import { XpackComponent } from '@/components/plugin'
  import { useEmbedded } from '@/store/modules/embedded'
  import { embeddedInitIframeApi } from '@/api/embedded'
  import { resolveEmbeddedOrigin } from '@/utils/embedded'
  import { isAllowedEmbeddedMessageOrigin } from '@/utils/embedded'
  import { useParentCommunication } from './composables/useParentCommunication'

  const embeddedStore = useEmbedded()
  const parentCommunication = useParentCommunication()

  const { close } = useLoading()

  const currentComponent = shallowRef()
  const Preview = defineAsyncComponent(() => import('@/views/data-visualization/PreviewCanvas.vue'))
  const VisualizationEditor = defineAsyncComponent(
    () => import('@/views/data-visualization/index.vue')
  )
  const DashboardEditor = defineAsyncComponent(() => import('@/views/dashboard/index.vue'))
  const ViewWrapper = defineAsyncComponent(() => import('@/pages/panel/ViewWrapper.vue'))
  const Dataset = defineAsyncComponent(() => import('@/views/visualized/data/dataset/index.vue'))
  const Datasource = defineAsyncComponent(
    () => import('@/views/visualized/data/datasource/index.vue')
  )
  const ScreenPanel = defineAsyncComponent(() => import('@/views/data-visualization/PreviewShow.vue'))
  const DashboardPanel = defineAsyncComponent(
    () => import('@/views/dashboard/DashboardPreviewShow.vue')
  )
  const TemplateManage = defineAsyncComponent(() => import('@/views/template/indexInject.vue'))

  const AsyncXpackComponent = defineAsyncComponent(() => import('@/components/plugin/src/index.vue'))

  const componentMap = {
    DashboardEditor,
    VisualizationEditor,
    ViewWrapper,
    Preview,
    Dashboard,
    Dataset,
    Datasource,
    ScreenPanel,
    DashboardPanel,
    TemplateManage
  }

  const iframeStyle = ref(null)
  const setStyle = debounce(() => {
    iframeStyle.value = {
      height: window.innerHeight + 'px',
      width: window.innerWidth + 'px'
    }
  }, 300)
  onBeforeMount(() => {
    window.addEventListener('resize', setStyle)
    setStyle()
  })
  onBeforeUnmount(() => {
    window.removeEventListener('resize', setStyle)
  })

  const showComponent = ref(false)
  const dataFillingPath = ref('')

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

        showComponent.value = false
        if (name && name.includes('DataFilling')) {
          if (name === 'DataFilling') {
            dataFillingPath.value = 'L21lbnUvZGF0YS9kYXRhLWZpbGxpbmcvbWFuYWdlL2Zvcm0vaW5kZXg='
          } else if (name === 'DataFillingEditor') {
            dataFillingPath.value = 'L21lbnUvZGF0YS9kYXRhLWZpbGxpbmcvbWFuYWdlL2Zvcm0vaW5kZXg='
          } else if (name === 'DataFillingHandler') {
            dataFillingPath.value = 'L21lbnUvZGF0YS9kYXRhLWZpbGxpbmcvbWFuYWdlL2Zvcm0vaW5kZXg='
          }
        }

        nextTick(() => {
          currentComponent.value = AsyncXpackComponent
          showComponent.value = true
        })
      }
    } else {
      nextTick(() => {
        currentComponent.value = componentMap[name || 'ViewWrapper']
        showComponent.value = true
      }
    }
  }

  // New bidirectional parameter handling
  const handleParamUpdate = (param: any, value: any) => {
    console.log('Bidirectional parameter update:', { param, value })

    // Update chart/dataset state via parent communication
    if (param && value !== undefined) {
      // Notify parent of parameter change
      const { resourceId, chartId } = embeddedStore
      if (resourceId) {
        embeddedStore.setResourceId(resourceId)
      }
      if (chartId) {
        embeddedStore.setChartId(chartId)
      }
    }
    }
  }

  // Initialize parent communication hook
  onMounted(() => {
    parentCommunication.initParentCommunication()

    // Listen for param updates from embedded
    parentCommunication.on('param_update', (param: any, value) => {
      handleParamUpdate(param, value)
    })

    // Listen for user interactions from embedded
    parentCommunication.on('user_interaction', (param: any, value) => {
      handleUserInteraction(param, value)
    })

    // Listen for ready events
    parentCommunication.on('ready', () => {
      parentCommunication.emitInitReady()
    })

    // Listen for errors
    parentCommunication.on('error', (error: string) => {
      parentCommunication.emitError(error)
    })
  })

  // Ready check
  watch: [ready] () => {
    if (parentCommunication.isReady()) {
      nextTick(() => {
        showComponent.value = false
      })
    }
  })
  </script>
```

### 3.2 Update `/views/data-visualization/PreviewCanvas.vue`

Enhance existing canvas communication with bidirectional parameter passing:

```vue
<script lang="ts" setup>
import { dvMainStoreWithOut } from '@/store/modules/data-visualization/dvMain'
import { eventBus } from '@/utils/eventBus'
import { listenGlobalKeyDown } from '@/utils/components/keyboard shortcuts'
import { useEmbedded } from '@/store/modules/embedded'
import { useParentCommunication } from '@/composables/useParentCommunication'

const embeddedStore = useEmbedded()
const parentCommunication = useParentCommunication()

const state = reactive({
  canvasCacheOutRef: null
  deCanvasRef = ref(null)
  canvasPreviewRef = ref(null)
})

// Initialize parent communication hook
onMounted(() => {
  parentCommunication.initParentCommunication()
})

// Listen for inner params updates from iframe or DIV
parentCommunication.on('param_update', (param: any, value) => {
  const { resourceId, dvId } = param
  if (resourceId) {
    embeddedStore.setResourceId(resourceId)
  }
  if (dvId) {
    embeddedStore.setDvId(dvId)
  }

  // Listen for parameter updates
parentCommunication.on('de_inner_params', (innerParams: Record<string, any>) => {
  const { resourceId, innerParams } = innerParams
  if (resourceId && resourceId === embeddedStore.resourceId) {
    dvMainStore.setNowTargetPanelJumpInfo({
      resourceId,
      dvId,
      target: 'canvas'
    })
  }

  // Listen for user interactions
parentCommunication.on('user_interaction', (param: any, value) => {
  const { param, value } = param
  // Handle user actions from embedded page
})
```

## Phase 4: Update /views/dashboard/DashboardPreviewShow.vue

Support DIV-based dashboard embedding with bidirectional parameter passing:

```vue
<script lang="ts" setup>
import { embeddedStore } from '@/store/modules/embedded'
import { isAllowedEmbeddedMessageOrigin, resolveEmbeddedOrigin } from '@/utils/embedded'
import { useParentCommunication } from '@/composables/useParentCommunication'

const embeddedStore = useEmbedded()
const parentCommunication = useParentCommunication()

// Existing iframe initialization
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

// New bidirectional parameter handling
const handleParamUpdate = (param: any, value: any) => {
  console.log('Bidirectional parameter update:', { param, value })

  // Update dashboard/dataset state via parent communication
  const { resourceId, chartId } = embeddedStore
  if (resourceId) {
    embeddedStore.setResourceId(resourceId)
  }
  if (chartId) {
    embeddedStore.setChartId(chartId)
  }
  }
}

// Initialize parent communication hook
onMounted(() => {
  parentCommunication.initParentCommunication()

  // Listen for param updates from DIV mode
parentCommunication.on('de_inner_params', (params: Record<string, any>) => {
  const { resourceId, innerParams } = params
  if (resourceId && resourceId === embeddedStore.resourceId) {
    // Update dashboard dataset state based on params
    dvMainStore.setNowTargetPanelJumpInfo({
      resourceId,
      dvId
    })
  }
})

// Listen for user interactions from DIV mode
parentCommunication.on('user_interaction', (param: any, value) => {
  const { param, value } = params
  // Handle user actions from embedded page
})
</script>
```

## Phase 5: Testing

### 5.1 Unit Tests

Test parent communication:

```typescript
// /tests/unit/parent-communication.spec.ts
describe('ParentCommunication', () => {
  let component = mountComponent()
  component.mounted = true
  expect(component.onMounted).toHaveBeenCalled() // Parent communication hook should be called
})
```

Test iframe communication:

```typescript
// /tests/unit/iframe-communication.spec.ts
describe('IframeCommunication', () => {
  const embeddedStore = mockEmbeddedStore()
  const iframe = mountIframeComponent()

  iframe.mounted = true

  await iframe.initIframe('test-token', 'http://test-origin')
  iframe.postMessage({ type: 'init_ready' }, '*')

  iframe.postMessage({ type: 'user_interaction', payload: { param: 'test-param', value: 'test-value' }, '*')

  // Verify iframe message was sent
})
})
```

Test DIV communication:

```typescript
// /tests/unit/div-communication.spec.ts
describe('DivCommunication', () => {
  const parentCommunication = mountDivComponent()

  parentCommunication.mounted = true

  parentCommunication.initIframe('test-token', 'http://test-origin')
  parentCommunication.postMessage({ type: 'module_init', payload: { mode: 'div' })

  // Verify init event was sent
  const initEvent = 'module_init'
})
```

Test module communication:

```typescript
// /tests/unit/module-communication.spec.ts
describe('ModuleCommunication', () => {
  const moduleCommunication = mountModuleComponent()

  moduleCommunication.mounted = true

  moduleCommunication.initIframe('test-token', 'http://test-origin')
  moduleCommunication.postMessage({ type: 'module_init', payload: { mode: 'module', resourceId: 'test-id' })
})
})
```

### 5.2 Integration Tests

Test bidirectional parameter passing:

```typescript
// /tests/integration/bidirectional-params.spec.ts
describe('BidirectionalParamPassing', () => {
  const parentCommunication = mountParentComponent()

  parentCommunication.mounted = true

  const resourceParams = {
    resourceId: 'test-resource-id',
    chartId: 'test-chart-id'
    dvId: 'test-dv-id',
    innerParams: {
      'test-param1': 'value1',
      'test-param2': 'value2'
    }
  }

  // Send param update
  parentCommunication.on('param_update', resourceParams)

  // Verify iframe message was sent
  const paramEvent = 'de_inner_params'
  const expectedEvents = ['param_update', 'ready']
})
})
```

## Phase 6: Documentation

### 6.1 Usage Guide

Add section to `/docs/embedded/bidirectional-params.md`:

```markdown
# Bidirectional Parameter Passing Guide

## Overview

This guide explains how to use bidirectional parameter passing for different embedding types.

## IFrame Embedding

### Token Management
Tokens are generated and validated via backend API and stored in `embeddedStore.token`.  
Origin allowlist is fetched via `embeddedInitIframeApi()` and stored in `embeddedStore.allowedOrigins`.

### Parent → Child Communication

Parent Component:
```vue
<template>
  <XpackComponent
    @init-iframe="initIframe"
    @bind:event-check="on('param_update')"
    @bind:event-check="on('user_interaction')"
    @bind:event-check="on('ready')"
    @bind:event-check="on('error')"
  />
</template>

<script setup>
  import { useEmbedded } from '@/store/modules/embedded'
  import { useParentCommunication } from '@/composables/useParentCommunication'

  const embeddedStore = useEmbedded()
  const parentCommunication = useParentCommunication()

  const emitParamUpdate = (params: Record<string, any>) => {
  parentCommunication.emit('param_update', params)
}

  onMounted(() => {
  parentCommunication.initParentCommunication()
})
</script>
```

### Child → Parent Communication

Child (iframe/DIV/module) emits parameters to parent:

```typescript
import { useEmbedded } from '@/store/modules/embedded'

const embeddedStore = useEmbedded()
const { resourceId, chartId } = embeddedStore

// Update parameters
const emitParamUpdate = (params: Record<string, any>) => {
  const payload = {
    resourceId: embeddedStore.resourceId,
    chartId: embeddedStore.chartId,
    innerParams: params
  }

parentCommunication.on('de_inner_params', emitParamUpdate)
```

## DIV Embedding

Parent Component (`/pages/panel/ViewWrapper.vue`):

```typescript
const embeddedStore = useEmbedded()
const { resourceId, chartId, busiFlag } = embeddedStore

const emitParamUpdate = (params: Record<string, any>) => {
  parentCommunication.emit('de_inner_params', emitParamUpdate)
}

parentCommunication.on('de_inner_params', emitParamUpdate)
```

## Module-Level Page Embedding

Parent Component (`/pages/panel/DashboardPreview.vue`):

```typescript
const { resourceId, chartId } = embeddedStore

const emitParamUpdate = (params: Record<string, any>) => {
  const payload = {
    resourceId,
    chartId,
    params
  }
}

parentCommunication.on('de_inner_params', emitParamUpdate)
```

## Event Types

| Type | Name | Purpose | Parameters | Handler |
|-----------|------|---------|----------|
| `init_ready` | Canvas init | resourceId, dvId | `parentCommunication.emitInitReady()` |
| `de_inner_params` | Inner params update | resourceId, chartId, innerParams | `parentCommunication.on('de_inner_params', emitParamUpdate()` |
| `user_interaction` | User action from embedded | param, value | `parentCommunication.on('user_interaction', emitParamUpdate()` |
| `ready` | Component ready | `parentCommunication.emitReady()` |
| `error` | Error handling | `parentCommunication.emitError()` |
| `module_init` | Module init | mode, resourceId, busiFlag | `module_init_ready` | `parentCommunication.on('module_init')` |

## Token Refresh (Future Enhancement)

Currently: Tokens are static (no refresh mechanism).
Future: Implement token refresh endpoint and interval-based token renewal.
Usage: In `embeddedStore`, setToken(newToken), `setExpiry(expiryTime)`.
```typescript
// Example
await embeddedStore.setToken(newToken)
await embeddedStore.setExpiry(Date.now() + 3600000) // 1 hour
```

## Open Questions

1. Do we need to create a unified `useBidirectionalCommunication()` composable?
2. Should DIV embedding have separate `useDivParentCommunication()` composable?
3. Should we create token refresh service?
4. How to handle token expiry in child components?
