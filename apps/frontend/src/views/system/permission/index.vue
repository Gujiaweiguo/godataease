<template>
  <div class="permission-config">
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
import { ref } from 'vue'
import MenuPermission from './MenuPermission.vue'
import ResourcePermission from './ResourcePermission.vue'
import DataPermission from './DataPermission.vue'

const activeTab = ref('menu')
</script>

<style scoped>
.permission-config {
  padding: 20px;
}

.permission-tabs {
  background: #fff;
  padding: 16px;
  border-radius: 4px;
}
</style>
