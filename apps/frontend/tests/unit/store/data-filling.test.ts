import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useDataFillingStore } from '@/store/modules/data-filling'
import type { DataFillingTreeNode } from '@/views/data-filling/types'

vi.mock('@/api/datafilling', () => ({
  getFormTree: vi.fn()
}))

import { getFormTree } from '@/api/datafilling'

function makeNode(
  id: number | string,
  name: string,
  overrides: Partial<DataFillingTreeNode> = {}
): DataFillingTreeNode {
  return {
    id,
    name,
    pid: 0,
    nodeType: 'folder',
    leaf: false,
    disabled: false,
    children: [],
    ...overrides
  }
}

describe('Data Filling Store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  describe('Initial State', () => {
    it('should have empty formTree', () => {
      const store = useDataFillingStore()
      expect(store.getFormTree).toEqual([])
    })

    it('should have null selectedNodeId', () => {
      const store = useDataFillingStore()
      expect(store.getSelectedNodeId).toBeNull()
    })

    it('should have empty expandedNodeIds', () => {
      const store = useDataFillingStore()
      expect(store.getExpandedNodeIds).toEqual([])
    })
  })

  describe('setFormTree', () => {
    it('should set formTree', () => {
      const store = useDataFillingStore()
      const tree = [makeNode(1, 'Root')]
      store.setFormTree(tree)
      expect(store.getFormTree).toEqual(tree)
    })

    it('should replace existing formTree', () => {
      const store = useDataFillingStore()
      store.setFormTree([makeNode(1, 'Old')])
      store.setFormTree([makeNode(2, 'New')])
      expect(store.getFormTree).toHaveLength(1)
      expect(store.getFormTree[0].name).toBe('New')
    })
  })

  describe('setSelectedNodeId', () => {
    it('should set selectedNodeId', () => {
      const store = useDataFillingStore()
      store.setSelectedNodeId(42)
      expect(store.getSelectedNodeId).toBe(42)
    })

    it('should clear selectedNodeId', () => {
      const store = useDataFillingStore()
      store.setSelectedNodeId(42)
      store.setSelectedNodeId(null)
      expect(store.getSelectedNodeId).toBeNull()
    })
  })

  describe('setExpandedNodeIds', () => {
    it('should set expandedNodeIds', () => {
      const store = useDataFillingStore()
      store.setExpandedNodeIds([1, 2, 3])
      expect(store.getExpandedNodeIds).toEqual([1, 2, 3])
    })
  })

  describe('fetchTree', () => {
    it('should normalize raw API data and set formTree', async () => {
      const rawTree = [
        {
          id: 1,
          name: 'Folder',
          pid: 0,
          nodeType: 'folder',
          leaf: false,
          disabled: false,
          children: [
            { id: 2, name: 'Form', pid: 1, nodeType: 'form', leaf: true, disabled: false }
          ]
        }
      ]
      vi.mocked(getFormTree).mockResolvedValueOnce(rawTree)

      const store = useDataFillingStore()
      const result = await store.fetchTree()

      expect(result).toHaveLength(1)
      expect(result[0].id).toBe(1)
      expect(result[0].name).toBe('Folder')
      expect(result[0].children).toHaveLength(1)
      expect(result[0].children[0].id).toBe(2)
      expect(store.getFormTree).toEqual(result)
    })

    it('should handle non-array API response', async () => {
      vi.mocked(getFormTree).mockResolvedValueOnce(null)

      const store = useDataFillingStore()
      const result = await store.fetchTree()

      expect(result).toEqual([])
      expect(store.getFormTree).toEqual([])
    })

    it('should handle undefined name and default to empty string', async () => {
      vi.mocked(getFormTree).mockResolvedValueOnce([{ id: 1, pid: 0 }])

      const store = useDataFillingStore()
      const result = await store.fetchTree()

      expect(result[0].name).toBe('')
      expect(result[0].nodeType).toBe('')
      expect(result[0].leaf).toBe(false)
      expect(result[0].disabled).toBe(false)
    })
  })

  describe('patchTreeNode (upsert)', () => {
    it('should update existing node by id', () => {
      const store = useDataFillingStore()
      store.setFormTree([makeNode(1, 'Old Name')])
      store.patchTreeNode(makeNode(1, 'New Name'))
      expect(store.getFormTree[0].name).toBe('New Name')
    })

    it('should insert as child when parentId is provided', () => {
      const store = useDataFillingStore()
      store.setFormTree([makeNode(1, 'Parent')])
      store.patchTreeNode(makeNode(2, 'Child'), 1)
      expect(store.getFormTree[0].children).toHaveLength(1)
      expect(store.getFormTree[0].children[0].name).toBe('Child')
    })

    it('should append to root when no parentId and no match', () => {
      const store = useDataFillingStore()
      store.setFormTree([makeNode(1, 'Existing')])
      store.patchTreeNode(makeNode(2, 'New Root'))
      expect(store.getFormTree).toHaveLength(2)
      expect(store.getFormTree[1].name).toBe('New Root')
    })

    it('should preserve children when updating existing node without children', () => {
      const store = useDataFillingStore()
      const parent = makeNode(1, 'Parent', {
        children: [makeNode(2, 'Child')]
      })
      store.setFormTree([parent])
      store.patchTreeNode({
        id: 1,
        name: 'Updated Parent',
        pid: 0,
        nodeType: 'folder',
        leaf: false,
        disabled: false,
        children: undefined
      })
      expect(store.getFormTree[0].name).toBe('Updated Parent')
      expect(store.getFormTree[0].children).toHaveLength(1)
      expect(store.getFormTree[0].children[0].name).toBe('Child')
    })

    it('should update nested node by id', () => {
      const store = useDataFillingStore()
      const root = makeNode(1, 'Root', {
        children: [makeNode(2, 'Nested Old')]
      })
      store.setFormTree([root])
      store.patchTreeNode(makeNode(2, 'Nested New'))
      expect(store.getFormTree[0].children[0].name).toBe('Nested New')
    })
  })

  describe('removeTreeNode', () => {
    it('should remove root-level node', () => {
      const store = useDataFillingStore()
      store.setFormTree([makeNode(1, 'A'), makeNode(2, 'B')])
      store.removeTreeNode(1)
      expect(store.getFormTree).toHaveLength(1)
      expect(store.getFormTree[0].name).toBe('B')
    })

    it('should remove nested node', () => {
      const store = useDataFillingStore()
      store.setFormTree([
        makeNode(1, 'Parent', {
          children: [makeNode(2, 'Child A'), makeNode(3, 'Child B')]
        })
      ])
      store.removeTreeNode(2)
      expect(store.getFormTree[0].children).toHaveLength(1)
      expect(store.getFormTree[0].children[0].name).toBe('Child B')
    })

    it('should clear selectedNodeId when removed node is selected', () => {
      const store = useDataFillingStore()
      store.setFormTree([makeNode(1, 'A')])
      store.setSelectedNodeId(1)
      store.removeTreeNode(1)
      expect(store.getSelectedNodeId).toBeNull()
    })

    it('should remove node id from expandedNodeIds', () => {
      const store = useDataFillingStore()
      store.setFormTree([makeNode(1, 'A'), makeNode(2, 'B')])
      store.setExpandedNodeIds([1, 2])
      store.removeTreeNode(1)
      expect(store.getExpandedNodeIds).toEqual([2])
    })

    it('should not affect selectedNodeId when different node is removed', () => {
      const store = useDataFillingStore()
      store.setFormTree([makeNode(1, 'A'), makeNode(2, 'B')])
      store.setSelectedNodeId(2)
      store.removeTreeNode(1)
      expect(store.getSelectedNodeId).toBe(2)
    })
  })

  describe('clear', () => {
    it('should reset all state', () => {
      const store = useDataFillingStore()
      store.setFormTree([makeNode(1, 'Root')])
      store.setSelectedNodeId(1)
      store.setExpandedNodeIds([1])

      store.clear()

      expect(store.getFormTree).toEqual([])
      expect(store.getSelectedNodeId).toBeNull()
      expect(store.getExpandedNodeIds).toEqual([])
    })
  })
})
