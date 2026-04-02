<template>
  <div class="resource-permission">
    <div class="toolbar">
      <el-radio-group v-model="viewMode" size="default">
        <el-radio-button value="carrier">按角色</el-radio-button>
        <el-radio-button value="user">按用户</el-radio-button>
        <el-radio-button value="resource">按资源</el-radio-button>
      </el-radio-group>

      <template v-if="viewMode === 'carrier'">
        <el-select v-model="selectedRoleId" placeholder="请选择角色" @change="handleRoleChange">
          <el-option
            v-for="role in roleList"
            :key="role.roleId"
            :label="role.roleName"
            :value="role.roleId"
          />
        </el-select>
      </template>

      <template v-else-if="viewMode === 'user'">
        <el-select v-model="selectedUserId" placeholder="请选择用户" @change="handleUserChange">
          <el-option
            v-for="user in userList"
            :key="user.userId"
            :label="user.nickName ? `${user.nickName} (${user.username})` : user.username"
            :value="user.userId"
          />
        </el-select>
        <el-select v-model="selectedResourceType" placeholder="请选择资源类型" @change="handleUserResourceTypeChange">
          <el-option
            v-for="item in resourceTypeOptions"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </el-select>
      </template>

      <template v-else>
        <el-select v-model="selectedResourceType" placeholder="请选择资源类型" @change="handleResourceTypeChange">
          <el-option
            v-for="item in resourceTypeOptions"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </el-select>
        <el-select
          v-model="selectedResourceId"
          placeholder="请选择资源"
          :disabled="resourceOptions.length === 0"
          @change="handleResourceSelectionChange"
        >
          <el-option
            v-for="item in resourceOptions"
            :key="item.id"
            :label="item.name"
            :value="item.id"
          />
        </el-select>
        <el-select v-model="selectedUserId" clearable placeholder="按用户校验（可选）">
          <el-option
            v-for="user in userList"
            :key="user.userId"
            :label="user.nickName ? `${user.nickName} (${user.username})` : user.username"
            :value="user.userId"
          />
        </el-select>
      </template>
    </div>

    <div class="content" v-if="activeSelectionReady">
      <template v-if="viewMode === 'carrier'">
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
      </template>

      <template v-else-if="viewMode === 'user'">
        <div class="resource-target-view-desc">
          按用户视角展示当前资源类型下该用户的有效授权结果，用于与按资源视角校验同一底层授权状态。当前切片只读展示，不在此处新增用户侧保存语义。
        </div>
        <el-table :data="userPerspectivePermissions" border v-loading="userPerspectiveLoading">
          <el-table-column prop="permName" label="权限" min-width="160" />
          <el-table-column prop="sourceType" label="来源类型" min-width="120" />
          <el-table-column prop="sourceName" label="来源名称" min-width="180" show-overflow-tooltip />
        </el-table>
      </template>

      <template v-else>
        <div class="resource-target-view-desc">
          资源视角展示当前资源类型下的有效授权结果。当前切片仅允许编辑角色来源的权限集合；直接授权用户继续只读展示，不在此处伪装为资源实例级独立授权。
        </div>
        <div class="resource-target-view-desc" v-if="selectedUserDisplayName">
          当前已按用户过滤有效授权明细：{{ selectedUserDisplayName }}。
        </div>
        <div class="resource-target-editor" v-if="resourceRoleTargets.length > 0">
          <div class="resource-target-editor-title">角色来源编辑</div>
          <el-table :data="resourceRoleTargets" border>
            <el-table-column prop="targetName" label="角色" min-width="160" />
            <el-table-column label="影响用户" min-width="220">
              <template #default="scope">
                {{ scope.row.userNames.length > 0 ? scope.row.userNames.join('、') : '—' }}
              </template>
            </el-table-column>
            <el-table-column label="权限集合" min-width="320">
              <template #default="scope">
                <el-select
                  v-model="scope.row.permIds"
                  multiple
                  collapse-tags
                  collapse-tags-tooltip
                  style="width: 100%"
                >
                  <el-option
                    v-for="item in permissionCatalog"
                    :key="item.permId"
                    :label="item.permName"
                    :value="item.permId"
                  />
                </el-select>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="120">
              <template #default="scope">
                <el-button
                  type="primary"
                  link
                  :loading="scope.row.saving"
                  @click="handleResourceRoleSave(scope.row)"
                >
                  保存
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </div>
        <el-empty
          v-else
          description="当前资源视角下暂无可编辑的角色来源授权；如需新增载体，请先在按角色视角配置。"
        />
        <div class="resource-target-editor-title">有效授权明细</div>
        <el-table :data="filteredResourceTargetPermissions" border v-loading="resourceTargetLoading">
          <el-table-column prop="username" label="账号" min-width="140" />
          <el-table-column prop="nickName" label="名称" min-width="140" />
          <el-table-column prop="permName" label="权限" min-width="120" />
          <el-table-column prop="sourceType" label="来源类型" min-width="120" />
          <el-table-column prop="sourceName" label="来源名称" min-width="160" show-overflow-tooltip />
        </el-table>
      </template>
    </div>

    <el-empty v-else :description="emptyDescription" />
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import { ElMessage } from 'element-plus-secondary'
import { queryTreeApi } from '@/api/visualization/dataVisualization'
import { getDsTree } from '@/api/datasource'
import { getDatasetTree } from '@/api/dataset'
import {
  busiTargetPerSaveApi,
  busiPerSaveApi,
  queryRoleApi,
  queryUserApi,
  resourcePerApi,
  resourceTargetPerApi,
  resourceTreeApi,
  userPerspectiveApi
} from '@/api/auth'

type RoleItem = {
  roleId: number
  roleName: string
}

type UserItem = {
  userId: number
  username: string
  nickName?: string
}

type PermissionItem = {
  permId: number
  permName: string
  permKey?: string
  parentId?: number | null
  children?: PermissionItem[]
}

type TreeCheckInfo = {
  checkedKeys: number[]
}

type PermissionTreeRef = {
  setCheckedKeys: (keys: number[]) => void
}

type ResourceOption = {
  id: number
  name: string
}

type ResourceTargetPermission = {
  userId: number
  username: string
  nickName: string
  permKey: string
  permName: string
  sourceType: string
  sourceId?: number
  sourceName: string
}

type FlatPermissionOption = {
  permId: number
  permKey: string
  permName: string
}

type ResourceRoleTarget = {
  targetId: number
  targetName: string
  userNames: string[]
  permIds: number[]
  saving: boolean
}

type UserPerspectivePermission = {
  permKey: string
  permName: string
  sourceType: string
  sourceId?: number
  sourceName?: string
  resourceType?: string
}

const roleList = ref<RoleItem[]>([])
const userList = ref<UserItem[]>([])
const viewMode = ref<'carrier' | 'user' | 'resource'>('carrier')
const selectedRoleId = ref<number | null>(null)
const selectedUserId = ref<number | null>(null)
const permissionTree = ref<PermissionItem[]>([])
const selectedPermIds = ref<number[]>([])
const permissionTreeRef = ref<PermissionTreeRef>()
const selectedResourceType = ref<'datasource' | 'dataset' | 'dashboard' | 'screen'>('datasource')
const selectedResourceId = ref<number | null>(null)
const resourceOptions = ref<ResourceOption[]>([])
const resourceTargetPermissions = ref<ResourceTargetPermission[]>([])
const resourceTargetLoading = ref(false)
const resourceRoleTargets = ref<ResourceRoleTarget[]>([])
const userPerspectivePermissions = ref<UserPerspectivePermission[]>([])
const userPerspectiveLoading = ref(false)

const resourceTypeOptions = [
  { label: '数据源', value: 'datasource' },
  { label: '数据集', value: 'dataset' },
  { label: '仪表板', value: 'dashboard' },
  { label: '数据大屏', value: 'screen' }
] as const

const syncCheckedPermissions = async () => {
  await nextTick()
  permissionTreeRef.value?.setCheckedKeys(selectedPermIds.value)
}

const flattenPermissionTree = (nodes: PermissionItem[]): FlatPermissionOption[] => {
  const result: FlatPermissionOption[] = []

  const walk = (items: PermissionItem[]) => {
    items.forEach(item => {
      result.push({
        permId: item.permId,
        permKey: item.permKey || '',
        permName: item.permName
      })

      if (Array.isArray(item.children) && item.children.length > 0) {
        walk(item.children)
      }
    })
  }

  walk(nodes)
  return result
}

const permissionCatalog = computed(() => {
  return flattenPermissionTree(permissionTree.value).filter(item =>
    item.permKey.startsWith(`${selectedResourceType.value}:`)
  )
})

const permissionIdByKey = computed(() => {
  return new Map(permissionCatalog.value.map(item => [item.permKey, item.permId]))
})

const selectedUser = computed(() => {
  return userList.value.find(item => item.userId === selectedUserId.value) || null
})

const selectedUserDisplayName = computed(() => {
  if (!selectedUser.value) {
    return ''
  }

  return selectedUser.value.nickName
    ? `${selectedUser.value.nickName} (${selectedUser.value.username})`
    : selectedUser.value.username
})

const filteredResourceTargetPermissions = computed(() => {
  if (!selectedUserId.value) {
    return resourceTargetPermissions.value
  }

  return resourceTargetPermissions.value.filter(item => item.userId === selectedUserId.value)
})

const activeSelectionReady = computed(() => {
  if (viewMode.value === 'carrier') {
    return !!selectedRoleId.value
  }
  if (viewMode.value === 'user') {
    return !!selectedUserId.value
  }
  return !!selectedResourceId.value
})

const emptyDescription = computed(() => {
  if (viewMode.value === 'carrier') {
    return '请选择角色'
  }
  if (viewMode.value === 'user') {
    return '请选择用户'
  }
  return '请选择资源'
})

const syncResourceRoleTargets = () => {
  const grouped = new Map<number, ResourceRoleTarget>()

  resourceTargetPermissions.value
    .filter(item => item.sourceType === 'role' && Number(item.sourceId || 0) > 0)
    .forEach(item => {
      const targetId = Number(item.sourceId || 0)
      const existing =
        grouped.get(targetId) || {
          targetId,
          targetName: item.sourceName,
          userNames: [],
          permIds: [],
          saving: false
        }

      if (item.username && !existing.userNames.includes(item.username)) {
        existing.userNames.push(item.username)
      }

      const permId = permissionIdByKey.value.get(item.permKey)
      if (permId && !existing.permIds.includes(permId)) {
        existing.permIds.push(permId)
      }

      grouped.set(targetId, existing)
    })

  resourceRoleTargets.value = Array.from(grouped.values()).sort((a, b) => a.targetId - b.targetId)
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

const normalizeUsers = (items: any[]): UserItem[] => {
  return items.map(item => ({
    userId: Number(item.userId || item.id),
    username: item.username || '',
    nickName: item.nickName || item.realName || ''
  }))
}

const loadUserList = async () => {
  try {
    const res = await queryUserApi({ current: 1, size: 1000 })
    if (res.code === '000000') {
      userList.value = normalizeUsers(res.data?.list || [])
    }
  } catch (_error) {
    userList.value = []
    ElMessage.error('加载用户列表失败')
  }
}

const loadPermissionTree = async () => {
  try {
    const res = await resourceTreeApi('all')
    if (res.code === '000000') {
      permissionTree.value = buildPermissionTree(res.data || [])
      await syncCheckedPermissions()
      syncResourceRoleTargets()
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
      if (selectedResourceId.value) {
        await loadResourceTargetPermissions()
      }
      window.dispatchEvent(new Event('de:permission-refresh'))
    }
  } catch (error) {
    ElMessage.error('资源权限保存失败')
  }
}

const normalizeResourceNodes = (nodes: any[] = []): ResourceOption[] => {
  const result: ResourceOption[] = []

  const walk = (items: any[]) => {
    items.forEach(item => {
      const children = Array.isArray(item?.children) ? item.children : []
      const id = Number(item?.id ?? 0)
      const isLeaf = item?.leaf === true || children.length === 0
      if (isLeaf && id > 0) {
        result.push({
          id,
          name: item?.name || item?.label || String(id)
        })
        return
      }

      if (children.length > 0) {
        walk(children)
      }
    })
  }

  walk(nodes)
  return result
}

const loadResourceOptions = async () => {
  selectedResourceId.value = null
  resourceTargetPermissions.value = []
  resourceRoleTargets.value = []
  try {
    if (selectedResourceType.value === 'datasource') {
      const res = await getDsTree({ leaf: false, weight: 0 })
      resourceOptions.value = normalizeResourceNodes(res as unknown as any[])
      return
    }

    if (selectedResourceType.value === 'dataset') {
      const res = await getDatasetTree({ leaf: false, weight: 0 } as any)
      resourceOptions.value = normalizeResourceNodes(res as unknown as any[])
      return
    }

    const res = await queryTreeApi({
      busiFlag: selectedResourceType.value === 'screen' ? 'dataV' : 'dashboard',
      resourceTable: 'core',
      leaf: false
    } as any)
    resourceOptions.value = normalizeResourceNodes((res as any) || [])
  } catch (_error) {
    resourceOptions.value = []
    ElMessage.error('加载资源列表失败')
  }
}

const loadResourceTargetPermissions = async () => {
  if (!selectedResourceId.value) {
    resourceTargetPermissions.value = []
    resourceRoleTargets.value = []
    return
  }

  resourceTargetLoading.value = true
  try {
    const res = await resourceTargetPerApi({
      id: selectedResourceId.value,
      type: 0,
      flag: selectedResourceType.value
    })
    if (res.code === '000000') {
      resourceTargetPermissions.value = Array.isArray(res.data) ? res.data : []
      syncResourceRoleTargets()
    }
  } catch (_error) {
    resourceTargetPermissions.value = []
    resourceRoleTargets.value = []
    ElMessage.error('加载资源授权对象失败')
  } finally {
    resourceTargetLoading.value = false
  }
}

const loadUserPerspectivePermissions = async () => {
  if (!selectedUserId.value) {
    userPerspectivePermissions.value = []
    return
  }

  userPerspectiveLoading.value = true
  try {
    const res = await userPerspectiveApi({
      userId: selectedUserId.value,
	      resourceId: selectedResourceId.value || undefined,
	      resourceType: selectedResourceType.value
    })
    if (res.code === '000000') {
      userPerspectivePermissions.value = Array.isArray(res.data) ? res.data : []
    }
  } catch (_error) {
    userPerspectivePermissions.value = []
    ElMessage.error('加载用户视角资源权限失败')
  } finally {
    userPerspectiveLoading.value = false
  }
}

const handleResourceRoleSave = async (row: ResourceRoleTarget) => {
  if (!selectedResourceId.value) {
    ElMessage.warning('请选择资源')
    return
  }

  row.saving = true
  try {
    const res = await busiTargetPerSaveApi({
      id: selectedResourceId.value,
      type: 0,
      flag: selectedResourceType.value,
      targetPerms: resourceRoleTargets.value.map(item => ({
        targetType: 'role',
        targetId: item.targetId,
        permIds: item.permIds
      }))
    })

    if (res.code === '000000') {
      ElMessage.success('资源视角角色权限保存成功')
      await loadResourceTargetPermissions()
      if (selectedRoleId.value === row.targetId) {
        await handleRoleChange()
      }
      window.dispatchEvent(new Event('de:permission-refresh'))
    }
  } catch (_error) {
    ElMessage.error('资源视角角色权限保存失败')
  } finally {
    row.saving = false
  }
}

const handleResourceTypeChange = () => {
  loadResourceOptions()
}

const handleUserChange = () => {
  selectedResourceId.value = null
  loadUserPerspectivePermissions()
}

const handleUserResourceTypeChange = () => {
  selectedResourceId.value = null
  loadUserPerspectivePermissions()
}

const handleResourceSelectionChange = () => {
  loadResourceTargetPermissions()
}

watch(viewMode, mode => {
  if (mode === 'resource') {
    loadResourceOptions()
    return
  }

  if (mode === 'user' && selectedUserId.value) {
    selectedResourceId.value = null
    loadUserPerspectivePermissions()
  }
})

watch(selectedResourceType, () => {
  syncResourceRoleTargets()
})

onMounted(() => {
  loadRoleList()
  loadUserList()
  loadPermissionTree()
  loadResourceOptions()
})
</script>

<style scoped>
.resource-permission {
  padding: 0;
}

.toolbar {
  display: flex;
  margin-bottom: 16px;
  gap: 12px;
  flex-wrap: wrap;
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

.resource-target-view-desc {
  margin-bottom: 12px;
  font-size: 13px;
  line-height: 20px;
  color: var(--ed-text-color-secondary);
}

.resource-target-editor {
  margin-bottom: 16px;
}

.resource-target-editor-title {
  margin: 12px 0;
  font-size: 13px;
  font-weight: 500;
  color: var(--ed-text-color-primary);
}
</style>
