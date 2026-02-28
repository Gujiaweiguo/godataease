import { describe, it, expect } from 'vitest'
import { isExternal, validUsername, PHONE_REGEX, EMAIL_REGEX } from '@/utils/validate'

describe('Validate Utils', () => {
  describe('isExternal', () => {
    it('should return true for https URLs', () => {
      expect(isExternal('https://example.com')).toBe(true)
    })

    it('should return true for http URLs', () => {
      expect(isExternal('http://example.com')).toBe(true)
    })

    it('should return true for mailto URLs', () => {
      expect(isExternal('mailto:test@example.com')).toBe(true)
    })

    it('should return true for tel URLs', () => {
      expect(isExternal('tel:1234567890')).toBe(true)
    })

    it('should return true for plugin staticInfo path', () => {
      expect(isExternal('/api/pluginCommon/staticInfo')).toBe(true)
    })

    it('should return false for internal paths', () => {
      expect(isExternal('/dashboard')).toBe(false)
    })

    it('should return false for relative paths', () => {
      expect(isExternal('./components/Test.vue')).toBe(false)
    })
  })

  describe('validUsername', () => {
    it('should return true for admin', () => {
      expect(validUsername('admin')).toBe(true)
    })

    it('should return true for cyw', () => {
      expect(validUsername('cyw')).toBe(true)
    })

    it('should return true for admin with whitespace', () => {
      expect(validUsername('  admin  ')).toBe(true)
    })

    it('should return false for invalid username', () => {
      expect(validUsername('invalid')).toBe(false)
    })

    it('should return false for empty string', () => {
      expect(validUsername('')).toBe(false)
    })
  })

  describe('PHONE_REGEX', () => {
    it('should match valid Chinese phone numbers', () => {
      expect(new RegExp(PHONE_REGEX).test('13800138000')).toBe(true)
      expect(new RegExp(PHONE_REGEX).test('15012345678')).toBe(true)
      expect(new RegExp(PHONE_REGEX).test('18812345678')).toBe(true)
    })

    it('should not match invalid phone numbers', () => {
      expect(new RegExp(PHONE_REGEX).test('12800138000')).toBe(false) // starts with 2
      expect(new RegExp(PHONE_REGEX).test('03800138000')).toBe(false) // starts with 3
      expect(new RegExp(PHONE_REGEX).test('1380013800')).toBe(false) // too short
    })
  })

  describe('EMAIL_REGEX', () => {
    it('should match valid email addresses', () => {
      const regex = new RegExp(EMAIL_REGEX)
      expect(regex.test('test@example.com')).toBe(true)
      expect(regex.test('user.name@example.com')).toBe(true)
      expect(regex.test('user_name@example.co.uk')).toBe(true)
      expect(regex.test('user-name@example.org')).toBe(true)
    })

    it('should not match invalid email addresses', () => {
      const regex = new RegExp(EMAIL_REGEX)
      expect(regex.test('invalid')).toBe(false)
      expect(regex.test('invalid@')).toBe(false)
      expect(regex.test('@example.com')).toBe(false)
    })
  })
})
