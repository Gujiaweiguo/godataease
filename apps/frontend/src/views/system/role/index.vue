<template>
  <div class="role-entry-bridge">
    <el-result icon="info" title="角色管理已合并到用户管理">
      <template #sub-title>
        <p>正在跳转到用户管理中的角色页签；若未自动跳转，可手动进入。</p>
      </template>
      <template #extra>
        <el-button type="primary" @click="goToHostedRoleTab">进入角色页签</el-button>
      </template>
    </el-result>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const router = useRouter()
const route = useRoute()

const goToHostedRoleTab = () => {
  router.replace({
    path: '/system/user',
    query: {
      ...route.query,
      tab: 'role'
    }
  })
}

onMounted(() => {
  goToHostedRoleTab()
})
</script>

<style scoped>
.role-entry-bridge {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 320px;
  padding: 24px;
  background: #fff;
}
</style>
