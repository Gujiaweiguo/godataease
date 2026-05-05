<script setup lang="ts">
import { ref, watch } from 'vue'
import FormFieldEditor from './FormFieldEditor.vue'
import type { DataFillingFieldType, FormFieldConfig } from '@/views/data-filling/types'

const props = withDefaults(
  defineProps<{
    fields: FormFieldConfig[]
  }>(),
  {
    fields: () => []
  }
)

const emit = defineEmits<{
  (e: 'update:fields', value: FormFieldConfig[]): void
}>()

const fieldTypeOptions: Array<{ label: string; value: DataFillingFieldType }> = [
  { label: '文本', value: 'text' },
  { label: '长文本', value: 'nvarchar' },
  { label: '整数', value: 'number' },
  { label: '小数', value: 'decimal' },
  { label: '日期', value: 'date' },
  { label: '日期时间', value: 'datetime' },
  { label: '下拉选择', value: 'select' }
]

const localFields = ref<FormFieldConfig[]>([])

const buildDefaultField = (type: DataFillingFieldType, index: number): FormFieldConfig => {
  const displayName = `字段 ${index + 1}`
  return {
    name: `field_${index + 1}`,
    label: displayName,
    type,
    required: false,
    order: index,
    placeholder: '',
    options: type === 'select' ? [] : undefined,
    precision: type === 'decimal' ? 2 : undefined,
    format: type === 'datetime' ? 'YYYY-MM-DD HH:mm:ss' : type === 'date' ? 'YYYY-MM-DD' : undefined,
    multiple: type === 'select' ? false : undefined
  }
}

const normalizeFields = (fields: FormFieldConfig[]) => {
  return fields.map((field, index) => ({
    ...field,
    name: field.name || `field_${index + 1}`,
    label: field.label || `字段 ${index + 1}`,
    order: index
  }))
}

const syncFields = (fields: FormFieldConfig[]) => {
  localFields.value = normalizeFields(fields)
  emit('update:fields', localFields.value)
}

const addField = (type: DataFillingFieldType) => {
  syncFields([...localFields.value, buildDefaultField(type, localFields.value.length)])
}

const removeField = (index: number) => {
  syncFields(localFields.value.filter((_, fieldIndex) => fieldIndex !== index))
}

const moveField = (index: number, direction: -1 | 1) => {
  const targetIndex = index + direction
  if (targetIndex < 0 || targetIndex >= localFields.value.length) {
    return
  }
  const nextFields = [...localFields.value]
  const current = nextFields[index]
  nextFields[index] = nextFields[targetIndex]
  nextFields[targetIndex] = current
  syncFields(nextFields)
}

const updateField = (index: number, field: FormFieldConfig) => {
  const nextFields = [...localFields.value]
  nextFields[index] = field
  syncFields(nextFields)
}

watch(
  () => props.fields,
  value => {
    localFields.value = normalizeFields(value)
  },
  { deep: true, immediate: true }
)
</script>

<template>
  <div class="df-field-list">
    <div class="df-field-list__header">
      <span class="df-field-list__title">字段配置</span>
      <el-dropdown @command="command => addField(command as DataFillingFieldType)">
        <el-button type="primary">添加字段</el-button>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item v-for="typeItem in fieldTypeOptions" :key="typeItem.value" :command="typeItem.value">
              {{ typeItem.label }}
            </el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>

    <el-empty v-if="!localFields.length" description="暂无字段，点击右上角添加字段" :image-size="72" />

    <div v-else class="df-field-list__items">
      <div v-for="(field, index) in localFields" :key="field.id || `${field.name}-${index}`" class="df-field-list__item-wrap">
        <div class="df-field-list__item-actions">
          <el-button text :disabled="index === 0" @click="moveField(index, -1)">上移</el-button>
          <el-button text :disabled="index === localFields.length - 1" @click="moveField(index, 1)">下移</el-button>
        </div>
        <FormFieldEditor
          :field="field"
          :index="index"
          @update:field="value => updateField(index, value)"
          @remove="removeField(index)"
        />
      </div>
    </div>
  </div>
</template>

<style lang="less" scoped>
.df-field-list {
  display: flex;
  flex-direction: column;
  gap: 16px;

  &__header {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  &__title {
    color: #1f2329;
    font-size: 16px;
    font-weight: 500;
    line-height: 24px;
  }

  &__items {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  &__item-wrap {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  &__item-actions {
    display: flex;
    justify-content: flex-end;
    gap: 8px;
  }
}
</style>
