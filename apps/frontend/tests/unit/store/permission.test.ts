import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { usePermissionStore, pathValid, getFirstAuthMenu } from '@/store/modules/permission'

vi.mock('@/router', async () => {
  const { createRouterModuleMock } = await import('../helpers')
  return createRouterModuleMock()
})

vi.mock('@/router/establish', async () => {
  const { createRouterEstablishModuleMock } = await import('../helpers')
  return createRouterEstablishModuleMock()
})

vi.mock('@/views/404/index.vue', () => ({
  default: {}
}))

describe('Permission Store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  describe('Initial State', () => {
    it('should have empty routers initially', () => {
      const store = usePermissionStore()
      expect(store.getRouters).toEqual([])
    })

    it('should have empty addRouters initially', () => {
      const store = usePermissionStore()
      expect(store.getAddRouters).toEqual([])
    })

    it('should have isAddRouters false initially', () => {
      const store = usePermissionStore()
      expect(store.getIsAddRouters).toBe(false)
    })

    it('should have empty currentPath initially', () => {
      const store = usePermissionStore()
      expect(store.getCurrentPath).toBe('')
    })
  })

  describe('setCurrentPath', () => {
    it('should set current path', () => {
      const store = usePermissionStore()
      store.setCurrentPath('/dashboard')
      expect(store.getCurrentPath).toBe('/dashboard')
    })

    it('should update current path on multiple calls', () => {
      const store = usePermissionStore()
      store.setCurrentPath('/dashboard')
      store.setCurrentPath('/dataset')
      expect(store.getCurrentPath).toBe('/dataset')
    })
  })

  describe('setIsAddRouters', () => {
    it('should set isAddRouters to true', () => {
      const store = usePermissionStore()
      store.setIsAddRouters(true)
      expect(store.getIsAddRouters).toBe(true)
    })

    it('should set isAddRouters to false', () => {
      const store = usePermissionStore()
      store.setIsAddRouters(true)
      store.setIsAddRouters(false)
      expect(store.getIsAddRouters).toBe(false)
    })
  })

  describe('generateRoutes', () => {
    it('should generate routes from role routers', async () => {
      const store = usePermissionStore()
      const roleRouters = [
        {
          path: '/dashboard',
          name: 'Dashboard',
          component: 'Layout',
          children: [
            {
              path: 'index',
              name: 'DashboardIndex',
              component: 'dashboard/index'
            }
          ]
        }
      ]

      await store.generateRoutes(roleRouters)

      expect(store.getAddRouters.length).toBeGreaterThan(0)
    })

    it('should handle empty role routers', async () => {
      const store = usePermissionStore()
      await store.generateRoutes([])
      expect(store.getAddRouters.length).toBe(1)
    })

    it('should add catch-all route', async () => {
      const store = usePermissionStore()
      await store.generateRoutes([])
      const addRouters = store.getAddRouters
      const catchAll = addRouters.find(r => r.path === '/:catchAll(.*)')
      expect(catchAll).toBeDefined()
    })

    it('should handle nested routes', async () => {
      const store = usePermissionStore()
      const roleRouters = [
        {
          path: '/data',
          name: 'Data',
          component: 'Layout',
          children: [
            {
              path: 'dataset',
              name: 'Dataset',
              children: [
                {
                  path: 'list',
                  name: 'DatasetList'
                }
              ]
            }
          ]
        }
      ]

      await store.generateRoutes(roleRouters)
      expect(store.getRouters.length).toBeGreaterThan(0)
    })
  })

  describe('clear', () => {
    it('should clear all state', async () => {
      const store = usePermissionStore()
      store.setCurrentPath('/dashboard')
      store.setIsAddRouters(true)
      await store.generateRoutes([
        { path: '/test', name: 'Test' }
      ])

      store.clear()

      expect(store.getIsAddRouters).toBe(false)
      expect(store.getCurrentPath).toBe('')
      expect(store.getAddRouters).toEqual([])
    })
  })

  describe('getRoutersNotHidden', () => {
    it('should filter out hidden routers', async () => {
      const store = usePermissionStore()
      const roleRouters = [
        { path: '/visible', name: 'Visible', hidden: false },
        { path: '/hidden', name: 'Hidden', hidden: true }
      ]

      await store.generateRoutes(roleRouters)
      const visibleRouters = store.getRoutersNotHidden

      expect(visibleRouters.some(r => r.path === '/hidden')).toBe(false)
    })
  })
})

describe('pathValid', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('should return false for empty path', () => {
    expect(pathValid('')).toBe(false)
  })

  it('should throw for null path', () => {
    expect(() => pathValid(null as any)).toThrow()
  })

  it('should throw for undefined path', () => {
    expect(() => pathValid(undefined as any)).toThrow()
  })

  it('should return false for non-existent path', () => {
    const store = usePermissionStore()
    store.generateRoutes([
      { path: '/dashboard', name: 'Dashboard' }
    ])

    expect(pathValid('/nonexistent')).toBe(false)
  })

  it('should handle dataset-form path normalization', () => {
    const store = usePermissionStore()
    const roleRouters = [
      {
        path: '/data',
        name: 'Data',
        children: [
          {
            path: 'dataset',
            name: 'Dataset'
          }
        ]
      }
    ]
    store.generateRoutes(roleRouters)

    expect(pathValid('/dataset-form')).toBe(false)
  })

  it('should deny direct access after menu authorization is revoked', async () => {
    const store = usePermissionStore()

    await store.generateRoutes([
      {
        path: '/system',
        name: 'System',
        hidden: false,
        children: [
          {
            path: 'menu',
            name: 'Menu',
            hidden: false
          }
        ]
      }
    ])
    expect(store.getRouters.some(route => route.path === '/system')).toBe(true)

    await store.generateRoutes([
      {
        path: '/system',
        name: 'System',
        hidden: false,
        children: [
          {
            path: 'user',
            name: 'User',
            hidden: false
          }
        ]
      }
    ])
    expect(pathValid('/system/menu')).toBe(false)
  })
})

describe('getFirstAuthMenu', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('should return null for empty routers', async () => {
    const store = usePermissionStore()
    await store.generateRoutes([])
    const firstMenu = getFirstAuthMenu()
    expect(firstMenu).toBeNull()
  })

  it('should return null when all routes are hidden', async () => {
    const store = usePermissionStore()
    const roleRouters = [
      {
        path: '/admin',
        name: 'Admin',
        hidden: true,
        children: [
          {
            path: 'settings',
            name: 'AdminSettings',
            hidden: true
          }
        ]
      }
    ]

    await store.generateRoutes(roleRouters)
    const firstMenu = getFirstAuthMenu()

    expect(firstMenu).toBeNull()
  })
})

describe('Permission Store Integration', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('should handle complete permission flow', async () => {
    const store = usePermissionStore()

    store.setCurrentPath('/initial')
    expect(store.getCurrentPath).toBe('/initial')

    const roleRouters = [
      {
        path: '/dashboard',
        name: 'Dashboard',
        component: 'Layout',
        children: [
          {
            path: 'index',
            name: 'DashboardIndex'
          }
        ]
      },
      {
        path: '/data',
        name: 'Data',
        component: 'Layout',
        children: [
          {
            path: 'dataset',
            name: 'Dataset'
          }
        ]
      }
    ]

    await store.generateRoutes(roleRouters)
    store.setIsAddRouters(true)

    expect(store.getIsAddRouters).toBe(true)
    expect(store.getAddRouters.length).toBeGreaterThan(0)
    expect(store.getRouters.length).toBeGreaterThan(0)

    store.clear()
    expect(store.getIsAddRouters).toBe(false)
    expect(store.getAddRouters).toEqual([])
  })

  it('should handle role router update', async () => {
    const store = usePermissionStore()

    const initialRouters = [
      { path: '/dashboard', name: 'Dashboard' }
    ]
    await store.generateRoutes(initialRouters)
    const initialCount = store.getRouters.length

    const updatedRouters = [
      { path: '/dashboard', name: 'Dashboard' },
      { path: '/data', name: 'Data' }
    ]
    await store.generateRoutes(updatedRouters)

    expect(store.getRouters.length).toBeGreaterThanOrEqual(initialCount)
  })

  it('should sync visible navigation with role-menu authorization updates', async () => {
    const store = usePermissionStore()

    await store.generateRoutes([
      {
        path: '/system',
        name: 'System',
        hidden: false,
        children: [
          { path: 'menu', name: 'Menu', hidden: false },
          { path: 'user', name: 'User', hidden: false }
        ]
      }
    ])
    expect(store.getRoutersNotHidden.some(route => route.path === '/system')).toBe(true)

    await store.generateRoutes([
      {
        path: '/system',
        name: 'System',
        hidden: true,
        children: [
          { path: 'menu', name: 'Menu', hidden: false },
          { path: 'user', name: 'User', hidden: false }
        ]
      }
    ])
    expect(store.getRoutersNotHidden.some(route => route.path === '/system')).toBe(false)
  })
})
