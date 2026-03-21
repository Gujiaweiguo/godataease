<template>
  <div class="data-permission">
    <div class="toolbar">
      <el-select v-model="selectedDatasetId" placeholder="请选择数据集" @change="handleDatasetChange">
        <el-option
          v-for="dataset in datasetList"
          :key="dataset.id"
          :label="dataset.name"
          :value="dataset.id"
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
          <el-select v-model="rowRuleForm.filterType" placeholder="请选择过滤类型">
            <el-option label="按角色" value="role" />
            <el-option label="按用户" value="user" />
            <el-option label="按系统变量" value="variable" />
          </el-select>
        </el-form-item>
        <el-form-item label="目标" prop="targetId">
          <el-select v-model="rowRuleForm.targetId" placeholder="请选择目标">
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
        <el-form-item label="白名单用户">
          <el-select v-model="rowRuleForm.whiteList" multiple placeholder="请选择白名单用户">
            <el-option
              v-for="user in userList"
              :key="user.id"
              :label="user.realName || user.username"
              :value="user.id"
            />
          </el-select>
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
            <el-radio label="disable">禁用</el-radio>
            <el-radio label="mask">脱敏</el-radio>
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
import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus-secondary'
import { queryUserApi, queryRoleApi } from '@/api/auth'
import {
  getDatasetTree,
  listFieldByDatasetGroup,
  rowPermissionList,
  columnPermissionList,
  saveRowPermission,
  saveColumnPermission,
  deleteRowPermission,
  deleteColumnPermission
} from '@/api/dataset'

const activeTab = ref('row')
const loading = ref(false)
const selectedDatasetId = ref(null)
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
  filterValue: '',
  whiteList: []
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

const targetOptions = computed(() => {
  if (rowRuleForm.value.filterType === 'role') {
    return roleList.value
  } else if (rowRuleForm.value.filterType === 'user') {
    return userList.value
  }
  return [
    { id: 'account', name: '账号' },
    { id: 'name', name: '姓名' },
    { id: 'email', name: '邮箱' }
  ]
})

const getFilterTypeName = (type: string) => {
  const map: Record<string, string> = {
    role: '按角色',
    user: '按用户',
    variable: '按系统变量'
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
  loading.value = true
  try {
    const res = await rowPermissionList(1, 100, selectedDatasetId.value)
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
    rowRuleDialogTitle.value = '添加行权限规则'
    rowRuleForm.value = {
      id: null,
      datasetId: selectedDatasetId.value,
      name: '',
      filterType: 'role',
      targetId: null,
      filterField: '',
      filterValue: '',
      whiteList: []
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
  rowRuleDialogTitle.value = '编辑行权限规则'
  rowRuleForm.value = { ...row }
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
    const res = await saveRowPermission(rowRuleForm.value)
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
</script>

<style scoped>
.data-permission {
  padding: 0;
}

.toolbar {
  margin-bottom: 16px;
  display: flex;
  gap: 16px;
}

.permission-tabs {
  background: #fff;
  padding: 16px;
  border-radius: 4px;
}
</style>
