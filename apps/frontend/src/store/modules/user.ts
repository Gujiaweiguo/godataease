import { defineStore } from 'pinia'
import { store } from '../index'
import { useCache } from '@/hooks/web/useCache'
import { useLocaleStoreWithOut } from './locale'
import { useLocale } from '@/hooks/web/useLocale'
const { wsCache } = useCache()
const { changeLocale } = useLocale()

interface UserState {
  token: string | null
  uid: string | number | null
  name: string | null
  oid: string | number | null
  language: string
  exp: number | null
  time: number | null
  currentOrg: OrgSummary | null
  availableOrgs: OrgSummary[]
}

interface OrgSummary {
  orgId: number
  orgName: string
}

export const userStore = defineStore('user', {
  state: (): UserState => {
    return {
      token: null,
      uid: null,
      name: null,
      oid: null,
      language: 'zh-CN',
      exp: null,
      time: null,
      currentOrg: null,
      availableOrgs: []
    }
  },
  getters: {
    getToken(): string | null {
      return this.token
    },
    getUid(): string | number | null {
      return this.uid
    },
    getName(): string | null {
      return this.name
    },
    getOid(): string | number | null {
      return this.oid
    },
    getLanguage(): string {
      return this.language
    },
    getExp(): number | null {
      return this.exp
    },
    getTime(): number | null {
      return this.time
    },
    getCurrentOrg(): OrgSummary | null {
      return this.currentOrg
    },
    getAvailableOrgs(): OrgSummary[] {
      return this.availableOrgs
    }
  },
  actions: {
    async setUser() {
      const user = await import('@/api/user')
      const res = await user.userInfo()
      const data = res.data
      data.token = wsCache.get('user.token')
      data.exp = wsCache.get('user.exp')
      data.time = wsCache.get('user.time')
      const keys: string[] = ['token', 'uid', 'name', 'oid', 'language', 'exp', 'time']

      for (const key of keys) {
        const dkey = key === 'uid' ? 'id' : key
        this[key] = data[dkey]
        wsCache.set('user.' + key, this[key])
      }
      this.setOrgContext(data.currentOrg || null, data.availableOrgs || [])
      this.setLanguage(this.language)
    },
    setToken(token: string) {
      wsCache.set('user.token', token)
      this.token = token
    },
    setExp(exp: number) {
      wsCache.set('user.exp', exp)
      this.exp = exp
    },
    setTime(time: number) {
      wsCache.set('user.time', time)
      this.time = time
    },
    setUid(uid: string) {
      wsCache.set('user.uid', uid)
      this.uid = uid
    },
    setName(name: string) {
      wsCache.set('user.name', name)
      this.name = name
    },
    setOid(oid: string | number | null) {
      wsCache.set('user.oid', oid)
      this.oid = oid
    },
    setOrgContext(currentOrg: OrgSummary | null, availableOrgs: OrgSummary[] = []) {
      wsCache.set('user.currentOrg', currentOrg)
      wsCache.set('user.availableOrgs', availableOrgs)
      this.currentOrg = currentOrg
      this.availableOrgs = availableOrgs
      if (currentOrg?.orgId !== undefined) {
        this.setOid(currentOrg.orgId)
      }
    },
    async setLanguage(language: string) {
      const locale = useLocaleStoreWithOut()
      if (!language || language === 'zh_CN') {
        language = 'zh-CN'
      }
      wsCache.set('user.language', language)
      this.language = language
      locale.setLang(language)
      await changeLocale(language as any)
    },
    clear() {
      const keys: string[] = ['token', 'uid', 'name', 'oid', 'language', 'exp', 'time']
      for (const key of keys) {
        wsCache.delete('user.' + key)
      }
      wsCache.delete('user.currentOrg')
      wsCache.delete('user.availableOrgs')
      this.currentOrg = null
      this.availableOrgs = []
    }
  }
})

export const useUserStoreWithOut = () => {
  return userStore(store)
}
