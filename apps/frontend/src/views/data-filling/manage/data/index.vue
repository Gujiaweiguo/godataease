<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus-secondary'
import { useRoute } from 'vue-router_2'
import type { DataFillingForm, SearchParam, TableDataResponse } from '@/api/datafilling'
import {
  batchDeleteRowData,
  deleteRowData,
  downloadExcelTemplate,
  exportFormData,
  getFormById,
  saveRowData,
  searchTableData,
  truncateTableData
} from '@/api/datafilling'
import DataGrid from '@/views/data-filling/components/DataGrid.vue'
import FormSchemaRenderer from '@/views/data-filling/components/FormSchemaRenderer.vue'
import SearchFilter from '@/views/data-filling/components/SearchFilter.vue'
import type {
  DataFillingColumnConfig,
  DataFillingFormSchema,
  DataFillingTableRow,
  FormFieldConfig,
  FormFieldOption,
  FormFieldValue
} from '@/views/data-filling/types'
import CommitLog from './CommitLog.vue'
import ExcelUploader from './ExcelUploader.vue'

type RowDialogMode = 'add' | 'edit'

const route = useRoute()

const props = withDefaults(
  defineProps<{
    formId?: number
  }>(),
  {
    formId: 0
  }
)

const currentFormId = ref<number | null>(null)
const formDetail = ref<DataFillingForm | null>(null)
const formSchema = ref<DataFillingFormSchema>([])
const columns = ref<DataFillingColumnConfig[]>([])
const tableData = ref<DataFillingTableRow[]>([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(10)
const rowKey = ref('id')
const searchParams = ref<SearchParam[]>([])
const selectedRows = ref<DataFillingTableRow[]>([])

const pageLoading = ref(false)
const tableLoading = ref(false)
const rowSaveLoading = ref(false)
const exportLoading = ref(false)
const templateLoading = ref(false)

const rowDialogVisible = ref(false)
const rowDialogMode = ref<RowDialogMode>('add')
const editingRow = ref<DataFillingTableRow | null>(null)
const rowFormModel = ref<Record<string, FormFieldValue>>({})

const commitLogVisible = ref(false)
const excelUploaderVisible = ref(false)
const tableRenderKey = ref(0)

const pageTitle = computed(() => {
  return formDetail.value?.name ? `${formDetail.value.name} · 数据管理` : 'Data Filling 数据管理'
})

const pageSubtitle = computed(() => {
  if (!currentFormId.value) {
    return '请从管理页选择一个表单后查看数据'
  }

  const parts = [`表单 ID：${currentFormId.value}`]
  if (formDetail.value?.tableName) {
    parts.push(`数据表：${formDetail.value.tableName}`)
  }
  return parts.join(' · ')
})

const pageReady = computed(() => Boolean(currentFormId.value && formDetail.value))
const hasSelection = computed(() => selectedRows.value.length > 0)
const canEditRows = computed(() => pageReady.value && formSchema.value.length > 0)
const rowDialogTitle = computed(() => (rowDialogMode.value === 'add' ? '新增数据行' : '编辑数据行'))

const isRecord = (value: unknown): value is Record<string, unknown> => {
  return value !== null && typeof value === 'object' && !Array.isArray(value)
}

const isCancelableAction = (error: unknown) => {
  return error === 'cancel' || error === 'close'
}

const parseRouteFormId = (value: unknown) => {
  const rawValue = Array.isArray(value) ? value[0] : value
  if (typeof rawValue !== 'string' || !rawValue.trim()) {
    return null
  }

  const parsedValue = Number(rawValue)
  return Number.isFinite(parsedValue) && parsedValue > 0 ? parsedValue : null
}

const normalizeOptions = (value: unknown): FormFieldOption[] => {
  if (!Array.isArray(value)) {
    return []
  }

  return value.reduce<FormFieldOption[]>((result, item) => {
    if (isRecord(item)) {
      const optionName = item.name ?? item.label ?? item.value
      const optionValue = item.value ?? item.name ?? item.label
      if (optionName != null && optionValue != null) {
        result.push({
          name: String(optionName),
          value: String(optionValue),
          disabled: Boolean(item.disabled),
          description: item.description ? String(item.description) : undefined
        })
      }
      return result
    }

    if (typeof item === 'string' || typeof item === 'number') {
      result.push({
        name: String(item),
        value: String(item)
      })
    }

    return result
  }, [])
}

const toFieldConfig = (value: unknown, index: number): FormFieldConfig | null => {
  if (!isRecord(value)) {
    return null
  }

  const label = String(value.label ?? value.name ?? value.field ?? `字段 ${index + 1}`)
  const name = String(value.name ?? value.field ?? value.label ?? `field_${index + 1}`)

  return {
    id: typeof value.id === 'string' || typeof value.id === 'number' ? value.id : undefined,
    field: value.field ? String(value.field) : undefined,
    name,
    label,
    type: String(value.type ?? value.fieldType ?? 'text'),
    required: Boolean(value.required),
    placeholder: value.placeholder ? String(value.placeholder) : undefined,
    defaultValue:
      typeof value.defaultValue === 'string' ||
      typeof value.defaultValue === 'number' ||
      typeof value.defaultValue === 'boolean' ||
      value.defaultValue == null ||
      Array.isArray(value.defaultValue) ||
      isRecord(value.defaultValue)
        ? (value.defaultValue as FormFieldValue)
        : undefined,
    order: typeof value.order === 'number' ? value.order : index,
    options: normalizeOptions(value.options ?? value.optionList),
    optionDatasource: value.optionDatasource ? String(value.optionDatasource) : undefined,
    optionTable: value.optionTable ? String(value.optionTable) : undefined,
    optionColumn: value.optionColumn ? String(value.optionColumn) : undefined,
    optionOrder: value.optionOrder ? String(value.optionOrder) : undefined,
    precision: typeof value.precision === 'number' ? value.precision : undefined,
    format: value.format ? String(value.format) : undefined,
    multiple: Boolean(value.multiple),
    extra: isRecord(value.extra) ? value.extra : undefined
  }
}

const parseFormSchema = (forms: string): DataFillingFormSchema => {
  if (!forms.trim()) {
    return []
  }

  try {
    const parsed = JSON.parse(forms) as unknown
    const source = Array.isArray(parsed)
      ? parsed
      : isRecord(parsed) && Array.isArray(parsed.fields)
      ? parsed.fields
      : isRecord(parsed) && Array.isArray(parsed.forms)
      ? parsed.forms
      : []

    return source
      .map((item, index) => toFieldConfig(item, index))
      .filter((item): item is FormFieldConfig => item !== null)
      .sort((prev, next) => (prev.order ?? 0) - (next.order ?? 0))
  } catch {
    return []
  }
}

const resolveFieldKey = (field: FormFieldConfig, index: number) => {
  return field.field || field.name || `field_${index + 1}`
}

const buildColumns = (schema: DataFillingFormSchema): DataFillingColumnConfig[] => {
  return schema.map((field, index) => ({
    field: resolveFieldKey(field, index),
    label: field.label,
    type: field.type,
    options: field.options,
    multiple: field.multiple,
    format: field.format,
    searchable: true,
    minWidth: field.type === 'datetime' ? 180 : 160
  }))
}

const buildRowFormModel = (row?: DataFillingTableRow | null) => {
  const nextModel: Record<string, FormFieldValue> = {}

  formSchema.value.forEach((field, index) => {
    const fieldKey = resolveFieldKey(field, index)
    const rowValue = row?.[fieldKey]
    if (rowValue !== undefined) {
      nextModel[fieldKey] = rowValue as FormFieldValue
      return
    }
    if (field.defaultValue !== undefined) {
      nextModel[fieldKey] = field.defaultValue
    }
  })

  const rawIdentifier = row?.[rowKey.value]
  if (rawIdentifier !== undefined && rawIdentifier !== null) {
    nextModel[rowKey.value] = rawIdentifier as FormFieldValue
  }

  return nextModel
}

const resolveRowIdentifier = (row: DataFillingTableRow) => {
  const rawIdentifier = row[rowKey.value]
  if (typeof rawIdentifier === 'string' || typeof rawIdentifier === 'number') {
    return rawIdentifier
  }

  const fallbackIdentifier = row.id
  if (typeof fallbackIdentifier === 'string' || typeof fallbackIdentifier === 'number') {
    return fallbackIdentifier
  }

  return null
}

const triggerBlobDownload = (blob: Blob, fileName: string) => {
  const url = window.URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = fileName
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  window.URL.revokeObjectURL(url)
}

const resetSelection = () => {
  selectedRows.value = []
  tableRenderKey.value += 1
}

const resetTableState = () => {
  formDetail.value = null
  formSchema.value = []
  columns.value = []
  tableData.value = []
  total.value = 0
  currentPage.value = 1
  pageSize.value = 10
  rowKey.value = 'id'
  searchParams.value = []
  resetSelection()
}

const loadFormDetail = async (targetFormId: number) => {
  pageLoading.value = true
  try {
    const detail = (await getFormById(targetFormId)) as DataFillingForm
    formDetail.value = detail
    formSchema.value = parseFormSchema(detail?.forms || '')
    columns.value = buildColumns(formSchema.value)
    return true
  } catch {
    formDetail.value = null
    formSchema.value = []
    columns.value = []
    ElMessage.error('加载表单定义失败')
    return false
  } finally {
    pageLoading.value = false
  }
}

const loadTableData = async (page = currentPage.value, size = pageSize.value) => {
  if (!currentFormId.value) {
    return
  }

  tableLoading.value = true
  try {
    const response = (await searchTableData({
      id: currentFormId.value,
      currentPage: page,
      pageSize: size,
      searchParams: searchParams.value
    })) as TableDataResponse

    tableData.value = Array.isArray(response?.data) ? (response.data as DataFillingTableRow[]) : []
    total.value = typeof response?.total === 'number' ? response.total : 0
    currentPage.value = typeof response?.currentPage === 'number' ? response.currentPage : page
    pageSize.value = typeof response?.pageSize === 'number' ? response.pageSize : size
    rowKey.value = response?.key || 'id'
    selectedRows.value = []
  } catch {
    tableData.value = []
    total.value = 0
    ElMessage.error('加载表格数据失败')
  } finally {
    tableLoading.value = false
  }
}

const initPage = async (targetFormId: number | null) => {
  currentFormId.value = targetFormId
  resetTableState()

  if (!targetFormId) {
    return
  }

  currentFormId.value = targetFormId
  const loaded = await loadFormDetail(targetFormId)
  if (!loaded) {
    return
  }

  await loadTableData(1, pageSize.value)
}

const handlePageChange = async (payload: { currentPage: number; pageSize: number }) => {
  currentPage.value = payload.currentPage
  pageSize.value = payload.pageSize
  await loadTableData(payload.currentPage, payload.pageSize)
}

const handleSearchChange = async (value: SearchParam[]) => {
  searchParams.value = value
  currentPage.value = 1
  await loadTableData(1, pageSize.value)
}

const handleSelectionChange = (rows: DataFillingTableRow[]) => {
  selectedRows.value = rows
}

const openAddRowDialog = () => {
  rowDialogMode.value = 'add'
  editingRow.value = null
  rowFormModel.value = buildRowFormModel()
  rowDialogVisible.value = true
}

const openEditRowDialog = (row: DataFillingTableRow) => {
  rowDialogMode.value = 'edit'
  editingRow.value = row
  rowFormModel.value = buildRowFormModel(row)
  rowDialogVisible.value = true
}

const resetRowDialog = () => {
  editingRow.value = null
  rowFormModel.value = {}
  rowDialogMode.value = 'add'
}

const isEmptyFieldValue = (value: FormFieldValue | undefined) => {
  if (Array.isArray(value)) {
    return value.length === 0
  }
  return value === undefined || value === null || value === ''
}

const validateRowForm = () => {
  for (let index = 0; index < formSchema.value.length; index += 1) {
    const field = formSchema.value[index]
    if (!field.required) {
      continue
    }

    const fieldKey = resolveFieldKey(field, index)
    if (isEmptyFieldValue(rowFormModel.value[fieldKey])) {
      ElMessage.warning(`请完善字段“${field.label}”`)
      return false
    }
  }

  return true
}

const buildSavePayload = () => {
  const payload: Record<string, FormFieldValue> = {}

  formSchema.value.forEach((field, index) => {
    const fieldKey = resolveFieldKey(field, index)
    if (Object.prototype.hasOwnProperty.call(rowFormModel.value, fieldKey)) {
      payload[fieldKey] = rowFormModel.value[fieldKey]
    }
  })

  if (rowDialogMode.value === 'edit' && editingRow.value) {
    const rawIdentifier = resolveRowIdentifier(editingRow.value)
    if (rawIdentifier !== null) {
      payload[rowKey.value] = rawIdentifier
    }
  }

  return payload
}

const handleSaveRow = async () => {
  if (!currentFormId.value) {
    ElMessage.warning('缺少表单 ID，无法保存数据')
    return
  }

  if (!validateRowForm()) {
    return
  }

  rowSaveLoading.value = true
  try {
    await saveRowData(currentFormId.value, buildSavePayload())
    ElMessage.success(rowDialogMode.value === 'add' ? '数据行新增成功' : '数据行更新成功')
    rowDialogVisible.value = false
    await loadTableData(currentPage.value, pageSize.value)
  } catch {
    ElMessage.error(rowDialogMode.value === 'add' ? '数据行新增失败' : '数据行更新失败')
  } finally {
    rowSaveLoading.value = false
  }
}

const handleDeleteRow = async (row: DataFillingTableRow) => {
  if (!currentFormId.value) {
    return
  }

  const rawIdentifier = resolveRowIdentifier(row)
  if (rawIdentifier === null) {
    ElMessage.warning('未找到当前数据行主键，无法删除')
    return
  }

  try {
    await ElMessageBox.confirm('确认删除当前数据行吗？该操作不可恢复。', '删除提示', {
      type: 'warning',
      confirmButtonType: 'danger',
      confirmButtonText: '确认删除',
      cancelButtonText: '取消',
      autofocus: false,
      showClose: false
    })

    await deleteRowData(currentFormId.value, String(rawIdentifier))
    ElMessage.success('数据行删除成功')

    const nextPage = tableData.value.length === 1 && currentPage.value > 1 ? currentPage.value - 1 : currentPage.value
    currentPage.value = nextPage
    resetSelection()
    await loadTableData(nextPage, pageSize.value)
  } catch (error) {
    if (!isCancelableAction(error)) {
      ElMessage.error('数据行删除失败')
    }
  }
}

const handleBatchDelete = async () => {
  if (!currentFormId.value || !selectedRows.value.length) {
    return
  }

  const ids = selectedRows.value
    .map(row => resolveRowIdentifier(row))
    .filter((value): value is string | number => value !== null)
    .map(value => String(value))

  if (!ids.length) {
    ElMessage.warning('未找到可删除的数据行主键')
    return
  }

  try {
    await ElMessageBox.confirm(`确认批量删除已选中的 ${ids.length} 条数据吗？`, '批量删除提示', {
      type: 'warning',
      confirmButtonType: 'danger',
      confirmButtonText: '确认删除',
      cancelButtonText: '取消',
      autofocus: false,
      showClose: false
    })

    await batchDeleteRowData(currentFormId.value, { ids })
    ElMessage.success('批量删除成功')

    const nextPage = ids.length >= tableData.value.length && currentPage.value > 1 ? currentPage.value - 1 : currentPage.value
    currentPage.value = nextPage
    resetSelection()
    await loadTableData(nextPage, pageSize.value)
  } catch (error) {
    if (!isCancelableAction(error)) {
      ElMessage.error('批量删除失败')
    }
  }
}

const handleExportData = async () => {
  if (!currentFormId.value) {
    return
  }

  exportLoading.value = true
  try {
    const blob = await exportFormData(currentFormId.value)
    triggerBlobDownload(blob, `data-filling-export-${currentFormId.value}.xlsx`)
  } catch {
    ElMessage.error('数据导出失败')
  } finally {
    exportLoading.value = false
  }
}

const handleDownloadTemplate = async () => {
  if (!currentFormId.value) {
    return
  }

  templateLoading.value = true
  try {
    const blob = await downloadExcelTemplate(currentFormId.value)
    triggerBlobDownload(blob, `data-filling-template-${currentFormId.value}.xlsx`)
  } catch {
    ElMessage.error('模板下载失败')
  } finally {
    templateLoading.value = false
  }
}

const handleTruncate = async () => {
  if (!currentFormId.value) {
    return
  }

  try {
    await ElMessageBox.confirm('此操作将清空当前表单下的全部数据，且无法恢复。是否继续？', '清空数据确认', {
      type: 'warning',
      confirmButtonType: 'danger',
      confirmButtonText: '继续',
      cancelButtonText: '取消',
      autofocus: false,
      showClose: false
    })

    await ElMessageBox.confirm('请再次确认：执行“清空数据”后，所有已填报数据都会被永久删除。', '二次确认', {
      type: 'warning',
      confirmButtonType: 'danger',
      confirmButtonText: '确认清空',
      cancelButtonText: '返回',
      autofocus: false,
      showClose: false
    })

    await truncateTableData(currentFormId.value)
    ElMessage.success('数据已清空')
    currentPage.value = 1
    resetSelection()
    await loadTableData(1, pageSize.value)
  } catch (error) {
    if (!isCancelableAction(error)) {
      ElMessage.error('清空数据失败')
    }
  }
}

const handleExcelConfirmed = async () => {
  currentPage.value = 1
  resetSelection()
  await loadTableData(1, pageSize.value)
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
  <div class="df-data-page" v-loading="pageLoading">
    <div class="df-data-page__header">
      <div>
        <div class="df-data-page__title">{{ pageTitle }}</div>
        <div class="df-data-page__subtitle">{{ pageSubtitle }}</div>
      </div>
      <div class="df-data-page__header-meta">
        <div class="df-data-page__meta-item">
          <span class="label">当前筛选</span>
          <span class="value">{{ searchParams.length }}</span>
        </div>
        <div class="df-data-page__meta-item">
          <span class="label">已选数据</span>
          <span class="value">{{ selectedRows.length }}</span>
        </div>
        <div class="df-data-page__meta-item">
          <span class="label">数据总量</span>
          <span class="value">{{ total }}</span>
        </div>
      </div>
    </div>

    <template v-if="pageReady">
      <el-card shadow="never" class="df-data-page__card df-data-page__toolbar-card">
        <div class="df-data-page__toolbar">
          <div class="df-data-page__toolbar-main">
            <el-button type="primary" :disabled="!canEditRows" @click="openAddRowDialog">新增数据行</el-button>
            <el-button type="danger" plain :disabled="!hasSelection" @click="handleBatchDelete">批量删除</el-button>
            <el-button :loading="exportLoading" @click="handleExportData">导出数据</el-button>
            <el-button :loading="templateLoading" @click="handleDownloadTemplate">下载模板</el-button>
            <el-button @click="excelUploaderVisible = true">上传 Excel</el-button>
          </div>

          <div class="df-data-page__toolbar-side">
            <el-button type="danger" plain @click="handleTruncate">清空数据</el-button>
            <el-button @click="commitLogVisible = true">提交日志</el-button>
          </div>
        </div>
      </el-card>

      <el-card shadow="never" class="df-data-page__card">
        <SearchFilter :columns="columns" :model-value="searchParams" @update:modelValue="handleSearchChange" />
      </el-card>

      <el-card shadow="never" class="df-data-page__card df-data-page__grid-card">
        <DataGrid
          :key="tableRenderKey"
          :columns="columns"
          :data="tableData"
          :total="total"
          :current-page="currentPage"
          :page-size="pageSize"
          :loading="tableLoading"
          :row-key="rowKey"
          @page-change="handlePageChange"
          @selection-change="handleSelectionChange"
        >
          <el-table-column label="操作" fixed="right" width="168" align="center">
            <template #default="{ row }">
              <div class="df-data-page__row-actions">
                <el-button link type="primary" @click="openEditRowDialog(row)">编辑</el-button>
                <el-button link type="danger" @click="handleDeleteRow(row)">删除</el-button>
              </div>
            </template>
          </el-table-column>
        </DataGrid>
      </el-card>
    </template>

    <el-card v-else shadow="never" class="df-data-page__card df-data-page__empty-card">
      <el-empty description="未选择表单，暂时无法加载数据管理页。" :image-size="80" />
    </el-card>

    <el-dialog v-model="rowDialogVisible" :title="rowDialogTitle" width="920px" append-to-body @closed="resetRowDialog">
      <div class="df-data-page__dialog-body">
        <FormSchemaRenderer :forms="formDetail?.forms || ''" v-model="rowFormModel" />
      </div>

      <template #footer>
        <span class="df-data-page__dialog-footer">
          <el-button @click="rowDialogVisible = false">取消</el-button>
          <el-button type="primary" :loading="rowSaveLoading" @click="handleSaveRow">保存</el-button>
        </span>
      </template>
    </el-dialog>

    <CommitLog v-model="commitLogVisible" :form-id="currentFormId || 0" />
    <ExcelUploader v-model="excelUploaderVisible" :form-id="currentFormId || 0" @confirmed="handleExcelConfirmed" />
  </div>
</template>

<style lang="less" scoped>
.df-data-page {
  display: flex;
  min-height: 100%;
  flex-direction: column;
  background: #f5f6f7;
  padding: 24px;
  gap: 16px;

  &__header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 16px;
    border-radius: 4px;
    background: #fff;
    padding: 20px 24px;
  }

  &__title {
    color: #1f2329;
    font-size: 18px;
    font-weight: 500;
    line-height: 28px;
  }

  &__subtitle {
    margin-top: 4px;
    color: #646a73;
    font-size: 14px;
    line-height: 22px;
  }

  &__header-meta {
    display: grid;
    grid-template-columns: repeat(3, minmax(100px, 1fr));
    gap: 12px;
    min-width: 360px;
  }

  &__meta-item {
    display: flex;
    flex-direction: column;
    gap: 4px;
    border: 1px solid #e8eaec;
    border-radius: 4px;
    background: #f5f6f7;
    padding: 12px 16px;

    .label {
      color: #646a73;
      font-size: 12px;
      line-height: 18px;
    }

    .value {
      color: var(--ed-color-primary, #3370ff);
      font-size: 18px;
      font-weight: 500;
      line-height: 28px;
    }
  }

  &__card {
    border: none;
    border-radius: 4px;

    :deep(.el-card__body),
    :deep(.ed-card__body) {
      padding: 20px 24px;
    }
  }

  &__toolbar-card {
    :deep(.el-card__body),
    :deep(.ed-card__body) {
      padding-top: 16px;
      padding-bottom: 16px;
    }
  }

  &__toolbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
  }

  &__toolbar-main,
  &__toolbar-side {
    display: flex;
    align-items: center;
    gap: 12px;
    flex-wrap: wrap;
  }

  &__grid-card {
    flex: 1;
    min-height: 0;
  }

  &__row-actions {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 4px;
  }

  &__empty-card {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  &__dialog-body {
    max-height: 60vh;
    overflow-y: auto;
    padding-right: 4px;
  }

  &__dialog-footer {
    display: inline-flex;
    align-items: center;
    gap: 12px;
  }
}
</style>
