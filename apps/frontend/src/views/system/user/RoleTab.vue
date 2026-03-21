<template>
  <div class="role-tab">
    <div class="toolbar">
      <el-button type="primary" @click="handleCreate">新建角色</el-button>
    </div>

    <el-table :data="roleList" border v-loading="loading">
      <el-table-column prop="roleName" label="角色名称" />
      <el-table-column prop="roleKey" label="角色标识" />
      <el-table-column prop="roleDesc" label="角色描述" />
      <el-table-column prop="status" label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'danger'">
            {{ row.status === 1 ? '启用' : '禁用' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="450" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="handleEdit(row)">编辑</el-button>
          <el-button link type="primary" @click="handleMenuAuth(row)">菜单授权</el-button>
          <el-button link type="primary" @click="handlePermissions(row)">权限设置</el-button>
          <el-button link type="danger" @click="handleDelete(row.roleId)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="500px">
      <el-form :model="form" :rules="rules" label-width="100px">
        <el-form-item label="角色名称" prop="roleName">
          <el-input v-model="form.roleName" placeholder="请输入角色名称" />
        </el-form-item>
        <el-form-item label="角色标识" prop="roleKey">
          <el-input v-model="form.roleKey" placeholder="请输入角色标识" />
        </el-form-item>
        <el-form-item label="角色描述" prop="roleDesc">
          <el-input v-model="form.roleDesc" type="textarea" placeholder="请输入角色描述" />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-switch v-model="form.status" :active-value="1" :inactive-value="0" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="permDialogVisible" title="权限设置" width="600px">
      <div style="padding: 20px">
        <el-tree
          :data="permissionTree"
          :props="{ label: 'permName', children: 'children' }"
          show-checkbox
          node-key="permId"
          :default-checked-keys="selectedPermissions"
          @check="handlePermissionCheck"
        />
      </div>
      <template #footer>
        <el-button @click="permDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handlePermissionSave">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="menuAuthDialogVisible" title="菜单授权" width="600px">
      <div style="padding: 20px">
        <el-tree
          ref="menuTreeRef"
          :data="menuTree"
          :props="{ label: 'meta title', children: 'children' }"
          show-checkbox
          node-key="id"
          :default-checked-keys="selectedMenuIds"
          @check="handleMenuCheck"
        />
      </div>
      <template #footer>
        <el-button @click="menuAuthDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleMenuAuthSave">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus-secondary'
import {
  queryRoleApi,
  roleCreateApi,
  roleUpdateApi,
  roleDeleteApi,
  resourceTreeApi,
  resourcePerApi,
  resourcePerSaveApi,
  menuTreeApi,
  roleMenuAuthApi,
  roleMenuAuthSaveApi
} from '@/api/auth'

const loading = ref(false)
const roleList = ref([])
const dialogVisible = ref(false)
const permDialogVisible = ref(false)
const menuAuthDialogVisible = ref(false)
const dialogTitle = ref('')
const currentRole = ref<any>(null)
const permissionTree = ref([])
const selectedPermissions = ref([])
const menuTree = ref([])
const selectedMenuIds = ref([])
const menuTreeRef = ref()

const form = ref({
  roleId: null,
  roleName: '',
  roleKey: '',
  roleDesc: '',
  status: 1
})

const rules = {
  roleName: [{ required: true, message: '请输入角色名称', trigger: 'blur' }],
  roleKey: [{ required: true, message: '请输入角色标识', trigger: 'blur' }],
  roleDesc: [{ required: true, message: '请输入角色描述', trigger: 'blur' }]
}

const loadRoleList = async () => {
  loading.value = true
  try {
    const res = await queryRoleApi({ current: 1, size: 100 })
    if (res.code === '000000') {
      roleList.value = res.data?.list || []
    }
  } catch (error) {
    ElMessage.error('加载角色列表失败')
  } finally {
    loading.value = false
  }
}

const loadPermissions = async () => {
  try {
    const res = await resourceTreeApi('1')
    if (res.code === '000000') {
      permissionTree.value = buildPermissionTree(res.data || [])
    }
  } catch (error) {
    ElMessage.error('加载权限列表失败')
  }
}

const buildPermissionTree = (permissions: any[]) => {
  const tree: any[] = []
  const map = new Map<number, any>()

  permissions.forEach((perm: any) => {
    map.set(perm.permId, { ...perm, children: [] })
  })

  permissions.forEach((perm: any) => {
    if (perm.parentId && map.has(perm.parentId)) {
      const parent = map.get(perm.parentId)
      if (parent) {
        parent.children.push(map.get(perm.permId))
      }
    } else {
      tree.push(map.get(perm.permId))
    }
  })

  return tree
}

const loadMenuTree = async () => {
  try {
    const res = await menuTreeApi()
    if (res.code === '000000') {
      menuTree.value = buildMenuTree(res.data || [])
    }
  } catch (error) {
    ElMessage.error('加载菜单列表失败')
  }
}

const buildMenuTree = (menus: any[]) => {
  return menus.map(menu => ({
    id: menu.id,
    path: menu.path,
    meta: { title: menu.meta?.title || menu.name },
    children: menu.children ? buildMenuTree(menu.children) : []
  }))
}

const handleCreate = () => {
  dialogTitle.value = '新建角色'
  form.value = {
    roleId: null,
    roleName: '',
    roleKey: '',
    roleDesc: '',
    status: 1
  }
  dialogVisible.value = true
}

const handleEdit = (row: any) => {
  dialogTitle.value = '编辑角色'
  form.value = { ...row }
  dialogVisible.value = true
}

const handlePermissions = async (row: any) => {
  currentRole.value = row
  selectedPermissions.value = []
  await loadPermissions()
  try {
    const res = await resourcePerApi({ roleId: row.roleId })
    if (res.code === '000000') {
      selectedPermissions.value = res.data?.permIds || []
    }
  } catch (error) {
    ElMessage.error('加载角色权限失败')
  }
  permDialogVisible.value = true
}

const handlePermissionCheck = (checkedKeys: any[]) => {
  selectedPermissions.value = checkedKeys
}

const handlePermissionSave = async () => {
  try {
    const res = await resourcePerSaveApi({
      roleId: currentRole.value.roleId,
      permIds: selectedPermissions.value
    })
    if (res.code === '000000') {
      ElMessage.success('权限设置成功')
      permDialogVisible.value = false
    }
  } catch (error) {
    ElMessage.error('权限设置失败')
  }
}

const handleMenuAuth = async (row: any) => {
  currentRole.value = row
  selectedMenuIds.value = []
  await loadMenuTree()
  try {
    const res = await roleMenuAuthApi(row.roleId)
    if (res.code === '000000') {
      selectedMenuIds.value = res.data?.menuIds || []
    }
  } catch (error) {
    ElMessage.error('加载菜单授权失败')
  }
  menuAuthDialogVisible.value = true
}

const handleMenuCheck = (checkedKeys: any[]) => {
  selectedMenuIds.value = checkedKeys
}

const handleMenuAuthSave = async () => {
  try {
    const res = await roleMenuAuthSaveApi({
      roleId: currentRole.value.roleId,
      menuIds: selectedMenuIds.value
    })
    if (res.code === '000000') {
      ElMessage.success('菜单授权成功')
      menuAuthDialogVisible.value = false
      window.dispatchEvent(new Event('de:permission-refresh'))
    }
  } catch (error) {
    ElMessage.error('菜单授权失败')
  }
}

const handleDelete = async (roleId: number) => {
  try {
    await ElMessageBox.confirm('确定要删除该角色吗?', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })

    const res = await roleDeleteApi(roleId)
    if (res.code === '000000') {
      ElMessage.success('删除成功')
      loadRoleList()
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

const handleSubmit = async () => {
  try {
    let res
    if (form.value.roleId) {
      res = await roleUpdateApi(form.value)
    } else {
      res = await roleCreateApi(form.value)
    }

    if (res.code === '000000') {
      ElMessage.success(form.value.roleId ? '更新成功' : '创建成功')
      dialogVisible.value = false
      loadRoleList()
    }
  } catch (error) {
    ElMessage.error('操作失败')
  }
}

onMounted(() => {
  loadRoleList()
})
</script>

<style scoped>
.role-tab {
  padding: 0;
}

.toolbar {
  margin-bottom: 16px;
  display: flex;
  justify-content: flex-end;
}
</style>
