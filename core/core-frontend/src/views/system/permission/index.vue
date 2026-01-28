<template>
  <div class="perm-management">
    <div class="header">
      <h2>权限管理</h2>
      <el-button type="primary" @click="handleCreate">新建权限</el-button>
    </div>

    <el-table
      :data="permList"
      border
      row-key="permId"
      :tree-props="{ children: 'children', hasChildren: 'hasChildren' }"
    >
      <el-table-column prop="permName" label="权限名称" width="200" />
      <el-table-column prop="permKey" label="权限标识" width="200" />
      <el-table-column prop="permType" label="权限类型" width="120">
        <template #default="{ row }">
          <el-tag size="small">{{ getPermTypeName(row.permType) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="permDesc" label="权限描述" />
      <el-table-column prop="parentId" label="父权限" width="120">
        <template #default="{ row }">
          {{ getParentName(row.parentId) }}
        </template>
      </el-table-column>
      <el-table-column prop="status" label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'danger'">
            {{ row.status === 1 ? '启用' : '禁用' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="handleEdit(row)">编辑</el-button>
          <el-button link type="danger" @click="handleDelete(row.permId)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="500px">
      <el-form :model="form" :rules="rules" label-width="120px">
        <el-form-item label="权限名称" prop="permName">
          <el-input v-model="form.permName" placeholder="请输入权限名称" />
        </el-form-item>
        <el-form-item label="权限标识" prop="permKey">
          <el-input v-model="form.permKey" placeholder="请输入权限标识" />
        </el-form-item>
        <el-form-item label="权限类型" prop="permType">
          <el-select v-model="form.permType" placeholder="请选择权限类型">
            <el-option label="菜单权限" value="menu" />
            <el-option label="按钮权限" value="button" />
            <el-option label="数据权限" value="data" />
          </el-select>
        </el-form-item>
        <el-form-item label="权限描述" prop="permDesc">
          <el-input v-model="form.permDesc" type="textarea" placeholder="请输入权限描述" />
        </el-form-item>
        <el-form-item label="父权限" prop="parentId">
          <el-tree-select
            v-model="form.parentId"
            :data="permTreeData"
            :props="{ label: 'permName', value: 'permId' }"
            placeholder="请选择父权限"
            clearable
            check-strictly
          />
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
import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus-secondary'
import { permListApi, permCreateApi, permUpdateApi, permDeleteApi } from '@/api/org'
import { resourceTreeApi } from '@/api/auth'

const permList = ref([])
const permTreeData = computed(() => buildPermTree(permList.value))
const dialogVisible = ref(false)
const dialogTitle = ref('')
const form = ref({
  permId: null,
  permName: '',
  permKey: '',
  permType: 'menu',
  permDesc: '',
  parentId: null,
  status: 1
})

const rules = {
  permName: [{ required: true, message: '请输入权限名称', trigger: 'blur' }],
  permKey: [{ required: true, message: '请输入权限标识', trigger: 'blur' }],
  permType: [{ required: true, message: '请选择权限类型', trigger: 'change' }],
  permDesc: [{ required: true, message: '请输入权限描述', trigger: 'blur' }]
}

const getPermTypeName = (type: string) => {
  const typeMap: Record<string, string> = {
    menu: '菜单',
    button: '按钮',
    data: '数据'
  }
  return typeMap[type] || type
}

const buildPermTree = (flatList: any[]) => {
  const tree: any[] = []
  const map = new Map()

  flatList.forEach(perm => {
    const node = {
      ...perm,
      children: []
    }
    map.set(perm.permId, node)
  })

  flatList.forEach(perm => {
    if (perm.parentId && map.has(perm.parentId)) {
      const parent = map.get(perm.parentId)
      if (parent) {
        parent.children.push(perm)
      }
    } else if (!perm.parentId) {
      tree.push(perm)
    }
  })

  return tree
}

const getParentName = (parentId: number | null) => {
  if (!parentId) return '无'
  const parent = permList.value.find((perm: any) => perm.permId === parentId)
  return parent ? parent.permName : '未知'
}

const loadPermList = async () => {
  try {
    const res = await permListApi({ current: 1, size: 100 })
    if (res.code === '000000') {
      permList.value = res.data?.list || []
    }
  } catch (error) {
    ElMessage.error('加载权限列表失败')
  }
}

const handleCreate = () => {
  dialogTitle.value = '新建权限'
  form.value = {
    permId: null,
    permName: '',
    permKey: '',
    permType: 'menu',
    permDesc: '',
    parentId: null,
    status: 1
  }
  dialogVisible.value = true
}

const handleEdit = (row: any) => {
  dialogTitle.value = '编辑权限'
  form.value = { ...row }
  dialogVisible.value = true
}

const handleDelete = async (permId: number) => {
  try {
    await ElMessageBox.confirm('确定要删除该权限吗?', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })

    const res = await permDeleteApi(permId)
    if (res.code === '000000') {
      ElMessage.success('删除成功')
      loadPermList()
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
    if (form.value.permId) {
      res = await permUpdateApi(form.value)
    } else {
      res = await permCreateApi(form.value)
    }

    if (res.code === '000000') {
      ElMessage.success(form.value.permId ? '更新成功' : '创建成功')
      dialogVisible.value = false
      loadPermList()
    }
  } catch (error) {
    ElMessage.error('操作失败')
  }
}

onMounted(() => {
  loadPermList()
})
</script>

<style scoped>
.perm-management {
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
