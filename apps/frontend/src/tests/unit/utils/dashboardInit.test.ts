import { describe, expect, it } from 'vitest'

import {
  createAsyncLoadGate,
  shouldInitializeCreateRoute,
  shouldInitializeDashboardCreate
} from '@/utils/dashboardInit'

describe('dashboardInit utilities', () => {
  describe('shouldInitializeCreateRoute', () => {
    it('returns true for the explicit create flow', () => {
      expect(shouldInitializeCreateRoute(undefined, 'create')).toBe(true)
      expect(shouldInitializeCreateRoute(null, 'create')).toBe(true)
    })

    it('returns true for bare routes without resourceId or opt', () => {
      expect(shouldInitializeCreateRoute(undefined, undefined)).toBe(true)
      expect(shouldInitializeCreateRoute(null, null)).toBe(true)
      expect(shouldInitializeCreateRoute('', '')).toBe(true)
    })

    it('returns false for edit flow', () => {
      expect(shouldInitializeCreateRoute('2', undefined)).toBe(false)
      expect(shouldInitializeCreateRoute(2, null)).toBe(false)
    })

    it('returns false for other opt modes without a resource id', () => {
      expect(shouldInitializeCreateRoute(undefined, 'copy')).toBe(false)
      expect(shouldInitializeCreateRoute(null, 'preview')).toBe(false)
    })
  })

  describe('shouldInitializeDashboardCreate', () => {
    it('reuses the shared create route rule', () => {
      expect(shouldInitializeDashboardCreate(undefined, 'create')).toBe(
        shouldInitializeCreateRoute(undefined, 'create')
      )
      expect(shouldInitializeDashboardCreate(undefined, undefined)).toBe(
        shouldInitializeCreateRoute(undefined, undefined)
      )
      expect(shouldInitializeDashboardCreate('2', undefined)).toBe(
        shouldInitializeCreateRoute('2', undefined)
      )
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
