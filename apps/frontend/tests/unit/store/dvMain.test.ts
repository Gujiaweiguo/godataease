import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { dvMainStore } from '@/store/modules/data-visualization/dvMain'
import { useEmitt } from '@/hooks/web/useEmitt'

// Mock dependencies
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

// Mock useI18n to avoid inject warnings during module initialization
vi.mock('@/hooks/web/useI18n', async () => {
  const { useI18nModuleMock } = await import('../helpers')
  return useI18nModuleMock
})

// Mock element-plus-secondary to avoid inject warnings
vi.mock('element-plus-secondary', async () => {
  const { elementPlusSecondaryModuleMock } = await import('../helpers')
  return elementPlusSecondaryModuleMock
})

describe('dvMain Store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  describe('Initial State', () => {
    it('should have correct initial editMode', () => {
      const store = dvMainStore()
      expect(store.editMode).toBe('preview')
    })

    it('should have correct initial mobileInPc', () => {
      const store = dvMainStore()
      expect(store.mobileInPc).toBe(false)
    })

    it('should have correct initial isInEditor', () => {
      const store = dvMainStore()
      expect(store.isInEditor).toBe(false)
    })

    it('should have correct initial isClickComponent', () => {
      const store = dvMainStore()
      expect(store.isClickComponent).toBe(false)
    })

    it('should have correct initial componentData', () => {
      const store = dvMainStore()
      expect(store.componentData).toEqual([])
    })

    it('should have correct initial curComponent', () => {
      const store = dvMainStore()
      expect(store.curComponent).toBeNull()
    })

    it('should have correct initial curComponentIndex', () => {
      const store = dvMainStore()
      expect(store.curComponentIndex).toBeNull()
    })

    it('should have correct initial canvasViewInfo', () => {
      const store = dvMainStore()
      expect(store.canvasViewInfo).toEqual({})
    })

    it('should have correct initial fullscreenFlag', () => {
      const store = dvMainStore()
      expect(store.fullscreenFlag).toBe(false)
    })

    it('should have correct initial inMobile', () => {
      const store = dvMainStore()
      expect(store.inMobile).toBe(false)
    })
  })

  describe('setEditMode', () => {
    it('should set editMode to edit', () => {
      const store = dvMainStore()
      store.setEditMode('edit')
      expect(store.editMode).toBe('edit')
    })

    it('should set editMode to preview', () => {
      const store = dvMainStore()
      store.setEditMode('edit')
      store.setEditMode('preview')
      expect(store.editMode).toBe('preview')
    })
  })

  describe('setMobileInPc', () => {
    it('should set mobileInPc to true', () => {
      const store = dvMainStore()
      store.setMobileInPc(true)
      expect(store.mobileInPc).toBe(true)
    })

    it('should set mobileInPc to false', () => {
      const store = dvMainStore()
      store.setMobileInPc(true)
      store.setMobileInPc(false)
      expect(store.mobileInPc).toBe(false)
    })
  })

  describe('setInEditorStatus', () => {
    it('should set isInEditor to true', () => {
      const store = dvMainStore()
      store.setInEditorStatus(true)
      expect(store.isInEditor).toBe(true)
    })

    it('should set isInEditor to false', () => {
      const store = dvMainStore()
      store.setInEditorStatus(true)
      store.setInEditorStatus(false)
      expect(store.isInEditor).toBe(false)
    })
  })

  describe('setClickComponentStatus', () => {
    it('should set isClickComponent to true', () => {
      const store = dvMainStore()
      store.setClickComponentStatus(true)
      expect(store.isClickComponent).toBe(true)
    })

    it('should set isClickComponent to false', () => {
      const store = dvMainStore()
      store.setClickComponentStatus(true)
      store.setClickComponentStatus(false)
      expect(store.isClickComponent).toBe(false)
    })
  })

  describe('setFullscreenFlag', () => {
    it('should set fullscreenFlag to true', () => {
      const store = dvMainStore()
      store.setFullscreenFlag(true)
      expect(store.fullscreenFlag).toBe(true)
    })

    it('should set fullscreenFlag to false', () => {
      const store = dvMainStore()
      store.setFullscreenFlag(true)
      store.setFullscreenFlag(false)
      expect(store.fullscreenFlag).toBe(false)
    })
  })

  describe('setCurTabName', () => {
    it('should set curTabName', () => {
      const store = dvMainStore()
      store.setCurTabName('tab-1')
      expect(store.curTabName).toBe('tab-1')
    })

    it('should set curTabName to null', () => {
      const store = dvMainStore()
      store.setCurTabName('tab-1')
      store.setCurTabName(null)
      expect(store.curTabName).toBeNull()
    })
  })

  describe('setInMobile', () => {
    it('should set inMobile to true', () => {
      const store = dvMainStore()
      store.setInMobile(true)
      expect(store.inMobile).toBe(true)
    })

    it('should set inMobile to false', () => {
      const store = dvMainStore()
      store.setInMobile(true)
      store.setInMobile(false)
      expect(store.inMobile).toBe(false)
    })
  })

  describe('setAppDataInfo / getAppDataInfo', () => {
    it('should set and get appData', () => {
      const store = dvMainStore()
      const appData = { id: 'app-1', name: 'Test App' }
      store.setAppDataInfo(appData)
      expect(store.getAppDataInfo()).toEqual(appData)
    })

    it('should return null initially', () => {
      const store = dvMainStore()
      expect(store.getAppDataInfo()).toBeNull()
    })
  })

  describe('setCanvasViewInfo', () => {
    it('should set canvasViewInfo', () => {
      const store = dvMainStore()
      const viewInfo = {
        'view-1': { id: 'view-1', title: 'Chart 1' },
        'view-2': { id: 'view-2', title: 'Chart 2' }
      }
      store.setCanvasViewInfo(viewInfo)
      expect(store.canvasViewInfo).toEqual(viewInfo)
    })
  })

  describe('setCanvasStyleScale', () => {
    it('should update canvas scale', () => {
      const store = dvMainStore()
      store.setCanvasStyleScale(1.5)
      expect(store.canvasStyleData.scale).toBe(1.5)
    })
  })

  describe('setLastHiddenComponent', () => {
    it('should set and clear last hidden component', () => {
      const store = dvMainStore()
      store.setLastHiddenComponent({ id: 'hidden-1' })
      expect(store.lastHiddenComponent).toEqual([{ id: 'hidden-1' }])

      store.setLastHiddenComponent()
      expect(store.lastHiddenComponent).toEqual([])
    })
  })

  describe('setCurComponent', () => {
    it('should update current component and index', () => {
      const store = dvMainStore()
      const components = [
        { id: 'comp-1', canvasId: 'canvas-main', category: 'base', style: {} },
        { id: 'comp-2', canvasId: 'canvas-main', category: 'base', style: {} }
      ]
      store.setComponentData(components)
      store.setCurComponent({ component: components[1], index: 1 })

      expect(store.curComponent).toEqual(components[1])
      expect(store.curComponentIndex).toBe(1)
      expect(store.canvasState.curPointArea).toBe('base')
    })
  })

  describe('copy component helpers', () => {
    it('should add copied component on main canvas and remap canvas view info', () => {
      const store = dvMainStore()
      store.setComponentData([])
      store.setCanvasViewInfo({
        'old-view-1': { id: 'old-view-1', title: 'Old Chart' }
      })

      store.addCopyComponent(
        { id: 'new-view-1', canvasId: 'canvas-main', component: 'UserView' },
        { 'old-view-1': 'new-view-1' },
        store.canvasViewInfo
      )

      expect(store.componentData).toHaveLength(1)
      expect(store.componentData[0].id).toBe('new-view-1')
      expect(store.canvasViewInfo['new-view-1']).toEqual({
        id: 'new-view-1',
        title: 'Old Chart',
        linkageActive: false,
        jumpActive: false,
        customAttrMobile: null,
        customStyleMobile: null
      })
    })

    it('should remap multiple canvas view infos directly', () => {
      const store = dvMainStore()
      store.setCanvasViewInfo({
        'old-a': { id: 'old-a', title: 'A' },
        'old-b': { id: 'old-b', title: 'B' }
      })

      store.updateCopyCanvasView(
        {
          'old-a': 'new-a',
          'old-b': 'new-b'
        },
        store.canvasViewInfo
      )

      expect(store.canvasViewInfo['new-a'].id).toBe('new-a')
      expect(store.canvasViewInfo['new-b'].id).toBe('new-b')
      expect(store.canvasViewInfo['new-a'].linkageActive).toBe(false)
      expect(store.canvasViewInfo['new-b'].jumpActive).toBe(false)
    })
  })

  describe('deletion and linkage helpers', () => {
    it('should delete component by id', () => {
      const store = dvMainStore()
      store.setComponentData([
        { id: 'keep', component: 'UserView' },
        { id: 'remove', component: 'UserView' }
      ])

      store.deleteComponentById('remove')

      expect(store.componentData).toHaveLength(1)
      expect(store.componentData[0].id).toBe('keep')
    })

    it('should clear matching linkage filters and emit refresh', () => {
      const store = dvMainStore()
      store.setComponentData([
        {
          id: 'view-1',
          component: 'UserView',
          innerType: 'bar',
          linkageFilters: [{ sourceViewId: 'source-1' }, { sourceViewId: 'source-2' }]
        }
      ])

      store.clearViewLinkage('source-1')

      expect(store.componentData[0].linkageFilters).toEqual([
        { sourceViewId: 'source-2' }
      ])
      expect(useEmitt().emitter.emit).toHaveBeenCalledWith('query-data-view-1')
    })
  })

  describe('addCanvasViewInfo', () => {
    it('should add a new view info', () => {
      const store = dvMainStore()
      store.addCanvasViewInfo('view-1', { id: 'view-1', title: 'Chart 1' })
      expect(store.canvasViewInfo['view-1']).toEqual({ id: 'view-1', title: 'Chart 1' })
    })

    it('should add multiple view infos', () => {
      const store = dvMainStore()
      store.addCanvasViewInfo('view-1', { id: 'view-1', title: 'Chart 1' })
      store.addCanvasViewInfo('view-2', { id: 'view-2', title: 'Chart 2' })
      expect(Object.keys(store.canvasViewInfo)).toHaveLength(2)
    })
  })

  describe('removeCanvasViewInfo', () => {
    it('should remove a view info', () => {
      const store = dvMainStore()
      store.addCanvasViewInfo('view-1', { id: 'view-1', title: 'Chart 1' })
      store.addCanvasViewInfo('view-2', { id: 'view-2', title: 'Chart 2' })
      store.removeCanvasViewInfo('view-1')
      expect(store.canvasViewInfo['view-1']).toBeUndefined()
      expect(store.canvasViewInfo['view-2']).toBeDefined()
    })
  })

  describe('setComponentData', () => {
    it('should set componentData', () => {
      const store = dvMainStore()
      const components = [
        { id: 'comp-1', name: 'Component 1' },
        { id: 'comp-2', name: 'Component 2' }
      ]
      store.setComponentData(components)
      expect(store.componentData).toEqual(components)
    })

    it('should set componentData to empty array', () => {
      const store = dvMainStore()
      store.setComponentData([{ id: 'comp-1' }])
      store.setComponentData([])
      expect(store.componentData).toEqual([])
    })
  })

  describe('setEmbeddedCallBack', () => {
    it('should set embeddedCallBack', () => {
      const store = dvMainStore()
      store.setEmbeddedCallBack('yes')
      expect(store.embeddedCallBack).toBe('yes')
    })
  })

  describe('setPublicLinkStatus', () => {
    it('should set publicLinkStatus', () => {
      const store = dvMainStore()
      store.setPublicLinkStatus(true)
      expect(store.publicLinkStatus).toBe(true)
    })
  })

  describe('Canvas collapse state', () => {
    it('should have correct initial canvasCollapse', () => {
      const store = dvMainStore()
      expect(store.canvasCollapse.defaultSide).toBe(false)
      expect(store.canvasCollapse.realTimeComponent).toBe(false)
      expect(store.canvasCollapse.canvas).toBe(false)
      expect(store.canvasCollapse.componentProp).toBe(false)
    })
  })

  describe('Canvas state', () => {
    it('should have correct initial canvasState', () => {
      const store = dvMainStore()
      expect(store.canvasState.curPointArea).toBe('base')
    })
  })

  describe('Preview canvas scale', () => {
    it('should have correct initial previewCanvasScale', () => {
      const store = dvMainStore()
      expect(store.previewCanvasScale.scalePointWidth).toBe(1)
      expect(store.previewCanvasScale.scalePointHeight).toBe(1)
    })
  })

  describe('dvInfo initial state', () => {
    it('should have correct initial dvInfo', () => {
      const store = dvMainStore()
      expect(store.dvInfo.id).toBeNull()
      expect(store.dvInfo.name).toBeNull()
      expect(store.dvInfo.pid).toBeNull()
      expect(store.dvInfo.type).toBeNull()
      expect(store.dvInfo.mobileLayout).toBe(false)
    })
  })

  describe('updateCurDvInfo', () => {
    it('should update dvInfo', () => {
      const store = dvMainStore()
      const newDvInfo = {
        id: 'dv-1',
        name: 'Test Dashboard',
        pid: 'folder-1',
        type: 'dashboard'
      }
      store.updateCurDvInfo(newDvInfo)
      expect(store.dvInfo.id).toBe('dv-1')
      expect(store.dvInfo.name).toBe('Test Dashboard')
    })
  })

  describe('Multiple actions', () => {
    it('should handle multiple state changes correctly', () => {
      const store = dvMainStore()
      store.setEditMode('edit')
      store.setInEditorStatus(true)
      store.setMobileInPc(false)
      store.setFullscreenFlag(false)

      expect(store.editMode).toBe('edit')
      expect(store.isInEditor).toBe(true)
      expect(store.mobileInPc).toBe(false)
      expect(store.fullscreenFlag).toBe(false)
    })
  })
})
