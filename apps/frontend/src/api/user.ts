import request from '@/config/axios'

export const mountedOrg = (keyword?: string) =>
  request.post({ url: '/org/mounted', data: { keyword } })

export const switchOrg = (id: number | string) => request.post({ url: `/user/switch/${id}` })

export const userInfo = () => request.get({ url: '/user/info' })

export const userOptionForRoleApi = data => request.post({ url: '/user/role/option', data })

export const userSelectedForRoleApi = (page: number, limit: number, data) =>
  request.post({ url: `/user/role/selected/${page}/${limit}`, data })

export const personInfoApi = () => request.get({ url: `/user/personInfo` })

export const ipInfoApi = () => request.get({ url: `/user/ipInfo` })

export const beforeUnmountInfoApi = data => request.post({ url: '/role/beforeUnmountInfo', data })

export const unMountUserApi = data => request.post({ url: '/role/unMountUser', data })

export const mountUserApi = data => request.post({ url: '/role/mountUser', data })

export const searchExternalUserApi = keyword =>
  request.get({ url: '/role/searchExternalUser/' + keyword })

export const mountExternalUserApi = data => request.post({ url: '/role/mountExternalUser', data })

export const switchLangApi = data => request.post({ url: '/user/switchLanguage', data })

export const defaultPwdApi = () => request.get({ url: '/user/defaultPwd' })

export const resetPwdApi = uid => request.post({ url: `/user/resetPwd/${uid}` })

export const switchEnableApi = data => request.post({ url: '/system/user/enable', data })
