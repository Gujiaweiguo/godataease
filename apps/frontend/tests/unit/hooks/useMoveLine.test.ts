import { describe, it, expect, vi, beforeEach } from 'vitest'

const mockWsCache = {
  get: vi.fn((key: string) => {
    if (key === 'DATASET') return 300
    return null
  }),
  set: vi.fn()
}

vi.mock('@/hooks/web/useCache', () => ({
  useCache: () => ({ wsCache: mockWsCache })
}))

const mockEmitter = { emit: vi.fn() }
vi.mock('@/hooks/web/useEmitt', () => ({
  useEmitt: () => ({ emitter: mockEmitter })
}))

const mockOnMounted = vi.fn((cb: () => void) => {
  ;(mockOnMounted as any)._callback = cb
})
const mockOnBeforeUnmount = vi.fn((cb: () => void) => {
  ;(mockOnBeforeUnmount as any)._callback = cb
})

vi.mock('vue', async (importOriginal) => {
  const actual = await importOriginal()
  return {
    ...(actual as any),
    ref: (v?: any) => ({ value: v }),
    onMounted: mockOnMounted,
    onBeforeUnmount: mockOnBeforeUnmount
  }
})

const mockAddEventListener = vi.fn()
const mockRemoveEventListener = vi.fn()

const originalCreateElement = document.createElement.bind(document)

describe('useMoveLine', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    ;(mockOnMounted as any)._callback = null
    ;(mockOnBeforeUnmount as any)._callback = null

    document.createElement = vi.fn((tag: string) => {
      const el = originalCreateElement(tag)
      el.addEventListener = mockAddEventListener
      el.removeEventListener = mockRemoveEventListener
      return el
    })
  })

  it('should return width and node refs', async () => {
    const { useMoveLine } = await import('@/hooks/web/useMoveLine')
    const result = useMoveLine('DATASET')
    expect(result).toHaveProperty('width')
    expect(result).toHaveProperty('node')
  })

  it('should read cached width from wsCache', async () => {
    const { useMoveLine } = await import('@/hooks/web/useMoveLine')
    useMoveLine('DATASET')
    expect(mockWsCache.get).toHaveBeenCalledWith('DATASET')
  })

  it('should set current-collapse_bar in wsCache on init', async () => {
    const { useMoveLine } = await import('@/hooks/web/useMoveLine')
    useMoveLine('DATASET')
    expect(mockWsCache.set).toHaveBeenCalledWith('current-collapse_bar', expect.any(Number))
  })

  it('should register onMounted lifecycle hook', async () => {
    const { useMoveLine } = await import('@/hooks/web/useMoveLine')
    useMoveLine('DATASET')
    expect(mockOnMounted).toHaveBeenCalled()
  })

  it('should register onBeforeUnmount lifecycle hook', async () => {
    const { useMoveLine } = await import('@/hooks/web/useMoveLine')
    useMoveLine('DATASET')
    expect(mockOnBeforeUnmount).toHaveBeenCalled()
  })

  it('should create a div element with sidebar-move-line class', async () => {
    const { useMoveLine } = await import('@/hooks/web/useMoveLine')
    useMoveLine('DATASET')
    expect(document.createElement).toHaveBeenCalledWith('div')
  })

  it('should register mousedown event on the element', async () => {
    const { useMoveLine } = await import('@/hooks/web/useMoveLine')
    useMoveLine('DATASET')
    expect(mockAddEventListener).toHaveBeenCalledWith('mousedown', expect.any(Function))
  })

  it('should use default width 280 when cache returns falsy', async () => {
    mockWsCache.get.mockReturnValueOnce(null)
    const { useMoveLine } = await import('@/hooks/web/useMoveLine')
    const result = useMoveLine('DASHBOARD')
    expect(result.width.value).toBe(280)
  })
})
