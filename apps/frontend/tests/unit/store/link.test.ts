import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useLinkStore } from '@/store/modules/link'

describe('Link Store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  describe('Initial State', () => {
    it('should have empty linkToken', () => {
      const store = useLinkStore()
      expect(store.linkToken).toBe('')
    })

    it('should return empty string from getLinkToken getter', () => {
      const store = useLinkStore()
      expect(store.getLinkToken).toBe('')
    })
  })

  describe('setLinkToken', () => {
    it('should set linkToken to the given value', () => {
      const store = useLinkStore()
      store.setLinkToken('abc123')
      expect(store.linkToken).toBe('abc123')
    })

    it('should return the set value from getter', () => {
      const store = useLinkStore()
      store.setLinkToken('token-xyz')
      expect(store.getLinkToken).toBe('token-xyz')
    })

    it('should replace previous token value', () => {
      const store = useLinkStore()
      store.setLinkToken('first-token')
      store.setLinkToken('second-token')
      expect(store.getLinkToken).toBe('second-token')
    })

    it('should allow setting to empty string', () => {
      const store = useLinkStore()
      store.setLinkToken('some-token')
      store.setLinkToken('')
      expect(store.getLinkToken).toBe('')
    })
  })
})
