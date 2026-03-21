import { beforeEach, describe, expect, it, vi } from 'vitest'

const { requestMock } = vi.hoisted(() => ({
  requestMock: {
    post: vi.fn(),
    get: vi.fn()
  }
}))

vi.mock('@/config/axios', () => ({
  default: requestMock
}))

import { exportRetry, exportTasks, exportTasksRecords, getDatasetTree } from '@/api/dataset'

describe('Dataset API wrappers', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('injects dataset busiFlag and normalizes leaf state from nodeType', async () => {
    requestMock.post.mockResolvedValueOnce({
      data: [
        {
          id: 1,
          name: 'Dataset Folder',
          nodeType: 'folder',
          children: [{ id: 2, name: 'Sales Dataset', nodeType: 'dataset' }]
        }
      ]
    })

    const result = await getDatasetTree({})

    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/datasetTree/tree',
      data: { busiFlag: 'dataset' }
    })
    expect(result[0].leaf).toBe(false)
    expect(result[0].children[0].leaf).toBe(true)
  })

  it('requests export-center task counters through the records alias', async () => {
    requestMock.post.mockResolvedValueOnce({ data: { ALL: 3, FAILED: 1 } })

    const result = await exportTasksRecords()

    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/api/exportCenter/exportTasks/records',
      data: {}
    })
    expect(result).toEqual({ data: { ALL: 3, FAILED: 1 } })
  })

  it('requests export-center tasks for the given status and pagination', async () => {
    requestMock.post.mockResolvedValueOnce({ data: { total: 1, records: [] } })

    const result = await exportTasks(2, 20, 'FAILED')

    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/api/exportCenter/exportTasks/FAILED/2/20',
      data: {}
    })
    expect(result).toEqual({ data: { total: 1, records: [] } })
  })

  it('retries an export task through the retry alias', async () => {
    requestMock.post.mockResolvedValueOnce({ data: { code: '000000' } })

    const result = await exportRetry('task-9')

    expect(requestMock.post).toHaveBeenCalledWith({
      url: '/api/exportCenter/retry/task-9',
      data: {}
    })
    expect(result).toEqual({ code: '000000' })
  })
})
