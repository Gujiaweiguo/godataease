import { describe, expect, it } from 'vitest'

import {
  createAsyncLoadGate,
  shouldInitializeCreateRoute,
  shouldInitializeDashboardCreate
} from '@/utils/dashboardInit'

describe('dashboardInit', () => {
  describe('shouldInitializeCreateRoute', () => {
    it('returns true when opt is "create"', () => {
      expect(shouldInitializeCreateRoute(undefined, 'create')).toBe(true)
    })

    it('returns true when opt is "create" even with resourceId', () => {
      expect(shouldInitializeCreateRoute('some-id', 'create')).toBe(true)
    })

    it('returns true when both resourceId and opt are missing', () => {
      expect(shouldInitializeCreateRoute()).toBe(true)
    })

    it('returns true when resourceId is undefined and opt is undefined', () => {
      expect(shouldInitializeCreateRoute(undefined, undefined)).toBe(true)
    })

    it('returns false when resourceId exists and opt is not "create"', () => {
      expect(shouldInitializeCreateRoute('abc-123', 'edit')).toBe(false)
    })

    it('returns false when resourceId exists and opt is undefined', () => {
      expect(shouldInitializeCreateRoute('abc-123')).toBe(false)
    })

    it('returns false when opt is a non-"create" string', () => {
      expect(shouldInitializeCreateRoute(undefined, 'edit')).toBe(false)
    })
  })

  describe('shouldInitializeDashboardCreate', () => {
    it('is the same function as shouldInitializeCreateRoute', () => {
      expect(shouldInitializeDashboardCreate).toBe(shouldInitializeCreateRoute)
    })
  })

  describe('createAsyncLoadGate', () => {
    it('starts as not loaded', () => {
      const gate = createAsyncLoadGate()
      expect(gate.isLoaded()).toBe(false)
    })

    it('marks as loaded after markLoaded', () => {
      const gate = createAsyncLoadGate()
      gate.markLoaded()
      expect(gate.isLoaded()).toBe(true)
    })

    it('resolves wait promise after markLoaded', async () => {
      const gate = createAsyncLoadGate()
      gate.markLoaded()
      await expect(gate.wait).resolves.toBeUndefined()
    })

    it('does not change state on double markLoaded', () => {
      const gate = createAsyncLoadGate()
      gate.markLoaded()
      gate.markLoaded()
      expect(gate.isLoaded()).toBe(true)
    })
  })
})
