<template>
  <div class="permission-config">
    <div class="permission-config-header">
      <div class="permission-config-title">统一权限配置中心</div>
      <div class="permission-config-desc">
        统一承载菜单权限、资源权限与行列权限，角色仅作为授权载体使用。
      </div>
    </div>
    <el-tabs v-model="activeTab" class="permission-tabs">
      <el-tab-pane label="菜单权限" name="menu">
        <MenuPermission />
      </el-tab-pane>
      <el-tab-pane label="资源权限" name="resource">
        <ResourcePermission />
      </el-tab-pane>
      <el-tab-pane label="行列权限" name="data">
        <DataPermission />
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import MenuPermission from './MenuPermission.vue'
import ResourcePermission from './ResourcePermission.vue'
import DataPermission from './DataPermission.vue'

const route = useRoute()
const router = useRouter()
const activeTab = ref('menu')

const normalizePermissionTab = (tab?: string) => {
  if (tab === 'resource' || tab === 'data') {
    return tab
  }
  return 'menu'
}

watch(
  () => route.query.tab,
  value => {
    const tab = typeof value === 'string' ? value : undefined
    activeTab.value = normalizePermissionTab(tab)
  },
  { immediate: true }
)

watch(activeTab, (tab: string) => {
  const nextTab = normalizePermissionTab(tab)
  const currentTab = typeof route.query.tab === 'string' ? route.query.tab : undefined
  if (normalizePermissionTab(currentTab) === nextTab) {
    return
  }

  router.replace({
    query: {
      ...route.query,
      tab: nextTab
    }
  })
})
</script>

<style scoped>
.permission-config {
  padding: 20px;
}

.permission-config-header {
  margin-bottom: 16px;
}

.permission-config-title {
  font-size: 16px;
  font-weight: 600;
  line-height: 24px;
  color: var(--ed-text-color-primary);
}

.permission-config-desc {
  margin-top: 4px;
  font-size: 13px;
  line-height: 20px;
  color: var(--ed-text-color-secondary);
}

.permission-tabs {
  padding: 16px;
  background: #fff;
  border-radius: 4px;
}
</style>
