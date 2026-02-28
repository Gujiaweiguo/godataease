/**
 * 图表对象
 */
declare interface Chart {
  id: string
  render: string
  name: string
  type: string
  title: string
  drill?: boolean
  refreshViewEnable: boolean
  refreshTime: number
  refreshUnit: string
  data: {
    data: any[]
    series?: any[]
    dynamicAssistLines?: AssistLine[]
    fields: ChartViewField[]
    tableRow: []
    //chart-mix
    left: {
      data: any[]
      series?: any[]
      dynamicAssistLines?: AssistLine[]
      fields: ChartViewField[]
      tableRow: []
    }
    right: {
      data: any[]
      series?: any[]
      dynamicAssistLines?: AssistLine[]
      fields: ChartViewField[]
      tableRow: []
    }
    customCalc: any
    customSumResult?: Record<string, any>
  }
  xAxis?: Axis[]
  xAxisExt?: Axis[]
  yAxis?: Axis[]
  yAxisExt?: Axis[]
  extStack?: Axis[]
  extBubble?: Axis[]
  extLabel?: Axis[]
  extTooltip?: Axis[]
  customFilter: { items?: { fieldId?: string }[] }
  senior: CustomSenior
  customAttr: CustomAttr
  customAttrMobile: CustomAttr
  customStyle: CustomStyle
  customStyleMobile: CustomStyle
  drillFields: ChartViewField[]
  drillFilters: Filter[]
  sortPriority: ChartViewField[]
  datasetMode: 0 | 1
  datasourceType: string
  totalItems: number
  tableId: number
  resultMode: string
  resultCount: number
  linkageActive: boolean
  jumpActive: boolean
  aggregate?: boolean
  plugin?: CustomPlugin
  isPlugin: boolean
  extremumValues?: Map<string, any>
  filteredData?: any[]
  container?: string
  /**
   * 针对不是序列字段的图表，通过获取分类字段的值作为序列字段
   */
  seriesFieldObjs?: any[]
  flowMapStartName?: Axis[]
  flowMapEndName?: Axis[]
  showPosition: string

  extColor: Axis[]

  fontFamily?: string
}
declare type CustomAttr = DeepPartial<ChartAttr> | JSONString<DeepPartial<ChartAttr>>
declare type CustomStyle = DeepPartial<ChartStyle> | JSONString<DeepPartial<ChartStyle>>
declare type CustomSenior = DeepPartial<ChartSenior> | JSONString<DeepPartial<ChartSenior>>
declare type CustomPlugin = DeepPartial<ChartPlugin> | JSONString<DeepPartial<ChartPlugin>>

declare type ChartObj = Omit<Chart, 'customAttr' | 'customStyle' | 'senior' | 'plugin'> & {
  customAttr: ChartAttr
  customStyle: ChartStyle
  senior: ChartSenior
  plugin?: ChartPlugin
}

/**
 * 格式化属性
 */
declare interface BaseFormatter {
  /**
   * 格式化类型：auto,value,percent
   */
  type: string
  /**
   * 单位换算
   */
  unitLanguage: 'ch' | 'en'
  unit: number
  /**
   * 单位后缀
   */
  suffix: string
  /**
   * 保留小数位数
   */
  decimalCount: number
  /**
   * 千分符
   */
  thousandSeparator: boolean
  /**
   * 显示总出占比
   */
  showTotalPercent?: boolean
}

/**
 * 多系列格式化属性
 */
declare interface SeriesFormatter extends Axis {
  /**
   * 是否显示
   */
  show: boolean
  /**
   * 字体颜色
   */
  color?: string
  /**
   * 字体大小
   */
  fontSize?: number
  /**
   * 序列id
   */
  seriesId: string
  /**
   * 轴类型
   */
  axisType: string
  /**
   * 显示极值
   */
  showExtremum?: boolean

  optionLabel?: string
  optionShowName?: string
  /**
   * 位置
   */
  position?: string
}

declare interface Axis extends ChartViewField {
  /**
   * 格式化设置
   */
  formatterCfg: BaseFormatter
  /**
   * 聚合方式
   */
  summary: string
  /**
   * 维度/指标分组类型
   */
  groupType: 'q' | 'd'
  /**
   * 排序规则
   */
  sort: 'asc' | 'desc' | 'none' | 'custom_sort'
  /**
   * 自定义排序项
   */
  customSort: string[]
  /**
   * 是否隐藏
   */
  hide: boolean
  /**
   * 日期样式
   */
  dateStyle?: string
  /**
   * 日期格式
   */
  datePattern?: string
  /**
   * 图表类型（用于混合图）
   */
  chartType?: 'bar' | 'line'
  /**
   * 过滤条件
   */
  filter?: AxisFilter[]
  /**
   * 逻辑关系
   */
  logic?: 'and' | 'or'
  /**
   * 同比/环比计算
   */
  compareCalc?: CompareCalc
  /**
   * 过滤类型
   */
  filterType?: string
  /**
   * 枚举选中字段
   */
  enumCheckField?: string[]
  /**
   * 轴类型
   */
  axisType?: string
  /**
   * 数据源ID
   */
  datasourceId?: number | string
  /**
   * 数据集表ID
   */
  datasetTableId?: number | string
  datasetGroupId?: number | string
  /**
   * 图表ID（用于计算字段）
   */
  chartId?: string
}
declare interface ChartViewField {
  /**
   * 字段名称
   */
  name: string
  /**
   * de名称
   */
  dataeaseName: string
  /**
   * id
   */
  id: string
  /**
   * 图表自定义字段名称
   */
  chartShowName: string
  /**
   * 字段类型
   */
  deType: number
  /**
   * 分组类型
   */
  groupType: 'q' | 'd'
}

declare interface AxisFilter {
  term?: string
  value?: string | number
}

declare interface PageInfo {
  currentPage: number
  pageSize: number
  total: number
}

declare interface CompareCalc {
  type?: string
  field?: string
  custom?: string
}
