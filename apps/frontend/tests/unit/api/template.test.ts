import { describe, it, expect, vi, beforeEach } from 'vitest'

const { requestMock } = vi.hoisted(() => ({
  requestMock: {
    post: vi.fn().mockResolvedValue({ data: {} }),
    get: vi.fn().mockResolvedValue({ data: {} })
  }
}))

vi.mock('@/config/axios', () => ({ default: requestMock }))

import {
  save,
  templateDelete,
  deleteCategory,
  showTemplateList,
  findOne,
  find,
  findCategories,
  nameCheck,
  categoryTemplateNameCheck,
  checkCategoryTemplateBatchNames,
  batchDelete,
  batchUpdate,
  findCategoriesByTemplateIds
} from '@/api/template'

describe('api/template', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('save posts to templateManage/save with loading', () => {
    const data = { name: 'Template1', categoryId: 1 }
    save(data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/templateManage/save',
      data,
      loading: true
    })
  })

  it('templateDelete posts to delete with id and categoryId', () => {
    templateDelete(10, 5)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/templateManage/delete/10/5'
    })
  })

  it('deleteCategory posts to deleteCategory with id', () => {
    deleteCategory(3)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/templateManage/deleteCategory/3'
    })
  })

  it('showTemplateList posts to templateList', () => {
    const data = { categoryId: 1 }
    showTemplateList(data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/templateManage/templateList',
      data
    })
  })

  it('findOne gets templateManage findOne with id', () => {
    findOne(7)
    expect(requestMock.get).toHaveBeenCalledWith({
      url: '/templateManage/findOne/7'
    })
  })

  it('find posts to templateManage find with loading', () => {
    const data = { keyword: 'test' }
    find(data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/templateManage/find',
      data,
      loading: true
    })
  })

  it('findCategories posts to templateManage findCategories with loading', () => {
    const data = { level: 1 }
    findCategories(data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/templateManage/findCategories',
      data,
      loading: true
    })
  })

  it('nameCheck posts to templateManage nameCheck', () => {
    const data = { name: 'MyTemplate' }
    nameCheck(data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/templateManage/nameCheck',
      data
    })
  })

  it('categoryTemplateNameCheck posts to categoryTemplateNameCheck', () => {
    const data = { name: 'Check', categoryId: 1 }
    categoryTemplateNameCheck(data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/templateManage/categoryTemplateNameCheck',
      data
    })
  })

  it('checkCategoryTemplateBatchNames posts to categoryTemplateNameCheck (batch endpoint)', () => {
    const data = { names: ['A', 'B'] }
    checkCategoryTemplateBatchNames(data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/templateManage/categoryTemplateNameCheck',
      data
    })
  })

  it('batchDelete posts to templateManage batchDelete', () => {
    const data = { ids: [1, 2, 3] }
    batchDelete(data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/templateManage/batchDelete',
      data
    })
  })

  it('batchUpdate posts to templateManage batchUpdate', () => {
    const data = { templates: [{ id: 1, categoryId: 2 }] }
    batchUpdate(data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/templateManage/batchUpdate',
      data
    })
  })

  it('findCategoriesByTemplateIds posts to findCategoriesByTemplateIds', () => {
    const data = { templateIds: [1, 2] }
    findCategoriesByTemplateIds(data)
    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/templateManage/findCategoriesByTemplateIds',
      data
    })
  })
})
