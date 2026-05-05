<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import dayjs from 'dayjs'
import { ElMessage, ElMessageBox } from 'element-plus-secondary'
import { clearCommitLog, getCommitLogPage } from '@/api/datafilling'

const props = withDefaults(
  defineProps<{
    modelValue: boolean
    formId: number
  }>(),
  {
    modelValue: false,
    formId: 0
  }
)

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
}>()

const loading = ref(false)
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)
const clearType = ref('all')
const logRows = ref<Array<Record<string, unknown>>>([])

const visible = computed({
  get: () => props.modelValue,
  set: value => emit('update:modelValue', value)
})

const formatCommitTime = (value: unknown) => {
  if (typeof value !== 'number') {
    return '--'
  }
  return dayjs(value).format('YYYY-MM-DD HH:mm:ss')
}

const formatOperate = (value: unknown) => {
  if (value === 0) {
    return '删除'
  }
  if (value === 1) {
    return '新增'
  }
  if (value === 2) {
    return '更新'
  }
  return '--'
}

const loadLogs = async () => {
  if (!props.formId || !visible.value) {
    return
  }

  loading.value = true
  try {
    const response = await getCommitLogPage(currentPage.value, pageSize.value, {
      formId: props.formId
    })
    const records = Array.isArray((response as { records?: unknown[] })?.records)
      ? ((response as { records?: unknown[] }).records ?? [])
      : Array.isArray(response)
      ? response
      : []
    logRows.value = records as Array<Record<string, unknown>>
    total.value =
      typeof (response as { total?: number })?.total === 'number'
        ? (response as { total?: number }).total ?? 0
        : records.length
  } catch {
    ElMessage.error('加载提交日志失败')
  } finally {
    loading.value = false
  }
}

const handleClear = async () => {
  try {
    await ElMessageBox.confirm('确认清空当前表单的提交日志吗？', '提示', {
      type: 'warning',
      confirmButtonType: 'danger',
      confirmButtonText: '确认清空',
      cancelButtonText: '取消',
      autofocus: false,
      showClose: false
    })
    await clearCommitLog({
      formId: props.formId,
      clearType: clearType.value
    })
    ElMessage.success('提交日志已清空')
    currentPage.value = 1
    await loadLogs()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('清空提交日志失败')
    }
  }
}

watch(
  () => visible.value,
  value => {
    if (value) {
      currentPage.value = 1
      void loadLogs()
    }
  }
)
</script>

<template>
  <el-drawer v-model="visible" title="提交日志" size="720px" direction="rtl" append-to-body>
    <div class="df-commit-log">
      <div class="df-commit-log__toolbar">
        <el-select v-model="clearType" placeholder="清空范围" style="width: 180px">
          <el-option label="全部日志" value="all" />
          <el-option label="最近日志" value="latest" />
        </el-select>
        <el-button type="danger" @click="handleClear">清空日志</el-button>
      </div>

      <el-table :data="logRows" border stripe v-loading="loading" class="df-commit-log__table">
        <el-table-column prop="committer" label="操作人" min-width="140" />
        <el-table-column label="操作类型" min-width="100">
          <template #default="{ row }">
            {{ formatOperate(row.operate) }}
          </template>
        </el-table-column>
        <el-table-column label="提交时间" min-width="180">
          <template #default="{ row }">
            {{ formatCommitTime(row.commitTime) }}
          </template>
        </el-table-column>
        <el-table-column prop="count" label="数量" min-width="90" />

        <template #empty>
          <el-empty description="暂无提交日志" :image-size="72" />
        </template>
      </el-table>

      <div class="df-commit-log__pagination">
        <el-pagination
          :current-page="currentPage"
          :page-size="pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          background
          layout="total, prev, pager, next, sizes, jumper"
          @current-change="value => { currentPage = value; loadLogs() }"
          @size-change="value => { pageSize = value; currentPage = 1; loadLogs() }"
        />
      </div>
    </div>
  </el-drawer>
</template>

<style lang="less" scoped>
.df-commit-log {
  display: flex;
  height: 100%;
  flex-direction: column;
  gap: 16px;

  &__toolbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
  }

  &__table {
    flex: 1;
  }

  &__pagination {
    display: flex;
    justify-content: flex-end;
  }
}
</style>
