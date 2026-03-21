import { useUserStoreWithOut } from '@/store/modules/user'
import router from '@/router'
import { usePermissionStoreWithOut } from '@/store/modules/permission'
import { interactiveStoreWithOut } from '@/store/modules/interactive'
import { useLocaleStoreWithOut } from '@/store/modules/locale'
import { useCache } from '@/hooks/web/useCache'
import { logoutApi } from '@/api/login'

const { wsCache } = useCache()
const permissionStore = usePermissionStoreWithOut()
const userStore = useUserStoreWithOut()
const interactiveStore = interactiveStoreWithOut()
const localeStore = useLocaleStoreWithOut()

export const logoutHandler = (justClean?: boolean, save_platform_status = false) => {
  userStore.clear()
  userStore.$reset()
  permissionStore.clear()
  permissionStore.$reset()
  interactiveStore.clear()
  interactiveStore.$reset()
  localeStore.$reset()
  removeCache()
  const currentFullPath = (router.currentRoute.value.fullPath || '/workbranch') as string
  let queryRedirectPath = resolveLogoutRedirectTarget(currentFullPath)
  let pathname = window.location.pathname
  if (pathname) {
    if (pathname.includes('oidcbi/')) {
      if (save_platform_status) {
        return
      }
      pathname = pathname.replace('oidcbi/', '')
      if (pathname.includes('mobile.html')) {
        pathname = pathname.replace('mobile.html', '')
      }
      pathname = pathname.substring(0, pathname.length - 1)
      window.location.href = pathname + '/oidcbi/oidc/logout'
      return
    } else if (pathname.includes('casbi/')) {
      if (save_platform_status) {
        return
      }
      pathname = pathname.replace('casbi/', '')
      if (pathname.includes('mobile.html')) {
        pathname = pathname.replace('mobile.html', '')
      }
      pathname = pathname.substring(0, pathname.length - 1)
      const uri = window.location.href
      window.location.href = pathname + '/casbi/cas/logout?service=' + uri
      return
    }
    pathname = pathname.substring(0, pathname.length - 1)
  }
  if (wsCache.get('custom_auth_logout_url')) {
    window.location.href = wsCache.get('custom_auth_logout_url')
  }
  router.push(justClean ? queryRedirectPath : `/login?redirect=${queryRedirectPath}`)
}

export const performLogout = async (justClean?: boolean, save_platform_status = false) => {
  try {
    if (!justClean) {
      await logoutApi()
    }
  } catch {
  } finally {
    logoutHandler(justClean, save_platform_status)
  }
}

const resolveLogoutRedirectTarget = (currentFullPath: string) => {
  const path = (currentFullPath || '/workbranch').split('?')[0]
  const disallowedPrefixes = ['/mine', '/help', '/401', '/404', '/login']
  if (disallowedPrefixes.some(prefix => path === prefix || path.startsWith(prefix + '/'))) {
    return '/workbranch'
  }
  return currentFullPath || '/workbranch'
}

const removeCache = () => {
  const keys = Object.keys(wsCache['storage'])
  keys.forEach(key => {
    if (
      key.startsWith('de-plugin-') ||
      key === 'de-platform-client' ||
      key === 'pwd-validity-period' ||
      key === 'xpack-model-distributed'
    ) {
      wsCache.delete(key)
    }
  })
}
