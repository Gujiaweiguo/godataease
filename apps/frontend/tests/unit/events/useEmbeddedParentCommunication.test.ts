import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { useEmbeddedParentCommunication } from '@/hooks/event/useEmbeddedParentCommunication'
import { EmbeddingEventType } from '@/events/embedding/types'
import { isAllowedEmbeddedMessageOrigin } from '@/utils/embedded'

const hoisted = vi.hoisted(() => ({
  mockStore: {
    allowedOrigins: ['https://trusted-origin.com'],
    parent: false,
    resourceId: 'test-id',
    dvId: '',
    chartId: '',
    baseUrl: '',
    setParam: vi.fn(),
    setResourceId: vi.fn(),
    setEmbedReady: vi.fn(),
    setJumpInfoParam: vi.fn()
  }
}))

vi.mock('@/utils/embedded', async () => {
  const { createEmbeddedUtilsModuleMock } = await import('../helpers')
  return createEmbeddedUtilsModuleMock(true)
})

vi.mock('@/store/modules/embedded', async () => {
  const { createEmbeddedModuleMock } = await import('../helpers')
  return createEmbeddedModuleMock(hoisted.mockStore)
})

describe('useEmbeddedParentCommunication', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.spyOn(console, 'warn').mockImplementation(() => undefined)
    vi.spyOn(console, 'error').mockImplementation(() => undefined)
    vi.spyOn(console, 'log').mockImplementation(() => undefined)
    vi.spyOn(console, 'debug').mockImplementation(() => undefined)
    hoisted.mockStore.parent = false
    hoisted.mockStore.resourceId = 'test-id'
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('registers message listener', () => {
    const { listenForChildMessages } = useEmbeddedParentCommunication()
    const addEventListenerSpy = vi.spyOn(window, 'addEventListener')

    listenForChildMessages()

    expect(addEventListenerSpy).toHaveBeenCalledWith('message', expect.any(Function))
  })

  it('blocks untrusted message origin', () => {
    vi.mocked(isAllowedEmbeddedMessageOrigin).mockReturnValue(false)
    const consoleWarnSpy = vi.spyOn(console, 'warn').mockImplementation(() => undefined)
    const { listenForChildMessages } = useEmbeddedParentCommunication()

    listenForChildMessages()
    window.dispatchEvent(
      new MessageEvent('message', {
        origin: 'https://bad-origin.com',
        data: JSON.stringify({ type: 'param_update', payload: { resourceId: 'test-id', a: 1 } })
      })
    )

    expect(consoleWarnSpy).toHaveBeenCalledWith(
      'Message from untrusted origin blocked: https://bad-origin.com'
    )
  })

  it('handles param_update and writes params to store', () => {
    vi.mocked(isAllowedEmbeddedMessageOrigin).mockReturnValue(true)
    const { listenForChildMessages } = useEmbeddedParentCommunication()

    listenForChildMessages()
    window.dispatchEvent(
      new MessageEvent('message', {
        origin: 'https://trusted-origin.com',
        data: JSON.stringify({
          type: 'param_update',
          payload: { resourceId: 'test-id', region: '华东', timestamp: Date.now() }
        })
      })
    )

    expect(hoisted.mockStore.setResourceId).toHaveBeenCalledWith('test-id')
    expect(hoisted.mockStore.setParam).toHaveBeenCalledWith('region', '华东')
  })

  it('skips emit when not embedded mode', () => {
    const consoleWarnSpy = vi.spyOn(console, 'warn').mockImplementation(() => undefined)
    const postMessageSpy = vi.spyOn(window.parent, 'postMessage')
    const { emitToChild } = useEmbeddedParentCommunication()

    emitToChild(EmbeddingEventType.INIT_READY, { resourceId: 'test-id' })

    expect(postMessageSpy).not.toHaveBeenCalled()
    expect(consoleWarnSpy).toHaveBeenCalledWith('Not in embedded mode, skipping emit to child')
  })
})
