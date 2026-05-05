<script setup lang="ts">
import { computed, ref } from 'vue'
import { ElMessage } from 'element-plus-secondary'
import {
  confirmUpload,
  downloadExcelTemplate,
  exportFormData,
  uploadExcelFile
} from '@/api/datafilling'

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
  (e: 'confirmed'): void
}>()

const uploadLoading = ref(false)
const confirmLoading = ref(false)
const previewColumns = ref<Array<{ key: string; label: string }>>([])
const previewRows = ref<Array<Record<string, unknown>>>([])
const uploadId = ref('')
const currentFileName = ref('')

const visible = computed({
  get: () => props.modelValue,
  set: value => emit('update:modelValue', value)
})

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

const resetPreview = () => {
  previewColumns.value = []
  previewRows.value = []
  uploadId.value = ''
  currentFileName.value = ''
}

const buildPreview = (response: unknown) => {
  const record = response as Record<string, unknown>
  uploadId.value = typeof record.id === 'string' ? record.id : ''
  currentFileName.value = typeof record.excelName === 'string' ? record.excelName : ''

  const formFields = Array.isArray(record.formFields) ? (record.formFields as Array<Record<string, unknown>>) : []
  previewColumns.value = formFields.map((field, index) => ({
    key: String(field.field ?? field.name ?? `field_${index + 1}`),
    label: String(field.label ?? field.name ?? field.field ?? `字段 ${index + 1}`)
  }))

  const dataList = Array.isArray(record.dataList) ? (record.dataList as Array<Record<string, unknown>>) : []
  previewRows.value = dataList.map((item, index) => {
    const data = item.data && typeof item.data === 'object' ? (item.data as Record<string, unknown>) : {}
    return {
      __previewId: item.id ?? index,
      __insert: Boolean(item.insert),
      ...data
    }
  })
}

const handleUpload = async (options: { file: File }) => {
  if (!beforeUpload(options.file)) {
    return
  }
  uploadLoading.value = true
  try {
    const response = await uploadExcelFile(props.formId, options.file)
    buildPreview(response)
    ElMessage.success('Excel 解析成功，请确认导入')
  } catch {
    resetPreview()
    ElMessage.error('Excel 上传失败')
  } finally {
    uploadLoading.value = false
  }
}

const handleConfirm = async () => {
  if (!uploadId.value) {
    ElMessage.warning('请先上传 Excel 文件')
    return
  }
  confirmLoading.value = true
  try {
    await confirmUpload(props.formId, uploadId.value)
    ElMessage.success('Excel 数据导入成功')
    emit('confirmed')
    visible.value = false
    resetPreview()
  } catch {
    ElMessage.error('确认导入失败')
  } finally {
    confirmLoading.value = false
  }
}

const handleDownloadTemplate = async () => {
  try {
    const blob = await downloadExcelTemplate(props.formId)
    triggerBlobDownload(blob, `data-filling-template-${props.formId}.xlsx`)
  } catch {
    ElMessage.error('模板下载失败')
  }
}

const handleExport = async () => {
  try {
    const blob = await exportFormData(props.formId)
    triggerBlobDownload(blob, `data-filling-export-${props.formId}.xlsx`)
  } catch {
    ElMessage.error('数据导出失败')
  }
}
</script>

<template>
  <el-dialog v-model="visible" title="Excel 导入" width="900px" append-to-body @closed="resetPreview">
    <div class="df-excel-uploader">
      <div class="df-excel-uploader__toolbar">
        <div class="df-excel-uploader__toolbar-left">
          <el-upload
            action=""
            :http-request="handleUpload"
            :show-file-list="false"
            accept=".xls,.xlsx,.csv"
            :disabled="uploadLoading || confirmLoading"
          >
            <el-button :loading="uploadLoading" type="primary">选择并上传文件</el-button>
          </el-upload>
          <span v-if="currentFileName" class="df-excel-uploader__file-name">{{ currentFileName }}</span>
        </div>
        <div class="df-excel-uploader__toolbar-right">
          <el-button @click="handleDownloadTemplate">下载模板</el-button>
          <el-button @click="handleExport">导出数据</el-button>
        </div>
      </div>

      <div v-if="previewRows.length" class="df-excel-uploader__preview">
        <div class="df-excel-uploader__summary">预览 {{ previewRows.length }} 行数据，确认后将写入表单数据。</div>
        <el-table :data="previewRows" border stripe height="420px">
          <el-table-column label="操作类型" width="100">
            <template #default="{ row }">
              {{ row.__insert ? '新增' : '更新' }}
            </template>
          </el-table-column>
          <el-table-column
            v-for="column in previewColumns"
            :key="column.key"
            :label="column.label"
            :prop="column.key"
            min-width="140"
            show-overflow-tooltip
          />
        </el-table>
      </div>

      <el-empty v-else description="请上传 Excel 文件以查看预览" :image-size="72" />
    </div>

    <template #footer>
      <span class="dialog-footer">
        <el-button @click="visible = false">取消</el-button>
        <el-button type="primary" :loading="confirmLoading" :disabled="!uploadId" @click="handleConfirm">
          确认导入
        </el-button>
      </span>
    </template>
  </el-dialog>
</template>

<style lang="less" scoped>
.df-excel-uploader {
  display: flex;
  flex-direction: column;
  gap: 16px;

  &__toolbar {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 16px;
  }

  &__toolbar-left,
  &__toolbar-right {
    display: flex;
    align-items: center;
    gap: 12px;
    flex-wrap: wrap;
  }

  &__file-name {
    color: #646a73;
    font-size: 14px;
    line-height: 22px;
  }

  &__summary {
    color: #646a73;
    font-size: 14px;
    line-height: 22px;
  }
}
</style>
