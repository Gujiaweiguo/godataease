<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import type { SearchParam } from '@/api/datafilling'
import type { DataFillingColumnConfig } from '@/views/data-filling/types'

type SearchFilterValue = string | number | Array<string | number> | null

const props = withDefaults(
  defineProps<{
    columns: DataFillingColumnConfig[]
    modelValue: SearchParam[]
  }>(),
  {
    columns: () => [],
    modelValue: () => []
  }
)

const emit = defineEmits<{
  (e: 'update:modelValue', value: SearchParam[]): void
  (e: 'clear'): void
}>()

const searchableColumns = computed(() => {
  return props.columns.filter(column => column.searchable !== false)
})

const filterState = ref<Record<string, SearchFilterValue>>({})

const isEmptyValue = (value: SearchFilterValue) => {
  if (Array.isArray(value)) {
    return value.length === 0
  }

  return value === null || value === '' || value === undefined
}

const buildFilterState = (
  columns: DataFillingColumnConfig[],
  searchParams: SearchParam[]
): Record<string, SearchFilterValue> => {
  const nextState: Record<string, SearchFilterValue> = {}

  columns.forEach(column => {
    const matchedParam = searchParams.find(item => item.field === column.field)
    if (!matchedParam) {
      nextState[column.field] = column.multiple ? [] : null
      return
    }

    if (matchedParam.multiple) {
      nextState[column.field] = matchedParam.values.filter(
        (item): item is string | number => typeof item === 'string' || typeof item === 'number'
      )
      return
    }

    nextState[column.field] =
      typeof matchedParam.value === 'string' || typeof matchedParam.value === 'number'
        ? matchedParam.value
        : null
  })

  return nextState
}

const resolveTerm = (column: DataFillingColumnConfig, value: SearchFilterValue) => {
  if (Array.isArray(value)) {
    return 'in'
  }

  if (column.type === 'text' || column.type === 'nvarchar') {
    return 'like'
  }

  return 'eq'
}

const buildSearchParam = (column: DataFillingColumnConfig, value: SearchFilterValue): SearchParam | null => {
  if (isEmptyValue(value)) {
    return null
  }

  if (Array.isArray(value)) {
    return {
      term: resolveTerm(column, value),
      field: column.field,
      value: '',
      values: value,
      multiple: true
    }
  }

  return {
    term: resolveTerm(column, value),
    field: column.field,
    value,
    values: [],
    multiple: false
  }
}

const applyFilters = () => {
  const searchParams = searchableColumns.value.reduce<SearchParam[]>((result, column) => {
    const searchParam = buildSearchParam(column, filterState.value[column.field] ?? null)
    if (searchParam) {
      result.push(searchParam)
    }
    return result
  }, [])

  emit('update:modelValue', searchParams)
}

const clearFilters = () => {
  filterState.value = buildFilterState(searchableColumns.value, [])
  emit('update:modelValue', [])
  emit('clear')
}

const handleEnter = (event?: KeyboardEvent) => {
  if (event?.isComposing) {
    return
  }
  applyFilters()
}

watch(
  [searchableColumns, () => props.modelValue],
  ([columns, modelValue]) => {
    filterState.value = buildFilterState(columns, modelValue)
  },
  { deep: true, immediate: true }
)
</script>

<template>
  <div class="df-search-filter">
    <el-form :inline="true" class="df-search-filter__form">
      <el-form-item
        v-for="column in searchableColumns"
        :key="column.field"
        :label="column.label"
        class="df-search-filter__item"
      >
        <el-input
          v-if="column.type !== 'select' && column.type !== 'date' && column.type !== 'datetime'"
          v-model="filterState[column.field]"
          :placeholder="`请输入${column.label}`"
          clearable
          @keydown.enter.exact.prevent="handleEnter"
        />

        <el-select
          v-else-if="column.type === 'select'"
          v-model="filterState[column.field]"
          :placeholder="`请选择${column.label}`"
          :multiple="column.multiple"
          clearable
          filterable
          collapse-tags
          collapse-tags-tooltip
          style="width: 220px"
        >
          <el-option
            v-for="option in column.options"
            :key="option.value"
            :label="option.name"
            :value="option.value"
            :disabled="option.disabled"
          />
        </el-select>

        <el-date-picker
          v-else
          v-model="filterState[column.field]"
          :type="column.type === 'datetime' ? 'datetime' : 'date'"
          :placeholder="`请选择${column.label}`"
          :format="column.format || (column.type === 'datetime' ? 'YYYY-MM-DD HH:mm:ss' : 'YYYY-MM-DD')"
          :value-format="column.format || (column.type === 'datetime' ? 'YYYY-MM-DD HH:mm:ss' : 'YYYY-MM-DD')"
          clearable
        />
      </el-form-item>

      <el-form-item class="df-search-filter__actions">
        <el-button type="primary" @click="applyFilters">搜索</el-button>
        <el-button @click="clearFilters">清空</el-button>
      </el-form-item>
    </el-form>
  </div>
</template>

<style lang="less" scoped>
.df-search-filter {
  width: 100%;

  &__form {
    display: flex;
    flex-wrap: wrap;
    gap: 16px 0;
  }

  &__item {
    margin-right: 16px;
    margin-bottom: 0;

    :deep(.ed-input),
    :deep(.ed-select),
    :deep(.ed-date-editor) {
      min-width: 220px;
    }
  }

  &__actions {
    margin-left: auto;
    margin-right: 0;
    margin-bottom: 0;
  }
}
</style>
