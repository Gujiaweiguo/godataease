import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useMapStore } from '@/store/modules/map'

describe('Map Store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  describe('Initial State', () => {
    it('should have empty mapCache', () => {
      const store = useMapStore()
      expect(store.mapCache).toEqual({})
    })

    it('should have empty mapKey', () => {
      const store = useMapStore()
      expect(store.mapKey).toEqual({ key: '', securityCode: '', mapType: '' })
    })
  })

  describe('setMap', () => {
    it('should store geoJson by id', () => {
      const store = useMapStore()
      const geoJson = { type: 'FeatureCollection', features: [] } as any
      store.setMap({ id: 'china', geoJson })
      expect(store.mapCache['china']).toStrictEqual(geoJson)
    })

    it('should store multiple geoJson entries', () => {
      const store = useMapStore()
      const geoJson1 = { type: 'FeatureCollection', features: [] } as any
      const geoJson2 = { type: 'FeatureCollection', features: [{ type: 'Feature' }] } as any
      store.setMap({ id: 'china', geoJson: geoJson1 })
      store.setMap({ id: 'world', geoJson: geoJson2 })
      expect(store.mapCache['china']).toStrictEqual(geoJson1)
      expect(store.mapCache['world']).toStrictEqual(geoJson2)
    })

    it('should overwrite existing geoJson for same id', () => {
      const store = useMapStore()
      const oldGeoJson = { type: 'FeatureCollection', features: [] } as any
      const newGeoJson = { type: 'FeatureCollection', features: [{ type: 'Feature' }] } as any
      store.setMap({ id: 'china', geoJson: oldGeoJson })
      store.setMap({ id: 'china', geoJson: newGeoJson })
      expect(store.mapCache['china']).toStrictEqual(newGeoJson)
    })
  })

  describe('setKey', () => {
    it('should update mapKey', () => {
      const store = useMapStore()
      store.setKey({ key: 'test-key', securityCode: 'secret', mapType: 'gaode' })
      expect(store.mapKey).toEqual({ key: 'test-key', securityCode: 'secret', mapType: 'gaode' })
    })

    it('should replace entire mapKey', () => {
      const store = useMapStore()
      store.setKey({ key: 'first', securityCode: 'code1', mapType: 'type1' })
      store.setKey({ key: 'second', securityCode: 'code2', mapType: 'type2' })
      expect(store.mapKey).toEqual({ key: 'second', securityCode: 'code2', mapType: 'type2' })
    })

    it('should update mapKey with partial fields', () => {
      const store = useMapStore()
      store.setKey({ key: 'only-key', securityCode: '', mapType: '' })
      expect(store.mapKey.key).toBe('only-key')
      expect(store.mapKey.securityCode).toBe('')
    })
  })
})
