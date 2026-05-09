import { describe, it, expect } from 'vitest'
import treeSort, { sortPer, treeParentWeight, weightCheckCircle } from '@/utils/treeSortUtils'
import type { BusiTreeNode } from '@/models/tree/TreeNode'

const mkNode = (overrides: Partial<BusiTreeNode> & { id: string | number; name: string }): BusiTreeNode => ({
  pid: 0, weight: 0, extraFlag: 0, extraFlag1: 0, ...overrides
})

describe('treeSortUtils', () => {
  const sampleTree: BusiTreeNode[] = [
    { id: 1, pid: 0, name: 'banana', weight: 10, extraFlag: 0, extraFlag1: 0, children: [] },
    { id: 2, pid: 0, name: 'apple', weight: 5, extraFlag: 0, extraFlag1: 0, children: [] },
    { id: 3, pid: 0, name: 'cherry', weight: 15, extraFlag: 0, extraFlag1: 0, children: [] }
  ]

  describe('sortPer', () => {
    it('should sort by name descending (name_desc)', () => {
      const tree = [
        mkNode({ id: 1, name: 'alpha' }),
        mkNode({ id: 2, name: 'zeta' }),
        mkNode({ id: 3, name: 'mid' })
      ]
      sortPer(tree, 'name_desc')
      expect(tree[0].name).toBe('zeta')
      expect(tree[1].name).toBe('mid')
      expect(tree[2].name).toBe('alpha')
    })

    it('should sort by name ascending (name_asc)', () => {
      const tree = [
        mkNode({ id: 1, name: 'zeta' }),
        mkNode({ id: 2, name: 'alpha' }),
        mkNode({ id: 3, name: 'mid' })
      ]
      sortPer(tree, 'name_asc')
      expect(tree[0].name).toBe('alpha')
      expect(tree[1].name).toBe('mid')
      expect(tree[2].name).toBe('zeta')
    })

    it('should reverse array for time_asc', () => {
      const tree = [
        mkNode({ id: 1, name: 'first' }),
        mkNode({ id: 2, name: 'second' }),
        mkNode({ id: 3, name: 'third' })
      ]
      sortPer(tree, 'time_asc')
      expect(tree[0].name).toBe('third')
      expect(tree[2].name).toBe('first')
    })

    it('should do nothing for unknown sortType', () => {
      const tree = [
        mkNode({ id: 1, name: 'a' }),
        mkNode({ id: 2, name: 'b' })
      ]
      sortPer(tree, 'unknown')
      expect(tree[0].name).toBe('a')
      expect(tree[1].name).toBe('b')
    })

    it('should handle empty array', () => {
      const tree: BusiTreeNode[] = []
      expect(() => sortPer(tree, 'name_asc')).not.toThrow()
    })
  })

  describe('treeSort (default export)', () => {
    it('should return a deep clone, not mutate original', () => {
      const original = JSON.parse(JSON.stringify(sampleTree))
      const result = treeSort(sampleTree, 'name_asc')
      expect(result).not.toBe(sampleTree)
      expect(sampleTree).toEqual(original)
    })

    it('should sort all levels recursively with name_asc', () => {
      const tree = [
        mkNode({
          id: 1,
          name: 'zParent',
          weight: 1,
          children: [
            mkNode({ id: 11, name: 'zChild', weight: 2 }),
            mkNode({ id: 12, name: 'aChild', weight: 3 })
          ]
        }),
        mkNode({ id: 2, name: 'aParent', weight: 4, children: [] })
      ]
      const result = treeSort(tree, 'name_asc')
      expect(result[0].name).toBe('aParent')
      expect(result[1].name).toBe('zParent')
      expect(result[1].children[0].name).toBe('aChild')
      expect(result[1].children[1].name).toBe('zChild')
    })

    it('should sort all levels recursively with name_desc', () => {
      const tree = [
        mkNode({
          id: 1,
          name: 'aParent',
          weight: 1,
          children: [
            mkNode({ id: 11, name: 'aChild', weight: 2 }),
            mkNode({ id: 12, name: 'zChild', weight: 3 })
          ]
        })
      ]
      const result = treeSort(tree, 'name_desc')
      expect(result[0].children[0].name).toBe('zChild')
      expect(result[0].children[1].name).toBe('aChild')
    })

    it('should return cloned tree as-is for unknown sortType', () => {
      const result = treeSort(sampleTree, 'unknown')
      expect(result.map(n => n.id)).toEqual([1, 2, 3])
    })

    it('should handle empty tree', () => {
      const result = treeSort([], 'name_asc')
      expect(result).toEqual([])
    })
  })

  describe('treeParentWeight', () => {
    it('should build weight map for flat tree', () => {
      const tree = [
        mkNode({ id: 'a', name: 'A', weight: 1 }),
        mkNode({ id: 'b', name: 'B', weight: 2 })
      ]
      const result = treeParentWeight(tree, 0)
      expect(result).toEqual({ a: 0, b: 0 })
    })

    it('should build weight map recursively for nested tree', () => {
      const tree = [
        mkNode({
          id: 'parent',
          name: 'Parent',
          weight: 10,
          children: [
            mkNode({ id: 'child1', name: 'C1', weight: 20 }),
            mkNode({ id: 'child2', name: 'C2', weight: 30 })
          ]
        })
      ]
      const result = treeParentWeight(tree, 0)
      expect(result).toEqual({
        parent: 0,
        child1: 10,
        child2: 10
      })
    })

    it('should handle deeply nested tree', () => {
      const tree = [
        mkNode({
          id: 'root',
          name: 'Root',
          weight: 100,
          children: [
            mkNode({
              id: 'mid',
              name: 'Mid',
              weight: 200,
              children: [mkNode({ id: 'leaf', name: 'Leaf', weight: 300 })]
            })
          ]
        })
      ]
      const result = treeParentWeight(tree, 0)
      expect(result).toEqual({
        root: 0,
        mid: 100,
        leaf: 200
      })
    })

    it('should handle empty tree', () => {
      const result = treeParentWeight([], 0)
      expect(result).toEqual({})
    })
  })

  describe('weightCheckCircle', () => {
    it('should populate pWeightResult with parent weights', () => {
      const tree = [
        mkNode({ id: 1, name: 'A', weight: 10 }),
        mkNode({ id: 2, name: 'B', weight: 20 })
      ]
      const result = {}
      weightCheckCircle(tree, result, 99)
      expect(result).toEqual({ 1: 99, 2: 99 })
    })

    it('should use node weight as pWeight for children', () => {
      const tree = [
        mkNode({
          id: 'p',
          name: 'P',
          weight: 42,
          children: [mkNode({ id: 'c', name: 'C', weight: 0 })]
        })
      ]
      const result = {}
      weightCheckCircle(tree, result, 0)
      expect(result['p']).toBe(0)
      expect(result['c']).toBe(42)
    })
  })
})
