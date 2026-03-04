import { describe, expect, it } from 'vitest'
import { performance } from 'node:perf_hooks'
import { dirname } from 'node:path'
import { mkdirSync, writeFileSync } from 'node:fs'
import type { BusiTreeNode } from '@/models/tree/TreeNode'
import {
  createFilterKeywordCache,
  filterNodeByKeyword,
  mutateExpandedKeySet,
  pruneExpandedKeySet
} from '@/views/visualized/data/datasource/treeState'

type SamplingConfig = {
  warmup: number
  samples: number
}

type SamplingResult = {
  medianMs: number
  meanMs: number
  minMs: number
  maxMs: number
  samples: number
  trimmedSamples: number
}

type BenchmarkResult = {
  scale: string
  samples: number
  trimmedSamples: number
  searchLegacyMedianMs: number
  searchOptimizedMedianMs: number
  searchSpeedup: number
  expandLegacyMedianMs: number
  expandOptimizedMedianMs: number
  expandSpeedup: number
  pruneLegacyMedianMs: number
  pruneOptimizedMedianMs: number
  pruneSpeedup: number
}

type BenchmarkReport = {
  generatedAt: string
  sampling: SamplingConfig
  results: BenchmarkResult[]
}

const SAMPLING_CONFIG: SamplingConfig = {
  warmup: 3,
  samples: 13
}

const createDatasourceTree = (totalNodes: number) => {
  const folderSize = 20
  const folderCount = Math.max(1, Math.ceil(totalNodes / folderSize))
  let nextId = 1
  const treeNodes: BusiTreeNode[] = []
  const allIds: Array<string | number> = []

  for (let folderIndex = 0; folderIndex < folderCount; folderIndex++) {
    const folderId = `folder-${nextId++}`
    const children: BusiTreeNode[] = []

    for (let leafIndex = 0; leafIndex < folderSize; leafIndex++) {
      if (allIds.length >= totalNodes) {
        break
      }
      const leafId = `leaf-${nextId++}`
      const leafNode: BusiTreeNode = {
        id: leafId,
        pid: folderId,
        name: `Datasource_${folderIndex}_${leafIndex}_MySQL`,
        leaf: true,
        weight: 7,
        extraFlag: 0,
        extraFlag1: 0
      }
      children.push(leafNode)
      allIds.push(leafId)
    }

    const folderNode: BusiTreeNode = {
      id: folderId,
      pid: '0',
      name: `Folder_${folderIndex}`,
      leaf: false,
      weight: 7,
      extraFlag: 0,
      extraFlag1: 0,
      children
    }
    treeNodes.push(folderNode)
    allIds.push(folderId)
  }

  return { treeNodes, allIds }
}

const flattenNodes = (treeNodes: BusiTreeNode[]) => {
  const result: BusiTreeNode[] = []
  const walk = (nodes: BusiTreeNode[]) => {
    nodes.forEach(node => {
      result.push(node)
      if (node.children?.length) {
        walk(node.children)
      }
    })
  }
  walk(treeNodes)
  return result
}

const formatMs = (ms: number) => Number(ms.toFixed(2))

const measureOnce = (fn: () => void) => {
  const start = performance.now()
  fn()
  return formatMs(performance.now() - start)
}

const median = (values: number[]) => {
  if (!values.length) return 0
  const mid = Math.floor(values.length / 2)
  if (values.length % 2 === 0) {
    return (values[mid - 1] + values[mid]) / 2
  }
  return values[mid]
}

const average = (values: number[]) => {
  if (!values.length) return 0
  const sum = values.reduce((acc, value) => acc + value, 0)
  return sum / values.length
}

const trimOutliers = (sortedValues: number[]) => {
  if (sortedValues.length <= 4) {
    return sortedValues
  }
  return sortedValues.slice(1, -1)
}

const measureSamples = (
  fn: () => void,
  samplingConfig: SamplingConfig = SAMPLING_CONFIG
): SamplingResult => {
  const { warmup, samples } = samplingConfig

  for (let index = 0; index < warmup; index++) {
    fn()
  }

  const rawSamples: number[] = []
  for (let index = 0; index < samples; index++) {
    rawSamples.push(measureOnce(fn))
  }

  const sortedSamples = [...rawSamples].sort((left, right) => left - right)
  const trimmedSamples = trimOutliers(sortedSamples)

  return {
    medianMs: formatMs(median(trimmedSamples)),
    meanMs: formatMs(average(trimmedSamples)),
    minMs: formatMs(sortedSamples[0] ?? 0),
    maxMs: formatMs(sortedSamples[sortedSamples.length - 1] ?? 0),
    samples,
    trimmedSamples: trimmedSamples.length
  }
}

const calcSpeedup = (legacyMs: number, optimizedMs: number) => {
  if (optimizedMs === 0) {
    return 0
  }
  return formatMs(legacyMs / optimizedMs)
}

const benchmarkSearch = (nodes: BusiTreeNode[]) => {
  const keywords = ['m', 'my', 'mys', 'mysql', 'mysql_main']
  const rounds = 20

  const legacy = measureSamples(() => {
    let hitCount = 0
    for (let round = 0; round < rounds; round++) {
      keywords.forEach(keyword => {
        nodes.forEach(node => {
          if (node.name.toLowerCase().includes(keyword.toLowerCase())) {
            hitCount++
          }
        })
      })
    }
    if (hitCount < 0) {
      throw new Error('unreachable')
    }
  })

  const optimized = measureSamples(() => {
    let hitCount = 0
    const keywordCache = createFilterKeywordCache()
    for (let round = 0; round < rounds; round++) {
      keywords.forEach(keyword => {
        nodes.forEach(node => {
          if (filterNodeByKeyword(keyword, node, keywordCache)) {
            hitCount++
          }
        })
      })
    }
    if (hitCount < 0) {
      throw new Error('unreachable')
    }
  })

  return { legacy, optimized }
}

const benchmarkExpandMutation = (allIds: Array<string | number>) => {
  const half = Math.floor(allIds.length / 2)
  const initialExpanded = allIds.slice(0, half)
  const operations = allIds.slice(0, Math.min(allIds.length, half + 1000))

  const legacy = measureSamples(() => {
    let expandedSet = new Set(initialExpanded)
    const cacheSink: Array<Array<string | number>> = []
    operations.forEach(id => {
      const nextExpandedSet = new Set(expandedSet)
      if (nextExpandedSet.has(id)) {
        nextExpandedSet.delete(id)
      } else {
        nextExpandedSet.add(id)
      }
      expandedSet = nextExpandedSet
      cacheSink.push(Array.from(expandedSet))
    })
    if (cacheSink.length < 0) {
      throw new Error('unreachable')
    }
  })

  const optimized = measureSamples(() => {
    const expandedSet = new Set(initialExpanded)
    const cacheSink: Array<Array<string | number>> = []
    let persistPending = false
    const batchSize = 50

    operations.forEach((id, index) => {
      const changed = mutateExpandedKeySet(expandedSet, id, !expandedSet.has(id))
      if (changed) {
        persistPending = true
      }

      if ((index + 1) % batchSize === 0 && persistPending) {
        cacheSink.push(Array.from(expandedSet))
        persistPending = false
      }
    })

    if (persistPending) {
      cacheSink.push(Array.from(expandedSet))
    }

    if (cacheSink.length < 0) {
      throw new Error('unreachable')
    }
  })

  return { legacy, optimized }
}

const collectValidIdsAsArray = (treeNodes: BusiTreeNode[]) => {
  const validIds: Array<string | number> = []
  const walk = (nodes: BusiTreeNode[]) => {
    nodes.forEach(node => {
      validIds.push(node.id)
      if (node.children?.length) {
        walk(node.children)
      }
    })
  }
  walk(treeNodes)
  return validIds
}

const benchmarkPrune = (treeNodes: BusiTreeNode[], allIds: Array<string | number>) => {
  const staleIds = Array.from({ length: Math.max(100, Math.floor(allIds.length * 0.3)) }, (_, index) => {
    return `stale-${index}`
  })
  const expandedIds = allIds.concat(staleIds)

  const legacy = measureSamples(() => {
    const validIds = collectValidIdsAsArray(treeNodes)
    const nextExpandedKeys = expandedIds.filter(id => validIds.includes(id))
    const changed = nextExpandedKeys.length !== expandedIds.length
    if (!changed && nextExpandedKeys.length < 0) {
      throw new Error('unreachable')
    }
  })

  const optimized = measureSamples(() => {
    const { nextExpandedKeySet, changed } = pruneExpandedKeySet(new Set(expandedIds), treeNodes)
    if (!changed && nextExpandedKeySet.size < 0) {
      throw new Error('unreachable')
    }
  })

  return { legacy, optimized }
}

const runBaselineAtScale = (nodeCount: number): BenchmarkResult => {
  const { treeNodes, allIds } = createDatasourceTree(nodeCount)
  const flattenedNodes = flattenNodes(treeNodes)
  const search = benchmarkSearch(flattenedNodes)
  const expand = benchmarkExpandMutation(allIds)
  const prune = benchmarkPrune(treeNodes, allIds)

  return {
    scale: `${nodeCount} nodes`,
    samples: search.legacy.samples,
    trimmedSamples: search.legacy.trimmedSamples,
    searchLegacyMedianMs: search.legacy.medianMs,
    searchOptimizedMedianMs: search.optimized.medianMs,
    searchSpeedup: calcSpeedup(search.legacy.medianMs, search.optimized.medianMs),
    expandLegacyMedianMs: expand.legacy.medianMs,
    expandOptimizedMedianMs: expand.optimized.medianMs,
    expandSpeedup: calcSpeedup(expand.legacy.medianMs, expand.optimized.medianMs),
    pruneLegacyMedianMs: prune.legacy.medianMs,
    pruneOptimizedMedianMs: prune.optimized.medianMs,
    pruneSpeedup: calcSpeedup(prune.legacy.medianMs, prune.optimized.medianMs)
  }
}

const writeBenchmarkReportIfNeeded = (results: BenchmarkResult[]) => {
  const outputPath = process.env.BENCH_OUTPUT_JSON
  if (!outputPath) {
    return
  }

  const report: BenchmarkReport = {
    generatedAt: new Date().toISOString(),
    sampling: SAMPLING_CONFIG,
    results
  }

  mkdirSync(dirname(outputPath), { recursive: true })
  writeFileSync(outputPath, `${JSON.stringify(report, null, 2)}\n`, 'utf-8')
  console.log(`[bench] JSON report written to ${outputPath}`)
}

describe('datasource tree performance baseline', () => {
  it('prints 1k/5k baseline with trimmed-median sampling for interactions', () => {
    const baselineResults = [runBaselineAtScale(1000), runBaselineAtScale(5000)]
    console.table(baselineResults)
    writeBenchmarkReportIfNeeded(baselineResults)

    baselineResults.forEach(result => {
      expect(result.searchLegacyMedianMs).toBeGreaterThanOrEqual(0)
      expect(result.searchOptimizedMedianMs).toBeGreaterThanOrEqual(0)
      expect(result.expandLegacyMedianMs).toBeGreaterThanOrEqual(0)
      expect(result.expandOptimizedMedianMs).toBeGreaterThanOrEqual(0)
      expect(result.pruneLegacyMedianMs).toBeGreaterThanOrEqual(0)
      expect(result.pruneOptimizedMedianMs).toBeGreaterThanOrEqual(0)
      expect(result.samples).toBe(SAMPLING_CONFIG.samples)
      expect(result.trimmedSamples).toBe(
        SAMPLING_CONFIG.samples > 4 ? SAMPLING_CONFIG.samples - 2 : SAMPLING_CONFIG.samples
      )
    })
  })
})
