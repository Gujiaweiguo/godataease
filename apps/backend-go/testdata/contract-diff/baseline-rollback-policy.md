# Baseline Rollback Policy

This document defines the rollback procedures for contract diff baselines when issues are detected.

## Rollback Triggers

A baseline rollback should be initiated when any of the following conditions occur:

| Trigger | Description | Severity |
|---------|-------------|----------|
| Gate Failure | Baseline regression causing CI gate failures | High |
| Baseline Drift | Significant drift producing false positives | High |
| Production Incident | Incident linked directly to a baseline change | Critical |
| Critical Misclassification | System marking breaking changes as compatible | Critical |

## Rollback Steps

### Step 1: Identify Stable Baseline Version
- Review git history to find the last stable baseline commit
- Check audit logs for successful gate passes
- Confirm version has no associated incidents

### Step 2: Execute Revert
```bash
# Option A: Revert specific commit
git revert <commit-sha>

# Option B: Checkout known stable version
git checkout <stable-commit-sha> -- backend-go/testdata/contract-diff/baselines/
```

### Step 3: Verify Baseline Integrity
- Run contract diff against reverted baseline
- Confirm all expected contracts are present
- Validate JSON schema compliance

### Step 4: Re-run Contract Diff
```bash
make contract-diff
```

### Step 5: Confirm Gate Passes
- Verify CI pipeline succeeds
- Confirm no new failures introduced
- Document resolution in incident ticket

## Verification Requirements

### Evidence to Capture
- [ ] Before/after baseline diff output
- [ ] Git log showing reverted commits
- [ ] Contract diff results post-rollback
- [ ] CI run confirmation (green build URL)

### Sign-off Requirements
- [ ] Engineer confirms rollback executed
- [ ] Tech lead reviews evidence
- [ ] QA validates no regressions

### Audit Log Requirements
- Timestamp of rollback initiation
- Reason code (from trigger table)
- Baseline versions (from -> to)
- Approver identity
- Evidence artifacts (links)

## Escalation Path

If rollback fails, follow this escalation sequence:

1. **L1: Engineering Lead** - Immediate triage (within 15 minutes)
2. **L2: Platform Team** - Infrastructure investigation (within 1 hour)
3. **L3: Incident Commander** - Full incident response (within 2 hours)

### Rollback Failure Actions
- Preserve current state (do not force changes)
- Create incident ticket with "baseline-rollback-failed" tag
- Notify on-call engineer via PagerDuty
- Document all attempted remediation steps
