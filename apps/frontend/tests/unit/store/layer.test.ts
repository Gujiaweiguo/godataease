import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia } from 'pinia'
import { store as globalStore } from '@/store/index'
import { dvMainStore } from '@/store/modules/data-visualization/dvMain'
import { layerStore } from '@/store/modules/data-visualization/layer'

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

vi.mock('@/views/chart/components/js/g2plot_tooltip_carousel', () => ({
  default: {
    paused: vi.fn(),
    resume: vi.fn()
  }
}))

describe('layer Store', () => {
  beforeEach(() => {
    setActivePinia(globalStore)
    dvMainStore(globalStore).$reset()
    vi.clearAllMocks()
  })

  const createComponents = () => [
    { id: 'comp-a', canvasId: 'canvas-main', category: 'base', style: {} },
    { id: 'comp-b', canvasId: 'canvas-main', category: 'base', style: {} },
    { id: 'comp-c', canvasId: 'canvas-main', category: 'base', style: {} }
  ]

  const setupWithComponents = (index = 1) => {
    const dvMain = dvMainStore(globalStore)
    const components = createComponents()
    dvMain.setComponentData(components)
    dvMain.setCurComponent({ component: components[index], index })
    return { dvMain, components }
  }

  describe('upComponent', () => {
    it('should swap component with next one (move up in layer)', () => {
      const { dvMain } = setupWithComponents(0)
      const store = layerStore(globalStore)

      store.upComponent()

      expect(dvMain.componentData[0].id).toBe('comp-b')
      expect(dvMain.componentData[1].id).toBe('comp-a')
      expect(dvMain.curComponentIndex).toBe(1)
    })

    it('should not move if component is already at the top', () => {
      const { dvMain } = setupWithComponents(2)
      const store = layerStore(globalStore)

      store.upComponent()

      expect(dvMain.componentData[2].id).toBe('comp-c')
      expect(dvMain.curComponentIndex).toBe(2)
    })

    it('should move specific component up by id', () => {
      const { dvMain } = setupWithComponents(0)
      const store = layerStore(globalStore)

      store.upComponent('comp-a')

      expect(dvMain.componentData[0].id).toBe('comp-b')
      expect(dvMain.componentData[1].id).toBe('comp-a')
    })
  })

  describe('downComponent', () => {
    it('should swap component with previous one (move down in layer)', () => {
      const { dvMain } = setupWithComponents(2)
      const store = layerStore(globalStore)

      store.downComponent()

      expect(dvMain.componentData[1].id).toBe('comp-c')
      expect(dvMain.componentData[2].id).toBe('comp-b')
      expect(dvMain.curComponentIndex).toBe(1)
    })

    it('should not move if component is already at the bottom', () => {
      const { dvMain } = setupWithComponents(0)
      const store = layerStore(globalStore)

      store.downComponent()

      expect(dvMain.componentData[0].id).toBe('comp-a')
      expect(dvMain.curComponentIndex).toBe(0)
    })

    it('should move specific component down by id', () => {
      const { dvMain } = setupWithComponents(0)
      const store = layerStore(globalStore)

      store.downComponent('comp-c')

      expect(dvMain.componentData[1].id).toBe('comp-c')
      expect(dvMain.componentData[2].id).toBe('comp-b')
    })
  })

  describe('topComponent', () => {
    it('should move component to the top (end of array)', () => {
      const { dvMain } = setupWithComponents(0)
      const store = layerStore(globalStore)

      store.topComponent()

      expect(dvMain.componentData[2].id).toBe('comp-a')
      expect(dvMain.curComponentIndex).toBe(2)
    })

    it('should not move if component is already at the top', () => {
      const { dvMain } = setupWithComponents(2)
      const store = layerStore(globalStore)

      store.topComponent()

      expect(dvMain.componentData[2].id).toBe('comp-c')
      expect(dvMain.curComponentIndex).toBe(2)
    })
  })

  describe('bottomComponent', () => {
    it('should move component to the bottom (start of array)', () => {
      const { dvMain } = setupWithComponents(2)
      const store = layerStore(globalStore)

      store.bottomComponent()

      expect(dvMain.componentData[0].id).toBe('comp-c')
      expect(dvMain.curComponentIndex).toBe(0)
    })

    it('should not move if component is already at the bottom', () => {
      const { dvMain } = setupWithComponents(0)
      const store = layerStore(globalStore)

      store.bottomComponent()

      expect(dvMain.componentData[0].id).toBe('comp-a')
      expect(dvMain.curComponentIndex).toBe(0)
    })
  })

  describe('hideComponent', () => {
    it('should set isShow to false for target component', () => {
      const { dvMain } = setupWithComponents(0)
      dvMain.componentData[1].isShow = true
      const store = layerStore(globalStore)

      store.hideComponent('comp-b')

      expect(dvMain.componentData[1].isShow).toBe(false)
    })

    it('should not throw when component id not found', () => {
      setupWithComponents(0)
      const store = layerStore(globalStore)

      expect(() => store.hideComponent('nonexistent')).not.toThrow()
    })
  })

  describe('hideComponentWithComponent', () => {
    it('should set isShow to false for target component', () => {
      const { dvMain } = setupWithComponents(0)
      dvMain.componentData[1].isShow = true
      const store = layerStore(globalStore)

      store.hideComponentWithComponent('comp-b')

      expect(dvMain.componentData[1].isShow).toBe(false)
    })
  })

  describe('showComponent', () => {
    it('should set isShow to true for target component', () => {
      const { dvMain } = setupWithComponents(0)
      dvMain.componentData[1].isShow = false
      const store = layerStore(globalStore)

      store.showComponent('comp-b')

      expect(dvMain.componentData[1].isShow).toBe(true)
    })

    it('should emit renderChart for Group children with table innerType', async () => {
      vi.useFakeTimers()
      const { useEmitt } = await import('@/hooks/web/useEmitt')
      const dvMain = dvMainStore(globalStore)
      const groupComponent = {
        id: 'group-1',
        canvasId: 'canvas-main',
        category: 'base',
        style: {},
        component: 'Group',
        propValue: [
          { id: 'child-table', innerType: 'table-info' },
          { id: 'child-other', innerType: 'bar' }
        ],
        isShow: false
      }
      dvMain.setComponentData([groupComponent])
      dvMain.setCurComponent({ component: groupComponent, index: 0 })
      const store = layerStore(globalStore)

      store.showComponent('group-1')

      expect(groupComponent.isShow).toBe(true)
      vi.advanceTimersByTime(400)
      expect(useEmitt().emitter.emit).toHaveBeenCalledWith('renderChart-child-table')
      vi.useRealTimers()
    })

    it('should not throw when component id not found', () => {
      setupWithComponents(0)
      const store = layerStore(globalStore)

      expect(() => store.showComponent('nonexistent')).not.toThrow()
    })
  })

  describe('pausedTooltipCarousel', () => {
    it('should call ChartCarouselTooltip.paused with component id', async () => {
      const ChartCarouselTooltip = (
        await import('@/views/chart/components/js/g2plot_tooltip_carousel')
      ).default
      setupWithComponents(0)
      const store = layerStore(globalStore)

      store.pausedTooltipCarousel('comp-a')

      expect(ChartCarouselTooltip.paused).toHaveBeenCalledWith('comp-a')
    })

    it('should not call paused when component not found', async () => {
      const ChartCarouselTooltip = (
        await import('@/views/chart/components/js/g2plot_tooltip_carousel')
      ).default
      setupWithComponents(0)
      const store = layerStore(globalStore)

      store.pausedTooltipCarousel('nonexistent')

      expect(ChartCarouselTooltip.paused).not.toHaveBeenCalled()
    })
  })

  describe('resumeTooltipCarousel', () => {
    it('should call ChartCarouselTooltip.resume with component id', async () => {
      const ChartCarouselTooltip = (
        await import('@/views/chart/components/js/g2plot_tooltip_carousel')
      ).default
      setupWithComponents(0)
      const store = layerStore(globalStore)

      store.resumeTooltipCarousel('comp-a')

      expect(ChartCarouselTooltip.resume).toHaveBeenCalledWith('comp-a')
    })

    it('should not call resume when component not found', async () => {
      const ChartCarouselTooltip = (
        await import('@/views/chart/components/js/g2plot_tooltip_carousel')
      ).default
      setupWithComponents(0)
      const store = layerStore(globalStore)

      store.resumeTooltipCarousel('nonexistent')

      expect(ChartCarouselTooltip.resume).not.toHaveBeenCalled()
    })
  })
})
