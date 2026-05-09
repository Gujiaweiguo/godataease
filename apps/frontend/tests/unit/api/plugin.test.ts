import { describe, it, expect, vi, beforeEach } from 'vitest'

const { requestMock } = vi.hoisted(() => ({
  requestMock: {
    post: vi.fn().mockResolvedValue({ data: {} }),
    get: vi.fn().mockResolvedValue({ data: {} })
  }
}))

vi.mock('@/config/axios', () => ({ default: requestMock }))

import { load, loadPluginApi, loadDistributed, xpackModelApi } from '@/api/plugin'

describe('api/plugin', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('load gets xpackComponent content by key', () => {
    load('component-a')
    expect(requestMock.get).toHaveBeenCalledWith({ url: '/xpackComponent/content/component-a' })
  })

  it('loadPluginApi gets xpackComponent contentPlugin by key', () => {
    loadPluginApi('plugin-b')
    expect(requestMock.get).toHaveBeenCalledWith({ url: '/xpackComponent/contentPlugin/plugin-b' })
  })

  it('loadDistributed gets DEXPack umd bundle', () => {
    loadDistributed()
    expect(requestMock.get).toHaveBeenCalledWith({ url: '/DEXPack.umd.js' })
  })

  it('xpackModelApi gets xpackModel', () => {
    xpackModelApi()
    expect(requestMock.get).toHaveBeenCalledWith({ url: '/xpackModel' })
  })
})
