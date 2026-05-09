import { describe, expect, it } from 'vitest'

import { isBootstrapSessionValid } from '@/utils/authBootstrap'

describe('authBootstrap', () => {
  describe('isBootstrapSessionValid', () => {
    it('returns true when isDesktop is true', () => {
      expect(isBootstrapSessionValid({ token: null, exp: null, isDesktop: true })).toBe(true)
    })

    it('returns true when isDesktop is true even without token/exp', () => {
      expect(isBootstrapSessionValid({ isDesktop: true })).toBe(true)
    })

    it('returns false when token is missing', () => {
      expect(
        isBootstrapSessionValid({ token: null, exp: Date.now() + 60000, isDesktop: false })
      ).toBe(false)
    })

    it('returns false when exp is missing', () => {
      expect(
        isBootstrapSessionValid({ token: 'abc', exp: null, isDesktop: false })
      ).toBe(false)
    })

    it('returns false when both token and exp are missing', () => {
      expect(isBootstrapSessionValid({ isDesktop: false })).toBe(false)
    })

    it('returns false when session is expired', () => {
      const now = 1000000
      const exp = now + 5000
      expect(
        isBootstrapSessionValid({ token: 'abc', exp, isDesktop: false, now })
      ).toBe(false)
    })

    it('returns true when session is valid and not near expiry', () => {
      const now = 1000000
      const exp = now + 60000
      expect(
        isBootstrapSessionValid({ token: 'abc', exp, isDesktop: false, now })
      ).toBe(true)
    })

    it('returns false when exp is exactly expWarningMs away from now', () => {
      const now = 1000000
      const exp = now + 9999
      expect(
        isBootstrapSessionValid({ token: 'abc', exp, isDesktop: false, now })
      ).toBe(false)
    })

    it('returns true when exp is exactly expWarningMs away from now', () => {
      const now = 1000000
      const exp = now + 10000
      expect(
        isBootstrapSessionValid({ token: 'abc', exp, isDesktop: false, now })
      ).toBe(true)
    })

    it('returns true when exp is above expWarningMs from now', () => {
      const now = 1000000
      const exp = now + 10001
      expect(
        isBootstrapSessionValid({ token: 'abc', exp, isDesktop: false, now })
      ).toBe(true)
    })

    it('uses Date.now() as default for now parameter', () => {
      const futureExp = Date.now() + 100000
      expect(
        isBootstrapSessionValid({ token: 'abc', exp: futureExp, isDesktop: false })
      ).toBe(true)
    })
  })
})
