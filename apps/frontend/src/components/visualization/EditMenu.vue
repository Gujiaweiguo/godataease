<script lang="ts" setup>
import languageSvg from '@/assets/svg/language.svg'
import { ref, onMounted } from 'vue'
import { Icon } from '@/components/icon-custom'
import { useUserStoreWithOut } from '@/store/modules/user'
const userStore = useUserStoreWithOut()
const language = ref(null)
onMounted(() => {
  language.value = userStore.getLanguage
})

const handleSetLanguage = (lang: string) => {
  userStore.setLanguage(lang)
}

</script>
<template>
  <el-dropdown
    style="display: flex; align-items: center"
    trigger="click"
    class="international"
    @command="handleSetLanguage"
  >
    <el-icon>
      <Icon name="language"><languageSvg class="svg-icon" /></Icon>
    <template #dropdown>
      <el-dropdown-menu>
        <el-dropdown-item :disabled="language === 'zh-CN'" command="zh-CN"
          >简体中文</el-dropdown-item
        >
        <el-dropdown-item :disabled="language === 'tw'" command="tw"> 繁體中文 </el-dropdown-item>
        <el-dropdown-item :disabled="language === 'en'" command="en"> English </el-dropdown-item>
      </el-dropdown-menu>
    </template>
  </el-dropdown>
</template>

<style lang="less">
.right-menu-item {
  display: inline-block;
  padding: 10px 8px;
  height: 100%;

  vertical-align: text-bottom;

  &.hover-effect {
    cursor: pointer;
    transition: background 0.3s;

    &:hover {
      background-color: rgba(0, 0, 0, 0.025);
    }
  }
}
</style>
