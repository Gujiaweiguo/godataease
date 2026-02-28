import { describe, it, expect, beforeEach, vi } from 'vitest'
import {
  deepCopy,
  swap,
  checkAddHttp,
  isBtnShow,
  isNull,
  isInIframe,
  isLink,
  getQueryString
} from '@/utils/utils'

// Mock dependencies

// Mock element-plus-secondary to avoid inject warnings
vi.mock('element-plus-secondary', async () => {
  const { elementPlusSecondaryModuleMock } = await import('../helpers')
  return elementPlusSecondaryModuleMock
})
vi.mock('@/hooks/web/useCache', async () => {
  const { createUseCacheModuleMock } = await import('../helpers')
  return createUseCacheModuleMock()
})

vi.mock('@/utils/RemoteJs', () => ({
  loadScript: vi.fn().mockResolvedValue(Promise.resolve())
}))

describe('Utils Functions', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('deepCopy', () => {
    it('should return null when target is null', () => {
      expect(deepCopy(null)).toBeNull()
    })

    it('should return undefined when target is undefined', () => {
      expect(deepCopy(undefined)).toBeUndefined()
    })

    it('should deep copy an object', () => {
      const original = { a: 1, b: { c: 2 } }
      const copy = deepCopy(original)
      
      expect(copy).toEqual(original)
      expect(copy).not.toBe(original)
      expect(copy.b).not.toBe(original.b)
    })

    it('should deep copy an array', () => {
      const original = [1, { a: 2 }, [3, 4]]
      const copy = deepCopy(original)
      
      expect(copy).toEqual(original)
      expect(copy).not.toBe(original)
      expect(copy[1]).not.toBe(original[1])
      expect(copy[2]).not.toBe(original[2])
    })

    it('should handle Date objects inside objects', () => {
      const date = new Date('2024-01-01')
      const original = { created: date }
      const copy = deepCopy(original) as any
      
      expect(copy.created).toBeInstanceOf(Date)
      expect((copy.created as Date).getTime()).toBe(date.getTime())
      expect(copy.created).not.toBe(date)
    })

    it('should return empty object for top-level Date (limitation)', () => {
      const date = new Date('2024-01-01')
      const copy = deepCopy(date) as any
      
      // deepCopy doesn't handle top-level Date objects, only Date values inside objects
      expect(copy).toEqual({})
    })
    it('should return primitive values as-is', () => {
      expect(deepCopy(42)).toBe(42)
      expect(deepCopy('string')).toBe('string')
      expect(deepCopy(true)).toBe(true)
    })

    it('should handle nested objects with null values', () => {
      const original = { a: null, b: { c: null } }
      const copy = deepCopy(original)
      
      expect(copy).toEqual(original)
    })
  })

  describe('swap', () => {
    it('should swap two elements in an array', () => {
      const arr = [1, 2, 3, 4]
      swap(arr, 0, 2)
      
      expect(arr).toEqual([3, 2, 1, 4])
    })

    it('should swap adjacent elements', () => {
      const arr = [1, 2]
      swap(arr, 0, 1)
      
      expect(arr).toEqual([2, 1])
    })

    it('should handle swapping with same index', () => {
      const arr = [1, 2, 3]
      swap(arr, 1, 1)
      
      expect(arr).toEqual([1, 2, 3])
    })
  })

  describe('checkAddHttp', () => {
    it('should add http:// to url without protocol', () => {
      expect(checkAddHttp('example.com')).toBe('http://example.com')
    })

    it('should not modify url with http://', () => {
      expect(checkAddHttp('http://example.com')).toBe('http://example.com')
    })

    it('should not modify url with https://', () => {
      expect(checkAddHttp('https://example.com')).toBe('https://example.com')
    })

    it('should handle uppercase HTTPS', () => {
      expect(checkAddHttp('HTTPS://example.com')).toBe('HTTPS://example.com')
    })

    it('should return empty string for empty input', () => {
      expect(checkAddHttp('')).toBe('')
    })

    it('should return null for null input', () => {
      expect(checkAddHttp(null)).toBeNull()
    })

    it('should return undefined for undefined input', () => {
      expect(checkAddHttp(undefined)).toBeUndefined()
    })
  })

  describe('isBtnShow', () => {
    it('should return true when value is 0', () => {
      expect(isBtnShow('0')).toBe(true)
    })

    it('should return true when value is empty', () => {
      expect(isBtnShow('')).toBe(true)
    })

    it('should return true when value is null', () => {
      expect(isBtnShow(null as any)).toBe(true)
    })

    it('should return false when value is 1', () => {
      expect(isBtnShow('1')).toBe(false)
    })
  })

  describe('isNull', () => {
    it('should return true for undefined', () => {
      expect(isNull(undefined)).toBe(true)
    })

    it('should return true for null', () => {
      expect(isNull(null)).toBe(true)
    })

    it('should return true for string "null"', () => {
      expect(isNull('null')).toBe(true)
    })

    it('should return false for empty string', () => {
      expect(isNull('')).toBe(false)
    })

    it('should return false for 0', () => {
      expect(isNull(0)).toBe(false)
    })

    it('should return false for false', () => {
      expect(isNull(false)).toBe(false)
    })

    it('should return false for object', () => {
      expect(isNull({})).toBe(false)
    })
  })

  describe('isInIframe', () => {
    it('should return true when window.top is not window.self', () => {
      // Mock window.top to be different from window.self
      Object.defineProperty(window, 'top', {
        value: {},
        writable: true
      })
      
      expect(isInIframe()).toBe(true)
    })

    it('should return false when window.top is window.self', () => {
      Object.defineProperty(window, 'top', {
        value: window,
        writable: true
      })
      
      expect(isInIframe()).toBe(false)
    })
  })

  describe('isLink', () => {
    it('should return true when hash starts with #/de-link/', () => {
      Object.defineProperty(window, 'location', {
        value: { hash: '#/de-link/abc123' },
        writable: true
      })
      
      expect(isLink()).toBe(true)
    })

    it('should return false when hash does not start with #/de-link/', () => {
      Object.defineProperty(window, 'location', {
        value: { hash: '#/dashboard' },
        writable: true
      })
      
      expect(isLink()).toBe(false)
    })

    it('should return false for empty hash', () => {
      Object.defineProperty(window, 'location', {
        value: { hash: '' },
        writable: true
      })
      
      expect(isLink()).toBe(false)
    })
  })

  describe('getQueryString', () => {
    it('should return parameter value from URL', () => {
      Object.defineProperty(window, 'location', {
        value: { search: '?name=test&age=25' },
        writable: true
      })
      
      expect(getQueryString('name')).toBe('test')
      expect(getQueryString('age')).toBe('25')
    })

    it('should return null for non-existent parameter', () => {
      Object.defineProperty(window, 'location', {
        value: { search: '?name=test' },
        writable: true
      })
      
      expect(getQueryString('nonexistent')).toBeNull()
    })

    it('should return null for empty search', () => {
      Object.defineProperty(window, 'location', {
        value: { search: '' },
        writable: true
      })
      
      expect(getQueryString('name')).toBeNull()
    })
  })
})
