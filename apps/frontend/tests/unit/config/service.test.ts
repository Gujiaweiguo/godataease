import { beforeAll, beforeEach, describe, expect, it, vi } from 'vitest'

const mocks = vi.hoisted(() => {
  const requestInterceptor = {
    onFulfilled: undefined as any,
    onRejected: undefined as any
  }
  const responseInterceptor = {
    onFulfilled: undefined as any,
    onRejected: undefined as any
  }

  const axiosInstance = {
    interceptors: {
      request: {
        use: vi.fn((onFulfilled, onRejected) => {
          requestInterceptor.onFulfilled = onFulfilled
          requestInterceptor.onRejected = onRejected
        })
      },
      response: {
        use: vi.fn((onFulfilled, onRejected) => {
          responseInterceptor.onFulfilled = onFulfilled
          responseInterceptor.onRejected = onRejected
        })
      }
    }
  }

  const cacheState: Record<string, any> = {
    'user.token': 'token-123',
    'de-platform-client': false
  }

  return {
    requestInterceptor,
    responseInterceptor,
    axiosCreateMock: vi.fn(() => axiosInstance),
    cancelTokenMock: vi.fn(function CancelToken(this: any, executor) {
      const cancel = vi.fn()
      executor(cancel)
      return { cancel }
    }),
    clearCacheMock: vi.fn(),
    tryShowLoadingMock: vi.fn(),
    tryHideLoadingMock: vi.fn(),
    elMessageMock: vi.fn(),
    elMessageErrorMock: vi.fn(),
    confirmMock: vi.fn(() => Promise.resolve()),
    routerPushMock: vi.fn(),
    permissionStore: { getCurrentPath: '/workbranch' },
    linkStore: { getLinkToken: '', setLinkToken: vi.fn() },
    embeddedStore: { baseUrl: '', token: '' },
    requestStore: { resetLoadingMap: vi.fn() },
    configHandlerMock: vi.fn(config => config),
    isMobileMock: vi.fn(() => false),
    getLocaleMock: vi.fn(() => 'en'),
    cacheState
  }
})

class MockXMLHttpRequest {
  readyState = 0
  status = 200
  responseText = JSON.stringify({ code: '000000', data: 100 })
  onreadystatechange: (() => void) | null = null

  open() {}

  send() {
    this.readyState = 4
    this.onreadystatechange?.()
  }
}

vi.stubGlobal('XMLHttpRequest', MockXMLHttpRequest as any)

vi.mock('axios', () => ({
  default: {
    create: mocks.axiosCreateMock,
    CancelToken: mocks.cancelTokenMock
  }
}))

vi.mock('../../../src/utils/loading', () => ({
  tryShowLoading: mocks.tryShowLoadingMock,
  tryHideLoading: mocks.tryHideLoadingMock
}))

vi.mock('../../../src/store/modules/permission', () => ({
  usePermissionStoreWithOut: () => mocks.permissionStore
}))

vi.mock('../../../src/store/modules/embedded', () => ({
  useEmbedded: () => mocks.embeddedStore
}))

vi.mock('../../../src/store/modules/link', () => ({
  useLinkStoreWithOut: () => mocks.linkStore
}))

vi.mock('../../../src/config/axios/refresh', () => ({
  configHandler: mocks.configHandlerMock
}))

vi.mock('../../../src/utils/utils', () => ({
  isMobile: mocks.isMobileMock,
  getLocale: mocks.getLocaleMock
}))

vi.mock('../../../src/store/modules/request', () => ({
  useRequestStoreWithOut: () => mocks.requestStore
}))

vi.mock('../../../src/utils/cacheUtil', () => ({
  clearCache: mocks.clearCacheMock
}))

vi.mock('element-plus-secondary', () => ({
  ElMessage: Object.assign(mocks.elMessageMock, { error: mocks.elMessageErrorMock }),
  ElMessageBox: {
    confirm: mocks.confirmMock
  }
}))

vi.mock('../../../src/router', () => ({
  default: {
    currentRoute: { value: { fullPath: '/workbranch' } },
    push: mocks.routerPushMock
  }
}))

vi.mock('../../../src/hooks/web/useCache', () => ({
  useCache: () => ({
    wsCache: {
      get: (key: string) => mocks.cacheState[key],
      set: (key: string, value: any) => {
        mocks.cacheState[key] = value
      }
    }
  })
}))

describe('axios service response interceptor', () => {
  beforeAll(async () => {
    vi.stubEnv('VITE_API_BASEPATH', '/api')
    await import('../../../src/config/axios/service')
  })

  beforeEach(() => {
    vi.clearAllMocks()
    localStorage.clear()
    mocks.cacheState['user.token'] = 'token-123'
    mocks.cacheState['de-platform-client'] = false
  })

  it('clears cache and redirects on 401 with DE-GATEWAY-FLAG', async () => {
    const error = {
      message: 'Request failed with status code 401',
      config: {
        url: '/typeface/listFont',
        loading: false
      },
      response: {
        status: 401,
        data: { msg: 'missing authorization header' },
        headers: {
          has: (key: string) => key === 'DE-GATEWAY-FLAG',
          get: () => '1'
        }
      }
    }

    await expect(mocks.responseInterceptor.onRejected(error)).rejects.toBe(error)

    expect(mocks.clearCacheMock).toHaveBeenCalledTimes(1)
    expect(localStorage.getItem('DE-GATEWAY-FLAG')).toBe('1')
    expect(mocks.routerPushMock).toHaveBeenCalledWith('/login?redirect=/workbranch')
    expect(mocks.elMessageMock).not.toHaveBeenCalled()
  })

  it('shows an error message and rejects 401 without gateway flag', async () => {
    const error = {
      message: 'Request failed with status code 401',
      config: {
        url: '/typeface/listFont',
        loading: false
      },
      response: {
        status: 401,
        data: { msg: 'missing authorization header' },
        headers: {
          has: () => false,
          get: () => undefined
        }
      }
    }

    await expect(mocks.responseInterceptor.onRejected(error)).rejects.toBe(error)

    expect(mocks.clearCacheMock).not.toHaveBeenCalled()
    expect(mocks.routerPushMock).not.toHaveBeenCalled()
    expect(mocks.elMessageMock).toHaveBeenCalledWith(
      expect.objectContaining({
        type: 'error',
        message: 'missing authorization header'
      })
    )
  })
})
