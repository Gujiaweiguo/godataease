import { describe, it, expect, vi, beforeEach } from 'vitest'

const deepCopyMock = vi.fn((obj: any) => JSON.parse(JSON.stringify(obj)))

vi.mock('@/utils/utils', () => ({
  deepCopy: (obj: any) => deepCopyMock(obj)
}))

import defaultConditionTrans from '@/utils/CanvasInfoTransUtils'

describe('CanvasInfoTransUtils', () => {
  beforeEach(() => {
    deepCopyMock.mockClear()
    deepCopyMock.mockImplementation((obj: any) => JSON.parse(JSON.stringify(obj)))
  })

  it('should return empty arrays for empty componentData', () => {
    const canvasInfo = {
      componentData: JSON.stringify([])
    }
    const result = defaultConditionTrans(canvasInfo)
    expect(result.sourceFilter).toEqual([])
    expect(result.defaultFilter).toEqual([])
    expect(result.sourceDefaultFilter).toEqual([])
    expect(result.componentMap).toEqual({})
  })

  it('should extract VQuery propValue as filters', () => {
    const filterItems = [
      { id: 'filter1', name: 'Filter One' },
      { id: 'filter2', name: 'Filter Two' }
    ]
    const canvasInfo = {
      componentData: JSON.stringify([
        { component: 'VQuery', propValue: filterItems }
      ])
    }
    const result = defaultConditionTrans(canvasInfo)
    expect(result.sourceFilter).toEqual(filterItems)
    expect(result.defaultFilter).toEqual(filterItems)
  })

  it('should skip non-VQuery components', () => {
    const canvasInfo = {
      componentData: JSON.stringify([
        { component: 'Chart', propValue: [{ id: 'chart1' }] },
        { component: 'Table', propValue: [{ id: 'table1' }] }
      ])
    }
    const result = defaultConditionTrans(canvasInfo)
    expect(result.sourceFilter).toEqual([])
    expect(result.defaultFilter).toEqual([])
  })

  it('should build componentMap for VQuery items', () => {
    const vqueryComponent = { component: 'VQuery', propValue: [{ id: 'f1' }, { id: 'f2' }] }
    const canvasInfo = {
      componentData: JSON.stringify([vqueryComponent])
    }
    const result = defaultConditionTrans(canvasInfo)
    expect(result.componentMap['f1']).toBeDefined()
    expect(result.componentMap['f2']).toBeDefined()
  })

  it('should apply reportFilterInfo overrides', () => {
    const filterItems = [
      { id: 'filter1', name: 'Original' },
      { id: 'filter2', name: 'Keep' }
    ]
    const overrideFilter = { id: 'filter1', name: 'Overridden' }
    const canvasInfo = {
      componentData: JSON.stringify([
        { component: 'VQuery', propValue: filterItems }
      ]),
      reportFilterInfo: {
        filter1: { filterInfo: JSON.stringify(overrideFilter) }
      }
    }
    const result = defaultConditionTrans(canvasInfo)
    // sourceFilter is allFilter — unchanged by reportFilterInfo
    expect(result.sourceFilter[0]).toEqual(filterItems[0])
    // defaultFilter gets spliced with the override
    expect(result.defaultFilter[0]).toEqual(overrideFilter)
    // filter2 should remain unchanged in both
    expect(result.sourceFilter[1]).toEqual(filterItems[1])
    expect(result.defaultFilter[1]).toEqual(filterItems[1])
  })

  it('should handle multiple VQuery components', () => {
    const canvasInfo = {
      componentData: JSON.stringify([
        { component: 'VQuery', propValue: [{ id: 'a1' }] },
        { component: 'Chart', propValue: [] },
        { component: 'VQuery', propValue: [{ id: 'b1' }, { id: 'b2' }] }
      ])
    }
    const result = defaultConditionTrans(canvasInfo)
    expect(result.sourceFilter).toHaveLength(3)
    expect(result.sourceFilter.map((f: any) => f.id)).toEqual(['a1', 'b1', 'b2'])
  })

  it('should return sourceDefaultFilter as a deep copy of defaultFilter', () => {
    const filterItems = [{ id: 'f1', value: 'test' }]
    const canvasInfo = {
      componentData: JSON.stringify([
        { component: 'VQuery', propValue: filterItems }
      ])
    }
    const result = defaultConditionTrans(canvasInfo)
    expect(result.sourceDefaultFilter).toEqual(result.defaultFilter)
    // Verify it's a separate copy
    expect(result.sourceDefaultFilter).not.toBe(result.defaultFilter)
  })

  it('should handle null reportFilterInfo gracefully', () => {
    const canvasInfo = {
      componentData: JSON.stringify([
        { component: 'VQuery', propValue: [{ id: 'x' }] }
      ]),
      reportFilterInfo: null
    }
    const result = defaultConditionTrans(canvasInfo)
    expect(result.sourceFilter).toEqual([{ id: 'x' }])
    expect(result.defaultFilter).toEqual([{ id: 'x' }])
  })

  it('should handle undefined reportFilterInfo', () => {
    const canvasInfo = {
      componentData: JSON.stringify([
        { component: 'VQuery', propValue: [{ id: 'y' }] }
      ])
    }
    const result = defaultConditionTrans(canvasInfo)
    expect(result.sourceFilter).toEqual([{ id: 'y' }])
    expect(result.defaultFilter).toEqual([{ id: 'y' }])
  })
})
