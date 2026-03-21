import { beforeEach, describe, expect, it, vi } from 'vitest'

const mocks = vi.hoisted(() => ({
  cacheState: {
    'app.desktop': true,
    'user.token': 'token-123',
    'user.exp': undefined,
    'user.time': undefined,
    'de-global-refresh': false
  },
  refreshApiMock: vi.fn(),
  addCacheRequestMock: vi.fn(),
  cleanCacheRequestMock: vi.fn(),
  setTokenMock: vi.fn(),
  setExpMock: vi.fn(),
  setTimeMock: vi.fn(),
  isLinkMock: vi.fn()
}))

vi.mock('../../../src/hooks/web/useCache', () => ({
  useCache: () => ({
    wsCache: {
      get: key => mocks.cacheState[key],
      set: (key, value) => {
        mocks.cacheState[key] = value
      }
    }
  })
}))

vi.mock('../../../src/api/login', () => ({
  refreshApi: mocks.refreshApiMock
}))

vi.mock('../../../src/store/modules/user', () => ({
  useUserStoreWithOut: () => ({
    setToken: mocks.setTokenMock,
    setExp: mocks.setExpMock,
    setTime: mocks.setTimeMock
  })
}))

vi.mock('../../../src/store/modules/request', () => ({
  useRequestStoreWithOut: () => ({
    getRequestList: [],
    addCacheRequest: mocks.addCacheRequestMock,
    cleanCacheRequest: mocks.cleanCacheRequestMock
  })
}))

vi.mock('../../../src/utils/utils', () => ({
  isLink: mocks.isLinkMock
}))

import { configHandler } from '../../../src/config/axios/refresh'

describe('axios refresh configHandler', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mocks.cacheState['app.desktop'] = true
    mocks.cacheState['user.token'] = 'token-123'
    mocks.cacheState['user.exp'] = undefined
    mocks.cacheState['user.time'] = undefined
    mocks.cacheState['de-global-refresh'] = false
    mocks.isLinkMock.mockReturnValue(false)
  })

  it('injects token in desktop mode without entering refresh flow', async () => {
    const config = { url: '/user/switchLanguage', headers: {} }

    const result = await configHandler(config)

    expect(result.headers['X-DE-TOKEN']).toBe('token-123')
    expect(mocks.refreshApiMock).not.toHaveBeenCalled()
  })
})
