<template>
  <div class="role-tab">
    <div class="toolbar">
      <div class="toolbar-content">
        <div class="toolbar-title">角色工作流</div>
        <div class="toolbar-desc">
          继承当前用户管理组织上下文，仅处理角色生命周期与成员维护。
        </div>
      </div>
      <el-button type="primary" @click="handleCreate">新建角色</el-button>
    </div>

    <el-table :data="roleList" border v-loading="loading">
      <el-table-column prop="roleName" label="角色名称" min-width="180" />
      <el-table-column prop="roleKey" label="角色标识" min-width="180" />
      <el-table-column label="角色类型" width="120">
        <template #default="{ row }">
          <el-tag size="small" effect="plain">{{ formatRoleType(row.roleType) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="roleDesc" label="角色描述" min-width="220" show-overflow-tooltip />
      <el-table-column prop="status" label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'danger'">
            {{ row.status === 1 ? '启用' : '禁用' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="300" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="handleManageMembers(row)">成员管理</el-button>
          <el-button link type="primary" @click="handleEdit(row)">编辑</el-button>
          <el-button
            link
            type="danger"
            :disabled="Boolean(row.readonly)"
            @click="handleDelete(row.roleId)"
          >
            删除
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="500px">
      <el-form ref="roleFormRef" :model="form" :rules="rules" label-width="100px">
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

    <el-dialog
      v-model="memberDialogVisible"
      destroy-on-close
      width="1180px"
      :title="currentRole ? `${currentRole.roleName} · 成员管理` : '成员管理'"
    >
      <div class="member-layout" v-loading="memberDialogLoading">
        <div class="member-layout-panel member-layout-panel--stacked">
          <div class="panel-card panel-card-summary">
            <div class="panel-card-title">当前角色</div>
            <div class="role-summary">
              <div class="role-summary-name">{{ currentRole?.roleName || '--' }}</div>
              <div class="role-summary-meta">
                <span>标识：{{ currentRole?.roleKey || '--' }}</span>
                <span>状态：{{ currentRole?.status === 1 ? '启用' : '禁用' }}</span>
              </div>
              <div class="role-summary-desc">{{ currentRole?.roleDesc || '未填写角色描述' }}</div>
            </div>
          </div>

          <div class="panel-card">
            <div class="panel-card-header">
              <div>
                <div class="panel-card-title">添加组织内用户</div>
                <div class="panel-card-subtitle">仅展示当前组织上下文下尚未绑定该角色的用户。</div>
              </div>
            </div>
            <div class="panel-card-toolbar">
              <el-input
                v-model="orgCandidateKeyword"
                clearable
                placeholder="搜索当前组织用户"
                @keyup.enter="loadOrgCandidates"
              />
              <el-button @click="loadOrgCandidates">查询</el-button>
            </div>
            <el-table
              :data="orgCandidateList"
              border
              height="260"
              v-loading="orgCandidateLoading"
              @selection-change="handleOrgSelectionChange"
            >
              <el-table-column type="selection" width="48" />
              <el-table-column prop="account" label="账号" min-width="140" />
              <el-table-column prop="name" label="名称" min-width="120" />
              <el-table-column prop="email" label="邮箱" min-width="180" show-overflow-tooltip />
            </el-table>
            <div class="panel-card-footer">
              <span class="panel-card-hint">已选 {{ selectedOrgUserIds.length }} 人</span>
              <el-button
                type="primary"
                :disabled="selectedOrgUserIds.length === 0"
                :loading="mountOrgLoading"
                @click="handleMountOrgUsers"
              >
                添加所选用户
              </el-button>
            </div>
          </div>

          <div class="panel-card">
            <div class="panel-card-header">
              <div>
                <div class="panel-card-title">添加外部用户</div>
                <div class="panel-card-subtitle">需按账号或唯一标识精确检索，再执行挂载。</div>
              </div>
            </div>
            <div class="panel-card-toolbar">
              <el-input
                v-model="externalKeyword"
                clearable
                placeholder="请输入精确账号或 ID"
                @keyup.enter="handleExternalSearch"
              />
              <el-button type="primary" @click="handleExternalSearch">搜索</el-button>
            </div>
            <el-table :data="externalUserList" border height="220" v-loading="externalSearchLoading">
              <el-table-column prop="account" label="账号" min-width="140" />
              <el-table-column prop="name" label="名称" min-width="120" />
              <el-table-column prop="email" label="邮箱" min-width="180" show-overflow-tooltip />
              <el-table-column label="操作" width="100" align="right">
                <template #default="{ row }">
                  <el-button
                    link
                    type="primary"
                    :loading="mountExternalLoading && externalMountingUserId === row.uid"
                    @click="handleMountExternalUser(row)"
                  >
                    添加
                  </el-button>
                </template>
              </el-table-column>
            </el-table>
          </div>
        </div>

        <div class="member-layout-panel member-layout-panel--wide">
          <div class="panel-card panel-card-fill">
            <div class="panel-card-header">
              <div>
                <div class="panel-card-title">角色成员</div>
                <div class="panel-card-subtitle">支持搜索当前角色成员并执行解绑前校验。</div>
              </div>
              <el-tag type="info" effect="plain">共 {{ memberTotal }} 人</el-tag>
            </div>
            <div class="panel-card-toolbar">
              <el-input
                v-model="memberKeyword"
                clearable
                placeholder="搜索角色成员"
                @keyup.enter="handleMemberSearch"
              />
              <el-button @click="handleMemberSearch">查询</el-button>
            </div>
            <el-table :data="memberList" border height="560" v-loading="memberLoading">
              <el-table-column prop="account" label="账号" min-width="150" />
              <el-table-column prop="name" label="名称" min-width="140" />
              <el-table-column prop="email" label="邮箱" min-width="220" show-overflow-tooltip />
              <el-table-column prop="phone" label="手机号" min-width="150" />
              <el-table-column label="操作" width="100" align="right">
                <template #default="{ row }">
                  <el-button
                    link
                    type="danger"
                    :loading="unmountLoading && unmountingUserId === row.uid"
                    @click="handleUnmountUser(row)"
                  >
                    移除
                  </el-button>
                </template>
              </el-table-column>
            </el-table>
            <div class="panel-card-footer panel-card-footer-pagination">
              <span class="panel-card-hint">移除前会自动执行最后角色保护预检查。</span>
              <el-pagination
                background
                layout="prev, pager, next"
                :current-page="memberPage"
                :page-size="memberPageSize"
                :total="memberTotal"
                @current-change="handleMemberPageChange"
              />
            </div>
          </div>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { nextTick, onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus-secondary'
import { queryRoleApi, roleCreateApi, roleDeleteApi, roleUpdateApi } from '@/api/auth'
import {
  beforeUnmountInfoApi,
  mountExternalUserApi,
  mountUserApi,
  searchExternalUserApi,
  userOptionForRoleApi,
  userSelectedForRoleApi,
  unMountUserApi
} from '@/api/user'

type RoleItem = {
  roleId: number
  roleName: string
  roleKey: string
  roleDesc: string
  roleType?: string
  status: number
  readonly?: boolean
}

type UserItem = {
  id: number
  uid: number
  account: string
  name: string
  email: string
  phone: string
}

type RoleForm = {
  roleId: number | null
  roleName: string
  roleKey: string
  roleDesc: string
  status: number
}

const createDefaultForm = (): RoleForm => ({
  roleId: null,
  roleName: '',
  roleKey: '',
  roleDesc: '',
  status: 1
})

const loading = ref(false)
const roleList = ref<RoleItem[]>([])
const dialogVisible = ref(false)
const dialogTitle = ref('')
const roleFormRef = ref()
const form = ref<RoleForm>(createDefaultForm())

const memberDialogVisible = ref(false)
const memberDialogLoading = ref(false)
const memberLoading = ref(false)
const orgCandidateLoading = ref(false)
const externalSearchLoading = ref(false)
const mountOrgLoading = ref(false)
const mountExternalLoading = ref(false)
const unmountLoading = ref(false)
const currentRole = ref<RoleItem | null>(null)
const orgCandidateKeyword = ref('')
const memberKeyword = ref('')
const externalKeyword = ref('')
const orgCandidateList = ref<UserItem[]>([])
const externalUserList = ref<UserItem[]>([])
const memberList = ref<UserItem[]>([])
const selectedOrgUserIds = ref<number[]>([])
const memberPage = ref(1)
const memberPageSize = 10
const memberTotal = ref(0)
const externalMountingUserId = ref<number | null>(null)
const unmountingUserId = ref<number | null>(null)

const rules = {
  roleName: [{ required: true, message: '请输入角色名称', trigger: 'blur' }],
  roleKey: [{ required: true, message: '请输入角色标识', trigger: 'blur' }],
  roleDesc: [{ required: true, message: '请输入角色描述', trigger: 'blur' }]
}

const extractList = (data: any): any[] => {
  if (Array.isArray(data)) {
    return data
  }
  if (Array.isArray(data?.records)) {
    return data.records
  }
  if (Array.isArray(data?.list)) {
    return data.list
  }
  if (Array.isArray(data?.items)) {
    return data.items
  }
  return []
}

const normalizeRole = (item: any): RoleItem => ({
  ...item,
  roleId: Number(item?.roleId ?? item?.id ?? 0),
  roleName: item?.roleName ?? item?.name ?? '',
  roleKey: item?.roleKey ?? item?.code ?? '',
  roleDesc: item?.roleDesc ?? item?.desc ?? '',
  roleType: item?.roleType,
  status: Number(item?.status ?? 1),
  readonly: Boolean(item?.readonly)
})

const normalizeUser = (item: any): UserItem => {
  const rawId = Number(item?.uid ?? item?.id ?? item?.userId ?? 0)
  return {
    ...item,
    id: rawId,
    uid: rawId,
    account: item?.account ?? item?.username ?? item?.userName ?? '',
    name: item?.name ?? item?.realName ?? item?.nickName ?? '',
    email: item?.email ?? '',
    phone: item?.phone ?? ''
  }
}

const formatRoleType = (roleType?: string) => {
  if (roleType === 'system') return '系统'
  if (roleType === 'organization') return '组织'
  if (roleType === 'custom') return '自定义'
  return '默认'
}

const getCurrentRoleId = () => currentRole.value?.roleId || 0

const loadRoleList = async () => {
  loading.value = true
  try {
    const res = await queryRoleApi({ current: 1, size: 100 })
    if (res.code === '000000') {
      roleList.value = extractList(res.data).map(normalizeRole)
    }
  } catch (_error) {
    ElMessage.error('加载角色列表失败')
  } finally {
    loading.value = false
  }
}

const loadOrgCandidates = async () => {
  const rid = getCurrentRoleId()
  if (!rid) return
  orgCandidateLoading.value = true
  try {
    const res = await userOptionForRoleApi({
      rid,
      keyword: orgCandidateKeyword.value.trim() || undefined
    })
    if (res.code === '000000') {
      orgCandidateList.value = extractList(res.data).map(normalizeUser)
    }
  } catch (_error) {
    ElMessage.error('加载可添加用户失败')
  } finally {
    orgCandidateLoading.value = false
  }
}

const loadRoleMembers = async () => {
  const rid = getCurrentRoleId()
  if (!rid) return
  memberLoading.value = true
  try {
    const res = await userSelectedForRoleApi(memberPage.value, memberPageSize, {
      rid,
      keyword: memberKeyword.value.trim() || undefined
    })
    if (res.code === '000000') {
      const list = extractList(res.data).map(normalizeUser)
      memberList.value = list
      memberTotal.value = Number(res.data?.total ?? list.length)
    }
  } catch (_error) {
    ElMessage.error('加载角色成员失败')
  } finally {
    memberLoading.value = false
  }
}

const resetMemberState = () => {
  orgCandidateKeyword.value = ''
  memberKeyword.value = ''
  externalKeyword.value = ''
  orgCandidateList.value = []
  externalUserList.value = []
  memberList.value = []
  selectedOrgUserIds.value = []
  memberPage.value = 1
  memberTotal.value = 0
  externalMountingUserId.value = null
  unmountingUserId.value = null
}

const openRoleDialog = async (title: string, role?: RoleItem) => {
  dialogTitle.value = title
  form.value = role
    ? {
        roleId: role.roleId,
        roleName: role.roleName,
        roleKey: role.roleKey,
        roleDesc: role.roleDesc,
        status: role.status
      }
    : createDefaultForm()
  dialogVisible.value = true
  await nextTick()
  roleFormRef.value?.clearValidate?.()
}

const handleCreate = () => {
  openRoleDialog('新建角色')
}

const handleEdit = (row: RoleItem) => {
  openRoleDialog('编辑角色', normalizeRole(row))
}

const handleManageMembers = async (row: RoleItem) => {
  currentRole.value = normalizeRole(row)
  resetMemberState()
  memberDialogVisible.value = true
  memberDialogLoading.value = true
  try {
    await Promise.all([loadOrgCandidates(), loadRoleMembers()])
  } finally {
    memberDialogLoading.value = false
  }
}

const handleOrgSelectionChange = (rows: UserItem[]) => {
  selectedOrgUserIds.value = rows.map(item => item.uid)
}

const handleMountOrgUsers = async () => {
  const rid = getCurrentRoleId()
  if (!rid || selectedOrgUserIds.value.length === 0) return
  mountOrgLoading.value = true
  try {
    const res = await mountUserApi({ rid, uids: selectedOrgUserIds.value })
    if (res.code === '000000') {
      ElMessage.success('组织用户添加成功')
      selectedOrgUserIds.value = []
      await Promise.all([loadOrgCandidates(), loadRoleMembers()])
    }
  } catch (_error) {
    ElMessage.error('组织用户添加失败')
  } finally {
    mountOrgLoading.value = false
  }
}

const handleExternalSearch = async () => {
  const keyword = externalKeyword.value.trim()
  if (!keyword) {
    externalUserList.value = []
    ElMessage.warning('请输入精确账号或唯一标识')
    return
  }

  externalSearchLoading.value = true
  try {
    const res = await searchExternalUserApi(keyword)
    if (res.code === '000000') {
      externalUserList.value = extractList(res.data).map(normalizeUser)
    }
  } catch (_error) {
    ElMessage.error('外部用户搜索失败')
  } finally {
    externalSearchLoading.value = false
  }
}

const handleMountExternalUser = async (user: UserItem) => {
  const rid = getCurrentRoleId()
  if (!rid || !user.uid) return
  mountExternalLoading.value = true
  externalMountingUserId.value = user.uid
  try {
    const res = await mountExternalUserApi({ rid, uid: user.uid })
    if (res.code === '000000') {
      ElMessage.success('外部用户添加成功')
      await Promise.all([loadRoleMembers(), loadOrgCandidates(), handleExternalSearch()])
    }
  } catch (_error) {
    ElMessage.error('外部用户添加失败')
  } finally {
    mountExternalLoading.value = false
    externalMountingUserId.value = null
  }
}

const handleMemberSearch = () => {
  memberPage.value = 1
  loadRoleMembers()
}

const handleMemberPageChange = (page: number) => {
  memberPage.value = page
  loadRoleMembers()
}

const handleUnmountUser = async (user: UserItem) => {
  const rid = getCurrentRoleId()
  if (!rid || !user.uid) return

  try {
    const beforeRes = await beforeUnmountInfoApi({ rid, uid: user.uid })
    const roleCount = Number(beforeRes.data ?? 0)

    if (roleCount <= 1) {
      ElMessage.warning('该用户仅剩当前角色，无法移除')
      return
    }

    await ElMessageBox.confirm(
      `确认将 ${user.name || user.account} 从当前角色中移除吗？移除后该用户仍保留 ${roleCount - 1} 个角色。`,
      '移除成员',
      {
        confirmButtonText: '确定移除',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    unmountLoading.value = true
    unmountingUserId.value = user.uid
    const res = await unMountUserApi({ rid, uid: user.uid })
    if (res.code === '000000') {
      ElMessage.success('成员移除成功')
      await Promise.all([loadRoleMembers(), loadOrgCandidates()])
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('成员移除失败')
    }
  } finally {
    unmountLoading.value = false
    unmountingUserId.value = null
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
  const valid = await roleFormRef.value?.validate?.().catch(() => false)
  if (valid === false) {
    return
  }

  try {
    const payload = { ...form.value }
    const res = payload.roleId ? await roleUpdateApi(payload) : await roleCreateApi(payload)

    if (res.code === '000000') {
      ElMessage.success(payload.roleId ? '更新成功' : '创建成功')
      dialogVisible.value = false
      loadRoleList()
    }
  } catch (_error) {
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
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 16px;
}

.toolbar-content {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.toolbar-title {
  font-size: 16px;
  font-weight: 600;
  line-height: 24px;
  color: var(--ed-text-color-primary);
}

.toolbar-desc,
.panel-card-subtitle,
.panel-card-hint,
.role-summary-desc,
.role-summary-meta {
  font-size: 13px;
  line-height: 20px;
  color: var(--ed-text-color-secondary);
}

.member-layout {
  display: grid;
  grid-template-columns: minmax(0, 0.95fr) minmax(0, 1.25fr);
  gap: 16px;
}

.member-layout-panel {
  display: flex;
  flex-direction: column;
  gap: 16px;
  min-width: 0;
}

.panel-card {
  display: flex;
  flex-direction: column;
  gap: 16px;
  min-height: 0;
  padding: 16px;
  background: var(--ed-bg-color);
  border: 1px solid var(--ed-border-color-light);
  border-radius: var(--ed-border-radius-base);
}

.panel-card-summary {
  background: var(--ed-fill-color-light);
}

.panel-card-fill {
  height: 100%;
}

.panel-card-header,
.panel-card-footer,
.panel-card-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.panel-card-toolbar > :deep(.ed-input),
.panel-card-toolbar > :deep(.el-input) {
  flex: 1;
}

.panel-card-title {
  font-size: 14px;
  font-weight: 600;
  line-height: 22px;
  color: var(--ed-text-color-regular);
}

.panel-card-footer-pagination {
  align-items: center;
}

.role-summary {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.role-summary-name {
  font-size: 18px;
  font-weight: 600;
  line-height: 26px;
  color: var(--ed-text-color-primary);
}

.role-summary-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

@media (max-width: 1200px) {
  .member-layout {
    grid-template-columns: minmax(0, 1fr);
  }
}
</style>
