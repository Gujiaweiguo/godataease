import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia } from 'pinia'
import { store as globalStore } from '@/store/index'
import { dvMainStore } from '@/store/modules/data-visualization/dvMain'
import { animationStore } from '@/store/modules/data-visualization/animation'

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

describe('animation Store', () => {
  beforeEach(() => {
    setActivePinia(globalStore)
    dvMainStore(globalStore).$reset()
    vi.clearAllMocks()
  })

  const setupWithAnimations = () => {
    const dvMain = dvMainStore(globalStore)
    const component = {
      id: 'comp-1',
      canvasId: 'canvas-main',
      category: 'base',
      style: {},
      animations: []
    }
    dvMain.setComponentData([component])
    dvMain.setCurComponent({ component, index: 0 })
    return { dvMain, component }
  }

  describe('addAnimation', () => {
    it('should add animation to curComponent animations array', () => {
      const { component } = setupWithAnimations()
      const store = animationStore(globalStore)

      const animation = { type: 'fade', duration: 500 }
      store.addAnimation(animation)

      expect(component.animations).toHaveLength(1)
      expect(component.animations[0]).toEqual({ type: 'fade', duration: 500 })
    })

    it('should add multiple animations in order', () => {
      const { component } = setupWithAnimations()
      const store = animationStore(globalStore)

      store.addAnimation({ type: 'fade' })
      store.addAnimation({ type: 'slide' })
      store.addAnimation({ type: 'zoom' })

      expect(component.animations).toHaveLength(3)
      expect(component.animations[0].type).toBe('fade')
      expect(component.animations[1].type).toBe('slide')
      expect(component.animations[2].type).toBe('zoom')
    })
  })

  describe('removeAnimation', () => {
    it('should remove animation at specified index', () => {
      const { component } = setupWithAnimations()
      const store = animationStore(globalStore)

      store.addAnimation({ type: 'fade' })
      store.addAnimation({ type: 'slide' })
      store.addAnimation({ type: 'zoom' })

      store.removeAnimation(1)

      expect(component.animations).toHaveLength(2)
      expect(component.animations[0].type).toBe('fade')
      expect(component.animations[1].type).toBe('zoom')
    })

    it('should remove first animation', () => {
      const { component } = setupWithAnimations()
      const store = animationStore(globalStore)

      store.addAnimation({ type: 'fade' })
      store.addAnimation({ type: 'slide' })

      store.removeAnimation(0)

      expect(component.animations).toHaveLength(1)
      expect(component.animations[0].type).toBe('slide')
    })

    it('should remove last animation', () => {
      const { component } = setupWithAnimations()
      const store = animationStore(globalStore)

      store.addAnimation({ type: 'fade' })
      store.addAnimation({ type: 'slide' })

      store.removeAnimation(1)

      expect(component.animations).toHaveLength(1)
      expect(component.animations[0].type).toBe('fade')
    })
  })

  describe('alterAnimation', () => {
    it('should merge data into existing animation at index', () => {
      const { component } = setupWithAnimations()
      const store = animationStore(globalStore)

      store.addAnimation({ type: 'fade', duration: 500 })
      store.alterAnimation({ index: 0, data: { duration: 1000, delay: 200 } })

      expect(component.animations[0]).toEqual({
        type: 'fade',
        duration: 1000,
        delay: 200
      })
    })

    it('should preserve original properties when merging partial data', () => {
      const { component } = setupWithAnimations()
      const store = animationStore(globalStore)

      store.addAnimation({ type: 'slide', duration: 300, direction: 'left' })
      store.alterAnimation({ index: 0, data: { direction: 'right' } })

      expect(component.animations[0]).toEqual({
        type: 'slide',
        duration: 300,
        direction: 'right'
      })
    })

    it('should handle empty data object in alterAnimation', () => {
      const { component } = setupWithAnimations()
      const store = animationStore(globalStore)

      store.addAnimation({ type: 'zoom', scale: 1.5 })
      store.alterAnimation({ index: 0, data: {} })

      expect(component.animations[0]).toEqual({ type: 'zoom', scale: 1.5 })
    })

    it('should not modify when index is not a number', () => {
      const { component } = setupWithAnimations()
      const store = animationStore(globalStore)

      store.addAnimation({ type: 'fade' })
      store.alterAnimation({ index: undefined, data: { type: 'slide' } })

      expect(component.animations[0]).toEqual({ type: 'fade' })
    })
  })

  describe('combined operations', () => {
    it('should handle add, alter, and remove sequence', () => {
      const { component } = setupWithAnimations()
      const store = animationStore(globalStore)

      store.addAnimation({ type: 'fade', duration: 500 })
      store.addAnimation({ type: 'slide', duration: 300 })
      store.alterAnimation({ index: 0, data: { duration: 1000 } })
      store.removeAnimation(1)

      expect(component.animations).toHaveLength(1)
      expect(component.animations[0]).toEqual({ type: 'fade', duration: 1000 })
    })
  })
})
