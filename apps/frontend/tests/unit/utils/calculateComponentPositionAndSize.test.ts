import { describe, it, expect, vi } from 'vitest'

// translate.ts has module-level side effects (dvMainStore), must mock before import chain
vi.mock('@/store/modules/data-visualization/dvMain', () => ({
  dvMainStoreWithOut: () => ({})
}))

vi.mock('pinia', async () => {
  const actual = await vi.importActual('pinia')
  return {
    ...actual,
    storeToRefs: () => ({ canvasStyleData: { value: { scale: 100 } } })
  }
})

import calculateComponentPositionAndSize, {
  calculateRadioComponentPositionAndSize
} from '@/utils/calculateComponentPositionAndSize'

function makeStyle(overrides = {}) {
  return { left: 0, top: 0, width: 100, height: 100, rotate: 0, ...overrides }
}

describe('calculateComponentPositionAndSize', () => {
  describe('corner handlers at 0 rotation', () => {
    it('should resize via lt (left-top) handle', () => {
      const style = makeStyle()
      const curPosition = { x: 10, y: 10 }
      const pointInfo = { symmetricPoint: { x: 100, y: 100 } }
      calculateComponentPositionAndSize('lt', style, curPosition, 1, false, pointInfo)
      expect(style.left).toBe(10)
      expect(style.top).toBe(10)
      expect(style.width).toBe(90)
      expect(style.height).toBe(90)
    })

    it('should resize via rt (right-top) handle', () => {
      const style = makeStyle()
      const curPosition = { x: 90, y: 10 }
      const pointInfo = { symmetricPoint: { x: 0, y: 100 } }
      calculateComponentPositionAndSize('rt', style, curPosition, 1, false, pointInfo)
      expect(style.left).toBe(0)
      expect(style.top).toBe(10)
      expect(style.width).toBe(90)
      expect(style.height).toBe(90)
    })

    it('should resize via rb (right-bottom) handle', () => {
      const style = makeStyle()
      const curPosition = { x: 110, y: 110 }
      const pointInfo = { symmetricPoint: { x: 0, y: 0 } }
      calculateComponentPositionAndSize('rb', style, curPosition, 1, false, pointInfo)
      expect(style.left).toBe(0)
      expect(style.top).toBe(0)
      expect(style.width).toBe(110)
      expect(style.height).toBe(110)
    })

    it('should resize via lb (left-bottom) handle', () => {
      const style = makeStyle()
      const curPosition = { x: -10, y: 110 }
      const pointInfo = { symmetricPoint: { x: 100, y: 0 } }
      calculateComponentPositionAndSize('lb', style, curPosition, 1, false, pointInfo)
      expect(style.left).toBe(-10)
      expect(style.top).toBe(0)
      expect(style.width).toBe(110)
      expect(style.height).toBe(110)
    })
  })

  describe('edge handlers at 0 rotation', () => {
    it('should resize via t (top) handle - height only', () => {
      const style = makeStyle()
      const curPosition = { x: 50, y: 20 }
      const pointInfo = {
        symmetricPoint: { x: 50, y: 100 },
        curPoint: { x: 50, y: 0 }
      }
      calculateComponentPositionAndSize('t', style, curPosition, 1, false, pointInfo)
      expect(style.height).toBe(80)
    })

    it('should resize via b (bottom) handle - height only', () => {
      const style = makeStyle()
      const curPosition = { x: 50, y: 120 }
      const pointInfo = {
        symmetricPoint: { x: 50, y: 0 },
        curPoint: { x: 50, y: 100 }
      }
      calculateComponentPositionAndSize('b', style, curPosition, 1, false, pointInfo)
      expect(style.height).toBe(120)
    })

    it('should resize via r (right) handle - width only', () => {
      const style = makeStyle()
      const curPosition = { x: 120, y: 50 }
      const pointInfo = {
        symmetricPoint: { x: 0, y: 50 },
        curPoint: { x: 100, y: 50 }
      }
      calculateComponentPositionAndSize('r', style, curPosition, 1, false, pointInfo)
      expect(style.width).toBe(120)
    })

    it('should resize via l (left) handle - width only', () => {
      const style = makeStyle()
      const curPosition = { x: -20, y: 50 }
      const pointInfo = {
        symmetricPoint: { x: 100, y: 50 },
        curPoint: { x: 0, y: 50 }
      }
      calculateComponentPositionAndSize('l', style, curPosition, 1, false, pointInfo)
      expect(style.width).toBe(120)
    })
  })

  describe('proportion locking', () => {
    it('should lock proportion when resizing via lt with needLockProportion=true', () => {
      const style = makeStyle()
      const curPosition = { x: 20, y: 20 }
      const pointInfo = { symmetricPoint: { x: 100, y: 100 } }
      const proportion = 1 // 1:1 aspect ratio
      calculateComponentPositionAndSize('lt', style, curPosition, proportion, true, pointInfo)
      expect(style.width).toBe(style.height)
    })

    it('should lock proportion when resizing via rb with needLockProportion=true', () => {
      const style = makeStyle()
      const curPosition = { x: 150, y: 130 }
      const pointInfo = { symmetricPoint: { x: 0, y: 0 } }
      const proportion = 1
      calculateComponentPositionAndSize('rb', style, curPosition, proportion, true, pointInfo)
      expect(style.width).toBe(style.height)
    })
  })

  describe('no update for negative dimensions', () => {
    it('should not update style when lt drag produces negative width', () => {
      const style = makeStyle()
      const curPosition = { x: 200, y: 200 }
      const pointInfo = { symmetricPoint: { x: 100, y: 100 } }
      const originalStyle = { ...style }
      calculateComponentPositionAndSize('lt', style, curPosition, 1, false, pointInfo)
      expect(style.left).toBe(originalStyle.left)
      expect(style.top).toBe(originalStyle.top)
      expect(style.width).toBe(originalStyle.width)
      expect(style.height).toBe(originalStyle.height)
    })
  })
})

describe('calculateRadioComponentPositionAndSize', () => {
  it('should position for "b" (bottom) direction', () => {
    const style = makeStyle({ width: 60, height: 40 })
    calculateRadioComponentPositionAndSize('b', style, { x: 100, y: 200 })
    expect(style.left).toBe(70) // 100 - 60/2
    expect(style.top).toBe(200)
  })

  it('should position for "t" (top) direction', () => {
    const style = makeStyle({ width: 60, height: 40 })
    calculateRadioComponentPositionAndSize('t', style, { x: 100, y: 200 })
    expect(style.left).toBe(70)
    expect(style.top).toBe(160) // 200 - 40
  })

  it('should position for "l" (left) direction', () => {
    const style = makeStyle({ width: 60, height: 40 })
    calculateRadioComponentPositionAndSize('l', style, { x: 100, y: 200 })
    expect(style.left).toBe(40) // 100 - 60
    expect(style.top).toBe(180) // 200 - 40/2
  })

  it('should position for "r" (right) direction', () => {
    const style = makeStyle({ width: 60, height: 40 })
    calculateRadioComponentPositionAndSize('r', style, { x: 100, y: 200 })
    expect(style.left).toBe(100)
    expect(style.top).toBe(180)
  })

  it('should position for "lt" (left-top) direction', () => {
    const style = makeStyle({ width: 60, height: 40 })
    calculateRadioComponentPositionAndSize('lt', style, { x: 100, y: 200 })
    expect(style.left).toBe(40)
    expect(style.top).toBe(160)
  })

  it('should position for "lb" (left-bottom) direction', () => {
    const style = makeStyle({ width: 60, height: 40 })
    calculateRadioComponentPositionAndSize('lb', style, { x: 100, y: 200 })
    expect(style.left).toBe(40)
    expect(style.top).toBe(200)
  })

  it('should position for "rt" (right-top) direction', () => {
    const style = makeStyle({ width: 60, height: 40 })
    calculateRadioComponentPositionAndSize('rt', style, { x: 100, y: 200 })
    expect(style.left).toBe(100)
    expect(style.top).toBe(160)
  })

  it('should position for "rb" (right-bottom) direction', () => {
    const style = makeStyle({ width: 60, height: 40 })
    calculateRadioComponentPositionAndSize('rb', style, { x: 100, y: 200 })
    expect(style.left).toBe(100)
    expect(style.top).toBe(200)
  })
})
