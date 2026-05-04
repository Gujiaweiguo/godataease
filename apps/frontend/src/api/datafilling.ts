import request from '@/config/axios'

// ===================== Form Types =====================
export interface DataFillingForm {
  id: number
  name: string
  pid: number
  nodeType: string
  tableName: string
  datasourceId: number
  forms: string
  createIndex: boolean
  tableIndexes: string
  useExistsTable: boolean
  createBy: number
  createTime: number
  updateBy: number
  updateTime: number
}

export interface CreateFormRequest {
  name: string
  pid: number
  nodeType: string
  tableName?: string
  datasourceId?: number
  forms?: string
  createIndex?: boolean
  tableIndexes?: string
  useExistsTable?: boolean
}

export interface UpdateFormRequest extends CreateFormRequest {
  id: number
}

// ===================== Table Data Types =====================
export interface SearchParam {
  term: string
  field: string
  value: any
  values: any[]
  multiple: boolean
}

export interface TableDataRequest {
  id: number
  currentPage: number
  pageSize: number
  searchParams: SearchParam[]
}

export interface TableDataResponse {
  data: Record<string, any>[]
  fields: string
  total: number
  currentPage: number
  pageSize: number
  key: string
}

// ===================== Commit Log Types =====================
export interface DfCommitLog {
  id: number
  formId: number
  dataId: string
  pid: number
  operate: number
  commitBy: number
  committer: string
  commitTime: number
  count: number
}

export interface CommitLogPageRequest {
  formId: number
  operate?: number
}

export interface ClearCommitLogRequest {
  formId: number
  clearType?: string
}

// ===================== Task Types =====================
export interface TaskSaveRequest {
  id?: number
  formId: number
  name: string
  reciFlagList: number[]
  uidList: number[]
  ridList: number[]
  fillType: number
  fitType: number
  fitColumn: string
  rateType: number
  rateVal: string
  oneTimeType: number
  startTime: number
  endTime: number
  publishRangeTime: number
  publishRangeTimeType: number
  formExtSetting: string
  formFilterSetting: string
}

export interface TaskInfoVO {
  id: number
  formId: number
  name: string
  reciFlagList: number[]
  uidList: number[]
  ridList: number[]
  fillType: number
  fitType: number
  fitColumn: string
  rateType: number
  rateVal: string
  oneTimeType: number
  startTime: number
  endTime: number
  publishRangeTime: number
  publishRangeTimeType: number
  status: number
  lastExecStatus: number
  lastExecTime: number
  nextExecTime: number
  createBy: number
  createTime: number
  updateBy: number
  updateTime: number
  formExtSetting: string
  formFilterSetting: string
}

export interface TaskPageResponse {
  records: TaskInfoVO[]
  total: number
  current: number
  size: number
}

export interface BatchDeleteTaskRequest {
  ids: number[]
}

// ===================== SubTask Types =====================
export interface DataFillingSubTask {
  id: number
  taskId: number
  startTime: number
  endTime: number
  execStatus: number
  status: number
  totalCount: number
  unfinishedCount: number
  totalUserCount: number
  unfinishedUserCount: number
  fillType: number
}

export interface SubTaskPageResponse {
  records: DataFillingSubTask[]
  total: number
  current: number
  size: number
}

export interface SubTaskUserItem {
  id: number
  taskId: number
  pid: number
  uid: number
  formId: number
  dataId: string
  finishTime: number
  status: number
}

// ===================== User Task Types =====================
export interface UserTaskPageRequest {
  type?: number
  taskName: string
}

export interface UserTaskVO {
  id: number
  taskId: number
  taskName: string
  formId: number
  startTime: number
  endTime: number
  status: number
  finishTime?: number
  fillType: number
  expired: boolean
  totalCount: number
  finishCount: number
}

export interface SubInstanceItem {
  id: number
  taskId: number
  pid: number
  uid: number
  formId: number
  dataId: string
  finishTime?: number
  status: number
}

export interface UserTaskData {
  formId: number
  formTitle: string
  dataIds: string[]
  subInstances: SubInstanceItem[]
  form: string
  formExtSetting: string
  fillType: number
}

// ===================== Excel Types =====================
export interface RowDataDatum {
  id: string
  data: Record<string, any>
  insert: boolean
}

export interface DfExcelData {
  formFields: Record<string, any>[]
  dataList: RowDataDatum[]
  id: string
  excelName: string
  path: string
  suffix: string
}

// ===================== Extra Details Types =====================
export interface ExtraDetailsRequest {
  optionDatasource: string
  optionTable: string
  optionColumn: string
  extraColumns: string[]
  value: string
}

export interface ExtraDetails {
  name: string
  value: string
}

export interface DatasourceOptionsRequest {
  optionTable: string
  optionColumn: string
  optionOrder: string
}

export interface ColumnOption {
  name: string
  value: string
}

const BASE = '/data-filling'
const unwrap = <T>(res: IResponse<T>) => res?.data

// ===================== Form Management =====================
export const getFormTree = () => request.post({ url: BASE + '/tree' }).then(unwrap)

export const getFormById = (id: number) => request.get({ url: BASE + '/get/' + id }).then(unwrap)

export const createForm = (data: CreateFormRequest) =>
  request.post({ url: BASE + '/save', data }).then(unwrap)

export const updateForm = (data: UpdateFormRequest) =>
  request.post({ url: BASE + '/update', data }).then(unwrap)

export const renameForm = (data: { id: number; name: string }) => request.post({ url: BASE + '/rename', data })

export const moveForm = (data: { id: number; pid: number }) => request.post({ url: BASE + '/move', data })

export const deleteForm = (id: number) => request.get({ url: BASE + '/delete/' + id })

export const listDatasourceList = () => request.get({ url: BASE + '/datasource/list' }).then(unwrap)

export const listDatasourceListAll = () =>
  request.get({ url: BASE + '/datasource/listAll' }).then(unwrap)

export const getBuiltInTables = (data = {}) =>
  request.post({ url: BASE + '/getBuiltInTables', data }).then(unwrap)

// ===================== Table Data (DML) =====================
export const searchTableData = (data: TableDataRequest) =>
  request.post({ url: BASE + '/form/' + data.id + '/tableData', data }).then(unwrap)

export const saveRowData = (formId: number, data: Record<string, any>) =>
  request.post({ url: BASE + '/form/' + formId + '/rowData/save', data }).then(unwrap)

export const deleteRowData = (formId: number, id: string) =>
  request.get({ url: BASE + '/form/' + formId + '/delete/' + id })

export const batchDeleteRowData = (formId: number, data: { ids: string[] }) =>
  request.post({ url: BASE + '/form/' + formId + '/batch-delete', data })

export const truncateTableData = (formId: number) => request.get({ url: BASE + '/form/' + formId + '/truncate' })

export const listColumnData = (formId: number, data: { columnName: string }) =>
  request.post({ url: BASE + '/form/' + formId + '/listColumnData', data }).then(unwrap)

// ===================== Commit Log =====================
export const getCommitLogPage = (page: number, pageSize: number, data: CommitLogPageRequest) =>
  request.post({ url: BASE + '/log/page/' + page + '/' + pageSize, data }).then(unwrap)

export const clearCommitLog = (data: ClearCommitLogRequest) => request.post({ url: BASE + '/log/clear', data })

// ===================== Task Management =====================
export const getTaskInfo = (taskId: number) =>
  request.get({ url: BASE + '/task/info/' + taskId }).then(unwrap)

export const saveTask = (data: TaskSaveRequest) => request.post({ url: BASE + '/task/save', data }).then(unwrap)

export const executeNowTask = (data: { taskId: number }) => request.post({ url: BASE + '/task/executeNow', data })

export const getTaskPageList = (formId: number, page: number, pageSize: number) =>
  request.post({ url: BASE + '/form/' + formId + '/task/page/' + page + '/' + pageSize }).then(unwrap)

export const startTask = (formId: number, id: number) =>
  request.get({ url: BASE + '/form/' + formId + '/task/' + id + '/start' })

export const stopTask = (formId: number, id: number) =>
  request.get({ url: BASE + '/form/' + formId + '/task/' + id + '/stop' })

export const deleteTasks = (formId: number, data: BatchDeleteTaskRequest) =>
  request.post({ url: BASE + '/form/' + formId + '/task/delete', data })

// ===================== SubTask Management =====================
export const getSubTaskPageList = (page: number, pageSize: number, data: { taskId: number }) =>
  request.post({ url: BASE + '/sub-task/page/' + page + '/' + pageSize, data }).then(unwrap)

export const deleteSubTasks = (formId: number, data: { ids: number[] }) =>
  request.post({ url: BASE + '/form/' + formId + '/sub-task/delete', data })

export const getSubTaskUsersList = (subTaskId: number, type: string) =>
  request.get({ url: BASE + '/sub-task/' + subTaskId + '/users/list/' + type }).then(unwrap)

// ===================== User Task (Filler) =====================
export const getUserTaskPageList = (page: number, pageSize: number, data: UserTaskPageRequest) =>
  request.post({ url: BASE + '/user-task/page/' + page + '/' + pageSize, data }).then(unwrap)

export const getUserTaskTodoCount = () => request.post({ url: BASE + '/user-task/todo/count' }).then(unwrap)

export const getUserTaskData = (subTaskId: number) =>
  request.get({ url: BASE + '/user-task/list/' + subTaskId }).then(unwrap)

export const saveUserTaskData = (subTaskId: number, data: Record<string, any>[]) =>
  request.post({ url: BASE + '/user-task/saveData/' + subTaskId, data: { subTaskId, data } })

export const appendUserTaskData = (subTaskId: number, data: Record<string, any>[]) =>
  request.post({ url: BASE + '/user-task/appendData/' + subTaskId, data: { subTaskId, data } })

export const deleteUserTaskData = (taskInstanceId: number, dataId: string) =>
  request.get({ url: BASE + '/user-task/' + taskInstanceId + '/deleteData/' + dataId })

// ===================== Excel =====================
export const downloadExcelTemplate = (formId: number): Promise<Blob> =>
  request.get({ url: BASE + '/form/' + formId + '/excelTemplate', responseType: 'blob' }).then(unwrap)

export const uploadExcelFile = (formId: number, file: File) => {
  const formData = new FormData()
  formData.append('file', file)
  return request
    .post({
      url: BASE + '/form/' + formId + '/uploadFile',
      data: formData,
      headersType: 'multipart/form-data;'
    })
    .then(unwrap) as Promise<DfExcelData>
}

export const confirmUpload = (formId: number, uploadId: string) =>
  request.post({ url: BASE + '/form/' + formId + '/confirmUpload', data: { id: uploadId } })

export const userTaskConfirmUpload = (subTaskId: number, formId: number, uploadId: string) =>
  request.post({
    url: BASE + '/user-task/appendData/' + subTaskId + '/form/' + formId + '/confirmUpload',
    data: { id: uploadId }
  })

// ===================== Extra Details & Options =====================
export const getExtraDetails = (data: ExtraDetailsRequest) =>
  request.post({ url: BASE + '/form/extraDetails', data }).then(unwrap)

export const getDatasourceOptions = (formId: number, data: DatasourceOptionsRequest) =>
  request.post({ url: BASE + '/form/' + formId + '/options', data }).then(unwrap)

// ===================== Template & Export =====================
export const getTemplateByUserTaskItem = (itemId: number) =>
  request.get({ url: BASE + '/template/' + itemId }).then(unwrap)

export const exportFormData = (formId: number, isDataEaseBi = false): Promise<Blob> =>
  request
    .post({ url: BASE + '/innerExport/' + (isDataEaseBi ? 1 : 0) + '/' + formId, responseType: 'blob' })
    .then(unwrap)
