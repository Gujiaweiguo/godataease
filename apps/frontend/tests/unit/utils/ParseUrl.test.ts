import { describe, it, expect } from 'vitest'
import { parseUrl } from '@/utils/ParseUrl'

describe('parseUrl', () => {
  it('should parse a URL with hash path and single query param', () => {
    const result = parseUrl('http://example.com#/dashboard?name=test')
    expect(result.path).toBe('dashboard')
    expect(result.query).toEqual({ name: 'test' })
  })

  it('should parse a URL with hash path and multiple query params', () => {
    const result = parseUrl('http://example.com#/user/list?page=1&size=10')
    expect(result.path).toBe('user/list')
    expect(result.query).toEqual({ page: '1', size: '10' })
  })

  it('should parse a URL with deep hash path', () => {
    const result = parseUrl('http://example.com#/system/user/detail?id=42')
    expect(result.path).toBe('system/user/detail')
    expect(result.query).toEqual({ id: '42' })
  })

  it('should handle query params with empty value', () => {
    const result = parseUrl('http://example.com#/view?keyword=')
    expect(result.path).toBe('view')
    expect(result.query).toEqual({ keyword: '' })
  })

  it('should throw when URL has no query params', () => {
    expect(() => parseUrl('http://example.com#/dashboard')).toThrow()
  })

  it('should parse a minimal hash URL with query', () => {
    const result = parseUrl('#/home?tab=overview')
    expect(result.path).toBe('home')
    expect(result.query).toEqual({ tab: 'overview' })
  })

  it('should split on first equal sign per param', () => {
    const result = parseUrl('#/page?expr=a=b')
    expect(result.query).toEqual({ expr: 'a' })
  })

  it('should return the path after #/', () => {
    const result = parseUrl('http://localhost:8080#/panel/edit?resourceId=abc123')
    expect(result.path).toBe('panel/edit')
    expect(result.query).toEqual({ resourceId: 'abc123' })
  })

  it('should handle URL with port and complex path', () => {
    const result = parseUrl('http://localhost:5173#/data/datasource?type=mysql&name=testdb')
    expect(result.path).toBe('data/datasource')
    expect(result.query).toEqual({ type: 'mysql', name: 'testdb' })
  })
})
