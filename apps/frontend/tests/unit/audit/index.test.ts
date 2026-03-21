import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

const hoisted = vi.hoisted(() => ({
  queryAuditLogsApi: vi.fn(),
  exportAuditLogsApi: vi.fn(),
  messageError: vi.fn(),
  messageSuccess: vi.fn(),
  messageInfo: vi.fn(),
  confirm: vi.fn(() => Promise.resolve())
}))

vi.mock('@/api/audit', () => ({
  queryAuditLogsApi: hoisted.queryAuditLogsApi,
  exportAuditLogsApi: hoisted.exportAuditLogsApi
}))

vi.mock('element-plus-secondary', () => ({
  ElMessage: {
    error: hoisted.messageError,
    success: hoisted.messageSuccess,
    info: hoisted.messageInfo
  },
  ElMessageBox: {
    confirm: hoisted.confirm
  }
}))

import AuditIndex from '../../../src/views/audit/index.vue'

describe('Audit page init', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('queries audit logs on mount and updates pagination metadata', async () => {
    hoisted.queryAuditLogsApi.mockResolvedValueOnce({
      code: '000000',
      data: {
        list: [{ id: 1, username: 'alice', actionName: 'Login', actionType: 'USER_ACTION' }],
        total: 1
      }
    })

    const wrapper = mount(AuditIndex, {
      global: {
        directives: {
          loading: () => undefined
        },
        stubs: {
          'el-form': true,
          'el-form-item': true,
          'el-select': true,
          'el-option': true,
          'el-date-picker': true,
          'el-button': true,
          'el-table': true,
          'el-table-column': true,
          'el-tag': true,
          'el-pagination': true
        }
      }
    })

    await flushPromises()

    expect(hoisted.queryAuditLogsApi).toHaveBeenCalledWith({})
    expect(wrapper.text()).toContain('审计日志管理')
    expect(wrapper.get('el-pagination-stub').attributes('total')).toBe('1')
    expect(wrapper.get('el-pagination-stub').attributes('current-page')).toBe('1')
    expect(wrapper.get('el-pagination-stub').attributes('page-size')).toBe('10')
  })

  it('surfaces query failure as an explicit error message', async () => {
    hoisted.queryAuditLogsApi.mockResolvedValueOnce({
      code: '500000',
      msg: '查询失败'
    })

    mount(AuditIndex, {
      global: {
        directives: {
          loading: () => undefined
        },
        stubs: {
          'el-form': true,
          'el-form-item': true,
          'el-select': true,
          'el-option': true,
          'el-date-picker': true,
          'el-button': true,
          'el-table': true,
          'el-table-column': true,
          'el-tag': true,
          'el-pagination': true
        }
      }
    })

    await flushPromises()

    expect(hoisted.messageError).toHaveBeenCalledWith('查询失败')
  })
})
