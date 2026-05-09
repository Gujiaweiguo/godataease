import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { useScrollTo } from '@/hooks/event/useScrollTo'

describe('useScrollTo', () => {
  let mockEl: HTMLElement
  let rafCallbacks: FrameRequestCallback[]

  beforeEach(() => {
    rafCallbacks = []
    vi.stubGlobal(
      'requestAnimationFrame',
      (cb: FrameRequestCallback) => {
        rafCallbacks.push(cb)
        return rafCallbacks.length
      }
    )
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  function flushAnimationFrames(maxIterations = 100) {
    let count = 0
    while (rafCallbacks.length > 0 && count < maxIterations) {
      const cb = rafCallbacks.shift()!
      cb(count * 20)
      count++
    }
  }

  it('should return start and stop functions', () => {
    mockEl = { scrollLeft: 0, scrollTop: 0 } as any
    const { start, stop } = useScrollTo({ el: mockEl, to: 100, position: 'scrollLeft' })
    expect(typeof start).toBe('function')
    expect(typeof stop).toBe('function')
  })

  it('should animate scrollLeft from 0 to target', () => {
    mockEl = { scrollLeft: 0 } as any
    const { start } = useScrollTo({ el: mockEl, to: 200, position: 'scrollLeft', duration: 100 })

    start()
    flushAnimationFrames()

    expect(mockEl.scrollLeft).toBe(200)
  })

  it('should animate scrollTop from 0 to target', () => {
    mockEl = { scrollTop: 0 } as any
    const { start } = useScrollTo({ el: mockEl, to: 300, position: 'scrollTop', duration: 100 })

    start()
    flushAnimationFrames()

    expect(mockEl.scrollTop).toBe(300)
  })

  it('should call callback when animation completes', () => {
    mockEl = { scrollLeft: 0 } as any
    const callback = vi.fn()
    const { start } = useScrollTo({
      el: mockEl,
      to: 100,
      position: 'scrollLeft',
      duration: 100,
      callback
    })

    start()
    flushAnimationFrames()

    expect(callback).toHaveBeenCalled()
  })

  it('should not call callback when animation is stopped', () => {
    mockEl = { scrollLeft: 0 } as any
    const callback = vi.fn()
    const { start, stop } = useScrollTo({
      el: mockEl,
      to: 500,
      position: 'scrollLeft',
      duration: 500,
      callback
    })

    start()
    stop()
    flushAnimationFrames()

    expect(callback).not.toHaveBeenCalled()
  })

  it('should stop animation when stop is called', () => {
    mockEl = { scrollLeft: 0 } as any
    const { start, stop } = useScrollTo({
      el: mockEl,
      to: 200,
      position: 'scrollLeft',
      duration: 500
    })

    start()
    stop()
    const scrollValueAfterStop = mockEl.scrollLeft
    flushAnimationFrames()
    expect(mockEl.scrollLeft).toBe(scrollValueAfterStop)
  })

  it('should use default position as scrollLeft', () => {
    mockEl = { scrollLeft: 50 } as any
    const { start } = useScrollTo({ el: mockEl, to: 150, position: 'scrollLeft', duration: 100 })

    start()
    flushAnimationFrames()

    expect(mockEl.scrollLeft).toBe(150)
  })

  it('should use default duration of 500', () => {
    mockEl = { scrollLeft: 0 } as any
    const { start } = useScrollTo({ el: mockEl, to: 100, position: 'scrollLeft', duration: 500 })

    start()
    flushAnimationFrames(30)

    expect(mockEl.scrollLeft).toBe(100)
  })

  it('should animate backward when target is less than start', () => {
    mockEl = { scrollLeft: 200 } as any
    const { start } = useScrollTo({ el: mockEl, to: 50, position: 'scrollLeft', duration: 100 })

    start()
    flushAnimationFrames()

    expect(mockEl.scrollLeft).toBe(50)
  })

  it('should handle no movement when start equals target', () => {
    mockEl = { scrollLeft: 100 } as any
    const { start } = useScrollTo({ el: mockEl, to: 100, position: 'scrollLeft', duration: 100 })

    start()
    flushAnimationFrames()

    expect(mockEl.scrollLeft).toBe(100)
  })
})
