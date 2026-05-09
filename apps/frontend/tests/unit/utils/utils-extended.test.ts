import { describe, it, expect, vi, beforeEach } from 'vitest'

vi.mock('element-plus-secondary', async () => {
  const { elementPlusSecondaryModuleMock } = await import('../helpers')
  return elementPlusSecondaryModuleMock
})

const { wsCacheMock } = vi.hoisted(() => ({
  wsCacheMock: {
    get: vi.fn(),
    set: vi.fn(),
    delete: vi.fn()
  }
}))

vi.mock('@/hooks/web/useCache', () => ({
  useCache: () => ({ wsCache: wsCacheMock })
}))

vi.mock('@/utils/RemoteJs', () => ({
  loadScript: vi.fn().mockResolvedValue(Promise.resolve())
}))

import {
  setColorName,
  isPreventDrop,
  formatExt,
  exportPermission,
  getBrowserLocale,
  nameTrim,
  getActiveCategories,
  cutTargetTree,
  isTablet,
  isFreeFolder,
  filterFreeFolder,
  cleanPlatformFlag
} from '@/utils/utils'

describe('utils-extended', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('setColorName', () => {
    it('should set colorName with highlighted keyword span', () => {
      const obj: any = { name: 'Hello World' }
      setColorName(obj, 'World')
      expect(obj.colorName).toBe('Hello <span class="search-key-span">World</span>')
    })

    it('should set colorName to null when keyword not found', () => {
      const obj: any = { name: 'Hello World' }
      setColorName(obj, 'xyz')
      expect(obj.colorName).toBeNull()
    })

    it('should set colorName to null when keyword is empty', () => {
      const obj: any = { name: 'Hello World' }
      setColorName(obj, '')
      expect(obj.colorName).toBeNull()
    })

    it('should use custom key and colorKey', () => {
      const obj: any = { title: 'My Test Title' }
      setColorName(obj, 'Test', 'title', 'highlight')
      expect(obj.highlight).toBe('My <span class="search-key-span">Test</span> Title')
    })

    it('should use default key "name" and colorKey "colorName"', () => {
      const obj: any = { name: 'abc' }
      setColorName(obj, 'a')
      expect(obj.colorName).toBe('<span class="search-key-span">a</span>bc')
    })
  })

  describe('isPreventDrop', () => {
    it('should return false for VText component', () => {
      expect(isPreventDrop('VText')).toBe(false)
    })

    it('should return false for RectShape component', () => {
      expect(isPreventDrop('RectShape')).toBe(false)
    })

    it('should return false for CircleShape component', () => {
      expect(isPreventDrop('CircleShape')).toBe(false)
    })

    it('should return false for SVG-prefixed components', () => {
      expect(isPreventDrop('SVGStar')).toBe(false)
      expect(isPreventDrop('SVGPolygon')).toBe(false)
    })

    it('should return true for other components', () => {
      expect(isPreventDrop('UserView')).toBe(true)
      expect(isPreventDrop('Group')).toBe(true)
    })
  })

  describe('formatExt', () => {
    it('should return null for falsy input', () => {
      expect(formatExt(0)).toBeNull()
      expect(formatExt(null as any)).toBeNull()
      expect(formatExt(undefined as any)).toBeNull()
    })

    it('should reverse digits into array', () => {
      expect(formatExt(123)).toEqual([3, 2, 1])
    })

    it('should handle single digit', () => {
      expect(formatExt(5)).toEqual([5])
    })

    it('should handle number with zeros', () => {
      expect(formatExt(100)).toEqual([0, 0, 1])
    })
  })

  describe('exportPermission', () => {
    it('should return [0,0,0] for weight 1', () => {
      expect(exportPermission(1, null)).toEqual([0, 0, 0])
    })

    it('should return [0,0,0] for weight 0', () => {
      expect(exportPermission(0, null)).toEqual([0, 0, 0])
    })

    it('should return [0,0,0] for falsy weight', () => {
      expect(exportPermission(null as any, null)).toEqual([0, 0, 0])
    })

    it('should return [1,1,1] for weight 9', () => {
      expect(exportPermission(9, null)).toEqual([1, 1, 1])
    })

    it('should return [0,0,0] when weight is valid but ext is missing', () => {
      expect(exportPermission(5, null)).toEqual([0, 0, 0])
    })

    it('should return ext values for non-trivial weight', () => {
      expect(exportPermission(5, 100)).toEqual([0, 0, 1])
    })
  })

  describe('getBrowserLocale', () => {
    it('should return zh-CN for Chinese language', () => {
      Object.defineProperty(navigator, 'language', { value: 'zh', configurable: true })
      expect(getBrowserLocale()).toBe('zh-CN')
    })

    it('should return zh-CN for zh-CN language', () => {
      Object.defineProperty(navigator, 'language', { value: 'zh-CN', configurable: true })
      expect(getBrowserLocale()).toBe('zh-CN')
    })

    it('should return tw for zh-TW language', () => {
      Object.defineProperty(navigator, 'language', { value: 'zh-TW', configurable: true })
      expect(getBrowserLocale()).toBe('tw')
    })

    it('should return en for English language', () => {
      Object.defineProperty(navigator, 'language', { value: 'en-US', configurable: true })
      expect(getBrowserLocale()).toBe('en')
    })

    it('should return zh-CN when no language is set', () => {
      Object.defineProperty(navigator, 'language', { value: '', configurable: true })
      expect(getBrowserLocale()).toBe('zh-CN')
    })

    it('should return raw language for other locales', () => {
      Object.defineProperty(navigator, 'language', { value: 'ja-JP', configurable: true })
      expect(getBrowserLocale()).toBe('ja-JP')
    })
  })

  describe('nameTrim', () => {
    it('should trim whitespace from name', () => {
      const target = { name: '  hello  ' }
      nameTrim(target)
      expect(target.name).toBe('hello')
    })

    it('should throw when name is empty after trim', () => {
      const target = { name: '   ' }
      expect(() => nameTrim(target)).toThrow('名称字段长度1-64个字符')
    })

    it('should throw when name exceeds 64 characters', () => {
      const target = { name: 'a'.repeat(65) }
      expect(() => nameTrim(target)).toThrow('名称字段长度1-64个字符')
    })

    it('should accept valid name with 64 characters', () => {
      const target = { name: 'a'.repeat(64) }
      expect(() => nameTrim(target)).not.toThrow()
      expect(target.name).toBe('a'.repeat(64))
    })

    it('should do nothing when name is undefined', () => {
      const target = { name: undefined }
      expect(() => nameTrim(target)).not.toThrow()
    })
  })

  describe('getActiveCategories', () => {
    it('should return set with "最近使用" by default', () => {
      const result = getActiveCategories(null)
      expect(result).toBeInstanceOf(Set)
      expect(result.has('最近使用')).toBe(true)
    })

    it('should extract category names from categories array', () => {
      const contents = [
        {
          categories: [{ name: '图表' }, { name: '指标' }],
          showFlag: true
        }
      ]
      const result = getActiveCategories(contents)
      expect(result.has('图表')).toBe(true)
      expect(result.has('指标')).toBe(true)
    })

    it('should extract names from categoryNames array', () => {
      const contents = [
        {
          categoryNames: ['报表', '仪表板'],
          showFlag: true
        }
      ]
      const result = getActiveCategories(contents)
      expect(result.has('报表')).toBe(true)
      expect(result.has('仪表板')).toBe(true)
    })

    it('should extract label from category object', () => {
      const contents = [
        {
          category: { label: '其他' },
          showFlag: true
        }
      ]
      const result = getActiveCategories(contents)
      expect(result.has('其他')).toBe(true)
    })

    it('should skip items with showFlag false', () => {
      const contents = [
        {
          categories: [{ name: '隐藏分类' }],
          showFlag: false
        }
      ]
      const result = getActiveCategories(contents)
      expect(result.has('隐藏分类')).toBe(false)
    })

    it('should deduplicate categories into a Set', () => {
      const contents = [
        {
          categoryNames: ['图表'],
          showFlag: true
        },
        {
          categories: [{ name: '图表' }],
          showFlag: true
        }
      ]
      const result = getActiveCategories(contents)
      const arr = [...result]
      expect(arr.filter(c => c === '图表').length).toBe(1)
    })
  })

  describe('cutTargetTree', () => {
    it('should remove target node from tree', () => {
      const tree: any[] = [
        { id: 1, children: [] },
        { id: 2, children: [] }
      ]
      cutTargetTree(tree, 1)
      expect(tree).toEqual([{ id: 2, children: [] }])
    })

    it('should search recursively in children', () => {
      const tree: any[] = [
        {
          id: 1,
          children: [
            { id: 3, children: [] },
            { id: 4, children: [] }
          ]
        }
      ]
      cutTargetTree(tree, 4)
      expect(tree[0].children).toEqual([{ id: 3, children: [] }])
    })

    it('should do nothing when target not found', () => {
      const tree: any[] = [{ id: 1, children: [] }]
      cutTargetTree(tree, 999)
      expect(tree).toEqual([{ id: 1, children: [] }])
    })
  })

  describe('isTablet', () => {
    it('should return true for iPad user agent', () => {
      const original = navigator.userAgent
      Object.defineProperty(navigator, 'userAgent', {
        value: 'Mozilla/5.0 (iPad; CPU OS 14_0 like Mac OS X)',
        configurable: true
      })
      expect(isTablet()).toBe(true)
      Object.defineProperty(navigator, 'userAgent', {
        value: original,
        configurable: true
      })
    })

    it('should return false for desktop user agent', () => {
      const original = navigator.userAgent
      Object.defineProperty(navigator, 'userAgent', {
        value: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64)',
        configurable: true
      })
      expect(isTablet()).toBe(false)
      Object.defineProperty(navigator, 'userAgent', {
        value: original,
        configurable: true
      })
    })
  })

  describe('isFreeFolder', () => {
    it('should return false when oid is not set', () => {
      wsCacheMock.get.mockReturnValue(null)
      expect(isFreeFolder({}, 'dashboard')).toBe(false)
    })

    it('should traverse parent chain to find freeRootId', () => {
      wsCacheMock.get.mockReturnValue('10')
      const leaf = { data: { id: '12' }, parent: null }
      expect(isFreeFolder(leaf, 2)).toBe(true)
    })

    it('should return false when no matching node found', () => {
      wsCacheMock.get.mockReturnValue('10')
      const node = { data: { id: '99' }, parent: null }
      expect(isFreeFolder(node, 2)).toBe(false)
    })
  })

  describe('filterFreeFolder', () => {
    it('should remove matching top-level node from list', () => {
      wsCacheMock.get.mockReturnValue('11')
      const list: any[] = [{ id: '12' }, { id: '99' }]
      filterFreeFolder(list, 'dashboard')
      expect(list).toEqual([{ id: '99' }])
    })

    it('should remove matching child when id is "0"', () => {
      wsCacheMock.get.mockReturnValue('11')
      const list: any[] = [
        {
          id: '0',
          children: [{ id: '13' }, { id: '12' }]
        }
      ]
      filterFreeFolder(list, 'dataV')
      expect(list[0].children).toEqual([{ id: '12' }])
    })

    it('should do nothing when oid is not set', () => {
      wsCacheMock.get.mockReturnValue(null)
      const list: any[] = [{ id: '12' }]
      filterFreeFolder(list, 'dashboard')
      expect(list).toEqual([{ id: '12' }])
    })

    it('should do nothing for invalid flagText', () => {
      wsCacheMock.get.mockReturnValue('10')
      const list: any[] = [{ id: '12' }]
      filterFreeFolder(list, 'invalid')
      expect(list).toEqual([{ id: '12' }])
    })
  })

  describe('cleanPlatformFlag', () => {
    it('should delete platform key and return false', () => {
      const result = cleanPlatformFlag()
      expect(wsCacheMock.delete).toHaveBeenCalledWith('out_auth_platform')
      expect(result).toBe(false)
    })
  })
})
