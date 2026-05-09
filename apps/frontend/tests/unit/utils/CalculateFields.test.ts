import { describe, it, expect } from 'vitest'
import { Base64 } from 'js-base64'
import {
  originNameHandle,
  originNameHandleBack,
  originNameHandleWithArr,
  originNameHandleBackWithArr
} from '@/utils/CalculateFields'

describe('CalculateFields', () => {
  describe('originNameHandle', () => {
    it('should encode originName for items with extField === 2', () => {
      const arr = [{ extField: 2, originName: 'SUM(amount)' }]
      originNameHandle(arr)
      expect(arr[0].originName).toBe(Base64.encode('SUM(amount)'))
    })

    it('should not encode originName for items with extField !== 2', () => {
      const arr = [
        { extField: 0, originName: 'field_a' },
        { extField: 1, originName: 'field_b' }
      ]
      originNameHandle(arr)
      expect(arr[0].originName).toBe('field_a')
      expect(arr[1].originName).toBe('field_b')
    })

    it('should handle mixed extField values', () => {
      const arr = [
        { extField: 0, originName: 'normal' },
        { extField: 2, originName: 'calculated' },
        { extField: 1, originName: 'other' }
      ]
      originNameHandle(arr)
      expect(arr[0].originName).toBe('normal')
      expect(arr[1].originName).toBe(Base64.encode('calculated'))
      expect(arr[2].originName).toBe('other')
    })

    it('should handle empty array without error', () => {
      const arr: any[] = []
      expect(() => originNameHandle(arr)).not.toThrow()
    })

    it('should handle undefined argument as empty array', () => {
      expect(() => originNameHandle(undefined as any)).not.toThrow()
    })

    it('should encode Chinese characters in originName', () => {
      const arr = [{ extField: 2, originName: '计算字段' }]
      originNameHandle(arr)
      expect(arr[0].originName).toBe(Base64.encode('计算字段'))
    })
  })

  describe('originNameHandleBack', () => {
    it('should decode originName for items with extField === 2', () => {
      const original = 'SUM(amount)'
      const arr = [{ extField: 2, originName: Base64.encode(original) }]
      originNameHandleBack(arr)
      expect(arr[0].originName).toBe(original)
    })

    it('should not decode originName for items with extField !== 2', () => {
      const arr = [{ extField: 0, originName: 'field_a' }]
      originNameHandleBack(arr)
      expect(arr[0].originName).toBe('field_a')
    })

    it('should handle empty array without error', () => {
      expect(() => originNameHandleBack([])).not.toThrow()
    })
  })

  describe('originNameHandleWithArr', () => {
    it('should encode originName across multiple field arrays', () => {
      const obj = {
        fieldsA: [{ extField: 2, originName: 'calc_a' }],
        fieldsB: [{ extField: 2, originName: 'calc_b' }]
      }
      originNameHandleWithArr(obj, ['fieldsA', 'fieldsB'])
      expect(obj.fieldsA[0].originName).toBe(Base64.encode('calc_a'))
      expect(obj.fieldsB[0].originName).toBe(Base64.encode('calc_b'))
    })

    it('should skip missing field keys gracefully', () => {
      const obj = { fieldsA: [{ extField: 2, originName: 'calc' }] }
      originNameHandleWithArr(obj, ['fieldsA', 'nonExistent'])
      expect(obj.fieldsA[0].originName).toBe(Base64.encode('calc'))
    })

    it('should handle undefined obj as empty object', () => {
      expect(() => originNameHandleWithArr(undefined, ['fields'])).not.toThrow()
    })

    it('should handle empty fields array', () => {
      const obj = { fieldsA: [{ extField: 2, originName: 'calc' }] }
      originNameHandleWithArr(obj, [])
      expect(obj.fieldsA[0].originName).toBe('calc')
    })
  })

  describe('originNameHandleBackWithArr', () => {
    it('should decode originName across multiple field arrays', () => {
      const obj = {
        fieldsA: [{ extField: 2, originName: Base64.encode('calc_a') }],
        fieldsB: [{ extField: 2, originName: Base64.encode('calc_b') }]
      }
      originNameHandleBackWithArr(obj, ['fieldsA', 'fieldsB'])
      expect(obj.fieldsA[0].originName).toBe('calc_a')
      expect(obj.fieldsB[0].originName).toBe('calc_b')
    })

    it('should handle undefined obj gracefully', () => {
      expect(() => originNameHandleBackWithArr(undefined, ['fields'])).not.toThrow()
    })
  })

  describe('encode-decode roundtrip', () => {
    it('should return original value after encode then decode', () => {
      const original = 'AVG(price * quantity)'
      const arr = [{ extField: 2, originName: original }]
      originNameHandle(arr)
      expect(arr[0].originName).not.toBe(original)
      originNameHandleBack(arr)
      expect(arr[0].originName).toBe(original)
    })
  })
})
