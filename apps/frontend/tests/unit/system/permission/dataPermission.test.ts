import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, nextTick } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'

const hoisted = vi.hoisted(() => ({
  queryUserApi: vi.fn(),
  queryRoleApi: vi.fn(),
  getDatasetTree: vi.fn(),
  listFieldByDatasetGroup: vi.fn(),
  rowPermissionList: vi.fn(),
  rowPermissionListByTarget: vi.fn(),
  columnPermissionList: vi.fn(),
  saveRowPermission: vi.fn(),
  saveColumnPermission: vi.fn(),
  deleteRowPermission: vi.fn(),
  deleteColumnPermission: vi.fn(),
  messageSuccess: vi.fn(),
  messageError: vi.fn(),
  messageWarning: vi.fn()
}))

vi.mock('@/api/auth', () => ({
  queryUserApi: hoisted.queryUserApi,
  queryRoleApi: hoisted.queryRoleApi
}))

vi.mock('@/api/dataset', () => ({
  getDatasetTree: hoisted.getDatasetTree,
  listFieldByDatasetGroup: hoisted.listFieldByDatasetGroup,
  rowPermissionList: hoisted.rowPermissionList,
  rowPermissionListByTarget: hoisted.rowPermissionListByTarget,
  columnPermissionList: hoisted.columnPermissionList,
  saveRowPermission: hoisted.saveRowPermission,
  saveColumnPermission: hoisted.saveColumnPermission,
  deleteRowPermission: hoisted.deleteRowPermission,
  deleteColumnPermission: hoisted.deleteColumnPermission
}))

vi.mock('element-plus-secondary', () => ({
  ElMessage: {
    success: hoisted.messageSuccess,
    error: hoisted.messageError,
    warning: hoisted.messageWarning
  },
  ElMessageBox: {
    confirm: vi.fn(() => Promise.resolve())
  }
}))

import DataPermission from '../../../../src/views/system/permission/DataPermission.vue'

const passthroughStub = defineComponent({
  template: '<div><slot /></div>'
})

const dialogStub = defineComponent({
  props: {
    modelValue: {
      type: Boolean,
      default: false
    }
  },
  template: '<div><slot /><slot name="footer" /></div>'
})

const optionStub = defineComponent({
  props: {
    label: {
      type: String,
      default: ''
    },
    value: {
      type: [String, Number],
      default: ''
    }
  },
  template: '<div class="option-stub" :data-label="label" :data-value="value" />'
})

const tableColumnStub = defineComponent({
  template: '<div />'
})

describe('DataPermission column mask keep-middle exposure', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    hoisted.getDatasetTree.mockResolvedValue([{ id: 1, name: '示例数据集', leaf: true }])
    hoisted.queryUserApi.mockResolvedValue({ code: '000000', data: { list: [] } })
    hoisted.queryRoleApi.mockResolvedValue({ code: '000000', data: { list: [] } })
    hoisted.listFieldByDatasetGroup.mockResolvedValue({
      code: '000000',
      data: [{ originName: 'phone', fieldType: 'STRING' }]
    })
    hoisted.rowPermissionList.mockResolvedValue({ code: '000000', data: { list: [] } })
    hoisted.rowPermissionListByTarget.mockResolvedValue({ code: '000000', data: { list: [] } })
    hoisted.columnPermissionList.mockResolvedValue({ code: '000000', data: { list: [] } })
    hoisted.saveRowPermission.mockResolvedValue({ code: '000000' })
    hoisted.saveColumnPermission.mockResolvedValue({ code: '000000' })
    hoisted.deleteRowPermission.mockResolvedValue({ code: '000000' })
    hoisted.deleteColumnPermission.mockResolvedValue({ code: '000000' })
  })

  it('shows keep_middle option and submits keep_middle maskRule', async () => {
    const wrapper = mount(DataPermission, {
      global: {
        directives: {
          loading: () => undefined
        },
        stubs: {
          'el-radio-group': passthroughStub,
          'el-radio-button': passthroughStub,
          'el-select': passthroughStub,
          'el-option': optionStub,
          'el-button': passthroughStub,
          'el-empty': true,
          'el-tabs': passthroughStub,
          'el-tab-pane': passthroughStub,
          'el-table': passthroughStub,
          'el-table-column': tableColumnStub,
          'el-tag': true,
          'el-dialog': dialogStub,
          'el-form': passthroughStub,
          'el-form-item': passthroughStub,
          'el-input': true,
          'el-input-number': true,
          'el-radio': passthroughStub
        }
      }
    })

    await flushPromises()

    const setupState = (wrapper.vm as any).$.setupState
    setupState.selectedDatasetId = 1
    setupState.activeTab = 'column'
    await nextTick()

    setupState.handleAddRule()
    setupState.columnRuleForm.ruleType = 'mask'
    setupState.columnRuleForm.maskRule = 'keep_middle'
    await nextTick()

    const keepMiddleOption = wrapper.find('.option-stub[data-value="keep_middle"]')
    expect(keepMiddleOption.attributes('data-label')).toBe('保留中间')

    await setupState.handleColumnRuleSubmit()
    await flushPromises()

    expect(hoisted.saveColumnPermission).toHaveBeenCalledWith(
      expect.objectContaining({ maskRule: 'keep_middle' })
    )
  })
})
