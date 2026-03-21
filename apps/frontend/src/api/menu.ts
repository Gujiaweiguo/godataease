import request from '@/config/axios'

const menuPathReg = /^\/?[A-Za-z0-9/_-]+$/
const componentPathReg = /^[A-Za-z0-9/_-]+$/

export const MENU_VALIDATION_MESSAGES = {
  nameRequired: '菜单名称不能为空',
  pathRequired: '菜单路径不能为空',
  pathInvalid: '菜单路径仅支持字母、数字、/、_、-',
  componentInvalid: '组件路径仅支持字母、数字、/、_、-',
  iconTooLong: '图标名称长度不能超过 128 个字符',
  sortInvalid: '排序值必须为非负整数',
  hiddenInvalid: '隐藏标记格式错误'
} as const

export type MenuLocation = 'sidebar' | 'user_menu' | 'help_menu'
export type MenuType = 'link' | 'action' | 'separator'

export interface MenuActionConfig {
  event?: string
  url?: string
  target?: '_blank' | '_self'
  [key: string]: unknown
}

export interface MenuSavePayload {
  id?: number
  pid: number
  type: number
  name: string
  component?: string
  menuSort: number
  icon?: string
  path: string
  hidden: boolean
  inLayout: boolean
  auth: boolean
  menuLocation?: MenuLocation
  menuType?: MenuType
  actionConfig?: MenuActionConfig
}

export interface MenuValidationResult {
  valid: boolean
  message?: string
}

export const validateMenuPayload = (payload: MenuSavePayload): MenuValidationResult => {
  if (!payload.name?.trim()) {
    return { valid: false, message: MENU_VALIDATION_MESSAGES.nameRequired }
  }
  if (!payload.path?.trim()) {
    return { valid: false, message: MENU_VALIDATION_MESSAGES.pathRequired }
  }
  if (!menuPathReg.test(payload.path.trim())) {
    return { valid: false, message: MENU_VALIDATION_MESSAGES.pathInvalid }
  }
  if (payload.component && !componentPathReg.test(payload.component.trim())) {
    return { valid: false, message: MENU_VALIDATION_MESSAGES.componentInvalid }
  }
  if ((payload.icon || '').length > 128) {
    return { valid: false, message: MENU_VALIDATION_MESSAGES.iconTooLong }
  }
  if (!Number.isInteger(payload.menuSort) || payload.menuSort < 0) {
    return { valid: false, message: MENU_VALIDATION_MESSAGES.sortInvalid }
  }
  if (typeof payload.hidden !== 'boolean') {
    return { valid: false, message: MENU_VALIDATION_MESSAGES.hiddenInvalid }
  }
  return { valid: true }
}

export const menuQueryApi = () => request.get({ url: '/menu/query' })

export const menuDetailApi = (id: number) => request.get({ url: `/menu/detail/${id}` })

export const menuCreateApi = (data: MenuSavePayload) => request.post({ url: '/menu/create', data })

export const menuUpdateApi = (data: MenuSavePayload) => request.post({ url: '/menu/update', data })

export const menuDeleteApi = (id: number) => request.post({ url: `/menu/delete/${id}` })

export const menuUpdateSortApi = (data: { id: number; sort: number }) =>
  request.post({ url: '/menu/updateSort', data })

export const menuUpdateHiddenApi = (data: { id: number; hidden: boolean }) =>
  request.post({ url: '/menu/updateHidden', data })
