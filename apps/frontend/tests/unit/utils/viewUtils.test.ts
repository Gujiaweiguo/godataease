import { describe, it, expect, vi, beforeEach } from 'vitest'

interface Dimension {
  id: string
  value: string
  timeValue?: string
}

vi.mock('@/utils/timeUitils', () => ({
  getRange: vi.fn((val, style) => `range(${val},${style})`)
}))

vi.mock('lodash-es', () => ({
  union: vi.fn((...args) => [...new Set(args.filter(Boolean).flat())])
}))

import { viewFieldTimeTrans } from '@/utils/viewUtils'
import { getRange } from '@/utils/timeUitils'
import { union } from 'lodash-es'

describe('viewUtils', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('viewFieldTimeTrans', () => {
    it('should set timeValue on date dimension fields (deType === 1)', () => {
      const viewDataInfo = {
        fields: [
          { id: 'f1', dataeaseName: 'create_time', deType: 1, dateStyle: 'y_M_d' },
          { id: 'f2', dataeaseName: 'amount', deType: 2, dateStyle: undefined }
        ]
      }
      const params = {
        dimensionList: [
          { id: 'f1', value: '2024-01-15' } as Dimension,
          { id: 'f2', value: '100' } as Dimension
        ]
      }
      viewFieldTimeTrans(viewDataInfo, params)
      expect((params.dimensionList[0] as Dimension).timeValue).toBe('range(2024-01-15,y_M_d)')
      expect((params.dimensionList[1] as Dimension).timeValue).toBeUndefined()
    })

    it('should not modify dimension when deType is not 1', () => {
      const viewDataInfo = {
        fields: [{ id: 'f1', dataeaseName: 'name', deType: 0, dateStyle: undefined }]
      }
      const params = { dimensionList: [{ id: 'f1', value: 'hello' } as Dimension] }
      viewFieldTimeTrans(viewDataInfo, params)
      expect((params.dimensionList[0] as Dimension).timeValue).toBeUndefined()
    })

    it('should handle null viewDataInfo gracefully', () => {
      const params = { dimensionList: [{ id: 'f1', value: 'test' }] }
      expect(() => viewFieldTimeTrans(null, params)).not.toThrow()
    })

    it('should handle null params gracefully', () => {
      expect(() => viewFieldTimeTrans({}, null)).not.toThrow()
    })

    it('should handle params without dimensionList', () => {
      expect(() => viewFieldTimeTrans({}, {})).not.toThrow()
    })

    it('should use left.fields when viewDataInfo.fields is missing', () => {
      const viewDataInfo = {
        left: {
          fields: [{ id: 'f1', dataeaseName: 'order_date', deType: 1, dateStyle: 'year' }]
        }
      }
      const params = { dimensionList: [{ id: 'f1', value: '2024' } as Dimension] }
      viewFieldTimeTrans(viewDataInfo, params)
      expect((params.dimensionList[0] as Dimension).timeValue).toBe('range(2024,year)')
    })

    it('should use right.fields when left is not available', () => {
      const viewDataInfo = {
        right: {
          fields: [{ id: 'f1', dataeaseName: 'update_date', deType: 1, dateStyle: 'month' }]
        }
      }
      const params = { dimensionList: [{ id: 'f1', value: '2024-03' } as Dimension] }
      viewFieldTimeTrans(viewDataInfo, params)
      expect((params.dimensionList[0] as Dimension).timeValue).toBe('range(2024-03,month)')
    })

    it('should use union of left and right fields when both exist and no top-level fields', () => {
      const viewDataInfo = {
        left: {
          fields: [{ id: 'f1', dataeaseName: 'date_a', deType: 1, dateStyle: 'date' }]
        },
        right: {
          fields: [{ id: 'f2', dataeaseName: 'date_b', deType: 1, dateStyle: 'hour' }]
        }
      }
      const params = {
        dimensionList: [
          { id: 'f1', value: '2024-01-01' } as Dimension,
          { id: 'f2', value: '2024-01-01 10' } as Dimension
        ]
      }
      viewFieldTimeTrans(viewDataInfo, params)
      expect(union).toHaveBeenCalled()
      expect((params.dimensionList[0] as Dimension).timeValue).toBe('range(2024-01-01,date)')
      expect((params.dimensionList[1] as Dimension).timeValue).toBe('range(2024-01-01 10,hour)')
    })

    it('should handle empty dimensionList without error', () => {
      const viewDataInfo = {
        fields: [{ id: 'f1', dataeaseName: 'date', deType: 1, dateStyle: 'year' }]
      }
      const params = { dimensionList: [] }
      expect(() => viewFieldTimeTrans(viewDataInfo, params)).not.toThrow()
    })

    it('should skip dimensions whose id is not in the field map', () => {
      const viewDataInfo = {
        fields: [{ id: 'f1', dataeaseName: 'date', deType: 1, dateStyle: 'year' }]
      }
      const params = { dimensionList: [{ id: 'unknown', value: '2024' }] }
      viewFieldTimeTrans(viewDataInfo, params)
      expect(getRange).not.toHaveBeenCalled()
    })

    it('should handle fields with empty array', () => {
      const viewDataInfo = { fields: [] }
      const params = { dimensionList: [{ id: 'f1', value: '2024' }] }
      viewFieldTimeTrans(viewDataInfo, params)
      expect(getRange).not.toHaveBeenCalled()
    })
  })
})
