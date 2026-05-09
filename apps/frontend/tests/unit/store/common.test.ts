import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia } from 'pinia'
import { store as globalStore } from '@/store/index'
import { dvMainStore } from '@/store/modules/data-visualization/dvMain'
import {
  getComponentById,
  getCurInfo,
  getCurInfoById,
  componentArraySort
} from '@/store/modules/data-visualization/common'

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

describe('common store helpers', () => {
  beforeEach(() => {
    setActivePinia(globalStore)
    dvMainStore(globalStore).$reset()
    vi.clearAllMocks()
  })

  const setupFlatComponents = () => {
    const dvMain = dvMainStore(globalStore)
    const components = [
      { id: 'comp-a', canvasId: 'canvas-main', category: 'base', style: {} },
      { id: 'comp-b', canvasId: 'canvas-main', category: 'base', style: {} },
      { id: 'comp-c', canvasId: 'canvas-main', category: 'base', style: {} }
    ]
    dvMain.setComponentData(components)
    dvMain.setCurComponent({ component: components[1], index: 1 })
    return { dvMain, components }
  }

  const setupGroupComponents = () => {
    const dvMain = dvMainStore(globalStore)
    const child1 = { id: 'child-1', canvasId: 'canvas-main' }
    const child2 = { id: 'child-2', canvasId: 'canvas-main' }
    const components = [
      {
        id: 'group-1',
        canvasId: 'canvas-main',
        component: 'Group',
        propValue: [child1, child2]
      },
      { id: 'standalone-1', canvasId: 'canvas-main' }
    ]
    dvMain.setComponentData(components)
    dvMain.setCurComponent({ component: components[0], index: 0 })
    return { dvMain, components, child1, child2 }
  }

  const setupDeTabsComponents = () => {
    const dvMain = dvMainStore(globalStore)
    const tabComponent = {
      id: 'tab-1',
      canvasId: 'canvas-main',
      component: 'DeTabs',
      propValue: [
        {
          componentData: [
            { id: 'tab-child-1', canvasId: 'canvas-main' },
            { id: 'tab-child-2', canvasId: 'canvas-main' }
          ]
        },
        {
          componentData: [{ id: 'tab-child-3', canvasId: 'canvas-main' }]
        }
      ]
    }
    const components = [tabComponent, { id: 'standalone-1', canvasId: 'canvas-main' }]
    dvMain.setComponentData(components)
    dvMain.setCurComponent({ component: components[0], index: 0 })
    return { dvMain, components, tabComponent }
  }

  describe('getComponentById', () => {
    it('should find a component by id in flat list', () => {
      setupFlatComponents()

      const result = getComponentById('comp-b')

      expect(result).toBeDefined()
      expect(result.id).toBe('comp-b')
    })

    it('should find a child component inside a Group', () => {
      setupGroupComponents()

      const result = getComponentById('child-1')

      expect(result).toBeDefined()
      expect(result.id).toBe('child-1')
    })

    it('should find a component inside DeTabs', () => {
      setupDeTabsComponents()

      const result = getComponentById('tab-child-2')

      expect(result).toBeDefined()
      expect(result.id).toBe('tab-child-2')
    })

    it('should return null for non-existent id', () => {
      setupFlatComponents()

      const result = getComponentById('nonexistent')

      expect(result).toBeNull()
    })

    it('should return curComponent when no id is provided', () => {
      setupFlatComponents()

      const result = getComponentById()

      expect(result).toBeDefined()
      expect(result.id).toBe('comp-b')
    })
  })

  describe('getCurInfo', () => {
    it('should return info for curComponent when no id provided', () => {
      setupFlatComponents()

      const info = getCurInfo()

      expect(info.index).toBe(1)
      expect(info.targetComponent.id).toBe('comp-b')
    })

    it('should return info for specific component by id', () => {
      setupFlatComponents()

      const info = getCurInfo('comp-a')

      expect(info.index).toBe(0)
      expect(info.targetComponent.id).toBe('comp-a')
    })
  })

  describe('getCurInfoById', () => {
    it('should return info for component at root level', () => {
      setupFlatComponents()

      const info = getCurInfoById('comp-c')

      expect(info.index).toBe(2)
      expect(info.targetComponent.id).toBe('comp-c')
    })

    it('should return info for component inside Group propValue', () => {
      const { child2 } = setupGroupComponents()

      const info = getCurInfoById('child-2')

      expect(info.targetComponent.id).toBe('child-2')
      expect(info.targetComponent).toEqual(child2)
    })

    it('should return info for component inside DeTabs', () => {
      setupDeTabsComponents()

      const info = getCurInfoById('tab-child-3')

      expect(info.targetComponent.id).toBe('tab-child-3')
      expect(info.tabIndex).toBe(1)
    })

    it('should return undefined when id is falsy', () => {
      setupFlatComponents()

      expect(getCurInfoById(null)).toBeUndefined()
      expect(getCurInfoById('')).toBeUndefined()
    })
  })

  describe('componentArraySort', () => {
    it('should sort array by componentData index in descending order (down)', () => {
      setupFlatComponents()

      const sortArray = [
        { id: 'comp-a' },
        { id: 'comp-c' },
        { id: 'comp-b' }
      ]

      componentArraySort(sortArray)

      expect(sortArray[0].id).toBe('comp-c')
      expect(sortArray[1].id).toBe('comp-b')
      expect(sortArray[2].id).toBe('comp-a')
    })

    it('should sort array by componentData index in ascending order (up)', () => {
      setupFlatComponents()

      const sortArray = [
        { id: 'comp-c' },
        { id: 'comp-a' },
        { id: 'comp-b' }
      ]

      componentArraySort(sortArray, 'up')

      expect(sortArray[0].id).toBe('comp-a')
      expect(sortArray[1].id).toBe('comp-b')
      expect(sortArray[2].id).toBe('comp-c')
    })

    it('should sort array with default direction down', () => {
      setupFlatComponents()

      const sortArray = [
        { id: 'comp-a' },
        { id: 'comp-b' },
        { id: 'comp-c' }
      ]

      componentArraySort(sortArray)

      expect(sortArray[0].id).toBe('comp-c')
      expect(sortArray[1].id).toBe('comp-b')
      expect(sortArray[2].id).toBe('comp-a')
    })
  })
})
