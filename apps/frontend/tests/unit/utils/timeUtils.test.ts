import { describe, it, expect } from 'vitest'
import { getRange, getTimeBegin } from '@/utils/timeUitils'

describe('timeUtils', () => {
  describe('getRange', () => {
    it('should return year range for "year" granularity', () => {
      const result = getRange('2023-06-15', 'year')
      expect(result).toBeInstanceOf(Array)
      expect(result).toHaveLength(2)
      // start = Jan 1 2023 00:00:00
      expect(result[0]).toBe(new Date(2023, 0, 1).getTime())
      // end = Dec 31 2023 23:59:59
      expect(result[1]).toBe(new Date(2023, 11, 31).getTime() + 86400000 - 1000)
    })

    it('should return year range for "y" granularity alias', () => {
      const result = getRange('2023-06-15', 'y')
      expect(result[0]).toBe(new Date(2023, 0, 1).getTime())
    })

    it('should return month range for "month" granularity', () => {
      const result = getRange('2023-03-15', 'month')
      expect(result).toHaveLength(2)
      // start = Mar 1 2023
      expect(result[0]).toBe(new Date(2023, 2, 1).getTime())
      // end = Mar 31 2023 23:59:59
      expect(result[1]).toBe(new Date(2023, 3, 1).getTime() - 1000)
    })

    it('should return month range for "y_M" granularity', () => {
      const result = getRange('2023-03-15', 'y_M')
      expect(result[0]).toBe(new Date(2023, 2, 1).getTime())
    })

    it('should return hour range for "hour" granularity', () => {
      const ts = '2023-06-15 10:30:00'
      const result = getRange(ts, 'hour')
      const expected = +new Date(ts)
      expect(result[0]).toBe(expected)
      expect(result[1]).toBe(expected + 3600000 - 1000)
    })

    it('should return minute range for "minute" granularity', () => {
      const ts = '2023-06-15 10:30:00'
      const result = getRange(ts, 'minute')
      const expected = +new Date(ts)
      expect(result[0]).toBe(expected)
      expect(result[1]).toBe(expected + 60000 - 1000)
    })

    it('should return second range for "y_M_d_H_m_s" granularity', () => {
      const ts = '2023-06-15 10:30:45'
      const result = getRange(ts, 'y_M_d_H_m_s')
      const expected = +new Date(ts)
      expect(result[0]).toBe(expected)
      expect(result[1]).toBe(expected + 999)
    })

    it('should return [start, end] same timestamp for "datetime"', () => {
      const ts = '2023-06-15 10:30:00'
      const result = getRange(ts, 'datetime')
      const expected = +new Date(ts)
      expect(result).toEqual([expected, expected])
    })

    it('should append ":" for "y_M_d_H" granularity before processing', () => {
      const ts = '2023-06-15 10'
      const result = getRange(ts + ':', 'y_M_d_H')
      const expected = +new Date(ts + ':')
      expect(result[0]).toBe(expected)
      expect(result[1]).toBe(expected + 3600000 - 1000)
    })

    it('should return input as-is for invalid date', () => {
      const result = getRange('not-a-date', 'year')
      expect(result).toBe('not-a-date')
    })

    it('should return input as-is for unknown granularity', () => {
      const result = getRange('2023-06-15', 'unknown')
      // valid date, falls through default => returns the processed selectValue string
      expect(result).toBe('2023-06-15')
    })
  })

  describe('getTimeBegin', () => {
    it('should return year range for "year"', () => {
      const result = getTimeBegin('2023-06-15', 'year')
      expect(result[0]).toBe(new Date(2023, 0, 1).getTime())
    })

    it('should return month range for "month"', () => {
      const result = getTimeBegin('2023-03-15', 'month')
      expect(result[0]).toBe(new Date(2023, 2, 1).getTime())
    })

    it('should return day range for "date"', () => {
      const ts = '2023-06-15'
      const result = getTimeBegin(ts, 'date')
      expect(result).toHaveLength(2)
    })

    it('should return input as-is for unknown granularity', () => {
      const result = getTimeBegin('2023-06-15', 'hour')
      expect(result).toBe('2023-06-15')
    })
  })
})
