<script setup lang="ts">
import { computed } from 'vue'
import type {
  DataFillingFormSchema,
  FormFieldConfig,
  FormFieldOption,
  FormFieldValue
} from '@/views/data-filling/types'

interface RenderableField extends FormFieldConfig {
  key: string
  componentType: 'input' | 'number' | 'decimal' | 'date' | 'datetime' | 'select'
  placeholderText: string
  dateFormat: string
  dateValueFormat: string
}

const props = withDefaults(
  defineProps<{
    forms: string
    modelValue: Record<string, FormFieldValue>
    disabled?: boolean
  }>(),
  {
    forms: '',
    modelValue: () => ({}),
    disabled: false
  }
)

const emit = defineEmits<{
  (e: 'update:modelValue', value: Record<string, FormFieldValue>): void
}>()

const isRecord = (value: unknown): value is Record<string, unknown> => {
  return value !== null && typeof value === 'object' && !Array.isArray(value)
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

  const label = String(value.label ?? value.name ?? value.field ?? `Field ${index + 1}`)
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

const parseSchema = (forms: string): DataFillingFormSchema => {
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
  } catch {
    return []
  }
}

const resolveComponentType = (type: string): RenderableField['componentType'] => {
  const normalizedType = type.toLowerCase()
  if (normalizedType === 'number') {
    return 'number'
  }
  if (normalizedType === 'decimal') {
    return 'decimal'
  }
  if (normalizedType === 'date') {
    return 'date'
  }
  if (normalizedType === 'datetime') {
    return 'datetime'
  }
  if (normalizedType === 'select') {
    return 'select'
  }
  return 'input'
}

const buildPlaceholder = (field: FormFieldConfig, componentType: RenderableField['componentType']) => {
  if (field.placeholder) {
    return field.placeholder
  }

  if (componentType === 'select') {
    return `请选择${field.label}`
  }

  if (componentType === 'date' || componentType === 'datetime') {
    return `请选择${field.label}`
  }

  return `请输入${field.label}`
}

const renderedFields = computed<RenderableField[]>(() => {
  return parseSchema(props.forms)
    .sort((prev, next) => (prev.order ?? 0) - (next.order ?? 0))
    .map((field, index) => {
      const componentType = resolveComponentType(field.type)
      const defaultDateFormat = componentType === 'datetime' ? 'YYYY-MM-DD HH:mm:ss' : 'YYYY-MM-DD'

      return {
        ...field,
        key: field.field || field.name || `field_${index + 1}`,
        componentType,
        placeholderText: buildPlaceholder(field, componentType),
        dateFormat: field.format || defaultDateFormat,
        dateValueFormat: field.format || defaultDateFormat
      }
    })
})

const getFieldValue = (field: RenderableField) => {
  return props.modelValue[field.key] ?? field.defaultValue
}

const handleFieldChange = (field: RenderableField, value: FormFieldValue) => {
  emit('update:modelValue', {
    ...props.modelValue,
    [field.key]: value
  })
}
</script>

<template>
  <div class="df-form-schema-renderer">
    <el-empty v-if="!renderedFields.length" description="暂无表单字段" :image-size="72" />
    <el-form v-else :model="modelValue" label-position="top" class="df-form-schema-renderer__form">
      <div class="df-form-schema-renderer__grid">
        <el-form-item
          v-for="field in renderedFields"
          :key="field.key"
          :label="field.label"
          :required="field.required"
          class="df-form-schema-renderer__item"
        >
          <el-input
            v-if="field.componentType === 'input'"
            :model-value="getFieldValue(field)"
            :placeholder="field.placeholderText"
            :disabled="disabled"
            clearable
            @update:model-value="value => handleFieldChange(field, value)"
          />

          <el-input-number
            v-else-if="field.componentType === 'number'"
            :model-value="typeof getFieldValue(field) === 'number' ? getFieldValue(field) : undefined"
            :placeholder="field.placeholderText"
            :disabled="disabled"
            controls-position="right"
            style="width: 100%"
            @update:model-value="value => handleFieldChange(field, value)"
          />

          <el-input-number
            v-else-if="field.componentType === 'decimal'"
            :model-value="typeof getFieldValue(field) === 'number' ? getFieldValue(field) : undefined"
            :placeholder="field.placeholderText"
            :disabled="disabled"
            :precision="field.precision"
            controls-position="right"
            style="width: 100%"
            @update:model-value="value => handleFieldChange(field, value)"
          />

          <el-date-picker
            v-else-if="field.componentType === 'date' || field.componentType === 'datetime'"
            :model-value="typeof getFieldValue(field) === 'string' ? getFieldValue(field) : undefined"
            :type="field.componentType"
            :placeholder="field.placeholderText"
            :disabled="disabled"
            :format="field.dateFormat"
            :value-format="field.dateValueFormat"
            style="width: 100%"
            @update:model-value="value => handleFieldChange(field, value)"
          />

          <el-select
            v-else-if="field.componentType === 'select'"
            :model-value="getFieldValue(field)"
            :placeholder="field.placeholderText"
            :disabled="disabled"
            :multiple="field.multiple"
            clearable
            filterable
            style="width: 100%"
            @update:model-value="value => handleFieldChange(field, value)"
          >
            <el-option
              v-for="option in field.options"
              :key="option.value"
              :label="option.name"
              :value="option.value"
              :disabled="option.disabled"
            />
          </el-select>
        </el-form-item>
      </div>
    </el-form>
  </div>
</template>

<style lang="less" scoped>
.df-form-schema-renderer {
  width: 100%;

  &__form {
    width: 100%;
  }

  &__grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
    gap: 16px;
  }

  &__item {
    margin-bottom: 0;
  }
}
</style>
