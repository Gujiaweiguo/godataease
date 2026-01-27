# Phase 2.1: Testing - Implementation Plan

## Overview

This document details the testing strategy for verifying the bidirectional parameter passing implementation in `add-multi-embed`.

## Testing Scope

### What to Test

1. **Event Type Registry**
   - Verify all 13 event types are correctly defined
   - Test event type validation function
   - Verify event direction mapping

2. **Payload Interfaces**
   - Verify TypeScript interfaces match expected structures
   - Test payload validation function
   - Test union types and utility types

3. **Parent Communication Composable**
   - Test `listenForChildMessages()` with valid/invalid origins
   - Test `emitToChild<T>()` with all event types
   - Test `postLegacyMessage()` for backward compatibility
   - Verify store integration (setResourceId, setParam, etc.)

4. **Component Integration**
   - Test PreviewCanvas.vue standardized event emission
   - Test DashboardPreviewShow.vue standardized event emission
   - Test dashboard/index.vue standardized event emission
   - Test data-visualization/index.vue standardized event emission
   - Verify backward compatibility (both new and legacy events sent)

## Test Cases

### Unit Tests

#### Test 1: Event Type Validation

```typescript
// Test: isValidEmbeddingEventType
import { isValidEmbeddingEventType, EmbeddingEventType } from '@/events/embedding/types'

test('validates known event types', () => {
  expect(isValidEmbeddingEventType('param_update')).toBe(true)
  expect(isValidEmbeddingEventType('user_interaction')).toBe(true)
  expect(isValidEmbeddingEventType('init_ready')).toBe(true)
  expect(isValidEmbeddingEventType('ready')).toBe(true)
  expect(isValidEmbeddingEventType('error')).toBe(true)
  expect(isValidEmbeddingEventType('de_init')).toBe(true)
  expect(isValidEmbeddingEventType('canvas_init')).toBe(true)
  expect(isValidEmbeddingEventType('attach_params')).toBe(true)
  expect(isValidEmbeddingEventType('jump_to_target')).toBe(true)
  expect(isValidEmbeddingEventType('module_init')).toBe(true)
  expect(isValidEmbeddingEventType('module_update')).toBe(true)
  expect(isValidEmbeddingEventType('module_interaction')).toBe(true)
})

test('rejects unknown event types', () => {
  expect(isValidEmbeddingEventType('unknown_event')).toBe(false)
  expect(isValidEmbeddingEventType('invalid')).toBe(false)
})
```

#### Test 2: Payload Validation

```typescript
import { validatePayload, EmbeddingEventType } from '@/events/embedding/payloads'

test('validates InitReadyPayload', () => {
  const payload = { resourceId: 'test-id' }
  expect(validatePayload(EmbeddingEventType.INIT_READY, payload)).toBe(true)
})

test('validates ErrorPayload', () => {
  const payload = { message: 'Test error' }
  expect(validatePayload(EmbeddingEventType.ERROR, payload)).toBe(true)
})

test('requires resourceId for InitReady', () => {
  const payload = {}
  expect(validatePayload(EmbeddingEventType.INIT_READY, payload)).toBe(false)
})

test('requires message for Error', () => {
  const payload = {}
  expect(validatePayload(EmbeddingEventType.ERROR, payload)).toBe(false)
})
```

#### Test 3: Composable Integration

```typescript
import { useEmbeddedParentCommunication } from '@/hooks/event/useEmbeddedParentCommunication'
import { useEmbedded } from '@/store/modules/embedded'

test('composable initializes message listener', () => {
  const { listenForChildMessages } = useEmbeddedParentCommunication()
  const addEventListenerSpy = vi.spyOn(window, 'addEventListener')
  listenForChildMessages()
  expect(addEventListenerSpy).toHaveBeenCalledWith('message', expect.any(Function), expect.any(Object))
})

test('composable validates origins', () => {
  const { listenForChildMessages } = useEmbeddedParentCommunication()
  const store = useEmbedded()
  store.setAllowedOrigins(['https://trusted-origin.com'])
  
  // Test with trusted origin
  const trustedEvent = new MessageEvent('message', { data: JSON.stringify({ type: 'param_update' }), origin: 'https://trusted-origin.com' })
  // Handler should accept this event
  
  // Test with untrusted origin
  const untrustedEvent = new MessageEvent('message', { data: JSON.stringify({ type: 'param_update' }), origin: 'https://untrusted-origin.com' })
  // Handler should reject this event
})
```

### Integration Tests

#### Test 4: Parent-Child Communication Flow

**Scenario: Parent frame receives init_ready event**

1. Parent loads embedded content in iframe
2. Child component (PreviewCanvas) sends INIT_READY event
3. Parent receives and parses event
4. Parent calls `handleInitReady()` handler
5. Verify `store.setResourceId()` is called
6. Verify parent can send response using `emitToChild()`

**Expected Result:**
- Event is successfully parsed
- Store is updated
- Parent can emit response events
- No errors in console

#### Test 5: Backward Compatibility

**Scenario: Legacy parent frame receives old event format**

1. Child component sends legacy event: `onInitReady({ resourceId: 'test' })`
2. Parent receives old format
3. Child also sends new event: `emitToChild(INIT_READY, payload)`
4. Verify both events work correctly

**Expected Result:**
- Legacy event still works (backward compatibility maintained)
- New standardized event also works
- Parent can handle both formats

#### Test 6: Event Type Registry Coverage

**Scenario: All 13 event types can be emitted and handled**

For each `EmbeddingEventType`:
1. Create payload matching event type
2. Call `emitToChild(eventType, payload)`
3. Verify message structure is correct
4. Verify parent can parse message

**Event Types to Test:**
1. PARAM_UPDATE
2. INTERACTION
3. INIT_READY
4. READY
5. ERROR
6. DE_INIT
7. CANVAS_INIT
8. ATTACH_PARAMS
9. JUMP_TO_TARGET
10. MODULE_INIT
11. MODULE_UPDATE
12. MODULE_INTERACTION

## Test Files Structure

```
core/core-frontend/tests/unit/events/
├── embedding/
│   ├── types.test.ts
│   ├── payloads.test.ts
│   └── useEmbeddedParentCommunication.test.ts
├── integration/
│   └── parent-child-communication.test.ts
```

## Success Criteria

- [ ] All unit tests pass
- [ ] All integration tests pass
- [ ] Backward compatibility verified
- [ ] Code coverage for new files > 80%
- [ ] No new linting errors
- [ ] TypeScript compilation succeeds

## Dependencies

### Required
- Vitest or Jest (check project configuration)
- Test utilities (vi.fn, vi.spyOn)
- Component testing utilities (mount components)

### Optional
- MSW or similar for mocking postMessage in tests
- Test storybook or similar for visual verification

## Tasks

### Unit Tests
- [ ] Create test file: types.test.ts
- [ ] Create test file: payloads.test.ts
- [ ] Create test file: useEmbeddedParentCommunication.test.ts
- [ ] Run unit tests and verify all pass

### Integration Tests
- [ ] Create test file: parent-child-communication.test.ts
- [ ] Run integration tests and verify all pass

### Verification
- [ ] Verify backward compatibility with legacy event format
- [ ] Check code coverage for new infrastructure
- [ ] Fix any failing tests
- [ ] Document test results

## Notes

- Tests should focus on verifying the new standardized infrastructure
- Backward compatibility is critical - must not break existing integrations
- Integration tests can use component mounting with test utilities
- Consider mocking window.parent.postMessage in tests to avoid cross-origin issues
