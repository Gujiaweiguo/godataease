import { describe, it, expect, vi, beforeEach, afterAll } from 'vitest'

describe('events', () => {
  let events: typeof import('@/utils/events')['events']
  let eventList: typeof import('@/utils/events')['eventList']
  let mixins: typeof import('@/utils/events')['mixins']

  const originalLocation = window.location
  const originalAlert = window.alert

  beforeEach(async () => {
    window.alert = vi.fn()
    delete (window as any).location
    ;(window as any).location = { href: '' }

    const mod = await import('@/utils/events')
    events = mod.events
    eventList = mod.eventList
    mixins = mod.mixins
  })

  afterAll(() => {
    window.alert = originalAlert
    ;(window as any).location = originalLocation
  })

  describe('events.redirect', () => {
    it('should set window.location.href to the given url', () => {
      events.redirect('https://example.com')
      expect(window.location.href).toBe('https://example.com')
    })

    it('should not change window.location.href when url is empty string', () => {
      window.location.href = 'original'
      events.redirect('')
      expect(window.location.href).toBe('original')
    })

    it('should not change window.location.href when url is null', () => {
      window.location.href = 'original'
      events.redirect(null as any)
      expect(window.location.href).toBe('original')
    })

    it('should not change window.location.href when url is undefined', () => {
      window.location.href = 'original'
      events.redirect(undefined as any)
      expect(window.location.href).toBe('original')
    })
  })

  describe('events.alert', () => {
    it('should call window.alert with the given message', () => {
      events.alert('hello world')
      expect(window.alert).toHaveBeenCalledWith('hello world')
    })

    it('should not call alert when message is empty string', () => {
      events.alert('')
      expect(window.alert).not.toHaveBeenCalled()
    })

    it('should not call alert when message is null', () => {
      events.alert(null as any)
      expect(window.alert).not.toHaveBeenCalled()
    })
  })

  describe('eventList', () => {
    it('should be an array with 2 items', () => {
      expect(Array.isArray(eventList)).toBe(true)
      expect(eventList).toHaveLength(2)
    })

    it('should contain redirect event with correct structure', () => {
      const redirect = eventList.find(e => e.key === 'redirect')
      expect(redirect).toBeDefined()
      expect(redirect!.label).toBe('跳转事件')
      expect(redirect!.param).toBe('')
      expect(typeof redirect!.event).toBe('function')
    })

    it('should contain alert event with correct structure', () => {
      const alertItem = eventList.find(e => e.key === 'alert')
      expect(alertItem).toBeDefined()
      expect(alertItem!.label).toBe('alert 事件')
      expect(alertItem!.param).toBe('')
      expect(typeof alertItem!.event).toBe('function')
    })
  })

  describe('mixins', () => {
    it('should have methods property equal to events', () => {
      expect(mixins.methods).toBe(events)
    })

    it('should expose redirect and alert as methods', () => {
      expect(typeof mixins.methods.redirect).toBe('function')
      expect(typeof mixins.methods.alert).toBe('function')
    })
  })
})
