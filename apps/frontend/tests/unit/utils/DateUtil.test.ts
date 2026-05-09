import { describe, it, expect } from 'vitest'

import '@/utils/DateUtil'

describe('Date.prototype.format', () => {
  it('should format date with default pattern yyyy-MM-dd hh:mm:ss', () => {
    const date = new Date(2024, 5, 15, 10, 30, 45)
    const result = (date as any).format()
    expect(result).toBe('2024-06-15 10:30:45')
  })

  it('should format date with custom pattern yyyy-MM-dd', () => {
    const date = new Date(2024, 0, 1)
    const result = (date as any).format('yyyy-MM-dd')
    expect(result).toBe('2024-01-01')
  })

  it('should format time with pattern hh:mm:ss', () => {
    const date = new Date(2024, 0, 1, 8, 5, 3)
    const result = (date as any).format('hh:mm:ss')
    expect(result).toBe('08:05:03')
  })

  it('should format with short year pattern yy', () => {
    const date = new Date(2024, 0, 1)
    const result = (date as any).format('yy-MM-dd')
    expect(result).toBe('24-01-01')
  })

  it('should pad single-digit month and day with leading zeros', () => {
    const date = new Date(2024, 2, 5)
    const result = (date as any).format('MM-dd')
    expect(result).toBe('03-05')
  })

  it('should format quarter correctly', () => {
    const q1 = new Date(2024, 1, 1)
    expect((q1 as any).format('q')).toBe('1')
    const q2 = new Date(2024, 4, 1)
    expect((q2 as any).format('q')).toBe('2')
    const q3 = new Date(2024, 7, 1)
    expect((q3 as any).format('q')).toBe('3')
    const q4 = new Date(2024, 10, 1)
    expect((q4 as any).format('q')).toBe('4')
  })

  it('should format milliseconds with S', () => {
    const date = new Date(2024, 0, 1, 0, 0, 0, 123)
    const result = (date as any).format('S')
    expect(result).toBe('123')
  })

  it('should handle single-digit seconds without padding', () => {
    const date = new Date(2024, 0, 1, 0, 0, 5)
    const result = (date as any).format('s')
    expect(result).toBe('5')
  })

  it('should handle full datetime with milliseconds', () => {
    const date = new Date(2024, 11, 31, 23, 59, 59, 999)
    const result = (date as any).format('yyyy-MM-dd hh:mm:ss S')
    expect(result).toBe('2024-12-31 23:59:59 999')
  })

  it('should return year when only yyyy is specified', () => {
    const date = new Date(2025, 0, 1)
    const result = (date as any).format('yyyy')
    expect(result).toBe('2025')
  })

  it('should handle midnight correctly', () => {
    const date = new Date(2024, 5, 15, 0, 0, 0)
    const result = (date as any).format('hh:mm:ss')
    expect(result).toBe('00:00:00')
  })

  it('should format double-digit month and day without extra padding', () => {
    const date = new Date(2024, 11, 25)
    const result = (date as any).format('MM-dd')
    expect(result).toBe('12-25')
  })
})
