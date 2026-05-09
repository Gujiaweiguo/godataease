import { describe, it, expect } from 'vitest'
import { colorStringToHex } from '@/utils/color'

describe('colorStringToHex', () => {
  it('should convert rgb(255,0,0) to #FF0000', () => {
    expect(colorStringToHex('rgb(255,0,0)')).toBe('#FF0000')
  })

  it('should convert rgb(0,128,255) to #0080FF', () => {
    expect(colorStringToHex('rgb(0,128,255)')).toBe('#0080FF')
  })

  it('should convert rgb(0,0,0) to #000000', () => {
    expect(colorStringToHex('rgb(0,0,0)')).toBe('#000000')
  })

  it('should convert rgb(255,255,255) to #FFFFFF', () => {
    expect(colorStringToHex('rgb(255,255,255)')).toBe('#FFFFFF')
  })

  it('should convert rgba(0,128,255,0.5) to #0080FF80', () => {
    expect(colorStringToHex('rgba(0,128,255,0.5)')).toBe('#0080FF80')
  })

  it('should convert rgba(255,0,0,1) to #FF0000FF', () => {
    expect(colorStringToHex('rgba(255,0,0,1)')).toBe('#FF0000FF')
  })

  it('should convert rgba(0,0,0,0) to #00000000', () => {
    expect(colorStringToHex('rgba(0,0,0,0)')).toBe('#00000000')
  })

  it('should handle rgb with spaces', () => {
    expect(colorStringToHex('rgb( 255 , 0 , 0 )')).toBe('#FF0000')
  })

  it('should return null for invalid input string', () => {
    expect(colorStringToHex('not-a-color')).toBeNull()
  })

  it('should return null for empty string', () => {
    expect(colorStringToHex('')).toBeNull()
  })

  it('should return null for hex color input', () => {
    expect(colorStringToHex('#FF0000')).toBeNull()
  })

  it('should return null for partial rgb string', () => {
    expect(colorStringToHex('rgb(255)')).toBeNull()
  })

  it('should handle rgba with decimal alpha like 0.75', () => {
    // 0.75 * 255 = 191.25 -> rounded to 191 -> hex BF
    expect(colorStringToHex('rgba(100,200,50,0.75)')).toBe('#64C832BF')
  })
})
