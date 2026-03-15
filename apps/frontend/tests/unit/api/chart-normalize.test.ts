import { describe, expect, it, vi } from 'vitest'

vi.mock('@/config/axios', () => {
  return {
    default: {
      post: vi.fn(),
      get: vi.fn()
    }
  }
})

import { __chartDataNormalizeTestUtils } from '../../../src/api/chart'

const viewInfo = {
  xAxis: [
    {
      id: 'x1',
      name: '地区',
      dataeaseName: 'region',
      groupType: 'd',
      deType: 0
    }
  ],
  xAxisExt: [],
  yAxis: [
    {
      id: 'y1',
      name: '销售额',
      dataeaseName: 'amount',
      summary: 'sum',
      groupType: 'q',
      deType: 2
    }
  ],
  yAxisExt: []
}

describe('chart api data normalize', () => {
  it('normalizes raw rows to chart data payload', () => {
    const payload = {
      rows: [
        { region: '华东', amount: 10 },
        { region: '华北', amount: 20 }
      ],
      data: {}
    }

    const normalized = __chartDataNormalizeTestUtils.normalizeChartDataResponse(payload, viewInfo) as {
      data?: {
        data?: Array<Record<string, unknown>>
        series?: Array<Record<string, unknown>>
        tableRow?: Array<Record<string, unknown>>
        fields?: Array<Record<string, unknown>>
      }
    }

    expect(normalized.data?.data?.length).toBe(2)
    expect(normalized.data?.data?.[0].field).toBe('华东')
    expect(normalized.data?.data?.[0].value).toBe(10)

    expect(normalized.data?.series?.length).toBe(1)
    expect(normalized.data?.series?.[0].name).toBe('销售额')
    expect(normalized.data?.series?.[0].data).toEqual([30])

    expect(normalized.data?.tableRow?.length).toBe(2)
    expect(normalized.data?.fields?.length).toBeGreaterThan(0)
  })

  it('keeps existing normalized data and series', () => {
    const payload = {
      rows: [{ region: '华东', amount: 10 }],
      data: {
        data: [{ field: '已处理', value: 99 }],
        series: [{ name: '已存在', data: [99] }],
        tableRow: [{ region: '华东', amount: 10 }],
        fields: [
          {
            id: 'f1',
            dataeaseName: 'region',
            name: '地区'
          }
        ]
      }
    }

    const normalized = __chartDataNormalizeTestUtils.normalizeChartDataResponse(payload, viewInfo) as {
      data?: {
        data?: Array<Record<string, unknown>>
        series?: Array<Record<string, unknown>>
      }
    }

    expect(normalized.data?.data).toEqual([{ field: '已处理', value: 99 }])
    expect(normalized.data?.series).toEqual([{ name: '已存在', data: [99] }])
  })

  it('builds fallback fields from axis config when fields missing', () => {
    const payload = {
      rows: [{ region: '华东', amount: 10 }],
      data: {}
    }

    const normalized = __chartDataNormalizeTestUtils.normalizeChartDataResponse(payload, viewInfo) as {
      data?: {
        fields?: Array<Record<string, unknown>>
      }
    }

    const fieldNames = (normalized.data?.fields || []).map(item => item.dataeaseName)
    expect(fieldNames).toContain('region')
    expect(fieldNames).toContain('amount')
  })
})
