<script lang="ts" setup>
import { computed } from 'vue'
import { ElMenu } from 'element-plus-secondary'
import { getCSSVariable } from '@/utils/color'
import { useRoute, useRouter } from 'vue-router_2'
import type { RouteRecordRaw } from 'vue-router_2'
import { isExternal } from '@/utils/validate'
import { useCache } from '@/hooks/web/useCache'
import MenuItem from './MenuItem.vue'
import { useAppearanceStoreWithOut } from '@/store/modules/appearance'
import { usePermissionStore } from '@/store/modules/permission'
import { buildMenuSelectPath, resolveActiveTopPath, resolveSideMenus, resolveTopMenus } from './menu-utils'
import { isDynamicNavigationEnabled } from '@/utils/featureFlags'
import { executeMenuAction } from '@/utils/menu-actions'
import type { UserMenuItem } from '@/store/modules/permission'

const appearanceStore = useAppearanceStoreWithOut()
const tempColor = computed(() => {
  return {
    '--temp-color':
      (appearanceStore.themeColor === 'custom' ? appearanceStore.customColor : getCSSVariable()) +
      '1A'
  }
})
defineProps({
  collapse: Boolean
})

const route = useRoute()
const { wsCache } = useCache('localStorage')
const { push } = useRouter()
const permissionStore = usePermissionStore()
const dynamicNavigationEnabled = computed(() => isDynamicNavigationEnabled())

type MenuRouteMeta = {
  hidden?: boolean
}

type MenuRoute = RouteRecordRaw & {
  hidden?: boolean
  meta?: MenuRouteMeta
}

const topMenus = computed(() => {
  return resolveTopMenus(permissionStore.getRoutersNotHidden as AppCustomRouteRecordRaw[])
})

const legacyPath = computed(() => route.matched[0]?.path || '')
const legacyMenuList = computed(() => {
  const root = route.matched[0]
  if (!root?.children?.length) {
    return []
  }
  return root.children.filter(item => {
    const menu = item as MenuRoute
    return !menu.hidden && !menu.meta?.hidden
  })
})

const path = computed(() => {
  if (!dynamicNavigationEnabled.value) {
    return legacyPath.value
  }
  return resolveActiveTopPath(route.path, topMenus.value) || route.matched[0]?.path || ''
})

const menuList = computed(() => {
  if (!dynamicNavigationEnabled.value) {
    return legacyMenuList.value
  }
  return resolveSideMenus(route.path, topMenus.value)
})

const activeIndex = computed(() => {
  const arr = route.path.split('/')
  return arr[arr.length - 1]
})
const findUserMenuItem = (fullPath: string, menus: UserMenuItem[]): UserMenuItem | null => {
  for (const menu of menus) {
    if (menu.path === fullPath) return menu
    if (menu.children?.length) {
      const found = findUserMenuItem(fullPath, menu.children)
      if (found) return found
    }
  }
  return null
}

const menuSelect = (index: string, indexPath: string[]) => {
  //   自定义事件
  if (isExternal(index)) {
    const openType = wsCache.get('open-backend') === '1' ? '_self' : '_blank'
    window.open(index, openType)
    return
  }
  // Build the full resolved path to look up menu metadata
  const resolvedPath = dynamicNavigationEnabled.value
    ? buildMenuSelectPath(path.value, index, indexPath)
    : `${path.value}/${indexPath.join('/')}`
  // Check if this menu item is an event-type menu (menuType === 'event')
  const menuItem = findUserMenuItem(resolvedPath, permissionStore.getUserMenus)
  if (menuItem?.menuType === 'event') {
    const event = menuItem.actionConfig?.event as string | undefined
    if (event) {
      executeMenuAction(event, menuItem.actionConfig)
      return
    }
  }
  push(resolvedPath)
}
</script>

<template>
  <el-menu
    :style="tempColor"
    @select="menuSelect"
    :default-active="activeIndex"
    class="el-menu-vertical"
    :collapse="collapse"
  >
    <MenuItem v-for="menu in menuList" :key="menu.path" :menu="menu"></MenuItem>
  </el-menu>
</template>

<style lang="less" scoped>
.ed-menu-vertical:not(.ed-menu--collapse) {
  width: 100%;
  min-height: 400px;
}

.ed-menu {
  border: none;
  .ed-menu-item:not(.is-active) {
    &:hover {
      background-color: #1f23291a !important;
    }
  }
  .is-active:not(.ed-sub-menu) {
    background-color: var(--temp-color);
  }
  :deep(.ed-sub-menu) {
    margin: 0;
    .ed-sub-menu__title {
      &:hover {
        background-color: #1f23291a;
      }
    }
    .ed-menu-item:not(.is-active) {
      &:hover {
        background-color: #1f23291a !important;
      }
    }
    ul.ed-menu {
      li.ed-menu-item {
        i {
          width: 4px !important;
        }
      }
    }
  }
  :deep(.ed-sub-menu.is-active) {
    .ed-sub-menu__title {
      color: var(--ed-color-primary);
    }
    .is-active {
      background-color: var(--temp-color);
    }
  }
}
</style>
