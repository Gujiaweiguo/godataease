import { describe, expect, it } from 'vitest'

import runAnimation from '@/utils/runAnimation'

function createMockElement() {
  const listeners: Record<string, ((...args: any[]) => void)[]> = {}
  return {
    style: {
      _props: {} as Record<string, string>,
      setProperty(prop: string, value: string) {
        this._props[prop] = value
      },
      removeProperty(prop: string) {
        delete this._props[prop]
      }
    },
    classList: {
      _classes: new Set<string>(),
      add(...tokens: string[]) {
        tokens.forEach(t => this._classes.add(t))
      },
      remove(...tokens: string[]) {
        tokens.forEach(t => this._classes.delete(t))
      },
      has(token: string) {
        return this._classes.has(token)
      }
    },
    addEventListener(event: string, fn: (...args: any[]) => void) {
      if (!listeners[event]) listeners[event] = []
      listeners[event].push(fn)
    },
    removeEventListener(event: string, fn: (...args: any[]) => void) {
      if (listeners[event]) {
        listeners[event] = listeners[event].filter(f => f !== fn)
      }
    },
    trigger(event: string) {
      if (listeners[event]) {
        listeners[event].forEach(fn => fn())
      }
    }
  }
}

describe('runAnimation', () => {
  it('resolves immediately with empty animations array', async () => {
    const $el = createMockElement()
    await runAnimation($el, [])
    expect($el.classList.has('animated')).toBe(false)
  })

  it('resolves immediately when animations is undefined', async () => {
    const $el = createMockElement()
    await runAnimation($el)
    expect($el.classList.has('animated')).toBe(false)
  })

  it('plays a single animation and removes classes after animationend', async () => {
    const $el = createMockElement()
    const animations = [{ animationTime: 1, value: 'fadeIn', isLoop: false }]

    const promise = runAnimation($el, animations)

    expect($el.classList.has('fadeIn')).toBe(true)
    expect($el.classList.has('animated')).toBe(true)
    expect($el.classList.has('no-infinite')).toBe(true)
    expect($el.style._props['--time']).toBe('1s')

    $el.trigger('animationend')

    await promise

    expect($el.classList.has('fadeIn')).toBe(false)
    expect($el.classList.has('animated')).toBe(false)
    expect($el.style._props['--time']).toBeUndefined()
  })

  it('adds infinite class when isLoop is true', async () => {
    const $el = createMockElement()
    const animations = [{ animationTime: 2, value: 'bounce', isLoop: true }]

    const promise = runAnimation($el, animations)

    expect($el.classList.has('infinite')).toBe(true)

    $el.trigger('animationend')
    await promise
  })

  it('plays multiple animations sequentially', async () => {
    const $el = createMockElement()
    const animations = [
      { animationTime: 1, value: 'fadeIn', isLoop: false },
      { animationTime: 1, value: 'bounce', isLoop: false }
    ]

    const promise = runAnimation($el, animations)

    expect($el.classList.has('fadeIn')).toBe(true)
    expect($el.classList.has('bounce')).toBe(false)

    $el.trigger('animationend')

    await Promise.resolve()

    expect($el.classList.has('fadeIn')).toBe(false)
    expect($el.classList.has('bounce')).toBe(true)

    $el.trigger('animationend')
    await promise

    expect($el.classList.has('bounce')).toBe(false)
  })

  it('cleans up on animationcancel event', async () => {
    const $el = createMockElement()
    const animations = [{ animationTime: 1, value: 'fadeIn', isLoop: false }]

    const promise = runAnimation($el, animations)

    expect($el.classList.has('fadeIn')).toBe(true)

    $el.trigger('animationcancel')

    await promise

    expect($el.classList.has('fadeIn')).toBe(false)
    expect($el.classList.has('animated')).toBe(false)
  })

  it('removes --time property after animation ends', async () => {
    const $el = createMockElement()
    const animations = [{ animationTime: 3, value: 'slideInLeft', isLoop: false }]

    const promise = runAnimation($el, animations)
    expect($el.style._props['--time']).toBe('3s')

    $el.trigger('animationend')
    await promise

    expect($el.style._props['--time']).toBeUndefined()
  })

  it('handles animation with empty value string', async () => {
    const $el = createMockElement()
    const animations = [{ animationTime: 1, value: '', isLoop: false }]

    const promise = runAnimation($el, animations)

    $el.trigger('animationend')
    await promise
  })
})
