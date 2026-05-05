import type {
  ColumnOption,
  DataFillingForm,
  DataFillingSubTask,
  RowDataDatum,
  SearchParam,
  SubTaskUserItem,
  TableDataResponse,
  TaskInfoVO,
  UserTaskData,
  UserTaskVO
} from '@/api/datafilling'

export type DataFillingNodeId = number | string

export type DataFillingFieldType =
  | 'text'
  | 'nvarchar'
  | 'number'
  | 'decimal'
  | 'date'
  | 'datetime'
  | 'select'
  | string

export type FormFieldPrimitive = string | number | boolean | null | undefined

export type FormFieldValue = FormFieldPrimitive | FormFieldPrimitive[] | Record<string, unknown>

export interface DataFillingTreeNode {
  id: DataFillingNodeId
  name: DataFillingForm['name']
  pid: DataFillingNodeId
  nodeType: DataFillingForm['nodeType']
  leaf?: boolean
  disabled?: boolean
  children?: DataFillingTreeNode[]
}

export interface FormFieldOption extends ColumnOption {
  disabled?: boolean
  description?: string
}

export interface FormFieldConfig {
  id?: DataFillingNodeId
  field?: string
  name: string
  label: string
  type: DataFillingFieldType
  required?: boolean
  placeholder?: string
  defaultValue?: FormFieldValue
  order?: number
  options?: FormFieldOption[]
  optionDatasource?: string
  optionTable?: string
  optionColumn?: string
  optionOrder?: string
  precision?: number
  format?: string
  multiple?: boolean
  extra?: Record<string, unknown>
}

export type DataFillingFormSchema = FormFieldConfig[]

export interface DataFillingColumnConfig {
  field: string
  label: string
  type?: DataFillingFieldType
  width?: number | string
  minWidth?: number | string
  align?: 'left' | 'center' | 'right'
  fixed?: boolean | 'left' | 'right'
  options?: FormFieldOption[]
  multiple?: boolean
  precision?: number
  format?: string
  searchable?: boolean
}

export interface DataFillingTableRow extends Record<string, unknown> {
  id?: string | number
}

export interface DataFillingTableView extends Omit<TableDataResponse, 'data' | 'fields'> {
  data: DataFillingTableRow[]
  fields: string[] | DataFillingFormSchema
}

export interface DataFillingTableQuery {
  currentPage: number
  pageSize: number
  searchParams: SearchParam[]
}

export interface DataFillingTaskItem extends TaskInfoVO {
  statusLabel?: string
  lastExecStatusLabel?: string
}

export interface DataFillingSubTaskItem extends DataFillingSubTask {
  execStatusLabel?: string
}

export interface DataFillingSubTaskUser extends SubTaskUserItem {
  userName?: string
}

export interface DataFillingUserTaskItem extends UserTaskVO {
  formName?: string
  statusLabel?: string
}

export interface DataFillingUserTaskDetail extends UserTaskData {
  schema?: DataFillingFormSchema
}

export interface DataFillingUploadPreview {
  uploadId: string
  rows: RowDataDatum[]
}
