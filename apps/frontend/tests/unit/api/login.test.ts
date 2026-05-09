import { describe, it, expect, vi, beforeEach } from 'vitest'

const { requestMock } = vi.hoisted(() => ({
  requestMock: {
    post: vi.fn().mockResolvedValue({ data: {} }),
    get: vi.fn().mockResolvedValue({ data: {} })
  }
}))

vi.mock('@/config/axios', () => ({ default: requestMock }))

import {
  loginApi,
  queryDekey,
  querySymmetricKey,
  modelApi,
  platformLoginApi,
  logoutApi,
  refreshApi,
  uiLoadApi,
  loginCategoryApi
} from '@/api/login'

describe('Login API wrappers', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('loginApi posts credentials to localLogin endpoint', () => {
    const credentials = { name: 'admin', pwd: 'admin123' }
    loginApi(credentials)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/login/localLogin',
      data: credentials
    })
  })

  it('loginApi passes arbitrary data object', () => {
    const data = { name: 'user', pwd: 'pass', token: 'abc' }
    loginApi(data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/login/localLogin',
      data
    })
  })

  it('queryDekey fetches dekey', () => {
    queryDekey()
    expect(requestMock.get).toHaveBeenCalledWith({ url: '/dekey' })
  })

  it('querySymmetricKey fetches symmetricKey', () => {
    querySymmetricKey()
    expect(requestMock.get).toHaveBeenCalledWith({ url: '/symmetricKey' })
  })

  it('modelApi fetches model', () => {
    modelApi()
    expect(requestMock.get).toHaveBeenCalledWith({ url: '/model' })
  })

  it('platformLoginApi posts to platform login with origin suffix', () => {
    platformLoginApi('cas')
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/login/platformLogin/cas'
    })
  })

  it('platformLoginApi appends different origin values', () => {
    platformLoginApi('ldap')
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/login/platformLogin/ldap'
    })
  })

  it('logoutApi fetches logout endpoint', () => {
    logoutApi()
    expect(requestMock.get).toHaveBeenCalledWith({ url: '/api/logout' })
  })

  it('refreshApi fetches refresh with time param', () => {
    refreshApi(1234567890)
    expect(requestMock.get).toHaveBeenCalledWith({
      url: '/login/refresh',
      params: { time: 1234567890 }
    })
  })

  it('refreshApi works without time argument', () => {
    refreshApi()
    expect(requestMock.get).toHaveBeenCalledWith({
      url: '/login/refresh',
      params: { time: undefined }
    })
  })

  it('uiLoadApi fetches sysParameter/ui', () => {
    uiLoadApi()
    expect(requestMock.get).toHaveBeenCalledWith({ url: '/sysParameter/ui' })
  })

  it('loginCategoryApi fetches defaultLogin', () => {
    loginCategoryApi()
    expect(requestMock.get).toHaveBeenCalledWith({ url: '/sysParameter/defaultLogin' })
  })

  it('loginApi returns the request promise', async () => {
    requestMock.post.mockResolvedValueOnce({ data: { token: 'abc' } })
    const result = await loginApi({ name: 'admin', pwd: 'pass' })
    expect(result).toEqual({ data: { token: 'abc' } })
  })

  it('queryDekey returns the request promise', async () => {
    requestMock.get.mockResolvedValueOnce({ data: { key: 'value' } })
    const result = await queryDekey()
    expect(result).toEqual({ data: { key: 'value' } })
  })
})
