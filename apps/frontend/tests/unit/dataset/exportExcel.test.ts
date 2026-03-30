import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, nextTick } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'

const hoisted = vi.hoisted(() => ({
  exportTasks: vi.fn(),
  exportTasksRecords: vi.fn(),
  exportRetry: vi.fn(),
  exportDelete: vi.fn(),
  exportDeleteAll: vi.fn(),
  exportDeletePost: vi.fn(),
  generateDownloadUri: vi.fn(),
  emitter: { emit: vi.fn() },
  wsCache: { get: vi.fn((key: string) => (key === 'open-backend' ? '0' : '0')) }
}))

vi.mock('@/api/dataset', () => ({
  exportTasks: hoisted.exportTasks,
  exportTasksRecords: hoisted.exportTasksRecords,
  exportRetry: hoisted.exportRetry,
  exportDelete: hoisted.exportDelete,
  exportDeleteAll: hoisted.exportDeleteAll,
  exportDeletePost: hoisted.exportDeletePost,
  generateDownloadUri: hoisted.generateDownloadUri
}))

vi.mock('@/hooks/web/useI18n', () => ({
  useI18n: () => ({ t: (value: string) => value })
}))

vi.mock('@/hooks/web/useEmitt', () => ({
  useEmitt: () => ({ emitter: hoisted.emitter })
}))

vi.mock('@/hooks/web/useCache', () => ({
  useCache: () => ({ wsCache: hoisted.wsCache })
}))

vi.mock('@/store/modules/link', () => ({
  useLinkStoreWithOut: () => ({ getLinkToken: '' })
}))

vi.mock('@/store/modules/app', () => ({
  useAppStoreWithOut: () => ({
    getIsDataEaseBi: false,
    getIsIframe: false
  })
}))

vi.mock('@/config/axios/service', () => ({
  PATH_URL: ''
}))

vi.mock('element-plus-secondary', () => {
  const ElDrawer = defineComponent({
    name: 'ElDrawer',
    props: { modelValue: { type: Boolean, default: false } },
    template: '<div data-test="drawer" :data-open="String(modelValue)"><slot /></div>'
  })
  const ElTabs = defineComponent({
    name: 'ElTabs',
    props: { modelValue: { type: String, default: '' } },
    template: '<div data-test="tabs"><slot /></div>'
  })
  const ElTabPane = defineComponent({
    name: 'ElTabPane',
    template: '<div data-test="tab-pane"></div>'
  })
  const ElButton = defineComponent({
    name: 'ElButton',
    template: '<button><slot /></button>'
  })

  return {
    ElDrawer,
    ElTabs,
    ElTabPane,
    ElButton,
    ElMessage: vi.fn(),
    ElMessageBox: { confirm: vi.fn(() => Promise.resolve()) }
  }
})

import ExportExcel from '../../../src/views/visualized/data/dataset/ExportExcel.vue'

type ExportExcelExposed = {
  init: (params?: { activeName?: string }) => void | Promise<void>
}

describe('ExportExcel page init', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    vi.clearAllMocks()
    hoisted.exportTasksRecords.mockResolvedValue({
      code: '000000',
      data: {
        ALL: 2,
        IN_PROGRESS: 1,
        SUCCESS: 1,
        FAILED: 0,
        PENDING: 0
      }
    })
    hoisted.exportTasks.mockResolvedValue({
      code: '000000',
      data: {
        total: 1,
        records: [{ id: 'task-1', exportStatus: 'FAILED' }]
      }
    })
  })

  afterEach(() => {
    vi.runOnlyPendingTimers()
    vi.useRealTimers()
  })

  it('opens drawer and queries the requested initial tab', async () => {
    const wrapper = mount(ExportExcel, {
      global: {
        directives: {
          loading: () => undefined
        },
        stubs: {
          EmptyBackground: true,
          Icon: true,
          GridTable: true,
          dvPreviewDownload: true,
          deDelete: true,
          icon_fileExcel_colorful: true,
	          icon_refresh_outlined: true,
	          'el-drawer': true,
	          'el-dialog': true,
	          'el-table-column': true,
	          'el-icon': true,
	          'el-progress': true,
	          'el-tooltip': true
        },
        mocks: {
          $t: (value: string) => value
        }
      }
    })

    const exposed = wrapper.vm.$.exposed as ExportExcelExposed

    await exposed.init({ activeName: 'FAILED' })
    await flushPromises()
    await nextTick()

    expect(wrapper.get('el-drawer-stub').attributes('modelvalue')).toBe('true')
    expect(hoisted.exportTasksRecords).toHaveBeenCalledTimes(1)
    expect(hoisted.exportTasks).toHaveBeenCalledWith(1, 10, 'FAILED')
  })

  it('polls records and tasks again for the IN_PROGRESS tab', async () => {
    const wrapper = mount(ExportExcel, {
      global: {
        directives: {
          loading: () => undefined
        },
        stubs: {
          EmptyBackground: true,
          Icon: true,
          GridTable: true,
          dvPreviewDownload: true,
          deDelete: true,
          icon_fileExcel_colorful: true,
	          icon_refresh_outlined: true,
	          'el-drawer': true,
	          'el-dialog': true,
	          'el-table-column': true,
	          'el-icon': true,
	          'el-progress': true,
	          'el-tooltip': true
        },
        mocks: {
          $t: (value: string) => value
        }
      }
    })

    const exposed = wrapper.vm.$.exposed as ExportExcelExposed

    await exposed.init({ activeName: 'IN_PROGRESS' })
    await flushPromises()

    expect(hoisted.exportTasksRecords).toHaveBeenCalledTimes(1)
    expect(hoisted.exportTasks).toHaveBeenCalledWith(1, 10, 'IN_PROGRESS')

    vi.advanceTimersByTime(5000)
    await flushPromises()

    expect(hoisted.exportTasksRecords).toHaveBeenCalledTimes(2)
    expect(hoisted.exportTasks).toHaveBeenLastCalledWith(1, 10, 'IN_PROGRESS')
  })
})
