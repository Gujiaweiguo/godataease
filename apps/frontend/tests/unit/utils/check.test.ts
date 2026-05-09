import { describe, it, expect } from 'vitest'
import checkArrayRepeat from '@/utils/check'

describe('checkArrayRepeat', () => {
  it('should return false for empty array', () => {
    expect(checkArrayRepeat([], 'name')).toBe(false)
  })

  it('should return false for single element array', () => {
    expect(checkArrayRepeat([{ name: 'a' }], 'name')).toBe(false)
  })

  it('should return false when no duplicates exist', () => {
    const data = [
      { name: 'alpha', id: 1 },
      { name: 'beta', id: 2 },
      { name: 'gamma', id: 3 }
    ]
    expect(checkArrayRepeat(data, 'name')).toBe(false)
  })

  it('should return true when duplicates exist at adjacent positions', () => {
    const data = [
      { name: 'alpha', id: 1 },
      { name: 'alpha', id: 2 }
    ]
    expect(checkArrayRepeat(data, 'name')).toBe(true)
  })

  it('should return true when duplicates exist at non-adjacent positions', () => {
    const data = [
      { name: 'alpha', id: 1 },
      { name: 'beta', id: 2 },
      { name: 'alpha', id: 3 }
    ]
    expect(checkArrayRepeat(data, 'name')).toBe(true)
  })

  it('should return false when values are different types but same string representation', () => {
    const data = [
      { value: 1 },
      { value: '1' }
    ]
    // 1 === '1' is false with strict equality
    expect(checkArrayRepeat(data, 'value')).toBe(false)
  })

  it('should detect duplicate by numeric key', () => {
    const data = [
      { id: 1, label: 'a' },
      { id: 2, label: 'b' },
      { id: 1, label: 'c' }
    ]
    expect(checkArrayRepeat(data, 'id')).toBe(true)
  })

  it('should detect duplicates at the last positions', () => {
    const data = [
      { name: 'a' },
      { name: 'b' },
      { name: 'c' },
      { name: 'b' }
    ]
    expect(checkArrayRepeat(data, 'name')).toBe(true)
  })

  it('should handle arrays with all identical elements', () => {
    const data = [
      { name: 'same' },
      { name: 'same' },
      { name: 'same' }
    ]
    expect(checkArrayRepeat(data, 'name')).toBe(true)
  })

  it('should not find duplicates with undefined key values', () => {
    const data = [
      { name: 'a' },
      { name: 'b' }
    ]
    // undefined === undefined is true, so this should detect duplicate
    expect(checkArrayRepeat(data, 'nonexistent')).toBe(true)
  })
})
