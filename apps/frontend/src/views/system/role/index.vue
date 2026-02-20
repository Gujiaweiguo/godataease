<template>
  <div class="role-management">
    <div class="header">
      <h2>角色管理</h2>
      <el-button type="primary" @click="handleCreate">新建角色</el-button>
    </div>

    <el-table :data="roleList" border>
      <el-table-column prop="roleName" label="角色名称" />
      <el-table-column prop="roleKey" label="角色标识" />
      <el-table-column prop="roleDesc" label="角色描述" />
      <el-table-column prop="status" label="状态">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'danger'">
            {{ row.status === 1 ? '启用' : '禁用' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="380">
        <template #default="{ row }">
          <el-button link type="primary" @click="handleEdit(row)">编辑</el-button>
          <el-button link type="primary" @click="handlePermissions(row)">权限设置</el-button>
          <el-button link type="primary" @click="handleViewAudit(row)">审计日志</el-button>
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
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus-secondary'
import { queryRoleApi, roleCreateApi, roleUpdateApi, roleDeleteApi } from '@/api/auth'
import { resourceTreeApi, resourcePerSaveApi } from '@/api/auth'

const router = useRouter()

const roleList = ref([])
const dialogVisible = ref(false)
const permDialogVisible = ref(false)
const dialogTitle = ref('')
const currentRole = ref<any>(null)
const permissionTree = ref([])
const selectedPermissions = ref([])

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
  try {
    const res = await queryRoleApi({ current: 1, size: 100 })
    if (res.code === '000000') {
      roleList.value = res.data?.list || []
    }
  } catch (error) {
    ElMessage.error('加载角色列表失败')
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
  const map = new Map()

  permissions.forEach(perm => {
    const node = {
      ...perm,
      children: []
    }
    map.set(perm.permId, node)
  })

  permissions.forEach(perm => {
    if (perm.parentId && map.has(perm.parentId)) {
      const parent = map.get(perm.parentId)
      if (parent) {
        parent.children.push(perm)
      }
    } else {
      tree.push(perm)
    }
  })

  return tree
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

const handlePermissions = (row: any) => {
  currentRole.value = row
  selectedPermissions.value = []
  loadPermissions()
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

const handleViewAudit = (row: any) => {
  router.push({
    path: '/system/audit',
    query: { resourceType: 'ROLE', resourceId: row.roleId }
  })
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
.role-management {
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
