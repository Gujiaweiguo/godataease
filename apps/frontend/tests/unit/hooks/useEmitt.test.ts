import { describe, it, expect, vi, beforeEach } from 'vitest'

const { mittInstance, mockOnBeforeUnmount } = vi.hoisted(() => {
  const instance = {
    on: vi.fn(),
    off: vi.fn(),
    emit: vi.fn(),
    all: new Map()
  }
  const mockFn = vi.fn((cb: () => void) => {
    ;(mockFn as any)._callback = cb
  })
  ;(mockFn as any)._callback = null
  return { mittInstance: instance, mockOnBeforeUnmount: mockFn }
})

vi.mock('mitt', () => {
  return {
    default: () => mittInstance
  }
})

vi.mock('vue', async (importOriginal) => {
  const actual = await importOriginal()
  return {
    ...(actual as any),
    onBeforeUnmount: mockOnBeforeUnmount
  }
})

import { useEmitt } from '@/hooks/web/useEmitt'

describe('useEmitt', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    ;(mockOnBeforeUnmount as any)._callback = null
  })

  it('should return emitter when called without option', () => {
    const { emitter } = useEmitt()
    expect(emitter).toBeDefined()
    expect(typeof emitter.on).toBe('function')
    expect(typeof emitter.off).toBe('function')
    expect(typeof emitter.emit).toBe('function')
  })

  it('should register event listener when option is provided', () => {
    const callback = vi.fn()
    useEmitt({ name: 'test-event', callback })

    expect(mittInstance.on).toHaveBeenCalledWith('test-event', callback)
  })

  it('should register cleanup on onBeforeUnmount when option is provided', () => {
    const callback = vi.fn()
    useEmitt({ name: 'test-event', callback })

    expect(mockOnBeforeUnmount).toHaveBeenCalled()
  })

  it('should unregister event listener on unmount', () => {
    const callback = vi.fn()
    useEmitt({ name: 'test-event', callback })

    const cleanup = (mockOnBeforeUnmount as any)._callback as () => void
    cleanup()

    expect(mittInstance.off).toHaveBeenCalledWith('test-event', callback)
  })

  it('should not register listener when no option is provided', () => {
    useEmitt()

    expect(mittInstance.on).not.toHaveBeenCalled()
  })

  it('should not register onBeforeUnmount when no option is provided', () => {
    useEmitt()

    expect(mockOnBeforeUnmount).not.toHaveBeenCalled()
  })

  it('should support emitting events via emitter', () => {
    const { emitter } = useEmitt()
    emitter.emit('custom-event', { data: 42 })

    expect(mittInstance.emit).toHaveBeenCalledWith('custom-event', { data: 42 })
  })

  it('should support registering multiple event listeners', () => {
    const cb1 = vi.fn()
    const cb2 = vi.fn()

    useEmitt({ name: 'event-1', callback: cb1 })
    useEmitt({ name: 'event-2', callback: cb2 })

    expect(mittInstance.on).toHaveBeenCalledWith('event-1', cb1)
    expect(mittInstance.on).toHaveBeenCalledWith('event-2', cb2)
  })

  it('should return same emitter instance across calls (singleton)', () => {
    const result1 = useEmitt()
    const result2 = useEmitt()

    expect(result1.emitter).toBe(result2.emitter)
  })

  it('should handle string event names', () => {
    const callback = vi.fn()
    useEmitt({ name: 'my-custom-event', callback })

    expect(mittInstance.on).toHaveBeenCalledWith('my-custom-event', callback)
  })
})
