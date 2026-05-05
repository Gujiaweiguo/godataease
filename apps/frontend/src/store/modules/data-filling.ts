import { defineStore } from 'pinia'
import { getFormTree } from '@/api/datafilling'
import { store } from '@/store'
import type { DataFillingNodeId, DataFillingTreeNode } from '@/views/data-filling/types'

interface DataFillingState {
  formTree: DataFillingTreeNode[]
  selectedNodeId: DataFillingNodeId | null
  expandedNodeIds: DataFillingNodeId[]
}

const normalizeTree = (nodes: unknown): DataFillingTreeNode[] => {
  if (!Array.isArray(nodes)) {
    return []
  }

  return nodes.map(node => {
    const current = node as Record<string, unknown>
    return {
      id: current.id as DataFillingNodeId,
      name: String(current.name ?? ''),
      pid: (current.pid ?? 0) as DataFillingNodeId,
      nodeType: String(current.nodeType ?? ''),
      leaf: Boolean(current.leaf),
      disabled: Boolean(current.disabled),
      children: normalizeTree(current.children)
    }
  })
}

const removeTreeNode = (
  nodes: DataFillingTreeNode[],
  nodeId: DataFillingNodeId
): DataFillingTreeNode[] => {
  return nodes
    .filter(node => node.id !== nodeId)
    .map(node => ({
      ...node,
      children: removeTreeNode(node.children ?? [], nodeId)
    }))
}

const upsertTreeNode = (
  nodes: DataFillingTreeNode[],
  nextNode: DataFillingTreeNode,
  parentId?: DataFillingNodeId | null
): DataFillingTreeNode[] => {
  const updateExistingNode = (
    source: DataFillingTreeNode[]
  ): [DataFillingTreeNode[], boolean] => {
    let matched = false

    const nextNodes = source.map(node => {
      if (node.id === nextNode.id) {
        matched = true
        return {
          ...node,
          ...nextNode,
          children: nextNode.children ?? node.children ?? []
        }
      }

      const [nextChildren, childMatched] = updateExistingNode(node.children ?? [])
      if (!childMatched) {
        return node
      }

      matched = true
      return {
        ...node,
        children: nextChildren
      }
    })

    return [nextNodes, matched]
  }

  const [updatedNodes, matched] = updateExistingNode(nodes)
  if (matched) {
    return updatedNodes
  }

  if (parentId == null) {
    return [...updatedNodes, nextNode]
  }

  return updatedNodes.map(node => {
    if (node.id !== parentId) {
      return node
    }

    return {
      ...node,
      children: [...(node.children ?? []), nextNode]
    }
  })
}

export const useDataFillingStore = defineStore('data-filling', {
  state: (): DataFillingState => ({
    formTree: [],
    selectedNodeId: null,
    expandedNodeIds: []
  }),
  getters: {
    getFormTree(state): DataFillingTreeNode[] {
      return state.formTree
    },
    getSelectedNodeId(state): DataFillingNodeId | null {
      return state.selectedNodeId
    },
    getExpandedNodeIds(state): DataFillingNodeId[] {
      return state.expandedNodeIds
    }
  },
  actions: {
    async fetchTree() {
      const tree = await getFormTree()
      this.formTree = normalizeTree(tree)
      return this.formTree
    },
    setFormTree(tree: DataFillingTreeNode[]) {
      this.formTree = tree
    },
    setSelectedNodeId(nodeId: DataFillingNodeId | null) {
      this.selectedNodeId = nodeId
    },
    setExpandedNodeIds(nodeIds: DataFillingNodeId[]) {
      this.expandedNodeIds = nodeIds
    },
    patchTreeNode(node: DataFillingTreeNode, parentId?: DataFillingNodeId | null) {
      this.formTree = upsertTreeNode(this.formTree, node, parentId)
    },
    removeTreeNode(nodeId: DataFillingNodeId) {
      this.formTree = removeTreeNode(this.formTree, nodeId)
      if (this.selectedNodeId === nodeId) {
        this.selectedNodeId = null
      }
      this.expandedNodeIds = this.expandedNodeIds.filter((id: DataFillingNodeId) => id !== nodeId)
    },
    clear() {
      this.formTree = []
      this.selectedNodeId = null
      this.expandedNodeIds = []
    }
  }
})

export const useDataFillingStoreWithOut = () => useDataFillingStore(store)
