import { describe, it, expect, vi, beforeEach } from 'vitest'

const { requestMock } = vi.hoisted(() => ({
  requestMock: {
    post: vi.fn().mockResolvedValue({ data: {} }),
    get: vi.fn().mockResolvedValue({ data: {} }),
    delete: vi.fn().mockResolvedValue({ data: {} }),
    put: vi.fn().mockResolvedValue({ data: {} }),
    download: vi.fn().mockResolvedValue({ data: {} })
  }
}))

vi.mock('@/config/axios', () => ({ default: requestMock }))

import {
  list,
  create,
  edit,
  deleteById,
  defaultFont,
  uploadFontFile
} from '@/api/font'

describe('Font API wrappers', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('list gets typeface/listFont and returns res.data', async () => {
    requestMock.get.mockResolvedValueOnce({ data: [{ id: '1', name: 'Arial' }] })
    const result = await list()
    expect(requestMock.get).toHaveBeenCalledWith({ url: '/typeface/listFont' })
    expect(result).toEqual([{ id: '1', name: 'Arial' }])
  })

  it('list returns undefined when res is null', async () => {
    requestMock.get.mockResolvedValueOnce(null)
    const result = await list()
    expect(result).toBeUndefined()
  })

  it('create posts to typeface/create and returns res.data', async () => {
    requestMock.post.mockResolvedValueOnce({ data: { id: '2' } })
    const result = await create({ name: 'NewFont' })
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/typeface/create',
      data: { name: 'NewFont' }
    })
    expect(result).toEqual({ id: '2' })
  })

  it('create uses default empty object when no data provided', () => {
    create()
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/typeface/create',
      data: {}
    })
  })

  it('edit posts to typeface/edit and returns res.data', async () => {
    requestMock.post.mockResolvedValueOnce({ data: { success: true } })
    const result = await edit({ id: '1', name: 'Updated' })
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/typeface/edit',
      data: { id: '1', name: 'Updated' }
    })
    expect(result).toEqual({ success: true })
  })

  it('edit uses default empty object when no data provided', () => {
    edit()
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/typeface/edit',
      data: {}
    })
  })

  it('deleteById posts to typeface/delete/:id and returns res.data', async () => {
    requestMock.post.mockResolvedValueOnce({ data: { success: true } })
    const result = await deleteById('font1')
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/typeface/delete/font1',
      data: {}
    })
    expect(result).toEqual({ success: true })
  })

  it('defaultFont gets typeface/defaultFont and returns res.data', async () => {
    requestMock.get.mockResolvedValueOnce({ data: { name: 'Default' } })
    const result = await defaultFont()
    expect(requestMock.get).toHaveBeenCalledWith({ url: '/typeface/defaultFont' })
    expect(result).toEqual({ name: 'Default' })
  })

  it('uploadFontFile posts to typeface/uploadFile with multipart headers', async () => {
    const formData = new FormData()
    requestMock.post.mockResolvedValueOnce({ data: { url: '/files/font.ttf' } })
    const result = await uploadFontFile(formData)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/typeface/uploadFile',
      data: formData,
      loading: true,
      headersType: 'multipart/form-data;'
    })
    expect(result).toEqual({ data: { url: '/files/font.ttf' } })
  })

  it('uploadFontFile returns raw res without .data extraction', async () => {
    const rawResponse = { data: { id: 'file1' } }
    requestMock.post.mockResolvedValueOnce(rawResponse)
    const result = await uploadFontFile({})
    expect(result).toEqual(rawResponse)
  })
})
