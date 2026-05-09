import { describe, it, expect, vi } from 'vitest'

vi.mock('@/api/dataset', () => ({
  enumValueObj: vi.fn()
}))

import { filterEnumParams, filterEnumParamsReduce, filterParamsOptions } from '@/utils/componentUtils'

describe('componentUtils', () => {
  describe('filterEnumParams', () => {
    it('should return params unchanged when fieldId is not in filterEnumMap', () => {
      const params = ['value1', 'value2']
      const result = filterEnumParams(params, 'nonexistent-field-id')
      expect(result).toEqual(['value1', 'value2'])
    })

    it('should return params unchanged for empty array', () => {
      const result = filterEnumParams([], 'some-id')
      expect(result).toEqual([])
    })

    it('should return string params unchanged when fieldId not in map', () => {
      const result = filterEnumParams(['hello'], 'missing-id')
      expect(result).toEqual(['hello'])
    })
  })

  describe('filterEnumParamsReduce', () => {
    it('should return params unchanged when fieldId is not in filterEnumMap', () => {
      const params = ['value1', 'value2']
      const result = filterEnumParamsReduce(params, 'nonexistent-field-id')
      expect(result).toEqual(['value1', 'value2'])
    })

    it('should return empty array unchanged', () => {
      const result = filterEnumParamsReduce([], 'some-id')
      expect(result).toEqual([])
    })

    it('should handle single-element arrays', () => {
      const result = filterEnumParamsReduce(['solo'], 'missing')
      expect(result).toEqual(['solo'])
    })
  })

  describe('filterParamsOptions', () => {
    it('should return null when params is null', () => {
      const result = filterParamsOptions(null, ['a', 'b'])
      expect(result).toBeNull()
    })

    it('should return null when params is undefined', () => {
      const result = filterParamsOptions(undefined, ['a', 'b'])
      expect(result).toBeNull()
    })

    it('should return null when params is an empty array', () => {
      const result = filterParamsOptions([], ['a', 'b'])
      expect(result).toBeNull()
    })

    it('should return null when paramsOption is empty', () => {
      const result = filterParamsOptions(['a'], [])
      expect(result).toBeNull()
    })

    it('should return null when paramsOption is null', () => {
      const result = filterParamsOptions(['a'], null)
      expect(result).toBeNull()
    })

    it('should return matching string when single string param matches option', () => {
      const result = filterParamsOptions('apple', ['apple', 'banana'])
      expect(result).toBe('apple')
    })

    it('should return null when single string param does not match any option', () => {
      const result = filterParamsOptions('cherry', ['apple', 'banana'])
      expect(result).toBeNull()
    })

    it('should filter array params to only matching options', () => {
      const result = filterParamsOptions(['apple', 'cherry', 'banana'], ['apple', 'banana'])
      expect(result).toEqual(['apple', 'banana'])
    })

    it('should return null when no array params match', () => {
      const result = filterParamsOptions(['cherry', 'date'], ['apple', 'banana'])
      expect(result).toBeNull()
    })

    it('should match hierarchical value with -de- separator', () => {
      const result = filterParamsOptions('香橙店-de-浓郁椰奶', ['香橙店-de-浓郁椰奶', '其他'])
      expect(result).toBe('香橙店-de-浓郁椰奶')
    })

    it('should match parent prefix when child options exist', () => {
      const result = filterParamsOptions('香橙店', ['香橙店-de-浓郁椰奶', '其他'])
      expect(result).toBe('香橙店')
    })

    it('should filter array with hierarchical matches', () => {
      const result = filterParamsOptions(
        ['香橙店', '其他'],
        ['香橙店-de-浓郁椰奶', '其他']
      )
      expect(result).toEqual(['香橙店', '其他'])
    })

    it('should match intermediate parent in -de- hierarchy', () => {
      const result = filterParamsOptions(
        'A-de-B',
        ['A-de-B-de-C', 'X-de-Y']
      )
      expect(result).toBe('A-de-B')
    })

    it('should return null for non-string array elements', () => {
      const result = filterParamsOptions([123, true] as any, ['a', 'b'])
      expect(result).toBeNull()
    })
  })
})
