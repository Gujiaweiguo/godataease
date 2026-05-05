<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import type { DataFillingFieldType, FormFieldConfig, FormFieldOption } from '@/views/data-filling/types'

const props = defineProps<{
  field: FormFieldConfig
  index: number
}>()

const emit = defineEmits<{
  (e: 'update:field', value: FormFieldConfig): void
  (e: 'remove'): void
}>()

const fieldTypes: Array<{ label: string; value: DataFillingFieldType }> = [
  { label: '文本', value: 'text' },
  { label: '长文本', value: 'nvarchar' },
  { label: '整数', value: 'number' },
  { label: '小数', value: 'decimal' },
  { label: '日期', value: 'date' },
  { label: '日期时间', value: 'datetime' },
  { label: '下拉选择', value: 'select' }
]

const localField = ref<FormFieldConfig>({ ...props.field })
const optionsText = ref('')

const isSelectField = computed(() => localField.value.type === 'select')
const isDecimalField = computed(() => localField.value.type === 'decimal')
const isDateField = computed(() => ['date', 'datetime'].includes(localField.value.type))

const serializeOptions = (options: FormFieldOption[] | undefined) => {
  return (options ?? []).map(option => `${option.name}:${option.value}`).join('\n')
}

const parseOptionsText = (value: string): FormFieldOption[] => {
  return value
    .split('\n')
    .map(line => line.trim())
    .filter(Boolean)
    .map(line => {
      const separatorIndex = line.indexOf(':')
      if (separatorIndex < 0) {
        return {
          name: line,
          value: line
        }
      }

      const label = line.slice(0, separatorIndex).trim()
      const optionValue = line.slice(separatorIndex + 1).trim()
      return {
        name: label || optionValue,
        value: optionValue || label
      }
    })
}

const buildFieldName = (label: string, fallbackName: string) => {
  const normalized = label
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9\u4e00-\u9fa5]+/g, '_')
    .replace(/^_+|_+$/g, '')

  return normalized || fallbackName
}

const emitField = (nextField?: FormFieldConfig) => {
  const field = nextField ?? localField.value
  emit('update:field', {
    ...field,
    name: field.name.trim(),
    label: field.label.trim(),
    placeholder: field.placeholder?.trim() || undefined,
    options: isSelectField.value ? parseOptionsText(optionsText.value) : undefined,
    precision: isDecimalField.value ? field.precision : undefined,
    format: isDateField.value ? field.format : undefined,
    multiple: isSelectField.value ? Boolean(field.multiple) : undefined,
    optionDatasource: isSelectField.value ? field.optionDatasource?.trim() || undefined : undefined,
    optionTable: isSelectField.value ? field.optionTable?.trim() || undefined : undefined,
    optionColumn: isSelectField.value ? field.optionColumn?.trim() || undefined : undefined,
    optionOrder: isSelectField.value ? field.optionOrder?.trim() || undefined : undefined
  })
}

const patchField = (patch: Partial<FormFieldConfig>) => {
  localField.value = {
    ...localField.value,
    ...patch
  }

  if (patch.type && patch.type !== 'select') {
    optionsText.value = ''
  }

  emitField()
}

const handleLabelBlur = () => {
  if (!localField.value.name.trim()) {
    localField.value = {
      ...localField.value,
      name: buildFieldName(localField.value.label, `field_${props.index + 1}`)
    }
  }
  emitField()
}

const handleOptionsBlur = () => {
  emitField({
    ...localField.value,
    options: parseOptionsText(optionsText.value)
  })
}

watch(
  () => props.field,
  value => {
    localField.value = { ...value }
    optionsText.value = serializeOptions(value.options)
  },
  { deep: true, immediate: true }
)
</script>

<template>
  <div class="df-field-editor">
    <div class="df-field-editor__header">
      <span class="df-field-editor__title">字段 {{ index + 1 }}</span>
      <el-button text type="danger" @click="emit('remove')">删除</el-button>
    </div>

    <div class="df-field-editor__grid">
      <el-form-item label="字段类型" required>
        <el-select
          :model-value="localField.type"
          style="width: 100%"
          @update:model-value="value => patchField({ type: value })"
        >
          <el-option v-for="typeItem in fieldTypes" :key="typeItem.value" :label="typeItem.label" :value="typeItem.value" />
        </el-select>
      </el-form-item>

      <el-form-item label="显示名称" required>
        <el-input
          v-model="localField.label"
          maxlength="64"
          placeholder="请输入字段显示名称"
          @blur="handleLabelBlur"
        />
      </el-form-item>

      <el-form-item label="字段编码" required>
        <el-input
          v-model="localField.name"
          maxlength="64"
          placeholder="请输入字段编码"
          @blur="emitField()"
        />
      </el-form-item>

      <el-form-item label="占位提示">
        <el-input
          v-model="localField.placeholder"
          maxlength="128"
          placeholder="可选，占位提示"
          @blur="emitField()"
        />
      </el-form-item>

      <el-form-item label="必填">
        <el-switch :model-value="Boolean(localField.required)" @update:model-value="value => patchField({ required: value })" />
      </el-form-item>

      <el-form-item v-if="isDecimalField" label="小数精度">
        <el-input-number
          :model-value="localField.precision ?? 2"
          :min="0"
          :max="8"
          controls-position="right"
          style="width: 100%"
          @update:model-value="value => patchField({ precision: typeof value === 'number' ? value : 2 })"
        />
      </el-form-item>

      <el-form-item v-if="isDateField" label="日期格式">
        <el-select
          :model-value="localField.format || (localField.type === 'datetime' ? 'YYYY-MM-DD HH:mm:ss' : 'YYYY-MM-DD')"
          style="width: 100%"
          @update:model-value="value => patchField({ format: value })"
        >
          <el-option label="YYYY-MM-DD" value="YYYY-MM-DD" />
          <el-option label="YYYY/MM/DD" value="YYYY/MM/DD" />
          <el-option label="YYYY-MM-DD HH:mm:ss" value="YYYY-MM-DD HH:mm:ss" />
          <el-option label="YYYY/MM/DD HH:mm:ss" value="YYYY/MM/DD HH:mm:ss" />
        </el-select>
      </el-form-item>
    </div>

    <div v-if="isSelectField" class="df-field-editor__select-config">
      <el-form-item label="允许多选">
        <el-switch :model-value="Boolean(localField.multiple)" @update:model-value="value => patchField({ multiple: value })" />
      </el-form-item>

      <div class="df-field-editor__grid">
        <el-form-item label="选项数据源配置">
          <el-input v-model="localField.optionDatasource" placeholder="可选，配置 option datasource" @blur="emitField()" />
        </el-form-item>
        <el-form-item label="选项表">
          <el-input v-model="localField.optionTable" placeholder="可选，配置 option table" @blur="emitField()" />
        </el-form-item>
        <el-form-item label="选项列">
          <el-input v-model="localField.optionColumn" placeholder="可选，配置 option column" @blur="emitField()" />
        </el-form-item>
        <el-form-item label="排序字段">
          <el-input v-model="localField.optionOrder" placeholder="可选，配置 option order" @blur="emitField()" />
        </el-form-item>
      </div>

      <el-form-item label="静态选项">
        <el-input
          v-model="optionsText"
          type="textarea"
          :rows="4"
          placeholder="每行一个选项，格式：标签:值。仅填一个值时默认 label/value 相同"
          @blur="handleOptionsBlur"
        />
      </el-form-item>
    </div>
  </div>
</template>

<style lang="less" scoped>
.df-field-editor {
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  padding: 16px;
  background: #fff;

  &__header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 16px;
  }

  &__title {
    color: #1f2329;
    font-size: 14px;
    font-weight: 500;
    line-height: 22px;
  }

  &__grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
    gap: 0 16px;
  }

  &__select-config {
    margin-top: 8px;
    padding-top: 8px;
    border-top: 1px dashed #dcdfe6;
  }
}
</style>
