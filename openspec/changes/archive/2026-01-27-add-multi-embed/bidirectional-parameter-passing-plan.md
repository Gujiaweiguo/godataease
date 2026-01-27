# Bidirectional Parameter Passing - Implementation Plan

## Overview

This document provides a detailed implementation plan for standardizing bidirectional parameter passing for dashboard/screen/chart/module-level page embedding.

## Current State

### Working Components
1. **Dashboard**:
   - `/views/dashboard/index.vue` - Uses `embeddedInitIframeApi()` and `useEmbedded()` store
   - `/views/dashboard/DashboardPreviewShow.vue` - Same pattern

2. **Screen/DataV**:
   - `/views/data-visualization/index.vue` - Designer editor with `embeddedInitIframeApi()`
   - `/views/data-visualization/PreviewCanvas.vue` - Screen preview with `embeddedInitIframeApi()`
   - `/views/data-visualization/PreviewShow.vue` - Screen show with `embeddedInitIframeApi()`

3. **Chart**:
   - `/views/chart/ChartView.vue` - Entry point for chart embedding
   - Uses `embeddedInitIframeApi()` and `embeddedStore.setAllowedOrigins()`

4. **Dataset/Datasource**:
   - `/views/visualized/data/dataset/index.vue` - Uses `embeddedInitIframeApi()` for iframe-based
   - `/views/visualized/data/dataset/form/index.vue` - Form-based entry

5. **Panel/Mobile**:
   - `/mobile/panel/index.vue` - Panel config with iframe communication
   - Uses `eventBus` for panel/screen-weight` events

## Problem Statement

The current bidirectional parameter passing implementation is fragmented:
- Multiple postMessage patterns without clear standardization
- Different payload structures across events
- Inconsistent error handling
- No unified interface for parent-child communication
- No centralized validation of parameters

## Solution Design

### Phase 1: Define Standardized Event Payloads

#### 1.1. Event Type Registry
```typescript
// /events/embedding/types.ts
export enum EmbeddingEventType {
  // Parameter update events
  PARAM_UPDATE = 'param_update',
  
  // Interaction events
  INTERACTION = 'user_interaction',
  
  // Lifecycle events
  INIT_READY = 'canvas_init_ready',
  ERROR = 'error',
  READY = 'ready',
  DE_INIT = 'de_inner_params',
  CANVAS_INIT = 'canvas_init_ready',
  JUMP_TO_TARGET = 'jump_to_target',
  ATTACH_PARAMS = 'attach_params'
}
```

#### 1.2. Standardized Payload Interfaces
```typescript
// /events/embedding/payloads.ts
interface InitReadyPayload {
  resourceId: string
}

interface ParamUpdatePayload {
  resourceId: string
  [key: string]: any
}

interface UserInteractionPayload {
  param: string
  value: any
}

interface ErrorPayload {
  message: string
}

interface ReadyPayload {
  component?: string
}

interface DeInnerParamsPayload {
  [key: string]: any
}
```

### Phase 2: Implement Parent Communication Hook

#### 2.1. Parent Hook Interface
```typescript
// /composables/useEmbeddedParentCommunication.ts
interface UseEmbeddedParentCommunication {
  postMessage(type: EmbeddingEventType, payload: InitReadyPayload | ParamUpdatePayload | UserInteractionPayload | ErrorPayload | ReadyPayload | DeInnerParamsPayload | DeInitParamsPayload): void
  
  listenForChildMessages(): void
  emitToChild(type: EmbeddingEventType, payload: InitReadyPayload | ParamUpdatePayload | UserInteractionPayload | ErrorPayload | ReadyPayload | DeInnerParamsPayload | DeInitParamsPayload): void
}
```

#### 2.2. Implementation
```typescript
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { useEmbedded } from '@/store/modules/embedded'

export function useEmbeddedParentCommunication() {
  const store = useEmbedded()
  
  const listenForChildMessages = () => {
    window.addEventListener('message', (event: MessageEvent) => {
      try {
        const parsed = JSON.parse(event.data)
        const type = parsed.type as EmbeddingEventType
        
        switch (type) {
          case 'param_update':
            if (parsed.resourceId === store.resourceId) {
              store.setResourceId(parsed.resourceId)
            }
            break
            
          case 'user_interaction':
            if (parsed.resourceId === store.resourceId) {
              const { param, value } = parsed
              store.setParam(param, value)
            }
            break
            
          case 'init_ready':
            // Canvas/Panel/ScreenPanel ready for initialization
            console.log('Parent received init_ready event')
            break
            
          case 'de_inner_params':
            // DIV-based embed mode - nested iframe receives inner params
            const innerParams = parsed
            if (parsed.dvId === store.dvId) {
              // Update local state with inner params
              // This needs to be expanded based on specific embed type
            }
            break
            
          case 'ready':
            store.setEmbedReady(true)
            break
            
          case 'de_init':
            // Component initialization complete
            console.log('Parent received de_init event')
            break
            
          case 'canvas_init_ready':
            // Canvas initialized, ready for user interaction
            break
            
          case 'jump_to_target':
            if (parsed.chartId === store.chartId && parsed.dvId === store.dvId) {
              const { jumpInfoParam } = parsed
              store.setJumpInfoParam(jumpInfoParam)
            }
            break
            
          default:
            console.warn('Unknown event type:', type)
        }
      } catch (error) {
        console.error('Failed to parse parent message:', error)
        // Handle error gracefully
      }
    }
  }
  
  const emitToChild = (type: EmbeddingEventType, payload: any) => {
    if (store.parent && window.parent !== window.top) {
      const targetPm = 'dataease-embedded-host' + (payload as string)
      window.parent.postMessage(targetPm, '*')
    }
  }
  
  return {
    listenForChildMessages,
    emitToChild
  }
}
```

### Phase 3: Update Components to Use Standardized Communication

#### 3.1. Update `/views/data-visualization/PreviewCanvas.vue`
Replace existing `postMessage` event handling with standardized event payloads

#### 3.2. Update `/views/dashboard/index.vue`
Replace iframe initialization with standard hook

#### 3.3. Update `/views/chart/ChartView.vue`
Use `useEmbeddedParentCommunication()` instead of raw event handling

#### 3.4. Update `/views/data-visualization/index.vue`
Same as PreviewCanvas.vue

#### 3.5. Update `/views/mobile/panel/index.vue`
Consolidate panel/screen-weight event handling

## Open Questions

1. Should we create a `/composables/useEmbeddedChildCommunication.ts` composable for child-side communication?
2. Should we create shared interfaces for parameter schemas that are currently ad-hoc?
3. Do dashboard/screen/chart components need to emit specific events beyond what's currently supported?
4. Should we create separate event handlers for different embed types (Dashboard vs DataV vs Dataset)?

## Success Criteria

- [ ] All components use `useEmbeddedParentCommunication()` for parent communication
- [ ] Parent hook initialized with `listenForChildMessages()`
- [ ] `emitToChild()` calls use standardized event types from registry
- [ ] Event payloads match defined TypeScript interfaces
- [ ] Error handling is centralized in parent hook
- [ ] Parent-child communication is consistent across all embed types
