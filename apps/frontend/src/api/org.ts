import request from '@/config/axios'

export const orgListApi = () => request.get({ url: '/system/organization/list' })
export const orgCreateApi = (data: any) =>
  request.post({ url: '/system/organization/create', data })
export const orgUpdateApi = (data: any) =>
  request.post({ url: '/system/organization/update', data })
export const orgDeleteApi = (id: number) =>
  request.post({ url: '/system/organization/delete/' + id })
export const orgTreeApi = () => request.get({ url: '/system/organization/tree' })
export const queryUserOptionsApi = () => request.get({ url: '/system/user/options' })

export const permListApi = (params?: any) =>
  request.post({ url: '/system/permission/list', data: params || {} })
export const permCreateApi = (data: any) => request.post({ url: '/system/permission/create', data })
export const permUpdateApi = (data: any) => request.post({ url: '/system/permission/update', data })
export const permDeleteApi = (id: number) =>
  request.post({ url: '/system/permission/delete/' + id })
