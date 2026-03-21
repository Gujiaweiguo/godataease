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
  '../views/system/role/index.vue': 'SystemRole',
  '../views/system/org/index.vue': 'SystemOrg',
  '../views/dashboard/index.vue': 'Dashboard',
  '../views/data-visualization/index.vue': 'BigScreen',
  '../views/visualized/data/datasource/index.vue': 'DatasourceIndex',
  '../views/visualized/data/dataset/index.vue': 'DatasetIndex',
  '../views/visualized/view/panel/index.vue': 'PanelIndex',
  '../views/visualized/view/screen/index.vue': 'ScreenIndex'
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

  it('应解析核心系统与 BI 业务页组件', () => {
    expect(resolveViewComponent('system/user', mockModules)).toBe('SystemUser')
    expect(resolveViewComponent('system/role', mockModules)).toBe('SystemRole')
    expect(resolveViewComponent('system/org', mockModules)).toBe('SystemOrg')
    expect(resolveViewComponent('visualized/data/datasource', mockModules)).toBe('DatasourceIndex')
    expect(resolveViewComponent('visualized/data/dataset', mockModules)).toBe('DatasetIndex')
    expect(resolveViewComponent('visualized/view/panel', mockModules)).toBe('PanelIndex')
    expect(resolveViewComponent('visualized/view/screen', mockModules)).toBe('ScreenIndex')
    expect(resolveViewComponent('data-visualization', mockModules)).toBe('BigScreen')
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

  it('保留可解析的核心系统与 BI 动态路由', () => {
    const routes = generateRoutesFn2([
      {
        path: '/system',
        name: 'System',
        hidden: false,
        children: [
          { path: 'user', name: 'User', hidden: false, component: 'system/user' },
          { path: 'role', name: 'Role', hidden: false, component: 'system/role' },
          { path: 'org', name: 'Org', hidden: false, component: 'system/org' }
        ]
      },
      {
        path: '/data',
        name: 'Data',
        hidden: false,
        children: [
          {
            path: 'datasource',
            name: 'Datasource',
            hidden: false,
            component: 'visualized/data/datasource'
          },
          {
            path: 'dataset',
            name: 'Dataset',
            hidden: false,
            component: 'visualized/data/dataset'
          }
        ]
      }
    ])

    expect(routes).toHaveLength(2)
    expect(routes[0].children?.map(route => route.path)).toEqual(['user', 'role', 'org'])
    expect(routes[1].children?.map(route => route.path)).toEqual(['datasource', 'dataset'])
  })

  it('归一化带父级前缀的子路由 path', () => {
    const routes = generateRoutesFn2([
      {
        path: '/mine',
        name: 'Mine',
        hidden: false,
        children: [
          {
            path: 'mine/modify-pwd',
            name: 'ModifyPwd',
            hidden: false,
            component: 'system/modify-pwd/index'
          }
        ]
      },
      {
        path: '/help',
        name: 'Help',
        hidden: false,
        children: [
          {
            path: 'help/doc',
            name: 'HelpDoc',
            hidden: false
          }
        ]
      }
    ])

    expect(routes[0].children?.[0].path).toBe('modify-pwd')
    expect(routes[1].children?.[0].path).toBe('doc')
  })
})
