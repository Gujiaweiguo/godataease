<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus-secondary'
import { useRoute, useRouter } from 'vue-router_2'
import icon_back from '@/assets/svg/icon_left_outlined.svg'
import FormSchemaRenderer from '@/views/data-filling/components/FormSchemaRenderer.vue'
import type { DataFillingFormSchema, FormFieldConfig } from '@/views/data-filling/types'
import {
  normalizeBuiltInTableOptions,
  normalizeDatasourceOptions
} from '@/views/data-filling/utils/dataHelpers'
import {
  isRecord,
  normalizeFieldType,
  resolveFieldTypeFromMapping
} from '@/views/data-filling/utils/schemaParser'
import {
  createForm,
  getBuiltInTables,
  getFormById,
  listDatasourceList,
  updateForm
} from '@/api/datafilling'
import FormFieldList from './FormFieldList.vue'

interface DatasourceOption {
  label: string
  value: number
}

interface BuiltInTableOption {
  label: string
  value: string
}

const route = useRoute()
const router = useRouter()

const loading = ref(false)
const datasourceLoading = ref(false)
const builtInTablesLoading = ref(false)
const isEdit = ref(false)
const formId = ref<number | null>(null)
const parentPid = ref<number>(0)
const formName = ref('')
const datasourceId = ref<number | null>(null)
const useExistsTable = ref(false)
const tableName = ref('')
const createIndex = ref(false)
const tableIndexes = ref('')
const fields = ref<FormFieldConfig[]>([])
const datasourceOptions = ref<DatasourceOption[]>([])
const builtInTableOptions = ref<BuiltInTableOption[]>([])

const previewModel = ref<Record<string, string | number | boolean | null>>({})

type BackendFieldOption = {
  name?: unknown
  value?: unknown
}

type BackendFieldMapping = {
  columnName?: unknown
  type?: unknown
  size?: unknown
  accuracy?: unknown
}

type BackendFieldSettings = {
  name?: unknown
  required?: unknown
  mapping?: BackendFieldMapping
  inputType?: unknown
  placeholder?: unknown
  optionDatasource?: unknown
  optionTable?: unknown
  optionColumn?: unknown
  optionOrder?: unknown
  multiple?: unknown
  options?: BackendFieldOption[]
}

const fieldTypeMeta: Record<
  string,
  {
    typeName: string
    mappingType: string
    size?: number
    accuracy?: number
  }
> = {
  text: {
    typeName: '文本',
    mappingType: 'nvarchar',
    size: 255
  },
  nvarchar: {
    typeName: '长文本',
    mappingType: 'nvarchar',
    size: 255
  },
  number: {
    typeName: '整数',
    mappingType: 'number'
  },
  decimal: {
    typeName: '小数',
    mappingType: 'decimal',
    size: 12,
    accuracy: 2
  },
  date: {
    typeName: '日期',
    mappingType: 'date'
  },
  datetime: {
    typeName: '日期时间',
    mappingType: 'datetime'
  },
  select: {
    typeName: '下拉选择',
    mappingType: 'nvarchar',
    size: 255
  }
}

const getFieldTypeMeta = (value: unknown) => {
  const normalizedType = normalizeFieldType(value)
  return {
    normalizedType,
    meta: fieldTypeMeta[normalizedType] ?? fieldTypeMeta.text
  }
}

const normalizeOptions = (value: unknown): FormFieldConfig['options'] => {
  if (!Array.isArray(value)) {
    return undefined
  }

  const options = value.reduce<NonNullable<FormFieldConfig['options']>>((result, item) => {
    if (!isRecord(item)) {
      return result
    }

    const optionName = item.name ?? item.label ?? item.value
    const optionValue = item.value ?? item.name ?? item.label

    if (optionName == null || optionValue == null) {
      return result
    }

    result.push({
      name: String(optionName),
      value: String(optionValue)
    })
    return result
  }, [])

  return options.length ? options : undefined
}

const toFormFieldConfig = (value: unknown, index: number): FormFieldConfig | null => {
  if (!isRecord(value)) {
    return null
  }

  const settings = isRecord(value.settings) ? (value.settings as BackendFieldSettings) : undefined
  const mapping = settings && isRecord(settings.mapping) ? (settings.mapping as BackendFieldMapping) : undefined
  const legacyName = value.name
  const legacyLabel = value.label
  const explicitType = value.type ?? settings?.inputType
  const normalizedType = explicitType != null ? getFieldTypeMeta(explicitType).normalizedType : resolveFieldTypeFromMapping(mapping?.type)
  const fieldName = typeof mapping?.columnName === 'string' && mapping.columnName.trim() ? mapping.columnName.trim() : undefined
  const legacyFieldName = typeof legacyName === 'string' && legacyName.trim() ? legacyName.trim() : undefined
  const labelSource = typeof settings?.name === 'string' && settings.name.trim() ? settings.name.trim() : undefined
  const legacyFieldLabel = typeof legacyLabel === 'string' && legacyLabel.trim() ? legacyLabel.trim() : undefined
  const name = fieldName ?? legacyFieldName ?? `field_${index + 1}`
  const label = labelSource ?? legacyFieldLabel ?? name

  return {
    id: typeof value.id === 'string' || typeof value.id === 'number' ? value.id : undefined,
    name,
    label,
    type: normalizedType,
    required: settings ? Boolean(settings.required) : Boolean(value.required),
    placeholder:
      typeof settings?.placeholder === 'string'
        ? settings.placeholder
        : typeof value.placeholder === 'string'
        ? value.placeholder
        : undefined,
    order: typeof value.order === 'number' ? value.order : index,
    options: normalizeOptions(settings?.options ?? value.options),
    optionDatasource:
      settings?.optionDatasource != null ? String(settings.optionDatasource) : value.optionDatasource ? String(value.optionDatasource) : undefined,
    optionTable:
      typeof settings?.optionTable === 'string'
        ? settings.optionTable
        : typeof value.optionTable === 'string'
        ? value.optionTable
        : undefined,
    optionColumn:
      typeof settings?.optionColumn === 'string'
        ? settings.optionColumn
        : typeof value.optionColumn === 'string'
        ? value.optionColumn
        : undefined,
    optionOrder:
      typeof settings?.optionOrder === 'string'
        ? settings.optionOrder
        : typeof value.optionOrder === 'string'
        ? value.optionOrder
        : undefined,
    precision:
      typeof value.precision === 'number'
        ? value.precision
        : typeof mapping?.accuracy === 'number'
        ? mapping.accuracy
        : undefined,
    multiple: settings ? Boolean(settings.multiple) : Boolean(value.multiple)
  }
}

const parseForms = (forms: string): DataFillingFormSchema => {
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
      .map((item, index) => toFormFieldConfig(item, index))
      .filter((item): item is FormFieldConfig => item !== null)
  } catch {
    return []
  }

  return []
}

const formsJson = computed(() => {
  return JSON.stringify(
    fields.value.map((field, index) => {
      const name = field.name.trim()
      const label = field.label.trim()
      const placeholder = field.placeholder?.trim() || undefined
      const options = field.options?.length
        ? field.options.map(option => ({
            name: String(option.name),
            value: String(option.value)
          }))
        : undefined
      const { normalizedType, meta } = getFieldTypeMeta(field.type)

      return {
        type: normalizedType,
        typeName: meta.typeName,
        icon: '',
        id: field.id != null ? String(field.id) : `field-${index + 1}`,
        settings: {
          name: label,
          required: Boolean(field.required),
          inputType: normalizedType,
          placeholder,
          optionDatasource: field.optionDatasource ? Number(field.optionDatasource) : undefined,
          optionTable: field.optionTable?.trim() || undefined,
          optionColumn: field.optionColumn?.trim() || undefined,
          optionOrder: field.optionOrder?.trim() || undefined,
          multiple: Boolean(field.multiple),
          options,
          mapping: {
            columnName: name,
            type: meta.mappingType,
            size: meta.size,
            accuracy: meta.accuracy
          }
        }
      }
    })
  )
})

const previewFormsJson = computed(() => {
  return JSON.stringify(
    fields.value.map((field, index) => ({
      ...field,
      name: field.name.trim(),
      label: field.label.trim(),
      placeholder: field.placeholder?.trim() || undefined,
      order: index
    }))
  )
})

const pageTitle = computed(() => (isEdit.value ? '编辑 Data Filling 表单' : '新建 Data Filling 表单'))

const canShowIndexSection = computed(() => !useExistsTable.value)

const handleFieldsUpdate = (value: FormFieldConfig[]) => {
  fields.value = value
}

const loadDatasourceOptions = async () => {
  datasourceLoading.value = true
  try {
    const response = await listDatasourceList()
    datasourceOptions.value = normalizeDatasourceOptions(response)
  } catch {
    ElMessage.error('加载数据源失败')
  } finally {
    datasourceLoading.value = false
  }
}

const loadBuiltInTableOptions = async () => {
  if (!datasourceId.value) {
    builtInTableOptions.value = []
    return
  }

  builtInTablesLoading.value = true
  try {
    const response = await getBuiltInTables({ datasourceId: datasourceId.value })
    builtInTableOptions.value = normalizeBuiltInTableOptions(response)
  } catch {
    ElMessage.error('加载内置表失败')
  } finally {
    builtInTablesLoading.value = false
  }
}

const handleDatasourceChange = async (value: number | null) => {
  datasourceId.value = value
  tableName.value = ''
  builtInTableOptions.value = []
  if (useExistsTable.value) {
    await loadBuiltInTableOptions()
  }
}

const handleUseExistingTableChange = async (value: boolean) => {
  useExistsTable.value = value
  tableName.value = ''
  if (value) {
    createIndex.value = false
    tableIndexes.value = ''
    await loadBuiltInTableOptions()
    return
  }
  builtInTableOptions.value = []
}

const backToManage = () => {
  router.push('/data-filling-manage')
}

const validateBeforeSave = () => {
  if (!formName.value.trim()) {
    ElMessage.warning('请输入表单名称')
    return false
  }
  if (!datasourceId.value) {
    ElMessage.warning('请选择数据源')
    return false
  }
  if (!fields.value.length) {
    ElMessage.warning('请至少添加一个字段')
    return false
  }
  if (!useExistsTable.value && !tableName.value.trim()) {
    ElMessage.warning('请输入新建表名')
    return false
  }
  if (useExistsTable.value && !tableName.value.trim()) {
    ElMessage.warning('请选择已有表')
    return false
  }
  return true
}

const handleSave = async () => {
  if (!validateBeforeSave()) {
    return
  }

  loading.value = true
  try {
    const payload = {
      name: formName.value.trim(),
      pid: parentPid.value,
      nodeType: 'form',
      tableName: tableName.value.trim(),
      datasourceId: datasourceId.value ?? undefined,
      forms: formsJson.value,
      createIndex: canShowIndexSection.value ? createIndex.value : false,
      tableIndexes: canShowIndexSection.value && createIndex.value ? tableIndexes.value.trim() : '',
      useExistsTable: useExistsTable.value
    }

    if (isEdit.value && formId.value) {
      await updateForm({
        id: formId.value,
        ...payload
      })
    } else {
      await createForm(payload)
    }

    ElMessage.success('表单保存成功')
    backToManage()
  } catch {
    ElMessage.error('表单保存失败')
  } finally {
    loading.value = false
  }
}

const initEditor = async () => {
  await loadDatasourceOptions()

  const routeFormId = route.query.formId
  const routePid = route.query.pid

  parentPid.value = typeof routePid === 'string' ? Number(routePid) : 0
  formId.value = typeof routeFormId === 'string' ? Number(routeFormId) : null
  isEdit.value = formId.value !== null

  if (!isEdit.value || formId.value === null) {
    return
  }

  loading.value = true
  try {
    const detail = await getFormById(formId.value)
    formName.value = detail?.name || ''
    parentPid.value = typeof detail?.pid === 'number' ? detail.pid : parentPid.value
    datasourceId.value = typeof detail?.datasourceId === 'number' ? detail.datasourceId : null
    tableName.value = detail?.tableName || ''
    useExistsTable.value = Boolean(detail?.useExistsTable)
    createIndex.value = Boolean(detail?.createIndex)
    tableIndexes.value = detail?.tableIndexes || ''
    fields.value = parseForms(detail?.forms || '')

    if (useExistsTable.value) {
      await loadBuiltInTableOptions()
    }
  } catch {
    ElMessage.error('加载表单详情失败')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  void initEditor()
})
</script>

<template>
  <div class="df-form-editor" v-loading="loading">
    <div class="df-form-editor__toolbar">
      <div>
        <div class="df-form-editor__title">{{ pageTitle }}</div>
        <div class="df-form-editor__subtitle">
          {{ isEdit ? `编辑表单 ID：${formId}` : `新建表单，父目录 ID：${parentPid}` }}
        </div>
      </div>
      <div class="df-form-editor__toolbar-actions">
        <el-button @click="backToManage">
          <template #icon>
            <Icon name="icon_left_outlined"><icon_back class="svg-icon" /></Icon>
          </template>
          返回管理页
        </el-button>
        <el-button type="primary" @click="handleSave">保存</el-button>
      </div>
    </div>

    <div class="df-form-editor__content">
      <div class="df-form-editor__main">
        <el-card shadow="never" class="df-form-editor__card">
          <template #header>
            <span>基础配置</span>
          </template>

          <el-form label-position="top">
            <div class="df-form-editor__grid">
              <el-form-item label="表单名称" required>
                <el-input v-model="formName" maxlength="64" placeholder="请输入表单名称" />
              </el-form-item>

              <el-form-item label="数据源" required>
                <el-select
                  :model-value="datasourceId"
                  clearable
                  filterable
                  :loading="datasourceLoading"
                  placeholder="请选择数据源"
                  style="width: 100%"
                  @update:model-value="value => handleDatasourceChange(typeof value === 'number' ? value : null)"
                >
                  <el-option
                    v-for="item in datasourceOptions"
                    :key="item.value"
                    :label="item.label"
                    :value="item.value"
                  />
                </el-select>
              </el-form-item>
            </div>

            <el-form-item label="使用已有表">
              <el-switch :model-value="useExistsTable" @update:model-value="handleUseExistingTableChange" />
            </el-form-item>

            <div class="df-form-editor__grid" v-if="useExistsTable">
              <el-form-item label="已有表" required>
                <el-select
                  v-model="tableName"
                  clearable
                  filterable
                  :loading="builtInTablesLoading"
                  placeholder="请选择已有表"
                  style="width: 100%"
                >
                  <el-option
                    v-for="item in builtInTableOptions"
                    :key="item.value"
                    :label="item.label"
                    :value="item.value"
                  />
                </el-select>
              </el-form-item>
            </div>

            <div class="df-form-editor__grid" v-else>
              <el-form-item label="新建表名" required>
                <el-input v-model="tableName" maxlength="128" placeholder="请输入物理表名" />
              </el-form-item>
            </div>

            <template v-if="canShowIndexSection">
              <el-form-item label="创建索引">
                <el-switch v-model="createIndex" />
              </el-form-item>

              <el-form-item v-if="createIndex" label="索引配置">
                <el-input
                  v-model="tableIndexes"
                  type="textarea"
                  :rows="4"
                  placeholder="请输入索引配置，可使用 JSON 或逗号分隔字段名"
                />
              </el-form-item>
            </template>
          </el-form>
        </el-card>

        <el-card shadow="never" class="df-form-editor__card">
          <FormFieldList :fields="fields" @update:fields="handleFieldsUpdate" />
        </el-card>
      </div>

      <div class="df-form-editor__preview">
        <el-card shadow="never" class="df-form-editor__card df-form-editor__preview-card">
          <template #header>
            <span>表单预览</span>
          </template>
          <FormSchemaRenderer :forms="previewFormsJson" :model-value="previewModel" disabled />
        </el-card>
      </div>
    </div>
  </div>
</template>

<style lang="less" scoped>
.df-form-editor {
  display: flex;
  height: 100%;
  flex-direction: column;
  gap: 24px;
  padding: 24px;
  background: #f5f6f7;

  &__toolbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
    background: #fff;
    border-radius: 4px;
    padding: 16px 24px;
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

  &__toolbar-actions {
    display: flex;
    gap: 12px;
  }

  &__content {
    display: grid;
    grid-template-columns: minmax(0, 2fr) minmax(320px, 1fr);
    gap: 24px;
    min-height: 0;
    flex: 1;
  }

  &__main,
  &__preview {
    display: flex;
    flex-direction: column;
    gap: 24px;
    min-height: 0;
  }

  &__card {
    border-radius: 4px;
  }

  &__grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
    gap: 0 16px;
  }

  &__preview-card {
    height: 100%;
  }
}
</style>
