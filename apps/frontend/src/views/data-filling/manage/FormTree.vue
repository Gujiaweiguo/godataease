<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRouter } from 'vue-router_2'
import { ElMessage, ElMessageBox } from 'element-plus-secondary'
import dvFolder from '@/assets/svg/dv-folder.svg'
import dvNewFolder from '@/assets/svg/dv-new-folder.svg'
import icon_deleteTrash_outlined from '@/assets/svg/icon_delete-trash_outlined.svg'
import icon_edit_outlined from '@/assets/svg/icon_edit_outlined.svg'
import icon_fileAdd_outlined from '@/assets/svg/icon_file-add_outlined.svg'
import icon_intoItem_outlined from '@/assets/svg/icon_into-item_outlined.svg'
import icon_rename_outlined from '@/assets/svg/icon_rename_outlined.svg'
import icon_searchOutline_outlined from '@/assets/svg/icon_search-outline_outlined.svg'
import icon_dataset from '@/assets/svg/icon_dataset.svg'
import { createForm, deleteForm, moveForm, renameForm } from '@/api/datafilling'
import { useDataFillingStore } from '@/store/modules/data-filling'
import type { DataFillingNodeId, DataFillingTreeNode } from '@/views/data-filling/types'

type TreeAction = 'create-folder' | 'rename' | 'move'

const emit = defineEmits<{
  (e: 'select', node: DataFillingTreeNode | null): void
  (e: 'loaded', loaded: boolean): void
}>()

const router = useRouter()
const dataFillingStore = useDataFillingStore()

const treeRef = ref()
const loading = ref(false)
const filterText = ref('')
const dialogVisible = ref(false)
const actionType = ref<TreeAction>('create-folder')
const actionNode = ref<DataFillingTreeNode | null>(null)
const currentNode = ref<DataFillingTreeNode | null>(null)
const dialogName = ref('')
const moveTargetId = ref<DataFillingNodeId>(0)

const defaultProps = {
  children: 'children',
  label: 'name'
}

const treeData = computed(() => dataFillingStore.getFormTree)

const dialogTitle = computed(() => {
  if (actionType.value === 'rename') {
    return '重命名'
  }
  if (actionType.value === 'move') {
    return '移动到'
  }
  return '新建文件夹'
})

const isFolderNode = (node: DataFillingTreeNode) => {
  return node.nodeType === 'folder' || (!node.leaf && !!node.children?.length)
}

const isLeafFormNode = (node: DataFillingTreeNode) => {
  return !isFolderNode(node)
}

const normalizeCreatedNode = (
  value: unknown,
  fallback: Pick<DataFillingTreeNode, 'name' | 'pid' | 'nodeType'>
): DataFillingTreeNode | null => {
  if (!value || typeof value !== 'object') {
    return null
  }

  const record = value as Record<string, unknown>
  const id = record.id
  if (typeof id !== 'string' && typeof id !== 'number') {
    return null
  }

  return {
    id,
    name: String(record.name ?? fallback.name),
    pid:
      typeof record.pid === 'string' || typeof record.pid === 'number' ? record.pid : fallback.pid,
    nodeType: String(record.nodeType ?? fallback.nodeType),
    leaf: Boolean(record.leaf ?? false),
    disabled: Boolean(record.disabled ?? false),
    children: []
  }
}

const filterNode = (value: string, data: DataFillingTreeNode) => {
  if (!value) {
    return true
  }
  return data.name.toLowerCase().includes(value.toLowerCase())
}

const collectDescendantIds = (node: DataFillingTreeNode): DataFillingNodeId[] => {
  const descendants: DataFillingNodeId[] = [node.id]
  node.children?.forEach(child => {
    descendants.push(...collectDescendantIds(child))
  })
  return descendants
}

const buildFolderTreeOptions = (
  nodes: DataFillingTreeNode[],
  excludedIds: DataFillingNodeId[]
): DataFillingTreeNode[] => {
  return nodes.reduce<DataFillingTreeNode[]>((result, node) => {
    if (!isFolderNode(node) || excludedIds.includes(node.id)) {
      return result
    }

    result.push({
      ...node,
      children: buildFolderTreeOptions(node.children ?? [], excludedIds)
    })
    return result
  }, [])
}

const moveTreeNode = (
  nodes: DataFillingTreeNode[],
  nodeId: DataFillingNodeId,
  targetPid: DataFillingNodeId
): DataFillingTreeNode[] => {
  const detachNode = (
    source: DataFillingTreeNode[]
  ): { tree: DataFillingTreeNode[]; detached: DataFillingTreeNode | null } => {
    let detached: DataFillingTreeNode | null = null
    const nextTree = source.reduce<DataFillingTreeNode[]>((result, node) => {
      if (node.id === nodeId) {
        detached = {
          ...node,
          pid: targetPid
        }
        return result
      }

      const childResult = detachNode(node.children ?? [])
      if (childResult.detached) {
        detached = childResult.detached
        result.push({
          ...node,
          children: childResult.tree
        })
        return result
      }

      result.push(node)
      return result
    }, [])

    return {
      tree: nextTree,
      detached
    }
  }

  const attachNode = (source: DataFillingTreeNode[], nextNode: DataFillingTreeNode): DataFillingTreeNode[] => {
    if (targetPid === 0 || targetPid === '0') {
      return [...source, nextNode]
    }

    return source.map(node => {
      if (node.id === targetPid) {
        return {
          ...node,
          children: [...(node.children ?? []), nextNode]
        }
      }

      if (!node.children?.length) {
        return node
      }

      return {
        ...node,
        children: attachNode(node.children, nextNode)
      }
    })
  }

  const { tree, detached } = detachNode(nodes)
  if (!detached) {
    return nodes
  }
  return attachNode(tree, detached)
}

const moveTargetOptions = computed(() => {
  const excludedIds = actionNode.value ? collectDescendantIds(actionNode.value) : []
  return buildFolderTreeOptions(treeData.value, excludedIds)
})

const navigateToCreateForm = (pid?: DataFillingNodeId) => {
  const targetPid = typeof pid === 'string' || typeof pid === 'number' ? pid : 0
  router.push({
    path: '/data-filling-editor',
    query: {
      pid: String(targetPid)
    }
  })
}

const navigateToEditForm = (formId: DataFillingNodeId) => {
  router.push({
    path: '/data-filling-editor',
    query: {
      formId: String(formId)
    }
  })
}

const syncSelectedNode = (node: DataFillingTreeNode | null) => {
  currentNode.value = node
  dataFillingStore.setSelectedNodeId(node?.id ?? null)
  emit('select', node)
}

const loadTree = async () => {
  loading.value = true
  try {
    await dataFillingStore.fetchTree()
    emit('loaded', true)
  } catch {
    ElMessage.error('加载表单树失败')
    emit('loaded', true)
  } finally {
    loading.value = false
  }
}

const nodeExpand = (node: DataFillingTreeNode) => {
  const nextIds = Array.from(new Set([...dataFillingStore.getExpandedNodeIds, node.id]))
  dataFillingStore.setExpandedNodeIds(nextIds)
}

const nodeCollapse = (node: DataFillingTreeNode) => {
  dataFillingStore.setExpandedNodeIds(dataFillingStore.getExpandedNodeIds.filter(id => id !== node.id))
}

const handleNodeClick = (node: DataFillingTreeNode) => {
  syncSelectedNode(node)
}

const handleNodeDoubleClick = (node: DataFillingTreeNode) => {
  if (isLeafFormNode(node)) {
    navigateToEditForm(node.id)
  }
}

const openCreateFolder = (parentNode?: DataFillingTreeNode | null) => {
  actionType.value = 'create-folder'
  actionNode.value = parentNode ?? null
  dialogName.value = ''
  moveTargetId.value = 0
  dialogVisible.value = true
}

const openRenameDialog = (node: DataFillingTreeNode) => {
  actionType.value = 'rename'
  actionNode.value = node
  dialogName.value = node.name
  moveTargetId.value = 0
  dialogVisible.value = true
}

const openMoveDialog = (node: DataFillingTreeNode) => {
  actionType.value = 'move'
  actionNode.value = node
  dialogName.value = ''
  moveTargetId.value = node.pid
  dialogVisible.value = true
}

const handleDelete = async (node: DataFillingTreeNode) => {
  const isFolder = isFolderNode(node)
  const hasChildren = Boolean(node.children?.length)
  const message = isFolder
    ? hasChildren
      ? '删除该文件夹将级联删除其下所有内容，是否继续？'
      : '确认删除该文件夹吗？'
    : '确认删除该表单吗？'

  try {
    await ElMessageBox.confirm(message, '提示', {
      type: 'warning',
      confirmButtonType: 'danger',
      confirmButtonText: '确认',
      cancelButtonText: '取消',
      autofocus: false,
      showClose: false
    })
    await deleteForm(Number(node.id))
    dataFillingStore.removeTreeNode(node.id)
    if (currentNode.value?.id === node.id) {
      syncSelectedNode(null)
    }
    ElMessage.success(isFolder ? '文件夹删除成功' : '表单删除成功')
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(isFolder ? '文件夹删除失败' : '表单删除失败')
    }
  }
}

const handleCommand = (command: string, node: DataFillingTreeNode) => {
  if (command === 'new-folder') {
    openCreateFolder(node)
    return
  }
  if (command === 'new-form') {
    navigateToCreateForm(node.id)
    return
  }
  if (command === 'edit') {
    navigateToEditForm(node.id)
    return
  }
  if (command === 'rename') {
    openRenameDialog(node)
    return
  }
  if (command === 'move') {
    openMoveDialog(node)
    return
  }
  if (command === 'delete') {
    void handleDelete(node)
  }
}

const handleDialogConfirm = async () => {
  const activeNode = actionNode.value

  if (actionType.value !== 'move' && !dialogName.value.trim()) {
    ElMessage.warning('请输入名称')
    return
  }

  try {
    if (actionType.value === 'create-folder') {
      const parentId = activeNode?.id ?? 0
      const created = await createForm({
        name: dialogName.value.trim(),
        pid: Number(parentId),
        nodeType: 'folder'
      })
      const createdNode = normalizeCreatedNode(created, {
        name: dialogName.value.trim(),
        pid: parentId,
        nodeType: 'folder'
      })
      if (createdNode) {
        dataFillingStore.patchTreeNode(createdNode, parentId)
        dataFillingStore.setExpandedNodeIds(
          Array.from(new Set([...dataFillingStore.getExpandedNodeIds, parentId]))
        )
      } else {
        await dataFillingStore.fetchTree()
      }
      ElMessage.success('文件夹创建成功')
    }

    if (actionType.value === 'rename' && activeNode) {
      await renameForm({
        id: Number(activeNode.id),
        name: dialogName.value.trim()
      })
      dataFillingStore.patchTreeNode({
        ...activeNode,
        name: dialogName.value.trim()
      })
      if (currentNode.value?.id === activeNode.id) {
        syncSelectedNode({
          ...activeNode,
          name: dialogName.value.trim()
        })
      }
      ElMessage.success('重命名成功')
    }

    if (actionType.value === 'move' && activeNode) {
      const targetPid = moveTargetId.value
      if (targetPid === activeNode.id) {
        ElMessage.warning('不能移动到当前节点下')
        return
      }
      await moveForm({
        id: Number(activeNode.id),
        pid: Number(targetPid)
      })
      dataFillingStore.setFormTree(moveTreeNode(treeData.value, activeNode.id, targetPid))
      if (currentNode.value?.id === activeNode.id) {
        syncSelectedNode({
          ...activeNode,
          pid: targetPid
        })
      }
      ElMessage.success('移动成功')
    }

    dialogVisible.value = false
  } catch {
    ElMessage.error(
      actionType.value === 'create-folder'
        ? '文件夹创建失败'
        : actionType.value === 'rename'
        ? '重命名失败'
        : '移动失败'
    )
  }
}

const folderActionMenu = [
  {
    label: '新建文件夹',
    svgName: dvNewFolder,
    command: 'new-folder'
  },
  {
    label: '新建表单',
    svgName: icon_fileAdd_outlined,
    command: 'new-form'
  },
  {
    label: '重命名',
    svgName: icon_rename_outlined,
    command: 'rename'
  },
  {
    label: '移动到',
    svgName: icon_intoItem_outlined,
    command: 'move'
  },
  {
    label: '删除',
    svgName: icon_deleteTrash_outlined,
    command: 'delete',
    divided: true
  }
]

const formActionMenu = [
  {
    label: '编辑',
    svgName: icon_edit_outlined,
    command: 'edit'
  },
  {
    label: '移动到',
    svgName: icon_intoItem_outlined,
    command: 'move'
  },
  {
    label: '删除',
    svgName: icon_deleteTrash_outlined,
    command: 'delete',
    divided: true
  }
]

watch(filterText, (value: string) => {
  treeRef.value?.filter(value)
})

void loadTree()

defineExpose({
  openCreateFolder
})
</script>

<template>
  <div class="df-form-tree" v-loading="loading">
    <div class="df-form-tree__header">
      <div class="df-form-tree__title-row">
        <span class="df-form-tree__title">Data Filling</span>
        <div class="df-form-tree__actions">
          <el-tooltip effect="dark" content="新建文件夹" placement="top">
            <el-icon class="df-form-tree__action-icon" @click="openCreateFolder()">
              <Icon name="dv-new-folder"><dvNewFolder class="svg-icon" /></Icon>
            </el-icon>
          </el-tooltip>
          <el-tooltip effect="dark" content="新建表单" placement="top">
            <el-icon class="df-form-tree__action-icon" @click="navigateToCreateForm(currentNode?.id)">
              <Icon name="icon_file-add_outlined"><icon_fileAdd_outlined class="svg-icon" /></Icon>
            </el-icon>
          </el-tooltip>
        </div>
      </div>

      <el-input v-model="filterText" clearable class="df-form-tree__search" placeholder="搜索目录或表单">
        <template #prefix>
          <el-icon>
            <Icon name="icon_search-outline_outlined"
              ><icon_searchOutline_outlined class="svg-icon"
            /></Icon>
          </el-icon>
        </template>
      </el-input>
    </div>

    <el-scrollbar class="df-form-tree__scrollbar">
      <el-tree
        ref="treeRef"
        :data="treeData"
        :props="defaultProps"
        node-key="id"
        menu
        highlight-current
        expand-on-click-node
        :default-expanded-keys="dataFillingStore.getExpandedNodeIds"
        :filter-node-method="filterNode"
        @node-expand="nodeExpand"
        @node-collapse="nodeCollapse"
        @node-click="handleNodeClick"
        @node-dblclick="handleNodeDoubleClick"
      >
        <template #default="{ node, data }">
          <span class="custom-tree-node">
            <el-icon style="font-size: 18px">
              <Icon :name="isFolderNode(data) ? 'dv-folder' : 'icon_dataset'"
                ><component :is="isFolderNode(data) ? dvFolder : icon_dataset" class="svg-icon"
              /></Icon>
            </el-icon>
            <span :title="node.label" class="label-tooltip ellipsis">{{ node.label }}</span>
            <div class="icon-more">
              <handle-more
                :menu-list="isFolderNode(data) ? folderActionMenu : formActionMenu"
                @handle-command="command => handleCommand(String(command), data)"
              />
            </div>
          </span>
        </template>
      </el-tree>
    </el-scrollbar>

    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="420px" append-to-body>
      <div v-if="actionType !== 'move'" class="df-form-tree__dialog-body">
        <el-input v-model="dialogName" maxlength="64" show-word-limit placeholder="请输入名称" />
      </div>
      <div v-else class="df-form-tree__dialog-body">
        <el-radio-group v-model="moveTargetId" class="df-form-tree__radio-group">
          <el-radio :label="0">根目录</el-radio>
        </el-radio-group>
        <el-tree-select
          v-model="moveTargetId"
          :data="moveTargetOptions"
          :props="defaultProps"
          check-strictly
          clearable
          default-expand-all
          style="width: 100%"
          placeholder="选择目标文件夹"
        />
      </div>

      <template #footer>
        <span class="dialog-footer">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleDialogConfirm">确认</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<style lang="less" scoped>
.df-form-tree {
  display: flex;
  height: 100%;
  flex-direction: column;

  &__header {
    padding: 0 16px 12px;
  }

  &__title-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding-bottom: 12px;
  }

  &__title {
    color: #1f2329;
    font-size: 16px;
    font-weight: 500;
    line-height: 24px;
  }

  &__actions {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  &__action-icon {
    color: var(--ed-color-primary);
    cursor: pointer;
    font-size: 18px;
  }

  &__search {
    width: 100%;
  }

  &__scrollbar {
    height: calc(100vh - 172px);
    padding: 0 8px 12px;
  }

  &__dialog-body {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  &__radio-group {
    display: flex;
  }
}

.custom-tree-node {
  display: flex;
  width: 100%;
  min-width: 0;
  align-items: center;

  .label-tooltip {
    margin-left: 6px;
    min-width: 0;
    flex: 1;
  }

  .icon-more {
    margin-left: auto;
    display: flex;
    align-items: center;
  }
}
</style>
