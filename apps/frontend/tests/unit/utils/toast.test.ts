import { describe, it, expect, vi, afterEach } from 'vitest'

vi.mock('element-plus-secondary', () => ({
  ElMessage: {
    error: vi.fn()
  }
}))

import { ElMessage } from 'element-plus-secondary'
import toast from '@/utils/toast'

describe('toast', () => {
  afterEach(() => {
    vi.clearAllMocks()
  })

  it('should call ElMessage.error with the provided message', () => {
    toast('Something went wrong')
    expect(ElMessage.error).toHaveBeenCalledWith('Something went wrong')
  })

  it('should call ElMessage.error with empty string when no argument', () => {
    toast()
    expect(ElMessage.error).toHaveBeenCalledWith('')
  })

  it('should call ElMessage.error with empty string when called with empty string', () => {
    toast('')
    expect(ElMessage.error).toHaveBeenCalledWith('')
  })

  it('should call ElMessage.error exactly once per call', () => {
    toast('error')
    expect(ElMessage.error).toHaveBeenCalledTimes(1)
  })

  it('should pass through multi-line messages', () => {
    const msg = 'line1\nline2\nline3'
    toast(msg)
    expect(ElMessage.error).toHaveBeenCalledWith(msg)
  })

  it('should pass through long messages', () => {
    const msg = 'a'.repeat(500)
    toast(msg)
    expect(ElMessage.error).toHaveBeenCalledWith(msg)
  })

  it('should handle multiple sequential calls', () => {
    toast('first')
    toast('second')
    expect(ElMessage.error).toHaveBeenCalledTimes(2)
    expect(ElMessage.error).toHaveBeenNthCalledWith(1, 'first')
    expect(ElMessage.error).toHaveBeenNthCalledWith(2, 'second')
  })
})
