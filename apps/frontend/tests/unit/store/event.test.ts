import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia } from 'pinia'
import { store as globalStore } from '@/store/index'
import { dvMainStore } from '@/store/modules/data-visualization/dvMain'
import { eventStore } from '@/store/modules/data-visualization/event'

vi.mock('@/hooks/web/useEmitt', async () => {
  const { createUseEmittModuleMock } = await import('../helpers')
  return createUseEmittModuleMock()
})

vi.mock('@/views/chart/components/editor/util/chart', async () => {
  const { createChartEditorUtilModuleMock } = await import('../helpers')
  return createChartEditorUtilModuleMock()
})

vi.mock('@/views/chart/components/editor/util/dataVisualization', async () => {
  const { createDataVisualizationEditorUtilModuleMock } = await import('../helpers')
  return createDataVisualizationEditorUtilModuleMock()
})

vi.mock('@/views/chart/components/js/panel', async () => {
  const { createPanelModuleMock } = await import('../helpers')
  return createPanelModuleMock()
})

vi.mock('@/custom-component/component-list', async () => {
  const { createComponentListModuleMock } = await import('../helpers')
  return createComponentListModuleMock()
})

vi.mock('@/utils/viewUtils', async () => {
  const { createViewUtilsModuleMock } = await import('../helpers')
  return createViewUtilsModuleMock()
})

vi.mock('@/store/modules/appearance', async () => {
  const { createAppearanceStoreWithOutModuleMock } = await import('../helpers')
  return createAppearanceStoreWithOutModuleMock()
})

vi.mock('@/utils/componentUtils', async () => {
  const { createComponentUtilsModuleMock } = await import('../helpers')
  return createComponentUtilsModuleMock()
})

vi.mock('@/views/chart/components/js/formatter', async () => {
  const { createFormatterModuleMock } = await import('../helpers')
  return createFormatterModuleMock()
})

vi.mock('@/hooks/web/useI18n', async () => {
  const { useI18nModuleMock } = await import('../helpers')
  return useI18nModuleMock
})

vi.mock('element-plus-secondary', async () => {
  const { elementPlusSecondaryModuleMock } = await import('../helpers')
  return elementPlusSecondaryModuleMock
})

describe('event Store', () => {
  beforeEach(() => {
    setActivePinia(globalStore)
    dvMainStore(globalStore).$reset()
    vi.clearAllMocks()
  })

  const setupWithEvents = (events: Record<string, any> = {}) => {
    const dvMain = dvMainStore(globalStore)
    const component = {
      id: 'comp-1',
      canvasId: 'canvas-main',
      category: 'base',
      style: {},
      events
    }
    dvMain.setComponentData([component])
    dvMain.setCurComponent({ component, index: 0 })
    return { dvMain, component }
  }

  describe('addEvent', () => {
    it('should add event to curComponent events', () => {
      const { component } = setupWithEvents()
      const store = eventStore(globalStore)

      store.addEvent({ event: 'onClick', param: { action: 'jump' } })

      expect(component.events.onClick).toEqual({ action: 'jump' })
    })

    it('should add multiple different events', () => {
      const { component } = setupWithEvents()
      const store = eventStore(globalStore)

      store.addEvent({ event: 'onClick', param: { action: 'jump' } })
      store.addEvent({ event: 'onHover', param: { action: 'tooltip' } })

      expect(component.events.onClick).toEqual({ action: 'jump' })
      expect(component.events.onHover).toEqual({ action: 'tooltip' })
    })

    it('should overwrite existing event with same key', () => {
      const { component } = setupWithEvents()
      const store = eventStore(globalStore)

      store.addEvent({ event: 'onClick', param: { action: 'jump' } })
      store.addEvent({ event: 'onClick', param: { action: 'dialog' } })

      expect(component.events.onClick).toEqual({ action: 'dialog' })
    })
  })

  describe('removeEvent', () => {
    it('should remove event from curComponent events', () => {
      const { component } = setupWithEvents({
        onClick: { action: 'jump' },
        onHover: { action: 'tooltip' }
      })
      const store = eventStore(globalStore)

      store.removeEvent('onClick')

      expect(component.events.onClick).toBeUndefined()
      expect(component.events.onHover).toEqual({ action: 'tooltip' })
    })

    it('should handle removing non-existent event', () => {
      const { component } = setupWithEvents({
        onClick: { action: 'jump' }
      })
      const store = eventStore(globalStore)

      store.removeEvent('onHover')

      expect(component.events.onClick).toEqual({ action: 'jump' })
    })

    it('should remove all events one by one', () => {
      const { component } = setupWithEvents({
        onClick: { action: 'jump' },
        onHover: { action: 'tooltip' }
      })
      const store = eventStore(globalStore)

      store.removeEvent('onClick')
      store.removeEvent('onHover')

      expect(Object.keys(component.events)).toHaveLength(0)
    })
  })

  describe('displayEventChange', () => {
    it('should toggle displayChange from false to true and show hidden components', () => {
      const dvMain = dvMainStore(globalStore)
      const component = {
        id: 'comp-1',
        canvasId: 'canvas-main',
        category: 'base',
        style: {},
        events: {
          displayChange: { value: false }
        }
      }
      const hiddenComp = {
        id: 'hidden-1',
        category: 'hidden',
        isShow: false
      }
      dvMain.setComponentData([component, hiddenComp])
      dvMain.setCurComponent({ component, index: 0 })
      const store = eventStore(globalStore)

      store.displayEventChange(component)

      expect(component.events.displayChange.value).toBe(true)
      expect(hiddenComp.isShow).toBe(true)
      expect(dvMain.canvasState.curPointArea).toBe('base')
    })

    it('should toggle displayChange from true to false and hide hidden components', () => {
      const dvMain = dvMainStore(globalStore)
      const component = {
        id: 'comp-1',
        canvasId: 'canvas-main',
        category: 'base',
        style: {},
        events: {
          displayChange: { value: true }
        }
      }
      const hiddenComp = {
        id: 'hidden-1',
        category: 'hidden',
        isShow: true
      }
      dvMain.setComponentData([component, hiddenComp])
      dvMain.setCurComponent({ component, index: 0 })
      const store = eventStore(globalStore)

      store.displayEventChange(component)

      expect(component.events.displayChange.value).toBe(false)
      expect(hiddenComp.isShow).toBe(false)
      expect(dvMain.canvasState.curPointArea).toBe('hidden')
    })

    it('should only affect components with category hidden', () => {
      const dvMain = dvMainStore(globalStore)
      const component = {
        id: 'comp-1',
        canvasId: 'canvas-main',
        category: 'base',
        style: {},
        events: {
          displayChange: { value: false }
        }
      }
      const hiddenComp = {
        id: 'hidden-1',
        category: 'hidden',
        isShow: false
      }
      const baseComp = {
        id: 'base-1',
        category: 'base',
        isShow: true
      }
      dvMain.setComponentData([component, hiddenComp, baseComp])
      dvMain.setCurComponent({ component, index: 0 })
      const store = eventStore(globalStore)

      store.displayEventChange(component)

      expect(hiddenComp.isShow).toBe(true)
      expect(baseComp.isShow).toBe(true)
    })

    it('should toggle back and forth correctly', () => {
      const dvMain = dvMainStore(globalStore)
      const component = {
        id: 'comp-1',
        canvasId: 'canvas-main',
        category: 'base',
        style: {},
        events: {
          displayChange: { value: false }
        }
      }
      const hiddenComp = {
        id: 'hidden-1',
        category: 'hidden',
        isShow: false
      }
      dvMain.setComponentData([component, hiddenComp])
      dvMain.setCurComponent({ component, index: 0 })
      const store = eventStore(globalStore)

      store.displayEventChange(component)
      expect(component.events.displayChange.value).toBe(true)
      expect(hiddenComp.isShow).toBe(true)

      store.displayEventChange(component)
      expect(component.events.displayChange.value).toBe(false)
      expect(hiddenComp.isShow).toBe(false)
    })
  })
})
