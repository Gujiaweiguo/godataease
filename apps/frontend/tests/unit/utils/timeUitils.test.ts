import { describe, expect, it } from 'vitest'

import { getRange, getTimeBegin } from '@/utils/timeUitils'

describe('timeUitils', () => {
  describe('getRange', () => {
    it('returns input value for invalid date', () => {
      expect(getRange('not-a-date', 'year')).toBe('not-a-date')
    })

    it('returns input value for unknown granularity', () => {
      expect(getRange('2024-06-15', 'unknown')).toBe('2024-06-15')
    })

    it('returns year range [Jan 1, Dec 31 23:59:59] for granularity "year"', () => {
      const result = getRange('2024-06-15', 'year') as number[]
      const start = new Date(result[0])
      const end = new Date(result[1])
      expect(start.getFullYear()).toBe(2024)
      expect(start.getMonth()).toBe(0)
      expect(start.getDate()).toBe(1)
      expect(start.getHours()).toBe(0)
      expect(end.getFullYear()).toBe(2024)
      expect(end.getMonth()).toBe(11)
      expect(end.getDate()).toBe(31)
      expect(end.getHours()).toBe(23)
      expect(end.getMinutes()).toBe(59)
      expect(end.getSeconds()).toBe(59)
    })

    it('returns year range for alias "y"', () => {
      const result = getRange('2023-03-10', 'y') as number[]
      expect(new Date(result[0]).getFullYear()).toBe(2023)
    })

    it('returns month range for granularity "month"', () => {
      const result = getRange('2024-06-15', 'month') as number[]
      const start = new Date(result[0])
      const end = new Date(result[1])
      expect(start.getMonth()).toBe(5)
      expect(start.getDate()).toBe(1)
      expect(end.getMonth()).toBe(5)
      expect(end.getDate()).toBe(30)
      expect(end.getHours()).toBe(23)
    })

    it('returns month range for alias "y_M"', () => {
      const result = getRange('2024-01-20', 'y_M') as number[]
      const start = new Date(result[0])
      expect(start.getMonth()).toBe(0)
      expect(start.getDate()).toBe(1)
    })

    it('returns day range for granularity "date"', () => {
      const result = getRange('2024-06-15T10:30:00', 'date') as number[]
      expect(Array.isArray(result)).toBe(true)
      expect(result).toHaveLength(2)
      expect(result[1] - result[0]).toBe(60 * 1000 * 60 * 24 - 1000)
    })

    it('returns hour range for granularity "hour"', () => {
      const result = getRange('2024-06-15T10:30:00', 'hour') as number[]
      expect(result[1] - result[0]).toBe(60 * 60 * 1000 - 1000)
    })

    it('returns minute range for granularity "minute"', () => {
      const result = getRange('2024-06-15T10:30:00', 'minute') as number[]
      expect(result[1] - result[0]).toBe(60 * 1000 - 1000)
    })

    it('returns second range for granularity "y_M_d_H_m_s"', () => {
      const result = getRange('2024-06-15T10:30:00', 'y_M_d_H_m_s') as number[]
      expect(result[1] - result[0]).toBe(999)
    })

    it('returns [same, same] for datetime granularity', () => {
      const result = getRange('2024-06-15T10:30:00', 'datetime') as number[]
      expect(result[0]).toBe(result[1])
    })

    it('handles y_M_d_H granularity by appending colon to value', () => {
      const result = getRange('2024-06-15 10', 'y_M_d_H') as number[]
      expect(Array.isArray(result)).toBe(true)
    })
  })

  describe('getTimeBegin', () => {
    it('returns year range for "year" granularity', () => {
      const result = getTimeBegin('2024-06-15', 'year') as number[]
      expect(new Date(result[0]).getFullYear()).toBe(2024)
      expect(new Date(result[0]).getMonth()).toBe(0)
    })

    it('returns month range for "month" granularity', () => {
      const result = getTimeBegin('2024-06-15', 'month') as number[]
      expect(new Date(result[0]).getMonth()).toBe(5)
    })

    it('returns day range for "date" granularity', () => {
      const result = getTimeBegin('2024-06-15T10:30:00', 'date') as number[]
      expect(Array.isArray(result)).toBe(true)
      expect(result).toHaveLength(2)
      expect(result[1] - result[0]).toBe(60 * 1000 * 60 * 24 - 1000)
    })

    it('returns input value for unknown granularity', () => {
      expect(getTimeBegin('2024-06-15', 'hour')).toBe('2024-06-15')
    })
  })
})
