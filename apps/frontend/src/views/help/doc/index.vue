<script lang="ts" setup>
import { computed } from 'vue'
import { useAppearanceStoreWithOut } from '@/store/modules/appearance'

const appearanceStore = useAppearanceStoreWithOut()
const helpUrl = computed(() => appearanceStore.getHelp || 'https://dataease.io/docs/v2/')
const openHelp = () => {
  window.open(helpUrl.value, '_blank')
}
</script>

<template>
  <div class="help-doc-page">
    <div class="help-card">
      <h2>Help Documentation</h2>
      <p>Click the button below to open the help documentation in a new tab.</p>
      <el-button type="primary" @click="openHelp">Open Help</el-button>
      <div class="help-link">{{ helpUrl }}</div>
    </div>
  </div>
</template>

<style lang="less" scoped>
.help-doc-page {
  padding: 24px;
  display: flex;
  justify-content: center;
}
.help-card {
  width: 840px;
  background: #fff;
  border-radius: 8px;
  padding: 32px;
}
.help-link {
  margin-top: 16px;
  color: #646a73;
  word-break: break-all;
}
</style>
