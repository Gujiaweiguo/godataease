import { beforeEach, describe, expect, it, vi } from 'vitest'

const mocks = vi.hoisted(() => {
  const pushMock = vi.fn()
  const clearMock = vi.fn()
  const resetMock = vi.fn()
  const wsDeleteMock = vi.fn()
  const wsGetMock = vi.fn()
  const logoutApiMock = vi.fn()

  return {
    pushMock,
    clearMock,
    resetMock,
    wsDeleteMock,
    wsGetMock,
    logoutApiMock
  }
})

vi.mock('../../../src/api/login', () => ({
  logoutApi: mocks.logoutApiMock
}))

vi.mock('../../../src/router', () => ({
  default: {
    currentRoute: { value: { fullPath: '/mine/about' } },
    push: mocks.pushMock
  }
}))

vi.mock('../../../src/store/modules/user', () => ({
  useUserStoreWithOut: () => ({ clear: mocks.clearMock, $reset: mocks.resetMock })
}))

vi.mock('../../../src/store/modules/permission', () => ({
  usePermissionStoreWithOut: () => ({ clear: mocks.clearMock, $reset: mocks.resetMock })
}))

vi.mock('../../../src/store/modules/interactive', () => ({
  interactiveStoreWithOut: () => ({ clear: mocks.clearMock, $reset: mocks.resetMock })
}))

vi.mock('../../../src/store/modules/locale', () => ({
  useLocaleStoreWithOut: () => ({ $reset: mocks.resetMock })
}))

vi.mock('../../../src/hooks/web/useCache', () => ({
  useCache: () => ({
    wsCache: {
      storage: {},
      delete: mocks.wsDeleteMock,
      get: mocks.wsGetMock
    }
  })
}))

import { logoutApi } from '../../../src/api/login'
import { performLogout } from '../../../src/utils/logout'

describe('logout utils', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mocks.wsGetMock.mockReturnValue(undefined)
  })

  it('should still clear local session and redirect when logout api rejects', async () => {
    mocks.logoutApiMock.mockRejectedValueOnce(new Error('unauthorized'))

    await performLogout()

    expect(logoutApi).toHaveBeenCalled()
    expect(mocks.clearMock).toHaveBeenCalled()
    expect(mocks.resetMock).toHaveBeenCalled()
    expect(mocks.pushMock).toHaveBeenCalledWith('/login?redirect=/workbranch')
  })
})
