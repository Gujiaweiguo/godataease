<template>
  <div class="org-management">
    <div class="header">
      <h2>组织管理</h2>
      <el-button type="primary" @click="handleCreate">新建组织</el-button>
    </div>

    <el-table
      :data="orgList"
      border
      row-key="orgId"
      :tree-props="{ children: 'children', hasChildren: 'hasChildren' }"
    >
      <el-table-column prop="orgName" label="组织名称" width="250" />
      <el-table-column prop="orgDesc" label="描述" />
      <el-table-column prop="parentId" label="父组织" width="150">
        <template #default="{ row }">
          {{ getParentName(row.parentId) }}
        </template>
      </el-table-column>
      <el-table-column prop="level" label="层级" width="100">
        <template #default="{ row }">
          {{ '第' + row.level + '级' }}
        </template>
      </el-table-column>
      <el-table-column prop="status" label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'danger'">
            {{ row.status === 1 ? '启用' : '禁用' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="300" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="handleEdit(row)">编辑</el-button>
          <el-button link type="primary" @click="handleAddChild(row)">添加子组织</el-button>
          <el-button link type="danger" @click="handleDelete(row.orgId)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="500px">
      <el-form :model="form" :rules="rules" label-width="100px">
        <el-form-item label="组织名称" prop="orgName">
          <el-input v-model="form.orgName" placeholder="请输入组织名称" />
        </el-form-item>
        <el-form-item label="组织描述" prop="orgDesc">
          <el-input v-model="form.orgDesc" type="textarea" placeholder="请输入组织描述" />
        </el-form-item>
        <el-form-item label="父组织" prop="parentId">
          <el-tree-select
            v-model="form.parentId"
            :data="orgTreeData"
            :props="{ label: 'orgName', value: 'orgId' }"
            placeholder="请选择父组织"
            clearable
            check-strictly
          />
        </el-form-item>
        <el-form-item label="层级" prop="level">
          <el-input-number v-model="form.level" :min="1" :max="10" placeholder="请输入层级" />
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
import { orgCreateApi, orgUpdateApi, orgDeleteApi, orgListApi } from '@/api/org'

const orgList = ref([])
const orgTreeData = computed(() => buildTreeData(orgList.value))
const dialogVisible = ref(false)
const dialogTitle = ref('')
const form = ref({
  orgId: null,
  orgName: '',
  orgDesc: '',
  parentId: null,
  level: 1,
  status: 1
})

const rules = {
  orgName: [{ required: true, message: '请输入组织名称', trigger: 'blur' }],
  parentId: [{ required: true, message: '请选择父组织', trigger: 'change' }],
  level: [{ required: true, message: '请输入层级', trigger: 'blur' }]
}

const buildTreeData = (flatList: any[]) => {
  const tree: any[] = []
  const map = new Map()

  flatList.forEach(org => {
    const node = {
      ...org,
      children: []
    }
    map.set(org.orgId, node)
  })

  flatList.forEach(org => {
    if (org.parentId && map.has(org.parentId)) {
      const parent = map.get(org.parentId)
      if (parent) {
        parent.children.push(org)
      }
    } else if (!org.parentId) {
      tree.push(org)
    }
  })

  return tree
}

const getParentName = (parentId: number | null) => {
  if (!parentId) return '无'
  const parent = orgList.value.find((org: any) => org.orgId === parentId)
  return parent ? parent.orgName : '未知'
}

const loadOrgList = async () => {
  try {
    const res = await orgListApi({ current: 1, size: 100 })
    if (res.code === '000000') {
      orgList.value = res.data?.list || []
    }
  } catch (error) {
    ElMessage.error('加载组织列表失败')
  }
}

const handleCreate = () => {
  dialogTitle.value = '新建组织'
  form.value = {
    orgId: null,
    orgName: '',
    orgDesc: '',
    parentId: null,
    level: 1,
    status: 1
  }
  dialogVisible.value = true
}

const handleEdit = (row: any) => {
  dialogTitle.value = '编辑组织'
  form.value = { ...row }
  dialogVisible.value = true
}

const handleAddChild = (row: any) => {
  dialogTitle.value = '添加子组织'
  form.value = {
    orgId: null,
    orgName: '',
    orgDesc: '',
    parentId: row.orgId,
    level: row.level + 1,
    status: 1
  }
  dialogVisible.value = true
}

const handleDelete = async (orgId: number) => {
  try {
    await ElMessageBox.confirm('确定要删除该组织吗?删除后所有子组织也将被删除。', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })

    const res = await orgDeleteApi(orgId)
    if (res.code === '000000') {
      ElMessage.success('删除成功')
      loadOrgList()
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
    if (form.value.orgId) {
      res = await orgUpdateApi(form.value)
    } else {
      res = await orgCreateApi(form.value)
    }

    if (res.code === '000000') {
      ElMessage.success(form.value.orgId ? '更新成功' : '创建成功')
      dialogVisible.value = false
      loadOrgList()
    }
  } catch (error) {
    ElMessage.error('操作失败')
  }
}

onMounted(() => {
  loadOrgList()
})
</script>

<style scoped>
.org-management {
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
