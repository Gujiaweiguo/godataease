import { describe, it, expect, vi, beforeEach } from 'vitest'

const { mockEmbeddedStore } = vi.hoisted(() => {
  vi.stubEnv('VITE_API_BASEPATH', '/api')
  return { mockEmbeddedStore: { baseUrl: '' } }
})

vi.mock('@/store/modules/embedded', () => ({
  useEmbedded: () => mockEmbeddedStore
}))

vi.mock('@/store/modules/data-visualization/dvMain', () => ({
  dvMainStoreWithOut: () => ({
    canvasStyleData: { value: {} },
    componentData: { value: [] },
    canvasViewInfo: { value: {} },
    canvasViewDataInfo: { value: {} },
    dvInfo: { value: {} }
  })
}))

vi.mock('pinia', () => ({
  storeToRefs: (store: any) => ({
    canvasStyleData: store.canvasStyleData,
    componentData: store.componentData,
    canvasViewInfo: store.canvasViewInfo,
    canvasViewDataInfo: store.canvasViewDataInfo,
    dvInfo: store.dvInfo
  })
}))

vi.mock('@/api/staticResource', () => ({
  findResourceAsBase64: vi.fn()
}))

vi.mock('html2canvas', () => ({
  default: vi.fn()
}))

vi.mock('jspdf', () => ({
  default: vi.fn()
}))

vi.mock('file-saver', () => ({
  default: { saveAs: vi.fn() }
}))

vi.mock('modern-screenshot', () => ({
  domToPng: vi.fn()
}))

vi.mock('@/utils/utils', () => ({
  deepCopy: (obj: any) => JSON.parse(JSON.stringify(obj))
}))

import { formatterUrl, imgUrlTrans, dataURLToBlob } from '@/utils/imgUtils'

describe('imgUtils', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockEmbeddedStore.baseUrl = ''
  })

  describe('formatterUrl', () => {
    it('should replace //de2api with /de2api', () => {
      expect(formatterUrl('https://example.com//de2api/resource')).toBe(
        'https://example.com/de2api/resource'
      )
    })

    it('should replace only first occurrence of //de2api', () => {
      // String.replace only replaces first match
      expect(formatterUrl('//de2api//de2api/test')).toBe('/de2api//de2api/test')
    })

    it('should return unchanged URL when //de2api is not present', () => {
      expect(formatterUrl('https://example.com/api/resource')).toBe(
        'https://example.com/api/resource'
      )
    })

    it('should handle URL with only /de2api (single slash)', () => {
      expect(formatterUrl('https://example.com/de2api/test')).toBe(
        'https://example.com/de2api/test'
      )
    })

    it('should handle empty string', () => {
      expect(formatterUrl('')).toBe('')
    })
  })

  describe('imgUrlTrans', () => {
    it('should return undefined for falsy url', () => {
      expect(imgUrlTrans(null)).toBeUndefined()
      expect(imgUrlTrans(undefined)).toBeUndefined()
      expect(imgUrlTrans('')).toBeUndefined()
    })

    it('should prepend basePath for static-resource URLs', () => {
      const url = '/static-resource/image.png'
      const result = imgUrlTrans(url)
      expect(result).toBe('/api/static-resource/image.png')
    })

    it('should use embeddedStore.baseUrl for static-resource URLs when set', () => {
      mockEmbeddedStore.baseUrl = 'https://embedded.com'
      const url = '/static-resource/img.png'
      const result = imgUrlTrans(url)
      expect(result).toContain('embedded.com')
    })

    it('should strip /api prefix when embeddedStore.baseUrl is set and url starts with /api', () => {
      mockEmbeddedStore.baseUrl = 'https://embedded.com'
      const url = '/static-resource/img.png'
      const result = imgUrlTrans(url)
      // rawUrl = '/api/static-resource/img.png', starts with '/api' -> slice(5) -> '/static-resource/img.png'
      // but baseUrl 'https://embedded.com' has no trailing slash, so concatenated as-is
      expect(result).toBe('https://embedded.comstatic-resource/img.png')
    })

    it('should apply formatterUrl to non-static-resource URLs', () => {
      const url = 'https://example.com/image.png'
      const result = imgUrlTrans(url)
      expect(result).toBe('https://example.com/image.png')
    })

    it('should replace com// with com/ for non-static URLs', () => {
      const url = 'https://example.com//path/img.png'
      const result = imgUrlTrans(url)
      expect(result).toBe('https://example.com/path/img.png')
    })
  })

  describe('dataURLToBlob', () => {
    it('should convert a PNG data URL to Blob with correct type', () => {
      const base64Data = btoa('test-image-data')
      const dataUrl = `data:image/png;base64,${base64Data}`
      const blob = dataURLToBlob(dataUrl) as Blob

      expect(blob).toBeInstanceOf(Blob)
      expect(blob.type).toBe('image/png')
    })

    it('should convert a JPEG data URL to Blob', () => {
      const base64Data = btoa('jpeg-data')
      const dataUrl = `data:image/jpeg;base64,${base64Data}`
      const blob = dataURLToBlob(dataUrl) as Blob

      expect(blob).toBeInstanceOf(Blob)
      expect(blob.type).toBe('image/jpeg')
    })

    it('should preserve binary data correctly', async () => {
      const original = 'Hello, DataEase!'
      const base64Data = btoa(original)
      const dataUrl = `data:text/plain;base64,${base64Data}`
      const blob = dataURLToBlob(dataUrl) as Blob

      const text = await blob.text()
      expect(text).toBe(original)
    })

    it('should handle SVG data URL', () => {
      const svgContent = '<svg><rect/></svg>'
      const base64Data = btoa(svgContent)
      const dataUrl = `data:image/svg+xml;base64,${base64Data}`
      const blob = dataURLToBlob(dataUrl) as Blob

      expect(blob.type).toBe('image/svg+xml')
    })

    it('should handle empty base64 data', () => {
      const dataUrl = 'data:image/png;base64,'
      const blob = dataURLToBlob(dataUrl) as Blob

      expect(blob).toBeInstanceOf(Blob)
      expect(blob.size).toBe(0)
    })
  })
})
