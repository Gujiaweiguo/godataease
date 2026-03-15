import { cloneDeep, defaultsDeep } from 'lodash-es'

import { BASE_VIEW_CONFIG } from '../../editor/util/chart'
import { deepCopy } from '@/utils/utils'

export const buildRenderedChart = (
  viewInfo: Chart,
  chartData: Partial<Chart['data']>,
  fontFamily?: string
): ChartObj => {
  return deepCopy({
    ...defaultsDeep(viewInfo, cloneDeep(BASE_VIEW_CONFIG)),
    data: chartData,
    ...(fontFamily && fontFamily !== 'inherit' ? { fontFamily } : {})
  } as ChartObj)
}
