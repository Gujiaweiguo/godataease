import request from '@/config/axios'

export const queryAuditLogsApi = (params = {}) => request.get({ url: '/audit/list', params })

export const exportAuditLogsApi = (ids: number[], format = 'csv') =>
  request.post({ url: `/audit/export?format=${format}`, data: ids })
