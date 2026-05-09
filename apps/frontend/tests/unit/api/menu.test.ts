import { describe, it, expect, vi, beforeEach } from 'vitest'

const { requestMock } = vi.hoisted(() => ({
  requestMock: {
    post: vi.fn().mockResolvedValue({ data: {} }),
    get: vi.fn().mockResolvedValue({ data: {} })
  }
}))

vi.mock('@/config/axios', () => ({ default: requestMock }))

import {
  menuQueryApi,
  menuDetailApi,
  menuCreateApi,
  menuUpdateApi,
  menuDeleteApi,
  menuUpdateSortApi,
  menuUpdateHiddenApi,
  validateMenuPayload,
  MENU_VALIDATION_MESSAGES
} from '@/api/menu'

describe('api/menu', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('menuQueryApi gets menu query', () => {
    menuQueryApi()
    expect(requestMock.get).toHaveBeenCalledWith({ url: '/menu/query' })
  })

  it('menuDetailApi gets menu detail with id', () => {
    menuDetailApi(5)
    expect(requestMock.get).toHaveBeenCalledWith({ url: '/menu/detail/5' })
  })

  it('menuCreateApi posts to menu create', () => {
    const data = {
      pid: 0,
      type: 0,
      name: 'TestMenu',
      path: '/test',
      menuSort: 0,
      hidden: false,
      inLayout: false,
      auth: false
    }
    menuCreateApi(data)
    expect(requestMock.post).toHaveBeenCalledWith({ url: '/menu/create', data })
  })

  it('menuUpdateApi posts to menu update', () => {
    const data = {
      id: 1,
      pid: 0,
      type: 0,
      name: 'Updated',
      path: '/updated',
      menuSort: 1,
      hidden: false,
      inLayout: false,
      auth: false
    }
    menuUpdateApi(data)
    expect(requestMock.post).toHaveBeenCalledWith({ url: '/menu/update', data })
  })

  it('menuDeleteApi posts to menu delete with id in url', () => {
    menuDeleteApi(3)
    expect(requestMock.post).toHaveBeenCalledWith({ url: '/menu/delete/3' })
  })

  it('menuUpdateSortApi posts to menu updateSort', () => {
    const data = { id: 1, sort: 5 }
    menuUpdateSortApi(data)
    expect(requestMock.post).toHaveBeenCalledWith({ url: '/menu/updateSort', data })
  })

  it('menuUpdateHiddenApi posts to menu updateHidden', () => {
    const data = { id: 1, hidden: true }
    menuUpdateHiddenApi(data)
    expect(requestMock.post).toHaveBeenCalledWith({ url: '/menu/updateHidden', data })
  })

  describe('validateMenuPayload', () => {
    const validPayload = {
      pid: 0,
      type: 0,
      name: 'Test',
      path: '/test',
      menuSort: 0,
      hidden: false,
      inLayout: false,
      auth: false
    }

    it('returns valid for correct payload', () => {
      expect(validateMenuPayload(validPayload)).toEqual({ valid: true })
    })

    it('rejects empty name', () => {
      const result = validateMenuPayload({ ...validPayload, name: '' })
      expect(result).toEqual({ valid: false, message: MENU_VALIDATION_MESSAGES.nameRequired })
    })

    it('rejects whitespace-only name', () => {
      const result = validateMenuPayload({ ...validPayload, name: '   ' })
      expect(result).toEqual({ valid: false, message: MENU_VALIDATION_MESSAGES.nameRequired })
    })

    it('rejects empty path', () => {
      const result = validateMenuPayload({ ...validPayload, path: '' })
      expect(result).toEqual({ valid: false, message: MENU_VALIDATION_MESSAGES.pathRequired })
    })

    it('rejects invalid path characters', () => {
      const result = validateMenuPayload({ ...validPayload, path: '/test path!' })
      expect(result).toEqual({ valid: false, message: MENU_VALIDATION_MESSAGES.pathInvalid })
    })

    it('rejects invalid component path characters', () => {
      const result = validateMenuPayload({ ...validPayload, component: 'comp name!' })
      expect(result).toEqual({ valid: false, message: MENU_VALIDATION_MESSAGES.componentInvalid })
    })

    it('rejects icon name over 128 chars', () => {
      const result = validateMenuPayload({ ...validPayload, icon: 'x'.repeat(129) })
      expect(result).toEqual({ valid: false, message: MENU_VALIDATION_MESSAGES.iconTooLong })
    })

    it('rejects negative menuSort', () => {
      const result = validateMenuPayload({ ...validPayload, menuSort: -1 })
      expect(result).toEqual({ valid: false, message: MENU_VALIDATION_MESSAGES.sortInvalid })
    })

    it('rejects non-integer menuSort', () => {
      const result = validateMenuPayload({ ...validPayload, menuSort: 1.5 as any })
      expect(result).toEqual({ valid: false, message: MENU_VALIDATION_MESSAGES.sortInvalid })
    })

    it('rejects non-boolean hidden', () => {
      const result = validateMenuPayload({ ...validPayload, hidden: 'yes' as any })
      expect(result).toEqual({ valid: false, message: MENU_VALIDATION_MESSAGES.hiddenInvalid })
    })
  })
})
