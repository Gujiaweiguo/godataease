import { throttle } from 'lodash-es'
import type { BusiTreeNode } from '@/models/tree/TreeNode'

type ExpandedKey = string | number

type WsCacheLike = {
  get: (key: string) => unknown
  set: (key: string, value: unknown) => void
}

type FilterKeywordCache = {
  raw: string
  lower: string
  nodeNameLowerMap: WeakMap<Pick<BusiTreeNode, 'name'>, string>
}

export const DATASOURCE_TREE_EXPANDED_KEY_CACHE_KEY = 'TreeExpanded-datasource'

export const getExpandedKeysArray = (expandedKeySet: Set<ExpandedKey>) => {
  return Array.from(expandedKeySet)
}

export const restoreExpandedKeySet = (cachedExpandedKeys: unknown) => {
  if (!Array.isArray(cachedExpandedKeys)) {
    return new Set<ExpandedKey>()
  }
  return new Set(cachedExpandedKeys as ExpandedKey[])
}

export const restoreExpandedKeysFromCache = (wsCache: WsCacheLike) => {
  return restoreExpandedKeySet(wsCache.get(DATASOURCE_TREE_EXPANDED_KEY_CACHE_KEY))
}

export const persistExpandedKeysToCache = (
  wsCache: WsCacheLike,
  expandedKeySet: Set<ExpandedKey>
) => {
  wsCache.set(DATASOURCE_TREE_EXPANDED_KEY_CACHE_KEY, getExpandedKeysArray(expandedKeySet))
}

export const mutateExpandedKeySet = (
  expandedKeySet: Set<ExpandedKey>,
  id: ExpandedKey,
  expand: boolean
) => {
  if (expand) {
    if (expandedKeySet.has(id)) {
      return false
    }
    expandedKeySet.add(id)
    return true
  }

  if (!expandedKeySet.has(id)) {
    return false
  }

  expandedKeySet.delete(id)
  return true
}

export const createExpandedKeyPersistScheduler = (persistFn: () => void) => {
  let pending = false

  return () => {
    if (pending) {
      return
    }
    pending = true
    queueMicrotask(() => {
      pending = false
      persistFn()
    })
  }
}

const collectTreeNodeIds = (treeNodes: BusiTreeNode[]) => {
  const idSet = new Set<ExpandedKey>()
  const walk = (nodes: BusiTreeNode[]) => {
    nodes.forEach(node => {
      if (node?.id !== undefined && node?.id !== null) {
        idSet.add(node.id)
      }
      if (node.children?.length) {
        walk(node.children)
      }
    })
  }
  walk(treeNodes)
  return idSet
}

export const pruneExpandedKeySet = (
  currentExpandedKeySet: Set<ExpandedKey>,
  treeNodes: BusiTreeNode[]
) => {
  if (!currentExpandedKeySet.size) {
    return { nextExpandedKeySet: currentExpandedKeySet, changed: false }
  }

  const validIdSet = collectTreeNodeIds(treeNodes)
  let changed = false
  const nextExpandedKeySet = new Set<ExpandedKey>()

  currentExpandedKeySet.forEach(id => {
    if (validIdSet.has(id)) {
      nextExpandedKeySet.add(id)
      return
    }
    changed = true
  })

  return { nextExpandedKeySet, changed }
}

export const createFilterKeywordCache = (): FilterKeywordCache => ({
  raw: '',
  lower: '',
  nodeNameLowerMap: new WeakMap()
})

const getLowerFilterKeyword = (value: string, filterKeywordCache: FilterKeywordCache) => {
  if (value !== filterKeywordCache.raw) {
    filterKeywordCache.raw = value
    filterKeywordCache.lower = value.toLowerCase()
  }
  return filterKeywordCache.lower
}

const getLowerNodeName = (
  data: Pick<BusiTreeNode, 'name'>,
  filterKeywordCache: FilterKeywordCache
) => {
  const cachedLowerName = filterKeywordCache.nodeNameLowerMap.get(data)
  if (cachedLowerName !== undefined) {
    return cachedLowerName
  }
  const lowerName = data.name.toLowerCase()
  filterKeywordCache.nodeNameLowerMap.set(data, lowerName)
  return lowerName
}

export const filterNodeByKeyword = (
  value: string,
  data: Pick<BusiTreeNode, 'name'>,
  filterKeywordCache: FilterKeywordCache
) => {
  if (!value) {
    return true
  }
  return getLowerNodeName(data, filterKeywordCache).includes(
    getLowerFilterKeyword(value, filterKeywordCache)
  )
}

export const createTreeFilterHandler = (filterFn: (value: string) => void, wait = 300) => {
  return throttle((value: string) => {
    filterFn(value)
  }, wait)
}
