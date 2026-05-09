import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'

const mocks = vi.hoisted(() => ({
  elMessageError: vi.fn(),
  elMessageBoxConfirm: vi.fn(),
  elMessageBoxClose: vi.fn()
}))

vi.mock('element-plus-secondary', () => ({
  ElMessage: {
    error: mocks.elMessageError
  },
  ElMessageBox: {
    confirm: mocks.elMessageBoxConfirm,
    close: mocks.elMessageBoxClose
  }
}))

import { ElMessage, ElMessageBox } from 'element-plus-secondary'
import { check, compareStorage } from '@/utils/CrossPermission'

describe('CrossPermission', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  afterEach(() => {
    // Clean up any window properties set during tests
    const keys = Object.keys(window).filter(k => k.startsWith('cross-panel-'))
    keys.forEach(k => delete window[k])
  })

  describe('check', () => {
    it('should return false and show error when id is not provided', () => {
      const result = check({ someId: 1 })
      expect(result).toBe(false)
      expect(ElMessage.error).toHaveBeenCalledWith('资源ID不能为空')
    })

    it('should return false and show error when id is empty string', () => {
      const result = check({ someId: 1 }, '')
      expect(result).toBe(false)
      expect(ElMessage.error).toHaveBeenCalledWith('资源ID不能为空')
    })

    it('should show confirm dialog when node is not found', () => {
      mocks.elMessageBoxConfirm.mockResolvedValue(undefined)
      const result = check({}, 'panel-123')
      expect(result).toBe(false)
      expect(ElMessageBox.confirm).toHaveBeenCalledWith(
        '无权访问当前资源，是否离开当前页面？系统将不保存您所做的更改',
        expect.objectContaining({
          confirmButtonType: 'primary',
          type: 'warning'
        })
      )
    })

    it('should show confirm dialog when node weight is less than required', () => {
      mocks.elMessageBoxConfirm.mockResolvedValue(undefined)
      const result = check({ 'panel-123': 0 }, 'panel-123', 1)
      expect(result).toBe(false)
      expect(ElMessageBox.confirm).toHaveBeenCalled()
    })

    it('should return true when node weight meets requirement', () => {
      const result = check({ 'panel-123': 1 }, 'panel-123', 1)
      expect(result).toBe(true)
      expect(ElMessageBox.confirm).not.toHaveBeenCalled()
    })

    it('should return true when node weight exceeds requirement', () => {
      const result = check({ 'panel-456': 5 }, 'panel-456', 2)
      expect(result).toBe(true)
    })

    it('should default weight to 1 when not provided', () => {
      const data = { 'panel-789': 1 }
      const result = check(data, 'panel-789')
      expect(result).toBe(true)
    })

    it('should close existing ElMessageBox when cross-panel flag is set and check passes', () => {
      window['cross-panel-test'] = Promise.resolve()
      const result = check({ test: 1 }, 'test', 1)
      expect(result).toBe(true)
      expect(ElMessageBox.close).toHaveBeenCalled()
      expect(window['cross-panel-test']).toBeNull()
    })

    it('should not show confirm dialog again if already showing for same id', () => {
      mocks.elMessageBoxConfirm.mockResolvedValue(undefined)
      // First call sets the flag
      check({}, 'dup-id')
      vi.clearAllMocks()
      // Second call should skip showMsg because flag exists
      check({}, 'dup-id')
      expect(ElMessageBox.confirm).not.toHaveBeenCalled()
    })
  })

  describe('compareStorage', () => {
    it('should return true when both values are equal strings', () => {
      expect(compareStorage('abc', 'abc')).toBe(true)
    })

    it('should return undefined when values differ', () => {
      expect(compareStorage('old', 'new')).toBeUndefined()
    })

    it('should return true when both values are undefined', () => {
      expect(compareStorage(undefined, undefined)).toBe(true)
    })

    it('should return true when both values are the same empty string', () => {
      expect(compareStorage('', '')).toBe(true)
    })

    it('should return undefined when one is undefined and other is a string', () => {
      expect(compareStorage(undefined, 'value')).toBeUndefined()
      expect(compareStorage('value', undefined)).toBeUndefined()
    })
  })
})
