import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

const mocks = vi.hoisted(() => ({
  cacheState: {
    'xpack-model-distributed': null
  } as Record<string, any>,
  xpackModelApiMock: vi.fn(),
  loadDistributedMock: vi.fn(),
  emitterEmitMock: vi.fn()
}))

vi.mock('../../../../src/api/plugin', () => ({
  xpackModelApi: mocks.xpackModelApiMock,
  loadDistributed: mocks.loadDistributedMock
}))

vi.mock('../../../../src/hooks/web/useCache', () => ({
  useCache: () => ({
    wsCache: {
      get: (key: string) => mocks.cacheState[key],
      set: (key: string, value: any) => {
        mocks.cacheState[key] = value
      }
    }
  })
}))

vi.mock('../../../../src/components/config-global/src/ConfigGlobal.vue', () => ({
  default: {
    template: '<div><slot /></div>'
  }
}))

vi.mock('../../../../src/plugins/vue-i18n', () => ({
  i18n: {}
}))

vi.mock('../../../../src/router', () => ({
  default: {}
}))

vi.mock('../../../../src/hooks/web/useEmitt', () => ({
  useEmitt: vi.fn(() => ({
    emitter: {
      all: {},
      emit: mocks.emitterEmitMock
    }
  }))
}))

vi.mock('../../../../src/utils/utils', async () => {
  const actual = await vi.importActual('../../../../src/utils/utils')
  return {
    ...actual,
    isNull: (value: unknown) => value === null || value === undefined
  }
})

import XpackComponent from '../../../../src/components/plugin/src/index.vue'

describe('XpackComponent', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mocks.cacheState['xpack-model-distributed'] = null
  })

  it('falls back gracefully when xpackModelApi rejects', async () => {
    mocks.xpackModelApiMock.mockRejectedValueOnce(new Error('unauthorized'))

    const wrapper = mount(XpackComponent, {
      attrs: {
        jsname: 'L2NvbXBvbmVudC9sb2dpbi9IYW5kbGVy'
      },
      global: {
        directives: {
          loading: () => undefined
        }
      }
    })

    await flushPromises()

    expect(mocks.xpackModelApiMock).toHaveBeenCalledTimes(1)
    expect(wrapper.emitted('loadFail')).toBeTruthy()
    expect(wrapper.findAll('div').length).toBeGreaterThan(1)
    expect(wrapper.text()).toBe('')
  })
})
