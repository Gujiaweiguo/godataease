import { describe, it, expect, vi, beforeEach } from 'vitest'

const { mockComponentData } = vi.hoisted(() => {
  return { mockComponentData: { value: [] as any[] } }
})

vi.mock('pinia', () => ({
  storeToRefs: () => ({ componentData: mockComponentData })
}))

const { mockStore } = vi.hoisted(() => {
  return { mockStore: { componentData: [] } }
})

vi.mock('@/store/modules/data-visualization/dvMain', () => ({
  dvMainStoreWithOut: () => mockStore
}))

vi.mock('@/utils/canvasUtils', () => ({
  isMainCanvas: (id: string) => id === 'canvas-main'
}))

import {
  checkJoinGroup,
  checkJoinTab,
  itemCanvasPathCheck,
  canvasIdMapCheck
} from '@/utils/ShapeUtils'

describe('ShapeUtils', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockComponentData.value = []
  })

  describe('checkJoinGroup', () => {
    it('should return true for non-DeTabs components', () => {
      expect(checkJoinGroup({ component: 'UserView' })).toBe(true)
    })

    it('should return true for DeTabs without Group children', () => {
      const item = {
        component: 'DeTabs',
        propValue: [
          { componentData: [{ component: 'UserView' }] },
          { componentData: [{ component: 'RectShape' }] }
        ]
      }
      expect(checkJoinGroup(item)).toBe(true)
    })

    it('should return false for DeTabs containing Group children', () => {
      const item = {
        component: 'DeTabs',
        propValue: [
          { componentData: [{ component: 'UserView' }] },
          { componentData: [{ component: 'Group' }] }
        ]
      }
      expect(checkJoinGroup(item)).toBe(false)
    })

    it('should return true for DeTabs with empty propValue', () => {
      const item = { component: 'DeTabs', propValue: [] }
      expect(checkJoinGroup(item)).toBe(true)
    })

    it('should return true for DeTabs with undefined propValue', () => {
      const item = { component: 'DeTabs' }
      expect(checkJoinGroup(item)).toBe(true)
    })
  })

  describe('checkJoinTab', () => {
    it('should return true for non-Group components', () => {
      expect(checkJoinTab({ component: 'UserView' })).toBe(true)
    })

    it('should return true for Group without DeTabs children', () => {
      const item = {
        component: 'Group',
        propValue: [{ component: 'UserView' }, { component: 'RectShape' }]
      }
      expect(checkJoinTab(item)).toBe(true)
    })

    it('should return false for Group containing DeTabs children', () => {
      const item = {
        component: 'Group',
        propValue: [{ component: 'UserView' }, { component: 'DeTabs' }]
      }
      expect(checkJoinTab(item)).toBe(false)
    })

    it('should return true for Group with empty propValue', () => {
      const item = { component: 'Group', propValue: [] }
      expect(checkJoinTab(item)).toBe(true)
    })
  })

  describe('canvasIdMapCheck', () => {
    it('should build path map for flat components', () => {
      const item = { id: 'a', component: 'UserView' }
      const pathMap: Record<string, any> = {}
      canvasIdMapCheck(item, null, pathMap)
      expect(pathMap['a']).toBeNull()
    })

    it('should recursively map DeTabs children', () => {
      const tabItem = {
        componentData: [{ id: 'child1', component: 'UserView' }]
      }
      const item = {
        id: 'tab1',
        component: 'DeTabs',
        propValue: [tabItem]
      }
      const pathMap: Record<string, any> = {}
      canvasIdMapCheck(item, null, pathMap)
      expect(pathMap['tab1']).toBeNull()
      expect(pathMap['child1']).toBe(item)
    })

    it('should recursively map Group children', () => {
      const item = {
        id: 'group1',
        component: 'Group',
        propValue: [{ id: 'child2', component: 'UserView' }]
      }
      const pathMap: Record<string, any> = {}
      canvasIdMapCheck(item, null, pathMap)
      expect(pathMap['group1']).toBeNull()
      expect(pathMap['child2']).toBe(item)
    })

    it('should handle nested Group in DeTabs', () => {
      const groupInTab = {
        id: 'group-inner',
        component: 'Group',
        propValue: [{ id: 'leaf', component: 'UserView' }]
      }
      const tabItem = {
        componentData: [groupInTab]
      }
      const tab = {
        id: 'tab-outer',
        component: 'DeTabs',
        propValue: [tabItem]
      }
      const pathMap: Record<string, any> = {}
      canvasIdMapCheck(tab, null, pathMap)
      expect(pathMap['tab-outer']).toBeNull()
      expect(pathMap['group-inner']).toBe(tab)
      expect(pathMap['leaf']).toBe(groupInTab)
    })
  })

  describe('itemCanvasPathCheck', () => {
    it('should return isMainCanvas check for canvas-main type', () => {
      const item = { canvasId: 'canvas-main', id: 'x' }
      expect(itemCanvasPathCheck(item, 'canvas-main')).toBe(true)

      const item2 = { canvasId: 'other-canvas', id: 'y' }
      expect(itemCanvasPathCheck(item2, 'canvas-main')).toBe(false)
    })

    it('should return false for unknown checkType', () => {
      mockComponentData.value = []
      expect(itemCanvasPathCheck({ id: 'x', component: 'UserView' }, 'unknown')).toBe(false)
    })

    it('should detect pTabGroup: child whose parent is Tab and grandparent is Group', () => {
      const child = { id: 'child1', component: 'UserView' }
      const tabItem = { componentData: [child] }
      const tab = { id: 'tab1', component: 'DeTabs', propValue: [tabItem] }
      const group = { id: 'group1', component: 'Group', propValue: [tab] }
      mockComponentData.value = [group]

      expect(itemCanvasPathCheck(child, 'pTabGroup')).toBe(true)
    })

    it('should detect groupInTab: Group whose grandparent is DeTabs', () => {
      const innerGroup = { id: 'g-inner', component: 'Group', propValue: [] }
      const parentGroup = { id: 'g-outer', component: 'Group', propValue: [innerGroup] }
      const tabItem = { componentData: [parentGroup] }
      const tab = { id: 't1', component: 'DeTabs', propValue: [tabItem] }
      mockComponentData.value = [tab]

      expect(itemCanvasPathCheck(innerGroup, 'groupInTab')).toBe(true)
    })

    it('should detect tabInGroup: Tab inside a Group', () => {
      const tab = {
        id: 'tab1',
        component: 'DeTabs',
        propValue: [],
        canvasId: 'group-canvas'
      }
      const group = {
        id: 'group1',
        component: 'Group',
        propValue: [tab],
        canvasId: 'canvas-main'
      }
      mockComponentData.value = [group]

      expect(itemCanvasPathCheck(tab, 'tabInGroup')).toBe(true)
    })

    it('should return false when component has no parent', () => {
      const item = { id: 'orphan', component: 'UserView' }
      mockComponentData.value = []
      expect(itemCanvasPathCheck(item, 'pTabGroup')).toBe(false)
      expect(itemCanvasPathCheck(item, 'groupInTab')).toBe(false)
      expect(itemCanvasPathCheck(item, 'tabInGroup')).toBe(false)
    })
  })
})
