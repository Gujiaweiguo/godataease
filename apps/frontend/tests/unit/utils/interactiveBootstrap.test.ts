import { describe, expect, it, vi } from 'vitest'

import { bootstrapInteractiveInBackground } from '@/utils/interactiveBootstrap'

describe('bootstrapInteractiveInBackground', () => {
  it('同步清空缓存并在后台启动交互树预热', async () => {
    const clear = vi.fn()
    const initInteractive = vi.fn().mockResolvedValue(undefined)

    bootstrapInteractiveInBackground({ clear, initInteractive })

    expect(clear).toHaveBeenCalledTimes(1)
    expect(initInteractive).toHaveBeenCalledWith(true)
  })

  it('交互树预热失败时不向外抛错', async () => {
    const clear = vi.fn()
    const initInteractive = vi.fn().mockRejectedValue(new Error('interactive tree timeout'))

    expect(() => bootstrapInteractiveInBackground({ clear, initInteractive })).not.toThrow()
    await Promise.resolve()
    await Promise.resolve()
    expect(clear).toHaveBeenCalledTimes(1)
  })
})
