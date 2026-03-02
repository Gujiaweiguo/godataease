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

const formatDate = raw => {
  if (raw) return raw
  return new Date().toISOString().slice(0, 10)
}

const shortSha = raw => {
  if (!raw) return 'n/a'
  return String(raw).slice(0, 7)
}

const buildRunUrl = () => {
  const server = process.env.GITHUB_SERVER_URL
  const repo = process.env.GITHUB_REPOSITORY
  const runId = process.env.GITHUB_RUN_ID
  if (!server || !repo || !runId) return 'n/a'
  return `${server}/${repo}/actions/runs/${runId}`
}

const buildMarkdown = ({ rows, buildStatus, affectedTarget, affectedStatus, baselineMeta }) => {
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
  lines.push('')
  lines.push('## Weekly Baseline Snapshot')
  lines.push('')
  lines.push('| Field | Value |')
  lines.push('| --- | --- |')
  lines.push(`| Date (UTC) | ${baselineMeta.date} |`)
  lines.push(`| Branch | ${baselineMeta.branch} |`)
  lines.push(`| Commit | ${baselineMeta.commit} |`)
  lines.push(`| TypeScript | ${statusLabel(baselineMeta.tsStatus)} |`)
  lines.push(`| ESLint | ${statusLabel(baselineMeta.lintStatus)} |`)
  lines.push(`| Tests (core) | ${statusLabel(baselineMeta.testCoreStatus)} |`)
  lines.push(`| Build | ${statusLabel(baselineMeta.buildStatus)} |`)
  lines.push(`| Affected target | ${baselineMeta.affectedTarget || 'none'} |`)
  lines.push(`| Affected tests | ${statusLabel(baselineMeta.affectedStatus)} |`)
  lines.push(`| Workflow run | ${baselineMeta.runUrl} |`)

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
  const buildStatus = normalizeStatus(getArg('build', process.env.DE_BUILD || 'unknown'))
  const affectedTarget = getArg('affected-target', process.env.DE_AFFECTED_TARGET || 'none')
  const affectedStatus = normalizeStatus(getArg('affected', process.env.DE_AFFECTED || 'unknown'))
  const tsStatus = normalizeStatus(getArg('ts', process.env.DE_TS || 'unknown'))
  const lintStatus = normalizeStatus(getArg('lint', process.env.DE_LINT || 'unknown'))
  const testCoreStatus = normalizeStatus(getArg('test-core', process.env.DE_TEST_CORE || 'unknown'))

  const rows = [
    {
      name: 'TypeScript',
      status: tsStatus,
      duration: getArg('ts-ms', '-')
    },
    {
      name: 'ESLint',
      status: lintStatus,
      duration: getArg('lint-ms', '-')
    },
    {
      name: 'Tests (core)',
      status: testCoreStatus,
      duration: getArg('test-core-ms', '-')
    }
  ]

  const markdown = buildMarkdown({
    rows,
    buildStatus,
    affectedTarget,
    affectedStatus,
    baselineMeta: {
      date: formatDate(getArg('date', process.env.DE_BASELINE_DATE || '')),
      branch:
        getArg('branch', process.env.DE_BASELINE_BRANCH || '') ||
        process.env.GITHUB_REF_NAME ||
        process.env.GITHUB_HEAD_REF ||
        'local',
      commit: shortSha(
        getArg('commit', process.env.DE_BASELINE_COMMIT || '') || process.env.GITHUB_SHA || 'local'
      ),
      tsStatus,
      lintStatus,
      testCoreStatus,
      buildStatus,
      affectedTarget,
      affectedStatus,
      runUrl: getArg('run-url', process.env.DE_BASELINE_RUN_URL || '') || buildRunUrl()
    }
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
  affectedStatus: 'n/a',
  baselineMeta: {
    date: formatDate(''),
    branch: 'local',
    commit: 'local',
    tsStatus: rows[0]?.status || 'unknown',
    lintStatus: rows[1]?.status || 'unknown',
    testCoreStatus: rows[2]?.status || 'unknown',
    buildStatus: 'n/a',
    affectedTarget: 'n/a',
    affectedStatus: 'n/a',
    runUrl: 'n/a'
  }
})

writeReport(markdown)
process.exit(hasFailure ? 1 : 0)
