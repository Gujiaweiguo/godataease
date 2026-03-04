const { readFileSync, writeFileSync, mkdirSync, appendFileSync } = require('node:fs')
const { dirname } = require('node:path')

const args = process.argv.slice(2)

const getArg = (name, fallback = '') => {
  const prefix = `--${name}=`
  const matched = args.find(arg => arg.startsWith(prefix))
  if (matched) {
    return matched.slice(prefix.length)
  }

  const key = `--${name}`
  const index = args.findIndex(arg => arg === key)
  if (index >= 0 && index + 1 < args.length) {
    return args[index + 1]
  }

  return fallback
}

const normalizePath = value => String(value || '').trim()

const readJson = filePath => {
  return JSON.parse(readFileSync(filePath, 'utf-8'))
}

const pct = value => `${value.toFixed(2)}%`

const metricConfig = [
  { key: 'searchOptimizedMedianMs', label: 'search optimized median (ms)' },
  { key: 'expandOptimizedMedianMs', label: 'expand optimized median (ms)' },
  { key: 'pruneOptimizedMedianMs', label: 'prune optimized median (ms)' }
]

const classify = (deltaPct, thresholdPct, baseline, current, ignoreBelowMs) => {
  if (Math.max(baseline, current) < ignoreBelowMs) {
    return 'stable'
  }
  if (deltaPct <= -thresholdPct) {
    return 'improved'
  }
  if (deltaPct >= thresholdPct) {
    return 'regressed'
  }
  return 'stable'
}

const asScaleMap = report => {
  const results = Array.isArray(report?.results) ? report.results : []
  return new Map(results.map(item => [item.scale, item]))
}

const buildMarkdown = ({ rows, thresholdPct, ignoreBelowMs, baselinePath, currentPath }) => {
  const lines = [
    '## Datasource Tree Benchmark Comparison',
    '',
    `- Threshold: ${thresholdPct}%`,
    `- Ignore below: ${ignoreBelowMs}ms`,
    `- Baseline: ${baselinePath}`,
    `- Current: ${currentPath}`,
    '',
    '| Scale | Metric | Baseline | Current | Delta | Trend |',
    '| --- | --- | ---: | ---: | ---: | --- |'
  ]

  rows.forEach(row => {
    lines.push(
      `| ${row.scale} | ${row.metric} | ${row.baseline.toFixed(2)} | ${row.current.toFixed(2)} | ${pct(row.deltaPct)} | ${row.trend} |`
    )
  })

  const improved = rows.filter(row => row.trend === 'improved').length
  const regressed = rows.filter(row => row.trend === 'regressed').length
  const stable = rows.filter(row => row.trend === 'stable').length

  lines.push('')
  lines.push(`- Improved: ${improved}`)
  lines.push(`- Stable: ${stable}`)
  lines.push(`- Regressed: ${regressed}`)

  return `${lines.join('\n')}\n`
}

const main = () => {
  const baselinePath = normalizePath(getArg('baseline'))
  const currentPath = normalizePath(getArg('current'))
  const outputPath = normalizePath(getArg('output'))
  const jsonOutputPath = normalizePath(getArg('json-output'))
  const thresholdPct = Number(getArg('threshold', '5'))
  const ignoreBelowMs = Number(getArg('ignore-below-ms', '0.5'))
  const failOnRegression = String(getArg('fail-on-regression', 'false')).toLowerCase() === 'true'

  if (!baselinePath || !currentPath) {
    throw new Error('Missing required args: --baseline=<path> --current=<path>')
  }

  const baselineReport = readJson(baselinePath)
  const currentReport = readJson(currentPath)

  const baselineMap = asScaleMap(baselineReport)
  const currentMap = asScaleMap(currentReport)

  const rows = []

  baselineMap.forEach((baselineItem, scale) => {
    const currentItem = currentMap.get(scale)
    if (!currentItem) {
      return
    }

    metricConfig.forEach(metric => {
      const baselineValue = Number(baselineItem[metric.key])
      const currentValue = Number(currentItem[metric.key])

      if (Number.isNaN(baselineValue) || Number.isNaN(currentValue) || baselineValue === 0) {
        return
      }

      const deltaPct = ((currentValue - baselineValue) / baselineValue) * 100
      rows.push({
        scale,
        metric: metric.label,
        baseline: baselineValue,
        current: currentValue,
        deltaPct,
        trend: classify(deltaPct, thresholdPct, baselineValue, currentValue, ignoreBelowMs)
      })
    })
  })

  const markdown = buildMarkdown({ rows, thresholdPct, ignoreBelowMs, baselinePath, currentPath })
  process.stdout.write(markdown)

  if (outputPath) {
    mkdirSync(dirname(outputPath), { recursive: true })
    writeFileSync(outputPath, markdown)
  }

  const jsonPayload = {
    generatedAt: new Date().toISOString(),
    thresholdPct,
    ignoreBelowMs,
    baselinePath,
    currentPath,
    rows
  }

  if (jsonOutputPath) {
    mkdirSync(dirname(jsonOutputPath), { recursive: true })
    writeFileSync(jsonOutputPath, `${JSON.stringify(jsonPayload, null, 2)}\n`)
  }

  if (process.env.GITHUB_STEP_SUMMARY) {
    appendFileSync(process.env.GITHUB_STEP_SUMMARY, `${markdown}\n`)
  }

  if (failOnRegression && rows.some(row => row.trend === 'regressed')) {
    process.exit(1)
  }
}

main()
