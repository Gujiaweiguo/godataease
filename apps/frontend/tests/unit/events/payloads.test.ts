import { describe, it, expect } from 'vitest'
import { validatePayload, EmbeddingEventType } from '@/events/embedding/payloads'
import type { InitReadyPayload, ErrorPayload } from '@/events/embedding/payloads'

describe('validatePayload', () => {
  describe('InitReadyPayload', () => {
    it('should validate correct InitReadyPayload', () => {
      const payload: InitReadyPayload = {
        resourceId: 'test-id'
      }
      expect(validatePayload(EmbeddingEventType.INIT_READY, payload)).toBe(true)
    })

    it('should require resourceId for InitReadyPayload', () => {
      const payload = {}
      expect(validatePayload(EmbeddingEventType.INIT_READY, payload)).toBe(false)
    })

    it('should allow optional timestamp in InitReadyPayload', () => {
      const payload: InitReadyPayload = {
        resourceId: 'test-id',
        timestamp: Date.now()
      }
      expect(validatePayload(EmbeddingEventType.INIT_READY, payload)).toBe(true)
    })
  })

  describe('ErrorPayload', () => {
    it('should validate correct ErrorPayload', () => {
      const payload: ErrorPayload = {
        message: 'Test error',
        code: 'TEST_001',
        context: 'test-scenario'
      }
      expect(validatePayload(EmbeddingEventType.ERROR, payload)).toBe(true)
    })

    it('should require message for ErrorPayload', () => {
      const payload = {}
      expect(validatePayload(EmbeddingEventType.ERROR, payload)).toBe(false)
    })

    it('should allow optional fields in ErrorPayload', () => {
      const payload: ErrorPayload = {
        message: 'Test error',
        details: { errorInfo: 'test' },
        context: 'test-scenario'
      }
      expect(validatePayload(EmbeddingEventType.ERROR, payload)).toBe(true)
    })
  })

  describe('ParamUpdatePayload', () => {
    it('should validate correct ParamUpdatePayload', () => {
      const payload = {
        resourceId: 'test-id',
        param1: 'value1',
        param2: 'value2'
      }
      expect(validatePayload(EmbeddingEventType.PARAM_UPDATE, payload)).toBe(true)
    })

    it('should allow dynamic parameters in ParamUpdatePayload', () => {
      const payload = {
        resourceId: 'test-id',
        customField: 'custom-value'
      }
      expect(validatePayload(EmbeddingEventType.PARAM_UPDATE, payload)).toBe(true)
    })
  })

  describe('InteractionPayload', () => {
    it('should validate correct InteractionPayload', () => {
      const payload = {
        interactionType: 'click',
        param: 'test-param',
        value: 'test-value'
      }
      expect(validatePayload(EmbeddingEventType.INTERACTION, payload)).toBe(true)
    })

    it('should allow optional fields in InteractionPayload', () => {
      const payload = {
        interactionType: 'click'
        metadata: { source: 'user' }
      }
      expect(validatePayload(EmbeddingEventType.INTERACTION, payload)).toBe(true)
    })
  })
})
