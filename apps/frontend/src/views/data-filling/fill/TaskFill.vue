<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useWindowSize } from '@vueuse/core'
import dayjs from 'dayjs'
import { ElMessage, ElMessageBox } from 'element-plus-secondary'
import { useRoute, useRouter } from 'vue-router_2'
import type { SearchParam, TableDataResponse, UserTaskData } from '@/api/datafilling'
import {
  deleteUserTaskData,
  getUserTaskData,
  saveUserTaskData,
  searchTableData,
  uploadExcelFile,
  userTaskConfirmUpload
} from '@/api/datafilling'
import FormSchemaRenderer from '@/views/data-filling/components/FormSchemaRenderer.vue'
import type { DataFillingFormSchema, FormFieldValue } from '@/views/data-filling/types'
import { isRecord, parseFormSchema as parseSchema, resolveFieldKey } from '@/views/data-filling/utils/schemaParser'

interface FillRowItem {
  localId: string
  dataId?: string
  taskInstanceId?: number
  model: Record<string, FormFieldValue>
}

const route = useRoute()
const router = useRouter()
const { width } = useWindowSize()

const pageLoading = ref(false)
const submitLoading = ref(false)
const uploadLoading = ref(false)
const uploadConfirmLoading = ref(false)

const formTitle = ref('')
const formId = ref(0)
const fillType = ref(0)
const formExtSetting = ref('')
const taskData = ref<UserTaskData | null>(null)
const formSchema = ref<DataFillingFormSchema>([])
const rows = ref<FillRowItem[]>([])
const uploadPreviewVisible = ref(false)
const uploadId = ref('')
const uploadFileName = ref('')
const previewColumns = ref<Array<{ key: string; label: string }>>([])
const previewRows = ref<Array<Record<string, unknown>>>([])

const isMobile = computed(() => width.value < 768)
const rendererForms = computed(() => JSON.stringify(formSchema.value))
const subTaskId = computed(() => {
  const rawValue = Array.isArray(route.query.subTaskId) ? route.query.subTaskId[0] : route.query.subTaskId
  const parsedValue = Number(rawValue)
  return Number.isFinite(parsedValue) && parsedValue > 0 ? parsedValue : 0
})

const singleFillCompleted = computed(() => {
  if (fillType.value !== 0 || !taskData.value) {
    return false
  }
  return taskData.value.subInstances.some(item => Boolean(item.finishTime) || item.status === 1)
})

const canEdit = computed(() => {
  return Boolean(subTaskId.value && formId.value && !singleFillCompleted.value)
})

const pageDescription = computed(() => {
  if (singleFillCompleted.value) {
    return '当前任务为单次填报，您已提交完成，本页仅支持查看。'
  }
  if (fillType.value === 0) {
    return '单次填报任务提交后即视为完成，请确认内容后再提交。'
  }
  return '支持逐行新增、编辑、删除与 Excel 追加，提交后会同步更新当前任务。'
})

const createEmptyModel = () => {
  return formSchema.value.reduce<Record<string, FormFieldValue>>((result, field, index) => {
    const fieldKey = resolveFieldKey(field, index)
    if (field.defaultValue !== undefined) {
      result[fieldKey] = field.defaultValue
    }
    return result
  }, {})
}

const createEmptyRow = (taskInstanceId?: number) => {
  return {
    localId: `row-${Date.now()}-${Math.random().toString(16).slice(2, 8)}`,
    taskInstanceId,
    model: createEmptyModel()
  }
}

const buildRowModel = (row: Record<string, unknown>) => {
  const nextModel = createEmptyModel()
  formSchema.value.forEach((field, index) => {
    const fieldKey = resolveFieldKey(field, index)
    if (Object.prototype.hasOwnProperty.call(row, fieldKey)) {
      nextModel[fieldKey] = row[fieldKey] as FormFieldValue
    }
  })

  if (row.id !== undefined && row.id !== null) {
    nextModel.id = row.id as FormFieldValue
  }
  return nextModel
}

const isEmptyFieldValue = (value: FormFieldValue | undefined) => {
  if (Array.isArray(value)) {
    return value.length === 0
  }
  return value === undefined || value === null || value === ''
}

const sanitizeRowPayload = (row: FillRowItem) => {
  const payload = formSchema.value.reduce<Record<string, FormFieldValue>>((result, field, index) => {
    const fieldKey = resolveFieldKey(field, index)
    if (Object.prototype.hasOwnProperty.call(row.model, fieldKey)) {
      result[fieldKey] = row.model[fieldKey]
    }
    return result
  }, {})

  if (row.model.id !== undefined && row.model.id !== null && row.model.id !== '') {
    payload.id = row.model.id
  }
  return payload
}

const isBlankPayload = (payload: Record<string, FormFieldValue>) => {
  const keys = Object.keys(payload).filter(key => key !== 'id')
  if (!keys.length) {
    return true
  }
  return keys.every(key => isEmptyFieldValue(payload[key]))
}

const validateRows = (payloadRows: Record<string, FormFieldValue>[]) => {
  if (!payloadRows.length) {
    ElMessage.warning('请至少填写一行有效数据')
    return false
  }

  for (let rowIndex = 0; rowIndex < payloadRows.length; rowIndex += 1) {
    const payload = payloadRows[rowIndex]
    for (let fieldIndex = 0; fieldIndex < formSchema.value.length; fieldIndex += 1) {
      const field = formSchema.value[fieldIndex]
      if (!field.required) {
        continue
      }
      const fieldKey = resolveFieldKey(field, fieldIndex)
      if (isEmptyFieldValue(payload[fieldKey])) {
        ElMessage.warning(`第 ${rowIndex + 1} 行缺少必填项“${field.label}”`)
        return false
      }
    }
  }

  return true
}

const resetUploadPreview = () => {
  uploadId.value = ''
  uploadFileName.value = ''
  previewColumns.value = []
  previewRows.value = []
}

const buildUploadPreview = (response: unknown) => {
  const record = response as Record<string, unknown>
  uploadId.value = typeof record.id === 'string' ? record.id : ''
  uploadFileName.value = typeof record.excelName === 'string' ? record.excelName : ''

  const formFields = Array.isArray(record.formFields) ? (record.formFields as Array<Record<string, unknown>>) : []
  previewColumns.value = formFields.map((field, index) => ({
    key: String(field.field ?? field.name ?? `field_${index + 1}`),
    label: String(field.label ?? field.name ?? field.field ?? `字段 ${index + 1}`)
  }))

  const dataList = Array.isArray(record.dataList) ? (record.dataList as Array<Record<string, unknown>>) : []
  previewRows.value = dataList.map((item, index) => {
    const data = isRecord(item.data) ? item.data : {}
    return {
      __previewId: item.id ?? index,
      __insert: Boolean(item.insert),
      ...data
    }
  })
}

const resolvePendingInstances = () => {
  if (!taskData.value) {
    return []
  }
  return taskData.value.subInstances.filter(instance => !String(instance.dataId || '').trim())
}

const loadPersistedRows = async (detail: UserTaskData) => {
  if (!detail.dataIds.length) {
    return []
  }

  const searchParams: SearchParam[] = [
    {
      term: 'in',
      field: 'id',
      value: '',
      values: detail.dataIds,
      multiple: true
    }
  ]

  const response = (await searchTableData({
    id: detail.formId,
    currentPage: 1,
    pageSize: Math.max(20, detail.dataIds.length),
    searchParams
  })) as TableDataResponse

  const records = Array.isArray(response?.data) ? response.data : []
  const orderMap = new Map(detail.dataIds.map((dataId, index) => [dataId, index]))
  return records
    .filter(item => {
      const rowId = item?.id
      return typeof rowId === 'string' || typeof rowId === 'number'
    })
    .sort((prev, next) => {
      const prevIndex = orderMap.get(String(prev.id)) ?? Number.MAX_SAFE_INTEGER
      const nextIndex = orderMap.get(String(next.id)) ?? Number.MAX_SAFE_INTEGER
      return prevIndex - nextIndex
    })
}

const rebuildRows = async (detail: UserTaskData) => {
  let persistedRows: Record<string, unknown>[] = []
  if (detail.dataIds.length) {
    try {
      persistedRows = await loadPersistedRows(detail)
    } catch {
      ElMessage.warning('任务信息已加载，但历史填报数据未完整返回，可继续新增或重新提交。')
    }
  }

  const instanceMap = new Map(detail.subInstances.map(instance => [instance.dataId, instance]))
  const savedRows = persistedRows.map(row => {
    const dataId = String(row.id ?? '')
    const matchedInstance = instanceMap.get(dataId)
    return {
      localId: `saved-${dataId}`,
      dataId,
      taskInstanceId: matchedInstance?.id,
      model: buildRowModel(row)
    }
  })

  const blankRows = resolvePendingInstances().map(instance => createEmptyRow(instance.id))
  const fallbackRows = !savedRows.length && !blankRows.length && !singleFillCompleted.value ? [createEmptyRow()] : []
  rows.value = [...savedRows, ...blankRows, ...fallbackRows]
}

const loadTaskDetail = async () => {
  if (!subTaskId.value) {
    ElMessage.error('缺少有效的子任务 ID')
    return
  }

  pageLoading.value = true
  try {
    const detail = (await getUserTaskData(subTaskId.value)) as UserTaskData
    taskData.value = detail
    formTitle.value = detail.formTitle || '数据填报'
    formId.value = detail.formId
    fillType.value = detail.fillType
    formExtSetting.value = detail.formExtSetting || ''
    formSchema.value = parseSchema(detail.form || '')
    await rebuildRows(detail)
  } catch {
    ElMessage.error('加载填报任务失败')
  } finally {
    pageLoading.value = false
  }
}

const handleAddRow = () => {
  if (!canEdit.value) {
    return
  }
  rows.value.push(createEmptyRow())
}

const handleDeleteRow = async (row: FillRowItem) => {
  if (!canEdit.value) {
    return
  }

  try {
    await ElMessageBox.confirm('确认删除当前数据行吗？删除后不可恢复。', '删除确认', {
      type: 'warning',
      autofocus: false
    })
  } catch {
    return
  }

  if (row.dataId) {
    try {
      await deleteUserTaskData(subTaskId.value, row.dataId)
      ElMessage.success('删除成功')
    } catch {
      ElMessage.error('删除失败')
      return
    }
  }

  rows.value = rows.value.filter(item => item.localId !== row.localId)
  if (!rows.value.length && canEdit.value) {
    rows.value.push(createEmptyRow())
  }
}

const handleSubmit = async () => {
  if (!canEdit.value) {
    ElMessage.warning('当前任务已完成，无法再次提交')
    return
  }

  const payloadRows = rows.value
    .map(item => sanitizeRowPayload(item))
    .filter(item => !isBlankPayload(item))

  if (!validateRows(payloadRows)) {
    return
  }

  submitLoading.value = true
  try {
    await saveUserTaskData(subTaskId.value, payloadRows)
    ElMessage.success('填报提交成功')
    router.push({ path: '/data-filling-fill' })
  } catch {
    ElMessage.error('填报提交失败')
  } finally {
    submitLoading.value = false
  }
}

const beforeUpload = (file: File) => {
  const validExtensions = ['.xls', '.xlsx', '.csv']
  const valid = validExtensions.some(ext => file.name.toLowerCase().endsWith(ext))
  if (!valid) {
    ElMessage.warning('仅支持上传 xls、xlsx、csv 文件')
    return false
  }
  if (file.size > 10 * 1024 * 1024) {
    ElMessage.warning('上传文件不能超过 10MB')
    return false
  }
  return true
}

const handleExcelUpload = async (options: { file: File }) => {
  if (!formId.value || !canEdit.value || !beforeUpload(options.file)) {
    return
  }

  uploadLoading.value = true
  try {
    const response = await uploadExcelFile(formId.value, options.file)
    buildUploadPreview(response)
    uploadPreviewVisible.value = true
    ElMessage.success('Excel 解析成功，请确认追加')
  } catch {
    resetUploadPreview()
    ElMessage.error('Excel 上传失败')
  } finally {
    uploadLoading.value = false
  }
}

const handleConfirmUpload = async () => {
  if (!uploadId.value || !subTaskId.value || !formId.value) {
    ElMessage.warning('请先上传 Excel 文件')
    return
  }

  uploadConfirmLoading.value = true
  try {
    await userTaskConfirmUpload(subTaskId.value, formId.value, uploadId.value)
    ElMessage.success('Excel 追加成功')
    uploadPreviewVisible.value = false
    resetUploadPreview()
    await loadTaskDetail()
  } catch {
    ElMessage.error('确认追加失败')
  } finally {
    uploadConfirmLoading.value = false
  }
}

const handleBack = () => {
  router.push({ path: '/data-filling-fill' })
}

const formatFinishTime = (value?: number) => {
  if (!value) {
    return '--'
  }
  return dayjs(value).format('YYYY-MM-DD HH:mm')
}

onMounted(() => {
  loadTaskDetail()
})
</script>

<template>
  <div class="df-fill-task" :class="{ 'df-fill-task--mobile': isMobile }" v-loading="pageLoading">
    <section class="df-fill-task__panel">
      <div class="df-fill-task__header">
        <div class="df-fill-task__heading">
          <div class="df-fill-task__eyebrow">Task Filling</div>
          <h2 class="df-fill-task__title">{{ formTitle || '填报任务' }}</h2>
          <p class="df-fill-task__subtitle">{{ pageDescription }}</p>
        </div>
        <div class="df-fill-task__header-actions">
          <el-button @click="handleBack">返回列表</el-button>
          <el-upload
            action=""
            :http-request="handleExcelUpload"
            :show-file-list="false"
            accept=".xls,.xlsx,.csv"
            :disabled="!canEdit || uploadLoading || uploadConfirmLoading"
          >
            <el-button :disabled="!canEdit" :loading="uploadLoading">Excel 追加</el-button>
          </el-upload>
          <el-button type="primary" :disabled="!canEdit" :loading="submitLoading" @click="handleSubmit">
            提交填报
          </el-button>
        </div>
      </div>

      <el-alert
        v-if="singleFillCompleted"
        title="当前任务为单次填报，您已提交完成。"
        type="success"
        :closable="false"
        show-icon
      />

      <div class="df-fill-task__summary">
        <div class="df-fill-task__summary-item">
          <span class="df-fill-task__summary-label">填报方式</span>
          <span class="df-fill-task__summary-value">{{ fillType === 0 ? '单次填报' : '多次填报' }}</span>
        </div>
        <div class="df-fill-task__summary-item">
          <span class="df-fill-task__summary-label">当前行数</span>
          <span class="df-fill-task__summary-value">{{ rows.length }}</span>
        </div>
        <div class="df-fill-task__summary-item">
          <span class="df-fill-task__summary-label">附加配置</span>
          <span class="df-fill-task__summary-value">{{ formExtSetting ? '已配置' : '未配置' }}</span>
        </div>
      </div>

      <div v-if="taskData?.subInstances?.length" class="df-fill-task__timeline">
        <div v-for="instance in taskData.subInstances" :key="instance.id" class="df-fill-task__timeline-item">
          <div class="df-fill-task__timeline-id">实例 #{{ instance.id }}</div>
          <div class="df-fill-task__timeline-meta">
            <span>数据 ID：{{ instance.dataId || '--' }}</span>
            <span>完成时间：{{ formatFinishTime(instance.finishTime) }}</span>
          </div>
        </div>
      </div>

      <div class="df-fill-task__toolbar">
        <div class="df-fill-task__toolbar-text">逐行填写表单内容，支持保留已填数据后继续补充。</div>
        <el-button :disabled="!canEdit" @click="handleAddRow">新增行</el-button>
      </div>

      <el-empty v-if="!rows.length && !pageLoading" description="暂无填报数据，请先新增一行" :image-size="72" />

      <div v-else class="df-fill-task__rows">
        <article v-for="(row, index) in rows" :key="row.localId" class="df-fill-task__row-card">
          <div class="df-fill-task__row-header">
            <div>
              <div class="df-fill-task__row-title">第 {{ index + 1 }} 行</div>
              <div class="df-fill-task__row-subtitle">
                {{ row.dataId ? `数据 ID：${row.dataId}` : '尚未保存到数据表' }}
              </div>
            </div>
            <el-button text type="danger" :disabled="!canEdit" @click="handleDeleteRow(row)">删除行</el-button>
          </div>

          <FormSchemaRenderer v-model="row.model" :forms="rendererForms" :disabled="!canEdit" />
        </article>
      </div>
    </section>

    <el-dialog
      v-model="uploadPreviewVisible"
      title="Excel 追加预览"
      :width="isMobile ? '100%' : '920px'"
      :fullscreen="isMobile"
      append-to-body
      @closed="resetUploadPreview"
    >
      <div class="df-fill-task__upload-preview">
        <div class="df-fill-task__upload-summary">
          <span>{{ uploadFileName || '待确认文件' }}</span>
          <span>共 {{ previewRows.length }} 行预览数据</span>
        </div>

        <el-table v-if="previewRows.length" :data="previewRows" border stripe max-height="420">
          <el-table-column label="操作类型" width="100">
            <template #default="{ row }">
              {{ row.__insert ? '新增' : '更新' }}
            </template>
          </el-table-column>
          <el-table-column
            v-for="column in previewColumns"
            :key="column.key"
            :prop="column.key"
            :label="column.label"
            min-width="140"
            show-overflow-tooltip
          />
        </el-table>

        <el-empty v-else description="暂无可确认的 Excel 数据" :image-size="72" />
      </div>

      <template #footer>
        <span class="df-fill-task__dialog-footer">
          <el-button @click="uploadPreviewVisible = false">取消</el-button>
          <el-button
            type="primary"
            :loading="uploadConfirmLoading"
            :disabled="!uploadId"
            @click="handleConfirmUpload"
          >
            确认追加
          </el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<style lang="less" scoped>
.df-fill-task {
  --df-fill-task-bg: var(--ed-fill-color-page, #f5f6f7);
  --df-fill-task-panel-bg: var(--ed-fill-color-blank, #ffffff);
  --df-fill-task-text-primary: var(--ed-text-color-primary, #1f2329);
  --df-fill-task-text-secondary: var(--ed-text-color-secondary, #646a73);
  --df-fill-task-border: var(--ed-border-color-light, #dcdfe6);
  --df-fill-task-border-strong: var(--ed-border-color, #d7d7d7);
  --df-fill-task-primary: var(--ed-color-primary, #3370ff);
  --df-fill-task-shadow: 0 12px 30px rgba(31, 35, 41, 0.08);
  --df-fill-task-radius: 12px;
  --df-fill-task-space-xs: 8px;
  --df-fill-task-space-sm: 12px;
  --df-fill-task-space-md: 16px;
  --df-fill-task-space-lg: 24px;
  --df-fill-task-space-xl: 32px;

  min-height: 100%;
  padding: var(--df-fill-task-space-lg);
  background: radial-gradient(circle at top left, rgba(51, 112, 255, 0.08), transparent 24%),
    var(--df-fill-task-bg);

  &__panel {
    display: flex;
    flex-direction: column;
    gap: var(--df-fill-task-space-lg);
    padding: var(--df-fill-task-space-xl);
    background: var(--df-fill-task-panel-bg);
    border: 1px solid var(--df-fill-task-border);
    border-radius: calc(var(--df-fill-task-radius) + 4px);
    box-shadow: var(--df-fill-task-shadow);
  }

  &__header,
  &__header-actions,
  &__toolbar,
  &__upload-summary,
  &__dialog-footer {
    display: flex;
    align-items: center;
  }

  &__header {
    justify-content: space-between;
    align-items: flex-start;
    gap: var(--df-fill-task-space-lg);
  }

  &__heading {
    min-width: 0;
  }

  &__eyebrow {
    margin-bottom: var(--df-fill-task-space-xs);
    color: var(--df-fill-task-primary);
    font-size: 12px;
    font-weight: 600;
    line-height: 18px;
    letter-spacing: 0.12em;
    text-transform: uppercase;
  }

  &__title {
    margin: 0;
    color: var(--df-fill-task-text-primary);
    font-size: 28px;
    line-height: 36px;
    font-weight: 600;
  }

  &__subtitle {
    margin: var(--df-fill-task-space-xs) 0 0;
    color: var(--df-fill-task-text-secondary);
    font-size: 14px;
    line-height: 22px;
  }

  &__header-actions {
    flex-wrap: wrap;
    justify-content: flex-end;
    gap: var(--df-fill-task-space-sm);
  }

  &__summary {
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: var(--df-fill-task-space-md);
  }

  &__summary-item,
  &__timeline-item,
  &__row-card {
    padding: var(--df-fill-task-space-md);
    background: linear-gradient(180deg, rgba(255, 255, 255, 0.96), rgba(245, 246, 247, 0.92));
    border: 1px solid var(--df-fill-task-border);
    border-radius: var(--df-fill-task-radius);
  }

  &__summary-label,
  &__row-subtitle,
  &__toolbar-text,
  &__timeline-meta,
  &__upload-summary {
    color: var(--df-fill-task-text-secondary);
    font-size: 13px;
    line-height: 20px;
  }

  &__summary-value,
  &__timeline-id,
  &__row-title {
    color: var(--df-fill-task-text-primary);
    font-size: 16px;
    line-height: 24px;
    font-weight: 600;
  }

  &__timeline {
    display: grid;
    gap: var(--df-fill-task-space-sm);
  }

  &__timeline-meta {
    display: flex;
    flex-wrap: wrap;
    gap: var(--df-fill-task-space-sm);
    margin-top: 4px;
  }

  &__toolbar {
    justify-content: space-between;
    gap: var(--df-fill-task-space-md);
    padding-bottom: var(--df-fill-task-space-sm);
    border-bottom: 1px dashed var(--df-fill-task-border-strong);
  }

  &__rows {
    display: flex;
    flex-direction: column;
    gap: var(--df-fill-task-space-md);
  }

  &__row-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    gap: var(--df-fill-task-space-md);
    margin-bottom: var(--df-fill-task-space-md);
  }

  &__upload-preview {
    display: flex;
    flex-direction: column;
    gap: var(--df-fill-task-space-md);
  }

  &__upload-summary,
  &__dialog-footer {
    justify-content: space-between;
    gap: var(--df-fill-task-space-sm);
  }

  :deep(.df-form-schema-renderer__grid) {
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: var(--df-fill-task-space-md);
  }

  :deep(.df-form-schema-renderer__item) {
    margin-bottom: 0;
  }

  &--mobile {
    padding: var(--df-fill-task-space-md);

    .df-fill-task__panel {
      padding: var(--df-fill-task-space-lg);
    }

    .df-fill-task__header,
    .df-fill-task__header-actions,
    .df-fill-task__toolbar,
    .df-fill-task__row-header,
    .df-fill-task__upload-summary,
    .df-fill-task__dialog-footer {
      flex-direction: column;
      align-items: stretch;
    }

    .df-fill-task__summary {
      grid-template-columns: 1fr;
    }

    :deep(.df-form-schema-renderer__grid) {
      grid-template-columns: 1fr;
    }
  }
}

@media screen and (max-width: 768px) {
  .df-fill-task {
    &__title {
      font-size: 24px;
      line-height: 32px;
    }

    &__panel {
      border-radius: var(--df-fill-task-radius);
    }
  }
}
</style>
