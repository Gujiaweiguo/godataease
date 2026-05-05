import { describe, expect, it } from 'vitest'
import type { FormFieldConfig } from '@/views/data-filling/types'
import {
  buildPlaceholder,
  normalizeOptions,
  parseFormSchema,
  resolveComponentType,
  resolveFieldKey,
  toFieldConfig
} from './schemaParser'

describe('toFieldConfig', () => {
  it('extracts field key from settings.mapping.columnName', () => {
    const backendField = {
      type: 'text',
      typeName: '文本',
      settings: {
        name: '姓名',
        mapping: { columnName: 'user_name', type: 'text' }
      }
    }

    const result = toFieldConfig(backendField, 0)

    expect(result).not.toBeNull()
    expect(result?.field).toBe('user_name')
    expect(result?.label).toBe('姓名')
    expect(result?.name).toBe('user_name')
  })

  it('does NOT default to field_1 when settings.mapping.columnName is present', () => {
    const backendField = {
      type: 'text',
      settings: {
        name: '姓名',
        mapping: { columnName: 'user_name', type: 'text' }
      }
    }

    const result = toFieldConfig(backendField, 0)

    expect(result?.field).toBe('user_name')
    expect(result?.field).not.toBe('field_1')
  })

  it('handles flat legacy format {name, label, type}', () => {
    const flatField = { name: 'user_name', label: '姓名', type: 'text' }

    const result = toFieldConfig(flatField, 0)

    expect(result).not.toBeNull()
    expect(result?.name).toBe('user_name')
    expect(result?.label).toBe('姓名')
  })

  it('extracts required from settings', () => {
    const field = {
      type: 'text',
      settings: { name: '姓名', required: true, mapping: { columnName: 'user_name', type: 'text' } }
    }

    expect(toFieldConfig(field, 0)?.required).toBe(true)
  })

  it('extracts options from settings', () => {
    const field = {
      type: 'select',
      settings: {
        name: '城市',
        mapping: { columnName: 'city', type: 'text' },
        options: [
          { name: '北京', value: 'bj' },
          { name: '上海', value: 'sh' }
        ]
      }
    }

    const result = toFieldConfig(field, 0)

    expect(result?.options).toHaveLength(2)
    expect(result?.options?.[0].name).toBe('北京')
  })

  it('extracts precision from mapping.accuracy', () => {
    const field = {
      type: 'decimal',
      settings: {
        name: '金额',
        mapping: { columnName: 'amount', type: 'decimal', accuracy: 2 }
      }
    }

    expect(toFieldConfig(field, 0)?.precision).toBe(2)
  })

  it('maps nvarchar to text type', () => {
    const field = {
      type: 'nvarchar',
      settings: { name: '名称', mapping: { columnName: 'name', type: 'nvarchar' } }
    }

    expect(toFieldConfig(field, 0)?.type).toBe('text')
  })

  it('returns null for non-record input', () => {
    expect(toFieldConfig(null, 0)).toBeNull()
    expect(toFieldConfig('string', 0)).toBeNull()
    expect(toFieldConfig(42, 0)).toBeNull()
  })

  it('uses field_${index+1} as fallback when no name or field exists', () => {
    const field = { type: 'text' }

    const result = toFieldConfig(field, 2)

    expect(result?.name).toBe('field_3')
    expect(result?.label).toBe('Field 3')
  })

  it('extracts placeholder from settings', () => {
    const field = {
      type: 'text',
      settings: {
        name: '备注',
        placeholder: '请输入备注信息',
        mapping: { columnName: 'remark', type: 'text' }
      }
    }

    expect(toFieldConfig(field, 0)?.placeholder).toBe('请输入备注信息')
  })

  it('extracts multiple from settings', () => {
    const field = {
      type: 'select',
      settings: {
        name: '标签',
        multiple: true,
        mapping: { columnName: 'tags', type: 'text' }
      }
    }

    expect(toFieldConfig(field, 0)?.multiple).toBe(true)
  })
})

describe('parseFormSchema', () => {
  it('parses JSON array', () => {
    const json = JSON.stringify([
      { type: 'text', settings: { name: '姓名', mapping: { columnName: 'user_name', type: 'text' } } }
    ])

    const result = parseFormSchema(json)

    expect(result).toHaveLength(1)
    expect(result[0].field).toBe('user_name')
  })

  it('parses {fields:[...]} format', () => {
    const json = JSON.stringify({
      fields: [{ type: 'text', settings: { name: '姓名', mapping: { columnName: 'user_name', type: 'text' } } }]
    })

    const result = parseFormSchema(json)

    expect(result).toHaveLength(1)
    expect(result[0].field).toBe('user_name')
  })

  it('parses {forms:[...]} format', () => {
    const json = JSON.stringify({
      forms: [{ type: 'text', settings: { name: '姓名', mapping: { columnName: 'user_name', type: 'text' } } }]
    })

    const result = parseFormSchema(json)

    expect(result).toHaveLength(1)
    expect(result[0].field).toBe('user_name')
  })

  it('returns empty array for empty string', () => {
    expect(parseFormSchema('')).toEqual([])
    expect(parseFormSchema('   ')).toEqual([])
  })

  it('returns empty array for invalid JSON', () => {
    expect(parseFormSchema('{invalid json}')).toEqual([])
  })

  it('filters out null entries', () => {
    const json = JSON.stringify([
      null,
      { type: 'text', settings: { name: '姓名', mapping: { columnName: 'user_name', type: 'text' } } },
      'invalid'
    ])

    const result = parseFormSchema(json)

    expect(result).toHaveLength(1)
    expect(result[0].field).toBe('user_name')
  })

  it('sorts by order field', () => {
    const json = JSON.stringify([
      { type: 'text', order: 2, settings: { name: 'B', mapping: { columnName: 'b', type: 'text' } } },
      { type: 'text', order: 1, settings: { name: 'A', mapping: { columnName: 'a', type: 'text' } } }
    ])

    const result = parseFormSchema(json)

    expect(result.map(item => item.field)).toEqual(['a', 'b'])
  })
})

describe('normalizeOptions', () => {
  it('normalizes {name, value} objects', () => {
    expect(normalizeOptions([{ name: '北京', value: 'bj' }])).toEqual([{ name: '北京', value: 'bj', disabled: false }])
  })

  it('normalizes string primitives', () => {
    expect(normalizeOptions(['北京'])).toEqual([{ name: '北京', value: '北京' }])
  })

  it('normalizes number primitives', () => {
    expect(normalizeOptions([1])).toEqual([{ name: '1', value: '1' }])
  })

  it('returns empty array for non-array input', () => {
    expect(normalizeOptions(null)).toEqual([])
    expect(normalizeOptions({})).toEqual([])
  })

  it('handles mixed valid/invalid entries', () => {
    expect(normalizeOptions([{ label: '上海' }, false, undefined, '北京', 2, { foo: 'bar' }])).toEqual([
      { name: '上海', value: '上海', disabled: false },
      { name: '北京', value: '北京' },
      { name: '2', value: '2' }
    ])
  })
})

describe('resolveFieldKey', () => {
  it('prefers field.field', () => {
    const field: FormFieldConfig = { field: 'user_name', name: 'name', label: '姓名', type: 'text' }

    expect(resolveFieldKey(field, 0)).toBe('user_name')
  })

  it('falls back to field.name', () => {
    const field: FormFieldConfig = { name: 'user_name', label: '姓名', type: 'text' }

    expect(resolveFieldKey(field, 0)).toBe('user_name')
  })

  it('falls back to field_${index+1}', () => {
    const field = { name: '', label: '姓名', type: 'text' } as FormFieldConfig

    expect(resolveFieldKey(field, 2)).toBe('field_3')
  })
})

describe('resolveComponentType', () => {
  it.each([
    ['text', 'input'],
    ['number', 'number'],
    ['decimal', 'decimal'],
    ['date', 'date'],
    ['datetime', 'datetime'],
    ['select', 'select'],
    ['unknown', 'input'],
    ['TEXT', 'input']
  ])('maps %s to %s', (input, expected) => {
    expect(resolveComponentType(input)).toBe(expected)
  })
})

describe('buildPlaceholder', () => {
  it('returns custom placeholder when provided', () => {
    expect(buildPlaceholder({ name: 'remark', label: '备注', type: 'text', placeholder: '请输入备注信息' }, 'input')).toBe(
      '请输入备注信息'
    )
  })

  it('returns 请选择 for select type', () => {
    expect(buildPlaceholder({ name: 'city', label: '城市', type: 'select' }, 'select')).toBe('请选择城市')
  })

  it('returns 请选择 for date type', () => {
    expect(buildPlaceholder({ name: 'date', label: '日期', type: 'date' }, 'date')).toBe('请选择日期')
  })

  it('returns 请输入 for input type', () => {
    expect(buildPlaceholder({ name: 'name', label: '姓名', type: 'text' }, 'input')).toBe('请输入姓名')
  })
})
