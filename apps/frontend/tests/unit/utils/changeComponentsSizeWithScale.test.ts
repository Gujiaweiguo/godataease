import { describe, it, expect, vi, beforeEach } from 'vitest'

const {
  mockComponentData,
  mockCanvasStyleData,
  mockCurComponentIndex,
  mockSetComponentData,
  mockSetCurComponent,
  mockSetCanvasStyle
} = vi.hoisted(() => {
  const componentData = { value: [] }
  const canvasStyleData: { value: Record<string, any> } = { value: { scale: 100 } }
  const curComponentIndex = { value: -1 }
  const setComponentData = vi.fn((data: any) => {
    componentData.value = data
  })
  return {
    mockComponentData: componentData,
    mockCanvasStyleData: canvasStyleData,
    mockCurComponentIndex: curComponentIndex,
    mockSetComponentData: setComponentData,
    mockSetCurComponent: vi.fn(),
    mockSetCanvasStyle: vi.fn()
  }
})

vi.mock('@/store/modules/data-visualization/dvMain', () => ({
  dvMainStoreWithOut: () => ({
    componentData: mockComponentData,
    canvasStyleData: mockCanvasStyleData,
    curComponentIndex: mockCurComponentIndex,
    setComponentData: mockSetComponentData,
    setCurComponent: mockSetCurComponent,
    setCanvasStyle: mockSetCanvasStyle
  })
}))

vi.mock('pinia', () => ({
  storeToRefs: (store: any) => ({
    componentData: store.componentData,
    curComponentIndex: store.curComponentIndex,
    canvasStyleData: store.canvasStyleData
  })
}))

vi.mock('@/utils/utils', () => ({
  deepCopy: (obj: any) => JSON.parse(JSON.stringify(obj))
}))

vi.mock('@/utils/style', () => ({
  groupItemStyleAdaptor: vi.fn(),
  groupSizeStyleAdaptor: vi.fn()
}))

vi.mock('vue', () => ({
  nextTick: (cb: any) => Promise.resolve().then(cb)
}))

import {
  changeRefComponentsSizeWithScale,
  changeRefComponentsSizeWithScalePointCircle,
  changeRefComponentsSizeWithScalePoint,
  changeComponentSizeWithScale,
  changeSizeWithScale,
  changeSizeWithScaleAdaptor
} from '@/utils/changeComponentsSizeWithScale'

function makeComponent(style: Record<string, any>, component = 'rect') {
  return { component, style: { ...style }, propValue: [] }
}

describe('changeComponentsSizeWithScale', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockComponentData.value = []
    mockCanvasStyleData.value = { scale: 100, scaleHeight: 100, scaleWidth: 100 }
    mockCurComponentIndex.value = -1
  })

  describe('changeRefComponentsSizeWithScale', () => {
    it('should scale width/height/top/left/fontSize by scale ratio', () => {
      const comp = makeComponent({ top: 100, left: 200, width: 300, height: 400, fontSize: 16 })
      const canvasRef = { scale: 100 }

      changeRefComponentsSizeWithScale([comp], canvasRef, 50)

      expect(comp.style.top).toBe(50)
      expect(comp.style.left).toBe(100)
      expect(comp.style.width).toBe(150)
      expect(comp.style.height).toBe(200)
      expect(comp.style.fontSize).toBe(8)
    })

    it('should update canvasRef.scale to new scale', () => {
      const canvasRef = { scale: 100 }
      changeRefComponentsSizeWithScale([], canvasRef, 200)
      expect(canvasRef.scale).toBe(200)
    })

    it('should skip fontSize when it is empty string', () => {
      const comp = makeComponent({ width: 100, fontSize: '' })
      const canvasRef = { scale: 100 }

      changeRefComponentsSizeWithScale([comp], canvasRef, 50)

      expect(comp.style.width).toBe(50)
      expect(comp.style.fontSize).toBe('')
    })

    it('should handle empty componentDataRef array', () => {
      const canvasRef = { scale: 100 }
      expect(() => changeRefComponentsSizeWithScale([], canvasRef, 50)).not.toThrow()
      expect(canvasRef.scale).toBe(50)
    })

    it('should not change attrs not in needToChangeAttrs', () => {
      const comp = makeComponent({ width: 100, color: 'red', opacity: 0.5 })
      const canvasRef = { scale: 100 }

      changeRefComponentsSizeWithScale([comp], canvasRef, 200)

      expect(comp.style.color).toBe('red')
      expect(comp.style.opacity).toBe(0.5)
      expect(comp.style.width).toBe(200)
    })

    it('should scale from previous scale correctly (non-100 origin)', () => {
      const comp = makeComponent({ width: 150 })
      const canvasRef = { scale: 150 }

      changeRefComponentsSizeWithScale([comp], canvasRef, 100)

      // getOriginStyle(150, 150) = 150 / (150/100) = 100
      // format(100, 100) = 100 * (100/100) = 100
      expect(comp.style.width).toBe(100)
    })
  })

  describe('changeRefComponentsSizeWithScalePointCircle', () => {
    it('should scale width attrs by scaleWidth and height attrs by scaleHeight', () => {
      const comp = makeComponent({ left: 100, width: 200, top: 100, height: 200 })
      const canvasRef = { scale: 100, scaleHeight: 100 }

      changeRefComponentsSizeWithScalePointCircle([comp], canvasRef, 50, 75, 100)

      expect(comp.style.left).toBe(50)
      expect(comp.style.width).toBe(100)
      expect(comp.style.top).toBe(75)
      expect(comp.style.height).toBe(150)
    })

    it('should recurse into Group components', () => {
      const inner = makeComponent({ width: 100, height: 100 })
      const group = makeComponent({ width: 200, height: 200 }, 'Group')
      group.propValue = [inner]

      const canvasRef = { scale: 100, scaleHeight: 100 }

      changeRefComponentsSizeWithScalePointCircle([group], canvasRef, 50, 50, 100)

      expect(group.style.width).toBe(100)
      expect(inner.style.width).toBe(50)
    })

    it('should use outScale when canvasRef.scale is falsy', () => {
      const comp = makeComponent({ left: 100 })
      const canvasRef = { scale: undefined as any, scaleHeight: undefined as any }

      changeRefComponentsSizeWithScalePointCircle([comp], canvasRef, 50, 50, 100)

      // getOriginStyle(100, undefined || 100) = 100 / 1 = 100
      // format(100, 50) = 100 * 0.5 = 50
      expect(comp.style.left).toBe(50)
    })
  })

  describe('changeRefComponentsSizeWithScalePoint', () => {
    it('should call PointCircle and update scale refs', () => {
      const comp = makeComponent({ width: 100, top: 100 })
      const canvasRef = { scale: 100, scaleWidth: 100, scaleHeight: 100 }

      changeRefComponentsSizeWithScalePoint([comp], canvasRef, 50, 75, 100)

      expect(canvasRef.scale).toBe(50)
      expect(canvasRef.scaleWidth).toBe(50)
      expect(canvasRef.scaleHeight).toBe(75)
    })

    it('should handle multiple components', () => {
      const comp1 = makeComponent({ width: 200, height: 100 })
      const comp2 = makeComponent({ left: 50, top: 50 })
      const canvasRef = { scale: 100, scaleWidth: 100, scaleHeight: 100 }

      changeRefComponentsSizeWithScalePoint([comp1, comp2], canvasRef, 200, 200, 100)

      expect(comp1.style.width).toBe(400)
      expect(comp2.style.left).toBe(100)
    })
  })

  describe('changeComponentSizeWithScale', () => {
    it('should scale width, height, fontSize of a single component', () => {
      const comp = makeComponent({ width: 200, height: 100, fontSize: 14 })

      changeComponentSizeWithScale(comp, 50)

      expect(comp.style.width).toBe(100)
      expect(comp.style.height).toBe(50)
      expect(comp.style.fontSize).toBe(7)
    })

    it('should skip fontSize when empty string', () => {
      const comp = makeComponent({ width: 200, fontSize: '' })

      changeComponentSizeWithScale(comp, 50)

      expect(comp.style.width).toBe(100)
      expect(comp.style.fontSize).toBe('')
    })

    it('should not change top/left/letterSpacing', () => {
      const comp = makeComponent({ top: 100, left: 200, letterSpacing: 5, width: 100 })

      changeComponentSizeWithScale(comp, 200)

      expect(comp.style.top).toBe(100)
      expect(comp.style.left).toBe(200)
      expect(comp.style.letterSpacing).toBe(5)
      expect(comp.style.width).toBe(200)
    })

    it('should use canvasStyleData.scale as default scale', () => {
      mockCanvasStyleData.value.scale = 200
      const comp = makeComponent({ width: 100 })

      changeComponentSizeWithScale(comp)

      expect(comp.style.width).toBe(200)
    })

    it('should handle zero values gracefully', () => {
      const comp = makeComponent({ width: 0, height: 0 })

      changeComponentSizeWithScale(comp, 50)

      expect(comp.style.width).toBe(0)
      expect(comp.style.height).toBe(0)
    })
  })

  describe('changeSizeWithScale / changeSizeWithScaleAdaptor', () => {
    it('changeSizeWithScale should update store with scaled component data', () => {
      mockComponentData.value = [makeComponent({ width: 100, height: 100 })]
      mockCurComponentIndex.value = 0

      changeSizeWithScale(50)

      expect(mockSetComponentData).toHaveBeenCalled()
      expect(mockSetCanvasStyle).toHaveBeenCalledWith(
        expect.objectContaining({ scale: 50, scaleWidth: 50, scaleHeight: 50 })
      )
    })

    it('changeSizeWithScaleAdaptor should delegate to changeComponentsSizeWithScale', () => {
      mockComponentData.value = [makeComponent({ width: 100 })]
      mockCurComponentIndex.value = 0

      changeSizeWithScaleAdaptor(200)

      expect(mockSetCanvasStyle).toHaveBeenCalledWith(
        expect.objectContaining({ scale: 200, scaleWidth: 200, scaleHeight: 200 })
      )
    })
  })
})
