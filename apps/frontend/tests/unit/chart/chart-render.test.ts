import { describe, expect, it, vi } from 'vitest'

vi.mock('@/hooks/web/useI18n', () => ({
  useI18n: () => ({ t: (value: string) => value })
}))

vi.mock('@/config/axios', () => ({
  default: {
    post: vi.fn(),
    get: vi.fn()
  }
}))

import { buildRenderedChart } from '../../../src/views/chart/components/views/components/chart-render'

describe('buildRenderedChart', () => {
  it('preserves gauge type instead of falling back to base bar type', () => {
    const chart = buildRenderedChart(
      {
        id: 'gauge-1',
        type: 'gauge',
        render: 'antv',
        yAxis: [{ id: 'quota-1', name: '完成率', summary: 'sum', groupType: 'q' }],
        customAttr: { label: {}, basicStyle: {}, misc: {}, tooltip: {}, tableTotal: {}, tableHeader: {}, tableCell: {}, indicator: {}, indicatorName: {}, map: {} },
        customStyle: { text: {}, legend: {}, xAxis: {}, yAxis: {}, yAxisExt: {}, misc: {} },
        senior: { functionCfg: {}, assistLineCfg: {}, threshold: {}, scrollCfg: {}, areaMapping: {}, bubbleCfg: {} }
      } as unknown as Chart,
      {
        series: [{ name: '完成率', data: [68, 0, 100] }],
        fields: []
      },
      'PingFang'
    )

    expect(chart.type).toBe('gauge')
    expect(chart.render).toBe('antv')
    expect(chart.data.series?.[0].data[0]).toBe(68)
    expect(chart.fontFamily).toBe('PingFang')
  })

  it('preserves liquid type instead of falling back to base bar type', () => {
    const chart = buildRenderedChart(
      {
        id: 'liquid-1',
        type: 'liquid',
        render: 'antv',
        yAxis: [{ id: 'quota-1', name: '库存水位', summary: 'sum', groupType: 'q' }],
        customAttr: { label: {}, basicStyle: {}, misc: {}, tooltip: {}, tableTotal: {}, tableHeader: {}, tableCell: {}, indicator: {}, indicatorName: {}, map: {} },
        customStyle: { text: {}, legend: {}, xAxis: {}, yAxis: {}, yAxisExt: {}, misc: {} },
        senior: { functionCfg: {}, assistLineCfg: {}, threshold: {}, scrollCfg: {}, areaMapping: {}, bubbleCfg: {} }
      } as unknown as Chart,
      {
        series: [{ name: '库存水位', data: [45, 100] }],
        fields: []
      }
    )

    expect(chart.type).toBe('liquid')
    expect(chart.render).toBe('antv')
    expect(chart.data.series?.[0].data[0]).toBe(45)
  })
})
