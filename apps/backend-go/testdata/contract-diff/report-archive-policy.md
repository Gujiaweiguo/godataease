# Report Archive Policy

## Purpose

Contract diff reports are archived to ensure traceability and enable historical analysis of API contract changes. This policy defines how reports are named, stored, and retained in the CI pipeline.

## Report Types

Two report formats are generated for each contract diff run:

| Format | Filename | Purpose |
|--------|----------|---------|
| JSON | `contract-diff-report.json` | Machine-readable, suitable for programmatic processing |
| Markdown | `contract-diff-report.md` | Human-readable, suitable for code review and documentation |

## Naming Convention

### For Commits

```
contract-diff-{commit-sha}-{timestamp}.{ext}
```

Example:
- `contract-diff-abc1234-20260218-143022.json`
- `contract-diff-abc1234-20260218-143022.md`

### For Pull Requests

```
contract-diff-pr{number}-{timestamp}.{ext}
```

Example:
- `contract-diff-pr42-20260218-143022.json`
- `contract-diff-pr42-20260218-143022.md`

### Timestamp Format

- Format: `YYYYMMDD-HHMMSS` (UTC)
- Example: `20260218-143022`

## Storage Location

### Primary: GitHub Actions Artifacts

Reports are uploaded as GitHub Actions artifacts:

- **Artifact Name**: `contract-diff-report-{run-id}`
- **Retention Period**: 30 days
- **Access**: Available in the workflow run summary under "Artifacts"

### Fallback: Job Logs

If artifact upload fails, the full report content is output to job logs:

```yaml
- name: Output Report to Logs (Fallback)
  if: failure()
  run: cat backend-go/testdata/contract-diff/contract-diff-report.md
```

## Retention Policy

| Storage Type | Retention Period | Notes |
|--------------|------------------|-------|
| GitHub Artifacts | 30 days | Auto-deleted after retention period |
| Job Logs | 90 days | GitHub default log retention |
| Git History | Indefinite | Commit SHA reference preserved in git |

## Traceability

### Finding Reports by Commit

1. Navigate to GitHub Actions workflow runs
2. Filter by commit SHA in the search box
3. Open the workflow run and download artifacts

### Finding Reports by PR

1. Navigate to the Pull Request
2. Check the "Checks" tab
3. Find the `contract-diff` job
4. Download artifacts from the workflow run

### Report Metadata

Each report includes metadata for verification:

```json
{
  "metadata": {
    "commit_sha": "abc1234...",
    "pr_number": 42,
    "generated_at": "2026-02-18T14:30:22Z",
    "workflow_run_id": 12345678901
  }
}
```

## Related Documents

- [Output Schema](./output-schema.md) - Report structure and fields
- [Critical Whitelist](./critical-whitelist.yaml) - Allowed breaking changes
- [Whitelist Governance](./whitelist-governance.md) - Whitelist management process
