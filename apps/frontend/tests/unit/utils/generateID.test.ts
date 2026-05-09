import { describe, it, expect, vi } from 'vitest'
import { generateID } from '@/utils/generateID'

vi.mock('@/views/visualized/data/dataset/form/util.js', () => ({
  guid: vi.fn(() => 'mock-guid-123')
}))

describe('generateID', () => {
  it('should return the result from guid()', () => {
    const result = generateID()
    expect(result).toBe('mock-guid-123')
  })

  it('should call guid each time generateID is called', async () => {
    generateID()
    generateID()
    const { guid } = await import('@/views/visualized/data/dataset/form/util.js')
    expect(guid).toHaveBeenCalledTimes(3)
  })

  it('should return different values when guid returns different values', async () => {
    const { guid } = await import('@/views/visualized/data/dataset/form/util.js')
    vi.mocked(guid).mockReturnValueOnce('id-aaa').mockReturnValueOnce('id-bbb')
    expect(generateID()).toBe('id-aaa')
    expect(generateID()).toBe('id-bbb')
  })
})
