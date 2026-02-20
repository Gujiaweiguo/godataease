# Baseline Policy

## Purpose

The baseline serves as the authoritative reference for API contract diff comparisons in the Go migration project. It captures the expected behavior of the Java backend, enabling detection of discrepancies when the Go implementation deviates from established contracts.

## Baseline Definition

A baseline consists of captured HTTP responses from the Java backend for key API endpoints. Each baseline represents:

- **Golden responses**: Expected JSON output structure and values
- **Response metadata**: HTTP status codes, headers, timing
- **Data contracts**: Field types, nullability, and value ranges

Baselines are considered frozen once established and require formal approval for any modifications.

## Establishing a Baseline

### When to Create

- Before starting Go implementation of an endpoint
- When adding new test coverage for existing endpoints
- After major Java backend changes that alter response structure

### Process

1. Execute Java backend against representative test scenarios
2. Capture responses using the baseline collection tool
3. Validate responses against known-good examples
4. Commit baseline files to version control
5. Tag commit with baseline version identifier

### Version Control

Baselines are stored in Git and tagged using semantic versioning:
- Format: `baseline-v{MAJOR}.{MINOR}`
- Example: `baseline-v1.0`, `baseline-v1.1`

## Updating Baselines

### Allowed Updates

- Bug fixes in Java backend that correct response structure
- Documented API changes with proper change management approval
- Quarterly maintenance reviews

### Approval Process

1. Submit change request with justification
2. Technical lead review and approval
3. Update baseline files
4. Increment version tag
5. Notify affected teams

### Restricted Changes

- No updates during active Go migration sprints
- No updates without documented rationale
- No partial updates to baseline sets

## Baseline Files

### Location

```
backend-go/testdata/contract-diff/baselines/
├── auth/
│   ├── login.json
│   └── logout.json
├── dataset/
│   ├── list.json
│   └── detail.json
└── dashboard/
    ├── create.json
    └── export.json
```

### Format

- File type: JSON
- Naming: `{endpoint-name}.json`
- Encoding: UTF-8
- Pretty-printed with 2-space indentation

### Structure

```json
{
  "endpoint": "/api/dataset/list",
  "method": "POST",
  "captured_at": "2025-01-15T10:30:00Z",
  "java_version": "2.10.0",
  "response": { ... }
}
```

## Maintenance Schedule

| Activity | Frequency | Owner |
|----------|-----------|-------|
| Baseline integrity check | Weekly | CI Pipeline |
| Drift analysis | Monthly | QA Team |
| Full baseline review | Quarterly | Tech Lead |
| Archive cleanup | Quarterly | DevOps |

### Quarterly Review Checklist

- [ ] Verify all active endpoints have baselines
- [ ] Remove baselines for deprecated endpoints
- [ ] Update baselines for documented API changes
- [ ] Validate baseline coverage metrics
