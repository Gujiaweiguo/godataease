<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import dayjs from 'dayjs'
import { ElMessage, ElMessageBox } from 'element-plus-secondary'
import type { DataFillingSubTaskItem } from '@/views/data-filling/types'
import type { SubTaskPageResponse } from '@/api/datafilling'
import { deleteSubTasks, getSubTaskPageList } from '@/api/datafilling'
import SubTaskUsers from './SubTaskUsers.vue'

const props = withDefaults(
  defineProps<{
    modelValue: boolean
    taskId: number
    formId: number
  }>(),
  {
    modelValue: false,
    taskId: 0,
    formId: 0
  }
)

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
}>()

const EXEC_STATUS_MAP: Record<number, string> = {
  0: '未执行',
  1: '执行中',
  2: '已完成',
  3: '失败'
}

const visible = computed({
  get: () => props.modelValue,
  set: value => emit('update:modelValue', value)
})

const loading = ref(false)
const rows = ref<DataFillingSubTaskItem[]>([])
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)

const usersVisible = ref(false)
const activeSubTaskId = ref(0)

const formatTime = (value: unknown) => {
  return typeof value === 'number' && value > 0 ? dayjs(value).format('YYYY-MM-DD HH:mm:ss') : '--'
}

const isCancelableAction = (error: unknown) => {
  return error === 'cancel' || error === 'close'
}

const normalizeRows = (records: DataFillingSubTaskItem[]) => {
  return records.map(item => ({
    ...item,
    execStatusLabel: EXEC_STATUS_MAP[item.execStatus] || '--'
  }))
}

const loadSubTasks = async (page = currentPage.value, size = pageSize.value) => {
  if (!visible.value || !props.taskId) {
    return
  }

  loading.value = true
  try {
    const response = (await getSubTaskPageList(page, size, { taskId: props.taskId })) as SubTaskPageResponse
    const records = Array.isArray(response?.records) ? response.records : []

    rows.value = normalizeRows(records)
    total.value = typeof response?.total === 'number' ? response.total : 0
    currentPage.value = typeof response?.current === 'number' ? response.current : page
    pageSize.value = typeof response?.size === 'number' ? response.size : size
  } catch {
    rows.value = []
    total.value = 0
    ElMessage.error('加载子任务列表失败')
  } finally {
    loading.value = false
  }
}

const openUsers = (subTaskId: number) => {
  activeSubTaskId.value = subTaskId
  usersVisible.value = true
}

const handleDelete = async (row: DataFillingSubTaskItem) => {
  if (!props.formId) {
    return
  }

  const nextPage = rows.value.length === 1 && currentPage.value > 1 ? currentPage.value - 1 : currentPage.value

  try {
    await ElMessageBox.confirm(`确认删除子任务 #${row.id} 吗？`, '删除子任务', {
      type: 'warning',
      confirmButtonType: 'danger',
      confirmButtonText: '确认删除',
      cancelButtonText: '取消',
      autofocus: false,
      showClose: false
    })

    await deleteSubTasks(props.formId, {
      ids: [row.id]
    })
    ElMessage.success('子任务已删除')
    currentPage.value = nextPage
    await loadSubTasks(nextPage, pageSize.value)
  } catch (error) {
    if (!isCancelableAction(error)) {
      ElMessage.error('删除子任务失败')
    }
  }
}

const handleCurrentChange = async (page: number) => {
  currentPage.value = page
  await loadSubTasks(page, pageSize.value)
}

const handleSizeChange = async (size: number) => {
  pageSize.value = size
  currentPage.value = 1
  await loadSubTasks(1, size)
}

watch(
  [() => visible.value, () => props.taskId],
  ([open]) => {
    if (!open) {
      rows.value = []
      total.value = 0
      currentPage.value = 1
      pageSize.value = 10
      return
    }

    currentPage.value = 1
    void loadSubTasks(1, pageSize.value)
  },
  { immediate: false }
)
</script>

<template>
  <el-drawer v-model="visible" title="子任务列表" size="800px" direction="rtl" append-to-body>
    <div class="df-sub-task-list">
      <div class="df-sub-task-list__header">
        <div>
          <div class="df-sub-task-list__title">任务 #{{ taskId }}</div>
          <div class="df-sub-task-list__subtitle">查看分发执行结果、进度统计与用户明细。</div>
        </div>

        <div class="df-sub-task-list__header-meta">
          <div class="df-sub-task-list__meta-item">
            <span class="label">当前页</span>
            <span class="value">{{ rows.length }}</span>
          </div>
          <div class="df-sub-task-list__meta-item">
            <span class="label">总数</span>
            <span class="value">{{ total }}</span>
          </div>
        </div>
      </div>

      <el-table :data="rows" border stripe v-loading="loading" class="df-sub-task-list__table">
        <el-table-column label="开始时间" min-width="172">
          <template #default="{ row }">
            {{ formatTime(row.startTime) }}
          </template>
        </el-table-column>
        <el-table-column label="结束时间" min-width="172">
          <template #default="{ row }">
            {{ formatTime(row.endTime) }}
          </template>
        </el-table-column>
        <el-table-column label="执行状态" min-width="110">
          <template #default="{ row }">
            <el-tag :type="row.execStatus === 2 ? 'success' : row.execStatus === 3 ? 'danger' : 'info'" effect="plain">
              {{ row.execStatusLabel || '--' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="totalCount" label="总数" min-width="90" />
        <el-table-column prop="unfinishedCount" label="未完成数" min-width="100" />
        <el-table-column label="用户进度" min-width="140">
          <template #default="{ row }">
            {{ row.totalUserCount - row.unfinishedUserCount }}/{{ row.totalUserCount }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="170" align="center" fixed="right">
          <template #default="{ row }">
            <div class="df-sub-task-list__row-actions">
              <el-button link type="danger" @click="handleDelete(row)">删除</el-button>
              <el-button link type="primary" @click="openUsers(row.id)">查看用户</el-button>
            </div>
          </template>
        </el-table-column>

        <template #empty>
          <el-empty description="当前任务暂无子任务记录" :image-size="72" />
        </template>
      </el-table>

      <div class="df-sub-task-list__pagination">
        <el-pagination
          :current-page="currentPage"
          :page-size="pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          background
          layout="total, prev, pager, next, sizes, jumper"
          @current-change="handleCurrentChange"
          @size-change="handleSizeChange"
        />
      </div>
    </div>

    <SubTaskUsers v-model="usersVisible" :sub-task-id="activeSubTaskId" />
  </el-drawer>
</template>

<style lang="less" scoped>
.df-sub-task-list {
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
  --df-color-primary: var(--ed-color-primary, #3370ff);

  display: flex;
  height: 100%;
  flex-direction: column;
  gap: var(--df-space-4);

  &__header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: var(--df-space-4);
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

  &__header-meta {
    display: grid;
    grid-template-columns: repeat(2, minmax(84px, 1fr));
    gap: var(--df-space-3);
  }

  &__meta-item {
    display: flex;
    min-width: 84px;
    flex-direction: column;
    gap: var(--df-space-1);
    border: 1px solid var(--df-border-color);
    border-radius: var(--df-radius-sm);
    background: #fff;
    padding: var(--df-space-3);

    .label {
      color: var(--df-text-secondary);
      font-size: 12px;
      line-height: 18px;
    }

    .value {
      color: var(--df-color-primary);
      font-size: 16px;
      font-weight: 500;
      line-height: 24px;
    }
  }

  &__table {
    flex: 1;
  }

  &__row-actions {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: var(--df-space-1);
  }

  &__pagination {
    display: flex;
    justify-content: flex-end;
  }
}
</style>
