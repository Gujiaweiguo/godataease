import { describe, expect, it } from 'vitest'

import animationClassData from '@/utils/animationClassData'

describe('animationClassData', () => {
  it('has exactly 3 categories', () => {
    expect(animationClassData).toHaveLength(3)
  })

  it('has correct category labels: 进入, 强调, 退出', () => {
    const labels = animationClassData.map(c => c.label)
    expect(labels).toEqual(['进入', '强调', '退出'])
  })

  it('each category has a non-empty children array', () => {
    animationClassData.forEach(category => {
      expect(category.children.length).toBeGreaterThan(0)
    })
  })

  it('each child has label and value properties', () => {
    animationClassData.forEach(category => {
      category.children.forEach(child => {
        expect(child).toHaveProperty('label')
        expect(child).toHaveProperty('value')
        expect(typeof child.label).toBe('string')
        expect(typeof child.value).toBe('string')
      })
    })
  })

  it('each child has pending set to false', () => {
    animationClassData.forEach(category => {
      category.children.forEach((child: any) => {
        expect(child.pending).toBe(false)
      })
    })
  })

  it('each child has animationTime set to 1', () => {
    animationClassData.forEach(category => {
      category.children.forEach((child: any) => {
        expect(child.animationTime).toBe(1)
      })
    })
  })

  it('进入 category contains fadeIn animation', () => {
    const enterCategory = animationClassData.find(c => c.label === '进入')!
    const fadeIn = enterCategory.children.find(c => c.value === 'fadeIn')
    expect(fadeIn).toBeDefined()
    expect(fadeIn!.label).toBe('渐显')
  })

  it('强调 category contains bounce animation', () => {
    const emphasisCategory = animationClassData.find(c => c.label === '强调')!
    const bounce = emphasisCategory.children.find(c => c.value === 'bounce')
    expect(bounce).toBeDefined()
    expect(bounce!.label).toBe('弹跳')
  })

  it('退出 category contains fadeOut animation', () => {
    const exitCategory = animationClassData.find(c => c.label === '退出')!
    const fadeOut = exitCategory.children.find(c => c.value === 'fadeOut')
    expect(fadeOut).toBeDefined()
    expect(fadeOut!.label).toBe('渐隐')
  })

  it('进入 category has more children than 强调 category', () => {
    const enterCategory = animationClassData.find(c => c.label === '进入')!
    const emphasisCategory = animationClassData.find(c => c.label === '强调')!
    expect(enterCategory.children.length).toBeGreaterThan(emphasisCategory.children.length)
  })

  it('all values are unique across all categories', () => {
    const allValues = animationClassData.flatMap(c => c.children.map(ch => ch.value))
    const uniqueValues = new Set(allValues)
    expect(uniqueValues.size).toBe(allValues.length)
  })
})
