<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import dayjs from 'dayjs'
import { ElMessage } from 'element-plus-secondary'
import type { DataFillingSubTaskUser } from '@/views/data-filling/types'
import { getSubTaskUsersList } from '@/api/datafilling'

type UserTabName = 'finished' | 'unfinished'

const props = withDefaults(
  defineProps<{
    modelValue: boolean
    subTaskId: number
  }>(),
  {
    modelValue: false,
    subTaskId: 0
  }
)

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
}>()

const STATUS_MAP: Record<number, string> = {
  0: '未完成',
  1: '已完成'
}

const visible = computed({
  get: () => props.modelValue,
  set: value => emit('update:modelValue', value)
})

const loading = ref(false)
const activeTab = ref<UserTabName>('finished')
const rows = ref<DataFillingSubTaskUser[]>([])

const formatFinishTime = (value: unknown) => {
  return typeof value === 'number' && value > 0 ? dayjs(value).format('YYYY-MM-DD HH:mm:ss') : '--'
}

const loadUsers = async () => {
  if (!visible.value || !props.subTaskId) {
    return
  }

  loading.value = true
  try {
    const response = await getSubTaskUsersList(props.subTaskId, activeTab.value)
    rows.value = Array.isArray(response) ? (response as DataFillingSubTaskUser[]) : []
  } catch {
    rows.value = []
    ElMessage.error('加载子任务用户失败')
  } finally {
    loading.value = false
  }
}

watch(
  () => visible.value,
  open => {
    if (!open) {
      rows.value = []
      activeTab.value = 'finished'
      return
    }

    activeTab.value = 'finished'
    void loadUsers()
  },
  { immediate: false }
)

watch(
  () => activeTab.value,
  () => {
    if (visible.value) {
      void loadUsers()
    }
  }
)
</script>

<template>
  <el-drawer v-model="visible" title="子任务用户" size="720px" direction="rtl" append-to-body>
    <div class="df-sub-task-users">
      <div class="df-sub-task-users__header">
        <div class="df-sub-task-users__title">子任务 #{{ subTaskId }}</div>
        <div class="df-sub-task-users__subtitle">在已完成与未完成间切换，快速查看用户分配结果。</div>
      </div>

      <el-tabs v-model="activeTab" class="df-sub-task-users__tabs">
        <el-tab-pane label="已完成" name="finished" />
        <el-tab-pane label="未完成" name="unfinished" />
      </el-tabs>

      <el-table :data="rows" border stripe v-loading="loading" class="df-sub-task-users__table">
        <el-table-column prop="uid" label="用户ID(uid)" min-width="120" />
        <el-table-column label="状态" min-width="110">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'" effect="plain">
              {{ STATUS_MAP[row.status] || '--' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="完成时间" min-width="180">
          <template #default="{ row }">
            {{ formatFinishTime(row.finishTime) }}
          </template>
        </el-table-column>
        <el-table-column prop="dataId" label="数据ID" min-width="180" show-overflow-tooltip />

        <template #empty>
          <el-empty description="当前筛选条件下暂无用户记录" :image-size="72" />
        </template>
      </el-table>
    </div>
  </el-drawer>
</template>

<style lang="less" scoped>
.df-sub-task-users {
  --df-space-1: 4px;
  --df-space-2: 8px;
  --df-space-3: 12px;
  --df-space-4: 16px;
  --df-space-5: 20px;
  --df-radius-sm: 4px;
  --df-text-primary: var(--deTextPrimary, #1f2329);
  --df-text-secondary: var(--N600, #646a73);
  --df-border-color: var(--deCardStrokeColor, #e8eaec);
  --df-bg-soft: var(--deTextPrimary5, #f5f6f7);

  display: flex;
  height: 100%;
  flex-direction: column;
  gap: var(--df-space-4);

  &__header {
    border: 1px solid var(--df-border-color);
    border-radius: var(--df-radius-sm);
    background: var(--df-bg-soft);
    padding: var(--df-space-4) var(--df-space-5);
  }

  &__title {
    color: var(--df-text-primary);
    font-size: 16px;
    font-weight: 500;
    line-height: 24px;
  }

  &__subtitle {
    margin-top: var(--df-space-1);
    color: var(--df-text-secondary);
    font-size: 13px;
    line-height: 20px;
  }

  &__tabs {
    :deep(.el-tabs__header),
    :deep(.ed-tabs__header) {
      margin-bottom: 0;
    }
  }

  &__table {
    flex: 1;
  }
}
</style>
