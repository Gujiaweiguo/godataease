import { describe, it, expect, vi, beforeEach } from 'vitest'

const { mockClose, mockService } = vi.hoisted(() => {
  const close = vi.fn()
  const service = vi.fn(() => ({ close }))
  return { mockClose: close, mockService: service }
})

vi.mock('element-plus-secondary', () => ({
  ElLoading: {
    service: mockService
  }
}))

import { useLoading } from '@/hooks/web/useLoading'

describe('useLoading', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should return open and close functions', () => {
    const { open, close } = useLoading()
    expect(typeof open).toBe('function')
    expect(typeof close).toBe('function')
  })

  it('should call ElLoading.service with fullscreen when open is called', () => {
    const { open } = useLoading()
    open()
    expect(mockService).toHaveBeenCalledWith({ fullscreen: true })
  })

  it('should call close on the loading instance when close is called', () => {
    const { open, close } = useLoading()
    open()
    close()
    expect(mockClose).toHaveBeenCalled()
  })

  it('should not throw when close is called without open', () => {
    const { close } = useLoading()
    expect(() => close()).not.toThrow()
  })

  it('should create new loading instance on each open call', () => {
    const { open } = useLoading()
    open()
    open()
    expect(mockService).toHaveBeenCalledTimes(2)
  })
})
