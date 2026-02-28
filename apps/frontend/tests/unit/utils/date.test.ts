import { describe, it, expect } from 'vitest'

// Import DateUtil to register format on Date.prototype
import '@/utils/DateUtil'

describe('DateUtil', () => {
  describe('format', () => {
    it('should format date with default pattern', () => {
      const date = new Date('2024-01-01')
      const formatted = date.format('yyyy-MM-dd')
      expect(formatted).toBe('2024-01-01')
    })

    it('should format date with custom pattern', () => {
      const date = new Date('2024-01-01')
      const formatted = date.format('yyyy/MM/dd')
      expect(formatted).toBe('2024/01/01')
    })

    it('should format date with full ISO format', () => {
      const date = new Date('2024-01-01')
      const formatted = date.format('yyyy-MM-dd hh:mm:ss')
      expect(formatted).toContain('2024-01-01')
    })

    it('should handle month format', () => {
      const date = new Date('2024-03-15')
      const formatted = date.format('MM/dd')
      expect(formatted).toBe('03/15')
    })

    it('should handle year format', () => {
      const date = new Date('2024-01-01')
      const formatted = date.format('yyyy')
      expect(formatted).toBe('2024')
    })
  })
})
