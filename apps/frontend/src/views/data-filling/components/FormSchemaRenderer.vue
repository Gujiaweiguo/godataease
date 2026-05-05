<script setup lang="ts">
import { computed } from 'vue'
import type {
  FormFieldConfig,
  FormFieldValue
} from '@/views/data-filling/types'
import {
  buildPlaceholder,
  parseFormSchema as parseSchema,
  resolveComponentType
} from '@/views/data-filling/utils/schemaParser'

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
