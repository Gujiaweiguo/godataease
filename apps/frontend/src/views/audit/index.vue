<template>
  <div class="audit-log-management">
    <div class="header">
      <h2>审计日志管理</h2>
    </div>

    <div class="filter-bar">
      <el-form :inline="true" :model="filterForm" label-width="120px">
        <el-form-item label="用户">
          <el-select v-model="filterForm.userId" placeholder="全部用户" clearable filterable>
            <el-option label="全部用户" :value="null" />
            <el-option
              v-for="user in userList"
              :key="user.userId"
              :label="user.username"
              :value="user.userId"
            />
          </el-select>
        </el-form-item>

        <el-form-item label="操作类型">
          <el-select v-model="filterForm.actionType" placeholder="全部类型" clearable filterable>
            <el-option label="全部类型" :value="null" />
            <el-option label="用户操作" value="USER_ACTION" />
            <el-option label="权限变更" value="PERMISSION_CHANGE" />
            <el-option label="数据访问" value="DATA_ACCESS" />
            <el-option label="系统配置" value="SYSTEM_CONFIG" />
          </el-select>
        </el-form-item>

        <el-form-item label="资源类型">
          <el-select v-model="filterForm.resourceType" placeholder="全部资源" clearable filterable>
            <el-option label="全部资源" :value="null" />
            <el-option label="用户" value="USER" />
            <el-option label="组织" value="ORGANIZATION" />
            <el-option label="角色" value="ROLE" />
            <el-option label="权限" value="PERMISSION" />
            <el-option label="数据集" value="DATASET" />
            <el-option label="仪表板" value="DASHBOARD" />
          </el-select>
        </el-form-item>

        <el-form-item label="日期范围">
          <el-date-picker
            v-model="filterForm.startTime"
            type="datetime"
            placeholder="开始时间"
            value-format="YYYY-MM-DD HH:mm:ss"
          />
          <span style="margin: 0 10px">至</span>
          <el-date-picker
            v-model="filterForm.endTime"
            type="datetime"
            placeholder="结束时间"
            value-format="YYYY-MM-DD HH:mm:ss"
          />
        </el-form-item>

        <el-form-item>
          <el-button type="primary" @click="handleQuery">查询</el-button>
          <el-button type="default" @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </div>

    <el-table :data="auditLogList" border v-loading="loading" stripe>
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="username" label="用户名" width="120" />
      <el-table-column prop="actionType" label="操作类型" width="120">
        <template #default="{ row }">
          <el-tag :type="getActionTypeTag(row.actionType)" size="small">
            {{ getActionTypeText(row.actionType) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="resourceType" label="资源类型" width="100">
        <template #default="{ row }">
          {{ getResourceTypeText(row.resourceType) }}
        </template>
      </el-table-column>
      <el-table-column prop="actionName" label="操作名称" width="150" />
      <el-table-column prop="resourceName" label="资源名称" width="120" />
      <el-table-column prop="resourceId" label="资源ID" width="80" />
      <el-table-column prop="operation" label="操作" width="80">
        <template #default="{ row }">
          {{ getOperationText(row.operation) }}
        </template>
      </el-table-column>
      <el-table-column prop="status" label="状态" width="80">
        <template #default="{ row }">
          <el-tag :type="row.status === 'SUCCESS' ? 'success' : 'danger'" size="small">
            {{ row.status === 'SUCCESS' ? '成功' : '失败' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="failureReason" label="失败原因" width="150" />
      <el-table-column prop="ipAddress" label="IP地址" width="120" />
      <el-table-column prop="createTime" label="创建时间" width="160">
        <template #default="{ row }">
          {{ formatDateTime(row.createTime) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="100" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" size="small" @click="handleViewDetail(row)">
            详情
          </el-button>
          <el-button link type="primary" size="small" @click="handleExport([row.id])">
            导出
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-pagination
      v-model:current-page="pagination.currentPage"
      v-model:page-size="pagination.pageSize"
      :total="pagination.total"
      @update:current-page="handlePageChange"
      @update:page-size="handleSizeChange"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus-secondary'
import { queryAuditLogsApi, exportAuditLogsApi } from '@/api/audit'

const loading = ref(false)
const auditLogList = ref([])

const filterForm = reactive({
  userId: null,
  actionType: null,
  resourceType: null,
  startTime: null,
  endTime: null
})

const pagination = reactive({
  currentPage: 1,
  pageSize: 10,
  total: 0
})

onMounted(() => {
  handleQuery()
})

const handleQuery = async () => {
  loading.value = true
  try {
    const params: any = {}
    if (filterForm.userId) params.userId = filterForm.userId
    if (filterForm.actionType) params.actionType = filterForm.actionType
    if (filterForm.resourceType) params.resourceType = filterForm.resourceType
    if (filterForm.startTime) params.startTime = filterForm.startTime
    if (filterForm.endTime) params.endTime = filterForm.endTime

    const res = await queryAuditLogsApi(params)
    if (res.code === '000000') {
      auditLogList.value = res.data?.list || []
      pagination.total = res.data?.total || 0
    } else {
      ElMessage.error(res.msg || '查询失败')
    }
  } catch (error) {
    ElMessage.error('查询审计日志失败')
  } finally {
    loading.value = false
  }
}

const handleReset = () => {
  Object.assign(filterForm, {
    userId: null,
    actionType: null,
    resourceType: null,
    startTime: null,
    endTime: null
  })
  pagination.currentPage = 1
}

const handlePageChange = (page: number) => {
  pagination.currentPage = page
  handleQuery()
}

const handleSizeChange = (size: number) => {
  pagination.pageSize = size
  handleQuery()
}

const handleViewDetail = (row: any) => {
  ElMessage.info(`查看详情: ID=${row.id}`)
}

const handleExport = async (ids: number[]) => {
  try {
    await ElMessageBox.confirm('确定要导出选中的审计日志吗?', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await exportAuditLogsApi(ids, 'csv')
    ElMessage.success('导出成功')
  } catch (error) {
    ElMessage.error('导出失败')
  }
}

const getActionTypeTag = (actionType: string) => {
  const typeMap: Record<string, string> = {
    USER_ACTION: 'info',
    PERMISSION_CHANGE: 'warning',
    DATA_ACCESS: 'primary',
    SYSTEM_CONFIG: 'default'
  }
  return typeMap[actionType] || 'default'
}

const getActionTypeText = (actionType: string) => {
  const textMap: Record<string, string> = {
    CREATE: '创建',
    UPDATE: '更新',
    DELETE: '删除',
    EXPORT: '导出',
    LOGIN: '登录',
    LOGOUT: '登出'
  }
  return textMap[actionType] || actionType
}

const getResourceTypeText = (resourceType: string) => {
  const typeMap: Record<string, string> = {
    USER: '用户',
    ORGANIZATION: '组织',
    ROLE: '角色',
    PERMISSION: '权限',
    DATASET: '数据集',
    DASHBOARD: '仪表板'
  }
  return typeMap[resourceType] || resourceType
}

const formatDateTime = (dateStr: string) => {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN', { hour12: false })
}
</script>

<style scoped lang="scss">
.audit-log-management {
  padding: 20px;

  .header {
    margin-bottom: 20px;
  }

  .filter-bar {
    background: #f5f7fa;
    padding: 20px;
    border-radius: 4px;
    display: flex;
    gap: 16px;
    align-items: center;
  }
}
</style>
