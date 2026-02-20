export interface EmbeddingOuterParams {
  resourceId?: string
  dvId?: string
  chartId?: string
  busiFlag?: string
  outerParams?: Record<string, any>
  callbackParams?: Record<string, any>
  jumpInfoParam?: Record<string, any>
}

export interface OuterParamsOptions {
  encodeAsBase64?: boolean
  format?: 'json' | 'base64'
}

export interface DecodedOuterParams {
  resourceId?: string
  dvId?: string
  chartId?: string
  busiFlag?: string
  params?: Record<string, any>
  [key: string]: any
}
