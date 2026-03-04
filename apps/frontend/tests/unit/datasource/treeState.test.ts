import { describe, it, expect, vi, afterEach } from 'vitest'
import {
  createExpandedKeyPersistScheduler,
  createFilterKeywordCache,
  createTreeFilterHandler,
  DATASOURCE_TREE_EXPANDED_KEY_CACHE_KEY,
  filterNodeByKeyword,
  mutateExpandedKeySet,
  persistExpandedKeysToCache,
  pruneExpandedKeySet,
  restoreExpandedKeysFromCache
} from '@/views/visualized/data/datasource/treeState'

describe('datasource tree state helpers', () => {
  afterEach(() => {
    vi.useRealTimers()
  })

  it('persists and restores expanded keys with datasource cache key', () => {
    const wsCache = {
      get: vi.fn().mockReturnValue(['folder-1', 2001]),
      set: vi.fn()
    }

    const restored = restoreExpandedKeysFromCache(wsCache)

    expect(wsCache.get).toHaveBeenCalledWith(DATASOURCE_TREE_EXPANDED_KEY_CACHE_KEY)
    expect(Array.from(restored)).toEqual(['folder-1', 2001])

    persistExpandedKeysToCache(wsCache, new Set(['folder-1', 'leaf-2']))

    expect(wsCache.set).toHaveBeenCalledWith(DATASOURCE_TREE_EXPANDED_KEY_CACHE_KEY, [
      'folder-1',
      'leaf-2'
    ])
  })

  it('prunes stale expanded keys when datasource tree refreshes', () => {
    const currentExpandedKeySet = new Set<string | number>(['folder-1', 'leaf-1', 'removed-id'])
    const treeNodes = [
      {
        id: 'folder-1',
        pid: '0',
        name: 'Folder',
        weight: 7,
        extraFlag: 0,
        extraFlag1: 0,
        children: [
          {
            id: 'leaf-1',
            pid: 'folder-1',
            name: 'Leaf',
            weight: 7,
            extraFlag: 0,
            extraFlag1: 0,
            leaf: true
          }
        ]
      }
    ]

    const { nextExpandedKeySet, changed } = pruneExpandedKeySet(currentExpandedKeySet, treeNodes)

    expect(changed).toBe(true)
    expect(Array.from(nextExpandedKeySet)).toEqual(['folder-1', 'leaf-1'])
  })

  it('mutates expanded set in-place for expand and collapse', () => {
    const expandedSet = new Set<string | number>(['folder-1'])

    const expandChanged = mutateExpandedKeySet(expandedSet, 'leaf-1', true)
    const collapseChanged = mutateExpandedKeySet(expandedSet, 'folder-1', false)

    expect(expandChanged).toBe(true)
    expect(collapseChanged).toBe(true)
    expect(Array.from(expandedSet)).toEqual(['leaf-1'])
  })

  it('batches persist calls in the same microtask tick', async () => {
    const persistFn = vi.fn()
    const schedulePersist = createExpandedKeyPersistScheduler(persistFn)

    schedulePersist()
    schedulePersist()
    schedulePersist()

    expect(persistFn).toHaveBeenCalledTimes(0)

    await Promise.resolve()

    expect(persistFn).toHaveBeenCalledTimes(1)

    schedulePersist()
    await Promise.resolve()

    expect(persistFn).toHaveBeenCalledTimes(2)
  })

  it('throttles tree search filter invocation to 300ms window', () => {
    vi.useFakeTimers()
    const filterFn = vi.fn()
    const filterHandler = createTreeFilterHandler(filterFn, 300)

    filterHandler('mysql')
    filterHandler('mysql-1')
    filterHandler('mysql-12')

    expect(filterFn).toHaveBeenCalledTimes(1)
    expect(filterFn).toHaveBeenNthCalledWith(1, 'mysql')

    vi.advanceTimersByTime(300)

    expect(filterFn).toHaveBeenCalledTimes(2)
    expect(filterFn).toHaveBeenNthCalledWith(2, 'mysql-12')
  })

  it('filters datasource node name case-insensitively with keyword cache', () => {
    const filterKeywordCache = createFilterKeywordCache()
    const datasourceNode = {
      id: 'leaf-1',
      pid: 'folder-1',
      name: 'MySQL_Main',
      weight: 7,
      extraFlag: 0,
      extraFlag1: 0,
      leaf: true
    }

    expect(filterNodeByKeyword('mysql', datasourceNode, filterKeywordCache)).toBe(true)
    expect(filterNodeByKeyword('MYSQL', datasourceNode, filterKeywordCache)).toBe(true)
    expect(filterNodeByKeyword('postgres', datasourceNode, filterKeywordCache)).toBe(false)
    expect(filterKeywordCache.nodeNameLowerMap.has(datasourceNode)).toBe(true)
  })
})
