import { describe, it, expect, vi } from 'vitest'

// Mock stores and dependencies BEFORE importing canvasUtils
vi.mock('@/store/modules/data-visualization/dvMain', async () => {
  const { createDvMainStoreWithOutModuleMock } = await import('../helpers')
  return createDvMainStoreWithOutModuleMock()
})

vi.mock('@/store/modules/data-visualization/snapshot', async () => {
  const { createSnapshotStoreWithOutModuleMock } = await import('../helpers')
  return createSnapshotStoreWithOutModuleMock()
})

vi.mock('@/store/modules/appearance', async () => {
  const { createAppearanceStoreWithOutModuleMock } = await import('../helpers')
  return createAppearanceStoreWithOutModuleMock()
})

vi.mock('@/hooks/web/useCache', async () => {
  const { createUseCacheModuleMock } = await import('../helpers')
  return createUseCacheModuleMock({
    get: vi.fn(() => '1.0.0')
  })
})

vi.mock('@/hooks/web/useI18n', async () => {
  const { useI18nModuleMock } = await import('../helpers')
  return useI18nModuleMock
})

vi.mock('@/utils/eventBus', async () => {
  const { createEventBusModuleMock } = await import('../helpers')
  return createEventBusModuleMock()
})

vi.mock('@/custom-component/component-list', async () => {
  const { createComponentListModuleMock } = await import('../helpers')
  return createComponentListModuleMock()
})

vi.mock('@/api/visualization/dataVisualization', async () => {
  const { createApiModuleMock } = await import('../helpers')
  return createApiModuleMock([
    'appCanvasNameCheck',
    'checkCanvasChange',
    'decompression',
    'dvNameCheck',
    'findById',
    'findCopyResource',
    'saveCanvas',
    'updateCanvas'
  ])
})

vi.mock('@/api/visualization/linkage', async () => {
  const { createResolvedApiModuleMock } = await import('../helpers')
  return createResolvedApiModuleMock({ getPanelAllLinkageInfo: { data: {} } })
})

vi.mock('@/api/visualization/linkJump', async () => {
  const { createResolvedApiModuleMock } = await import('../helpers')
  return createResolvedApiModuleMock({ queryVisualizationJumpInfo: { data: {} } })
})

vi.mock('@/views/chart/components/editor/util/chart', async () => {
  const { createChartEditorUtilModuleMock } = await import('../helpers')
  return createChartEditorUtilModuleMock()
})

vi.mock('@/views/visualized/data/dataset/form/util', async () => {
  const { createDatasetFormUtilModuleMock } = await import('../helpers')
  return createDatasetFormUtilModuleMock()
})

vi.mock('@/utils/ModelUtil', async () => {
  const { createModelUtilModuleMock } = await import('../helpers')
  return createModelUtilModuleMock()
})

vi.mock('@/views/chart/components/js/formatter', async () => {
  const { createFormatterModuleMock } = await import('../helpers')
  return createFormatterModuleMock()
})

vi.mock('element-plus-secondary', async () => {
  const { elementPlusSecondaryModuleMock } = await import('../helpers')
  return elementPlusSecondaryModuleMock
})

// Import after mocking
import {
  checkAddHttp,
  isMainCanvas,
  isGroupCanvas,
  isTabCanvas,
  isGroupOrTabCanvas,
  isSameCanvas,
  setIdValueTrans,
  filterEmptyFolderTree,
  findParentIdByChildIdRecursive,
  componentPreSort,
  getMapElementIds,
  findComponentIndexByIdWithFilterHidden,
  findComponentIndexById
} from '@/utils/canvasUtils'

describe('canvasUtils', () => {
  describe('checkAddHttp', () => {
    it('should return url as-is if already has http://', () => {
      expect(checkAddHttp('http://example.com')).toBe('http://example.com')
    })

    it('should return url as-is if already has https://', () => {
      expect(checkAddHttp('https://example.com')).toBe('https://example.com')
    })

    it('should add http:// prefix if missing', () => {
      expect(checkAddHttp('example.com')).toBe('http://example.com')
    })

    it('should return empty string as-is', () => {
      expect(checkAddHttp('')).toBe('')
    })

    it('should return undefined as-is', () => {
      expect(checkAddHttp(undefined as any)).toBeUndefined()
    })

    it('should handle uppercase HTTP', () => {
      expect(checkAddHttp('HTTP://example.com')).toBe('HTTP://example.com')
    })

    it('should handle uppercase HTTPS', () => {
      expect(checkAddHttp('HTTPS://example.com')).toBe('HTTPS://example.com')
    })
  })

  describe('isMainCanvas', () => {
    it('should return true for canvas-main', () => {
      expect(isMainCanvas('canvas-main')).toBe(true)
    })

    it('should return false for other canvas IDs', () => {
      expect(isMainCanvas('canvas-1')).toBe(false)
      expect(isMainCanvas('Group-123')).toBe(false)
      expect(isMainCanvas('tab-canvas')).toBe(false)
    })

    it('should return false for empty string', () => {
      expect(isMainCanvas('')).toBe(false)
    })
  })

  describe('isGroupCanvas', () => {
    it('should return true for canvas IDs containing "Group"', () => {
      expect(isGroupCanvas('Group-123')).toBe(true)
      expect(isGroupCanvas('canvas-Group-test')).toBe(true)
      expect(isGroupCanvas('Group')).toBe(true)
    })

    it('should return false for canvas IDs without "Group"', () => {
      expect(isGroupCanvas('canvas-main')).toBe(false)
      expect(isGroupCanvas('tab-123')).toBe(false)
    })

    it('should return falsy for undefined', () => {
      expect(isGroupCanvas(undefined as any)).toBeFalsy()
    })

  })

  describe('isTabCanvas', () => {
    it('should return true for non-main, non-group canvas IDs', () => {
      expect(isTabCanvas('tab-123')).toBe(true)
      expect(isTabCanvas('canvas-tab')).toBe(true)
    })

    it('should return false for main canvas', () => {
      expect(isTabCanvas('canvas-main')).toBe(false)
    })

    it('should return false for group canvas', () => {
      expect(isTabCanvas('Group-123')).toBe(false)
    })
  })

  describe('isGroupOrTabCanvas', () => {
    it('should return true for group canvas', () => {
      expect(isGroupOrTabCanvas('Group-123')).toBe(true)
    })

    it('should return true for tab canvas', () => {
      expect(isGroupOrTabCanvas('tab-456')).toBe(true)
    })

    it('should return false for main canvas', () => {
      expect(isGroupOrTabCanvas('canvas-main')).toBe(false)
    })
  })

  describe('component location helpers', () => {
    it('should return filtered index for visible components', () => {
      const components = [
        { id: 'hidden-1', dashboardHidden: true },
        { id: 'visible-1', dashboardHidden: false },
        { id: 'visible-2' }
      ]

      expect(findComponentIndexByIdWithFilterHidden('visible-1', components)).toBe(0)
      expect(findComponentIndexByIdWithFilterHidden('visible-2', components)).toBe(1)
      expect(findComponentIndexByIdWithFilterHidden('hidden-1', components)).toBe(-1)
    })

    it('should return raw index for components', () => {
      const components = [
        { id: 'first' },
        { id: 'second' },
        { id: 'third' }
      ]

      expect(findComponentIndexById('second', components)).toBe(1)
      expect(findComponentIndexById('missing', components)).toBe(-1)
    })
  })

  describe('isSameCanvas', () => {
    it('should return true when canvas IDs match', () => {
      const item = { canvasId: 'canvas-main' } as any
      expect(isSameCanvas(item, 'canvas-main')).toBe(true)
    })

    it('should return false when canvas IDs do not match', () => {
      const item = { canvasId: 'canvas-main' } as any
      expect(isSameCanvas(item, 'canvas-1')).toBe(false)
    })
  })

  describe('setIdValueTrans', () => {
    it('should replace name placeholders with IDs', () => {
      const content = '[name1] and [name2]'
      const colList = [
        { id: '1', name: 'name1' },
        { id: '2', name: 'name2' }
      ]
      const result = setIdValueTrans('name', 'id', content, colList)
      expect(result).toBe('1 and 2')
    })

    it('should return content as-is if no matches', () => {
      const content = 'no placeholders here'
      const colList = [
        { id: '1', name: 'name1' }
      ]
      const result = setIdValueTrans('name', 'id', content, colList)
      expect(result).toBe('no placeholders here')
    })

    it('should return content as-is if empty', () => {
      expect(setIdValueTrans('name', 'id', '', [])).toBe('')
    })

    it('should return content as-is if null', () => {
      expect(setIdValueTrans('name', 'id', null as any, [])).toBeNull()
    })

    it('should handle partial matches', () => {
      const content = '[name1] and [unknown]'
      const colList = [
        { id: '1', name: 'name1' }
      ]
      const result = setIdValueTrans('name', 'id', content, colList)
      expect(result).toBe('1 and undefined')
    })
  })

  describe('filterEmptyFolderTree', () => {
    it('should keep leaf nodes', () => {
      const nodes = [
        { id: '1', leaf: true, name: 'file1' }
      ]
      const result = filterEmptyFolderTree([...nodes])
      expect(result.length).toBe(1)
      expect(result[0].id).toBe('1')
    })

    it('should remove empty folders', () => {
      const nodes = [
        { id: '1', leaf: false, children: [], name: 'empty-folder' }
      ]
      const result = filterEmptyFolderTree([...nodes])
      expect(result.length).toBe(0)
    })

    it('should keep folders with children', () => {
      const nodes = [
        {
          id: '1',
          leaf: false,
          children: [
            { id: '2', leaf: true, name: 'file1' }
          ],
          name: 'folder'
        }
      ]
      const result = filterEmptyFolderTree(JSON.parse(JSON.stringify(nodes)))
      expect(result.length).toBe(1)
      expect(result[0].children.length).toBe(1)
    })

    it('should handle nested empty folders', () => {
      const nodes = [
        {
          id: '1',
          leaf: false,
          children: [
            { id: '2', leaf: false, children: [], name: 'empty-subfolder' }
          ],
          name: 'folder'
        }
      ]
      const result = filterEmptyFolderTree(JSON.parse(JSON.stringify(nodes)))
      // Parent folder should still be kept since it has children (even if nested empty)
      expect(result.length).toBe(1)
    })

    it('should handle mixed nodes', () => {
      const nodes = [
        { id: '1', leaf: true, name: 'file1' },
        { id: '2', leaf: false, children: [], name: 'empty-folder' },
        {
          id: '3',
          leaf: false,
          children: [
            { id: '4', leaf: true, name: 'file2' }
          ],
          name: 'folder-with-files'
        }
      ]
      const result = filterEmptyFolderTree(JSON.parse(JSON.stringify(nodes)))
      expect(result.length).toBe(2)
    })
  })

  describe('findParentIdByChildIdRecursive', () => {
    it('should find parent ID for child', () => {
      const tree = [
        {
          id: 'parent1',
          children: [
            { id: 'child1', children: [] }
          ]
        }
      ]
      const result = findParentIdByChildIdRecursive(tree, 'child1')
      expect(result).toBe('parent1')
    })

    it('should return null if child not found', () => {
      const tree = [
        {
          id: 'parent1',
          children: [
            { id: 'child1', children: [] }
          ]
        }
      ]
      const result = findParentIdByChildIdRecursive(tree, 'nonexistent')
      expect(result).toBeNull()
    })

    it('should handle deeply nested children', () => {
      const tree = [
        {
          id: 'root',
          children: [
            {
              id: 'level1',
              children: [
                {
                  id: 'level2',
                  children: [
                    { id: 'target', children: [] }
                  ]
                }
              ]
            }
          ]
        }
      ]
      const result = findParentIdByChildIdRecursive(tree, 'target')
      expect(result).toBe('level2')
    })

    it('should handle empty tree', () => {
      const result = findParentIdByChildIdRecursive([], 'any')
      expect(result).toBeNull()
    })

    it('should handle multiple root nodes', () => {
      const tree = [
        {
          id: 'root1',
          children: [
            { id: 'child1', children: [] }
          ]
        },
        {
          id: 'root2',
          children: [
            { id: 'child2', children: [] }
          ]
        }
      ]
      expect(findParentIdByChildIdRecursive(tree, 'child1')).toBe('root1')
      expect(findParentIdByChildIdRecursive(tree, 'child2')).toBe('root2')
    })
  })

  describe('componentPreSort', () => {
    it('should sort components by y position', () => {
      const components = [
        { id: '1', y: 3 },
        { id: '2', y: 1 },
        { id: '3', y: 2 }
      ]
      componentPreSort(components)
      expect(components[0].id).toBe('2')
      expect(components[1].id).toBe('3')
      expect(components[2].id).toBe('1')
    })

    it('should handle empty array', () => {
      const components: any[] = []
      expect(() => componentPreSort(components)).not.toThrow()
    })

    it('should handle null', () => {
      expect(() => componentPreSort(null as any)).not.toThrow()
    })

    it('should sort DeTabs children recursively', () => {
      const components = [
        {
          id: 'tab1',
          y: 1,
          component: 'DeTabs',
          propValue: [
            {
              name: 'tab1',
              componentData: [
                { id: 'inner1', y: 3 },
                { id: 'inner2', y: 1 }
              ]
            }
          ]
        }
      ]
      componentPreSort(components)
      expect(components[0].propValue[0].componentData[0].id).toBe('inner2')
      expect(components[0].propValue[0].componentData[1].id).toBe('inner1')
    })
  })

  describe('getMapElementIds', () => {
    it('should return empty array for empty canvas data', () => {
      expect(getMapElementIds([])).toEqual([])
      expect(getMapElementIds(null as any)).toEqual([])
      expect(getMapElementIds(undefined as any)).toEqual([])
    })

    it('should collect map chart element IDs', () => {
      const canvasData = [
        { id: '1', innerType: 'map' },
        { id: '2', innerType: 'bubble-map' },
        { id: '3', innerType: 'line' } // not a map
      ]
      const result = getMapElementIds(canvasData)
      expect(result).toContain('1')
      expect(result).toContain('2')
      expect(result).not.toContain('3')
    })

    it('should collect map element IDs from DeTabs', () => {
      const canvasData = [
        {
          id: 'tab1',
          component: 'DeTabs',
          propValue: [
            {
              name: 'tab1',
              componentData: [
                { id: 'map-in-tab', innerType: 'heat-map' }
              ]
            }
          ]
        }
      ]
      const result = getMapElementIds(canvasData)
      expect(result).toContain('map-in-tab')
    })

    it('should handle all map types', () => {
      const canvasData = [
        { id: '1', innerType: 'bubble-map' },
        { id: '2', innerType: 'flow-map' },
        { id: '3', innerType: 'heat-map' },
        { id: '4', innerType: 'map' },
        { id: '5', innerType: 'symbolic-map' }
      ]
      const result = getMapElementIds(canvasData)
      expect(result.length).toBe(5)
    })

    it('should ignore non-map chart types', () => {
      const canvasData = [
        { id: '1', innerType: 'line' },
        { id: '2', innerType: 'bar' },
        { id: '3', innerType: 'pie' }
      ]
      const result = getMapElementIds(canvasData)
      expect(result.length).toBe(0)
    })
  })
})
