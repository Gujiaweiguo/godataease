import request from '@/config/axios'

export const queryUserApi = data => request.post({ url: '/user/byCurOrg', data })
export const queryUserOptionsApi = () => request.get({ url: '/user/org/option' })
export const queryRoleApi = data => request.post({ url: '/role/byCurOrg', data })

export const userCreateApi = (data: any) => request.post({ url: '/api/system/user/create', data })
export const userUpdateApi = (data: any) => request.post({ url: '/api/system/user/update', data })
export const userDeleteApi = (id: number) => request.post({ url: '/api/system/user/delete/' + id })

export const roleCreateApi = (data: any) => request.post({ url: '/api/system/role/create', data })
export const roleUpdateApi = (data: any) => request.post({ url: '/api/system/role/update', data })
export const roleDeleteApi = (roleId: number) => request.post({ url: '/api/system/role/delete/' + roleId })

export const resourceTreeApi = (flag: string) => request.get({ url: '/auth/busiResource/' + flag })

export const menuTreeApi = () => request.get({ url: '/auth/menuResource' })

export const resourcePerApi = data => request.post({ url: '/auth/busiPermission', data })

export const menuPerApi = data => request.post({ url: '/auth/menuPermission', data })

export const busiPerSaveApi = data => request.post({ url: '/auth/saveBusiPer', data })
export const menuPerSaveApi = data => request.post({ url: '/auth/saveMenuPer', data })

export const resourcePerSaveApi = data => request.post({ url: '/api/system/role/permission/save', data })

export const resourceTargetPerApi = data =>
  request.post({ url: '/auth/busiTargetPermission', data })

export const menuTargetPerApi = data => request.post({ url: '/auth/menuTargetPermission', data })

export const busiTargetPerSaveApi = data => request.post({ url: '/auth/saveBusiTargetPer', data })
export const menuTargetPerSaveApi = data => request.post({ url: '/auth/saveMenuTargetPer', data })
