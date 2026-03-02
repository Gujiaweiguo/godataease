<script lang="ts" setup>
import { ref, reactive, computed } from 'vue'
import { ElDrawer, ElButton } from 'element-plus-secondary'
import DrawerFilter from '@/components/drawer-filter/src/DrawerFilter.vue'
import DrawerEnumFilter from '@/components/drawer-filter/src/DrawerEnumFilter.vue'
import DrawerTimeFilter from '@/components/drawer-filter/src/DrawerTimeFilter.vue'
import DrawerTreeFilter from '@/components/drawer-filter/src/DrawerTreeFilter.vue'
import { useI18n } from '@/hooks/web/useI18n'
const { t } = useI18n()

type TreeOption = {
  value: string
  label: string
  children: TreeOption[]
  disabled: boolean
}

type SelectOption = {
  id: string
  name: string
}

type TreeConfig = {
  checkStrictly: boolean
  showCheckbox: boolean
  checkOnClickNode: boolean
  placeholder: string
}

type SelectConfig = {
  placeholder: string
}

type Config = {
  placeholder: string
  showType: string
  rangeSeparator: string
  startPlaceholder: string
  endPlaceholder: string
  format: string
  valueFormat: string
  size: string
  placement: string
}

type TreeFilterOption = {
  type: 'tree-select'
  field: string
  option: TreeOption[]
  title: string
  property?: TreeConfig
  operator?: string
}

type SelectFilterOption = {
  type: 'select'
  field: string
  option: SelectOption[]
  title: string
  property?: SelectConfig
  operator?: string
}

type EnumFilterOption = {
  type: 'enum'
  field: string
  option: SelectOption[]
  title: string
  operator?: string
}

type TimeFilterOption = {
  type: 'time'
  field: string
  title: string
  property?: Config
  operator?: string
}

type FilterOption = TreeFilterOption | SelectFilterOption | EnumFilterOption | TimeFilterOption

const props = defineProps<{
  filterOptions: FilterOption[]
  title?: string
}>()

const myRefs = ref<Array<unknown>>([])
const componentList = computed(() => {
  return props.filterOptions || []
})

const state = reactive({
  conditions: [] as Array<{ field: string; value: unknown[]; operator?: string }>
})
const userDrawer = ref(false)

const init = () => {
  userDrawer.value = true
}
const cleanrInnerValue = (index?: number) => {
  const field = componentList.value[index]?.field
  if (!field) {
    return
  }
  const targetRef = myRefs.value[index] as { clear?: () => void } | undefined
  targetRef?.clear?.()
  for (let i = 0; i < state.conditions.length; i++) {
    if (state.conditions[i].field === field) {
      state.conditions[i].value = []
    }
  }
}
const clearInnerTag = (index?: number) => {
  if (isNaN(index)) {
    for (let i = 0; i < componentList.value.length; i++) {
      const targetRef = myRefs.value[i] as { clear?: () => void } | undefined
      targetRef?.clear?.()
    }
    return
  }
  const condition = state.conditions[index]
  const field = condition?.field
  for (let i = 0; i < componentList.value.length; i++) {
    if (componentList.value[i].field === field) {
      const targetRef = myRefs.value[i] as { clear?: () => void } | undefined
      targetRef?.clear?.()
    }
  }
}
const clearFilter = (id?: number | null) => {
  clearInnerTag(id)
  if (isNaN(id)) {
    const len = state.conditions.length
    state.conditions.splice(0, len)
  } else {
    state.conditions.splice(id, 1)
  }
  trigger()
}
const filterChange = (value, field, operator) => {
  let exits = false
  let len = state.conditions.length
  while (len--) {
    const condition = state.conditions[len]
    if (condition.field === field) {
      exits = true
      condition['value'] = value
    }
    if (!condition?.value?.length) {
      state.conditions.splice(len, 1)
    }
  }
  if (!exits && value?.length) {
    state.conditions.push({ field, value, operator })
  }
  treeFilterChange(value, field, operator)
}
const reset = () => {
  clearFilter()
  userDrawer.value = false
}
const close = () => {
  userDrawer.value = false
}
const emits = defineEmits(['trigger-filter', 'tree-filter-change'])
const trigger = () => {
  emits('trigger-filter', state.conditions)
}
const treeFilterChange = (value, field, operator) => {
  emits('tree-filter-change', {
    value,
    field,
    operator
  })
}
defineExpose({
  init,
  clearFilter,
  close,
  cleanrInnerValue
})
</script>

<template>
  <el-drawer
    :title="t('common.filter_condition')"
    v-model="userDrawer"
    size="600px"
    modal-class="drawer-main-container"
    direction="rtl"
  >
    <div v-for="(component, index) in componentList" :key="index">
      <drawer-tree-filter
        :ref="el => (myRefs[index] = el)"
        v-if="component.type === 'tree-select'"
        :option-list="component.option"
        :title="component.title"
        :property="component.property"
        @filter-change="v => filterChange(v, component.field, 'in')"
      />
      <drawer-filter
        :ref="el => (myRefs[index] = el)"
        v-if="component.type === 'select'"
        :option-list="component.option"
        :title="component.title"
        :property="component.property"
        @filter-change="v => filterChange(v, component.field, 'in')"
      />
      <drawer-enum-filter
        :ref="el => (myRefs[index] = el)"
        v-if="component.type === 'enum'"
        :option-list="component.option"
        :title="component.title"
        @filter-change="v => filterChange(v, component.field, 'in')"
      />
      <drawer-time-filter
        :ref="el => (myRefs[index] = el)"
        v-if="component.type === 'time'"
        :title="component.title"
        :property="component.property"
        @filter-change="v => filterChange(v, component.field, component.operator)"
      />
    </div>

    <template #footer>
      <el-button secondary @click="reset">{{ t('commons.reset') }}</el-button>
      <el-button @click="trigger" type="primary">{{ t('commons.adv_search.search') }}</el-button>
    </template>
  </el-drawer>
</template>

<style lang="less">
.drawer-main-container {
  .ed-drawer__body {
    padding: 16px 24px 80px !important;
  }
  .ed-drawer__footer {
    padding: 16px 24px;
    height: 64px;
  }
}
</style>
