<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useWindowSize } from '@vueuse/core'
import dayjs from 'dayjs'
import { useRouter } from 'vue-router_2'
import type { UserTaskPageRequest } from '@/api/datafilling'
import { getUserTaskPageList, getUserTaskTodoCount } from '@/api/datafilling'
import type { DataFillingUserTaskItem } from '@/views/data-filling/types'

interface UserTaskPageResult {
  records?: DataFillingUserTaskItem[]
  total?: number
  current?: number
  size?: number
}

const emit = defineEmits<{
  (e: 'todo-count-change', value: number): void
  (e: 'loaded', value: { name: string; title: string; shortName: string; count: number }): void
}>()

const router = useRouter()
const { width } = useWindowSize()

const searchKeyword = ref('')
const loading = ref(false)
const countLoading = ref(false)
const currentPage = ref(1)
const pageSize = ref(20)
const total = ref(0)
const todoCount = ref(0)
const tasks = ref<DataFillingUserTaskItem[]>([])

const isMobile = computed(() => width.value < 768)
const pageLayout = computed(() => {
  return isMobile.value ? 'prev, pager, next' : 'total, prev, pager, next, sizes, jumper'
})

const requestPayload = computed<UserTaskPageRequest>(() => ({
  taskName: searchKeyword.value.trim()
}))

const resolveStatusLabel = (item: DataFillingUserTaskItem) => {
  if (item.expired || item.status === 2) {
    return '已过期'
  }
  if (item.status === 1) {
    return '已填报'
  }
  return '待填报'
}

const resolveStatusType = (item: DataFillingUserTaskItem) => {
  if (item.expired || item.status === 2) {
    return 'danger'
  }
  if (item.status === 1) {
    return 'success'
  }
  return 'warning'
}

const resolveProgressPercentage = (item: DataFillingUserTaskItem) => {
  if (!item.totalCount) {
    return 0
  }
  const percent = Math.round((item.finishCount / item.totalCount) * 100)
  return Math.min(100, Math.max(0, percent))
}

const formatDateTime = (value?: number) => {
  if (!value) {
    return '--'
  }
  return dayjs(value).format('YYYY-MM-DD HH:mm')
}

const decorateTasks = (records: DataFillingUserTaskItem[]) => {
  return records.map(item => ({
    ...item,
    statusLabel: resolveStatusLabel(item)
  }))
}

const emitTodoCount = (count: number) => {
  emit('todo-count-change', count)
  emit('loaded', {
    name: 'data-filling',
    title: '数据填报',
    shortName: '填报',
    count
  })
}

const loadTodoCount = async () => {
  countLoading.value = true
  try {
    const response = await getUserTaskTodoCount()
    const nextCount = typeof response === 'number' ? response : Number(response || 0)
    todoCount.value = Number.isFinite(nextCount) ? nextCount : 0
    emitTodoCount(todoCount.value)
  } finally {
    countLoading.value = false
  }
}

const loadTaskList = async (page = currentPage.value, size = pageSize.value) => {
  loading.value = true
  try {
    const response = (await getUserTaskPageList(page, size, requestPayload.value)) as UserTaskPageResult
    const records = Array.isArray(response?.records) ? response.records : []
    tasks.value = decorateTasks(records)
    total.value = typeof response?.total === 'number' ? response.total : records.length
    currentPage.value = typeof response?.current === 'number' ? response.current : page
    pageSize.value = typeof response?.size === 'number' ? response.size : size
  } catch {
    tasks.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

const refreshPage = async (page = currentPage.value, size = pageSize.value) => {
  await Promise.all([loadTodoCount(), loadTaskList(page, size)])
}

const handleSearch = async () => {
  currentPage.value = 1
  await refreshPage(1, pageSize.value)
}

const handlePageChange = async (page: number) => {
  currentPage.value = page
  await loadTaskList(page, pageSize.value)
}

const handlePageSizeChange = async (size: number) => {
  pageSize.value = size
  currentPage.value = 1
  await loadTaskList(1, size)
}

const openTask = (item: DataFillingUserTaskItem) => {
  if (item.expired) {
    return
  }
  router.push({
    path: '/data-filling-fill-task',
    query: {
      subTaskId: String(item.id)
    }
  })
}

defineExpose({
  refresh: refreshPage,
  refreshTodoCount: loadTodoCount
})

onMounted(() => {
  refreshPage()
})
</script>

<template>
  <div class="df-fill-list" :class="{ 'df-fill-list--mobile': isMobile }">
    <section class="df-fill-list__panel">
      <div class="df-fill-list__header">
        <div>
          <div class="df-fill-list__eyebrow">Data Filling</div>
          <h2 class="df-fill-list__title">我的填报任务</h2>
          <p class="df-fill-list__subtitle">查看待填报、已完成与已过期任务，支持按任务名称快速筛选。</p>
        </div>
        <div class="df-fill-list__badge-card">
          <span class="df-fill-list__badge-label">待处理</span>
          <span class="df-fill-list__badge-value">{{ countLoading ? '--' : todoCount }}</span>
        </div>
      </div>

      <div class="df-fill-list__toolbar">
        <el-input
          v-model="searchKeyword"
          clearable
          class="df-fill-list__search"
          placeholder="请输入任务名称"
          @clear="handleSearch"
          @keydown.enter.exact.prevent="handleSearch"
        >
          <template #append>
            <el-button @click="handleSearch">搜索</el-button>
          </template>
        </el-input>
      </div>

      <div class="df-fill-list__content" v-loading="loading">
        <el-empty v-if="!tasks.length" description="暂无可处理的填报任务" :image-size="72" />

        <div v-else class="df-fill-list__grid">
          <article
            v-for="task in tasks"
            :key="task.id"
            class="df-fill-list__card"
            :class="{
              'df-fill-list__card--expired': task.expired,
              'df-fill-list__card--clickable': !task.expired
            }"
            :tabindex="task.expired ? -1 : 0"
            @click="openTask(task)"
            @keydown.enter.prevent="openTask(task)"
          >
            <div class="df-fill-list__card-top">
              <div class="df-fill-list__card-heading">
                <h3 class="df-fill-list__card-title">{{ task.taskName }}</h3>
                <p v-if="task.formName" class="df-fill-list__card-subtitle">{{ task.formName }}</p>
              </div>
              <div class="df-fill-list__card-tags">
                <el-tag :type="resolveStatusType(task)" effect="light">{{ task.statusLabel }}</el-tag>
                <el-tag v-if="task.expired" type="danger" effect="dark">已过期</el-tag>
              </div>
            </div>

            <div class="df-fill-list__meta">
              <div class="df-fill-list__meta-item">
                <span class="df-fill-list__meta-label">时间范围</span>
                <span class="df-fill-list__meta-value">
                  {{ formatDateTime(task.startTime) }} → {{ formatDateTime(task.endTime) }}
                </span>
              </div>
              <div class="df-fill-list__meta-item">
                <span class="df-fill-list__meta-label">填报进度</span>
                <span class="df-fill-list__meta-value">{{ task.finishCount }}/{{ task.totalCount }}</span>
              </div>
            </div>

            <div class="df-fill-list__progress">
              <el-progress
                :percentage="resolveProgressPercentage(task)"
                :stroke-width="8"
                :show-text="false"
              />
              <span class="df-fill-list__progress-text">{{ resolveProgressPercentage(task) }}%</span>
            </div>

            <div class="df-fill-list__footer">
              <span v-if="task.expired" class="df-fill-list__footer-text df-fill-list__footer-text--expired">
                该任务已超出填报时间，当前不可继续处理
              </span>
              <span v-else class="df-fill-list__footer-text">点击卡片进入填报页面</span>
            </div>
          </article>
        </div>
      </div>

      <div v-if="total > pageSize" class="df-fill-list__pagination">
        <el-pagination
          :current-page="currentPage"
          :page-size="pageSize"
          :page-sizes="[20, 40, 60, 100]"
          :layout="pageLayout"
          :small="isMobile"
          :total="total"
          background
          @current-change="handlePageChange"
          @size-change="handlePageSizeChange"
        />
      </div>
    </section>
  </div>
</template>

<style lang="less" scoped>
.df-fill-list {
  --df-fill-bg: var(--ed-fill-color-page, #f5f6f7);
  --df-fill-panel-bg: var(--ed-fill-color-blank, #ffffff);
  --df-fill-text-primary: var(--ed-text-color-primary, #1f2329);
  --df-fill-text-secondary: var(--ed-text-color-secondary, #646a73);
  --df-fill-text-placeholder: var(--ed-text-color-placeholder, #8f959e);
  --df-fill-border: var(--ed-border-color-light, #dcdfe6);
  --df-fill-border-strong: var(--ed-border-color, #d7d7d7);
  --df-fill-primary: var(--ed-color-primary, #3370ff);
  --df-fill-danger: var(--ed-color-danger, #f04438);
  --df-fill-shadow: 0 10px 24px rgba(31, 35, 41, 0.06);
  --df-fill-radius: 12px;
  --df-fill-space-xs: 8px;
  --df-fill-space-sm: 12px;
  --df-fill-space-md: 16px;
  --df-fill-space-lg: 24px;
  --df-fill-space-xl: 32px;

  min-height: 100%;
  padding: var(--df-fill-space-lg);
  background: radial-gradient(circle at top right, rgba(51, 112, 255, 0.08), transparent 22%),
    var(--df-fill-bg);

  &__panel {
    display: flex;
    flex-direction: column;
    gap: var(--df-fill-space-lg);
    padding: var(--df-fill-space-xl);
    background: var(--df-fill-panel-bg);
    border: 1px solid var(--df-fill-border);
    border-radius: calc(var(--df-fill-radius) + 4px);
    box-shadow: var(--df-fill-shadow);
  }

  &__header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: var(--df-fill-space-lg);
  }

  &__eyebrow {
    margin-bottom: var(--df-fill-space-xs);
    color: var(--df-fill-primary);
    font-size: 12px;
    font-weight: 600;
    line-height: 18px;
    letter-spacing: 0.12em;
    text-transform: uppercase;
  }

  &__title {
    margin: 0;
    color: var(--df-fill-text-primary);
    font-size: 28px;
    line-height: 36px;
    font-weight: 600;
  }

  &__subtitle {
    margin: var(--df-fill-space-xs) 0 0;
    color: var(--df-fill-text-secondary);
    font-size: 14px;
    line-height: 22px;
  }

  &__badge-card {
    min-width: 140px;
    display: flex;
    flex-direction: column;
    gap: 4px;
    padding: var(--df-fill-space-md);
    background: linear-gradient(135deg, rgba(51, 112, 255, 0.12), rgba(51, 112, 255, 0.04));
    border: 1px solid rgba(51, 112, 255, 0.14);
    border-radius: var(--df-fill-radius);
  }

  &__badge-label {
    color: var(--df-fill-text-secondary);
    font-size: 13px;
    line-height: 20px;
  }

  &__badge-value {
    color: var(--df-fill-text-primary);
    font-size: 30px;
    line-height: 38px;
    font-weight: 700;
  }

  &__toolbar {
    display: flex;
    justify-content: flex-end;
  }

  &__search {
    width: min(100%, 360px);
  }

  &__content {
    min-height: 280px;
  }

  &__grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
    gap: var(--df-fill-space-md);
  }

  &__card {
    display: flex;
    min-height: 232px;
    flex-direction: column;
    gap: var(--df-fill-space-md);
    padding: 20px;
    background: linear-gradient(180deg, rgba(255, 255, 255, 0.94), rgba(245, 246, 247, 0.9));
    border: 1px solid var(--df-fill-border);
    border-radius: var(--df-fill-radius);
    transition: transform 0.2s ease, box-shadow 0.2s ease, border-color 0.2s ease;
    outline: none;

    &--clickable {
      cursor: pointer;

      &:hover,
      &:focus-visible {
        transform: translateY(-2px);
        border-color: rgba(51, 112, 255, 0.32);
        box-shadow: 0 14px 28px rgba(31, 35, 41, 0.1);
      }
    }

    &--expired {
      opacity: 0.6;
      cursor: not-allowed;
    }
  }

  &__card-top,
  &__card-tags,
  &__footer,
  &__progress {
    display: flex;
    align-items: center;
  }

  &__card-top {
    justify-content: space-between;
    gap: var(--df-fill-space-md);
    align-items: flex-start;
  }

  &__card-heading {
    min-width: 0;
  }

  &__card-title {
    margin: 0;
    color: var(--df-fill-text-primary);
    font-size: 18px;
    line-height: 28px;
    font-weight: 600;
    word-break: break-word;
  }

  &__card-subtitle {
    margin: 4px 0 0;
    color: var(--df-fill-text-secondary);
    font-size: 13px;
    line-height: 20px;
  }

  &__card-tags {
    flex-wrap: wrap;
    justify-content: flex-end;
    gap: var(--df-fill-space-xs);
  }

  &__meta {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: var(--df-fill-space-sm);
  }

  &__meta-item {
    display: flex;
    flex-direction: column;
    gap: 4px;
    padding: var(--df-fill-space-sm);
    background: rgba(255, 255, 255, 0.74);
    border: 1px solid var(--df-fill-border);
    border-radius: 10px;
  }

  &__meta-label,
  &__progress-text,
  &__footer-text {
    color: var(--df-fill-text-secondary);
    font-size: 13px;
    line-height: 20px;
  }

  &__meta-value {
    color: var(--df-fill-text-primary);
    font-size: 14px;
    line-height: 22px;
    font-weight: 500;
    word-break: break-word;
  }

  &__progress {
    gap: var(--df-fill-space-sm);
  }

  &__footer {
    margin-top: auto;
    padding-top: var(--df-fill-space-sm);
    border-top: 1px dashed var(--df-fill-border-strong);
  }

  &__footer-text--expired {
    color: var(--df-fill-danger);
  }

  &__pagination {
    display: flex;
    justify-content: flex-end;
  }

  &--mobile {
    padding: var(--df-fill-space-md);

    .df-fill-list__panel {
      padding: var(--df-fill-space-lg);
    }

    .df-fill-list__header {
      flex-direction: column;
    }

    .df-fill-list__badge-card,
    .df-fill-list__search {
      width: 100%;
    }

    .df-fill-list__grid,
    .df-fill-list__meta {
      grid-template-columns: 1fr;
    }

    .df-fill-list__card-top,
    .df-fill-list__pagination {
      flex-direction: column;
      align-items: stretch;
    }

    .df-fill-list__card-tags {
      justify-content: flex-start;
    }
  }
}

@media screen and (max-width: 768px) {
  .df-fill-list {
    &__title {
      font-size: 24px;
      line-height: 32px;
    }

    &__panel {
      border-radius: var(--df-fill-radius);
    }
  }
}
</style>
