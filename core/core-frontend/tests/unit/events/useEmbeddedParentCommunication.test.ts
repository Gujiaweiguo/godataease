import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { useEmbeddedParentCommunication } from '@/hooks/event/useEmbeddedParentCommunication'
import { useEmbedded } from '@/store/modules/embedded'
import { resolveEmbeddedOrigin } from '@/utils/embedded'
import type { InitReadyPayload } from '@/events/embedding/payloads'
import { EmbeddingEventType } from '@/events/embedding/types'

vi.mock('@/utils/embedded', () => ({
  resolveEmbeddedOrigin: vi.fn(() => 'https://test-origin.com'),
  isAllowedEmbeddedMessageOrigin: vi.fn(() => true)
}))

vi.mock('@/store/modules/embedded', () => ({
  useEmbedded: vi.fn(() => ({
    getToken: 'test-token',
    getAllowedOrigins: () => ['https://test-origin.com'],
    parent: true
  }))
}))

describe('useEmbeddedParentCommunication', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  describe('listenForChildMessages', () => {
    it('should initialize message listener on mount', () => {
      const { listenForChildMessages } = useEmbeddedParentCommunication()
      const addEventListenerSpy = vi.spyOn(window, 'addEventListener')

      listenForChildMessages()

      expect(addEventListenerSpy).toHaveBeenCalledWith('message', expect.any(Function), expect.any(Object))
    })

    it('should validate message origin before processing', () => {
      const { listenForChildMessages } = useEmbeddedParentCommunication()
      const { isAllowedEmbeddedMessageOrigin } = require('@/utils/embedded')
      const mockEvent = new MessageEvent('message', {
        data: JSON.stringify({ type: 'param_update' }),
        origin: 'https://trusted-origin.com'
      })

      window.dispatchEvent(mockEvent)

      // Should process the message since origin is allowed
      expect(isAllowedEmbeddedMessageOrigin).toHaveBeenCalledWith(
        'https://trusted-origin.com',
        ['https://trusted-origin.com'],
        true
      )
    })

    it('should reject message from untrusted origin', () => {
      const { listenForChildMessages } = useEmbeddedParentCommunication()
      const { isAllowedEmbeddedMessageOrigin } = require('@/utils/embedded')
      const consoleWarnSpy = vi.spyOn(console, 'warn')
      const mockEvent = new MessageEvent('message', {
        data: JSON.stringify({ type: 'param_update' }),
        origin: 'https://untrusted-origin.com'
      })

      window.dispatchEvent(mockEvent)

      // Should reject the message
      expect(consoleWarnSpy).toHaveBeenCalledWith('Message from untrusted origin blocked: https://untrusted-origin.com')
    })
  })

  describe('emitToChild', () => {
    it('should emit event to parent window', () => {
      const { emitToChild } = useEmbeddedParentCommunication()
      const postMessageSpy = vi.spyOn(window.parent, 'postMessage')

      emitToChild(EmbeddingEventType.INIT_READY, { resourceId: 'test' })

      const expectedMessage = 'dataease-embedded-host:{"type":"init_ready","payload":{"resourceId":"test"}}'
      expect(postMessageSpy).toHaveBeenCalledWith(expectedMessage, '*')
    })

    it('should skip emit if not in embedded mode', () => {
      const { useEmbedded: mockUseEmbedded } = require('@/store/modules/embedded')
      mockUseEmbedded.mockReturnValue({ getToken: '', parent: false })

      const { emitToChild } = useEmbeddedParentCommunication()
      const postMessageSpy = vi.spyOn(window.parent, 'postMessage')

      emitToChild(EmbeddingEventType.INIT_READY, { resourceId: 'test' })

      expect(postMessageSpy).not.toHaveBeenCalled()
    })
  })

  describe('event handlers', () => {
    it('should handle param_update event and update store', () => {
      const { listenForChildMessages } = useEmbeddedParentCommunication()
      const mockSetParam = vi.fn()

      const mockEvent = new MessageEvent('message', {
        data: JSON.stringify({ type: 'param_update', resourceId: 'test-id', param1: 'value1' })
      })

      window.dispatchEvent(mockEvent)

      // Check if setParam was called
      expect(mockSetParam).toHaveBeenCalledWith('param1', 'value1')
    })

    it('should handle user_interaction event and update store', () => {
      const { listenForChildMessages } = useEmbeddedParentCommunication()
      const mockSetParam = vi.fn()

      const mockEvent = new MessageEvent('message', {
        data: JSON.stringify({ type: 'user_interaction', param: 'test-param', value: 'test-value' })
      })

      window.dispatchEvent(mockEvent)

      expect(mockSetParam).toHaveBeenCalledWith('test-param', 'test-value')
    })

    it('should handle init_ready event', () => {
      const { listenForChildMessages } = useEmbeddedParentCommunication()
      const consoleLogSpy = vi.spyOn(console, 'log')

      const mockEvent = new MessageEvent('message', {
        data: JSON.stringify({ type: 'init_ready', resourceId: 'test-id' })
      })

      window.dispatchEvent(mockEvent)

      expect(consoleLogSpy).toHaveBeenCalledWith('Parent received init_ready event')
    })

    it('should handle error event', () => {
      const { listenForChildMessages } = useEmbeddedParentCommunication()
      const consoleErrorSpy = vi.spyOn(console, 'error')

      const mockEvent = new MessageEvent('message', {
        data: JSON.stringify({ type: 'error', message: 'Test error', context: 'test-scenario' })
      })

      window.dispatchEvent(mockEvent)

      expect(consoleErrorSpy).toHaveBeenCalledWith('Child frame error: Test error', 'test-scenario')
    })
  })
})
