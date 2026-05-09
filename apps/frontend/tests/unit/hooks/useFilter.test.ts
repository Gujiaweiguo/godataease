import { describe, it, expect, vi, beforeEach } from 'vitest'

const { mockComponentData, mockCanvasStyleData } = vi.hoisted(() => ({
  mockComponentData: { value: [] as any[] },
  mockCanvasStyleData: { value: { popupAvailable: false } }
}))

vi.mock('@/store/modules/data-visualization/dvMain', () => ({
  dvMainStoreWithOut: () => ({
    componentData: mockComponentData,
    canvasStyleData: mockCanvasStyleData
  })
}))

vi.mock('pinia', () => ({
  storeToRefs: (_store: any) => ({
    componentData: mockComponentData,
    canvasStyleData: mockCanvasStyleData
  })
}))

vi.mock('@/custom-component/v-query/time-format', () => ({
  getDynamicRange: vi.fn(() => [Date.now(), Date.now()]),
  getCustomTime: vi.fn(() => Date.now())
}))

vi.mock('@/custom-component/v-query/time-format-dayjs', () => ({
  getCustomRange: vi.fn(() => [Date.now(), Date.now()])
}))

import { getRange, searchQuery, useFilter } from '@/hooks/web/useFilter'

describe('useFilter', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockComponentData.value = []
    mockCanvasStyleData.value = { popupAvailable: false }
  })

  describe('getRange', () => {
    it('should return year range for year granularity', () => {
      const timestamp = new Date('2024-06-15').getTime()
      const result = getRange(timestamp, 'year')

      expect(result).toHaveLength(2)
      expect(result[0]).toBeLessThanOrEqual(result[1])
    })

    it('should return month range for month granularity', () => {
      const timestamp = new Date('2024-06-15').getTime()
      const result = getRange(timestamp, 'month')

      expect(result).toHaveLength(2)
      expect(result[0]).toBeLessThanOrEqual(result[1])
    })

    it('should return day range for date granularity', () => {
      const timestamp = new Date('2024-06-15').getTime()
      const result = getRange(timestamp, 'date')

      expect(result).toHaveLength(2)
      expect(result[0]).toBeLessThanOrEqual(result[1])
    })

    it('should return same timestamp for datetime granularity', () => {
      const timestamp = new Date('2024-06-15T10:30:00').getTime()
      const result = getRange(timestamp, 'datetime')

      expect(result).toHaveLength(2)
      expect(result[0]).toBe(result[1])
    })

    it('should return undefined for unknown granularity', () => {
      const timestamp = new Date('2024-06-15').getTime()
      const result = getRange(timestamp, 'unknown')

      expect(result).toBeUndefined()
    })
  })

  describe('useFilter', () => {
    it('should return filter array', () => {
      const { filter } = useFilter('chart-1')
      expect(Array.isArray(filter)).toBe(true)
    })

    it('should return empty filter when no VQuery components exist', () => {
      mockComponentData.value = [
        { component: 'UserView', id: 'comp-1' },
        { component: 'Group', id: 'comp-2', propValue: [] }
      ]

      const { filter } = useFilter('chart-1')
      expect(filter).toHaveLength(0)
    })

    it('should process VQuery components in componentData', () => {
      mockComponentData.value = [
        {
          component: 'VQuery',
          id: 'vquery-1',
          propValue: []
        }
      ]

      const { filter } = useFilter('chart-1')
      expect(filter).toHaveLength(0)
    })
  })

  describe('searchQuery', () => {
    it('should push filter for optionFilter when present', () => {
      const filter: any[] = []
      const queryComponentList = [
        {
          id: 'vquery-1',
          propValue: [
            {
              displayType: '2',
              checkedFields: ['chart-1'],
              checkedFieldsMap: { 'chart-1': 'field-1' },
              optionFilter: ['opt1', 'opt2'],
              id: 'filter-1',
              parameters: []
            }
          ]
        }
      ]

      searchQuery(queryComponentList, filter, 'chart-1', false)

      expect(filter).toHaveLength(1)
      expect(filter[0].filterId).toBe('filter-1')
      expect(filter[0].operator).toBe('in')
      expect(filter[0].value).toEqual(['opt1', 'opt2'])
    })

    it('should skip components not matching curComponentId', () => {
      const filter: any[] = []
      const queryComponentList = [
        {
          id: 'vquery-1',
          propValue: [
            {
              displayType: '2',
              checkedFields: ['chart-2'],
              checkedFieldsMap: { 'chart-2': 'field-2' },
              id: 'filter-1'
            }
          ]
        }
      ]

      searchQuery(queryComponentList, filter, 'chart-1', false)

      expect(filter).toHaveLength(0)
    })

    it('should skip when checkedFieldsMap for curComponentId is falsy', () => {
      const filter: any[] = []
      const queryComponentList = [
        {
          id: 'vquery-1',
          propValue: [
            {
              displayType: '2',
              checkedFields: ['chart-1'],
              checkedFieldsMap: { 'chart-1': null },
              id: 'filter-1'
            }
          ]
        }
      ]

      searchQuery(queryComponentList, filter, 'chart-1', false)

      expect(filter).toHaveLength(0)
    })

    it('should handle empty propValue gracefully', () => {
      const filter: any[] = []
      const queryComponentList = [{ id: 'vquery-1', propValue: [] }]

      searchQuery(queryComponentList, filter, 'chart-1', false)

      expect(filter).toHaveLength(0)
    })

    it('should use "in" operator for multiple select with displayType 2', () => {
      const filter: any[] = []
      const queryComponentList = [
        {
          id: 'vquery-1',
          propValue: [
            {
              displayType: '2',
              checkedFields: ['chart-1'],
              checkedFieldsMap: { 'chart-1': 'field-1' },
              id: 'filter-1',
              parameters: [],
              multiple: true,
              defaultValueCheck: true,
              defaultValue: ['val1', 'val2']
            }
          ]
        }
      ]

      searchQuery(queryComponentList, filter, 'chart-1', true)

      expect(filter).toHaveLength(1)
      expect(filter[0].operator).toBe('in')
    })
  })
})
