<script setup lang="ts">
import type { DataFillingColumnConfig, DataFillingTableRow } from '@/views/data-filling/types'

const props = withDefaults(
  defineProps<{
    columns: DataFillingColumnConfig[]
    data: DataFillingTableRow[]
    total: number
    currentPage: number
    pageSize: number
    loading?: boolean
    rowKey?: string
    showSelection?: boolean
  }>(),
  {
    columns: () => [],
    data: () => [],
    total: 0,
    currentPage: 1,
    pageSize: 10,
    loading: false,
    rowKey: 'id',
    showSelection: true
  }
)

const emit = defineEmits<{
  (e: 'page-change', value: { currentPage: number; pageSize: number }): void
  (e: 'selection-change', value: DataFillingTableRow[]): void
}>()

const resolveCellValue = (row: DataFillingTableRow, column: DataFillingColumnConfig) => {
  const value = row[column.field]
  if (value == null || value === '') {
    return '--'
  }

  if (column.type === 'select' && column.options?.length) {
    const matchedOption = column.options.find(option => option.value === String(value))
    return matchedOption?.name ?? value
  }

  return value
}

const handleSelectionChange = (selection: DataFillingTableRow[]) => {
  emit('selection-change', selection)
}

const handleCurrentChange = (currentPage: number) => {
  emit('page-change', {
    currentPage,
    pageSize: props.pageSize
  })
}

const handleSizeChange = (pageSize: number) => {
  emit('page-change', {
    currentPage: 1,
    pageSize
  })
}
</script>

<template>
  <div class="df-data-grid">
    <el-table
      :data="data"
      :row-key="rowKey"
      border
      stripe
      v-loading="loading"
      class="df-data-grid__table"
      @selection-change="handleSelectionChange"
    >
      <el-table-column v-if="showSelection" type="selection" width="48" />
      <el-table-column
        v-for="column in columns"
        :key="column.field"
        :prop="column.field"
        :label="column.label"
        :width="column.width"
        :min-width="column.minWidth || 160"
        :align="column.align || 'left'"
        :fixed="column.fixed"
        show-overflow-tooltip
      >
        <template #default="{ row }">
          {{ resolveCellValue(row, column) }}
        </template>
      </el-table-column>

      <slot />

      <template #empty>
        <el-empty description="暂无数据" :image-size="72" />
      </template>
    </el-table>

    <div class="df-data-grid__pagination">
      <el-pagination
        :current-page="currentPage"
        :page-size="pageSize"
        :page-sizes="[10, 20, 50, 100]"
        :total="total"
        background
        layout="total, prev, pager, next, sizes, jumper"
        @current-change="handleCurrentChange"
        @size-change="handleSizeChange"
      />
    </div>
  </div>
</template>

<style lang="less" scoped>
.df-data-grid {
  display: flex;
  width: 100%;
  flex-direction: column;
  gap: 16px;

  &__table {
    width: 100%;
  }

  &__pagination {
    display: flex;
    justify-content: flex-end;
  }
}
</style>
