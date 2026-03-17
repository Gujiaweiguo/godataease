import { describe, expect, it, vi } from 'vitest'

vi.mock('@/components/plugin', () => ({
  XpackComponent: 'XpackComponent'
}))

import { generateRoutesFn2, resolveViewComponent } from '@/router/establish'

const mockModules: Record<string, unknown> = {
  '../views/audit/dashboard.vue': 'AuditDashboard',
  '../views/audit/settings.vue': 'AuditSettings',
  '../views/system/menu/index.vue': 'SystemMenu',
  '../views/system/user/index.vue': 'SystemUser',
  '../views/dashboard/index.vue': 'Dashboard'
}

describe('resolveViewComponent', () => {
  it('优先解析扁平 vue 文件路径', () => {
    expect(resolveViewComponent('audit/dashboard', mockModules)).toBe('AuditDashboard')
    expect(resolveViewComponent('audit/settings', mockModules)).toBe('AuditSettings')
  })

  it('兼容 component 以 /index 结尾的菜单配置', () => {
    expect(resolveViewComponent('system/user/index', mockModules)).toBe('SystemUser')
  })

  it('在扁平文件不存在时回退到目录 index.vue', () => {
    expect(resolveViewComponent('system/menu', mockModules)).toBe('SystemMenu')
    expect(resolveViewComponent('dashboard', mockModules)).toBe('Dashboard')
  })

  it('找不到组件时返回 null', () => {
    expect(resolveViewComponent('nonexistent/path', mockModules)).toBeNull()
  })
})

describe('generateRoutesFn2', () => {
  it('跳过声明了 component 但无法解析的叶子路由', () => {
    const routes = generateRoutesFn2([
      {
        path: '/broken',
        name: 'Broken',
        hidden: false,
        component: 'nonexistent/path'
      }
    ])

    expect(routes).toEqual([])
  })

  it('保留无 component 的目录路由及其可用子路由', () => {
    const routes = generateRoutesFn2([
      {
        path: '/system',
        name: 'System',
        hidden: false,
        children: [
          {
            path: 'menu',
            name: 'Menu',
            hidden: false,
            component: 'system/menu'
          }
        ]
      }
    ])

    expect(routes).toHaveLength(1)
    expect(routes[0].path).toBe('/system')
    expect(routes[0].children).toHaveLength(1)
  })

  it('过滤掉目录下无法解析的坏子路由', () => {
    const routes = generateRoutesFn2([
      {
        path: '/system',
        name: 'System',
        hidden: false,
        children: [
          {
            path: 'missing',
            name: 'Missing',
            hidden: false,
            component: 'nonexistent/path'
          },
          {
            path: 'menu',
            name: 'Menu',
            hidden: false,
            component: 'system/menu'
          }
        ]
      }
    ])

    expect(routes[0].children?.map(route => route.path)).toEqual(['menu'])
  })
})
