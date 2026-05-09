import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { loadScript } from '@/utils/RemoteJs'

describe('RemoteJs', () => {
  let createdScripts: { id: string; src: string; onload: ((...args: any[]) => void) | null; onerror: ((...args: any[]) => void) | null }[] = []

  const originalCreateElement = document.createElement.bind(document)

  beforeEach(() => {
    createdScripts = []
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  function setupMocks(appendTarget: any) {
    vi.spyOn(document, 'getElementById').mockReturnValue(null)
    vi.spyOn(document, 'createElement').mockImplementation((tag: string) => {
      if (tag === 'script') {
        const mock: any = {
          id: '',
          src: '',
          onload: null as ((...args: any[]) => void) | null,
          onerror: null as ((...args: any[]) => void) | null
        }
        createdScripts.push(mock)
        return mock
      }
      return originalCreateElement(tag)
    })
    vi.spyOn(document, 'body', 'get').mockReturnValue(appendTarget)
    vi.spyOn(document, 'head', 'get').mockReturnValue(appendTarget)
  }

  it('creates a script element and appends to body', async () => {
    const appendSpy = vi.fn()
    setupMocks({ appendChild: appendSpy })

    const promise = loadScript('https://cdn.example.com/lib.js')

    expect(appendSpy).toHaveBeenCalledWith(createdScripts[0])

    createdScripts[0].onload?.(null as any)
    await expect(promise).resolves.toBeNull()
  })

  it('uses default jsId when not provided', async () => {
    setupMocks({ appendChild: vi.fn() })

    const promise = loadScript('https://cdn.example.com/lib.js')
    expect(createdScripts[0].id).toBe('de-fit2cloud-script-id')

    createdScripts[0].onload?.(null as any)
    await expect(promise).resolves.toBeNull()
  })

  it('uses custom jsId when provided', async () => {
    setupMocks({ appendChild: vi.fn() })

    const promise = loadScript('https://cdn.example.com/lib.js', 'my-custom-id')
    expect(createdScripts[0].id).toBe('my-custom-id')

    createdScripts[0].onload?.(null as any)
    await expect(promise).resolves.toBeNull()
  })

  it('removes existing script with same id before creating new one', async () => {
    const removeChild = vi.fn()
    const existingScript: any = {
      id: 'de-fit2cloud-script-id',
      parentElement: { removeChild }
    }
    setupMocks({ appendChild: vi.fn() })
    vi.spyOn(document, 'getElementById').mockReturnValue(existingScript)

    const promise = loadScript('https://cdn.example.com/lib.js')

    expect(removeChild).toHaveBeenCalledWith(existingScript)

    createdScripts[0].onload?.(null as any)
    await expect(promise).resolves.toBeNull()
  })

  it('rejects promise on script load error', async () => {
    setupMocks({ appendChild: vi.fn() })

    const promise = loadScript('https://cdn.example.com/broken.js')

    createdScripts[0].onerror?.(null as any)
    await expect(promise).rejects.toThrow(
      'Load script from https://cdn.example.com/broken.js failed'
    )
  })

  it('sets the script src to the provided url', async () => {
    setupMocks({ appendChild: vi.fn() })

    const promise = loadScript('https://cdn.example.com/lib.js')

    expect(createdScripts[0].src).toBe('https://cdn.example.com/lib.js')

    createdScripts[0].onload?.(null as any)
    await expect(promise).resolves.toBeNull()
  })
})
