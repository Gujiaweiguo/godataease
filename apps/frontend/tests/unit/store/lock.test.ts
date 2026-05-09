import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia } from 'pinia'
import { store as globalStore } from '@/store/index'
import { dvMainStore } from '@/store/modules/data-visualization/dvMain'
import { lockStore } from '@/store/modules/data-visualization/lock'

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

describe('lock Store', () => {
  beforeEach(() => {
    setActivePinia(globalStore)
    dvMainStore(globalStore).$reset()
    vi.clearAllMocks()
  })

  const setupWithComponent = (component, index = 0) => {
    const dvMain = dvMainStore(globalStore)
    dvMain.setComponentData([component])
    dvMain.setCurComponent({ component, index })
    return dvMain
  }

  describe('lock', () => {
    it('should set isLock to true on current component', () => {
      const component = {
        id: 'comp-1',
        canvasId: 'canvas-main',
        category: 'base',
        style: {},
        isLock: false
      }
      setupWithComponent(component)
      const store = lockStore(globalStore)

      store.lock()

      expect(component.isLock).toBe(true)
    })

    it('should set isLock on explicitly passed component', () => {
      const dvMain = dvMainStore(globalStore)
      dvMain.setComponentData([])
      const store = lockStore(globalStore)
      const otherComponent = { id: 'other', isLock: false }

      store.lock(otherComponent)

      expect(otherComponent.isLock).toBe(true)
    })

    it('should lock Group component and all its children', () => {
      const child1 = { id: 'child-1', isLock: false }
      const child2 = { id: 'child-2', isLock: false }
      const groupComponent = {
        id: 'group-1',
        canvasId: 'canvas-main',
        category: 'base',
        style: {},
        component: 'Group',
        propValue: [child1, child2],
        isLock: false
      }
      setupWithComponent(groupComponent)
      const store = lockStore(globalStore)

      store.lock()

      expect(groupComponent.isLock).toBe(true)
      expect(child1.isLock).toBe(true)
      expect(child2.isLock).toBe(true)
    })

    it('should lock a Group passed as argument with children', () => {
      const dvMain = dvMainStore(globalStore)
      dvMain.setComponentData([])
      const store = lockStore(globalStore)
      const child = { id: 'c', isLock: false }
      const group = { component: 'Group', propValue: [child], isLock: false }

      store.lock(group)

      expect(group.isLock).toBe(true)
      expect(child.isLock).toBe(true)
    })
  })

  describe('unlock', () => {
    it('should set isLock to false on current component', () => {
      const component = {
        id: 'comp-1',
        canvasId: 'canvas-main',
        category: 'base',
        style: {},
        isLock: true
      }
      setupWithComponent(component)
      const store = lockStore(globalStore)

      store.unlock()

      expect(component.isLock).toBe(false)
    })

    it('should set isLock to false on explicitly passed component', () => {
      const dvMain = dvMainStore(globalStore)
      dvMain.setComponentData([])
      const store = lockStore(globalStore)
      const otherComponent = { id: 'other', isLock: true }

      store.unlock(otherComponent)

      expect(otherComponent.isLock).toBe(false)
    })

    it('should unlock Group component and all its children', () => {
      const child1 = { id: 'child-1', isLock: true }
      const child2 = { id: 'child-2', isLock: true }
      const groupComponent = {
        id: 'group-1',
        canvasId: 'canvas-main',
        category: 'base',
        style: {},
        component: 'Group',
        propValue: [child1, child2],
        isLock: true
      }
      setupWithComponent(groupComponent)
      const store = lockStore(globalStore)

      store.unlock()

      expect(groupComponent.isLock).toBe(false)
      expect(child1.isLock).toBe(false)
      expect(child2.isLock).toBe(false)
    })

    it('should unlock a Group passed as argument with children', () => {
      const dvMain = dvMainStore(globalStore)
      dvMain.setComponentData([])
      const store = lockStore(globalStore)
      const child = { id: 'c', isLock: true }
      const group = { component: 'Group', propValue: [child], isLock: true }

      store.unlock(group)

      expect(group.isLock).toBe(false)
      expect(child.isLock).toBe(false)
    })
  })

  describe('lock/unlock round-trip', () => {
    it('should toggle lock state correctly', () => {
      const component = {
        id: 'comp-1',
        canvasId: 'canvas-main',
        category: 'base',
        style: {},
        isLock: false
      }
      setupWithComponent(component)
      const store = lockStore(globalStore)

      store.lock()
      expect(component.isLock).toBe(true)

      store.unlock()
      expect(component.isLock).toBe(false)

      store.lock()
      expect(component.isLock).toBe(true)
    })
  })
})
