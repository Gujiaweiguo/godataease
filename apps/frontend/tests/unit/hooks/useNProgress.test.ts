import { describe, it, expect, vi, beforeEach } from 'vitest'

const { mockConfigure, mockStart, mockDone, mockNProgress } = vi.hoisted(() => {
  const configure = vi.fn()
  const start = vi.fn()
  const done = vi.fn()
  return {
    mockConfigure: configure,
    mockStart: start,
    mockDone: done,
    mockNProgress: { configure, start, done }
  }
})

vi.mock('nprogress', () => ({
  __esModule: true,
  default: mockNProgress
}))

vi.mock('nprogress/nprogress.css', () => ({}))

vi.mock('@vueuse/core', () => ({
  useCssVar: () => ({ value: '#409EFF' })
}))

vi.mock('vue', () => ({
  nextTick: () => Promise.resolve(),
  unref: (v: any) => (typeof v === 'object' && v !== null && 'value' in v ? v.value : v)
}))

import { useNProgress } from '@/hooks/web/useNProgress'

describe('useNProgress', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should return start and done functions', () => {
    const { start, done } = useNProgress()
    expect(typeof start).toBe('function')
    expect(typeof done).toBe('function')
  })

  it('should configure NProgress with showSpinner false', () => {
    useNProgress()
    expect(mockConfigure).toHaveBeenCalledWith({ showSpinner: false })
  })

  it('should call NProgress.start when start is called', () => {
    const { start } = useNProgress()
    start()
    expect(mockStart).toHaveBeenCalled()
  })

  it('should call NProgress.done when done is called', () => {
    const { done } = useNProgress()
    done()
    expect(mockDone).toHaveBeenCalled()
  })

  it('should not throw when initialized', () => {
    expect(() => useNProgress()).not.toThrow()
  })

  it('should allow calling start and done in sequence', () => {
    const { start, done } = useNProgress()
    start()
    done()
    expect(mockStart).toHaveBeenCalledTimes(1)
    expect(mockDone).toHaveBeenCalledTimes(1)
  })
})
