# Drift Detection Rules

Definitions and procedures for detecting and responding to API contract drift.

## Drift Types

### Response Structure Drift
- **Field Added**: New field appears in response not in baseline
- **Field Removed**: Expected field missing from response
- **Field Renamed**: Same value under different key name
- **Type Changed**: Field type differs (string → number, etc.)

### Value Drift
- **Non-deterministic Values**: Fields that change on each request (IDs, UUIDs)
- **Timestamp Drift**: Date/time fields that vary legitimately
- **Dynamic Data**: User-specific or context-dependent values

### Behavior Drift
- **Status Code Change**: Different HTTP status than baseline
- **Error Message Change**: Different error text or structure
- **Header Change**: Response headers differ from baseline

### Frequency Drift
- **High Update Rate**: Too many baseline updates in short period
- **Rapid Cycling**: Baseline updated then reverted repeatedly

## Detection Rules

| Rule ID | Condition | Severity | Action |
|---------|-----------|----------|--------|
| DR-001 | >5 baseline updates/week per endpoint | Warning | Require review |
| DR-002 | >10 baseline updates/week per endpoint | Critical | Escalate + freeze |
| DR-003 | Structural change (add/remove field) | High | Alert team |
| DR-004 | Type change in existing field | Critical | Block auto-approve |
| DR-005 | Status code differs from baseline | High | Immediate alert |
| DR-006 | 3+ FPs on same field in 7 days | Warning | Review field rules |

## False Positive / Negative Handling

### Classification Criteria

**False Positive (FP)**: Drift flagged but behavior is expected/valid
- Planned API changes documented in changelog
- Dynamic fields missing ignore rules
- Test environment inconsistencies

**False Negative (FN)**: Real drift not detected
- Type coercion (number as string)
- Nested object changes
- Array order differences

### Dispute Resolution Process
1. Developer flags FP/FN in review comment
2. Maintainer validates within 24 hours
3. Update detection rules if needed
4. Document case in archive

### Archive Location
- FP/FN cases: `testdata/contract-diff/drift-archive/`
- Format: `YYYY-MM-DD_drift-case-{id}.json`

## Response Procedures

### Automated Alerts
- Slack: `#api-contract-alerts` channel
- Email: api-team@company.com
- CI: Block merge on Critical severity

### Manual Review Requirements
- All High/Critical drift requires human approval
- Review window: 4 hours during business hours
- After hours: Auto-escalate to on-call

### Escalation Paths
1. **L1**: Contract diff maintainer (initial review)
2. **L2**: API team lead (approval authority)
3. **L3**: Architecture review board (breaking changes)
