import { describe, expect, it } from 'vitest'

import decomposeComponent from '@/utils/decomposeComponent'

describe('decomposeComponent', () => {
  const createComponent = (left = 10, top = 20, groupStyle?: Record<string, number>) => ({
    style: { left, top },
    groupStyle,
    canvasId: ''
  })

  it('adds parentStyle left and top to component style', () => {
    const component = createComponent()
    decomposeComponent(component, null, { left: 100, top: 50 })
    expect(component.style.left).toBe(110)
    expect(component.style.top).toBe(70)
  })

  it('sets canvasId to provided value', () => {
    const component = createComponent()
    decomposeComponent(component, null, { left: 0, top: 0 }, 'canvas-0')
    expect(component.canvasId).toBe('canvas-0')
  })

  it('defaults canvasId to "canvas-main"', () => {
    const component = createComponent()
    decomposeComponent(component, null, { left: 0, top: 0 })
    expect(component.canvasId).toBe('canvas-main')
  })

  it('does not modify groupStyle when parentGroupStyle is not provided', () => {
    const component = createComponent(10, 20, { left: 5, top: 10, width: 100, height: 200 })
    decomposeComponent(component, null, { left: 50, top: 30 })
    expect(component.groupStyle.left).toBe(5)
    expect(component.groupStyle.top).toBe(10)
  })

  it('does not modify groupStyle when component has no groupStyle', () => {
    const component = createComponent()
    decomposeComponent(component, null, { left: 10, top: 20 }, 'canvas-main', {
      left: 50,
      top: 60,
      width: 200,
      height: 300
    })
    expect(component.groupStyle).toBeUndefined()
  })

  it('recalculates groupStyle when both parentGroupStyle and component groupStyle exist', () => {
    const component = createComponent(10, 20, { left: 25, top: 50, width: 100, height: 100 })
    const parentStyle = { left: 200, top: 100 }
    const parentGroupStyle = { left: 0, top: 0, width: 2, height: 3 }

    decomposeComponent(component, null, parentStyle, 'canvas-main', parentGroupStyle)

    expect(component.groupStyle.width).toBe(200)
    expect(component.groupStyle.height).toBe(300)
    expect(component.groupStyle.left).toBe(250)
    expect(component.groupStyle.top).toBe(250)
  })

  it('updates component style left/top and also sets groupStyle', () => {
    const component = createComponent(5, 10, { left: 10, top: 20, width: 50, height: 50 })
    decomposeComponent(component, null, { left: 100, top: 200 }, 'canvas-main', {
      left: 0,
      top: 0,
      width: 3,
      height: 2
    })
    expect(component.style.left).toBe(105)
    expect(component.style.top).toBe(210)
    expect(component.groupStyle.width).toBe(150)
    expect(component.groupStyle.height).toBe(100)
  })

  it('handles zero parentStyle values', () => {
    const component = createComponent(10, 20)
    decomposeComponent(component, null, { left: 0, top: 0 })
    expect(component.style.left).toBe(10)
    expect(component.style.top).toBe(20)
  })

  it('handles negative component positions', () => {
    const component = createComponent(-5, -10)
    decomposeComponent(component, null, { left: 100, top: 50 })
    expect(component.style.left).toBe(95)
    expect(component.style.top).toBe(40)
  })
})
