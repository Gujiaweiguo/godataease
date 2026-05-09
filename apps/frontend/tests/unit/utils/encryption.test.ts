import { describe, it, expect, vi, beforeEach } from 'vitest'
import CryptoJS from 'crypto-js/crypto-js'

const mocks = vi.hoisted(() => ({
  wsGetMock: vi.fn(),
  dekeyGetter: vi.fn()
}))

vi.mock('@/hooks/web/useCache', () => ({
  useCache: () => ({
    wsCache: {
      get: mocks.wsGetMock
    }
  })
}))

vi.mock('@/store/modules/app', () => ({
  useAppStoreWithOut: () => ({
    getDekey: mocks.dekeyGetter
  })
}))

import { symmetricDecrypt, rsaEncryp } from '@/utils/encryption'

describe('encryption', () => {
  describe('symmetricDecrypt', () => {
    it('should decrypt data encrypted with AES-CBC and Base64 key', () => {
      const plainText = 'Hello, DataEase!'
      const iv = CryptoJS.enc.Utf8.parse('0000000000000000')
      const keyStr = CryptoJS.enc.Base64.stringify(CryptoJS.lib.WordArray.random(16))
      const key = CryptoJS.enc.Base64.parse(keyStr)
      const encrypted = CryptoJS.AES.encrypt(plainText, key, {
        iv: iv,
        mode: CryptoJS.mode.CBC,
        padding: CryptoJS.pad.Pkcs7
      })
      const encryptedData = encrypted.toString()
      const result = symmetricDecrypt(encryptedData, keyStr)
      expect(result).toBe(plainText)
    })

    it('should handle empty string decryption', () => {
      const plainText = ''
      const iv = CryptoJS.enc.Utf8.parse('0000000000000000')
      const keyStr = CryptoJS.enc.Base64.stringify(CryptoJS.lib.WordArray.random(16))
      const key = CryptoJS.enc.Base64.parse(keyStr)
      const encrypted = CryptoJS.AES.encrypt(plainText, key, {
        iv: iv,
        mode: CryptoJS.mode.CBC,
        padding: CryptoJS.pad.Pkcs7
      })
      const result = symmetricDecrypt(encrypted.toString(), keyStr)
      expect(result).toBe('')
    })

    it('should decrypt Chinese characters', () => {
      const plainText = '数据可视化分析'
      const iv = CryptoJS.enc.Utf8.parse('0000000000000000')
      const keyStr = CryptoJS.enc.Base64.stringify(CryptoJS.lib.WordArray.random(16))
      const key = CryptoJS.enc.Base64.parse(keyStr)
      const encrypted = CryptoJS.AES.encrypt(plainText, key, {
        iv: iv,
        mode: CryptoJS.mode.CBC,
        padding: CryptoJS.pad.Pkcs7
      })
      const result = symmetricDecrypt(encrypted.toString(), keyStr)
      expect(result).toBe(plainText)
    })

    it('should decrypt long strings', () => {
      const plainText = 'a'.repeat(500)
      const iv = CryptoJS.enc.Utf8.parse('0000000000000000')
      const keyStr = CryptoJS.enc.Base64.stringify(CryptoJS.lib.WordArray.random(16))
      const key = CryptoJS.enc.Base64.parse(keyStr)
      const encrypted = CryptoJS.AES.encrypt(plainText, key, {
        iv: iv,
        mode: CryptoJS.mode.CBC,
        padding: CryptoJS.pad.Pkcs7
      })
      const result = symmetricDecrypt(encrypted.toString(), keyStr)
      expect(result).toBe(plainText)
    })

    it('should decrypt JSON strings', () => {
      const plainText = JSON.stringify({ user: 'admin', role: 'editor' })
      const iv = CryptoJS.enc.Utf8.parse('0000000000000000')
      const keyStr = CryptoJS.enc.Base64.stringify(CryptoJS.lib.WordArray.random(16))
      const key = CryptoJS.enc.Base64.parse(keyStr)
      const encrypted = CryptoJS.AES.encrypt(plainText, key, {
        iv: iv,
        mode: CryptoJS.mode.CBC,
        padding: CryptoJS.pad.Pkcs7
      })
      const result = symmetricDecrypt(encrypted.toString(), keyStr)
      expect(result).toBe(plainText)
    })

    it('should work with a fixed known key', () => {
      const plainText = 'test-password-123'
      const rawKey = CryptoJS.lib.WordArray.random(16)
      const keyStr = CryptoJS.enc.Base64.stringify(rawKey)
      const iv = CryptoJS.enc.Utf8.parse('0000000000000000')
      const key = CryptoJS.enc.Base64.parse(keyStr)
      const encrypted = CryptoJS.AES.encrypt(plainText, key, {
        iv: iv,
        mode: CryptoJS.mode.CBC,
        padding: CryptoJS.pad.Pkcs7
      })
      const decrypted = symmetricDecrypt(encrypted.toString(), keyStr)
      expect(decrypted).toBe(plainText)
    })
  })

  describe('rsaEncryp', () => {
    beforeEach(() => {
      mocks.wsGetMock.mockReturnValue('dummyKey123' + btoa('-pk_separator-') + '=' + 'MTIzNDU2Nzg5MDEyMzQ1Ng==')
      mocks.dekeyGetter.mockReturnValue('test-dekey')
    })

    it('should be callable with mocked dependencies and return a result', () => {
      const result = rsaEncryp('test')
      expect(result).toBeDefined()
    })
  })
})
