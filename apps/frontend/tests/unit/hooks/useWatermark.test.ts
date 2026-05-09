import { describe, it, expect, vi, beforeEach } from 'vitest'

const mockGetContext = vi.fn(() => ({
  rotate: vi.fn(),
  font: '',
  fillStyle: '',
  textAlign: '',
  textBaseline: '',
  fillText: vi.fn()
}))
const mockToDataURL = vi.fn(() => 'data:image/png;base64,watermark')

const mockAppendChild = vi.fn()
const mockRemoveChild = vi.fn()
const mockBody = { appendChild: mockAppendChild, removeChild: mockRemoveChild }

const mockGetElementById = vi.fn()
const origCreateElement = document.createElement.bind(document)

describe('useWatermark', () => {
  let createdElements: any[] = []

  beforeEach(() => {
    vi.clearAllMocks()
    createdElements = []

    document.createElement = vi.fn((tag: string) => {
      const el = origCreateElement(tag)
      if (tag === 'canvas') {
        el.getContext = mockGetContext
        el.toDataURL = mockToDataURL
      }
      createdElements.push({ tag, el })
      return el
    }) as any

    document.getElementById = mockGetElementById
  })

  it('should return setWatermark and clear functions', async () => {
    const { useWatermark } = await import('@/hooks/web/useWatermark')
    const result = useWatermark(mockBody as any)
    expect(typeof result.setWatermark).toBe('function')
    expect(typeof result.clear).toBe('function')
  })

  it('should create canvas and div elements when setWatermark is called', async () => {
    const { useWatermark } = await import('@/hooks/web/useWatermark')
    const { setWatermark } = useWatermark(mockBody as any)
    setWatermark('Test')
    expect(document.createElement).toHaveBeenCalledWith('canvas')
    expect(document.createElement).toHaveBeenCalledWith('div')
  })

  it('should call canvas getContext with 2d', async () => {
    const { useWatermark } = await import('@/hooks/web/useWatermark')
    const { setWatermark } = useWatermark(mockBody as any)
    setWatermark('Hello')
    expect(mockGetContext).toHaveBeenCalledWith('2d')
  })

  it('should call canvas toDataURL to generate background image', async () => {
    const { useWatermark } = await import('@/hooks/web/useWatermark')
    const { setWatermark } = useWatermark(mockBody as any)
    setWatermark('Test')
    expect(mockToDataURL).toHaveBeenCalledWith('image/png')
  })

  it('should append div to appendEl when setWatermark is called', async () => {
    const { useWatermark } = await import('@/hooks/web/useWatermark')
    const { setWatermark } = useWatermark(mockBody as any)
    setWatermark('Test')
    expect(mockAppendChild).toHaveBeenCalled()
  })

  it('should add resize event listener when setWatermark is called', async () => {
    const addSpy = vi.spyOn(window, 'addEventListener')
    const { useWatermark } = await import('@/hooks/web/useWatermark')
    const { setWatermark } = useWatermark(mockBody as any)
    setWatermark('Test')
    expect(addSpy).toHaveBeenCalledWith('resize', expect.any(Function))
    addSpy.mockRestore()
  })

  it('should remove resize listener when clear is called', async () => {
    const removeSpy = vi.spyOn(window, 'removeEventListener')
    const { useWatermark } = await import('@/hooks/web/useWatermark')
    const { setWatermark, clear } = useWatermark(mockBody as any)
    setWatermark('Test')
    clear()
    expect(removeSpy).toHaveBeenCalledWith('resize', expect.any(Function))
    removeSpy.mockRestore()
  })

  it('should call createWatermark which returns an id', async () => {
    const { useWatermark } = await import('@/hooks/web/useWatermark')
    const { setWatermark } = useWatermark(mockBody as any)
    setWatermark('Test')
    expect(mockAppendChild).toHaveBeenCalled()
    expect(mockToDataURL).toHaveBeenCalled()
  })

  it('should handle clear when no watermark DOM element exists', async () => {
    mockGetElementById.mockReturnValue(null)
    const { useWatermark } = await import('@/hooks/web/useWatermark')
    const { clear } = useWatermark(mockBody as any)
    expect(() => clear()).not.toThrow()
  })

  it('should remove watermark element when clear finds it', async () => {
    const mockDom = { id: 'test' }
    mockGetElementById.mockReturnValue(mockDom)
    const { useWatermark } = await import('@/hooks/web/useWatermark')
    const { setWatermark, clear } = useWatermark(mockBody as any)
    setWatermark('Test')
    clear()
    expect(mockRemoveChild).toHaveBeenCalledWith(mockDom)
  })
})
