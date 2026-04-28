import { beforeAll, beforeEach, describe, expect, it, vi } from 'vitest'

const mocks = vi.hoisted(() => {
  const cacheState: Record<string, any> = {
    'app.desktop': false,
    'user.token': undefined,
    'user.exp': undefined,
    'user.time': undefined
  }

  const routerState = {
    currentRoute: {
      value: {
        fullPath: '/workbranch',
        path: '/workbranch'
      }
    },
    beforeEachHandler: undefined as any
  }

  const routerMock = {
    beforeEach: vi.fn(handler => {
      routerState.beforeEachHandler = handler
    }),
    afterEach: vi.fn(),
    hasRoute: vi.fn(() => false),
    removeRoute: vi.fn(),
    addRoute: vi.fn(),
    replace: vi.fn(),
    push: vi.fn()
  }

  const appearanceStore = {
    setAppearance: vi.fn().mockResolvedValue(undefined),
    setFontList: vi.fn().mockResolvedValue(undefined)
  }

  const permissionStore = {
    getIsAddRouters: true,
    getAddRouters: [],
    getCurrentPath: '/workbranch',
    clear: vi.fn(),
    generateRoutes: vi.fn().mockResolvedValue(undefined),
    setIsAddRouters: vi.fn(),
    setCurrentPath: vi.fn()
  }

  const interactiveStore = {
    clear: vi.fn()
  }

  const userStore = {
    getUid: 1,
    clear: vi.fn(),
    setUser: vi.fn().mockResolvedValue(undefined)
  }

  const appStore = {
    getDesktop: false,
    setAppModel: vi.fn().mockResolvedValue(undefined)
  }

  return {
    cacheState,
    routerState,
    routerMock,
    appearanceStore,
    permissionStore,
    interactiveStore,
    userStore,
    appStore,
    getDefaultSettingsMock: vi.fn().mockResolvedValue({
      'basic.defaultSort': '1',
      'basic.defaultOpen': '0'
    }),
    getRoleRoutersMock: vi.fn().mockResolvedValue([]),
    startMock: vi.fn(),
    doneMock: vi.fn(),
    openMock: vi.fn(),
    loadStartMock: vi.fn(),
    loadDoneMock: vi.fn(),
    isMobileMock: vi.fn(() => false),
    checkPlatformMock: vi.fn(() => false),
    isLarkPlatformMock: vi.fn(() => false),
    isPlatformClientMock: vi.fn(() => false),
    pathValidMock: vi.fn(() => true),
    isDynamicNavigationEnabledMock: vi.fn(() => true),
    isBootstrapSessionValidMock: vi.fn(() => false),
    bootstrapInteractiveInBackgroundMock: vi.fn()
  }
})

vi.mock('../../../src/router', () => ({
  default: {
    ...mocks.routerMock,
    currentRoute: mocks.routerState.currentRoute
  }
}))

vi.mock('../../../src/store/modules/user', () => ({
  useUserStoreWithOut: () => mocks.userStore
}))

vi.mock('../../../src/store/modules/app', () => ({
  useAppStoreWithOut: () => mocks.appStore
}))

vi.mock('../../../src/api/common', () => ({
  getDefaultSettings: mocks.getDefaultSettingsMock,
  getRoleRouters: mocks.getRoleRoutersMock
}))

vi.mock('../../../src/hooks/web/useNProgress', () => ({
  useNProgress: () => ({ start: mocks.startMock, done: mocks.doneMock })
}))

vi.mock('../../../src/store/modules/permission', () => ({
  DYNAMIC_NOT_FOUND_ROUTE_NAME: 'DYNAMIC_NOT_FOUND_ROUTE_NAME',
  usePermissionStoreWithOut: () => mocks.permissionStore,
  pathValid: mocks.pathValidMock
}))

vi.mock('../../../src/hooks/web/usePageLoading', () => ({
  usePageLoading: () => ({ loadStart: mocks.loadStartMock, loadDone: mocks.loadDoneMock })
}))

vi.mock('../../../src/hooks/web/useCache', () => ({
  useCache: () => ({
    wsCache: {
      get: (key: string) => mocks.cacheState[key],
      set: (key: string, value: any) => {
        mocks.cacheState[key] = value
      },
      delete: vi.fn()
    }
  })
}))

vi.mock('../../../src/utils/utils', () => ({
  isMobile: mocks.isMobileMock,
  checkPlatform: mocks.checkPlatformMock,
  isLarkPlatform: mocks.isLarkPlatformMock,
  isPlatformClient: mocks.isPlatformClientMock
}))

vi.mock('../../../src/store/modules/interactive', () => ({
  interactiveStoreWithOut: () => mocks.interactiveStore
}))

vi.mock('../../../src/store/modules/appearance', () => ({
  useAppearanceStoreWithOut: () => mocks.appearanceStore
}))

vi.mock('../../../src/store/modules/embedded', () => ({
  useEmbedded: () => ({ getToken: '', baseUrl: '' })
}))

vi.mock('../../../src/hooks/web/useLoading', () => ({
  useLoading: () => ({ open: mocks.openMock })
}))

vi.mock('../../../src/utils/featureFlags', () => ({
  isDynamicNavigationEnabled: mocks.isDynamicNavigationEnabledMock
}))

vi.mock('../../../src/utils/authBootstrap', () => ({
  isBootstrapSessionValid: mocks.isBootstrapSessionValidMock
}))

vi.mock('../../../src/utils/interactiveBootstrap', () => ({
  bootstrapInteractiveInBackground: mocks.bootstrapInteractiveInBackgroundMock
}))

describe('permission router guard', () => {
  beforeAll(async () => {
    await import('../../../src/permission')
  })

  beforeEach(() => {
    vi.clearAllMocks()
    mocks.cacheState['app.desktop'] = false
    mocks.cacheState['user.token'] = undefined
    mocks.cacheState['user.exp'] = undefined
    mocks.cacheState['user.time'] = undefined
    mocks.permissionStore.getIsAddRouters = true
    mocks.permissionStore.getCurrentPath = '/workbranch'
    mocks.userStore.getUid = 1
    mocks.userStore.setUser.mockResolvedValue(undefined)
    mocks.isBootstrapSessionValidMock.mockReturnValue(false)
    mocks.pathValidMock.mockReturnValue(true)
    mocks.routerState.currentRoute.value = {
      fullPath: '/workbranch',
      path: '/workbranch'
    }
  })

  it('skips protected font bootstrap calls on unauthenticated login route', async () => {
    const next = vi.fn()

    await mocks.routerState.beforeEachHandler(
      { path: '/login', fullPath: '/login', name: 'login', query: {} },
      { path: '/', fullPath: '/', query: {} },
      next
    )

    expect(mocks.appearanceStore.setAppearance).toHaveBeenCalledTimes(1)
    expect(mocks.appearanceStore.setFontList).not.toHaveBeenCalled()
    expect(mocks.getDefaultSettingsMock).not.toHaveBeenCalled()
    expect(mocks.permissionStore.setCurrentPath).toHaveBeenCalledWith('/login')
    expect(next).toHaveBeenCalledWith()
  })

  it('loads protected bootstrap data when session is valid', async () => {
    mocks.cacheState['user.token'] = 'token-123'
    mocks.cacheState['user.exp'] = Date.now() + 60000
    mocks.cacheState['user.time'] = Date.now()
    mocks.isBootstrapSessionValidMock.mockReturnValue(true)

    const next = vi.fn()

    await mocks.routerState.beforeEachHandler(
      { path: '/workbranch', fullPath: '/workbranch', name: 'workbranch', query: {} },
      { path: '/login', fullPath: '/login', query: {} },
      next
    )

    expect(mocks.appearanceStore.setAppearance).toHaveBeenCalledTimes(1)
    expect(mocks.appearanceStore.setFontList).toHaveBeenCalledTimes(1)
    expect(mocks.getDefaultSettingsMock).toHaveBeenCalledTimes(1)
    expect(mocks.permissionStore.setCurrentPath).toHaveBeenCalledWith('/workbranch')
    expect(next).toHaveBeenCalledWith()
  })

  it('redirects to login when bootstrap auth fails after session validation', async () => {
    mocks.cacheState['user.token'] = 'token-123'
    mocks.cacheState['user.exp'] = Date.now() + 60000
    mocks.cacheState['user.time'] = Date.now()
    mocks.isBootstrapSessionValidMock.mockReturnValue(true)
    mocks.userStore.getUid = null
    mocks.userStore.setUser.mockRejectedValue({
      response: {
        status: 401,
        headers: {
          has: (key: string) => key === 'DE-GATEWAY-FLAG'
        }
      }
    })

    const next = vi.fn()

    await mocks.routerState.beforeEachHandler(
      { path: '/workbranch', fullPath: '/workbranch', name: 'workbranch', query: {} },
      { path: '/login', fullPath: '/login', query: {} },
      next
    )

    expect(mocks.userStore.clear).toHaveBeenCalledTimes(1)
    expect(mocks.permissionStore.clear).toHaveBeenCalledTimes(1)
    expect(mocks.interactiveStore.clear).toHaveBeenCalledTimes(1)
    expect(next).toHaveBeenCalledWith('/login?redirect=/workbranch')
  })
})
