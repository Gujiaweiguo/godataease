import { describe, expect, it } from 'vitest'

import { createAsyncLoadGate, shouldInitializeDashboardCreate } from '@/utils/dashboardInit'

describe('dashboardInit utilities', () => {
  describe('shouldInitializeDashboardCreate', () => {
    it('returns true for the explicit create flow', () => {
      expect(shouldInitializeDashboardCreate(undefined, 'create')).toBe(true)
      expect(shouldInitializeDashboardCreate(null, 'create')).toBe(true)
    })

    it('returns true for bare dashboard routes without resourceId or opt', () => {
      expect(shouldInitializeDashboardCreate(undefined, undefined)).toBe(true)
      expect(shouldInitializeDashboardCreate(null, null)).toBe(true)
      expect(shouldInitializeDashboardCreate('', '')).toBe(true)
    })

    it('returns false for existing dashboard edit flow', () => {
      expect(shouldInitializeDashboardCreate('2', undefined)).toBe(false)
      expect(shouldInitializeDashboardCreate(2, null)).toBe(false)
    })

    it('returns false for other opt modes without a resource id', () => {
      expect(shouldInitializeDashboardCreate(undefined, 'copy')).toBe(false)
      expect(shouldInitializeDashboardCreate(null, 'preview')).toBe(false)
    })
  })

  describe('createAsyncLoadGate', () => {
    it('resolves waiting consumers after markLoaded', async () => {
      const gate = createAsyncLoadGate()

      expect(gate.isLoaded()).toBe(false)

      gate.markLoaded()
      await expect(gate.wait).resolves.toBeUndefined()

      expect(gate.isLoaded()).toBe(true)
    })

    it('does not lose an early loaded signal before awaiting', async () => {
      const gate = createAsyncLoadGate()

      gate.markLoaded()

      await expect(gate.wait).resolves.toBeUndefined()
      expect(gate.isLoaded()).toBe(true)
    })

    it('is idempotent when markLoaded is called multiple times', async () => {
      const gate = createAsyncLoadGate()

      gate.markLoaded()
      gate.markLoaded()

      await expect(gate.wait).resolves.toBeUndefined()
      expect(gate.isLoaded()).toBe(true)
    })
  })
})
