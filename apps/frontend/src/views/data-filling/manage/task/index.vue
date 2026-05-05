<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import dayjs from 'dayjs'
import { ElMessage, ElMessageBox } from 'element-plus-secondary'
import { useRoute } from 'vue-router_2'
import type { TaskPageResponse } from '@/api/datafilling'
import { deleteTasks, executeNowTask, getTaskPageList, startTask, stopTask } from '@/api/datafilling'
import type { DataFillingTaskItem } from '@/views/data-filling/types'
import SubTaskList from './SubTaskList.vue'
import TaskEditor from './TaskEditor.vue'

const props = withDefaults(
  defineProps<{
    formId?: number
  }>(),
  {
    formId: 0
  }
)

const route = useRoute()

const TASK_STATUS_MAP: Record<number, string> = {
  0: '未开始',
  1: '执行中',
  2: '已停止',
  3: '已完成'
}

const RATE_TYPE_MAP: Record<number, string> = {
  0: '单次',
  1: '每日',
  2: '每周',
  3: '每月',
  4: '每年'
}

const LAST_EXEC_STATUS_MAP: Record<number, string> = {
  0: '未执行',
  1: '成功',
  2: '失败',
  3: '执行中'
}

const currentFormId = ref<number | null>(null)
const loading = ref(false)
const taskRows = ref<DataFillingTaskItem[]>([])
const selectedRows = ref<DataFillingTaskItem[]>([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(10)

const editorVisible = ref(false)
const editingTaskId = ref<number>()
const subTaskVisible = ref(false)
const activeTaskId = ref(0)
const activeTaskName = ref('')

const pageReady = computed(() => Boolean(currentFormId.value))
const hasSelection = computed(() => selectedRows.value.length > 0)
const runningCount = computed(() => taskRows.value.filter(item => item.status === 1).length)

const pageTitle = computed(() => 'Data Filling 任务管理')

const pageSubtitle = computed(() => {
  if (!currentFormId.value) {
    return '请先从路由参数或父级上下文传入 formId。'
  }

  return `表单 ID：${currentFormId.value} · 统一管理任务调度、执行状态与子任务进度`
})

const parseRouteFormId = (value: unknown) => {
  const rawValue = Array.isArray(value) ? value[0] : value
  if (typeof rawValue !== 'string' || !rawValue.trim()) {
    return null
  }

  const parsedValue = Number(rawValue)
  return Number.isFinite(parsedValue) && parsedValue > 0 ? parsedValue : null
}

const isCancelableAction = (error: unknown) => {
  return error === 'cancel' || error === 'close'
}

const formatTime = (value: unknown) => {
  return typeof value === 'number' && value > 0 ? dayjs(value).format('YYYY-MM-DD HH:mm:ss') : '--'
}

const formatStatus = (value: unknown, map: Record<number, string>) => {
  return typeof value === 'number' ? map[value] || '--' : '--'
}

const normalizeTaskRows = (records: DataFillingTaskItem[]) => {
  return records.map(item => ({
    ...item,
    statusLabel: TASK_STATUS_MAP[item.status] || '--',
    lastExecStatusLabel: LAST_EXEC_STATUS_MAP[item.lastExecStatus] || '--'
  }))
}

const resetTaskState = () => {
  taskRows.value = []
  selectedRows.value = []
  total.value = 0
  currentPage.value = 1
  pageSize.value = 10
}

const loadTasks = async (page = currentPage.value, size = pageSize.value) => {
  if (!currentFormId.value) {
    return
  }

  loading.value = true
  try {
    const response = (await getTaskPageList(currentFormId.value, page, size)) as TaskPageResponse
    const records = Array.isArray(response?.records) ? response.records : []

    taskRows.value = normalizeTaskRows(records)
    total.value = typeof response?.total === 'number' ? response.total : 0
    currentPage.value = typeof response?.current === 'number' ? response.current : page
    pageSize.value = typeof response?.size === 'number' ? response.size : size
    selectedRows.value = []
  } catch {
    taskRows.value = []
    total.value = 0
    ElMessage.error('加载任务列表失败')
  } finally {
    loading.value = false
  }
}

const initPage = async (formId: number | null) => {
  currentFormId.value = formId
  resetTaskState()

  if (!formId) {
    return
  }

  currentFormId.value = formId
  await loadTasks(1, pageSize.value)
}

const reloadCurrentPage = async () => {
  await loadTasks(currentPage.value, pageSize.value)
}

const handleSelectionChange = (rows: DataFillingTaskItem[]) => {
  selectedRows.value = rows
}

const handleCurrentChange = async (page: number) => {
  currentPage.value = page
  await loadTasks(page, pageSize.value)
}

const handleSizeChange = async (size: number) => {
  pageSize.value = size
  currentPage.value = 1
  await loadTasks(1, size)
}

const openCreateEditor = () => {
  editingTaskId.value = undefined
  editorVisible.value = true
}

const openEditEditor = (taskId: number) => {
  editingTaskId.value = taskId
  editorVisible.value = true
}

const openSubTasks = (row: DataFillingTaskItem) => {
  activeTaskId.value = row.id
  activeTaskName.value = row.name
  subTaskVisible.value = true
}

const runConfirmedAction = async (options: {
  title: string
  message: string
  confirmButtonText: string
  successMessage: string
  errorMessage: string
  action: () => Promise<unknown>
  reloadPage?: boolean
}) => {
  try {
    await ElMessageBox.confirm(options.message, options.title, {
      type: 'warning',
      confirmButtonText: options.confirmButtonText,
      cancelButtonText: '取消',
      autofocus: false,
      showClose: false
    })

    await options.action()
    ElMessage.success(options.successMessage)

    if (options.reloadPage !== false) {
      await reloadCurrentPage()
    }
  } catch (error) {
    if (!isCancelableAction(error)) {
      ElMessage.error(options.errorMessage)
    }
  }
}

const handleStart = async (row: DataFillingTaskItem) => {
  if (!currentFormId.value) {
    return
  }

  await runConfirmedAction({
    title: '启动任务',
    message: `确认启动任务“${row.name}”吗？`,
    confirmButtonText: '确认启动',
    successMessage: '任务已启动',
    errorMessage: '启动任务失败',
    action: () => startTask(currentFormId.value as number, row.id)
  })
}

const handleStop = async (row: DataFillingTaskItem) => {
  if (!currentFormId.value) {
    return
  }

  await runConfirmedAction({
    title: '停止任务',
    message: `确认停止任务“${row.name}”吗？`,
    confirmButtonText: '确认停止',
    successMessage: '任务已停止',
    errorMessage: '停止任务失败',
    action: () => stopTask(currentFormId.value as number, row.id)
  })
}

const handleExecuteNow = async (row: DataFillingTaskItem) => {
  await runConfirmedAction({
    title: '立即执行',
    message: `确认立即执行任务“${row.name}”吗？`,
    confirmButtonText: '立即执行',
    successMessage: '任务已触发执行',
    errorMessage: '立即执行失败',
    action: () => executeNowTask({ taskId: row.id })
  })
}

const handleBatchDelete = async () => {
  if (!currentFormId.value || !selectedRows.value.length) {
    return
  }

  const ids = selectedRows.value.map(item => item.id)
  const nextPage = ids.length >= taskRows.value.length && currentPage.value > 1 ? currentPage.value - 1 : currentPage.value

  try {
    await ElMessageBox.confirm(`确认删除已选中的 ${ids.length} 个任务吗？`, '批量删除任务', {
      type: 'warning',
      confirmButtonType: 'danger',
      confirmButtonText: '确认删除',
      cancelButtonText: '取消',
      autofocus: false,
      showClose: false
    })

    await deleteTasks(currentFormId.value, { ids })
    ElMessage.success('任务已删除')
    currentPage.value = nextPage
    await loadTasks(nextPage, pageSize.value)
  } catch (error) {
    if (!isCancelableAction(error)) {
      ElMessage.error('删除任务失败')
    }
  }
}

const handleEditorSaved = async () => {
  currentPage.value = 1
  await loadTasks(1, pageSize.value)
}

watch(
  [() => props.formId, () => route.query.formId],
  ([propFormId, routeFormId]) => {
    const targetFormId = propFormId > 0 ? propFormId : parseRouteFormId(routeFormId)
    void initPage(targetFormId)
  },
  { immediate: true }
)
</script>

<template>
  <div class="df-task-page" v-loading="loading">
    <div class="df-task-page__header">
      <div>
        <div class="df-task-page__title">{{ pageTitle }}</div>
        <div class="df-task-page__subtitle">{{ pageSubtitle }}</div>
      </div>

      <div class="df-task-page__header-meta">
        <div class="df-task-page__meta-item">
          <span class="label">当前页任务</span>
          <span class="value">{{ taskRows.length }}</span>
        </div>
        <div class="df-task-page__meta-item">
          <span class="label">运行中</span>
          <span class="value">{{ runningCount }}</span>
        </div>
        <div class="df-task-page__meta-item">
          <span class="label">已选择</span>
          <span class="value">{{ selectedRows.length }}</span>
        </div>
      </div>
    </div>

    <template v-if="pageReady">
      <el-card shadow="never" class="df-task-page__card df-task-page__toolbar-card">
        <div class="df-task-page__toolbar">
          <div class="df-task-page__toolbar-main">
            <el-button type="primary" @click="openCreateEditor">创建任务</el-button>
            <el-button type="danger" plain :disabled="!hasSelection" @click="handleBatchDelete">批量删除</el-button>
          </div>

          <div class="df-task-page__toolbar-tip">子任务与执行明细可在“查看子任务”中继续下钻。</div>
        </div>
      </el-card>

      <el-card shadow="never" class="df-task-page__card df-task-page__grid-card">
        <el-table
          :data="taskRows"
          border
          stripe
          class="df-task-page__table"
          @selection-change="handleSelectionChange"
        >
          <el-table-column type="selection" width="48" />
          <el-table-column prop="name" label="任务名称" min-width="220" show-overflow-tooltip />
          <el-table-column label="任务状态" min-width="110">
            <template #default="{ row }">
              <el-tag :type="row.status === 1 ? 'success' : row.status === 2 ? 'warning' : 'info'" effect="plain">
                {{ row.statusLabel || formatStatus(row.status, TASK_STATUS_MAP) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="调度类型" min-width="110">
            <template #default="{ row }">
              {{ formatStatus(row.rateType, RATE_TYPE_MAP) }}
            </template>
          </el-table-column>
          <el-table-column label="最近执行状态" min-width="120">
            <template #default="{ row }">
              <span
                :class="[
                  'df-task-page__status-text',
                  {
                    'is-success': row.lastExecStatus === 1,
                    'is-danger': row.lastExecStatus === 2,
                    'is-primary': row.lastExecStatus === 3
                  }
                ]"
              >
                {{ row.lastExecStatusLabel || formatStatus(row.lastExecStatus, LAST_EXEC_STATUS_MAP) }}
              </span>
            </template>
          </el-table-column>
          <el-table-column label="下次执行时间" min-width="180">
            <template #default="{ row }">
              {{ formatTime(row.nextExecTime) }}
            </template>
          </el-table-column>
          <el-table-column label="操作" fixed="right" width="300" align="center">
            <template #default="{ row }">
              <div class="df-task-page__row-actions">
                <el-button link type="primary" @click="openEditEditor(row.id)">编辑</el-button>
                <el-button link type="success" :disabled="row.status === 1" @click="handleStart(row)">启动</el-button>
                <el-button link type="warning" :disabled="row.status !== 1" @click="handleStop(row)">停止</el-button>
                <el-button link type="primary" @click="handleExecuteNow(row)">立即执行</el-button>
                <el-button link type="info" @click="openSubTasks(row)">查看子任务</el-button>
              </div>
            </template>
          </el-table-column>

          <template #empty>
            <el-empty description="当前表单暂无任务，请先创建任务。" :image-size="80">
              <el-button type="primary" @click="openCreateEditor">创建任务</el-button>
            </el-empty>
          </template>
        </el-table>

        <div class="df-task-page__pagination">
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
      </el-card>
    </template>

    <el-card v-else shadow="never" class="df-task-page__card df-task-page__empty-card">
      <el-empty description="未获取到有效 formId，暂时无法加载任务管理页。" :image-size="80" />
    </el-card>

    <TaskEditor
      v-model="editorVisible"
      :form-id="currentFormId || 0"
      :task-id="editingTaskId"
      @saved="handleEditorSaved"
    />

    <SubTaskList
      v-model="subTaskVisible"
      :task-id="activeTaskId"
      :form-id="currentFormId || 0"
    />
  </div>
</template>

<style lang="less" scoped>
.df-task-page {
  --df-space-1: 4px;
  --df-space-2: 8px;
  --df-space-3: 12px;
  --df-space-4: 16px;
  --df-space-5: 20px;
  --df-space-6: 24px;
  --df-radius-sm: 4px;
  --df-text-primary: var(--deTextPrimary, #1f2329);
  --df-text-secondary: var(--N600, #646a73);
  --df-border-color: var(--deCardStrokeColor, #e8eaec);
  --df-bg-page: var(--deTextPrimary5, #f5f6f7);
  --df-bg-surface: #fff;
  --df-bg-soft: var(--deTextPrimary5, #f5f6f7);
  --df-color-primary: var(--ed-color-primary, #3370ff);
  --df-color-primary-soft: var(--ed-color-primary-1a, rgba(51, 112, 255, 0.1));
  --df-color-success: var(--ed-color-success, #34c724);
  --df-color-danger: var(--ed-color-danger, #f54a45);

  display: flex;
  min-height: 100%;
  flex-direction: column;
  gap: var(--df-space-4);
  background: var(--df-bg-page);
  padding: var(--df-space-6);

  &__header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: var(--df-space-4);
    border-radius: var(--df-radius-sm);
    background: var(--df-bg-surface);
    padding: var(--df-space-5) var(--df-space-6);
  }

  &__title {
    color: var(--df-text-primary);
    font-size: 18px;
    font-weight: 500;
    line-height: 28px;
  }

  &__subtitle {
    margin-top: var(--df-space-1);
    color: var(--df-text-secondary);
    font-size: 14px;
    line-height: 22px;
  }

  &__header-meta {
    display: grid;
    min-width: 360px;
    grid-template-columns: repeat(3, minmax(100px, 1fr));
    gap: var(--df-space-3);
  }

  &__meta-item {
    display: flex;
    flex-direction: column;
    gap: var(--df-space-1);
    border: 1px solid var(--df-border-color);
    border-radius: var(--df-radius-sm);
    background: var(--df-bg-soft);
    padding: var(--df-space-3) var(--df-space-4);

    .label {
      color: var(--df-text-secondary);
      font-size: 12px;
      line-height: 18px;
    }

    .value {
      color: var(--df-color-primary);
      font-size: 18px;
      font-weight: 500;
      line-height: 28px;
    }
  }

  &__card {
    border: none;
    border-radius: var(--df-radius-sm);

    :deep(.el-card__body),
    :deep(.ed-card__body) {
      padding: var(--df-space-5) var(--df-space-6);
    }
  }

  &__toolbar-card {
    :deep(.el-card__body),
    :deep(.ed-card__body) {
      padding-top: var(--df-space-4);
      padding-bottom: var(--df-space-4);
    }
  }

  &__toolbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--df-space-4);
  }

  &__toolbar-main {
    display: flex;
    align-items: center;
    gap: var(--df-space-3);
    flex-wrap: wrap;
  }

  &__toolbar-tip {
    color: var(--df-text-secondary);
    font-size: 13px;
    line-height: 20px;
    text-align: right;
  }

  &__grid-card {
    flex: 1;
    min-height: 0;
  }

  &__table {
    width: 100%;
  }

  &__status-text {
    color: var(--df-text-secondary);

    &.is-success {
      color: var(--df-color-success);
    }

    &.is-danger {
      color: var(--df-color-danger);
    }

    &.is-primary {
      color: var(--df-color-primary);
    }
  }

  &__row-actions {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: var(--df-space-1);
    flex-wrap: wrap;
  }

  &__pagination {
    display: flex;
    justify-content: flex-end;
    margin-top: var(--df-space-4);
  }

  &__empty-card {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
  }
}
</style>
