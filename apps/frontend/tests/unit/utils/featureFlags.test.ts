import { afterEach, describe, expect, it } from 'vitest'
import {
  DYNAMIC_NAVIGATION_FLAG_KEY,
  isDynamicNavigationEnabled
} from '@/utils/featureFlags'

describe('featureFlags', () => {
  afterEach(() => {
    window.localStorage.removeItem(DYNAMIC_NAVIGATION_FLAG_KEY)
  })

  it('should enable dynamic navigation by default', () => {
    expect(isDynamicNavigationEnabled()).toBe(true)
  })

  it('should disable dynamic navigation for false-like values', () => {
    window.localStorage.setItem(DYNAMIC_NAVIGATION_FLAG_KEY, 'false')
    expect(isDynamicNavigationEnabled()).toBe(false)

    window.localStorage.setItem(DYNAMIC_NAVIGATION_FLAG_KEY, '0')
    expect(isDynamicNavigationEnabled()).toBe(false)
  })

  it('should keep dynamic navigation enabled for truthy values', () => {
    window.localStorage.setItem(DYNAMIC_NAVIGATION_FLAG_KEY, 'true')
    expect(isDynamicNavigationEnabled()).toBe(true)

    window.localStorage.setItem(DYNAMIC_NAVIGATION_FLAG_KEY, 'on')
    expect(isDynamicNavigationEnabled()).toBe(true)
  })
})
