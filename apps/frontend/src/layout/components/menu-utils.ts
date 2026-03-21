import { formatRoute } from '@/router/establish'

type RouteMetaLike = {
  hidden?: boolean
}

export type LayoutMenuRoute = AppCustomRouteRecordRaw & {
  hidden?: boolean
  meta?: RouteMetaLike
  children?: LayoutMenuRoute[]
}

const isVisible = (route: LayoutMenuRoute) => !route.hidden && !route.meta?.hidden

const toAbsolutePath = (path: string) => {
  if (!path) {
    return '/'
  }
  return path.startsWith('/') ? path : `/${path}`
}

const isAccountTopMenu = (path: string) => {
  const absolutePath = toAbsolutePath(path)
  return absolutePath === '/mine' || absolutePath.startsWith('/mine/')
}

export const resolveTopMenus = (routes: AppCustomRouteRecordRaw[]): LayoutMenuRoute[] => {
  const normalized = formatRoute(routes)
  return normalized.filter(route => !isAccountTopMenu(route.path) && isVisible(route as LayoutMenuRoute)) as LayoutMenuRoute[]
}

export const resolveActiveTopPath = (currentPath: string, topMenus: LayoutMenuRoute[]): string | null => {
  const targetPath = toAbsolutePath(currentPath)
  let matchedPath: string | null = null
  let matchedLength = -1
  topMenus.forEach(menu => {
    const menuPath = toAbsolutePath(menu.path)
    const hit = targetPath === menuPath || targetPath.startsWith(`${menuPath}/`)
    if (hit && menuPath.length > matchedLength) {
      matchedPath = menuPath
      matchedLength = menuPath.length
    }
  })
  return matchedPath
}

export const resolveSideMenus = (
  currentPath: string,
  topMenus: LayoutMenuRoute[]
): LayoutMenuRoute[] => {
  const activeTopPath = resolveActiveTopPath(currentPath, topMenus)
  if (!activeTopPath) {
    return []
  }
  const top = topMenus.find(menu => toAbsolutePath(menu.path) === activeTopPath)
  if (!top?.children?.length) {
    return []
  }
  return top.children.filter(route => isVisible(route as LayoutMenuRoute)) as LayoutMenuRoute[]
}

export const buildMenuSelectPath = (
  topPath: string,
  index: string,
  indexPath: string[]
): string => {
  if (!index) {
    return topPath
  }
  if (index.startsWith('/')) {
    return index
  }
  const baseSegments = toAbsolutePath(topPath)
    .split('/')
    .filter(Boolean)
  const chain = [...indexPath]
  const rootSegment = baseSegments[baseSegments.length - 1]
  if (chain.length && rootSegment && chain[0] === rootSegment) {
    chain.shift()
  }
  const relativePath = chain.join('/')
  return `/${[...baseSegments, relativePath].filter(Boolean).join('/')}`
}
