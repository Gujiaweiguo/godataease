<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router_2'
import icon_add_outlined from '@/assets/svg/icon_add_outlined.svg'
import icon_edit_outlined from '@/assets/svg/icon_edit_outlined.svg'
import EmptyBackground from '@/components/empty-background/src/EmptyBackground.vue'
import { useMoveLine } from '@/hooks/web/useMoveLine'
import { useAppStoreWithOut } from '@/store/modules/app'
import { useDataFillingStore } from '@/store/modules/data-filling'
import type { DataFillingNodeId, DataFillingTreeNode } from '@/views/data-filling/types'
import ArrowSide from '@/views/common/DeResourceArrow.vue'
import FormTree from './FormTree.vue'
import DataPage from './data/index.vue'
import TaskPage from './task/index.vue'

const router = useRouter()
const appStore = useAppStoreWithOut()
const dataFillingStore = useDataFillingStore()
const { width, node } = useMoveLine('DATA-FILLING')

const treeLoaded = ref(false)
const sideTreeStatus = ref(true)
const rightTab = ref<'data' | 'tasks'>('data')
const formTreeRef = ref<{ openCreateFolder: (parentNode?: DataFillingTreeNode | null) => void } | null>(null)

const changeSideTreeStatus = (value: boolean) => {
  sideTreeStatus.value = value
}

const mouseenter = () => {
  appStore.setArrowSide(true)
}

const mouseleave = () => {
  appStore.setArrowSide(false)
}

const isFolderNode = (nodeValue: DataFillingTreeNode | null | undefined) => {
  if (!nodeValue) {
    return false
  }
  return nodeValue.nodeType === 'folder' || (!nodeValue.leaf && !!nodeValue.children?.length)
}

const findNodeById = (
  nodes: DataFillingTreeNode[],
  nodeId: DataFillingNodeId | null
): DataFillingTreeNode | null => {
  if (nodeId == null) {
    return null
  }

  for (const nodeValue of nodes) {
    if (nodeValue.id === nodeId) {
      return nodeValue
    }
    const found = findNodeById(nodeValue.children ?? [], nodeId)
    if (found) {
      return found
    }
  }

  return null
}

const selectedNode = computed(() => {
  return findNodeById(dataFillingStore.getFormTree, dataFillingStore.getSelectedNodeId)
})

const selectedLeaf = computed(() => {
  if (!selectedNode.value || isFolderNode(selectedNode.value)) {
    return null
  }
  return selectedNode.value
})

const currentCreatePid = computed(() => {
  if (selectedNode.value) {
    return isFolderNode(selectedNode.value) ? selectedNode.value.id : selectedNode.value.pid
  }
  return 0
})

const openCreateFolder = () => {
  formTreeRef.value?.openCreateFolder(null)
}

const navigateToCreateForm = () => {
  router.push({
    path: '/data-filling-editor',
    query: {
      pid: String(currentCreatePid.value)
    }
  })
}

const navigateToEditForm = () => {
  if (!selectedLeaf.value) {
    return
  }
  router.push({
    path: '/data-filling-editor',
    query: {
      formId: String(selectedLeaf.value.id)
    }
  })
}
</script>

<template>
  <div class="df-manage" v-loading="!treeLoaded">
    <ArrowSide
      :style="{ left: (sideTreeStatus ? width - 12 : 0) + 'px' }"
      @change-side-tree-status="changeSideTreeStatus"
      :isInside="!sideTreeStatus"
    />

    <el-aside
      ref="node"
      class="resource-area"
      :class="{ retract: !sideTreeStatus }"
      :style="{ width: width + 'px' }"
      @mouseenter="mouseenter"
      @mouseleave="mouseleave"
    >
      <ArrowSide
        :isInside="!sideTreeStatus"
        :style="{ left: (sideTreeStatus ? width - 12 : 0) + 'px' }"
        @change-side-tree-status="changeSideTreeStatus"
      />

      <div class="resource-tree">
        <FormTree ref="formTreeRef" @loaded="treeLoaded = $event" />
      </div>
    </el-aside>

    <div class="df-manage__content">
      <div class="df-manage__toolbar">
        <span class="df-manage__toolbar-title">Data Filling 管理</span>
        <div class="df-manage__toolbar-actions">
          <el-button @click="openCreateFolder">
            <template #icon>
              <Icon name="icon_add_outlined"><icon_add_outlined class="svg-icon" /></Icon>
            </template>
            新建文件夹
          </el-button>
          <el-button type="primary" @click="navigateToCreateForm">
            <template #icon>
              <Icon name="icon_add_outlined"><icon_add_outlined class="svg-icon" /></Icon>
            </template>
            新建表单
          </el-button>
        </div>
      </div>

      <template v-if="treeLoaded && !dataFillingStore.getFormTree.length">
        <empty-background description="暂无 Data Filling 表单目录，请先创建文件夹或表单。" img-type="none">
          <div class="df-manage__empty-actions">
            <el-button @click="openCreateFolder">创建第一个文件夹</el-button>
            <el-button type="primary" @click="navigateToCreateForm">创建第一个表单</el-button>
          </div>
        </empty-background>
      </template>

      <template v-else-if="selectedLeaf">
        <div class="df-manage__summary-card">
          <div class="df-manage__summary-header">
            <div>
              <div class="df-manage__summary-title">{{ selectedLeaf.name }}</div>
              <div class="df-manage__summary-subtitle">已选中叶子表单，可继续前往编辑页。</div>
            </div>
            <div class="df-manage__summary-actions">
              <el-button @click="navigateToCreateForm">同目录新建表单</el-button>
              <el-button type="primary" @click="navigateToEditForm">
                <template #icon>
                  <Icon name="icon_edit_outlined"><icon_edit_outlined class="svg-icon" /></Icon>
                </template>
                编辑表单
              </el-button>
            </div>
          </div>

          <div class="df-manage__summary-grid">
            <div class="df-manage__summary-item">
              <span class="label">表单 ID</span>
              <span class="value">{{ selectedLeaf.id }}</span>
            </div>
            <div class="df-manage__summary-item">
              <span class="label">父目录 ID</span>
              <span class="value">{{ selectedLeaf.pid }}</span>
            </div>
            <div class="df-manage__summary-item">
              <span class="label">节点类型</span>
              <span class="value">{{ selectedLeaf.nodeType }}</span>
            </div>
            <div class="df-manage__summary-item">
              <span class="label">子节点数</span>
              <span class="value">{{ selectedLeaf.children?.length || 0 }}</span>
            </div>
          </div>

          <el-tabs v-model="rightTab" class="df-manage__tabs">
            <el-tab-pane label="数据管理" name="data" lazy>
              <DataPage :form-id="Number(selectedLeaf.id)" />
            </el-tab-pane>
            <el-tab-pane label="任务管理" name="tasks" lazy>
              <TaskPage :form-id="Number(selectedLeaf.id)" />
            </el-tab-pane>
          </el-tabs>
        </div>
      </template>

      <template v-else>
        <empty-background description="请选择左侧一个叶子表单以查看摘要，文件夹仅用于组织与新建。" img-type="select" />
      </template>
    </div>
  </div>
</template>

<style lang="less" scoped>
.df-manage {
  display: flex;
  width: 100%;
  height: 100%;
  position: relative;
  background: #fff;

  .resource-area {
    position: relative;
    height: 100%;
    padding: 0;
    border-right: 1px solid #d7d7d7;
    overflow: visible;

    &.retract {
      display: none;
    }
  }

  .resource-tree {
    width: 100%;
    height: 100%;
    padding-top: 16px;
  }

  &__content {
    flex: 1;
    display: flex;
    flex-direction: column;
    min-width: 0;
    background: #f5f6f7;
    padding: 24px;
    gap: 24px;
  }

  &__toolbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    background: #fff;
    border-radius: 4px;
    padding: 16px 24px;
  }

  &__toolbar-title {
    color: #1f2329;
    font-size: 16px;
    font-weight: 500;
    line-height: 24px;
  }

  &__toolbar-actions {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  &__empty-actions {
    display: flex;
    gap: 12px;
  }

  &__summary-card {
    display: flex;
    flex-direction: column;
    gap: 24px;
    background: #fff;
    border-radius: 4px;
    padding: 24px;
  }

  &__summary-header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 16px;
  }

  &__summary-title {
    color: #1f2329;
    font-size: 18px;
    font-weight: 500;
    line-height: 28px;
  }

  &__summary-subtitle {
    margin-top: 4px;
    color: #646a73;
    font-size: 14px;
    line-height: 22px;
  }

  &__summary-actions {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  &__summary-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
    gap: 16px;
  }

  &__summary-item {
    display: flex;
    flex-direction: column;
    gap: 6px;
    border: 1px solid #e8eaec;
    border-radius: 4px;
    padding: 16px;

    .label {
      color: #646a73;
      font-size: 13px;
      line-height: 20px;
    }

    .value {
      color: #1f2329;
      font-size: 14px;
      font-weight: 500;
      line-height: 22px;
      word-break: break-all;
    }
  }

  &__tabs {
    :deep(.el-tabs__header),
    :deep(.ed-tabs__header) {
      margin-bottom: 16px;
    }

    :deep(.el-tabs__content),
    :deep(.ed-tabs__content) {
      overflow: visible;
    }
  }
}
</style>
