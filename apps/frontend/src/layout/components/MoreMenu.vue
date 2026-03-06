<script lang="ts" setup>
import { computed } from 'vue'
import { useAppearanceStoreWithOut } from '@/store/modules/appearance'
import { usePermissionStore } from '@/store/modules/permission'
import { useI18n } from '@/hooks/web/useI18n'
import { useEmitt } from '@/hooks/web/useEmitt'
import { useRouter } from 'vue-router_2'
import iconMore from '@/assets/svg/icon_more_outlined.svg'
import dvPreviewDownload from '@/assets/svg/icon_download_outlined.svg'
import toolboxLog from '@/assets/svg/toolbox-log.svg'
import toolboxData_fill from '@/assets/svg/toolbox-data_fill.svg'
import toolboxIcon_template from '@/assets/svg/toolbox-icon_template.svg'
import topHelpDoc from '@/assets/svg/top-help-doc.svg'
import topProductBbs from '@/assets/svg/top-product-bbs.svg'
import topTechnology from '@/assets/svg/top-technology.svg'
import topEnterpriseTrial from '@/assets/svg/top-enterprise-trial.svg'

const appearanceStore = useAppearanceStoreWithOut()
const permissionStore = usePermissionStore()
const { t } = useI18n()
const { push, resolve } = useRouter()
const navigateBg = computed(() => appearanceStore.getNavigateBg)
const help = computed(() => appearanceStore.getHelp)

// 导出中心
const downloadClick = () => {
  useEmitt().emitter.emit('data-export-center')
}

// 工具箱
const showToolbox = computed(() => {
  return permissionStore.getRouters.some(route => route.path === '/toolbox')
})

const toolboxItems = computed(() => {
  const items: { name: string; rName: string; icon: any }[] = []
  const toolboxMenu = resolve('/toolbox')
  if (!toolboxMenu) return items
  
  const children = toolboxMenu.matched[0]?.children
  if (!children?.length) return items

  const iconMap: Record<string, any> = {
    'toolbox-data_fill': toolboxData_fill,
    'toolbox-icon_template': toolboxIcon_template,
    'toolbox-log': toolboxLog
  }

  children.forEach(item => {
    items.push({
      name: String(item.meta?.title || ''),
      rName: String(item.name || ''),
      icon: iconMap['toolbox-' + item.meta?.icon]
    })
  })
  return items
})

const toToolbox = (item: { rName: string }) => {
  push({ name: item.rName })
}

// 帮助文档
const showDoc = computed(() => appearanceStore.getShowDoc)

const helpItems = computed(() => [
  {
    name: t('api_pagination.help_documentation'),
    url: help.value || 'https://dataease.io/docs/v2/',
    icon: topHelpDoc
  },
  {
    name: t('api_pagination.product_forum'),
    url: 'https://bbs.fit2cloud.com/c/de/6',
    icon: topProductBbs
  },
  {
    name: t('api_pagination.technical_blog'),
    url: 'https://blog.fit2cloud.com/categories/dataease',
    icon: topTechnology
  },
  {
    name: t('api_pagination.enterprise_edition_trial'),
    url: 'https://jinshuju.net/f/TK5TTd',
    icon: topEnterpriseTrial
  }
])

const openHelp = (item: { url: string }) => {
  window.open(item.url, '_blank')
}
</script>

<template>
  <el-popover
    :show-arrow="false"
    popper-class="more-menu-popover"
    placement="bottom-end"
    :width="280"
    trigger="hover"
  >
    <div class="more-menu-content">
      <!-- 导出中心 -->
      <div class="more-menu-item" @click="downloadClick">
        <el-icon><dvPreviewDownload class="svg-icon menu-icon" /></el-icon>
        <span>{{ t('data_export.export_center') }}</span>
      </div>

      <!-- 工具箱 -->
      <template v-if="showToolbox && toolboxItems.length">
        <div class="more-menu-divider" />
        <div class="more-menu-section-title">{{ t('toolbox.name') }}</div>
        <div class="more-menu-grid">
          <div
            v-for="(item, index) in toolboxItems"
            :key="index"
            class="more-menu-grid-item"
            @click="toToolbox(item)"
          >
            <el-icon><component :is="item.icon" class="svg-icon" /></el-icon>
            <span>{{ item.name }}</span>
          </div>
        </div>
      </template>

      <!-- 帮助 -->
      <template v-if="showDoc">
        <div class="more-menu-divider" />
        <div class="more-menu-section-title">{{ t('api_pagination.help_documentation') }}</div>
        <div class="more-menu-grid">
          <div
            v-for="(item, index) in helpItems"
            :key="index"
            class="more-menu-grid-item"
            @click="openHelp(item)"
          >
            <el-icon><component :is="item.icon" class="svg-icon" /></el-icon>
            <span>{{ item.name }}</span>
          </div>
        </div>
      </template>
    </div>

    <template #reference>
      <div
        class="more-menu-trigger"
        :class="{ 'is-light-setting': navigateBg === 'light' }"
      >
        <el-icon>
          <iconMore class="svg-icon" />
        </el-icon>
      </div>
    </template>
  </el-popover>
</template>

<style lang="less" scoped>
.more-menu-trigger {
  margin: 0 10px;
  padding: 5px;
  height: 28px;
  width: 28px;
  border-radius: 4px;
  overflow: hidden;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;

  &:hover {
    background-color: #1e2738;
  }

  &.is-light-setting:hover {
    background-color: #1f23291a !important;
  }
}

.more-menu-content {
  padding: 8px 0;
}

.more-menu-item {
  display: flex;
  align-items: center;
  padding: 10px 16px;
  cursor: pointer;
  color: #1f2329;

  &:hover {
    background: #1f23291a;
  }

  .menu-icon {
    width: 18px;
    height: 18px;
    margin-right: 8px;
  }

  span {
    font-size: 14px;
  }
}

.more-menu-divider {
  height: 1px;
  background: #1f232926;
  margin: 8px 0;
}

.more-menu-section-title {
  padding: 4px 16px;
  font-size: 12px;
  color: #8f959e;
}

.more-menu-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 8px;
  padding: 8px 12px;
}

.more-menu-grid-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 12px 8px;
  cursor: pointer;
  border-radius: 4px;

  &:hover {
    background: #1f23291a;
  }

  .el-icon {
    margin-bottom: 4px;
  }

  span {
    font-size: 12px;
    color: #1f2329;
    text-align: center;
  }
}
</style>

<style lang="less">
.more-menu-popover {
  padding: 0 !important;
}
</style>
