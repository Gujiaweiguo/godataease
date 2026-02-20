import request from '@/config/axios'

export const embeddedQueryGridApi = (page: number, pageSize: number, data) =>
  request.post({ url: `/embedded/pager/${page}/${pageSize}`, data })

export const embeddedCreateApi = data => request.post({ url: '/embedded/create', data })

export const embeddedEditApi = data => request.post({ url: '/embedded/edit', data })

export const embeddedDeleteApi = id => request.post({ url: `/embedded/delete/${id}` })

export const embeddedBatchDeleteApi = ids =>
  request.post({ url: '/embedded/batchDelete', data: ids })

export const embeddedResetApi = data => request.post({ url: '/embedded/reset', data })

export const embeddedDomainListApi = () => request.get({ url: '/embedded/domainList' })

export const embeddedInitIframeApi = data => request.post({ url: '/embedded/initIframe', data })

export const embeddedGetTokenArgsApi = () => request.get({ url: '/embedded/getTokenArgs' })
