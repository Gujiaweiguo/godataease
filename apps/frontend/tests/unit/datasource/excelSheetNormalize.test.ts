import { describe, it, expect } from 'vitest'
import {
  buildSheetJsonArray,
  isSheetEnabled,
  normalizeUploadedSheets
} from '@/views/visualized/data/datasource/form/excelSheetNormalize'

describe('excel sheet normalize helpers', () => {
  it('maps values by fieldIndex when header has empty columns', () => {
    const sheets = normalizeUploadedSheets([
      {
        sheetId: 's1',
        fields: [
          { originName: 'name', checked: true, fieldIndex: 0 },
          { originName: 'age', fieldIndex: 2 }
        ],
        data: [['Alice', 'ignored', '18']]
      }
    ])

    expect(sheets).toHaveLength(1)
    expect(sheets[0].fields?.[0].checked).toBe(true)
    expect(sheets[0].fields?.[1].checked).toBe(true)
    expect(sheets[0].jsonArray).toEqual([{ name: 'Alice', age: '18' }])
  })

  it('keeps existing jsonArray and does not rebuild', () => {
    const sheet = {
      fields: [{ originName: 'name', fieldIndex: 0 }],
      data: [['Alice']],
      jsonArray: [{ name: 'from-server' }]
    }

    const jsonArray = buildSheetJsonArray(sheet)
    expect(jsonArray).toEqual([{ name: 'from-server' }])
  })

  it('recognizes sheet and isSheet flags', () => {
    expect(isSheetEnabled({ sheet: true })).toBe(true)
    expect(isSheetEnabled({ isSheet: true })).toBe(true)
    expect(isSheetEnabled({})).toBe(false)
    expect(isSheetEnabled(undefined)).toBe(false)
  })
})
