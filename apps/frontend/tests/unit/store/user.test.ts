import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { userStore } from '@/store/modules/user'

const { wsCacheMock, changeLocaleMock, setLangMock } = vi.hoisted(() => ({
  wsCacheMock: { get: vi.fn(), set: vi.fn(), delete: vi.fn() },
  changeLocaleMock: vi.fn().mockResolvedValue(undefined),
  setLangMock: vi.fn()
}))

vi.mock('@/hooks/web/useCache', () => ({
  useCache: () => ({ wsCache: wsCacheMock })
}))

vi.mock('@/hooks/web/useLocale', () => ({
  useLocale: () => ({ changeLocale: changeLocaleMock })
}))

vi.mock('@/store/modules/locale', () => ({
  useLocaleStoreWithOut: () => ({ setLang: setLangMock })
}))

describe('User Store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  describe('Initial State', () => {
    it('should have null token initially', () => {
      const store = userStore()
      expect(store.getToken).toBeNull()
    })

    it('should have null uid initially', () => {
      const store = userStore()
      expect(store.getUid).toBeNull()
    })

    it('should have null name initially', () => {
      const store = userStore()
      expect(store.getName).toBeNull()
    })

    it('should have null oid initially', () => {
      const store = userStore()
      expect(store.getOid).toBeNull()
    })

    it('should have zh-CN language initially', () => {
      const store = userStore()
      expect(store.getLanguage).toBe('zh-CN')
    })

    it('should have null exp initially', () => {
      const store = userStore()
      expect(store.getExp).toBeNull()
    })

    it('should have null time initially', () => {
      const store = userStore()
      expect(store.getTime).toBeNull()
    })

    it('should have null currentOrg initially', () => {
      const store = userStore()
      expect(store.getCurrentOrg).toBeNull()
    })

    it('should have empty availableOrgs initially', () => {
      const store = userStore()
      expect(store.getAvailableOrgs).toEqual([])
    })
  })

  describe('setToken', () => {
    it('should set token in state and wsCache', () => {
      const store = userStore()
      store.setToken('test-token-123')
      expect(store.getToken).toBe('test-token-123')
      expect(wsCacheMock.set).toHaveBeenCalledWith('user.token', 'test-token-123')
    })
  })

  describe('setExp', () => {
    it('should set exp in state and wsCache', () => {
      const store = userStore()
      store.setExp(1700000000)
      expect(store.getExp).toBe(1700000000)
      expect(wsCacheMock.set).toHaveBeenCalledWith('user.exp', 1700000000)
    })
  })

  describe('setTime', () => {
    it('should set time in state and wsCache', () => {
      const store = userStore()
      store.setTime(1700000000)
      expect(store.getTime).toBe(1700000000)
      expect(wsCacheMock.set).toHaveBeenCalledWith('user.time', 1700000000)
    })
  })

  describe('setUid', () => {
    it('should set uid in state and wsCache', () => {
      const store = userStore()
      store.setUid('user-001')
      expect(store.getUid).toBe('user-001')
      expect(wsCacheMock.set).toHaveBeenCalledWith('user.uid', 'user-001')
    })
  })

  describe('setName', () => {
    it('should set name in state and wsCache', () => {
      const store = userStore()
      store.setName('Admin User')
      expect(store.getName).toBe('Admin User')
      expect(wsCacheMock.set).toHaveBeenCalledWith('user.name', 'Admin User')
    })
  })

  describe('setOid', () => {
    it('should set oid in state and wsCache', () => {
      const store = userStore()
      store.setOid(42)
      expect(store.getOid).toBe(42)
      expect(wsCacheMock.set).toHaveBeenCalledWith('user.oid', 42)
    })

    it('should set oid to null', () => {
      const store = userStore()
      store.setOid(42)
      store.setOid(null)
      expect(store.getOid).toBeNull()
      expect(wsCacheMock.set).toHaveBeenCalledWith('user.oid', null)
    })
  })

  describe('setOrgContext', () => {
    it('should set currentOrg and availableOrgs in state and wsCache', () => {
      const store = userStore()
      const org = { orgId: 1, orgName: 'Test Org' }
      const orgs = [
        { orgId: 1, orgName: 'Test Org' },
        { orgId: 2, orgName: 'Other Org' }
      ]
      store.setOrgContext(org, orgs)
      expect(store.getCurrentOrg).toEqual(org)
      expect(store.getAvailableOrgs).toEqual(orgs)
      expect(wsCacheMock.set).toHaveBeenCalledWith('user.currentOrg', org)
      expect(wsCacheMock.set).toHaveBeenCalledWith('user.availableOrgs', orgs)
    })

    it('should call setOid when currentOrg has orgId', () => {
      const store = userStore()
      const org = { orgId: 5, orgName: 'Org Five' }
      store.setOrgContext(org, [])
      expect(store.getOid).toBe(5)
    })

    it('should not call setOid when currentOrg is null', () => {
      const store = userStore()
      store.setOrgContext(null, [])
      expect(store.getOid).toBeNull()
    })

    it('should default availableOrgs to empty array', () => {
      const store = userStore()
      store.setOrgContext(null)
      expect(store.getAvailableOrgs).toEqual([])
    })
  })

  describe('setLanguage', () => {
    it('should set language in state and wsCache', async () => {
      const store = userStore()
      await store.setLanguage('en')
      expect(store.getLanguage).toBe('en')
      expect(wsCacheMock.set).toHaveBeenCalledWith('user.language', 'en')
    })

    it('should normalize zh_CN to zh-CN', async () => {
      const store = userStore()
      await store.setLanguage('zh_CN')
      expect(store.getLanguage).toBe('zh-CN')
    })

    it('should normalize empty string to zh-CN', async () => {
      const store = userStore()
      await store.setLanguage('')
      expect(store.getLanguage).toBe('zh-CN')
    })

    it('should call locale store setLang', async () => {
      const store = userStore()
      await store.setLanguage('en')
      expect(setLangMock).toHaveBeenCalledWith('en')
    })

    it('should call changeLocale', async () => {
      const store = userStore()
      await store.setLanguage('en')
      expect(changeLocaleMock).toHaveBeenCalledWith('en')
    })
  })

  describe('setUser', () => {
    it('should populate state from API response and wsCache', async () => {
      vi.doMock('@/api/user', () => ({
        userInfo: vi.fn().mockResolvedValue({
          data: {
            id: 'uid-100',
            name: 'Admin',
            oid: 1,
            language: 'en',
            currentOrg: { orgId: 1, orgName: 'Default Org' },
            availableOrgs: [{ orgId: 1, orgName: 'Default Org' }]
          }
        })
      }))

      wsCacheMock.get.mockImplementation((key: string) => {
        if (key === 'user.token') return 'cached-token'
        if (key === 'user.exp') return 1700000000
        if (key === 'user.time') return 1700000001
        return undefined
      })

      const store = userStore()
      await store.setUser()

      expect(store.getUid).toBe('uid-100')
      expect(store.getName).toBe('Admin')
      expect(store.getToken).toBe('cached-token')
      expect(store.getExp).toBe(1700000000)
      expect(store.getTime).toBe(1700000001)
      expect(store.getCurrentOrg).toEqual({ orgId: 1, orgName: 'Default Org' })
    })
  })

  describe('clear', () => {
    it('should delete all wsCache keys and reset state', () => {
      const store = userStore()
      store.setToken('some-token')
      store.setName('User')
      store.setOid(1)
      store.setOrgContext({ orgId: 1, orgName: 'Org' }, [{ orgId: 1, orgName: 'Org' }])

      store.clear()

      expect(wsCacheMock.delete).toHaveBeenCalledWith('user.token')
      expect(wsCacheMock.delete).toHaveBeenCalledWith('user.uid')
      expect(wsCacheMock.delete).toHaveBeenCalledWith('user.name')
      expect(wsCacheMock.delete).toHaveBeenCalledWith('user.oid')
      expect(wsCacheMock.delete).toHaveBeenCalledWith('user.language')
      expect(wsCacheMock.delete).toHaveBeenCalledWith('user.exp')
      expect(wsCacheMock.delete).toHaveBeenCalledWith('user.time')
      expect(wsCacheMock.delete).toHaveBeenCalledWith('user.currentOrg')
      expect(wsCacheMock.delete).toHaveBeenCalledWith('user.availableOrgs')
      expect(store.getCurrentOrg).toBeNull()
      expect(store.getAvailableOrgs).toEqual([])
    })
  })
})
