<template>
  <div class="data-permission">
    <div class="toolbar">
      <el-radio-group v-if="activeTab === 'row'" v-model="rowViewMode">
        <el-radio-button value="dataset">按数据集</el-radio-button>
        <el-radio-button value="role">按角色</el-radio-button>
      </el-radio-group>
      <el-select v-model="selectedDatasetId" placeholder="请选择数据集" @change="handleDatasetChange">
        <el-option
          v-for="dataset in datasetList"
          :key="dataset.id"
          :label="dataset.name"
          :value="dataset.id"
        />
      </el-select>
      <el-select
        v-if="activeTab === 'row' && rowViewMode === 'role'"
        v-model="selectedRoleFilterId"
        placeholder="请选择角色"
        @change="loadRowRules"
      >
        <el-option
          v-for="role in roleList"
          :key="role.roleId"
          :label="role.roleName"
          :value="role.roleId"
        />
      </el-select>
      <el-button type="primary" @click="handleAddRule" :disabled="!selectedDatasetId">添加规则</el-button>
    </div>

    <el-empty v-if="!selectedDatasetId" description="请选择数据集" />

    <el-tabs v-else v-model="activeTab" class="permission-tabs">
      <el-tab-pane label="行权限" name="row">
        <el-table :data="rowRules" border v-loading="loading">
          <el-table-column prop="name" label="规则名称" />
          <el-table-column prop="filterType" label="过滤类型" width="120">
            <template #default="{ row }">
              <el-tag size="small">{{ getFilterTypeName(row.filterType) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="target" label="目标" width="150">
            <template #default="{ row }">
              {{ getTargetName(row) }}
            </template>
          </el-table-column>
          <el-table-column prop="filterField" label="过滤字段" />
          <el-table-column prop="filterValue" label="过滤值" />
          <el-table-column label="操作" width="200" fixed="right">
            <template #default="{ row }">
              <el-button link type="primary" @click="handleEditRowRule(row)">编辑</el-button>
              <el-button link type="danger" @click="handleDeleteRowRule(row.id)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <el-tab-pane label="列权限" name="column">
        <div class="column-scope-desc">列权限当前按数据集统一治理，不区分角色单独视角。</div>
        <el-table :data="columnRules" border v-loading="loading">
          <el-table-column prop="fieldName" label="字段名称" />
          <el-table-column prop="fieldType" label="字段类型" width="120" />
          <el-table-column prop="ruleType" label="规则类型" width="120">
            <template #default="{ row }">
              <el-tag :type="row.ruleType === 'disable' ? 'danger' : 'warning'">
                {{ row.ruleType === 'disable' ? '禁用' : '脱敏' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="maskRule" label="脱敏规则" />
          <el-table-column label="操作" width="200" fixed="right">
            <template #default="{ row }">
              <el-button link type="primary" @click="handleEditColumnRule(row)">编辑</el-button>
              <el-button link type="danger" @click="handleDeleteColumnRule(row.id)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>
    </el-tabs>

    <el-dialog v-model="rowRuleDialogVisible" :title="rowRuleDialogTitle" width="600px">
      <el-form :model="rowRuleForm" :rules="rowRuleRules" label-width="100px">
        <el-form-item label="规则名称" prop="name">
          <el-input v-model="rowRuleForm.name" placeholder="请输入规则名称" />
        </el-form-item>
        <el-form-item label="过滤类型" prop="filterType">
          <el-select v-model="rowRuleForm.filterType" placeholder="请选择过滤类型" :disabled="isRowRoleMode">
            <el-option label="按角色" value="role" />
            <el-option label="按用户" value="user" />
          </el-select>
        </el-form-item>
        <el-form-item label="目标" prop="targetId">
          <el-select v-model="rowRuleForm.targetId" placeholder="请选择目标" :disabled="isRowRoleMode">
            <el-option
              v-for="item in targetOptions"
              :key="item.id"
              :label="item.name"
              :value="item.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="过滤字段" prop="filterField">
          <el-select v-model="rowRuleForm.filterField" placeholder="请选择过滤字段">
            <el-option
              v-for="field in datasetFields"
              :key="field.fieldName"
              :label="field.fieldName"
              :value="field.fieldName"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="过滤值" prop="filterValue">
          <el-input v-model="rowRuleForm.filterValue" placeholder="请输入过滤值" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="rowRuleDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleRowRuleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="columnRuleDialogVisible" :title="columnRuleDialogTitle" width="600px">
      <el-form :model="columnRuleForm" :rules="columnRuleRules" label-width="100px">
        <el-form-item label="字段名称" prop="fieldName">
          <el-select v-model="columnRuleForm.fieldName" placeholder="请选择字段">
            <el-option
              v-for="field in datasetFields"
              :key="field.fieldName"
              :label="field.fieldName"
              :value="field.fieldName"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="规则类型" prop="ruleType">
            <el-radio-group v-model="columnRuleForm.ruleType">
              <el-radio value="disable">禁用</el-radio>
              <el-radio value="mask">脱敏</el-radio>
            </el-radio-group>
        </el-form-item>
        <el-form-item label="脱敏规则" v-if="columnRuleForm.ruleType === 'mask'">
          <el-select v-model="columnRuleForm.maskRule" placeholder="请选择脱敏规则">
            <el-option label="全部隐藏" value="all" />
            <el-option label="保留首尾" value="keep_ends" />
            <el-option label="自定义位置" value="custom" />
          </el-select>
        </el-form-item>
        <el-form-item label="起始位置" v-if="columnRuleForm.maskRule === 'custom'">
          <el-input-number v-model="columnRuleForm.maskStart" :min="0" />
        </el-form-item>
        <el-form-item label="结束位置" v-if="columnRuleForm.maskRule === 'custom'">
          <el-input-number v-model="columnRuleForm.maskEnd" :min="0" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="columnRuleDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleColumnRuleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus-secondary'
import { queryUserApi, queryRoleApi } from '@/api/auth'
import {
  getDatasetTree,
  listFieldByDatasetGroup,
  rowPermissionList,
  rowPermissionListByTarget,
  columnPermissionList,
  saveRowPermission,
  saveColumnPermission,
  deleteRowPermission,
  deleteColumnPermission
} from '@/api/dataset'

const activeTab = ref('row')
const rowViewMode = ref<'dataset' | 'role'>('dataset')
const loading = ref(false)
const selectedDatasetId = ref(null)
const selectedRoleFilterId = ref<number | null>(null)
const datasetList = ref([])
const datasetFields = ref([])
const userList = ref([])
const roleList = ref([])

const rowRules = ref([])
const columnRules = ref([])

const rowRuleDialogVisible = ref(false)
const columnRuleDialogVisible = ref(false)
const rowRuleDialogTitle = ref('')
const columnRuleDialogTitle = ref('')

const rowRuleForm = ref({
  id: null,
  datasetId: null,
  name: '',
  filterType: 'role',
  targetId: null,
  filterField: '',
  filterValue: ''
})

const columnRuleForm = ref({
  id: null,
  datasetId: null,
  fieldName: '',
  fieldType: '',
  ruleType: 'disable',
  maskRule: 'all',
  maskStart: 0,
  maskEnd: 0
})

const rowRuleRules = {
  name: [{ required: true, message: '请输入规则名称', trigger: 'blur' }],
  filterType: [{ required: true, message: '请选择过滤类型', trigger: 'change' }],
  targetId: [{ required: true, message: '请选择目标', trigger: 'change' }],
  filterField: [{ required: true, message: '请选择过滤字段', trigger: 'change' }],
  filterValue: [{ required: true, message: '请输入过滤值', trigger: 'blur' }]
}

const columnRuleRules = {
  fieldName: [{ required: true, message: '请选择字段', trigger: 'change' }],
  ruleType: [{ required: true, message: '请选择规则类型', trigger: 'change' }]
}

const isRowRoleMode = computed(() => activeTab.value === 'row' && rowViewMode.value === 'role')

const targetOptions = computed(() => {
  if (rowRuleForm.value.filterType === 'role') {
    return roleList.value
  } else if (rowRuleForm.value.filterType === 'user') {
    return userList.value
  }
  return []
})

const getFilterTypeName = (type: string) => {
  const map: Record<string, string> = {
    role: '按角色',
    user: '按用户',
    variable: '按系统变量（暂不支持编辑）'
  }
  return map[type] || type
}

const getTargetName = (row: any) => {
  if (row.filterType === 'role') {
    const role = roleList.value.find((r: any) => r.roleId === row.targetId)
    return role?.roleName || row.targetId
  } else if (row.filterType === 'user') {
    const user = userList.value.find((u: any) => u.id === row.targetId)
    return user?.realName || user?.username || row.targetId
  }
  return row.targetId
}

const handleDatasetChange = () => {
  loadDatasetFields()
  loadRowRules()
  loadColumnRules()
}

const loadDatasetFields = async () => {
  if (!selectedDatasetId.value) return
  try {
    const res = await listFieldByDatasetGroup(selectedDatasetId.value)
    if (res.code === '000000') {
      datasetFields.value = (res.data || []).map((f: any) => ({
        fieldName: f.originName || f.name,
        fieldType: f.fieldType
      }))
    }
  } catch (error) {
    datasetFields.value = []
  }
}

const loadRowRules = async () => {
  if (!selectedDatasetId.value) return
  if (rowViewMode.value === 'role' && !selectedRoleFilterId.value) {
    rowRules.value = []
    return
  }
  loading.value = true
  try {
    const res =
      rowViewMode.value === 'role'
        ? await rowPermissionListByTarget(1, 100, selectedDatasetId.value, 'role', selectedRoleFilterId.value)
        : await rowPermissionList(1, 100, selectedDatasetId.value)
    if (res.code === '000000') {
      rowRules.value = res.data?.list || []
    }
  } catch (error) {
    rowRules.value = []
  } finally {
    loading.value = false
  }
}

const loadColumnRules = async () => {
  if (!selectedDatasetId.value) return
  loading.value = true
  try {
    const res = await columnPermissionList(1, 100, selectedDatasetId.value)
    if (res.code === '000000') {
      columnRules.value = res.data?.list || []
    }
  } catch (error) {
    columnRules.value = []
  } finally {
    loading.value = false
  }
}

const handleAddRule = () => {
  if (activeTab.value === 'row') {
    if (isRowRoleMode.value && !selectedRoleFilterId.value) {
      ElMessage.warning('请先选择角色')
      return
    }
    rowRuleDialogTitle.value = '添加行权限规则'
    rowRuleForm.value = {
      id: null,
      datasetId: selectedDatasetId.value,
      name: '',
      filterType: isRowRoleMode.value ? 'role' : 'role',
      targetId: isRowRoleMode.value ? selectedRoleFilterId.value : null,
      filterField: '',
      filterValue: ''
    }
    rowRuleDialogVisible.value = true
  } else {
    columnRuleDialogTitle.value = '添加列权限规则'
    columnRuleForm.value = {
      id: null,
      datasetId: selectedDatasetId.value,
      fieldName: '',
      fieldType: '',
      ruleType: 'disable',
      maskRule: 'all',
      maskStart: 0,
      maskEnd: 0
    }
    columnRuleDialogVisible.value = true
  }
}

const handleEditRowRule = (row: any) => {
	if (row.filterType === 'variable') {
		ElMessage.warning('系统变量行权限暂不支持在权限中心编辑，请使用原有变量配置入口维护')
		return
	}
  if (isRowRoleMode.value && row.filterType !== 'role') {
    ElMessage.warning('当前角色视角仅支持编辑角色行权限规则')
    return
  }
  rowRuleDialogTitle.value = '编辑行权限规则'
  const formRow = { ...row }
  delete formRow.whiteList
  rowRuleForm.value = {
    ...formRow,
    filterType: isRowRoleMode.value ? 'role' : row.filterType,
    targetId: isRowRoleMode.value ? selectedRoleFilterId.value : row.targetId
  }
  rowRuleDialogVisible.value = true
}

const handleEditColumnRule = (row: any) => {
  columnRuleDialogTitle.value = '编辑列权限规则'
  columnRuleForm.value = { ...row }
  columnRuleDialogVisible.value = true
}

const handleDeleteRowRule = async (id: number) => {
  try {
    await ElMessageBox.confirm('确定要删除该行权限规则吗?', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    const res = await deleteRowPermission({ id })
    if (res.code === '000000') {
      ElMessage.success('删除成功')
      loadRowRules()
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

const handleDeleteColumnRule = async (id: number) => {
  try {
    await ElMessageBox.confirm('确定要删除该列权限规则吗?', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    const res = await deleteColumnPermission({ id })
    if (res.code === '000000') {
      ElMessage.success('删除成功')
      loadColumnRules()
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

const handleRowRuleSubmit = async () => {
  try {
    if (isRowRoleMode.value && !selectedRoleFilterId.value) {
      ElMessage.warning('请先选择角色')
      return
    }
    const rowRuleFormValue = { ...rowRuleForm.value } as Record<string, unknown>
    delete rowRuleFormValue.whiteList
    const payload = {
      ...rowRuleFormValue,
      filterType: isRowRoleMode.value ? 'role' : rowRuleForm.value.filterType,
      targetId: isRowRoleMode.value ? selectedRoleFilterId.value : rowRuleForm.value.targetId
    }
    const res = await saveRowPermission(payload)
    if (res.code === '000000') {
      ElMessage.success('保存成功')
      rowRuleDialogVisible.value = false
      loadRowRules()
    }
  } catch (error) {
    ElMessage.error('保存失败')
  }
}

const handleColumnRuleSubmit = async () => {
  try {
    const res = await saveColumnPermission(columnRuleForm.value)
    if (res.code === '000000') {
      ElMessage.success('保存成功')
      columnRuleDialogVisible.value = false
      loadColumnRules()
    }
  } catch (error) {
    ElMessage.error('保存失败')
  }
}

const loadDatasetList = async () => {
  try {
    const res = await getDatasetTree({ leaf: false, weight: 0 })
    if (res) {
      datasetList.value = flattenDatasetTree(res as unknown as any[])
    }
  } catch (error) {
    datasetList.value = []
  }
}

const flattenDatasetTree = (nodes: any[]): any[] => {
  const result: any[] = []
  for (const node of nodes) {
    if (node.leaf) {
      result.push(node)
    }
    if (node.children) {
      result.push(...flattenDatasetTree(node.children))
    }
  }
  return result
}

const loadUserList = async () => {
  try {
    const res = await queryUserApi({ current: 1, size: 1000 })
    if (res.code === '000000') {
      userList.value = res.data?.list || []
    }
  } catch (error) {
    userList.value = []
  }
}

const loadRoleList = async () => {
  try {
    const res = await queryRoleApi({ current: 1, size: 100 })
    if (res.code === '000000') {
      roleList.value = res.data?.list || []
    }
  } catch (error) {
    roleList.value = []
  }
}

onMounted(() => {
  loadDatasetList()
  loadUserList()
  loadRoleList()
})

watch(rowViewMode, mode => {
  if (mode === 'dataset') {
    selectedRoleFilterId.value = null
    if (rowRuleDialogVisible.value) {
      rowRuleForm.value.filterType = 'role'
      rowRuleForm.value.targetId = null
    }
  } else if (rowRuleDialogVisible.value) {
    rowRuleForm.value.filterType = 'role'
    rowRuleForm.value.targetId = selectedRoleFilterId.value
  }
  loadRowRules()
})

watch(activeTab, tab => {
  if (tab === 'row') {
    loadRowRules()
  }
})
</script>

<style scoped>
.data-permission {
  padding: 0;
}

.toolbar {
  display: flex;
  margin-bottom: 16px;
  gap: 16px;
}

.permission-tabs {
  padding: 16px;
  background: #fff;
  border-radius: 4px;
}

.column-scope-desc {
  margin-bottom: 12px;
  font-size: 13px;
  line-height: 20px;
  color: var(--ed-text-color-secondary);
}
</style>
