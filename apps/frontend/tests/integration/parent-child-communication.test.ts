import { describe, it, expect, vi, beforeEach } from 'vitest'
import { defineComponent, h, nextTick } from 'vue'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { useEmbeddedParentCommunication } from '@/hooks/event/useEmbeddedParentCommunication'
import { EmbeddingEventType } from '@/events/embedding/types'

const hoisted = vi.hoisted(() => ({
  mockStore: {
    allowedOrigins: ['https://trusted-origin.com'],
    parent: true,
    resourceId: 'integration-resource',
    dvId: 'dv-1',
    chartId: 'chart-1',
    baseUrl: '',
    setParam: vi.fn(),
    setResourceId: vi.fn(),
    setEmbedReady: vi.fn(),
    setJumpInfoParam: vi.fn()
  }
}))

vi.mock('@/utils/embedded', async () => {
  const { createEmbeddedUtilsModuleMock } = await import('../unit/helpers')
  return createEmbeddedUtilsModuleMock(true)
})

vi.mock('@/store/modules/embedded', async () => {
  const { createEmbeddedModuleMock } = await import('../unit/helpers')
  return createEmbeddedModuleMock(hoisted.mockStore)
})

const Harness = defineComponent({
  setup() {
    const { emitToChild } = useEmbeddedParentCommunication()
    return () =>
      h('div', {
        id: 'harness',
        onClick: () =>
          emitToChild(EmbeddingEventType.INIT_READY, { resourceId: 'integration-resource' })
      })
  }
})

describe('Parent-Child Communication Integration', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    setActivePinia(createPinia())
    hoisted.mockStore.parent = true
    hoisted.mockStore.resourceId = 'integration-resource'
  })

  it('processes trusted param_update message and syncs store', async () => {
    mount(Harness)
    window.dispatchEvent(
      new MessageEvent('message', {
        origin: 'https://trusted-origin.com',
        data: JSON.stringify({
          type: 'param_update',
          payload: {
            resourceId: 'integration-resource',
            region: '华北'
          }
        })
      })
    )

    await nextTick()
    expect(hoisted.mockStore.setResourceId).toHaveBeenCalledWith('integration-resource')
    expect(hoisted.mockStore.setParam).toHaveBeenCalledWith('region', '华北')
  })

  it('skips emit when not in iframe embedded context', async () => {
    const consoleWarnSpy = vi.spyOn(console, 'warn').mockImplementation(() => undefined)
    const postMessageSpy = vi.spyOn(window.parent, 'postMessage')
    const wrapper = mount(Harness)

    await wrapper.trigger('click')

    expect(postMessageSpy).not.toHaveBeenCalled()
    expect(consoleWarnSpy).toHaveBeenCalledWith('Not in embedded mode, skipping emit to child')
  })
})
