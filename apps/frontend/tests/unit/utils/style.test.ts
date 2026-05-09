import { describe, it, expect, vi, beforeEach } from 'vitest'

const { mockStore } = vi.hoisted(() => {
  const store = {
    dvInfo: { type: 'dashboard' },
    editMode: 'preview',
    mobileInPc: false,
    getComponentData: () => []
  }
  return { mockStore: store }
})

vi.mock('@/store/modules/data-visualization/dvMain', () => ({
  dvMainStoreWithOut: () => mockStore
}))

vi.mock('@/utils/translate', () => ({
  sin: (r: number) => Math.abs(Math.sin((r * Math.PI) / 180)),
  cos: (r: number) => Math.abs(Math.cos((r * Math.PI) / 180)),
  toPercent: (v: number) => v * 100 + '%'
}))

vi.mock('@/utils/imgUtils', () => ({
  imgUrlTrans: (url: string) => url
}))

vi.mock('@/views/chart/components/js/util', () => ({
  hexColorToRGBA: (_hex: string, alpha: number) => `rgba(0,0,0,${alpha})`
}))

vi.mock('@/utils/canvasUtils', () => ({
  isMainCanvas: (id: string) => id === 'canvas-main',
  isTabCanvas: (id: string) => id?.includes('tab')
}))

import {
  getShapeStyle,
  getShapeItemStyle,
  syncShapeItemStyle,
  getSVGStyle,
  getItemAllStyle,
  getStyle,
  getComponentRotatedStyle,
  getCanvasStyle,
  createGroupStyle,
  groupItemStyleAdaptor,
  groupStyleRevertBatch,
  tabInnerStyleRevert,
  groupStyleRevert,
  groupSizeStyleAdaptor,
  dataVTabComponentAdd
} from '@/utils/style'

describe('style utils', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockStore.dvInfo.type = 'dashboard'
    mockStore.editMode = 'preview'
    mockStore.mobileInPc = false
  })

  describe('getShapeStyle', () => {
    it('should convert style dimensions to px and rotation to transform', () => {
      const result = getShapeStyle({ width: 100, height: 200, top: 10, left: 20, rotate: 45 })
      expect(result).toEqual({
        width: '100px',
        height: '200px',
        top: '10px',
        left: '20px',
        transform: 'rotate(45deg)'
      })
    })

    it('should handle zero rotation', () => {
      const result = getShapeStyle({ width: 50, height: 50, top: 0, left: 0, rotate: 0 }) as any
      expect(result.transform).toBe('rotate(0deg)')
      expect(result.width).toBe('50px')
    })
  })

  describe('getShapeItemStyle', () => {
    it('should calculate dashboard style based on grid position', () => {
      const item = { x: 2, y: 3, sizeX: 4, sizeY: 2, style: {} }
      const result = getShapeItemStyle(item, {
        dvModel: 'dashboard',
        cellWidth: 100,
        cellHeight: 50,
        curGap: 5
      })
      expect(result).toEqual({
        padding: '5px!important',
        width: '400px',
        height: '100px',
        left: '100px',
        top: '100px'
      })
    })

    it('should use pixel style for non-dashboard non-tab dataV items', () => {
      const item = {
        style: { width: 300, height: 200, left: 50, top: 80 },
        canvasId: 'canvas-main'
      }
      const result = getShapeItemStyle(item, {
        dvModel: 'dataV',
        cellWidth: 100,
        cellHeight: 50,
        curGap: 8
      })
      expect(result).toEqual({
        padding: '8px!important',
        width: '300px',
        height: '200px',
        left: '50px',
        top: '80px'
      })
    })

    it('should use percentage style for dataV tab canvas in preview mode', () => {
      const item = {
        style: { width: 300, height: 200, left: 50, top: 80 },
        canvasId: 'tab-001',
        groupStyle: { width: 0.5, height: 0.3, top: 0.1, left: 0.2 }
      }
      const result = getShapeItemStyle(item, {
        dvModel: 'dataV',
        cellWidth: 100,
        cellHeight: 50,
        curGap: 4,
        showPosition: 'preview'
      }) as any
      expect(result.width).toBe('50%')
      expect(result.height).toBe('30%')
      expect(result.top).toBe('10%')
      expect(result.left).toBe('20%')
    })
  })

  describe('syncShapeItemStyle', () => {
    it('should sync item style from grid position', () => {
      const item = { x: 2, y: 3, sizeX: 4, sizeY: 2, style: {} } as any
      syncShapeItemStyle(item, 100, 50)
      expect(item.style).toEqual({ left: 100, top: 100, width: 400, height: 100 })
    })
  })

  describe('getSVGStyle', () => {
    it('should build SVG style with px units for dimension keys', () => {
      const result = getSVGStyle({ fontSize: 14, width: 200, height: 100, rotate: 30 }) as any
      expect(result.fontSize).toBe('14px')
      expect(result.width).toBe('200px')
      expect(result.height).toBe('100px')
      expect(result.transform).toBe('rotate(30deg)')
    })

    it('should filter out specified keys', () => {
      const result = getSVGStyle({ opacity: 0.5, width: 200, color: 'red' }, ['width', 'color']) as any
      expect(result.opacity).toBe(0.5)
      expect(result.width).toBeUndefined()
      expect(result.color).toBeUndefined()
    })

    it('should skip keys with empty string values', () => {
      const result = getSVGStyle({ fontSize: '', width: 100 }) as any
      expect(result.fontSize).toBeUndefined()
      expect(result.width).toBe('100px')
    })
  })

  describe('getStyle', () => {
    it('should return style with px units for unit-required keys', () => {
      const result = getStyle({ width: 100, top: 50, borderRadius: 8 }) as any
      expect(result.width).toBe('100px')
      expect(result.top).toBe('50px')
      expect(result.borderRadius).toBe('8px')
    })

    it('should exclude filtered keys', () => {
      const result = getStyle({ width: 100, top: 50, opacity: 0.5 }, ['top']) as any
      expect(result.top).toBeUndefined()
      expect(result.width).toBe('100px')
      expect(result.opacity).toBe(0.5)
    })

    it('should handle rotation as transform', () => {
      const result = getStyle({ rotate: 90 }) as any
      expect(result.transform).toBe('rotate(90deg)')
      expect(result.rotate).toBeUndefined()
    })
  })

  describe('getComponentRotatedStyle', () => {
    it('should return unchanged bounds for zero rotation', () => {
      const result = getComponentRotatedStyle({
        width: 100,
        height: 50,
        top: 10,
        left: 20,
        rotate: 0
      })
      expect(result.bottom).toBe(60)
      expect(result.right).toBe(120)
      expect(result.width).toBe(100)
      expect(result.height).toBe(50)
    })

    it('should calculate rotated bounding box', () => {
      const result = getComponentRotatedStyle({
        width: 100,
        height: 50,
        top: 200,
        left: 100,
        rotate: 45
      })
      expect(result.width).toBeCloseTo(106.07, 0)
      expect(result.height).toBeCloseTo(106.07, 0)
      expect(result.right).toBeDefined()
      expect(result.bottom).toBeDefined()
    })
  })

  describe('getCanvasStyle', () => {
    it('should return default dashboard background for main canvas', () => {
      const result = getCanvasStyle(
        {
          backgroundColorSelect: false,
          background: '',
          backgroundColor: '',
          backgroundImageEnable: false,
          fontSize: 14,
          color: '#333',
          fontFamily: 'Arial'
        },
        'canvas-main'
      ) as any
      expect(result['background-color']).toBe('#f5f6f7')
      expect(result.fontSize).toBe('14px')
      expect(result.color).toBe('#333')
      expect(result['font-family']).toBe('Arial!important')
    })

    it('should return dark background for dataV type', () => {
      mockStore.dvInfo.type = 'dataV'
      const result = getCanvasStyle(
        {
          backgroundColorSelect: false,
          background: '',
          backgroundColor: '',
          backgroundImageEnable: false,
          fontSize: 12,
          color: '#fff',
          fontFamily: 'sans-serif'
        },
        'canvas-main'
      ) as any
      expect(result['background-color']).toBe('#1a1a1a')
    })

    it('should not set background for non-main canvas', () => {
      const result = getCanvasStyle(
        {
          backgroundColorSelect: true,
          background: 'test.png',
          backgroundColor: '#ff0000',
          backgroundImageEnable: true,
          fontSize: 14,
          color: '#333',
          fontFamily: 'Arial'
        },
        'tab-canvas-001'
      ) as any
      expect(result['background-color']).toBeUndefined()
      expect(result['background']).toBeUndefined()
    })

    it('should apply selected background color when enabled', () => {
      const result = getCanvasStyle(
        {
          backgroundColorSelect: true,
          background: '',
          backgroundColor: '#abcdef',
          backgroundImageEnable: false,
          fontSize: 14,
          color: '#333',
          fontFamily: 'Arial'
        },
        'canvas-main'
      ) as any
      expect(result['background-color']).toBe('#abcdef')
    })

    it('should apply background image when enabled', () => {
      const result = getCanvasStyle(
        {
          backgroundColorSelect: true,
          background: 'https://example.com/bg.png',
          backgroundColor: '#ffffff',
          backgroundImageEnable: true,
          fontSize: 14,
          color: '#333',
          fontFamily: 'Arial'
        },
        'canvas-main'
      ) as any
      expect(result['background']).toBe('url(https://example.com/bg.png) no-repeat #ffffff')
    })

    it('should apply mobile settings when mobileInPc and customSetting', () => {
      mockStore.mobileInPc = true
      const result = getCanvasStyle(
        {
          backgroundColorSelect: false,
          background: '',
          backgroundColor: '',
          backgroundImageEnable: false,
          fontSize: 14,
          color: '#333',
          fontFamily: 'Arial',
          mobileSetting: {
            customSetting: true,
            backgroundColorSelect: true,
            color: '#000000',
            backgroundImageEnable: false,
            background: ''
          }
        },
        'canvas-main'
      ) as any
      expect(result['background-color']).toBe('#000000')
    })
  })

  describe('createGroupStyle', () => {
    it('should calculate proportional groupStyle and offset component positions', () => {
      const parentStyle = { left: 100, top: 50, width: 200, height: 150 }
      const component1 = {
        style: { left: 120, top: 70, width: 80, height: 60 },
        groupStyle: {} as any
      }
      const groupComponent = { style: parentStyle, propValue: [component1] }

      createGroupStyle(groupComponent as any)

      expect(component1.groupStyle.left).toBeCloseTo(0.1)
      expect(component1.groupStyle.top).toBeCloseTo(0.133, 2)
      expect(component1.groupStyle.width).toBe(0.4)
      expect(component1.groupStyle.height).toBe(0.4)
      expect(component1.style.left).toBe(20)
      expect(component1.style.top).toBe(20)
    })
  })

  describe('groupItemStyleAdaptor', () => {
    it('should scale component style from groupStyle proportions', () => {
      const component = {
        groupStyle: { left: 0.1, top: 0.2, width: 0.5, height: 0.3 },
        style: {} as any
      }
      groupItemStyleAdaptor(component as any, { width: 400, height: 300 })
      expect(component.style.left).toBe(40)
      expect(component.style.top).toBe(60)
      expect(component.style.width).toBe(200)
      expect(component.style.height).toBe(90)
    })
  })

  describe('groupStyleRevert', () => {
    it('should calculate proportional groupStyle from absolute positions', () => {
      const component = {
        style: { left: 40, top: 60, width: 200, height: 90 },
        groupStyle: {} as any
      }
      groupStyleRevert(component as any, { width: 400, height: 300 })
      expect(component.groupStyle.left).toBe(0.1)
      expect(component.groupStyle.top).toBe(0.2)
      expect(component.groupStyle.width).toBe(0.5)
      expect(component.groupStyle.height).toBe(0.3)
    })
  })

  describe('groupStyleRevertBatch', () => {
    it('should revert styles for DeTabs children', () => {
      const tabComponent = {
        component: 'DeTabs',
        propValue: [
          {
            componentData: [
              {
                style: { left: 20, top: 30, width: 100, height: 50 },
                groupStyle: {} as any
              }
            ]
          }
        ]
      }
      const parentStyle = { width: 200, height: 100 }
      groupStyleRevertBatch(tabComponent as any, parentStyle)
      expect(tabComponent.propValue[0].componentData[0].groupStyle.left).toBe(0.1)
      expect(tabComponent.propValue[0].componentData[0].groupStyle.top).toBe(0.3)
    })

    it('should not process non-DeTabs components', () => {
      const groupComponent = {
        component: 'Group',
        propValue: []
      }
      expect(() => groupStyleRevertBatch(groupComponent as any, {})).not.toThrow()
    })
  })

  describe('tabInnerStyleRevert', () => {
    it('should revert tab inner component styles with title offset', () => {
      const component = {
        style: { left: 0, top: 0, width: 200, height: 146, showTabTitle: true }
      }
      const innerComponent = {
        style: { left: 20, top: 30, width: 100, height: 50 },
        groupStyle: {} as any
      }
      const tabOuter = {
        ...component,
        propValue: [{ componentData: [innerComponent] }]
      }

      tabInnerStyleRevert(tabOuter as any)

      expect(innerComponent.groupStyle.left).toBe(0.1)
      expect(innerComponent.groupStyle.width).toBe(0.5)
      // 146 - 46(showTabTitle) = 100
      expect(innerComponent.groupStyle.top).toBe(0.3)
      expect(innerComponent.groupStyle.height).toBe(0.5)
    })

    it('should not offset height when showTabTitle is false', () => {
      const innerComponent = {
        style: { left: 20, top: 30, width: 100, height: 50 },
        groupStyle: {} as any
      }
      const tabOuter = {
        style: { left: 0, top: 0, width: 200, height: 100, showTabTitle: false },
        propValue: [{ componentData: [innerComponent] }]
      }

      tabInnerStyleRevert(tabOuter as any)

      expect(innerComponent.groupStyle.height).toBe(0.5)
    })
  })

  describe('groupSizeStyleAdaptor', () => {
    it('should adapt Group children using groupItemStyleAdaptor', () => {
      const child = {
        groupStyle: { left: 0.5, top: 0.5, width: 0.5, height: 0.5 },
        style: {} as any
      }
      const group = {
        component: 'Group',
        style: { width: 200, height: 200 },
        propValue: [child]
      }

      groupSizeStyleAdaptor(group as any)

      expect(child.style.left).toBe(100)
      expect(child.style.top).toBe(100)
      expect(child.style.width).toBe(100)
      expect(child.style.height).toBe(100)
    })
  })

  describe('dataVTabComponentAdd', () => {
    it('should reset position and calculate groupStyle for tab component', () => {
      const inner = {
        style: { top: 50, left: 30, width: 100, height: 80 },
        groupStyle: {} as any
      }
      const parent = {
        style: { width: 200, height: 200 },
        showTabTitle: true
      }

      dataVTabComponentAdd(inner as any, parent as any)

      expect(inner.style.top).toBe(0)
      expect(inner.style.left).toBe(0)
      // width / parentWidth(200) = 0.5, height / reducedHeight(154)
      expect(inner.groupStyle.width).toBe(0.5)
      expect(inner.groupStyle.height).toBeCloseTo(80 / 154)
    })

    it('should not offset when showTabTitle is false', () => {
      const inner = {
        style: { top: 50, left: 30, width: 100, height: 80 },
        groupStyle: {} as any
      }
      const parent = {
        style: { width: 200, height: 200 },
        showTabTitle: false
      }

      dataVTabComponentAdd(inner as any, parent as any)

      expect(inner.groupStyle.width).toBe(0.5)
      expect(inner.groupStyle.height).toBe(0.4)
    })
  })

  describe('getItemAllStyle', () => {
    it('should include background color when backgroundColorSelect is enabled', () => {
      const item = {
        style: { width: 100 },
        commonBackground: {
          backgroundColorSelect: true,
          backgroundColor: '#ff0000',
          alpha: 0.8,
          backgroundImageEnable: false
        }
      }
      const result = getItemAllStyle(item as any) as any
      expect(result.width).toBe('100px')
      expect(result['background-color']).toBe('rgba(0,0,0,0.8)')
    })

    it('should include background image when backgroundImageEnable and outerImage type', () => {
      const item = {
        style: { height: 200 },
        commonBackground: {
          backgroundColorSelect: true,
          backgroundColor: '#00ff00',
          alpha: 1,
          backgroundImageEnable: true,
          backgroundType: 'outerImage',
          outerImage: 'https://example.com/img.png'
        }
      }
      const result = getItemAllStyle(item as any) as any
      expect(result.background).toContain('url(https://example.com/img.png)')
    })

    it('should use background-color when image enabled but not outerImage type', () => {
      const item = {
        style: { top: 10 },
        commonBackground: {
          backgroundColorSelect: true,
          backgroundColor: '#0000ff',
          alpha: 0.5,
          backgroundImageEnable: true,
          backgroundType: 'innerImage'
        }
      }
      const result = getItemAllStyle(item as any) as any
      expect(result['background-color']).toBe('rgba(0,0,0,0.5)')
    })
  })
})
