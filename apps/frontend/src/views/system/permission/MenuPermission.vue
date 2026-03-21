<template>
  <div class="menu-permission">
    <div class="toolbar">
      <el-select v-model="selectedRoleId" placeholder="请选择角色" @change="handleRoleChange">
        <el-option
          v-for="role in roleList"
          :key="role.roleId"
          :label="role.roleName"
          :value="role.roleId"
        />
      </el-select>
    </div>

    <div class="content" v-if="selectedRoleId">
      <el-tree
        ref="menuTreeRef"
        :data="menuTree"
        :props="{ label: 'meta title', children: 'children' }"
        show-checkbox
        node-key="id"
        :default-checked-keys="selectedMenuIds"
        :default-expand-all="true"
        @check="handleMenuCheck"
      />
      <div class="footer">
        <el-button type="primary" @click="handleSave">保存</el-button>
      </div>
    </div>

    <el-empty v-else description="请选择角色" />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus-secondary'
import { queryRoleApi, menuTreeApi, roleMenuAuthApi, roleMenuAuthSaveApi } from '@/api/auth'

const roleList = ref([])
const selectedRoleId = ref(null)
const menuTree = ref([])
const selectedMenuIds = ref([])
const menuTreeRef = ref()

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

const handleRoleChange = async () => {
  if (!selectedRoleId.value) {
    selectedMenuIds.value = []
    return
  }

  try {
    const res = await roleMenuAuthApi(selectedRoleId.value)
    if (res.code === '000000') {
      selectedMenuIds.value = res.data?.menuIds || []
    }
  } catch (error) {
    ElMessage.error('加载菜单授权失败')
  }
}

const handleMenuCheck = (checkedKeys: any[]) => {
  selectedMenuIds.value = checkedKeys
}

const handleSave = async () => {
  if (!selectedRoleId.value) {
    ElMessage.warning('请选择角色')
    return
  }

  try {
    const res = await roleMenuAuthSaveApi({
      roleId: selectedRoleId.value,
      menuIds: selectedMenuIds.value
    })
    if (res.code === '000000') {
      ElMessage.success('菜单授权成功')
      window.dispatchEvent(new Event('de:permission-refresh'))
    }
  } catch (error) {
    ElMessage.error('菜单授权失败')
  }
}

onMounted(() => {
  loadRoleList()
  loadMenuTree()
})
</script>

<style scoped>
.menu-permission {
  padding: 0;
}

.toolbar {
  margin-bottom: 16px;
}

.content {
  padding: 16px;
  border: 1px solid #eee;
  border-radius: 4px;
}

.footer {
  margin-top: 16px;
  text-align: right;
}
</style>
