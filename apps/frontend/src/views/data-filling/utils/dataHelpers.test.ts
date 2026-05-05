import { describe, expect, it } from 'vitest'
import {
  isBlankPayload,
  isCancelableAction,
  isEmptyFieldValue,
  normalizeBuiltInTableOptions,
  normalizeDatasourceOptions,
  parseRouteFormId
} from './dataHelpers'

describe('isEmptyFieldValue', () => {
  it('returns true for empty values', () => {
    expect(isEmptyFieldValue(undefined)).toBe(true)
    expect(isEmptyFieldValue(null)).toBe(true)
    expect(isEmptyFieldValue('')).toBe(true)
    expect(isEmptyFieldValue([])).toBe(true)
  })

  it('returns false for non-empty values', () => {
    expect(isEmptyFieldValue('text')).toBe(false)
    expect(isEmptyFieldValue(0)).toBe(false)
    expect(isEmptyFieldValue(false)).toBe(false)
    expect(isEmptyFieldValue([1, 2])).toBe(false)
    expect(isEmptyFieldValue({ key: 'val' } as never)).toBe(false)
  })
})

describe('isCancelableAction', () => {
  it('returns true for cancel and close', () => {
    expect(isCancelableAction('cancel')).toBe(true)
    expect(isCancelableAction('close')).toBe(true)
  })

  it('returns false for other values', () => {
    expect(isCancelableAction('other')).toBe(false)
    expect(isCancelableAction(new Error('x'))).toBe(false)
    expect(isCancelableAction(undefined)).toBe(false)
    expect(isCancelableAction(null)).toBe(false)
  })
})

describe('parseRouteFormId', () => {
  it('parses valid ids', () => {
    expect(parseRouteFormId('6')).toBe(6)
    expect(parseRouteFormId(['6'])).toBe(6)
  })

  it('returns null for invalid values', () => {
    expect(parseRouteFormId('')).toBeNull()
    expect(parseRouteFormId('abc')).toBeNull()
    expect(parseRouteFormId('0')).toBeNull()
    expect(parseRouteFormId('-1')).toBeNull()
    expect(parseRouteFormId(undefined)).toBeNull()
    expect(parseRouteFormId(null)).toBeNull()
  })
})

describe('isBlankPayload', () => {
  it('returns true for blank payloads', () => {
    expect(isBlankPayload({})).toBe(true)
    expect(isBlankPayload({ id: '123' })).toBe(true)
    expect(isBlankPayload({ id: '1', name: '' })).toBe(true)
  })

  it('returns false for non-blank payloads', () => {
    expect(isBlankPayload({ id: '1', name: '张三' })).toBe(false)
    expect(isBlankPayload({ name: 'hello' })).toBe(false)
  })
})

describe('normalizeDatasourceOptions', () => {
  it('normalizes datasource options', () => {
    expect(
      normalizeDatasourceOptions([
        { id: 1, name: 'MySQL' },
        { id: 2, name: 'PostgreSQL' },
        { value: 3, label: 'Oracle' }
      ])
    ).toEqual([
      { value: 1, label: 'MySQL' },
      { value: 2, label: 'PostgreSQL' },
      { value: 3, label: 'Oracle' }
    ])
  })

  it('returns empty array for invalid input', () => {
    expect(normalizeDatasourceOptions(null)).toEqual([])
    expect(normalizeDatasourceOptions({})).toEqual([])
    expect(normalizeDatasourceOptions('string')).toEqual([])
  })

  it('skips invalid entries', () => {
    expect(
      normalizeDatasourceOptions([
        { id: '1', name: 'bad-id' },
        { id: 2 },
        { value: 3, label: {} },
        null,
        'x'
      ])
    ).toEqual([])
  })
})

describe('normalizeBuiltInTableOptions', () => {
  it('normalizes string and object entries', () => {
    expect(
      normalizeBuiltInTableOptions(['table1', 'table2', { tableName: 'users' }, { name: 'orders' }])
    ).toEqual([
      { label: 'table1', value: 'table1' },
      { label: 'table2', value: 'table2' },
      { label: 'users', value: 'users' },
      { label: 'orders', value: 'orders' }
    ])
  })

  it('returns empty array for invalid input', () => {
    expect(normalizeBuiltInTableOptions(null)).toEqual([])
    expect(normalizeBuiltInTableOptions({})).toEqual([])
    expect(normalizeBuiltInTableOptions(1)).toEqual([])
  })

  it('skips entries without extractable string values', () => {
    expect(normalizeBuiltInTableOptions([{ value: 1 }, { label: 2 }, false, undefined])).toEqual([])
  })
})
