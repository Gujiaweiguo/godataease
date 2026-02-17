<template>
  <div class="user-management">
    <div class="header">
      <h2>用户管理</h2>
      <el-button type="primary" @click="handleCreate">新建用户</el-button>
    </div>

    <el-table :data="userList" border>
      <el-table-column prop="username" label="用户名" />
      <el-table-column prop="realName" label="姓名" />
      <el-table-column prop="email" label="邮箱" />
      <el-table-column prop="phone" label="手机号" />
      <el-table-column prop="status" label="状态">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'danger'">
            {{ row.status === 1 ? '启用' : '禁用' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="280">
        <template #default="{ row }">
          <el-button link type="primary" @click="handleEdit(row)">编辑</el-button>
          <el-button link type="primary" @click="handleViewAudit(row)">审计日志</el-button>
          <el-button link type="danger" @click="handleDelete(row.id)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="500px">
      <el-form :model="form" :rules="rules" label-width="100px">
        <el-form-item label="用户名" prop="username">
          <el-input v-model="form.username" placeholder="请输入用户名" />
        </el-form-item>
        <el-form-item label="姓名" prop="realName">
          <el-input v-model="form.realName" placeholder="请输入姓名" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="form.email" placeholder="请输入邮箱" />
        </el-form-item>
        <el-form-item label="密码" prop="password" v-if="!form.id">
          <el-input v-model="form.password" type="password" placeholder="请输入密码" />
        </el-form-item>
        <el-form-item label="组织" prop="organizationId">
          <el-select v-model="form.organizationId" placeholder="请选择组织">
            <el-option
              v-for="org in organizationList"
              :key="org.orgId"
              :label="org.orgName"
              :value="org.orgId"
            />
          </el-select>
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
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus-secondary'
import { queryUserApi, userCreateApi, userUpdateApi, userDeleteApi } from '@/api/auth'
import { queryUserOptionsApi } from '@/api/org'

const router = useRouter()

const userList = ref([])
const dialogVisible = ref(false)
const dialogTitle = ref('')
const form = ref({
  id: null,
  username: '',
  realName: '',
  email: '',
  password: '',
  organizationId: null,
  status: 1
})
const organizationList = ref([])

const rules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  realName: [{ required: true, message: '请输入姓名', trigger: 'blur' }],
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '请输入正确的邮箱格式', trigger: 'blur' }
  ],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
  organizationId: [{ required: true, message: '请选择组织', trigger: 'change' }]
}

const loadUserList = async () => {
  try {
    const res = await queryUserApi({ current: 1, size: 100 })
    if (res.code === '000000') {
      userList.value = res.data?.list || []
    }
  } catch (error) {
    ElMessage.error('加载用户列表失败')
  }
}

const loadOrganizationList = async () => {
  try {
    const res = await queryUserOptionsApi()
    if (res.code === '000000') {
      organizationList.value = res.data || []
    }
  } catch (error) {
    ElMessage.error('加载组织列表失败')
  }
}

const handleCreate = () => {
  dialogTitle.value = '新建用户'
  form.value = {
    id: null,
    username: '',
    realName: '',
    email: '',
    password: '',
    organizationId: null,
    status: 1
  }
  dialogVisible.value = true
}

const handleEdit = (row: any) => {
  dialogTitle.value = '编辑用户'
  form.value = { ...row, password: '' }
  dialogVisible.value = true
}

const handleDelete = async (id: number) => {
  try {
    await ElMessageBox.confirm('确定要删除该用户吗?', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })

    const res = await userDeleteApi(id)
    if (res.code === '000000') {
      ElMessage.success('删除成功')
      loadUserList()
    }
  } catch (error) {
    ElMessage.error('删除失败')
  }
}

const handleViewAudit = (row: any) => {
  router.push({
    path: '/system/audit',
    query: { userId: row.userId, username: row.username }
  })
}

const handleSubmit = async () => {
  try {
    let res
    if (form.value.id) {
      res = await userUpdateApi(form.value)
    } else {
      res = await userCreateApi(form.value)
    }

    if (res.code === '000000') {
      ElMessage.success(form.value.id ? '更新成功' : '创建成功')
      dialogVisible.value = false
      loadUserList()
    }
  } catch (error) {
    ElMessage.error('操作失败')
  }
}

onMounted(() => {
  loadUserList()
  loadOrganizationList()
})
</script>

<style scoped>
.user-management {
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
