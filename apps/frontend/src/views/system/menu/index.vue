<template>
  <div class="menu-management">
    <div class="header">
      <h2>菜单管理</h2>
      <el-button type="primary" @click="handleCreateRoot">新建根菜单</el-button>
    </div>

    <el-table
      v-loading="loading"
      :data="menuList"
      border
      row-key="id"
      :tree-props="{ children: 'children', hasChildren: 'hasChildren' }"
    >
      <el-table-column prop="name" label="菜单名称" min-width="180" />
      <el-table-column prop="path" label="路径" min-width="180" />
      <el-table-column prop="component" label="组件" min-width="180">
        <template #default="{ row }">
          {{ row.component || '-' }}
        </template>
      </el-table-column>
      <el-table-column prop="icon" label="图标" width="140">
        <template #default="{ row }">
          {{ row.icon || '-' }}
        </template>
      </el-table-column>
      <el-table-column prop="type" label="类型" width="120">
        <template #default="{ row }">
          <el-tag size="small">{{ getTypeLabel(row.type) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="menuSort" label="排序" width="160">
        <template #default="{ row }">
          <el-input-number
            v-model="row.menuSort"
            :min="0"
            :controls="false"
            size="small"
            @change="value => handleSortChange(row, value)"
          />
        </template>
      </el-table-column>
      <el-table-column prop="hidden" label="隐藏" width="120">
        <template #default="{ row }">
          <el-switch :model-value="row.hidden" @change="value => handleHiddenChange(row, value)" />
        </template>
      </el-table-column>
      <el-table-column label="操作" width="320" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="handleCreateChild(row)">新增子菜单</el-button>
          <el-button link type="primary" @click="handleEdit(row)">编辑</el-button>
          <el-button link type="danger" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="560px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="110px">
        <el-form-item label="父级菜单" prop="pid">
          <el-tree-select
            v-model="form.pid"
            :data="menuTreeOptions"
            :props="{ label: 'name', value: 'id', children: 'children' }"
            check-strictly
            clearable
            placeholder="请选择父级菜单"
          />
        </el-form-item>
        <el-form-item label="菜单名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入菜单名称" />
        </el-form-item>
        <el-form-item label="路径" prop="path">
          <el-input v-model="form.path" placeholder="例如：/system/menu 或 menu" />
        </el-form-item>
        <el-form-item label="组件" prop="component">
          <el-input v-model="form.component" placeholder="例如：system/menu" />
        </el-form-item>
        <el-form-item label="图标" prop="icon">
          <el-input v-model="form.icon" placeholder="请输入图标名称" />
        </el-form-item>
        <el-form-item label="菜单类型" prop="type">
          <el-select v-model="form.type" placeholder="请选择菜单类型">
            <el-option :value="0" label="目录" />
            <el-option :value="1" label="菜单" />
            <el-option :value="2" label="按钮" />
          </el-select>
        </el-form-item>
        <el-form-item label="排序" prop="menuSort">
          <el-input-number v-model="form.menuSort" :min="0" />
        </el-form-item>
        <el-form-item label="隐藏" prop="hidden">
          <el-switch v-model="form.hidden" />
        </el-form-item>
        <el-form-item label="布局内" prop="inLayout">
          <el-switch v-model="form.inLayout" />
        </el-form-item>
        <el-form-item label="鉴权" prop="auth">
          <el-switch v-model="form.auth" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus-secondary'
import type { FormInstance } from 'element-plus-secondary'
import {
  menuCreateApi,
  menuDeleteApi,
  menuDetailApi,
  menuQueryApi,
  menuUpdateApi,
  menuUpdateHiddenApi,
  menuUpdateSortApi,
  validateMenuPayload,
  MENU_VALIDATION_MESSAGES,
  type MenuSavePayload
} from '@/api/menu'

type MenuItem = {
  id: number
  pid: number
  type: number
  name: string
  component: string
  menuSort: number
  icon: string
  path: string
  hidden: boolean
  inLayout: boolean
  auth: boolean
  children?: MenuItem[]
}

const loading = ref(false)
const menuList = ref<MenuItem[]>([])
const dialogVisible = ref(false)
const dialogTitle = ref('')
const editingId = ref<number | null>(null)
const formRef = ref<FormInstance>()

const form = ref<MenuSavePayload>({
  pid: 0,
  type: 0,
  name: '',
  component: '',
  menuSort: 0,
  icon: '',
  path: '',
  hidden: false,
  inLayout: true,
  auth: true
})

const rules = {
  name: [{ required: true, message: MENU_VALIDATION_MESSAGES.nameRequired, trigger: 'blur' }],
  path: [{ required: true, message: MENU_VALIDATION_MESSAGES.pathRequired, trigger: 'blur' }],
  type: [{ required: true, message: '请选择菜单类型', trigger: 'change' }],
  menuSort: [{ required: true, message: MENU_VALIDATION_MESSAGES.sortInvalid, trigger: 'change' }]
}

const menuTreeOptions = computed(() => {
  const toOption = (nodes: MenuItem[]): MenuItem[] => {
    return nodes.map(node => ({
      ...node,
      children: node.children?.length ? toOption(node.children) : []
    }))
  }
  return [
    {
      id: 0,
      pid: -1,
      type: 0,
      name: '根节点',
      component: '',
      menuSort: 0,
      icon: '',
      path: '/',
      hidden: false,
      inLayout: true,
      auth: true,
      children: toOption(menuList.value)
    }
  ]
})

const getTypeLabel = (type: number) => {
  const map: Record<number, string> = {
    0: '目录',
    1: '菜单',
    2: '按钮'
  }
  return map[type] || `类型${type}`
}

const normalizeMenus = (items: any[]): MenuItem[] => {
  return (items || []).map(item => ({
    id: Number(item.id),
    pid: Number(item.pid || 0),
    type: Number(item.type || 0),
    name: item.name || '',
    component: item.component || '',
    menuSort: Number(item.menuSort || 0),
    icon: item.icon || item.meta?.icon || '',
    path: item.path || '',
    hidden: !!item.hidden,
    inLayout: item.inLayout !== false,
    auth: item.auth !== false,
    children: normalizeMenus(item.children || [])
  }))
}

const resetForm = () => {
  form.value = {
    pid: 0,
    type: 0,
    name: '',
    component: '',
    menuSort: 0,
    icon: '',
    path: '',
    hidden: false,
    inLayout: true,
    auth: true
  }
}

const loadMenuList = async () => {
  loading.value = true
  try {
    const res = await menuQueryApi()
    if (res.code === '000000') {
      menuList.value = normalizeMenus(res.data || [])
    }
  } catch (_error) {
    ElMessage.error('加载菜单列表失败')
  } finally {
    loading.value = false
  }
}

const handleCreateRoot = () => {
  editingId.value = null
  dialogTitle.value = '新建根菜单'
  resetForm()
  form.value.pid = 0
  dialogVisible.value = true
}

const handleCreateChild = (row: MenuItem) => {
  editingId.value = null
  dialogTitle.value = '新增子菜单'
  resetForm()
  form.value.pid = row.id
  dialogVisible.value = true
}

const handleEdit = async (row: MenuItem) => {
  editingId.value = row.id
  dialogTitle.value = '编辑菜单'
  try {
    const res = await menuDetailApi(row.id)
    if (res.code !== '000000') {
      ElMessage.error('加载菜单详情失败')
      return
    }
    const detail = res.data || {}
    form.value = {
      id: detail.id,
      pid: Number(detail.pid || 0),
      type: Number(detail.type || 0),
      name: detail.name || '',
      component: detail.component || '',
      menuSort: Number(detail.menuSort || 0),
      icon: detail.icon || '',
      path: detail.path || '',
      hidden: !!detail.hidden,
      inLayout: detail.inLayout !== false,
      auth: detail.auth !== false
    }
    dialogVisible.value = true
  } catch (_error) {
    ElMessage.error('加载菜单详情失败')
  }
}

const validateBeforeSubmit = async () => {
  try {
    await formRef.value?.validate()
  } catch (_error) {
    return false
  }
  const result = validateMenuPayload(form.value)
  if (!result.valid) {
    ElMessage.error(result.message || '菜单配置校验失败')
    return false
  }
  if (form.value.component && form.value.component.startsWith('/')) {
    ElMessage.error('组件路径不能以 / 开头')
    return false
  }
  return true
}

const handleSubmit = async () => {
  const valid = await validateBeforeSubmit()
  if (!valid) {
    return
  }
  try {
    const payload: MenuSavePayload = {
      ...(editingId.value ? { id: editingId.value } : {}),
      pid: Number(form.value.pid || 0),
      type: Number(form.value.type || 0),
      name: form.value.name.trim(),
      component: (form.value.component || '').trim(),
      menuSort: Number(form.value.menuSort || 0),
      icon: (form.value.icon || '').trim(),
      path: form.value.path.trim(),
      hidden: !!form.value.hidden,
      inLayout: !!form.value.inLayout,
      auth: !!form.value.auth
    }
    const res = editingId.value ? await menuUpdateApi(payload) : await menuCreateApi(payload)
    if (res.code === '000000') {
      ElMessage.success(editingId.value ? '菜单更新成功' : '菜单创建成功')
      dialogVisible.value = false
      await loadMenuList()
    }
  } catch (_error) {
    ElMessage.error('菜单保存失败')
  }
}

const handleDelete = async (row: MenuItem) => {
  try {
    await ElMessageBox.confirm(`确定删除菜单「${row.name}」吗？`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    const res = await menuDeleteApi(row.id)
    if (res.code === '000000') {
      ElMessage.success('删除成功')
      await loadMenuList()
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败，请先清理子菜单')
    }
  }
}

const handleSortChange = async (row: MenuItem, value: number) => {
  const sort = Number(value)
  if (!Number.isInteger(sort) || sort < 0) {
    ElMessage.error(MENU_VALIDATION_MESSAGES.sortInvalid)
    await loadMenuList()
    return
  }
  try {
    const res = await menuUpdateSortApi({ id: row.id, sort })
    if (res.code === '000000') {
      ElMessage.success('排序更新成功')
      await loadMenuList()
    }
  } catch (_error) {
    ElMessage.error('排序更新失败')
    await loadMenuList()
  }
}

const handleHiddenChange = async (row: MenuItem, value: boolean | string | number) => {
  const hidden = !!value
  try {
    const res = await menuUpdateHiddenApi({ id: row.id, hidden })
    if (res.code === '000000') {
      ElMessage.success('隐藏状态更新成功')
      await loadMenuList()
    }
  } catch (_error) {
    ElMessage.error('隐藏状态更新失败')
    await loadMenuList()
  }
}

onMounted(() => {
  loadMenuList()
})
</script>

<style scoped>
.menu-management {
  padding: 20px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.header h2 {
  margin: 0;
}
</style>
