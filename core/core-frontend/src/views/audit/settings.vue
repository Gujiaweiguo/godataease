<template>
  <div class="audit-settings">
    <div class="header">
      <h2>审计设置</h2>
    </div>

    <el-form :model="settings" label-width="200px" class="settings-form">
      <el-card class="setting-card">
        <template #header>
          <span class="card-title">日志保留策略</span>
        </template>

        <el-form-item label="日志保留天数">
          <el-input-number
            v-model="settings.retentionDays"
            :min="7"
            :max="365"
            :step="7"
            controls-position="right"
          />
          <span class="unit">天</span>
          <div class="form-tip">超过此天数的审计日志将被自动清理</div>
        </el-form-item>

        <el-form-item label="自动清理频率">
          <el-select v-model="settings.cleanupFrequency">
            <el-option label="每日" value="daily" />
            <el-option label="每周" value="weekly" />
            <el-option label="每月" value="monthly" />
          </el-select>
          <div class="form-tip">系统将按设定频率自动清理过期日志</div>
        </el-form-item>

        <el-form-item>
          <el-button type="primary" @click="handleSaveRetentionSettings"> 保存保留策略 </el-button>
          <el-button @click="handleCleanupNow"> 立即清理 </el-button>
        </el-form-item>
      </el-card>

      <el-card class="setting-card">
        <template #header>
          <span class="card-title">告警设置</span>
        </template>

        <el-form-item label="启用告警">
          <el-switch v-model="settings.enableAlerts" />
          <div class="form-tip">开启后系统将检测异常操作并发送告警</div>
        </el-form-item>

        <el-form-item label="失败登录告警阈值">
          <el-input-number
            v-model="settings.failedLoginThreshold"
            :min="3"
            :max="20"
            :step="1"
            controls-position="right"
            :disabled="!settings.enableAlerts"
          />
          <span class="unit">次/5分钟</span>
          <div class="form-tip">5分钟内超过此阈值的失败登录将触发告警</div>
        </el-form-item>

        <el-form-item label="权限变更告警">
          <el-switch
            v-model="settings.alertOnPermissionChange"
            :disabled="!settings.enableAlerts"
          />
          <div class="form-tip">权限变更操作发生时发送告警通知</div>
        </el-form-item>

        <el-form-item label="敏感数据访问告警">
          <el-switch v-model="settings.alertOnSensitiveAccess" :disabled="!settings.enableAlerts" />
          <div class="form-tip">访问敏感数据集或仪表板时发送告警</div>
        </el-form-item>

        <el-form-item label="批量操作告警阈值">
          <el-input-number
            v-model="settings.batchOperationThreshold"
            :min="10"
            :max="100"
            :step="5"
            controls-position="right"
            :disabled="!settings.enableAlerts"
          />
          <span class="unit">次/小时</span>
          <div class="form-tip">1小时内执行超过此阈值的操作将触发告警</div>
        </el-form-item>

        <el-form-item>
          <el-button
            type="primary"
            @click="handleSaveAlertSettings"
            :disabled="!settings.enableAlerts"
          >
            保存告警设置
          </el-button>
        </el-form-item>
      </el-card>

      <el-card class="setting-card">
        <template #header>
          <span class="card-title">通知方式</span>
        </template>

        <el-form-item label="邮件通知">
          <el-switch v-model="settings.enableEmailNotification" />
          <div class="form-tip">通过邮件发送告警通知</div>
        </el-form-item>

        <el-form-item label="通知邮箱" v-if="settings.enableEmailNotification">
          <el-input
            v-model="settings.notificationEmail"
            placeholder="请输入接收告警的邮箱地址"
            type="email"
          />
        </el-form-item>

        <el-form-item label="系统通知">
          <el-switch v-model="settings.enableSystemNotification" />
          <div class="form-tip">在系统内显示告警通知</div>
        </el-form-item>

        <el-form-item>
          <el-button type="primary" @click="handleSaveNotificationSettings">
            保存通知设置
          </el-button>
          <el-button @click="handleTestNotification"> 发送测试通知 </el-button>
        </el-form-item>
      </el-card>

      <el-card class="setting-card">
        <template #header>
          <span class="card-title">日志导出设置</span>
        </template>

        <el-form-item label="默认导出格式">
          <el-radio-group v-model="settings.defaultExportFormat">
            <el-radio value="csv">CSV</el-radio>
            <el-radio value="json">JSON</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item label="单次导出上限">
          <el-input-number
            v-model="settings.exportLimit"
            :min="100"
            :max="10000"
            :step="100"
            controls-position="right"
          />
          <span class="unit">条</span>
          <div class="form-tip">单次最多导出的日志条数</div>
        </el-form-item>

        <el-form-item>
          <el-button type="primary" @click="handleSaveExportSettings"> 保存导出设置 </el-button>
        </el-form-item>
      </el-card>
    </el-form>
  </div>
</template>

<script setup lang="ts">
import { reactive } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus-secondary'

const settings = reactive({
  retentionDays: 90,
  cleanupFrequency: 'weekly',
  enableAlerts: true,
  failedLoginThreshold: 5,
  alertOnPermissionChange: true,
  alertOnSensitiveAccess: false,
  batchOperationThreshold: 50,
  enableEmailNotification: false,
  notificationEmail: '',
  enableSystemNotification: true,
  defaultExportFormat: 'csv',
  exportLimit: 1000
})

const handleSaveRetentionSettings = () => {
  ElMessage.success('保留策略已保存')
}

const handleCleanupNow = async () => {
  try {
    await ElMessageBox.confirm('确定要立即清理过期审计日志吗？此操作不可撤销。', '确认清理', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    ElMessage.success('清理完成')
  } catch {
    // 用户取消
  }
}

const handleSaveAlertSettings = () => {
  ElMessage.success('告警设置已保存')
}

const handleSaveNotificationSettings = () => {
  ElMessage.success('通知设置已保存')
}

const handleTestNotification = () => {
  ElMessage.success('测试通知已发送')
}

const handleSaveExportSettings = () => {
  ElMessage.success('导出设置已保存')
}
</script>

<style scoped lang="scss">
.audit-settings {
  padding: 20px;

  .header {
    margin-bottom: 24px;
  }

  .settings-form {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 20px;

    .setting-card {
      .card-title {
        font-size: 16px;
        font-weight: bold;
        color: #303133;
      }

      .unit {
        margin-left: 8px;
        color: #909399;
      }

      .form-tip {
        margin-top: 8px;
        font-size: 12px;
        color: #909399;
        line-height: 1.5;
      }

      :deep(.el-form-item) {
        margin-bottom: 24px;
      }
    }
  }
}
</style>
