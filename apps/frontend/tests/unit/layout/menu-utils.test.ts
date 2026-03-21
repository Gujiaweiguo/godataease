import { describe, expect, it, vi } from 'vitest'

vi.mock('@/router/establish', () => ({
  formatRoute: (routes: any[]) => routes
}))
import {
  buildMenuSelectPath,
  resolveActiveTopPath,
  resolveTopMenus,
  resolveSideMenus
} from '../../../src/layout/components/menu-utils'

const topMenus = [
  {
    path: '/workbranch/index',
    hidden: false,
    meta: { hidden: false }
  },
  {
    path: '/system',
    hidden: false,
    meta: { hidden: false },
    children: [
      { path: 'user', hidden: false, meta: { hidden: false } },
      { path: 'menu', hidden: false, meta: { hidden: false } },
      { path: 'hidden-child', hidden: true, meta: { hidden: false } }
    ]
  },
  {
    path: '/toolbox',
    hidden: false,
    meta: { hidden: false },
    children: [{ path: 'log', hidden: false, meta: { hidden: false } }]
  }
] as any

describe('menu-utils', () => {
  it('resolveTopMenus should exclude mine entry while keeping other top menus', () => {
    const routes = [
      { path: '/workbranch', hidden: false, meta: { hidden: false } },
      { path: '/panel', hidden: false, meta: { hidden: false } },
      { path: '/screen', hidden: false, meta: { hidden: false } },
      { path: '/data', hidden: false, meta: { hidden: false } },
      { path: '/system', hidden: false, meta: { hidden: false } },
      { path: '/mine', hidden: false, meta: { hidden: false } },
      { path: '/help', hidden: false, meta: { hidden: false } }
    ] as any

    expect(resolveTopMenus(routes).map(item => item.path)).toEqual([
      '/workbranch',
      '/panel',
      '/screen',
      '/data',
      '/system',
      '/help'
    ])
  })

  it('resolveActiveTopPath should match top menu by current route', () => {
    expect(resolveActiveTopPath('/system/user', topMenus)).toBe('/system')
    expect(resolveActiveTopPath('/workbranch/index', topMenus)).toBe('/workbranch/index')
    expect(resolveActiveTopPath('/unknown/path', topMenus)).toBeNull()
  })

  it('resolveSideMenus should return visible submenu from same top tree', () => {
    const sideMenus = resolveSideMenus('/system/menu', topMenus)
    expect(sideMenus.map(item => item.path)).toEqual(['user', 'menu'])
  })

  it('buildMenuSelectPath should avoid duplicated top segment', () => {
    expect(buildMenuSelectPath('/system', 'user', ['system', 'user'])).toBe('/system/user')
    expect(buildMenuSelectPath('/system', 'user', ['user'])).toBe('/system/user')
    expect(buildMenuSelectPath('/system', '/external/path', ['external'])).toBe('/external/path')
  })
})
