import router from './router'
import { useUserStoreWithOut } from '@/store/modules/user'
import { useAppStoreWithOut } from '@/store/modules/app'
import type { RouteRecordRaw } from 'vue-router_2'
import { getDefaultSettings } from '@/api/common'
import { useNProgress } from '@/hooks/web/useNProgress'
import {
  DYNAMIC_NOT_FOUND_ROUTE_NAME,
  usePermissionStoreWithOut,
  pathValid
} from '@/store/modules/permission'
import { usePageLoading } from '@/hooks/web/usePageLoading'
import { getRoleRouters } from '@/api/common'
import { useCache } from '@/hooks/web/useCache'
import { isMobile, checkPlatform, isLarkPlatform, isPlatformClient } from '@/utils/utils'
import { interactiveStoreWithOut } from '@/store/modules/interactive'
import { useAppearanceStoreWithOut } from '@/store/modules/appearance'
import { useEmbedded } from '@/store/modules/embedded'
import { useLoading } from '@/hooks/web/useLoading'
import { isDynamicNavigationEnabled } from '@/utils/featureFlags'
import { isBootstrapSessionValid } from '@/utils/authBootstrap'
const appearanceStore = useAppearanceStoreWithOut()
const { wsCache } = useCache()
const permissionStore = usePermissionStoreWithOut()
const interactiveStore = interactiveStoreWithOut()
const userStore = useUserStoreWithOut()
const appStore = useAppStoreWithOut()

const { start, done } = useNProgress()
const { open } = useLoading()
const { loadStart, loadDone } = usePageLoading()

const whiteList = ['/login', '/de-link', '/chart-view', '/admin-login', '/401', '/404'] // 不重定向白名单
const embeddedWindowWhiteList = ['/dvCanvas', '/dashboard', '/preview', '/dataset-embedded-form']
const embeddedRouteWhiteList = [
  '/dataset-embedded',
  '/dataset-form',
  '/dataset-embedded-form',
  '/datasource-embedded'
]
const PERMISSION_REFRESH_EVENT = 'de:permission-refresh'
const MIN_PERMISSION_REFRESH_INTERVAL = 3000
let permissionRefreshPromise: Promise<void> | null = null
let lastPermissionRefreshAt = 0

const resolveUnauthorizedTarget = (path: string) => {
  if (whiteList.includes(path) || path.startsWith('/de-link')) {
    return '/404'
  }
  return '/401'
}

type BootstrapErrorResponse = {
  status?: number
  headers?: {
    has?: (key: string) => boolean
    [key: string]: unknown
  }
}

const hasResponseHeader = (headers: BootstrapErrorResponse['headers'], key: string) => {
  if (!headers) {
    return false
  }
  if (typeof headers.has === 'function') {
    return headers.has(key)
  }
  const targetKey = key.toLowerCase()
  return Object.keys(headers).some(headerKey => headerKey.toLowerCase() === targetKey)
}

const isAuthBootstrapError = (error: unknown) => {
  const response = (error as { response?: BootstrapErrorResponse } | undefined)?.response
  if (!response) {
    return false
  }
  return response.status === 401 || hasResponseHeader(response.headers, 'DE-GATEWAY-FLAG')
}

const resetBootstrapAuthState = () => {
  userStore.clear()
  permissionStore.clear()
  interactiveStore.clear()
}

const hasValidBootstrapSession = (isDesktop: boolean) => {
  return isBootstrapSessionValid({
    token: wsCache.get('user.token'),
    exp: wsCache.get('user.exp'),
    time: wsCache.get('user.time'),
    isDesktop
  })
}

const removeDynamicRoutes = () => {
  permissionStore.getAddRouters.forEach(route => {
    const routeName = route.name as string | undefined
    if (routeName && router.hasRoute(routeName)) {
      router.removeRoute(routeName)
    }
  })
  if (router.hasRoute(DYNAMIC_NOT_FOUND_ROUTE_NAME)) {
    router.removeRoute(DYNAMIC_NOT_FOUND_ROUTE_NAME)
  }
}

const loadAuthorizedRoutes = async () => {
  let roleRouters = (await getRoleRouters()) || []
  if (!isDynamicNavigationEnabled() && wsCache.get('app.desktop')) {
    roleRouters = roleRouters.filter(item => item.name !== 'system')
  }
  const routers: AppCustomRouteRecordRaw[] = (roleRouters as AppCustomRouteRecordRaw[]).map(item => ({
    ...item,
    top: true
  }))

  removeDynamicRoutes()
  permissionStore.clear()
  await permissionStore.generateRoutes(routers)
  permissionStore.getAddRouters.forEach(route => {
    router.addRoute(route as unknown as RouteRecordRaw)
  })
  permissionStore.setIsAddRouters(true)
  interactiveStore.clear()
  await interactiveStore.initInteractive(true)
}

const refreshPermissionRoutes = async (force = false) => {
  const isDesktop = !!wsCache.get('app.desktop')
  const hasSession = hasValidBootstrapSession(isDesktop)
  if (!hasSession || (!force && !permissionStore.getIsAddRouters)) {
    if (!hasSession && (wsCache.get('user.token') || isDesktop)) {
      resetBootstrapAuthState()
      const redirectPath = router.currentRoute.value.fullPath || router.currentRoute.value.path || '/workbranch'
      if (router.currentRoute.value.path !== '/login') {
        await router.replace(`/login?redirect=${redirectPath}`)
      }
    }
    return
  }

  const now = Date.now()
  if (!force && now - lastPermissionRefreshAt < MIN_PERMISSION_REFRESH_INTERVAL) {
    return
  }

  if (permissionRefreshPromise) {
    return permissionRefreshPromise
  }

  permissionRefreshPromise = (async () => {
    try {
      await loadAuthorizedRoutes()
      lastPermissionRefreshAt = Date.now()

      const currentPath = router.currentRoute.value.path
      if (
        currentPath &&
        !whiteList.includes(currentPath) &&
        !currentPath.startsWith('/de-link') &&
        !pathValid(currentPath)
      ) {
	      await router.replace(resolveUnauthorizedTarget(currentPath))
      }
    } catch (error) {
      if (isAuthBootstrapError(error)) {
        resetBootstrapAuthState()
        const redirectPath = router.currentRoute.value.fullPath || router.currentRoute.value.path || '/workbranch'
        if (router.currentRoute.value.path !== '/login') {
          await router.replace(`/login?redirect=${redirectPath}`)
        }
        return
      }
      throw error
    }
  })().finally(() => {
    permissionRefreshPromise = null
  })

  return permissionRefreshPromise
}

if (typeof window !== 'undefined') {
  window.addEventListener('focus', () => {
    void refreshPermissionRoutes()
  })
  window.addEventListener(PERMISSION_REFRESH_EVENT, () => {
    void refreshPermissionRoutes(true)
  })
}

router.beforeEach(async (to, from, next) => {
  if (['/chart-view'].includes(to.path) || to.path.startsWith('/de-link/')) {
    open()
  }
  start()
  loadStart()
  const platform = checkPlatform()
  let isDesktop = wsCache.get('app.desktop')
  if (isDesktop === null) {
    await appStore.setAppModel()
    isDesktop = appStore.getDesktop
  }
  const hasCachedSession = !!(wsCache.get('user.token') || isDesktop)
  const hasValidSession = hasValidBootstrapSession(!!isDesktop)
  if (hasCachedSession && !hasValidSession) {
    resetBootstrapAuthState()
  }
  if (isMobile() && !['/chart-view'].includes(to.path)) {
    done()
    loadDone()
    if (to.name === 'link') {
      let linkQuery = ''
      if (Object.keys(to.query)) {
        const tempQuery = Object.keys(to.query)
          .map(key => key + '=' + to.query[key])
          .join('&')
        if (tempQuery) {
          linkQuery = '?' + tempQuery
        }
      }
      let pathname = window.location.pathname
      pathname = pathname.replace('casbi/', '')
      pathname = pathname.replace('oidc/', '')
      pathname = pathname.substring(0, pathname.length - 1)
      const prefix = window.origin + pathname
      let toPath = to.fullPath
      if (toPath.includes('?')) {
        toPath = to.fullPath.substring(0, to.fullPath.lastIndexOf('?'))
      }
      window.location.href = (prefix + '/mobile.html#' + toPath + linkQuery).replace(/\+/g, '%2B')
    } else if (
      wsCache.get('user.token') ||
      isDesktop ||
      (!isPlatformClient() && !isLarkPlatform())
    ) {
      let pathname = window.location.pathname
      pathname = pathname.substring(0, pathname.length - 1)
      let url = window.origin + pathname + '/mobile.html#/index'
      if (location.hash?.startsWith('#/preview')) {
        url = window.origin + pathname + '/mobile.html' + location.hash
      }
      if (window.location.search) {
        url += window.location.search
      }
      window.location.href = url
    }
  }
  await appearanceStore.setAppearance()
  await appearanceStore.setFontList()
  const defaultSort = await getDefaultSettings()
  wsCache.set('TreeSort-backend', defaultSort['basic.defaultSort'] ?? '1')
  wsCache.set('open-backend', defaultSort['basic.defaultOpen'] ?? '0')
  if (hasValidSession && !to.path.startsWith('/de-link/')) {
    try {
      if (!userStore.getUid) {
        await userStore.setUser()
      }
      if (to.path === '/login') {
        next({ path: '/workbranch' })
      } else {
        permissionStore.setCurrentPath(to.path)
        if (permissionStore.getIsAddRouters) {
          let str = ''
          if (((from.query.redirect as string) || '?').split('?')[0] === to.path) {
            str = ((window.location.hash as string) || '?').split('?').reverse()[0]
            if (str.includes('redirect=')) {
              str = ''
            }
          }
          if (str) {
            to.fullPath += '?' + str
            to.query = str.split('&').reduce((pre, itx) => {
              const [key, val] = itx.split('=')
              pre[key] = val
              return pre
            }, {})
          }
          if (!pathValid(to.path) && to.path !== '/404' && !to.path.startsWith('/de-link')) {
	            next({ path: resolveUnauthorizedTarget(to.path), replace: true })
            return
          }
          next()
          return
        }

        await loadAuthorizedRoutes()

        const redirectPath = (from.query.redirect as string) || to.fullPath || to.path
        const redirect = decodeURIComponent(redirectPath)
        const resolvedRedirect = router.resolve(redirect)

        if (!pathValid(to.path) && to.path !== '/404' && !to.path.startsWith('/de-link')) {
	          next({ path: resolveUnauthorizedTarget(to.path), replace: true })
          return
        }
        if (to.path === redirect && resolvedRedirect.matched.length > 0) {
          next()
          return
        }
        next({ path: redirect, replace: true })
      }
    } catch (error) {
      if (isAuthBootstrapError(error)) {
        resetBootstrapAuthState()
        if (to.path === '/login') {
          next()
          return
        }
        next(`/login?redirect=${to.fullPath || to.path}`)
        return
      }
      throw error
    }
  } else {
    const embeddedStore = useEmbedded()
    if (
      embeddedStore.getToken &&
      appStore.getIsIframe &&
      embeddedRouteWhiteList.includes(to.path)
    ) {
      if (to.path.includes('/dataset-form')) {
        next({ path: '/dataset-embedded-form', query: to.query })
        return
      }
      permissionStore.setCurrentPath(to.path)
      next()
    } else if (
      (!platform && embeddedWindowWhiteList.includes(to.path)) ||
      whiteList.includes(to.path) ||
      to.path.startsWith('/de-link/')
    ) {
      await appearanceStore.setFontList()
      permissionStore.setCurrentPath(to.path)
      next()
    } else {
      next(`/login?redirect=${to.fullPath || to.path}`) // 否则全部重定向到登录页
    }
  }
})

router.afterEach(() => {
  done()
  loadDone()
})
