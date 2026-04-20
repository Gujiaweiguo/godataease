import { defineStore } from 'pinia'
import { routes } from '@/router'

import { generateRoutesFn2 } from '@/router/establish'
import { store } from '../index'
import { cloneDeep } from 'lodash-es'
import NotFoundPage from '@/views/404/index.vue'

export const DYNAMIC_NOT_FOUND_ROUTE_NAME = 'dynamic-not-found'

export interface UserMenuItem {
  id: number
  name: string
  path: string
  icon?: string
  menuType?: string
  actionConfig?: Record<string, unknown>
  menuSort?: number
  hidden?: boolean
  children?: UserMenuItem[]
}

export interface PermissionState {
  routers: AppRouteRecordRaw[]
  addRouters: AppRouteRecordRaw[]
  isAddRouters: boolean
  currentPath: string
  userMenus: UserMenuItem[]
  helpMenus: UserMenuItem[]
}

export const usePermissionStore = defineStore('permission', {
  state: (): PermissionState => ({
    routers: [],
    addRouters: [],
    isAddRouters: false,
    currentPath: '',
    userMenus: [],
    helpMenus: []
  }),
  getters: {
    getRouters(): AppRouteRecordRaw[] {
      return this.routers
    },
    getRoutersNotHidden(): AppRouteRecordRaw[] {
      return this.routers.filter(ele => !ele.hidden)
    },
    getAddRouters(): AppRouteRecordRaw[] {
      return cloneDeep(this.addRouters)
    },
    getIsAddRouters(): boolean {
      return this.isAddRouters
    },
    getCurrentPath(): boolean {
      return this.currentPath
    },
    getUserMenus(): UserMenuItem[] {
      return this.userMenus
    },
    getHelpMenus(): UserMenuItem[] {
      return this.helpMenus
    }
  },
  actions: {
    clear() {
      this.routers = cloneDeep(routes)
      this.addRouters = []
      this.isAddRouters = false
      this.currentPath = ''
    },
    generateRoutes(routers?: AppCustomRouteRecordRaw[] | string[]): Promise<unknown> {
      return new Promise<void>(resolve => {
        let routerMap: AppRouteRecordRaw[] = []
        routerMap = generateRoutesFn2(routers as AppCustomRouteRecordRaw[]) || []

        this.addRouters = routerMap.concat([
          {
            path: '/:catchAll(.*)',
            name: DYNAMIC_NOT_FOUND_ROUTE_NAME,
            component: NotFoundPage,
            meta: {
              hidden: true
            }
          }
        ])
        // 渲染菜单的所有路由
        this.routers = cloneDeep(routes).concat(routerMap)
        resolve()
      })
    },
    setCurrentPath(currentPath: string): void {
      this.currentPath = currentPath
    },
    setIsAddRouters(state: boolean): void {
      this.isAddRouters = state
    },
    setUserMenus(menus: UserMenuItem[]): void {
      this.userMenus = menus
    },
    setHelpMenus(menus: UserMenuItem[]): void {
      this.helpMenus = menus
    }
  }
})

export const usePermissionStoreWithOut = () => {
  return usePermissionStore(store)
}

export const pathValid = path => {
  if (path?.startsWith('/dataset-form')) {
    path = '/data/dataset'
  }
  const permissionStore = usePermissionStore(store)
  const routers = permissionStore.getRouters
  const temp = path.startsWith('/') ? path.substr(1) : path
  const locations = temp.split('/')
  if (locations.length === 0) {
    return false
  }

  return hasCurrentRouter(locations, routers, 0)
}
/**
 * 递归验证every level
 * @param {*} locations
 * @param {*} routers
 * @param {*} index
 * @returns
 */
const hasCurrentRouter = (locations, routers, index) => {
  if (!routers?.length) {
    return false
  }

  const location = locations[index]
  const matchedRouters = routers.filter(router => {
    return router.path === location || '/' + location === router.path
  })

  if (!matchedRouters.length) {
    return false
  }

  if (index >= locations.length - 1) {
    return true
  }

  return matchedRouters.some(router => {
    return hasCurrentRouter(locations, router.children || [], index + 1)
  })
}

export const getFirstAuthMenu = () => {
  const permissionStore = usePermissionStore(store)
  const routers = permissionStore.getRouters
  const nodePathArray = []
  getPathway(routers, nodePathArray)
  if (nodePathArray.length) {
    nodePathArray.reverse()
    return nodePathArray.join('/')
  }
  return null
}

const getPathway = (tree, nodePathArray) => {
  for (let index = 0; index < tree.length; index++) {
    if (tree[index].children) {
      const endRecursiveLoop = getPathway(tree[index].children, nodePathArray)
      if (endRecursiveLoop) {
        nodePathArray.push(tree[index].path)
        return true
      }
    }
    if (!tree[index].children?.length && !tree[index].hidden) {
      nodePathArray.push(tree[index].path)
      return true
    }
  }
}
