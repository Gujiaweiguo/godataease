import { describe, it, expect } from 'vitest'
import { isValidEmbeddingEventType, EmbeddingEventType } from '@/events/embedding/types'

describe('EmbeddingEventType', () => {
  describe('isValidEmbeddingEventType', () => {
    it('should return true for known event types', () => {
      const knownTypes: string[] = Object.values(EmbeddingEventType)
      knownTypes.forEach(type => {
        expect(isValidEmbeddingEventType(type)).toBe(true)
      })
    })

    it('should return false for unknown event types', () => {
      expect(isValidEmbeddingEventType('unknown_event')).toBe(false)
      expect(isValidEmbeddingEventType('')).toBe(false)
      expect(isValidEmbeddingEventType('invalid')).toBe(false)
      expect(isValidEmbeddingEventType(undefined as any)).toBe(false)
    })

    it('should be case-sensitive', () => {
      expect(isValidEmbeddingEventType('PARAM_UPDATE')).toBe(false)
      expect(isValidEmbeddingEventType('param_update')).toBe(true)
    })
  })

  describe('EmbeddingEventType values', () => {
    it('should contain all required event types', () => {
      expect(EmbeddingEventType.PARAM_UPDATE).toBe('param_update')
      expect(EmbeddingEventType.INTERACTION).toBe('user_interaction')
      expect(EmbeddingEventType.INIT_READY).toBe('init_ready')
      expect(EmbeddingEventType.READY).toBe('ready')
      expect(EmbeddingEventType.ERROR).toBe('error')
      expect(EmbeddingEventType.DE_INIT).toBe('de_init')
      expect(EmbeddingEventType.CANVAS_INIT).toBe('canvas_init')
      expect(EmbeddingEventType.ATTACH_PARAMS).toBe('attach_params')
      expect(EmbeddingEventType.JUMP_TO_TARGET).toBe('jump_to_target')
      expect(EmbeddingEventType.MODULE_INIT).toBe('module_init')
      expect(EmbeddingEventType.MODULE_UPDATE).toBe('module_update')
      expect(EmbeddingEventType.MODULE_INTERACTION).toBe('module_interaction')
    })
  })

