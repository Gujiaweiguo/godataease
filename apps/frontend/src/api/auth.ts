import request from '@/config/axios'

export const queryUserApi = data => request.post({ url: '/system/user/list', data })
export const queryRoleApi = data => request.post({ url: '/role/byCurOrg', data })

export const userCreateApi = (data: any) => request.post({ url: '/system/user/create', data })
export const userUpdateApi = (data: any) => request.post({ url: '/system/user/update', data })
export const userDeleteApi = (id: number) => request.post({ url: '/system/user/delete/' + id })

export const roleCreateApi = (data: any) => request.post({ url: '/system/role/create', data })
export const roleUpdateApi = (data: any) => request.post({ url: '/system/role/update', data })
export const roleDeleteApi = (roleId: number) =>
  request.post({ url: '/system/role/delete/' + roleId })

export const resourceTreeApi = (flag: string) => request.get({ url: '/system/permission/busiResource/' + flag })

export const menuTreeApi = () => request.get({ url: '/auth/menuResource' })

export const resourcePerApi = data => request.post({ url: '/system/permission/busiPermission', data })

export const menuPerApi = data => request.post({ url: '/system/permission/menuPermission', data })

export const busiPerSaveApi = data => request.post({ url: '/system/permission/saveBusiPer', data })
export const menuPerSaveApi = data => request.post({ url: '/system/permission/saveMenuPer', data })

export const resourcePerSaveApi = data =>
  request.post({ url: '/system/role/permission/save', data })

export const resourceTargetPerApi = data =>
  request.post({ url: '/system/permission/busiTargetPermission', data })

export const userPerspectiveApi = data => request.post({ url: '/system/permission/userPerspective', data })

export const menuTargetPerApi = data => request.post({ url: '/system/permission/menuTargetPermission', data })

export const busiTargetPerSaveApi = data => request.post({ url: '/system/permission/saveBusiTargetPer', data })
export const menuTargetPerSaveApi = data => request.post({ url: '/system/permission/saveMenuTargetPer', data })

export const roleMenuAuthApi = (roleId: number) => request.get({ url: '/roleMenu/auth/' + roleId })
export const roleMenuAuthSaveApi = (data: { roleId: number; menuIds: number[] }) =>
  request.post({ url: '/roleMenu/auth', data })
