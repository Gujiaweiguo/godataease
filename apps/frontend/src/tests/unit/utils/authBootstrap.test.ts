import { describe, expect, it } from 'vitest'
import { isBootstrapSessionValid } from '@/utils/authBootstrap'

describe('isBootstrapSessionValid', () => {
  it('returns true for desktop mode without token', () => {
    expect(isBootstrapSessionValid({ isDesktop: true })).toBe(true)
  })

  it('returns false when token is missing expiry', () => {
    expect(isBootstrapSessionValid({ token: 'token-only', now: 1000 })).toBe(false)
  })

  it('returns false when token is near expiry without refresh timestamp', () => {
    expect(
      isBootstrapSessionValid({
        token: 'token',
        exp: 10999,
        now: 1000
      })
    ).toBe(false)
  })

  it('returns true when token expiry is still safely in the future', () => {
    expect(
      isBootstrapSessionValid({
        token: 'token',
        exp: 12000,
        now: 1000
      })
    ).toBe(true)
  })

  it('returns true when session age exceeds refresh window but token expiry is still safe', () => {
    expect(
      isBootstrapSessionValid({
        token: 'token',
        exp: 200000,
        time: 1000,
        now: 92001
      })
    ).toBe(true)
  })

  it('returns false when refresh time is recent but token is already near expiry', () => {
    expect(
      isBootstrapSessionValid({
        token: 'token',
        exp: 10999,
        time: 500,
        now: 1000
      })
    ).toBe(false)
  })

  it('returns true when session age is still within refresh window', () => {
    expect(
      isBootstrapSessionValid({
        token: 'token',
        exp: 200000,
        time: 1000,
        now: 91000
      })
    ).toBe(true)
  })
})
