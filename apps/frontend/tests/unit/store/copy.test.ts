import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { setActivePinia } from 'pinia'
import { store as globalStore } from '@/store/index'
import { dvMainStore } from '@/store/modules/data-visualization/dvMain'
import { copyStore } from '@/store/modules/data-visualization/copy'
import { composeStore } from '@/store/modules/data-visualization/compose'

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
  return {
    ...createDataVisualizationEditorUtilModuleMock(),
    BASE_THEMES: {
      dark: { background: '#1a1a1a' },
      light: { background: '#ffffff' }
    }
  }
})

vi.mock('@/views/chart/components/js/panel', async () => {
  const { createPanelModuleMock } = await import('../helpers')
  return createPanelModuleMock()
})

vi.mock('@/custom-component/component-list', async () => {
  const { createComponentListModuleMock } = await import('../helpers')
  return {
    ...createComponentListModuleMock(),
    commonStyle: { rotate: 0, opacity: 1 },
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

vi.mock('@/utils/generateID', () => ({
  generateID: vi.fn(() => 'new-generated-id')
}))

vi.mock('@/utils/canvasStyle', () => ({
  adaptCurThemeCommonStyle: vi.fn()
}))

vi.mock('@/utils/canvasUtils', () => ({
  maxYComponentCount: vi.fn(() => 20)
}))

vi.mock('@/utils/decomposeComponent', () => ({
  default: vi.fn()
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

vi.mock('@/utils/utils', async (importOriginal) => {
  const actual = await importOriginal()
  return { ...(actual as any), deepCopy: (obj: any) => JSON.parse(JSON.stringify(obj)) }
})

vi.mock('@/plugins/vue-i18n', () => ({
  i18n: { global: { t: (key: string) => key } }
}))

vi.mock('@/config/axios/service', () => ({
  PATH_URL: '/api',
  service: vi.fn()
}))

describe('Copy Store', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.useFakeTimers()
    setActivePinia(globalStore)
    dvMainStore(globalStore).$reset()
    composeStore(globalStore).$reset()
    const store = copyStore(globalStore)
    store.$reset()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  const setupWithComponent = (component, index = 0) => {
    const dvMain = dvMainStore(globalStore)
    dvMain.setComponentData([{ ...component }])
    dvMain.setCurComponent({ component, index })
    dvMain.updateCurDvInfo({
      id: 'dv-001',
      pid: 'folder-001',
      name: 'Test Dashboard',
      type: 'dataV',
      dataState: null,
      optType: null,
      status: null,
      selfWatermarkStatus: null,
      watermarkInfo: {},
      mobileLayout: false,
      datasetFolderPid: null,
      datasetFolderName: null,
      creatorName: null,
      updateName: null,
      createTime: null,
      updateTime: null
    })
    dvMain.setCanvasStyle({
      component: {},
      width: 1920,
      height: 1080,
      scale: 100,
      popupAvailable: false
    })
    return dvMain
  }

  describe('Initial State', () => {
    it('should have empty copyDataArray', () => {
      const store = copyStore(globalStore)
      expect(store.copyDataArray).toEqual([])
    })

    it('should have null copyData', () => {
      const store = copyStore(globalStore)
      expect(store.copyData).toBeNull()
    })

    it('should have isCut as false', () => {
      const store = copyStore(globalStore)
      expect(store.isCut).toBe(false)
    })
  })

  describe('copy', () => {
    it('should copy current component when curComponent exists and is not GroupArea', () => {
      const component = {
        id: 'comp-1',
        component: 'UserView',
        canvasId: 'canvas-main',
        style: { top: 100, left: 200 }
      }
      setupWithComponent(component)

      const store = copyStore(globalStore)
      store.copy()

      expect(store.copyData).toBeDefined()
      expect(store.copyData.data).toHaveLength(1)
      expect(store.copyData.data[0].id).toBe('comp-1')
      expect(store.isCut).toBe(false)
    })

    it('should not copy GroupArea components', () => {
      const component = {
        id: 'ga-1',
        component: 'GroupArea',
        canvasId: 'canvas-main',
        style: {}
      }
      setupWithComponent(component)

      const composeStr = composeStore(globalStore)
      composeStr.setAreaData({
        style: { top: 0, left: 0, width: 0, height: 0 },
        components: []
      })

      const store = copyStore(globalStore)
      store.copy()

      expect(store.copyData).toBeNull()
    })

    it('should copy area components when curComponent is null', () => {
      const dvMain = dvMainStore(globalStore)
      dvMain.setComponentData([])
      dvMain.setCurComponent({ component: null, index: null })

      const composeStr = composeStore(globalStore)
      const areaComponents = [
        { id: 'a1', component: 'UserView', canvasId: 'canvas-main', style: {} },
        { id: 'a2', component: 'UserView', canvasId: 'canvas-main', style: {} }
      ]
      composeStr.setAreaData({
        style: { top: 0, left: 0, width: 0, height: 0 },
        components: areaComponents
      })

      const store = copyStore(globalStore)
      store.copy()

      expect(store.copyData).toBeDefined()
      expect(store.copyData.data).toHaveLength(2)
    })
  })

  describe('cut', () => {
    it('should copy and delete current component', () => {
      const component = {
        id: 'comp-1',
        component: 'UserView',
        canvasId: 'canvas-main',
        style: { top: 100, left: 200 }
      }
      const dvMain = setupWithComponent(component)

      const store = copyStore(globalStore)
      store.cut()

      expect(store.copyData).toBeDefined()
      expect(store.copyData.data).toHaveLength(1)
      expect(store.isCut).toBe(true)
      expect(dvMain.componentData).toHaveLength(0)
    })

    it('should not cut GroupArea components when no area components exist', () => {
      const component = {
        id: 'ga-1',
        component: 'GroupArea',
        canvasId: 'canvas-main',
        style: {}
      }
      const dvMain = setupWithComponent(component)

      const composeStr = composeStore(globalStore)
      composeStr.setAreaData({
        style: { top: 0, left: 0, width: 0, height: 0 },
        components: []
      })

      const store = copyStore(globalStore)
      store.cut()

      expect(store.copyData).toBeNull()
      expect(dvMain.componentData).toHaveLength(1)
    })
  })

  describe('copyDataInfo', () => {
    it('should set copyData with deep copy of data', () => {
      const dvMain = dvMainStore(globalStore)
      dvMain.setCurComponent({ component: null, index: 2 })

      const store = copyStore(globalStore)
      const data = [{ id: 'c1', name: 'test' }]
      store.copyDataInfo(data)

      expect(store.copyData).toBeDefined()
      expect(store.copyData.data).toHaveLength(1)
      expect(store.copyData.data[0].id).toBe('c1')
    })
  })

  describe('copyDataArrayInfo', () => {
    it('should deep copy area components into copyDataArray', () => {
      const composeStr = composeStore(globalStore)
      const areaComponents = [
        { id: 'a1', component: 'UserView', style: {} },
        { id: 'a2', component: 'UserView', style: {} }
      ]
      composeStr.setAreaData({
        style: { top: 0, left: 0, width: 0, height: 0 },
        components: areaComponents
      })

      const store = copyStore(globalStore)
      store.copyDataArrayInfo()

      expect(store.copyDataArray).toHaveLength(2)
      expect(store.copyDataArray[0].id).toBe('a1')
    })
  })

  describe('restorePreCutData', () => {
    it('should not restore when isCut is false', () => {
      const dvMain = dvMainStore(globalStore)
      dvMain.setComponentData([])

      const store = copyStore(globalStore)
      store.isCut = false
      store.copyData = {
        data: [{ id: 'restored-1' }],
        index: 0
      }

      store.restorePreCutData()

      expect(dvMain.componentData).toHaveLength(0)
    })

    it('should not restore when copyData is null', () => {
      const dvMain = dvMainStore(globalStore)
      dvMain.setComponentData([])

      const store = copyStore(globalStore)
      store.isCut = true
      store.copyData = null

      store.restorePreCutData()

      expect(dvMain.componentData).toHaveLength(0)
    })
  })

  describe('paste', () => {
    it('should do nothing when copyData is null', () => {
      const store = copyStore(globalStore)
      store.copyData = null

      store.paste()

      expect(store.copyData).toBeNull()
    })
  })

  describe('deepCopyTabItemHelper', () => {
    it('should deep copy tab items with new canvasId', async () => {
      const { deepCopyTabItemHelper } = await import(
        '@/store/modules/data-visualization/copy'
      )
      const idMap = {}
      const tabComponentData = [
        { id: 'tc-1', canvasId: 'old-canvas', style: {} }
      ]

      const result = deepCopyTabItemHelper('new-canvas', tabComponentData, idMap)

      expect(result).toHaveLength(1)
      expect(result[0].canvasId).toBe('new-canvas')
      expect(result[0].id).toBe('new-generated-id')
    })
  })
})
