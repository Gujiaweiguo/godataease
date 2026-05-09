import { describe, it, expect, vi, beforeEach } from 'vitest'

import { withDatasourceError } from '@/api/decorators/datasourceErrorDecorator'

describe('withDatasourceError decorator', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('returns resolved value when apiCall succeeds', async () => {
    const apiCall = vi.fn().mockResolvedValue({ data: { id: '1' } })
    const result = await withDatasourceError(apiCall)
    expect(result).toEqual({ data: { id: '1' } })
  })

  it('calls the apiCall function exactly once', async () => {
    const apiCall = vi.fn().mockResolvedValue('ok')
    await withDatasourceError(apiCall)
    expect(apiCall).toHaveBeenCalledTimes(1)
  })

  it('re-throws Error instances without wrapping', async () => {
    const originalError = new Error('Network timeout')
    const apiCall = vi.fn().mockRejectedValue(originalError)
    await expect(withDatasourceError(apiCall)).rejects.toThrow('Network timeout')
  })

  it('re-throws Error instances preserving the original reference', async () => {
    const originalError = new TypeError('Invalid data')
    const apiCall = vi.fn().mockRejectedValue(originalError)
    await expect(withDatasourceError(apiCall)).rejects.toBe(originalError)
  })

  it('wraps non-Error rejections in a generic Error', async () => {
    const apiCall = vi.fn().mockRejectedValue('string error')
    await expect(withDatasourceError(apiCall)).rejects.toThrow('Datasource request failed')
  })

  it('wraps null rejections in a generic Error', async () => {
    const apiCall = vi.fn().mockRejectedValue(null)
    await expect(withDatasourceError(apiCall)).rejects.toThrow('Datasource request failed')
  })

  it('wraps numeric rejections in a generic Error', async () => {
    const apiCall = vi.fn().mockRejectedValue(500)
    await expect(withDatasourceError(apiCall)).rejects.toThrow('Datasource request failed')
  })

  it('wraps object rejections in a generic Error', async () => {
    const apiCall = vi.fn().mockRejectedValue({ code: 'ERR', msg: 'fail' })
    await expect(withDatasourceError(apiCall)).rejects.toThrow('Datasource request failed')
  })

  it('wraps undefined rejections in a generic Error', async () => {
    const apiCall = vi.fn().mockRejectedValue(undefined)
    await expect(withDatasourceError(apiCall)).rejects.toThrow('Datasource request failed')
  })

  it('wrapped non-Error rejection is an Error instance', async () => {
    const apiCall = vi.fn().mockRejectedValue('raw string')
    try {
      await withDatasourceError(apiCall)
      expect.fail('Should have thrown')
    } catch (e) {
      expect(e).toBeInstanceOf(Error)
      expect((e as Error).message).toBe('Datasource request failed')
    }
  })

  it('preserves Error subclass when re-throwing', async () => {
    const syntaxError = new SyntaxError('bad syntax')
    const apiCall = vi.fn().mockRejectedValue(syntaxError)
    try {
      await withDatasourceError(apiCall)
      expect.fail('Should have thrown')
    } catch (e) {
      expect(e).toBeInstanceOf(SyntaxError)
      expect((e as SyntaxError).message).toBe('bad syntax')
    }
  })

  it('works with resolved undefined value', async () => {
    const apiCall = vi.fn().mockResolvedValue(undefined)
    const result = await withDatasourceError(apiCall)
    expect(result).toBeUndefined()
  })
})
