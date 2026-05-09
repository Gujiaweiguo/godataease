import { describe, it, expect, vi, beforeEach } from 'vitest'

const mocks = vi.hoisted(() => ({
  emitterEmit: vi.fn(),
  performLogout: vi.fn()
}))

vi.mock('@/hooks/web/useEmitt', () => ({
  useEmitt: () => ({
    emitter: {
      emit: mocks.emitterEmit
    }
  })
}))

vi.mock('@/utils/logout', () => ({
  performLogout: mocks.performLogout
}))

import {
  executeMenuAction,
  isEventMenu,
  isExternalMenu,
  isRouteMenu
} from '@/utils/menu-actions'

describe('menu-actions', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('executeMenuAction', () => {
    it('should emit open-about-dialog event', () => {
      executeMenuAction('open-about-dialog')
      expect(mocks.emitterEmit).toHaveBeenCalledWith('open-about-dialog')
    })

    it('should emit data-export-center event', () => {
      executeMenuAction('data-export-center')
      expect(mocks.emitterEmit).toHaveBeenCalledWith('data-export-center')
    })

    it('should call performLogout for user-logout action', () => {
      executeMenuAction('user-logout')
      expect(mocks.performLogout).toHaveBeenCalled()
    })

    it('should handle open-language-selector without error', () => {
      expect(() => executeMenuAction('open-language-selector')).not.toThrow()
      expect(mocks.emitterEmit).not.toHaveBeenCalled()
      expect(mocks.performLogout).not.toHaveBeenCalled()
    })

    it('should log warning for unknown action', () => {
      const warnSpy = vi.spyOn(console, 'warn').mockImplementation(() => {})
      executeMenuAction('unknown-action')
      expect(warnSpy).toHaveBeenCalledWith('Unknown menu action: unknown-action')
      warnSpy.mockRestore()
    })

    it('should not throw for unknown action', () => {
      vi.spyOn(console, 'warn').mockImplementation(() => {})
      expect(() => executeMenuAction('nonexistent')).not.toThrow()
      vi.restoreAllMocks()
    })
  })

  describe('isEventMenu', () => {
    it('should return true for event type', () => {
      expect(isEventMenu('event')).toBe(true)
    })

    it('should return false for non-event type', () => {
      expect(isEventMenu('route')).toBe(false)
    })

    it('should return false for undefined', () => {
      expect(isEventMenu(undefined)).toBe(false)
    })
  })

  describe('isExternalMenu', () => {
    it('should return true for external type', () => {
      expect(isExternalMenu('external')).toBe(true)
    })

    it('should return false for non-external type', () => {
      expect(isExternalMenu('route')).toBe(false)
    })

    it('should return false for undefined', () => {
      expect(isExternalMenu(undefined)).toBe(false)
    })
  })

  describe('isRouteMenu', () => {
    it('should return true for route type', () => {
      expect(isRouteMenu('route')).toBe(true)
    })

    it('should return true for undefined', () => {
      expect(isRouteMenu(undefined)).toBe(true)
    })

    it('should return false for other types', () => {
      expect(isRouteMenu('event')).toBe(false)
      expect(isRouteMenu('external')).toBe(false)
    })
  })
})
