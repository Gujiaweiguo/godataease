import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia } from 'pinia'
import { store as globalStore } from '@/store/index'
import { dvMainStore } from '@/store/modules/data-visualization/dvMain'
import { composeStore } from '@/store/modules/data-visualization/compose'
import eventBus from '@/utils/eventBus'

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
  return {
    ...createComponentListModuleMock(),
    commonStyle: {
      rotate: 0,
      opacity: 1,
      borderActive: false,
      borderWidth: 1,
      borderRadius: 5,
      borderStyle: 'solid',
      borderColor: '#cccccc'
    },
    commonAttr: { animations: [], events: {}, groupStyle: {} },
    COMMON_COMPONENT_BACKGROUND_MAP: {
      dark: { backgroundColor: '#1a1a1a' },
      light: { backgroundColor: '#ffffff' }
    }
  }
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

vi.mock('@/utils/eventBus', async () => {
  const { createEventBusModuleMock } = await import('../helpers')
  return createEventBusModuleMock()
})

vi.mock('@/utils/decomposeComponent', () => ({
  default: vi.fn()
}))

vi.mock('@/utils/generateID', () => ({
  generateID: vi.fn(() => 'generated-id-123')
}))

vi.mock('@/utils/style', () => ({
  createGroupStyle: vi.fn(),
  getComponentRotatedStyle: vi.fn((style) => ({
    left: style.left || 0,
    top: style.top || 0,
    right: (style.left || 0) + (style.width || 0),
    bottom: (style.top || 0) + (style.height || 0)
  }))
}))

vi.mock('@/utils/canvasUtils', () => ({
  canvasIdMapCheck: vi.fn(),
  checkJoinGroup: vi.fn(() => true),
  isTabCanvas: vi.fn(() => false)
}))

describe('Compose Store', () => {
  beforeEach(() => {
    setActivePinia(globalStore)
    dvMainStore(globalStore).$reset()
    vi.clearAllMocks()
  })

  const setupWithComponents = (components) => {
    const dvMain = dvMainStore(globalStore)
    dvMain.setComponentData([...components])
    return dvMain
  }

  describe('Initial State', () => {
    it('should have empty areaData', () => {
      const store = composeStore(globalStore)
      expect(store.areaData).toEqual({
        style: { top: 0, left: 0, width: 0, height: 0 },
        components: []
      })
    })

    it('should have isCtrlOrCmdDown as false', () => {
      const store = composeStore(globalStore)
      expect(store.isCtrlOrCmdDown).toBe(false)
    })

    it('should have isSpaceDown as false', () => {
      const store = composeStore(globalStore)
      expect(store.isSpaceDown).toBe(false)
    })

    it('should have isShiftDown as false', () => {
      const store = composeStore(globalStore)
      expect(store.isShiftDown).toBe(false)
    })

    it('should have laterIndex as null', () => {
      const store = composeStore(globalStore)
      expect(store.laterIndex).toBeNull()
    })
  })

  describe('setAreaData', () => {
    it('should set areaData', () => {
      const store = composeStore(globalStore)
      const data = {
        style: { top: 10, left: 20, width: 100, height: 200 },
        components: [{ id: 'comp-1' }]
      }
      store.setAreaData(data)
      expect(store.areaData).toEqual(data)
    })
  })

  describe('setLaterIndex', () => {
    it('should set laterIndex', () => {
      const store = composeStore(globalStore)
      store.setLaterIndex(3)
      expect(store.laterIndex).toBe(3)
    })
  })

  describe('setIsCtrlOrCmdDownStatus', () => {
    it('should set isCtrlOrCmdDown', () => {
      const store = composeStore(globalStore)
      store.setIsCtrlOrCmdDownStatus(true)
      expect(store.isCtrlOrCmdDown).toBe(true)
    })
  })

  describe('setSpaceDownStatus', () => {
    it('should set isSpaceDown', () => {
      const store = composeStore(globalStore)
      store.setSpaceDownStatus(true)
      expect(store.isSpaceDown).toBe(true)
    })
  })

  describe('setIsShiftDownStatus', () => {
    it('should set isShiftDown', () => {
      const store = composeStore(globalStore)
      store.setIsShiftDownStatus(true)
      expect(store.isShiftDown).toBe(true)
    })
  })

  describe('alignment', () => {
    it('should release when only one component in areaData', () => {
      const store = composeStore(globalStore)
      store.setAreaData({
        style: { top: 0, left: 0, width: 0, height: 0 },
        components: [{ id: 'comp-1', style: {} }]
      })

      store.alignment('left')

      expect(store.areaData.components).toEqual([])
    })

    it('should align components to left', () => {
      const store = composeStore(globalStore)
      const comp1 = { id: 'c1', style: { left: 10, top: 0, width: 50, height: 50 } }
      const comp2 = { id: 'c2', style: { left: 30, top: 0, width: 50, height: 50 } }
      store.setAreaData({
        style: { top: 0, left: 10, width: 70, height: 50 },
        components: [comp1, comp2]
      })

      store.alignment('left')

      expect(comp1.style.left).toBe(10)
      expect(comp2.style.left).toBe(10)
    })

    it('should align components to right', () => {
      const store = composeStore(globalStore)
      const comp1 = { id: 'c1', style: { left: 10, top: 0, width: 50, height: 50 } }
      const comp2 = { id: 'c2', style: { left: 30, top: 0, width: 60, height: 50 } }
      store.setAreaData({
        style: { top: 0, left: 10, width: 80, height: 50 },
        components: [comp1, comp2]
      })

      store.alignment('right')

      expect(comp1.style.left).toBe(90 - 50)
      expect(comp2.style.left).toBe(90 - 60)
    })

    it('should align components to top', () => {
      const store = composeStore(globalStore)
      const comp1 = { id: 'c1', style: { left: 0, top: 5, width: 50, height: 50 } }
      const comp2 = { id: 'c2', style: { left: 0, top: 20, width: 50, height: 50 } }
      store.setAreaData({
        style: { top: 5, left: 0, width: 50, height: 65 },
        components: [comp1, comp2]
      })

      store.alignment('top')

      expect(comp1.style.top).toBe(5)
      expect(comp2.style.top).toBe(5)
    })

    it('should align components to bottom', () => {
      const store = composeStore(globalStore)
      const comp1 = { id: 'c1', style: { left: 0, top: 5, width: 50, height: 30 } }
      const comp2 = { id: 'c2', style: { left: 0, top: 10, width: 50, height: 40 } }
      store.setAreaData({
        style: { top: 5, left: 0, width: 50, height: 70 },
        components: [comp1, comp2]
      })

      store.alignment('bottom')

      expect(comp1.style.top).toBe(5 + 70 - 30)
      expect(comp2.style.top).toBe(5 + 70 - 40)
    })

    it('should align components to transverse center', () => {
      const store = composeStore(globalStore)
      const comp1 = { id: 'c1', style: { left: 10, top: 0, width: 40, height: 50 } }
      const comp2 = { id: 'c2', style: { left: 50, top: 0, width: 30, height: 50 } }
      store.setAreaData({
        style: { top: 0, left: 10, width: 100, height: 50 },
        components: [comp1, comp2]
      })

      store.alignment('transverse')

      expect(comp1.style.left).toBe(10 + 100 / 2 - 40 / 2)
      expect(comp2.style.left).toBe(10 + 100 / 2 - 30 / 2)
    })

    it('should align components to direction center', () => {
      const store = composeStore(globalStore)
      const comp1 = { id: 'c1', style: { left: 0, top: 10, width: 50, height: 40 } }
      const comp2 = { id: 'c2', style: { left: 0, top: 30, width: 50, height: 20 } }
      store.setAreaData({
        style: { top: 10, left: 0, width: 50, height: 80 },
        components: [comp1, comp2]
      })

      store.alignment('direction')

      expect(comp1.style.top).toBe(10 + 80 / 2 - 40 / 2)
      expect(comp2.style.top).toBe(10 + 80 / 2 - 20 / 2)
    })
  })

  describe('compose', () => {
    it('should release when only one component in areaData', () => {
      const store = composeStore(globalStore)
      store.setAreaData({
        style: { top: 0, left: 0, width: 0, height: 0 },
        components: [{ id: 'c1', component: 'UserView', style: {} }]
      })

      store.compose('canvas-main')

      expect(store.areaData.components).toEqual([])
    })

    it('should emit hideArea event when composing multiple components', () => {
      const comp1 = {
        id: 'c1',
        component: 'UserView',
        canvasId: 'canvas-main',
        style: { left: 0, top: 0, width: 100, height: 100 }
      }
      const comp2 = {
        id: 'c2',
        component: 'UserView',
        canvasId: 'canvas-main',
        style: { left: 50, top: 50, width: 100, height: 100 }
      }
      const dvMain = setupWithComponents([comp1, comp2])

      const store = composeStore(globalStore)
      store.setAreaData({
        style: { top: 0, left: 0, width: 150, height: 150 },
        components: [comp1, comp2]
      })
      store.editorMap['canvas-main'] = { getBoundingClientRect: () => ({ left: 0, top: 0 }) }

      const addComponentSpy = vi.spyOn(dvMain, 'addComponent')
      addComponentSpy.mockImplementation(() => {})
      const batchSpy = vi.spyOn(store, 'batchDeleteComponent')

      store.compose('canvas-main')

      expect(eventBus.emit).toHaveBeenCalledWith('hideArea-canvas-main')
      expect(store.areaData.components).toEqual([])
      expect(batchSpy).toHaveBeenCalled()
    })
  })

  describe('batchDeleteComponent', () => {
    it('should remove components from componentData', () => {
      const comp1 = { id: 'c1', component: 'UserView' }
      const comp2 = { id: 'c2', component: 'UserView' }
      const comp3 = { id: 'c3', component: 'UserView' }
      setupWithComponents([comp1, comp2, comp3])

      const store = composeStore(globalStore)
      store.batchDeleteComponent([comp1, comp3])

      const dvMain = dvMainStore(globalStore)
      expect(dvMain.componentData).toHaveLength(1)
      expect(dvMain.componentData[0].id).toBe('c2')
    })

    it('should skip GroupArea components during batch delete', () => {
      const comp1 = { id: 'c1', component: 'UserView' }
      const groupArea = { id: 'ga1', component: 'GroupArea' }
      setupWithComponents([comp1, groupArea])

      const store = composeStore(globalStore)
      store.batchDeleteComponent([comp1, groupArea])

      const dvMain = dvMainStore(globalStore)
      expect(dvMain.componentData).toHaveLength(1)
      expect(dvMain.componentData[0].id).toBe('ga1')
    })
  })

  describe('calcComposeArea', () => {
    it('should calculate bounding box from components', () => {
      const store = composeStore(globalStore)
      store.setAreaData({
        style: { top: 0, left: 0, width: 0, height: 0 },
        components: [
          { style: { left: 10, top: 20, width: 100, height: 50 } },
          { style: { left: 30, top: 5, width: 80, height: 60 } }
        ]
      })

      store.calcComposeArea()

      expect(store.areaData.style.left).toBe(10)
      expect(store.areaData.style.top).toBe(5)
      expect(store.areaData.style.width).toBe(100)
      expect(store.areaData.style.height).toBe(65)
    })
  })
})
