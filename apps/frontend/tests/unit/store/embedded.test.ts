import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useEmbedded } from '@/store/modules/embedded'

describe('Embedded Store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  describe('Initial State', () => {
    it('should have correct initial values', () => {
      const store = useEmbedded()
      expect(store.getType).toBe('')
      expect(store.getToken).toBe('')
      expect(store.getBusiFlag).toBe('')
      expect(store.getOuterParams).toBe('')
      expect(store.getSuffixId).toBe('')
      expect(store.getBaseUrl).toBe('')
      expect(store.getDvId).toBe('')
      expect(store.getPid).toBe('')
      expect(store.getChartId).toBe('')
      expect(store.getResourceId).toBe('')
      expect(store.getDfId).toBe('')
      expect(store.getOpt).toBe('')
      expect(store.getCreateType).toBe('')
      expect(store.getTemplateParams).toBe('')
      expect(store.getOuterUrl).toBe('')
      expect(store.getJumpInfoParam).toBe('')
      expect(store.getEmbedReady).toBe(false)
      expect(store.getAllowedOrigins).toEqual([])
      expect(store.getParams).toEqual({})
    })

    it('should have empty tokenInfo Map', () => {
      const store = useEmbedded()
      expect(store.getTokenInfo).toBeInstanceOf(Map)
      expect(store.getTokenInfo.size).toBe(0)
    })
  })

  describe('setType', () => {
    it('should set type', () => {
      const store = useEmbedded()
      store.setType('dashboard')
      expect(store.getType).toBe('dashboard')
    })
  })

  describe('setToken', () => {
    it('should set token', () => {
      const store = useEmbedded()
      store.setToken('test-token-123')
      expect(store.getToken).toBe('test-token-123')
    })
  })

  describe('setBusiFlag', () => {
    it('should set busiFlag', () => {
      const store = useEmbedded()
      store.setBusiFlag('dashboard')
      expect(store.getBusiFlag).toBe('dashboard')
    })
  })

  describe('setOuterParams', () => {
    it('should set outerParams', () => {
      const store = useEmbedded()
      store.setOuterParams('{"key":"value"}')
      expect(store.getOuterParams).toBe('{"key":"value"}')
    })
  })

  describe('setSuffixId', () => {
    it('should set suffixId', () => {
      const store = useEmbedded()
      store.setSuffixId('suffix-123')
      expect(store.getSuffixId).toBe('suffix-123')
    })
  })

  describe('setBaseUrl', () => {
    it('should set baseUrl', () => {
      const store = useEmbedded()
      store.setBaseUrl('https://example.com')
      expect(store.getBaseUrl).toBe('https://example.com')
    })
  })

  describe('setDvId', () => {
    it('should set dvId', () => {
      const store = useEmbedded()
      store.setDvId('dv-123')
      expect(store.getDvId).toBe('dv-123')
    })
  })

  describe('setPid', () => {
    it('should set pid', () => {
      const store = useEmbedded()
      store.setPid('pid-123')
      expect(store.getPid).toBe('pid-123')
    })
  })

  describe('setChartId', () => {
    it('should set chartId', () => {
      const store = useEmbedded()
      store.setChartId('chart-123')
      expect(store.getChartId).toBe('chart-123')
    })
  })

  describe('setResourceId', () => {
    it('should set resourceId', () => {
      const store = useEmbedded()
      store.setResourceId('resource-123')
      expect(store.getResourceId).toBe('resource-123')
    })
  })

  describe('setDfId', () => {
    it('should set dfId', () => {
      const store = useEmbedded()
      store.setDfId('df-123')
      expect(store.getDfId).toBe('df-123')
    })
  })

  describe('setCreateType', () => {
    it('should set createType', () => {
      const store = useEmbedded()
      store.setCreateType('template')
      expect(store.getCreateType).toBe('template')
    })
  })

  describe('setTemplateParams', () => {
    it('should set templateParams', () => {
      const store = useEmbedded()
      store.setTemplateParams('{"template":"data"}')
      expect(store.getTemplateParams).toBe('{"template":"data"}')
    })
  })

  describe('setOuterUrl', () => {
    it('should set outerUrl', () => {
      const store = useEmbedded()
      store.setOuterUrl('https://outer.example.com')
      expect(store.getOuterUrl).toBe('https://outer.example.com')
    })
  })

  describe('setJumpInfoParam', () => {
    it('should set jumpInfoParam', () => {
      const store = useEmbedded()
      store.setJumpInfoParam('jump-info')
      expect(store.getJumpInfoParam).toBe('jump-info')
    })
  })

  describe('setAllowedOrigins', () => {
    it('should set allowedOrigins', () => {
      const store = useEmbedded()
      store.setAllowedOrigins(['https://example.com', 'https://test.com'])
      expect(store.getAllowedOrigins).toEqual(['https://example.com', 'https://test.com'])
    })
  })

  describe('setDatasetId', () => {
    it('should set datasetId', () => {
      const store = useEmbedded()
      store.setDatasetId('dataset-123')
      expect(store.datasetId).toBe('dataset-123')
    })
  })

  describe('setDatasourceId', () => {
    it('should set datasourceId', () => {
      const store = useEmbedded()
      store.setDatasourceId('ds-123')
      expect(store.datasourceId).toBe('ds-123')
    })
  })

  describe('setTableName', () => {
    it('should set tableName', () => {
      const store = useEmbedded()
      store.setTableName('users')
      expect(store.tableName).toBe('users')
    })
  })

  describe('getIframeData', () => {
    it('should return all iframe data', () => {
      const store = useEmbedded()
      store.setToken('token-123')
      store.setBusiFlag('dashboard')
      store.setDvId('dv-456')
      store.setChartId('chart-789')

      const iframeData = store.getIframeData
      expect(iframeData.embeddedToken).toBe('token-123')
      expect(iframeData.busiFlag).toBe('dashboard')
      expect(iframeData.dvId).toBe('dv-456')
      expect(iframeData.chartId).toBe('chart-789')
    })
  })

  describe('Multiple setters', () => {
    it('should set multiple values correctly', () => {
      const store = useEmbedded()
      store.setType('panel')
      store.setToken('multi-token')
      store.setBusiFlag('dataV')
      store.setDvId('multi-dv')
      store.setPid('multi-pid')

      expect(store.getType).toBe('panel')
      expect(store.getToken).toBe('multi-token')
      expect(store.getBusiFlag).toBe('dataV')
      expect(store.getDvId).toBe('multi-dv')
      expect(store.getPid).toBe('multi-pid')
    })
  })
})
