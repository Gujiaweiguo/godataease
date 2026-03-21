<template>
  <div class="resource-permission">
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
      <el-empty v-if="permissionTree.length === 0" description="暂无资源权限数据" />
      <template v-else>
        <el-tree
          ref="permissionTreeRef"
          :data="permissionTree"
          :props="{ label: 'permName', children: 'children' }"
          show-checkbox
          node-key="permId"
          :default-checked-keys="selectedPermIds"
          :default-expand-all="true"
          @check="handlePermissionCheck"
        />
        <div class="footer">
          <el-button type="primary" @click="handleSave">保存</el-button>
        </div>
      </template>
    </div>

    <el-empty v-else description="请选择角色" />
  </div>
</template>

<script setup lang="ts">
import { nextTick, onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus-secondary'
import { busiPerSaveApi, queryRoleApi, resourcePerApi, resourceTreeApi } from '@/api/auth'

type RoleItem = {
  roleId: number
  roleName: string
}

type PermissionItem = {
  permId: number
  permName: string
  parentId?: number | null
  children?: PermissionItem[]
}

type TreeCheckInfo = {
  checkedKeys: number[]
}

type PermissionTreeRef = {
  setCheckedKeys: (keys: number[]) => void
}

const roleList = ref<RoleItem[]>([])
const selectedRoleId = ref<number | null>(null)
const permissionTree = ref<PermissionItem[]>([])
const selectedPermIds = ref<number[]>([])
const permissionTreeRef = ref<PermissionTreeRef>()

const syncCheckedPermissions = async () => {
  await nextTick()
  permissionTreeRef.value?.setCheckedKeys(selectedPermIds.value)
}

const buildPermissionTree = (permissions: PermissionItem[]) => {
  const tree: PermissionItem[] = []
  const map = new Map<number, PermissionItem>()

  permissions.forEach(permission => {
    map.set(permission.permId, {
      ...permission,
      children: []
    })
  })

  map.forEach(permission => {
    if (permission.parentId && map.has(permission.parentId)) {
      map.get(permission.parentId)?.children?.push(permission)
    } else {
      tree.push(permission)
    }
  })

  return tree
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

const loadPermissionTree = async () => {
  try {
    const res = await resourceTreeApi('1')
    if (res.code === '000000') {
      permissionTree.value = buildPermissionTree(res.data || [])
      await syncCheckedPermissions()
    }
  } catch (error) {
    ElMessage.error('加载资源权限失败')
  }
}

const handleRoleChange = async () => {
  if (!selectedRoleId.value) {
    selectedPermIds.value = []
    await syncCheckedPermissions()
    return
  }

  try {
    const res = await resourcePerApi({ roleId: selectedRoleId.value })
    if (res.code === '000000') {
      selectedPermIds.value = res.data?.permIds || []
      await syncCheckedPermissions()
    }
  } catch (error) {
    ElMessage.error('加载资源权限失败')
  }
}

const handlePermissionCheck = (_: PermissionItem, checkedInfo: TreeCheckInfo) => {
  selectedPermIds.value = checkedInfo.checkedKeys
}

const handleSave = async () => {
  if (!selectedRoleId.value) {
    ElMessage.warning('请选择角色')
    return
  }

  try {
    const res = await busiPerSaveApi({
      roleId: selectedRoleId.value,
      permIds: selectedPermIds.value
    })
    if (res.code === '000000') {
      ElMessage.success('资源权限保存成功')
      window.dispatchEvent(new Event('de:permission-refresh'))
    }
  } catch (error) {
    ElMessage.error('资源权限保存失败')
  }
}

onMounted(() => {
  loadRoleList()
  loadPermissionTree()
})
</script>

<style scoped>
.resource-permission {
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
