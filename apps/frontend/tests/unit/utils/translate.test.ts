import { describe, it, expect, vi } from 'vitest'

// Mock dvMainStore before importing translate (module-level side effects)
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

import {
  calculateRotatedPointCoordinate,
  sin,
  cos,
  mod360,
  toPercent,
  getCenterPoint,
  getRotatedPointCoordinate,
  changeStyleWithScale
} from '@/utils/translate'

describe('translate utils', () => {
  const EPSILON = 1e-10

  describe('calculateRotatedPointCoordinate', () => {
    it('should return the same point when rotated by 0 degrees', () => {
      const result = calculateRotatedPointCoordinate({ x: 1, y: 0 }, { x: 0, y: 0 }, 0)
      expect(result.x).toBeCloseTo(1, 10)
      expect(result.y).toBeCloseTo(0, 10)
    })

    it('should rotate (1,0) around origin by 90 degrees to ~(0,1)', () => {
      const result = calculateRotatedPointCoordinate({ x: 1, y: 0 }, { x: 0, y: 0 }, 90)
      expect(result.x).toBeCloseTo(0, 10)
      expect(result.y).toBeCloseTo(1, 10)
    })

    it('should rotate (1,1) around origin by 180 degrees to (-1,-1)', () => {
      const result = calculateRotatedPointCoordinate({ x: 1, y: 1 }, { x: 0, y: 0 }, 180)
      expect(result.x).toBeCloseTo(-1, 10)
      expect(result.y).toBeCloseTo(-1, 10)
    })

    it('should rotate around a non-origin center', () => {
      // Point (3,2) rotated 90° around center (2,2)
      const result = calculateRotatedPointCoordinate({ x: 3, y: 2 }, { x: 2, y: 2 }, 90)
      // After rotation: relative (1,0) → (0,1), then +center = (2,3)
      expect(result.x).toBeCloseTo(2, 10)
      expect(result.y).toBeCloseTo(3, 10)
    })

    it('should be identity for 360 degree rotation', () => {
      const point = { x: 5, y: 7 }
      const center = { x: 2, y: 3 }
      const result = calculateRotatedPointCoordinate(point, center, 360)
      expect(result.x).toBeCloseTo(point.x, 10)
      expect(result.y).toBeCloseTo(point.y, 10)
    })

    it('should handle negative rotation angles', () => {
      // Rotating -90° is same as 270°
      const result = calculateRotatedPointCoordinate({ x: 1, y: 0 }, { x: 0, y: 0 }, -90)
      expect(result.x).toBeCloseTo(0, 10)
      expect(result.y).toBeCloseTo(-1, 10)
    })
  })

  describe('sin', () => {
    it('should return 0 for 0 degrees', () => {
      expect(sin(0)).toBeCloseTo(0, 10)
    })

    it('should return 1 for 90 degrees', () => {
      expect(sin(90)).toBeCloseTo(1, 10)
    })

    it('should return 0 for 180 degrees (abs)', () => {
      expect(sin(180)).toBeCloseTo(0, 10)
    })

    it('should return approximately 0.707 for 45 degrees', () => {
      expect(sin(45)).toBeCloseTo(Math.SQRT1_2, 10)
    })

    it('should always return non-negative values (abs)', () => {
      expect(sin(270)).toBeCloseTo(1, 10)
    })
  })

  describe('cos', () => {
    it('should return 1 for 0 degrees', () => {
      expect(cos(0)).toBeCloseTo(1, 10)
    })

    it('should return 0 for 90 degrees', () => {
      expect(Math.abs(cos(90) - 0)).toBeLessThan(EPSILON)
    })

    it('should return 1 for 180 degrees (abs)', () => {
      expect(cos(180)).toBeCloseTo(1, 10)
    })

    it('should return approximately 0.707 for 45 degrees', () => {
      expect(cos(45)).toBeCloseTo(Math.SQRT1_2, 10)
    })
  })

  describe('mod360', () => {
    it('should return 0 for 0', () => {
      expect(mod360(0)).toBe(0)
    })

    it('should return 0 for 360', () => {
      expect(mod360(360)).toBe(0)
    })

    it('should convert -90 to 270', () => {
      expect(mod360(-90)).toBe(270)
    })

    it('should convert 720 to 0', () => {
      expect(mod360(720)).toBe(0)
    })

    it('should leave 180 unchanged', () => {
      expect(mod360(180)).toBe(180)
    })

    it('should convert -1 to 359', () => {
      expect(mod360(-1)).toBe(359)
    })
  })

  describe('toPercent', () => {
    it('should convert 0.5 to 50%', () => {
      expect(toPercent(0.5)).toBe('50%')
    })

    it('should convert 1 to 100%', () => {
      expect(toPercent(1)).toBe('100%')
    })

    it('should convert 0 to 0%', () => {
      expect(toPercent(0)).toBe('0%')
    })

    it('should convert 0.255 to 25.5%', () => {
      expect(toPercent(0.255)).toBe('25.5%')
    })
  })

  describe('getCenterPoint', () => {
    it('should return midpoint of (0,0) and (2,2)', () => {
      const result = getCenterPoint({ x: 0, y: 0 }, { x: 2, y: 2 })
      expect(result).toEqual({ x: 1, y: 1 })
    })

    it('should return midpoint of (-1,-1) and (1,1)', () => {
      const result = getCenterPoint({ x: -1, y: -1 }, { x: 1, y: 1 })
      expect(result).toEqual({ x: 0, y: 0 })
    })

    it('should return midpoint for same point', () => {
      const result = getCenterPoint({ x: 5, y: 5 }, { x: 5, y: 5 })
      expect(result).toEqual({ x: 5, y: 5 })
    })
  })

  describe('getRotatedPointCoordinate', () => {
    const style = { left: 0, top: 0, width: 100, height: 100, rotate: 0 }
    const center = { x: 50, y: 50 }

    it('should return top-center point for "t"', () => {
      const result = getRotatedPointCoordinate(style, center, 't')
      expect(result.x).toBeCloseTo(50, 10)
      expect(result.y).toBeCloseTo(0, 10)
    })

    it('should return bottom-center point for "b"', () => {
      const result = getRotatedPointCoordinate(style, center, 'b')
      expect(result.x).toBeCloseTo(50, 10)
      expect(result.y).toBeCloseTo(100, 10)
    })

    it('should return left-center point for "l"', () => {
      const result = getRotatedPointCoordinate(style, center, 'l')
      expect(result.x).toBeCloseTo(0, 10)
      expect(result.y).toBeCloseTo(50, 10)
    })

    it('should return right-center point for "r"', () => {
      const result = getRotatedPointCoordinate(style, center, 'r')
      expect(result.x).toBeCloseTo(100, 10)
      expect(result.y).toBeCloseTo(50, 10)
    })

    it('should return top-left point for "lt"', () => {
      const result = getRotatedPointCoordinate(style, center, 'lt')
      expect(result.x).toBeCloseTo(0, 10)
      expect(result.y).toBeCloseTo(0, 10)
    })

    it('should return top-right point for "rt"', () => {
      const result = getRotatedPointCoordinate(style, center, 'rt')
      expect(result.x).toBeCloseTo(100, 10)
      expect(result.y).toBeCloseTo(0, 10)
    })

    it('should return bottom-left point for "lb"', () => {
      const result = getRotatedPointCoordinate(style, center, 'lb')
      expect(result.x).toBeCloseTo(0, 10)
      expect(result.y).toBeCloseTo(100, 10)
    })

    it('should return bottom-right point for "rb" (default)', () => {
      const result = getRotatedPointCoordinate(style, center, 'rb')
      expect(result.x).toBeCloseTo(100, 10)
      expect(result.y).toBeCloseTo(100, 10)
    })

    it('should default to bottom-right for unknown name', () => {
      const result = getRotatedPointCoordinate(style, center, 'unknown')
      expect(result.x).toBeCloseTo(100, 10)
      expect(result.y).toBeCloseTo(100, 10)
    })

    it('should rotate lt point when style has rotation', () => {
      const rotatedStyle = { left: 0, top: 0, width: 100, height: 100, rotate: 90 }
      const result = getRotatedPointCoordinate(rotatedStyle, center, 'lt')
      // lt=(0,0) rotated 90° around center (50,50) → (100, 0)
      expect(result.x).toBeCloseTo(100, 10)
      expect(result.y).toBeCloseTo(0, 10)
    })
  })

  describe('changeStyleWithScale', () => {
    it('should return floor(value) when scale is 100 (default)', () => {
      expect(changeStyleWithScale(150)).toBe(150)
    })

    it('should scale value by factor when explicit scale provided', () => {
      expect(changeStyleWithScale(100, 200)).toBe(200)
    })

    it('should handle scale as string', () => {
      expect(changeStyleWithScale(100, '200')).toBe(200)
    })

    it('should floor the result', () => {
      // 33 * (150/100) = 49.5 → floor = 49
      expect(changeStyleWithScale(33, 150)).toBe(49)
    })

    it('should return 0 for value 0', () => {
      expect(changeStyleWithScale(0)).toBe(0)
    })
  })
})
