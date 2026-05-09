import {
  positionData,
  multiDimensionalData,
  styleData,
  styleMap,
  textAlignOptions,
  borderStyleOptions,
  verticalAlignOptions,
  selectKey,
  horizontalPosition,
  fieldType,
  fieldTypeText,
  optionMap
} from '@/utils/attr'
import { describe, it, expect } from 'vitest'

describe('positionData', () => {
  it('should have 4 position entries', () => {
    expect(positionData).toHaveLength(4)
  })

  it('should contain left, width, top, height keys', () => {
    const keys = positionData.map(item => item.key)
    expect(keys).toEqual(['left', 'width', 'top', 'height'])
  })

  it('should have correct labels', () => {
    const labels = positionData.map(item => item.label)
    expect(labels).toEqual(['X', 'W', 'Y', 'H'])
  })

  it('should have min/max/step on all entries', () => {
    positionData.forEach(item => {
      expect(item).toHaveProperty('min')
      expect(item).toHaveProperty('max')
      expect(item).toHaveProperty('step')
    })
  })
})

describe('multiDimensionalData', () => {
  it('should have 3 dimension entries', () => {
    expect(multiDimensionalData).toHaveLength(3)
  })

  it('should contain x, y, z keys', () => {
    const keys = multiDimensionalData.map(item => item.key)
    expect(keys).toEqual(['x', 'y', 'z'])
  })
})

describe('styleData', () => {
  it('should have 19 style entries', () => {
    expect(styleData).toHaveLength(19)
  })

  it('should include required style keys', () => {
    const keys = styleData.map(item => item.key)
    expect(keys).toContain('fontSize')
    expect(keys).toContain('color')
    expect(keys).toContain('backgroundColor')
    expect(keys).toContain('opacity')
    expect(keys).toContain('borderRadius')
  })

  it('should have min/max/step for numeric entries', () => {
    const numericEntries = styleData.filter(item => 'min' in item)
    expect(numericEntries.length).toBeGreaterThan(0)
    numericEntries.forEach(item => {
      expect(typeof item.min).toBe('number')
      expect(typeof item.max).toBe('number')
      expect(typeof item.step).toBe('number')
    })
  })
})

describe('styleMap', () => {
  it('should be an object with Chinese labels', () => {
    expect(typeof styleMap).toBe('object')
    expect(styleMap.left).toBe('x 坐标')
    expect(styleMap.color).toBe('颜色')
    expect(styleMap.fontSize).toBe('字体大小')
  })

  it('should have entries for all common style properties', () => {
    expect(Object.keys(styleMap)).toContain('width')
    expect(Object.keys(styleMap)).toContain('height')
    expect(Object.keys(styleMap)).toContain('opacity')
    expect(Object.keys(styleMap)).toContain('lineHeight')
  })
})

describe('textAlignOptions', () => {
  it('should have 3 options', () => {
    expect(textAlignOptions).toHaveLength(3)
  })

  it('should include left, center, right', () => {
    const values = textAlignOptions.map(o => o.value)
    expect(values).toEqual(['left', 'center', 'right'])
  })
})

describe('borderStyleOptions', () => {
  it('should have 2 options', () => {
    expect(borderStyleOptions).toHaveLength(2)
  })

  it('should include solid and dashed', () => {
    const values = borderStyleOptions.map(o => o.value)
    expect(values).toEqual(['solid', 'dashed'])
  })
})

describe('verticalAlignOptions', () => {
  it('should have 3 options', () => {
    expect(verticalAlignOptions).toHaveLength(3)
  })

  it('should include top, middle, bottom', () => {
    const values = verticalAlignOptions.map(o => o.value)
    expect(values).toEqual(['top', 'middle', 'bottom'])
  })
})

describe('selectKey and horizontalPosition', () => {
  it('selectKey should contain textAlign, borderStyle, verticalAlign', () => {
    expect(selectKey).toEqual(['textAlign', 'borderStyle', 'verticalAlign'])
  })

  it('horizontalPosition should contain headHorizontalPosition', () => {
    expect(horizontalPosition).toEqual(['headHorizontalPosition'])
  })
})

describe('fieldType and fieldTypeText', () => {
  it('should have equal length arrays', () => {
    expect(fieldType).toHaveLength(fieldTypeText.length)
  })

  it('fieldType should contain expected types', () => {
    expect(fieldType).toContain('text')
    expect(fieldType).toContain('time')
    expect(fieldType).toContain('value')
    expect(fieldType).toContain('location')
    expect(fieldType).toContain('binary')
    expect(fieldType).toContain('url')
  })
})

describe('optionMap', () => {
  it('should map textAlign to textAlignOptions', () => {
    expect(optionMap.textAlign).toBe(textAlignOptions)
  })

  it('should map borderStyle to borderStyleOptions', () => {
    expect(optionMap.borderStyle).toBe(borderStyleOptions)
  })

  it('should map verticalAlign to verticalAlignOptions', () => {
    expect(optionMap.verticalAlign).toBe(verticalAlignOptions)
  })
})
