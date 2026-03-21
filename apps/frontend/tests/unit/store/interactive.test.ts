import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { interactiveStore } from '@/store/modules/interactive'
import { queryTreeApi, queryBusiTreeApi } from '@/api/visualization/dataVisualization'
import { getDatasetTree } from '@/api/dataset'
import { listDatasources } from '@/api/datasource'

// Mock API dependencies
vi.mock('@/api/visualization/dataVisualization', async () => {
  const { createResolvedApiModuleMock } = await import('../helpers')
  return createResolvedApiModuleMock({
    queryTreeApi: { data: [] },
    queryBusiTreeApi: {}
  })
})

vi.mock('@/api/dataset', async () => {
  const { createResolvedApiModuleMock } = await import('../helpers')
  return createResolvedApiModuleMock({ getDatasetTree: { data: [] } })
})

vi.mock('@/api/datasource', async () => {
  const { createResolvedApiModuleMock } = await import('../helpers')
  return createResolvedApiModuleMock({ listDatasources: { data: { list: [] } } })
})

vi.mock('@/hooks/web/useCache', async () => {
  const { createUseCacheModuleMock } = await import('../helpers')
  return createUseCacheModuleMock()
})

vi.mock('@/store/modules/app', async () => {
  const { createAppStoreWithOutModuleMock } = await import('../helpers')
  return createAppStoreWithOutModuleMock({ getIsIframe: false })
})

vi.mock('@/store/modules/permission', async () => {
  const { createPermissionModuleMock } = await import('../helpers')
  return createPermissionModuleMock()
})

describe('Interactive Store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  describe('Initial State', () => {
    it('should have empty data initially', () => {
      const store = interactiveStore()
      expect(store.getData).toEqual({})
    })

    it('should have undefined panel initially', () => {
      const store = interactiveStore()
      expect(store.getPanel).toBeUndefined()
    })

    it('should have undefined screen initially', () => {
      const store = interactiveStore()
      expect(store.getScreen).toBeUndefined()
    })

    it('should have undefined dataset initially', () => {
      const store = interactiveStore()
      expect(store.getDataset).toBeUndefined()
    })

    it('should have undefined datasource initially', () => {
      const store = interactiveStore()
      expect(store.getDatasource).toBeUndefined()
    })
  })

  describe('setInteractive', () => {
    it('should set interactive data for panel', async () => {
      const store = interactiveStore()
      const mockData = [
        { id: '1', pid: '0', name: 'Test Panel', weight: 7, leaf: true, extraFlag: 0, extraFlag1: 0 }
      ]

      await store.setInteractive({ busiFlag: 'dashboard' }, mockData)

      expect(store.getPanel).toBeDefined()
      expect(store.getPanel.rootManage).toBe(true)
    })

    it('should calculate leafNodeCount correctly', async () => {
      const store = interactiveStore()
      const mockData = [
        { id: '1', pid: '0', name: 'Folder', weight: 7, leaf: false, extraFlag: 0, extraFlag1: 0, children: [
          { id: '2', pid: '1', name: 'Leaf 1', weight: 7, leaf: true, extraFlag: 0, extraFlag1: 0 },
          { id: '3', pid: '1', name: 'Leaf 2', weight: 5, leaf: true, extraFlag: 0, extraFlag1: 0 }
        ]}
      ]

      await store.setInteractive({ busiFlag: 'dashboard' }, mockData)

      expect(store.getPanel.leafNodeCount).toBe(2)
    })

    it('should detect anyManage correctly', async () => {
      const store = interactiveStore()
      const mockData = [
        { id: '1', pid: '0', name: 'Folder', weight: 7, leaf: false, extraFlag: 0, extraFlag1: 0, children: [
          { id: '2', pid: '1', name: 'Leaf', weight: 3, leaf: true, extraFlag: 0, extraFlag1: 0 }
        ]}
      ]

      await store.setInteractive({ busiFlag: 'dashboard' }, mockData)

      expect(store.getPanel.anyManage).toBe(true)
    })

    it('should handle empty data', async () => {
      const store = interactiveStore()
      await store.setInteractive({ busiFlag: 'dashboard' }, [])

      expect(store.getPanel.treeNodes).toEqual([])
      expect(store.getPanel.leafNodeCount).toBe(0)
    })

    it('should fall back to empty state when api rejects', async () => {
      const store = interactiveStore()
      vi.mocked(queryTreeApi).mockRejectedValueOnce(new Error('compatibility failed'))

      const res = await store.setInteractive({ busiFlag: 'dashboard' })

      expect(res).toEqual([])
      expect(store.getPanel.treeNodes).toEqual([])
      expect(store.getPanel.menuAuth).toBe(true)
    })
  })

  describe('clear', () => {
    it('should clear all data', async () => {
      const store = interactiveStore()
      
      const mockData = [
        { id: '1', pid: '0', name: 'Test', weight: 7, leaf: true, extraFlag: 0, extraFlag1: 0 }
      ]

      await store.setInteractive({ busiFlag: 'dashboard' }, mockData)
      await store.setInteractive({ busiFlag: 'dataV' }, mockData)

      store.clear()

      expect(store.getData).toEqual({})
    })
  })

  describe('Multiple busiFlag types', () => {
    it('should handle screen (dataV) type', async () => {
      const store = interactiveStore()
      const mockData = [
        { id: '1', pid: '0', name: 'Screen', weight: 7, leaf: true, extraFlag: 0, extraFlag1: 0 }
      ]

      await store.setInteractive({ busiFlag: 'dataV' }, mockData)

      expect(store.getScreen).toBeDefined()
    })

    it('should handle dataset type', async () => {
      const store = interactiveStore()
      const mockData = [
        { id: '1', pid: '0', name: 'Dataset', weight: 7, leaf: true, extraFlag: 0, extraFlag1: 0 }
      ]

      await store.setInteractive({ busiFlag: 'dataset' }, mockData)

      expect(store.getDataset).toBeDefined()
    })

    it('should handle datasource type', async () => {
      const store = interactiveStore()
      const mockData = [
        { id: '1', pid: '0', name: 'Datasource', weight: 7, leaf: true, extraFlag: 0, extraFlag1: 0 }
      ]

      await store.setInteractive({ busiFlag: 'datasource' }, mockData)

      expect(store.getDatasource).toBeDefined()
    })
  })

  describe('Node normalization', () => {
    it('should normalize node ids to strings', async () => {
        const store = interactiveStore()
        const mockData = [
          { id: 123, pid: 0, name: 'Test', weight: 7, leaf: true, extraFlag: 0, extraFlag1: 0 }
        ]

        await store.setInteractive({ busiFlag: 'dashboard' }, mockData)

        expect(store.getPanel.treeNodes[0].id).toBe('123')
        expect(store.getPanel.treeNodes[0].pid).toBe('0')
      })

    it('should normalize nested children ids', async () => {
        const store = interactiveStore()
        const mockData = [
          { id: 1, pid: 0, name: 'Parent', weight: 7, leaf: false, extraFlag: 0, extraFlag1: 0, children: [
            { id: 2, pid: 1, name: 'Child', weight: 7, leaf: true, extraFlag: 0, extraFlag1: 0 }
          ]}
        ]

        await store.setInteractive({ busiFlag: 'dashboard' }, mockData)

        const parent = store.getPanel.treeNodes[0]
        expect(parent.id).toBe('1')
        expect(parent.children[0].id).toBe('2')
        expect(parent.children[0].pid).toBe('1')
      })
  })

  describe('loadBusiInteractive', () => {
    it('should backfill missing compatibility tree entries with empty state', async () => {
      const store = interactiveStore()
      vi.mocked(queryBusiTreeApi).mockResolvedValueOnce({
        dashboard: [{ id: '1', pid: '0', name: 'Dashboard', weight: 7, leaf: true, extraFlag: 0, extraFlag1: 0 }]
      } as unknown as Awaited<ReturnType<typeof queryBusiTreeApi>>)

      await store.loadBusiInteractive()

      expect(store.getPanel.treeNodes).toHaveLength(1)
      expect(store.getScreen.treeNodes).toEqual([])
      expect(store.getDataset.treeNodes).toEqual([])
      expect(store.getDatasource.treeNodes).toEqual([])
    })

    it('should preserve real dashboard and screen resource nodes from interactiveTree', async () => {
      const store = interactiveStore()
      vi.mocked(queryBusiTreeApi).mockResolvedValueOnce({
        dashboard: [
          {
            id: '10',
            pid: '0',
            name: 'Dashboard Folder',
            weight: 9,
            leaf: false,
            extraFlag: 0,
            extraFlag1: 1,
            children: [
              { id: '11', pid: '10', name: 'Revenue Dashboard', weight: 9, leaf: true, extraFlag: 1, extraFlag1: 1 }
            ]
          }
        ],
        dataV: [{ id: '21', pid: '0', name: 'Executive Screen', weight: 9, leaf: true, extraFlag: 0, extraFlag1: 1 }]
      } as unknown as Awaited<ReturnType<typeof queryBusiTreeApi>>)

      await store.loadBusiInteractive()

      expect(store.getPanel.treeNodes[0].id).toBe('10')
      expect(store.getPanel.treeNodes[0].children[0].id).toBe('11')
      expect(store.getPanel.leafNodeCount).toBe(1)
      expect(store.getPanel.anyManage).toBe(true)
      expect(store.getScreen.treeNodes[0].name).toBe('Executive Screen')
    })

    it('should preserve dataset and datasource nodes from batched interactive loading', async () => {
      const store = interactiveStore()
      vi.mocked(queryBusiTreeApi).mockResolvedValueOnce({
        dataset: [
          {
            id: '31',
            pid: '0',
            name: 'Dataset Folder',
            weight: 9,
            leaf: false,
            extraFlag: 0,
            extraFlag1: 0,
            children: [{ id: '32', pid: '31', name: 'Sales Dataset', weight: 9, leaf: true, extraFlag: 0, extraFlag1: 0 }]
          }
        ],
        datasource: [{ id: '41', pid: '0', name: 'MySQL DS', weight: 9, leaf: true, extraFlag: 1, extraFlag1: 0 }]
      } as unknown as Awaited<ReturnType<typeof queryBusiTreeApi>>)

      await store.loadBusiInteractive()

      expect(store.getDataset.treeNodes[0].id).toBe('31')
      expect(store.getDataset.treeNodes[0].children[0].name).toBe('Sales Dataset')
      expect(store.getDatasource.treeNodes[0].name).toBe('MySQL DS')
    })

    it('should initialize through batched interactive loading instead of dataset or datasource direct tree calls', async () => {
      const store = interactiveStore()
      vi.mocked(queryBusiTreeApi).mockResolvedValueOnce({
        dashboard: [],
        dataV: [],
        dataset: [{ id: '51', pid: '0', name: 'Dataset A', weight: 9, leaf: true, extraFlag: 0, extraFlag1: 0 }],
        datasource: [{ id: '61', pid: '0', name: 'Datasource A', weight: 9, leaf: true, extraFlag: 1, extraFlag1: 0 }]
      } as unknown as Awaited<ReturnType<typeof queryBusiTreeApi>>)

      await store.initInteractive()

      expect(queryBusiTreeApi).toHaveBeenCalledTimes(1)
      expect(getDatasetTree).not.toHaveBeenCalled()
      expect(listDatasources).not.toHaveBeenCalled()
      expect(store.getDataset.treeNodes[0].id).toBe('51')
      expect(store.getDatasource.treeNodes[0].id).toBe('61')
    })

    it('should keep create permission when batched interactive scopes are authorized but empty', async () => {
      const store = interactiveStore()
      vi.mocked(queryBusiTreeApi).mockResolvedValueOnce({
        dashboard: [],
        dataV: [],
        dataset: [],
        datasource: []
      } as unknown as Awaited<ReturnType<typeof queryBusiTreeApi>>)

      await store.loadBusiInteractive()

      expect(store.getPanel.menuAuth).toBe(true)
      expect(store.getPanel.anyManage).toBe(true)
      expect(store.getScreen.anyManage).toBe(true)
      expect(store.getDataset.anyManage).toBe(true)
      expect(store.getDatasource.anyManage).toBe(true)
    })
  })
})
