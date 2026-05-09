import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia } from 'pinia'
import { store as sharedPinia } from '@/store'
import { snapshotStore } from '@/store/modules/data-visualization/snapshot'
import { dvMainStore } from '@/store/modules/data-visualization/dvMain'
import eventBus from '@/utils/eventBus'

const { wsCacheMock } = vi.hoisted(() => ({
  wsCacheMock: { set: vi.fn(), get: vi.fn() }
}))

vi.mock('@/utils/utils', async (importOriginal) => {
  const actual = await importOriginal()
  return { ...(actual as any), deepCopy: (obj: any) => JSON.parse(JSON.stringify(obj)) }
})

vi.mock('@/hooks/web/useEmitt', async () => {
  const { createUseEmittModuleMock } = await import('../helpers')
  return createUseEmittModuleMock()
})

vi.mock('@/hooks/web/useCache', () => ({
  useCache: () => ({ wsCache: wsCacheMock })
}))

vi.mock('@/utils/eventBus', async () => {
  const { createEventBusModuleMock } = await import('../helpers')
  return createEventBusModuleMock()
})

vi.mock('@/hooks/web/useEmitt', async () => {
  const { createUseEmittModuleMock } = await import('../helpers')
  return createUseEmittModuleMock()
})

vi.mock('@/hooks/web/useCache', () => ({
  useCache: () => ({
    wsCache: {
      set: vi.fn(),
      get: vi.fn()
    }
  })
}))

vi.mock('@/utils/eventBus', async () => {
  const { createEventBusModuleMock } = await import('../helpers')
  return createEventBusModuleMock()
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

function resetDvMainState() {
  const dv = dvMainStore()
  dv.setDataPrepareState(false)
  dv.setMobileInPc(false)
  dv.setComponentData([])
  dv.setCanvasViewInfo({})
  dv.setCurComponent({ component: null, index: null })
  dv.updateCurDvInfo({
    dataState: null,
    optType: null,
    id: null,
    name: null,
    pid: null,
    status: null,
    selfWatermarkStatus: null,
    watermarkInfo: {},
    type: null,
    mobileLayout: false,
    datasetFolderPid: null,
    datasetFolderName: null,
    creatorName: null,
    updateName: null,
    createTime: null,
    updateTime: null
  })
  dv.setNowPanelTrackInfo({})
  dv.setNowPanelJumpInfoInner({})
}

function setupDvMainState() {
  const dv = dvMainStore()
  dv.setDataPrepareState(true)
  dv.setMobileInPc(false)
  dv.setComponentData([
    { id: 'comp-1', name: 'Component 1' },
    { id: 'comp-2', name: 'Component 2' }
  ])
  dv.setCanvasViewInfo({ 'view-1': { id: 'view-1', title: 'Chart 1' } })
  dv.aceSetCanvasData({ component: {} })
  dv.updateCurDvInfo({
    id: 'dv-001',
    pid: 'folder-001',
    name: 'Test Dashboard',
    type: 'dashboard',
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
  dv.setNowPanelTrackInfo({})
  dv.setNowPanelJumpInfoInner({})
  return dv
}

describe('Snapshot Store', () => {
  beforeEach(() => {
    setActivePinia(sharedPinia)
    resetDvMainState()
    const store = snapshotStore()
    store.initSnapShot()
    vi.clearAllMocks()
  })

  describe('Initial State', () => {
    it('should have snapshotDisableTime of 1', () => {
      const store = snapshotStore()
      expect(store.snapshotDisableTime).toBe(1)
    })

    it('should have styleChangeTimes of -1', () => {
      const store = snapshotStore()
      expect(store.styleChangeTimes).toBe(-1)
    })

    it('should have cacheStyleChangeTimes of 0', () => {
      const store = snapshotStore()
      expect(store.cacheStyleChangeTimes).toBe(0)
    })

    it('should have snapshotCacheTimes of 0', () => {
      const store = snapshotStore()
      expect(store.snapshotCacheTimes).toBe(0)
    })

    it('should have empty snapshotData', () => {
      const store = snapshotStore()
      expect(store.snapshotData).toEqual([])
    })

    it('should have snapshotIndex of -1', () => {
      const store = snapshotStore()
      expect(store.snapshotIndex).toBe(-1)
    })

    it('should have empty cacheViewIdInfo', () => {
      const store = snapshotStore()
      expect(store.cacheViewIdInfo).toEqual({
        snapshotCacheViewCalc: [],
        snapshotCacheViewRender: []
      })
    })
  })

  describe('initSnapShot', () => {
    it('should reset all counters and data', () => {
      const store = snapshotStore()
      store.styleChangeTimes = 5
      store.cacheStyleChangeTimes = 3
      store.snapshotCacheTimes = 2
      store.cacheViewIdInfo = {
        snapshotCacheViewCalc: ['view-1'],
        snapshotCacheViewRender: ['view-2']
      }
      store.snapshotData = [{ componentData: [] }]
      store.snapshotIndex = 0

      store.initSnapShot()

      expect(store.styleChangeTimes).toBe(-1)
      expect(store.cacheStyleChangeTimes).toBe(0)
      expect(store.snapshotCacheTimes).toBe(0)
      expect(store.cacheViewIdInfo).toEqual({
        snapshotCacheViewCalc: [],
        snapshotCacheViewRender: []
      })
      expect(store.snapshotData).toEqual([])
      expect(store.snapshotIndex).toBe(-1)
    })
  })

  describe('recordSnapshot', () => {
    it('should create snapshot from current dvMain state', () => {
      setupDvMainState()
      const store = snapshotStore()
      store.snapshotDisableTime = 0

      store.recordSnapshot()

      expect(store.snapshotIndex).toBe(0)
      expect(store.snapshotData).toHaveLength(1)
      expect(store.snapshotData[0].componentData).toHaveLength(2)
      expect(store.snapshotData[0].dvInfo.id).toBe('dv-001')
    })

    it('should increment snapshotIndex for multiple snapshots', () => {
      setupDvMainState()
      const store = snapshotStore()
      store.snapshotDisableTime = 0

      store.recordSnapshot()
      store.recordSnapshot()

      expect(store.snapshotIndex).toBe(1)
      expect(store.snapshotData).toHaveLength(2)
    })

    it('should clear snapshotCacheTimes after recording', () => {
      setupDvMainState()
      const store = snapshotStore()
      store.snapshotDisableTime = 0
      store.snapshotCacheTimes = 5

      store.recordSnapshot()

      expect(store.snapshotCacheTimes).toBe(0)
    })

    it('should not record when dataPrepareState is false', () => {
      const dv = dvMainStore()
      dv.setDataPrepareState(false)
      dv.setMobileInPc(false)
      const store = snapshotStore()
      store.snapshotDisableTime = 0

      store.recordSnapshot()

      expect(store.snapshotIndex).toBe(-1)
      expect(store.snapshotData).toHaveLength(0)
    })

    it('should not record when mobileInPc is true', () => {
      const dv = dvMainStore()
      dv.setDataPrepareState(true)
      dv.setMobileInPc(true)
      const store = snapshotStore()
      store.snapshotDisableTime = 0

      store.recordSnapshot()

      expect(store.snapshotIndex).toBe(-1)
    })

    it('should not record when within snapshotDisableTime', () => {
      setupDvMainState()
      const store = snapshotStore()
      store.snapshotDisableTime = Date.now() + 30000

      store.recordSnapshot()

      expect(store.snapshotIndex).toBe(-1)
    })

    it('should trim future snapshots when recording after undo', () => {
      setupDvMainState()
      const store = snapshotStore()
      store.snapshotDisableTime = 0

      store.recordSnapshot()
      store.recordSnapshot()
      // Simulate undo by going back
      store.snapshotIndex = 0
      // Record new snapshot
      store.recordSnapshot()

      expect(store.snapshotData).toHaveLength(2)
      expect(store.snapshotIndex).toBe(1)
    })

    it('should store multiple snapshots with correct data', () => {
      setupDvMainState()
      const store = snapshotStore()
      store.snapshotDisableTime = 0

      store.recordSnapshot()
      store.recordSnapshot()

      expect(store.snapshotData).toHaveLength(2)
      expect(store.snapshotData[1].dvInfo.id).toBe('dv-001')
      expect(store.snapshotData[1].componentData).toHaveLength(2)
    })
  })

  describe('snapshotCatchToStore', () => {
    it('should call recordSnapshot when snapshotCacheTimes > 0', () => {
      setupDvMainState()
      const store = snapshotStore()
      store.snapshotDisableTime = 0
      store.snapshotCacheTimes = 3
      const spy = vi.spyOn(store, 'recordSnapshot')

      store.snapshotCatchToStore()

      expect(spy).toHaveBeenCalled()
    })

    it('should not call recordSnapshot when snapshotCacheTimes is 0', () => {
      const store = snapshotStore()
      store.snapshotCacheTimes = 0
      const spy = vi.spyOn(store, 'recordSnapshot')

      store.snapshotCatchToStore()

      expect(spy).not.toHaveBeenCalled()
    })
  })

  describe('recordSnapshotCache', () => {
    it('should increment snapshotCacheTimes when dataPrepareState is true', () => {
      const dv = dvMainStore()
      dv.setDataPrepareState(true)
      const store = snapshotStore()

      store.recordSnapshotCache()

      expect(store.snapshotCacheTimes).toBe(1)
    })

    it('should not increment snapshotCacheTimes when dataPrepareState is false', () => {
      const dv = dvMainStore()
      dv.setDataPrepareState(false)
      const store = snapshotStore()

      store.recordSnapshotCache()

      expect(store.snapshotCacheTimes).toBe(0)
    })

    it('should track calcData view ids', () => {
      const dv = dvMainStore()
      dv.setDataPrepareState(true)
      const store = snapshotStore()

      store.recordSnapshotCache('calcData', 'view-1')

      expect(store.cacheViewIdInfo.snapshotCacheViewCalc).toContain('view-1')
    })

    it('should track renderChart view ids', () => {
      const dv = dvMainStore()
      dv.setDataPrepareState(true)
      const store = snapshotStore()

      store.recordSnapshotCache('renderChart', 'view-2')

      expect(store.cacheViewIdInfo.snapshotCacheViewRender).toContain('view-2')
    })
  })

  describe('undo', () => {
    it('should decrement snapshotIndex and publish snapshot', () => {
      setupDvMainState()
      const store = snapshotStore()
      store.snapshotDisableTime = 0

      store.recordSnapshot()
      store.recordSnapshot()

      store.undo()

      expect(store.snapshotIndex).toBe(0)
      expect(store.styleChangeTimes).toBe(2)
    })

    it('should not undo when snapshotIndex is 0', () => {
      setupDvMainState()
      const store = snapshotStore()
      store.snapshotDisableTime = 0

      store.recordSnapshot()

      store.undo()

      // snapshotIndex was 0, undo should not go below 0
      expect(store.snapshotIndex).toBe(0)
    })

    it('should set snapshotDisableTime to future', () => {
      setupDvMainState()
      const store = snapshotStore()
      store.snapshotDisableTime = 0

      store.recordSnapshot()
      store.recordSnapshot()

      const beforeUndo = Date.now()
      store.undo()

      expect(store.snapshotDisableTime).toBeGreaterThanOrEqual(beforeUndo + 2000)
    })

    it('should emit snapshotChange via eventBus', () => {
      setupDvMainState()
      const store = snapshotStore()
      store.snapshotDisableTime = 0

      store.recordSnapshot()
      store.recordSnapshot()

      store.undo()

      expect(eventBus.emit).toHaveBeenCalledWith('snapshotChange')
    })
  })

  describe('redo', () => {
    it('should increment snapshotIndex and publish snapshot', () => {
      setupDvMainState()
      const store = snapshotStore()
      store.snapshotDisableTime = 0

      store.recordSnapshot()
      store.recordSnapshot()
      store.snapshotIndex = 0

      store.redo()

      expect(store.snapshotIndex).toBe(1)
      expect(store.styleChangeTimes).toBe(2)
    })

    it('should not redo when at the last snapshot', () => {
      setupDvMainState()
      const store = snapshotStore()
      store.snapshotDisableTime = 0

      store.recordSnapshot()

      const originalIndex = store.snapshotIndex
      store.redo()

      expect(store.snapshotIndex).toBe(originalIndex)
    })

    it('should set snapshotDisableTime to future', () => {
      setupDvMainState()
      const store = snapshotStore()
      store.snapshotDisableTime = 0

      store.recordSnapshot()
      store.recordSnapshot()
      store.snapshotIndex = 0

      const beforeRedo = Date.now()
      store.redo()

      expect(store.snapshotDisableTime).toBeGreaterThanOrEqual(beforeRedo + 2000)
    })
  })

  describe('resetStyleChangeTimes', () => {
    it('should reset styleChangeTimes and snapshotCacheTimes to 0', () => {
      const store = snapshotStore()
      store.styleChangeTimes = 5
      store.snapshotCacheTimes = 3

      store.resetStyleChangeTimes()

      expect(store.styleChangeTimes).toBe(0)
      expect(store.snapshotCacheTimes).toBe(0)
    })
  })

  describe('resetSnapshot', () => {
    it('should clear all snapshot data and record a fresh snapshot', () => {
      setupDvMainState()
      const store = snapshotStore()
      store.snapshotDisableTime = 0
      store.styleChangeTimes = 10
      store.cacheStyleChangeTimes = 5
      store.snapshotCacheTimes = 3
      store.cacheViewIdInfo = {
        snapshotCacheViewCalc: ['a'],
        snapshotCacheViewRender: ['b']
      }
      store.snapshotData = [{ componentData: [] }, { componentData: [] }]
      store.snapshotIndex = 1

      store.resetSnapshot()

      expect(store.styleChangeTimes).toBe(0)
      expect(store.cacheStyleChangeTimes).toBe(0)
      expect(store.snapshotCacheTimes).toBe(0)
      expect(store.cacheViewIdInfo).toEqual({
        snapshotCacheViewCalc: [],
        snapshotCacheViewRender: []
      })
      expect(store.snapshotData).toHaveLength(1)
      expect(store.snapshotIndex).toBe(0)
    })
  })

  describe('recordSnapshotCacheWithPositionChange', () => {
    it('should call setLastHiddenComponent and recordSnapshotCache', () => {
      const dv = dvMainStore()
      dv.setDataPrepareState(true)
      const dvSpy = vi.spyOn(dv, 'setLastHiddenComponent')
      const store = snapshotStore()

      store.recordSnapshotCacheWithPositionChange('calcData', 'view-1')

      expect(dvSpy).toHaveBeenCalled()
      expect(store.snapshotCacheTimes).toBe(1)
    })
  })

  describe('full undo/redo cycle', () => {
    it('should support recording, undoing, and redoing correctly', () => {
      setupDvMainState()
      const store = snapshotStore()
      store.snapshotDisableTime = 0

      store.recordSnapshot()
      const dv = dvMainStore()
      dv.setComponentData([
        { id: 'comp-1', name: 'Component 1' },
        { id: 'comp-2', name: 'Component 2' },
        { id: 'comp-3', name: 'Component 3' }
      ])
      store.recordSnapshot()

      // Undo should go back to first snapshot
      store.undo()
      expect(store.snapshotIndex).toBe(0)

      // Redo should go forward to second snapshot
      store.redo()
      expect(store.snapshotIndex).toBe(1)
    })
  })
})
