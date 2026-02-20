<template>
  <div class="audit-dashboard">
    <div class="header">
      <h2>审计仪表板</h2>
      <el-button type="primary" @click="handleRefresh">刷新</el-button>
    </div>

    <!-- 统计卡片 -->
    <div class="stats-cards">
      <div class="stat-card">
        <div class="stat-icon total">
          <el-icon><Document /></el-icon>
        </div>
        <div class="stat-content">
          <div class="stat-label">总日志数</div>
          <div class="stat-value">{{ statistics.totalLogs }}</div>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon today">
          <el-icon><Calendar /></el-icon>
        </div>
        <div class="stat-content">
          <div class="stat-label">今日操作</div>
          <div class="stat-value">{{ statistics.todayOperations }}</div>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon failed">
          <el-icon><Warning /></el-icon>
        </div>
        <div class="stat-content">
          <div class="stat-label">失败操作</div>
          <div class="stat-value">{{ statistics.failedOperations }}</div>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon alert">
          <el-icon><Bell /></el-icon>
        </div>
        <div class="stat-content">
          <div class="stat-label">异常告警</div>
          <div class="stat-value">{{ statistics.alerts }}</div>
        </div>
      </div>
    </div>

    <!-- 图表区域 -->
    <div class="charts-container">
      <div class="chart-row">
        <div class="chart-box">
          <div class="chart-title">操作类型分布</div>
          <div ref="actionTypeChartRef" class="chart" style="height: 300px"></div>
        </div>

        <div class="chart-box">
          <div class="chart-title">资源类型分布</div>
          <div ref="resourceTypeChartRef" class="chart" style="height: 300px"></div>
        </div>
      </div>

      <div class="chart-row">
        <div class="chart-box full-width">
          <div class="chart-title">操作时间趋势（近7天）</div>
          <div ref="trendChartRef" class="chart" style="height: 350px"></div>
        </div>
      </div>
    </div>

    <!-- 最近操作列表 -->
    <div class="recent-operations">
      <div class="section-title">
        <h3>最近操作</h3>
        <el-button link type="primary" @click="goToAuditLogList">查看全部</el-button>
      </div>
      <el-table :data="recentLogs" border v-loading="loading">
        <el-table-column prop="username" label="用户名" width="120" />
        <el-table-column prop="actionType" label="操作类型" width="120">
          <template #default="{ row }">
            <el-tag :type="getActionTypeTag(row.actionType)" size="small">
              {{ getActionTypeText(row.actionType) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="actionName" label="操作名称" width="150" />
        <el-table-column prop="resourceType" label="资源类型" width="100">
          <template #default="{ row }">
            {{ getResourceTypeText(row.resourceType) }}
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 'SUCCESS' ? 'success' : 'danger'" size="small">
              {{ row.status === 'SUCCESS' ? '成功' : '失败' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createTime" label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatDateTime(row.createTime) }}
          </template>
        </el-table-column>
      </el-table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus-secondary'
import { Document, Calendar, Warning, Bell } from '@element-plus/icons-vue'
import { queryAuditLogsApi } from '@/api/audit'
import * as echarts from 'echarts'

const router = useRouter()
const loading = ref(false)
const actionTypeChartRef = ref()
const resourceTypeChartRef = ref()
const trendChartRef = ref()

let actionTypeChart: echarts.ECharts | null = null
let resourceTypeChart: echarts.ECharts | null = null
let trendChart: echarts.ECharts | null = null

const statistics = reactive({
  totalLogs: 0,
  todayOperations: 0,
  failedOperations: 0,
  alerts: 0
})

const recentLogs = ref([])

onMounted(() => {
  initCharts()
  loadStatistics()
  loadRecentLogs()
  window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  disposeCharts()
})

const initCharts = () => {
  if (actionTypeChartRef.value) {
    actionTypeChart = echarts.init(actionTypeChartRef.value)
  }
  if (resourceTypeChartRef.value) {
    resourceTypeChart = echarts.init(resourceTypeChartRef.value)
  }
  if (trendChartRef.value) {
    trendChart = echarts.init(trendChartRef.value)
  }
}

const disposeCharts = () => {
  actionTypeChart?.dispose()
  resourceTypeChart?.dispose()
  trendChart?.dispose()
}

const handleResize = () => {
  actionTypeChart?.resize()
  resourceTypeChart?.resize()
  trendChart?.resize()
}

const loadStatistics = async () => {
  try {
    const res = await queryAuditLogsApi({ page: 1, pageSize: 1 })
    if (res.code === '000000') {
      statistics.totalLogs = res.data?.total || 0
    }
  } catch (error) {
    ElMessage.error('加载统计数据失败')
  }
}

const loadRecentLogs = async () => {
  loading.value = true
  try {
    const res = await queryAuditLogsApi({ page: 1, pageSize: 10 })
    if (res.code === '000000') {
      recentLogs.value = res.data?.list || []
      updateCharts(recentLogs.value)
      calculateStatistics(recentLogs.value)
    }
  } catch (error) {
    ElMessage.error('加载最近操作失败')
  } finally {
    loading.value = false
  }
}

const calculateStatistics = (logs: any[]) => {
  const today = new Date().toDateString()
  statistics.todayOperations = logs.filter(
    log => new Date(log.createTime).toDateString() === today
  ).length
  statistics.failedOperations = logs.filter(log => log.status === 'FAILED').length
  statistics.alerts = logs.filter(
    log => log.status === 'FAILED' && log.failureReason?.includes('频繁')
  ).length
}

const updateCharts = (logs: any[]) => {
  const actionTypeMap: Record<string, number> = {}
  const resourceTypeMap: Record<string, number> = {}

  logs.forEach(log => {
    actionTypeMap[log.actionType] = (actionTypeMap[log.actionType] || 0) + 1
    if (log.resourceType) {
      resourceTypeMap[log.resourceType] = (resourceTypeMap[log.resourceType] || 0) + 1
    }
  })

  updateActionTypeChart(actionTypeMap)
  updateResourceTypeChart(resourceTypeMap)
  updateTrendChart(logs)
}

const updateActionTypeChart = (data: Record<string, number>) => {
  if (!actionTypeChart) return

  const option = {
    tooltip: {
      trigger: 'item',
      formatter: '{b}: {c} ({d}%)'
    },
    legend: {
      orient: 'vertical',
      right: 10,
      top: 'center'
    },
    series: [
      {
        name: '操作类型',
        type: 'pie',
        radius: ['40%', '70%'],
        avoidLabelOverlap: false,
        itemStyle: {
          borderRadius: 10,
          borderColor: '#fff',
          borderWidth: 2
        },
        label: {
          show: false,
          position: 'center'
        },
        emphasis: {
          label: {
            show: true,
            fontSize: 16,
            fontWeight: 'bold'
          }
        },
        data: Object.entries(data).map(([key, value]) => ({
          name: getActionTypeText(key),
          value
        }))
      }
    ]
  }

  actionTypeChart.setOption(option)
}

const updateResourceTypeChart = (data: Record<string, number>) => {
  if (!resourceTypeChart) return

  const option = {
    tooltip: {
      trigger: 'item',
      formatter: '{b}: {c} ({d}%)'
    },
    legend: {
      orient: 'vertical',
      right: 10,
      top: 'center'
    },
    series: [
      {
        name: '资源类型',
        type: 'pie',
        radius: ['40%', '70%'],
        avoidLabelOverlap: false,
        itemStyle: {
          borderRadius: 10,
          borderColor: '#fff',
          borderWidth: 2
        },
        data: Object.entries(data).map(([key, value]) => ({
          name: getResourceTypeText(key),
          value
        }))
      }
    ]
  }

  resourceTypeChart.setOption(option)
}

const updateTrendChart = (logs: any[]) => {
  if (!trendChart) return

  const dateMap: Record<string, Record<string, number>> = {}

  logs.forEach(log => {
    const date = new Date(log.createTime).toLocaleDateString('zh-CN')
    if (!dateMap[date]) {
      dateMap[date] = { success: 0, failed: 0 }
    }
    if (log.status === 'SUCCESS') {
      dateMap[date].success++
    } else {
      dateMap[date].failed++
    }
  })

  const dates = Object.keys(dateMap).sort()
  const successData = dates.map(d => dateMap[d].success)
  const failedData = dates.map(d => dateMap[d].failed)

  const option = {
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'cross'
      }
    },
    legend: {
      data: ['成功', '失败']
    },
    xAxis: {
      type: 'category',
      data: dates
    },
    yAxis: {
      type: 'value'
    },
    series: [
      {
        name: '成功',
        type: 'line',
        data: successData,
        smooth: true,
        itemStyle: { color: '#67C23A' }
      },
      {
        name: '失败',
        type: 'line',
        data: failedData,
        smooth: true,
        itemStyle: { color: '#F56C6C' }
      }
    ]
  }

  trendChart.setOption(option)
}

const handleRefresh = () => {
  loadStatistics()
  loadRecentLogs()
}

const goToAuditLogList = () => {
  router.push('/system/audit')
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
    USER_ACTION: '用户操作',
    PERMISSION_CHANGE: '权限变更',
    DATA_ACCESS: '数据访问',
    SYSTEM_CONFIG: '系统配置'
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
.audit-dashboard {
  padding: 20px;

  .header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 24px;
  }

  .stats-cards {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 20px;
    margin-bottom: 24px;

    .stat-card {
      background: #fff;
      border-radius: 8px;
      padding: 20px;
      display: flex;
      align-items: center;
      gap: 16px;
      box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);

      .stat-icon {
        width: 60px;
        height: 60px;
        border-radius: 50%;
        display: flex;
        align-items: center;
        justify-content: center;
        font-size: 28px;

        &.total {
          background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
          color: #fff;
        }

        &.today {
          background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
          color: #fff;
        }

        &.failed {
          background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
          color: #fff;
        }

        &.alert {
          background: linear-gradient(135deg, #fa709a 0%, #fee140 100%);
          color: #fff;
        }
      }

      .stat-content {
        flex: 1;

        .stat-label {
          font-size: 14px;
          color: #909399;
          margin-bottom: 8px;
        }

        .stat-value {
          font-size: 28px;
          font-weight: bold;
          color: #303133;
        }
      }
    }
  }

  .charts-container {
    margin-bottom: 24px;

    .chart-row {
      display: grid;
      grid-template-columns: repeat(2, 1fr);
      gap: 20px;
      margin-bottom: 20px;

      .chart-box {
        background: #fff;
        border-radius: 8px;
        padding: 20px;
        box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);

        &.full-width {
          grid-column: span 2;
        }

        .chart-title {
          font-size: 16px;
          font-weight: bold;
          margin-bottom: 16px;
          color: #303133;
        }

        .chart {
          width: 100%;
        }
      }
    }
  }

  .recent-operations {
    background: #fff;
    border-radius: 8px;
    padding: 20px;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);

    .section-title {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 16px;

      h3 {
        margin: 0;
        font-size: 16px;
        font-weight: bold;
        color: #303133;
      }
    }
  }
}
</style>
