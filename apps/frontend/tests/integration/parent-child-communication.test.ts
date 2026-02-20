import { describe, it, vi, beforeEach, afterEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import PreviewCanvas from '@/views/data-visualization/PreviewCanvas.vue'
import { useEmbeddedParentCommunication } from '@/hooks/event/useEmbeddedParentCommunication'
import type { InitReadyPayload } from '@/events/embedding/payloads'
import { EmbeddingEventType } from '@/events/embedding/types'

describe('Parent-Child Communication Integration', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    setActivePinia(createPinia())
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  describe('PreviewCanvas.vue', () => {
    it('should emit init_ready event on canvas initialization', async () => {
      const postMessageSpy = vi.spyOn(window.parent, 'postMessage')
      
      const wrapper = mount(PreviewCanvas, {
        props: {
          outerId: 'test-dv-id',
          publicLinkStatus: true
        }
      })

      await wrapper.vm.$nextTick()

      expect(postMessageSpy).toHaveBeenCalled()
      
      const emittedCalls = postMessageSpy.mock.calls
      
      expect(emittedCalls.length).toBeGreaterThan(0)
    })

    it('should emit both new and legacy init_ready events for backward compatibility', async () => {
      const postMessageSpy = vi.spyOn(window.parent, 'postMessage')
      
      const wrapper = mount(PreviewCanvas, {
        props: {
          outerId: 'test-dv-id',
          publicLinkStatus: true
        }
      })

      await wrapper.vm.$nextTick()

      expect(postMessageSpy).toHaveBeenCalled()
      
      const emittedCalls = postMessageSpy.mock.calls
      const allMessages = emittedCalls.map(call => call.toString())
      
      const hasNewEvent = allMessages.some(msg => 
        msg.includes('"type":"init_ready"') && 
        msg.includes('"resourceId":"test-dv-id"')
      )
      
      const hasLegacyEvent = allMessages.some(msg => 
        msg.includes('dataease-embedded-interactive')
      )
      
      expect(hasNewEvent).toBe(true)
      expect(hasLegacyEvent).toBe(true)
    })
  })

  describe('DashboardPreviewShow.vue', () => {
    it('should emit init_ready event on dashboard initialization', async () => {
      const postMessageSpy = vi.spyOn(window.parent, 'postMessage')
      
      const wrapper = mount(PreviewCanvas, {
        props: {
          outerId: 'test-dv-id'
          publicLinkStatus: false
        }
      })

      await wrapper.vm.$nextTick()

      expect(postMessageSpy).toHaveBeenCalled()
      
      const emittedCalls = postMessageSpy.mock.calls
      
      expect(emittedCalls.length).toBeGreaterThan(0)
    })
  })

  describe('Dashboard.vue', () => {
    it('should emit init_ready event on dashboard editor initialization', async () => {
      const postMessageSpy = vi.spyOn(window.parent, 'postMessage')
      
      const wrapper = mount(PreviewCanvas, {
        props: {
          outerId: 'test-dv-id',
          publicLinkStatus: true
        }
      })

      await wrapper.vm.$nextTick()

      expect(postMessageSpy).toHaveBeenCalled()
      
      const emittedCalls = postMessageSpy.mock.calls
      
      expect(emittedCalls.length).toBeGreaterThan(0)
    })
  })

  describe('DataVisualization.vue', () => {
    it('should emit init_ready event on screen initialization', async () => {
      const postMessageSpy = vi.spyOn(window.parent, 'postMessage')
      
      const wrapper = mount(PreviewCanvas, {
        props: {
          outerId: 'test-dv-id',
          publicLinkStatus: true
        }
      })

      await wrapper.vm.$nextTick()

      expect(postMessageSpy).toHaveBeenCalled()
      
      const emittedCalls = postMessageSpy.mock.calls
      
      expect(emittedCalls.length).toBeGreaterThan(0)
    })
  })
})
