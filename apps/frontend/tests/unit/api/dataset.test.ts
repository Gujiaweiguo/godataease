import { describe, it, expect, vi, beforeEach } from 'vitest'

const { requestMock } = vi.hoisted(() => ({
  requestMock: {
    post: vi.fn().mockResolvedValue({ data: {} }),
    get: vi.fn().mockResolvedValue({ data: {} })
  }
}))

vi.mock('@/config/axios', () => ({ default: requestMock }))

import {
  getDatasourceList,
  getTables,
  getDsDetailsWithPerm,
  rowPermissionList,
  rowPermissionListByTarget,
  columnPermissionList,
  listFieldByDatasetGroup,
  multFieldValuesForPermissions,
  listFieldsWithPermissions,
  copilotFields,
  saveRowPermission,
  saveColumnPermission,
  deleteRowPermission,
  deleteColumnPermission,
  saveField,
  getFunction,
  downloadFile,
  exportDelete,
  generateDownloadUri,
  exportDeleteAll,
  exportDeletePost,
  listByDsIds,
  getFieldTree,
  copilotChat,
  getListCopilot,
  clearAllCopilot
} from '@/api/dataset'

describe('api/dataset (remaining)', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    requestMock.post.mockResolvedValue({ data: {} })
    requestMock.get.mockResolvedValue({ data: {} })
  })

  it('getDatasourceList posts to ds/tree with busiFlag datasource', async () => {
    requestMock.post.mockResolvedValueOnce({ data: [] })
    await getDatasourceList()
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/ds/tree',
      data: { busiFlag: 'datasource' }
    })
  })

  it('getDatasourceList posts to ds/tree with optional weight', async () => {
    requestMock.post.mockResolvedValueOnce({ data: [] })
    await getDatasourceList(5)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/ds/tree',
      data: { busiFlag: 'datasource', weight: 5 }
    })
  })

  it('getTables posts to ds/tables', async () => {
    requestMock.post.mockResolvedValueOnce({ data: [] })
    const data = { datasourceId: 'ds-1', tableName: 'orders' }
    await getTables(data)
    expect(requestMock.post).toHaveBeenCalledWith({ url: '/ds/tables', data })
  })

  it('getDsDetailsWithPerm posts to detailWithPerm', async () => {
    requestMock.post.mockResolvedValueOnce({ data: [] })
    await getDsDetailsWithPerm({ ids: [1] })
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/dataset/detailWithPerm',
      data: { ids: [1] }
    })
  })

  it('rowPermissionList gets rowPermissions pager', () => {
    rowPermissionList(1, 20, 5)
    expect(requestMock.get).toHaveBeenCalledWith({
      url: '/dataset/rowPermissions/pager/5/1/20'
    })
  })

  it('rowPermissionListByTarget gets rowPermissions pagerByTarget', () => {
    rowPermissionListByTarget(1, 20, 5, 'role', 3)
    expect(requestMock.get).toHaveBeenCalledWith({
      url: '/dataset/rowPermissions/pagerByTarget/5/role/3/1/20'
    })
  })

  it('columnPermissionList gets columnPermissions pager', () => {
    columnPermissionList(1, 20, 5)
    expect(requestMock.get).toHaveBeenCalledWith({
      url: '/dataset/columnPermissions/pager/5/1/20'
    })
  })

  it('listFieldByDatasetGroup posts to datasetField listByDatasetGroup', async () => {
    requestMock.post.mockResolvedValueOnce({ data: [] })
    await listFieldByDatasetGroup(10)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/datasetField/listByDatasetGroup/10'
    })
  })

  it('multFieldValuesForPermissions posts to datasetField multFieldValuesForPermissions', () => {
    const data = { fieldIds: [1] }
    multFieldValuesForPermissions(data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/datasetField/multFieldValuesForPermissions',
      data
    })
  })

  it('listFieldsWithPermissions gets datasetField listWithPermissions', async () => {
    requestMock.get.mockResolvedValueOnce({ data: [] })
    await listFieldsWithPermissions(10)
    expect(requestMock.get).toHaveBeenCalledWith({
      url: '/datasetField/listWithPermissions/10'
    })
  })

  it('copilotFields posts to datasetField copilotFields', () => {
    copilotFields(10)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/datasetField/copilotFields/10'
    })
  })

  it('saveRowPermission posts to rowPermissions save', () => {
    const data = { datasetId: 1, roleId: 2 }
    saveRowPermission(data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/dataset/rowPermissions/save',
      data
    })
  })

  it('saveColumnPermission posts to columnPermissions save', () => {
    const data = { datasetId: 1 }
    saveColumnPermission(data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/dataset/columnPermissions/save',
      data
    })
  })

  it('deleteRowPermission posts to rowPermissions delete', () => {
    const data = { id: 1 }
    deleteRowPermission(data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/dataset/rowPermissions/delete',
      data
    })
  })

  it('deleteColumnPermission posts to columnPermissions delete', () => {
    const data = { id: 1 }
    deleteColumnPermission(data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/dataset/columnPermissions/delete',
      data
    })
  })

  it('saveField posts to datasetField save', async () => {
    requestMock.post.mockResolvedValueOnce({ data: [] })
    const data = { datasetId: 1, fields: [] }
    await saveField(data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/datasetField/save',
      data
    })
  })

  it('getFunction posts to datasetField getFunction', async () => {
    requestMock.post.mockResolvedValueOnce({ data: [] })
    await getFunction()
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/datasetField/getFunction',
      data: {}
    })
  })

  it('downloadFile gets exportCenter download with blob', async () => {
    requestMock.get.mockResolvedValueOnce({ data: new Blob() })
    await downloadFile('task-1')
    expect(requestMock.get).toHaveBeenCalledWith({
      url: '/exportCenter/download/task-1',
      responseType: 'blob'
    })
  })

  it('exportDelete gets exportCenter delete', async () => {
    requestMock.get.mockResolvedValueOnce({ data: null })
    await exportDelete(5)
    expect(requestMock.get).toHaveBeenCalledWith({ url: '/exportCenter/delete/5' })
  })

  it('generateDownloadUri gets exportCenter generateDownloadUri', async () => {
    requestMock.get.mockResolvedValueOnce({ data: { uri: 'http://...' } })
    await generateDownloadUri(5)
    expect(requestMock.get).toHaveBeenCalledWith({ url: '/exportCenter/generateDownloadUri/5' })
  })

  it('exportDeleteAll posts to exportCenter deleteAll with type', async () => {
    requestMock.post.mockResolvedValueOnce({ data: null })
    await exportDeleteAll('dataset', { ids: [1] })
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/exportCenter/deleteAll/dataset',
      data: { ids: [1] }
    })
  })

  it('exportDeletePost posts to exportCenter delete', async () => {
    requestMock.post.mockResolvedValueOnce({ data: null })
    await exportDeletePost({ ids: [1] })
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/exportCenter/delete',
      data: { ids: [1] }
    })
  })

  it('listByDsIds posts to datasetField listByDsIds', async () => {
    requestMock.post.mockResolvedValueOnce({ data: [] })
    await listByDsIds({ ids: [1, 2] })
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/datasetField/listByDsIds',
      data: { ids: [1, 2] }
    })
  })

  it('getFieldTree posts to dataset fieldTree', async () => {
    requestMock.post.mockResolvedValueOnce({ data: [] })
    await getFieldTree({ ids: [1] })
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/dataset/fieldTree',
      data: { ids: [1] }
    })
  })

  it('copilotChat posts to copilot chat', async () => {
    requestMock.post.mockResolvedValueOnce({ data: {} })
    await copilotChat({ message: 'hello' })
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/copilot/chat',
      data: { message: 'hello' }
    })
  })

  it('getListCopilot posts to copilot getList', async () => {
    requestMock.post.mockResolvedValueOnce({ data: [] })
    await getListCopilot()
    expect(requestMock.post).toHaveBeenCalledWith({ url: '/copilot/getList' })
  })

  it('clearAllCopilot posts to copilot clearAll', async () => {
    requestMock.post.mockResolvedValueOnce({ data: null })
    await clearAllCopilot()
    expect(requestMock.post).toHaveBeenCalledWith({ url: '/copilot/clearAll' })
  })
})
