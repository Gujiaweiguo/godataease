import request from '@/config/axios'

export const orgListApi = (params?: any) => request.get({ url: '/api/system/organization/list' })
export const orgCreateApi = (data: any) =>
  request.post({ url: '/api/system/organization/create', data })
export const orgUpdateApi = (data: any) =>
  request.post({ url: '/api/system/organization/update', data })
export const orgDeleteApi = (id: number) =>
  request.post({ url: '/api/system/organization/delete/' + id })
export const orgTreeApi = () => request.get({ url: '/api/system/organization/tree' })
export const queryUserOptionsApi = () => request.get({ url: '/user/org/option' })

export const permListApi = (params?: any) =>
  request.post({ url: '/api/system/permission/list', data: params || {} })
export const permCreateApi = (data: any) =>
  request.post({ url: '/api/system/permission/create', data })
export const permUpdateApi = (data: any) =>
  request.post({ url: '/api/system/permission/update', data })
export const permDeleteApi = (id: number) =>
  request.post({ url: '/api/system/permission/delete/' + id })
