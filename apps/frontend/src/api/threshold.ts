import request from '@/config/axios'

// Threshold CRUD
export const thresholdSave = (params: any) => request.post({ url: '/threshold/save', data: params })

export const thresholdEdit = (params: any) => request.post({ url: '/threshold/edit', data: params })

export const thresholdDelete = (resourceTable: string, ids: number[]) =>
  request.post({ url: `/threshold/delete/${resourceTable}`, data: ids })

export const thresholdFormInfo = (id: string, resourceTable: string) =>
  request.get({ url: `/threshold/formInfo/${id}/${resourceTable}` })

// Threshold switch
export const thresholdSwitch = (params: any) => request.post({ url: '/threshold/switch', data: params })

// Threshold pager
export const thresholdPager = (goPage: number, pageSize: number, params: any) =>
  request.post({ url: `/threshold/pager/${goPage}/${pageSize}`, data: params })

// Threshold preview
export const thresholdPreview = (params: any) =>
  request.post({ url: '/threshold/preview', data: params })

// Threshold recipients
export const thresholdBatchReci = (params: any) =>
  request.post({ url: '/threshold/batchReci', data: params })

// Threshold instance pager
export const thresholdInstancePager = (goPage: number, pageSize: number, params: any) =>
  request.post({ url: `/threshold/instancePager/${goPage}/${pageSize}`, data: params })

// Threshold chart association
export const thresholdAnyThreshold = (chartId: string, resourceTable: string) =>
  request.get({ url: `/threshold/anyThreshold/${chartId}/${resourceTable}` })

export const thresholdDeleteWithChart = (chartId: string, resourceTable: string) =>
  request.get({ url: `/threshold/deleteWithChart/${chartId}/${resourceTable}` })
