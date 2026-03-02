const { spawnSync } = require('node:child_process')
const { appendFileSync } = require('node:fs')

const args = process.argv.slice(2)

const hasFlag = flag => args.includes(`--${flag}`)

const getArg = (name, fallback = '') => {
  const key = `--${name}=`
  const found = args.find(arg => arg.startsWith(key))
  return found ? found.slice(key.length) : fallback
}

const normalizeStatus = raw => {
  const value = String(raw || '').trim().toLowerCase()
  if (!value) return 'unknown'
  if (['success', 'pass', 'passed'].includes(value)) return 'pass'
  if (['failure', 'failed', 'error'].includes(value)) return 'fail'
  if (['pending', 'in_progress', 'queued'].includes(value)) return 'pending'
  if (['cancelled', 'canceled'].includes(value)) return 'cancelled'
  if (['skipped', 'neutral'].includes(value)) return 'skipped'
  return value
}

const statusLabel = status => {
  if (status === 'pass') return 'pass'
  if (status === 'fail') return 'fail'
  return status
}

const formatDuration = ms => {
  if (typeof ms !== 'number' || Number.isNaN(ms)) return '-'
  if (ms < 1000) return `${ms}ms`
  return `${(ms / 1000).toFixed(2)}s`
}

const buildMarkdown = ({ rows, buildStatus, affectedTarget, affectedStatus }) => {
  const lines = [
    '## Frontend Quality Report',
    '',
    '| Check | Status | Duration |',
    '| --- | --- | --- |'
  ]

  rows.forEach(row => {
    lines.push(`| ${row.name} | ${statusLabel(row.status)} | ${row.duration} |`)
  })

  lines.push('')
  lines.push(`- Build: ${statusLabel(buildStatus || 'unknown')}`)
  lines.push(`- Affected target: ${affectedTarget || 'none'}`)
  lines.push(`- Affected tests: ${statusLabel(affectedStatus || (affectedTarget === 'none' ? 'skipped' : 'unknown'))}`)

  return `${lines.join('\n')}\n`
}

const writeReport = markdown => {
  process.stdout.write(`${markdown}\n`)
  if (process.env.GITHUB_STEP_SUMMARY) {
    appendFileSync(process.env.GITHUB_STEP_SUMMARY, `${markdown}\n`)
  }
}

const summaryMode = hasFlag('summary')

if (summaryMode) {
  const rows = [
    {
      name: 'TypeScript',
      status: normalizeStatus(getArg('ts', process.env.DE_TS || 'unknown')),
      duration: getArg('ts-ms', '-')
    },
    {
      name: 'ESLint',
      status: normalizeStatus(getArg('lint', process.env.DE_LINT || 'unknown')),
      duration: getArg('lint-ms', '-')
    },
    {
      name: 'Tests (core)',
      status: normalizeStatus(getArg('test-core', process.env.DE_TEST_CORE || 'unknown')),
      duration: getArg('test-core-ms', '-')
    }
  ]

  const markdown = buildMarkdown({
    rows,
    buildStatus: normalizeStatus(getArg('build', process.env.DE_BUILD || 'unknown')),
    affectedTarget: getArg('affected-target', process.env.DE_AFFECTED_TARGET || 'none'),
    affectedStatus: normalizeStatus(getArg('affected', process.env.DE_AFFECTED || 'unknown'))
  })

  writeReport(markdown)
  process.exit(0)
}

const checks = [
  { name: 'TypeScript', command: 'npm run ts:check' },
  { name: 'ESLint', command: 'npm run lint' },
  { name: 'Tests (core)', command: 'npm run test:core' }
]

let hasFailure = false
const rows = []

checks.forEach(check => {
  const start = Date.now()
  const result = spawnSync(check.command, {
    shell: true,
    stdio: 'inherit',
    env: process.env
  })
  const duration = Date.now() - start
  const status = result.status === 0 ? 'pass' : 'fail'

  if (status === 'fail') hasFailure = true

  rows.push({
    name: check.name,
    status,
    duration: formatDuration(duration)
  })
})

const markdown = buildMarkdown({
  rows,
  buildStatus: 'n/a',
  affectedTarget: 'n/a',
  affectedStatus: 'n/a'
})

writeReport(markdown)
process.exit(hasFailure ? 1 : 0)
