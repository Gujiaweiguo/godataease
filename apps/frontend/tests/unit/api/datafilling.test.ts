import { describe, it, expect, vi, beforeEach } from 'vitest'

const { requestMock } = vi.hoisted(() => ({
  requestMock: {
    post: vi.fn().mockResolvedValue({ data: {} }),
    get: vi.fn().mockResolvedValue({ data: {} })
  }
}))

vi.mock('@/config/axios', () => ({ default: requestMock }))

import {
  getFormTree,
  getFormById,
  createForm,
  updateForm,
  renameForm,
  moveForm,
  deleteForm,
  listDatasourceList,
  listDatasourceListAll,
  getBuiltInTables,
  searchTableData,
  saveRowData,
  deleteRowData,
  batchDeleteRowData,
  truncateTableData,
  listColumnData,
  getCommitLogPage,
  clearCommitLog,
  getTaskInfo,
  saveTask,
  executeNowTask,
  getTaskPageList,
  startTask,
  stopTask,
  deleteTasks,
  getSubTaskPageList,
  deleteSubTasks,
  getSubTaskUsersList,
  getUserTaskPageList,
  getUserTaskTodoCount,
  getUserTaskData,
  saveUserTaskData,
  appendUserTaskData,
  deleteUserTaskData,
  downloadExcelTemplate,
  confirmUpload,
  userTaskConfirmUpload,
  getExtraDetails,
  getDatasourceOptions,
  getTemplateByUserTaskItem,
  exportFormData
} from '@/api/datafilling'

describe('api/datafilling', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    requestMock.post.mockResolvedValue({ data: {} })
    requestMock.get.mockResolvedValue({ data: {} })
  })

  it('getFormTree posts to tree', () => {
    getFormTree()
    expect(requestMock.post).toHaveBeenCalledWith({ url: '/data-filling/tree' })
  })

  it('getFormById gets form by id', () => {
    getFormById(5)
    expect(requestMock.get).toHaveBeenCalledWith({ url: '/data-filling/get/5' })
  })

  it('createForm posts to save', () => {
    const data = { name: 'Form1', pid: 0, nodeType: 'folder' }
    createForm(data)
    expect(requestMock.post).toHaveBeenCalledWith({ url: '/data-filling/save', data })
  })

  it('updateForm posts to update', () => {
    const data = { id: 1, name: 'Updated', pid: 0, nodeType: 'folder' }
    updateForm(data)
    expect(requestMock.post).toHaveBeenCalledWith({ url: '/data-filling/update', data })
  })

  it('renameForm posts to rename', () => {
    const data = { id: 1, name: 'NewName' }
    renameForm(data)
    expect(requestMock.post).toHaveBeenCalledWith({ url: '/data-filling/rename', data })
  })

  it('moveForm posts to move', () => {
    const data = { id: 1, pid: 2 }
    moveForm(data)
    expect(requestMock.post).toHaveBeenCalledWith({ url: '/data-filling/move', data })
  })

  it('deleteForm gets delete by id', () => {
    deleteForm(3)
    expect(requestMock.get).toHaveBeenCalledWith({ url: '/data-filling/delete/3' })
  })

  it('listDatasourceList gets datasource list', () => {
    listDatasourceList()
    expect(requestMock.get).toHaveBeenCalledWith({ url: '/data-filling/datasource/list' })
  })

  it('listDatasourceListAll gets datasource listAll', () => {
    listDatasourceListAll()
    expect(requestMock.get).toHaveBeenCalledWith({ url: '/data-filling/datasource/listAll' })
  })

  it('getBuiltInTables posts to getBuiltInTables', () => {
    const data = { keyword: 'test' }
    getBuiltInTables(data)
    expect(requestMock.post).toHaveBeenCalledWith({ url: '/data-filling/getBuiltInTables', data })
  })

  it('searchTableData posts to form tableData', () => {
    const data = { id: 10, currentPage: 1, pageSize: 20, searchParams: [] }
    searchTableData(data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/data-filling/form/10/tableData',
      data
    })
  })

  it('saveRowData posts to form rowData save', () => {
    const data = { name: 'row1' }
    saveRowData(5, data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/data-filling/form/5/rowData/save',
      data
    })
  })

  it('deleteRowData gets form delete with formId and id', () => {
    deleteRowData(5, 'abc-123')
    expect(requestMock.get).toHaveBeenCalledWith({
      url: '/data-filling/form/5/delete/abc-123'
    })
  })

  it('batchDeleteRowData posts to form batch-delete', () => {
    const data = { ids: ['a', 'b'] }
    batchDeleteRowData(5, data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/data-filling/form/5/batch-delete',
      data
    })
  })

  it('truncateTableData gets form truncate', () => {
    truncateTableData(7)
    expect(requestMock.get).toHaveBeenCalledWith({ url: '/data-filling/form/7/truncate' })
  })

  it('listColumnData posts to form listColumnData', () => {
    const data = { columnName: 'col1' }
    listColumnData(5, data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/data-filling/form/5/listColumnData',
      data
    })
  })

  it('getCommitLogPage posts to log page with pagination', () => {
    const data = { formId: 1 }
    getCommitLogPage(1, 20, data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/data-filling/log/page/1/20',
      data
    })
  })

  it('clearCommitLog posts to log clear', () => {
    const data = { formId: 1, clearType: 'all' }
    clearCommitLog(data)
    expect(requestMock.post).toHaveBeenCalledWith({ url: '/data-filling/log/clear', data })
  })

  it('getTaskInfo gets task info by id', () => {
    getTaskInfo(10)
    expect(requestMock.get).toHaveBeenCalledWith({ url: '/data-filling/task/info/10' })
  })

  it('saveTask posts to task save', () => {
    const data = { formId: 1, name: 'Task1' } as any
    saveTask(data)
    expect(requestMock.post).toHaveBeenCalledWith({ url: '/data-filling/task/save', data })
  })

  it('executeNowTask posts to task executeNow', () => {
    const data = { taskId: 5 }
    executeNowTask(data)
    expect(requestMock.post).toHaveBeenCalledWith({ url: '/data-filling/task/executeNow', data })
  })

  it('getTaskPageList posts to form task page with pagination', () => {
    getTaskPageList(1, 2, 10)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/data-filling/form/1/task/page/2/10'
    })
  })

  it('startTask gets form task start', () => {
    startTask(1, 5)
    expect(requestMock.get).toHaveBeenCalledWith({
      url: '/data-filling/form/1/task/5/start'
    })
  })

  it('stopTask gets form task stop', () => {
    stopTask(1, 5)
    expect(requestMock.get).toHaveBeenCalledWith({
      url: '/data-filling/form/1/task/5/stop'
    })
  })

  it('deleteTasks posts to form task delete', () => {
    const data = { ids: [1, 2] }
    deleteTasks(1, data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/data-filling/form/1/task/delete',
      data
    })
  })

  it('getSubTaskPageList posts to sub-task page with pagination', () => {
    const data = { taskId: 1 }
    getSubTaskPageList(1, 10, data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/data-filling/sub-task/page/1/10',
      data
    })
  })

  it('deleteSubTasks posts to form sub-task delete', () => {
    const data = { ids: [1, 2] }
    deleteSubTasks(1, data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/data-filling/form/1/sub-task/delete',
      data
    })
  })

  it('getSubTaskUsersList gets sub-task users list', () => {
    getSubTaskUsersList(5, 'filler')
    expect(requestMock.get).toHaveBeenCalledWith({
      url: '/data-filling/sub-task/5/users/list/filler'
    })
  })

  it('getUserTaskPageList posts to user-task page with pagination', () => {
    const data = { taskName: 'test' }
    getUserTaskPageList(1, 20, data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/data-filling/user-task/page/1/20',
      data
    })
  })

  it('getUserTaskTodoCount posts to user-task todo count', () => {
    getUserTaskTodoCount()
    expect(requestMock.post).toHaveBeenCalledWith({ url: '/data-filling/user-task/todo/count' })
  })

  it('getUserTaskData gets user-task list by subTaskId', () => {
    getUserTaskData(5)
    expect(requestMock.get).toHaveBeenCalledWith({ url: '/data-filling/user-task/list/5' })
  })

  it('saveUserTaskData posts to user-task saveData with subTaskId', () => {
    const data = [{ name: 'row1' }]
    saveUserTaskData(5, data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/data-filling/user-task/saveData/5',
      data: { subTaskId: 5, data }
    })
  })

  it('appendUserTaskData posts to user-task appendData with subTaskId', () => {
    const data = [{ name: 'row2' }]
    appendUserTaskData(5, data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/data-filling/user-task/appendData/5',
      data: { subTaskId: 5, data }
    })
  })

  it('deleteUserTaskData gets user-task deleteData', () => {
    deleteUserTaskData(3, 'data-1')
    expect(requestMock.get).toHaveBeenCalledWith({
      url: '/data-filling/user-task/3/deleteData/data-1'
    })
  })

  it('downloadExcelTemplate gets form excelTemplate', () => {
    downloadExcelTemplate(5)
    expect(requestMock.get).toHaveBeenCalledWith({
      url: '/data-filling/form/5/excelTemplate',
      responseType: 'blob'
    })
  })

  it('confirmUpload posts to form confirmUpload', () => {
    confirmUpload(5, 'upload-123')
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/data-filling/form/5/confirmUpload',
      data: { id: 'upload-123' }
    })
  })

  it('userTaskConfirmUpload posts to user-task appendData form confirmUpload', () => {
    userTaskConfirmUpload(5, 10, 'upload-456')
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/data-filling/user-task/appendData/5/form/10/confirmUpload',
      data: { id: 'upload-456' }
    })
  })

  it('getExtraDetails posts to form extraDetails', () => {
    const data = { value: 'test', extraColumns: [], optionDatasource: 'ds', optionTable: 't', optionColumn: 'c' }
    getExtraDetails(data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/data-filling/form/extraDetails',
      data
    })
  })

  it('getDatasourceOptions posts to form options', () => {
    const data = { optionTable: 't', optionColumn: 'c', optionOrder: 'asc' }
    getDatasourceOptions(5, data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/data-filling/form/5/options',
      data
    })
  })

  it('getTemplateByUserTaskItem gets template by itemId', () => {
    getTemplateByUserTaskItem(7)
    expect(requestMock.get).toHaveBeenCalledWith({ url: '/data-filling/template/7' })
  })

  it('exportFormData posts to innerExport with isDataEaseBi=false', () => {
    exportFormData(5)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/data-filling/innerExport/0/5',
      responseType: 'blob'
    })
  })

  it('exportFormData posts to innerExport with isDataEaseBi=true', () => {
    exportFormData(5, true)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/data-filling/innerExport/1/5',
      responseType: 'blob'
    })
  })
})
