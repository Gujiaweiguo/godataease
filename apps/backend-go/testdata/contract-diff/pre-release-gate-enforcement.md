# Pre-Release Gate Enforcement Policy

## Purpose

This policy defines the mandatory gate enforcement rules that MUST be satisfied before any release can proceed to publish stage. It establishes the blocking conditions, evidence requirements, CI integration points, and exception handling for pre-release gate validation during the Java-to-Go migration.

---

## Mandatory Gate Requirements

### Gate Definition

A pre-release gate is a validation checkpoint that verifies required-gate interfaces pass compatibility checks. Gates are executed as part of the release pipeline and MUST return a passing status for release approval.

### Gate Types

| Gate Type | Scope | Blocking Level | Description |
|-----------|-------|----------------|-------------|
| Compatibility Gate | Required-gate interfaces | Per interface priority | Validates Go implementation matches Java baseline |
| Contract Gate | API contract diff | Critical | Ensures response structure matches expected schema |
| Performance Gate | Response time thresholds | High | Validates latency within acceptable bounds |

### Gate Execution Requirements

Every release MUST execute the following gates:

1. **Compatibility Gate**: All required-gate interfaces defined in `critical-whitelist.yaml` MUST pass
2. **Contract Gate**: No breaking schema changes detected in baseline comparison
3. **Performance Gate**: Response times within defined thresholds (P0: 2x baseline, P1: 3x baseline)

---

## Blocking Conditions

### 未过 Gate 不得发布 (Gate Failure Blocks Release)

The following conditions MUST block release to publish stage:

| Condition | Blocking Behavior | Override Allowed |
|-----------|-------------------|------------------|
| Any P0 interface fails compatibility check | Immediate block | No (waiver required) |
| Any P0 interface not evaluated | Immediate block | No (waiver required) |
| P1 interface fails for > 24 hours | Block after grace period | Tech Lead approval |
| P1 interface not evaluated | Block after grace period | Tech Lead approval |
| Missing gate evidence | Immediate block | No |
| Expired waiver active on blocking interface | Immediate block | No (renewal required) |

### Explicit Blocking Statement

**未满足以下条件禁止发布：**

1. 所有 P0 级别接口的兼容性检查必须通过
2. 所有 P0 级别接口必须被评估（不可跳过）
3. Gate 执行证据必须完整（包含状态、范围、时间戳）
4. 有效的豁免（如有）必须在有效期内

**Release MUST NOT proceed when:**
- Any P0 interface gate returns `failed` status
- Any P0 interface gate returns `not_evaluated` status
- Gate evidence is missing or incomplete
- Active waiver has expired

### Manual Override Restriction

Bypass through manual override MUST be denied unless an approved exception waiver is active:

```
IF gate_status == "failed" AND active_waiver == null:
    RETURN "BLOCKED: Gate failure without approved waiver"

IF gate_status == "failed" AND active_waiver.status == "approved":
    IF active_waiver.expiry > current_time:
        RETURN "ALLOWED: Approved waiver active"
    ELSE:
        RETURN "BLOCKED: Waiver expired, renewal required"
```

---

## Evidence Requirements

### Mandatory Evidence Fields

Gate evidence MUST include the following fields for release approval:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `result_status` | enum | Yes | `passed`, `failed`, `not_evaluated` |
| `evaluated_scope` | array | Yes | List of interface paths that were evaluated |
| `execution_timestamp` | ISO 8601 | Yes | UTC timestamp of gate execution |
| `interface_results` | array | Yes | Per-interface result details |
| `baseline_version` | string | Yes | Version of baseline used for comparison |
| `gate_version` | string | Yes | Version of gate execution engine |

### Evidence Schema

```json
{
  "gate_evidence": {
    "result_status": "passed|failed|not_evaluated",
    "evaluated_scope": [
      "/api/auth/login",
      "/api/dataset/list",
      "/api/chart/data"
    ],
    "execution_timestamp": "2026-02-18T10:30:00Z",
    "baseline_version": "baseline-v1.2",
    "gate_version": "gate-v2.0.0",
    "interface_results": [
      {
        "path": "/api/auth/login",
        "method": "POST",
        "status": "passed",
        "priority": "P0",
        "response_time_ms": 45,
        "diff_summary": null
      }
    ],
    "summary": {
      "total_interfaces": 50,
      "passed": 48,
      "failed": 2,
      "not_evaluated": 0
    }
  }
}
```

### Missing Evidence Handling

When gate evidence is missing or incomplete, the release MUST be blocked:

| Missing Field | Behavior |
|---------------|----------|
| `result_status` | Block - cannot determine gate outcome |
| `evaluated_scope` | Block - cannot verify coverage |
| `execution_timestamp` | Block - cannot verify freshness |
| `interface_results` | Block - cannot identify failures |
| `baseline_version` | Block - cannot verify comparison validity |

**Missing evidence MUST be treated as release blocking.**

No release approval SHALL be granted without complete gate evidence.

---

## CI Integration Points

### Pipeline Checkpoint Mapping

| Pipeline Stage | Gate Check | Action on Failure |
|----------------|------------|-------------------|
| Pre-commit | Contract lint (fast) | Block commit |
| PR Validation | Compatibility Gate (P0 only) | Block merge |
| Build | Performance Gate | Block artifact creation |
| Pre-release | Full Gate Suite | Block publish stage |
| Post-release | Regression Gate | Alert only |

### CI Stage Configuration

```yaml
# .github/workflows/release.yml (example)
stages:
  - name: pre-release-gate
    gates:
      - type: compatibility
        scope: required-gate
        blocking: true
      - type: contract
        scope: required-gate
        blocking: true
      - type: performance
        scope: required-gate
        blocking: false  # Warning only
    on_failure: block_release
    evidence_output: gate-evidence.json
```

### Release Pipeline Flow

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│   Build Stage   │────▶│  Gate Execution  │────▶│  Gate Review    │
└─────────────────┘     └──────────────────┘     └─────────────────┘
                               │                        │
                               ▼                        ▼
                        ┌──────────────┐        ┌──────────────┐
                        │ Any Failed?  │        │ Evidence OK? │
                        └──────────────┘        └──────────────┘
                          │         │             │         │
                         Yes        No           No        Yes
                          │         │             │         │
                          ▼         ▼             ▼         ▼
                    ┌─────────┐ ┌────────┐  ┌─────────┐ ┌────────┐
                    │  BLOCK  │ │ Continue│  │  BLOCK  │ │Approve │
                    │ RELEASE │ │  to     │  │ RELEASE │ │Release │
                    └─────────┘ │ Review  │  └─────────┘ └────────┘
                                └────────┘
```

### Checkpoint Enforcement

| Checkpoint | Enforcement | Evidence Required |
|------------|-------------|-------------------|
| Code Merge | Soft block (can override with approval) | PR gate summary |
| Artifact Build | Hard block (no override) | Build gate log |
| Release Approval | Hard block (no override without waiver) | Full gate evidence |
| Publish | Hard block (no override) | Approved gate evidence |

---

## Exception Handling

### Waiver Governance

Temporary waivers for gate failures follow strict governance:

#### Waiver Request Process

1. **Submit Request**: Create waiver request with justification
2. **Document Impact**: Include risk assessment and mitigation plan
3. **Specify Duration**: Maximum 7 days, explicit expiry required
4. **Obtain Approvals**: Tech Lead + QA Lead signatures required
5. **Activate**: Waiver inactive until all approvals complete

#### Waiver Activation Rules

```
waiver_status transitions:
  requested -> pending_approval -> approved -> active -> expired
                         │              │
                      rejected      cancelled

Activation requirement:
  IF all_required_approvers_signed:
      SET waiver_status = "active"
      SET activated_at = current_time
  ELSE:
      waiver_status REMAINS "pending_approval"
      RELEASE REMAINS BLOCKED
```

#### Waiver Expiry Enforcement

| Condition | Behavior |
|-----------|----------|
| Waiver `expiry_time > current_time` | Waiver valid, release allowed |
| Waiver `expiry_time <= current_time` | Waiver auto-expired, release blocked |
| Expired waiver on file | Requires renewal or gate resolution |

**Expired waivers MUST automatically stop unblocking release.**

### Exception Categories

| Exception Type | Max Duration | Required Approvers | Notes |
|----------------|--------------|-------------------|-------|
| Critical Hotfix | 24 hours | Tech Lead only | Must link to incident |
| Planned Migration | 7 days | Tech Lead + QA Lead | Requires rollback plan |
| Feature Flag Gate | 7 days | Tech Lead + QA Lead | For gradual rollout |
| Performance Exception | 3 days | Tech Lead + QA Lead | Requires optimization plan |

### Waiver Schema

```yaml
waiver:
  id: "waiver-2026-02-18-001"
  interface: "/api/dataset/list"
  reason: "Performance optimization in progress"
  risk_assessment: "Low - feature behind feature flag"
  mitigation: "Feature flag rollback available"
  requested_by: "dataset-team"
  approvers:
    - name: "Tech Lead"
      status: "approved"
      timestamp: "2026-02-18T09:00:00Z"
    - name: "QA Lead"
      status: "pending"
  status: "pending_approval"  # inactive until all approved
  created_at: "2026-02-18T08:00:00Z"
  expires_at: "2026-02-25T08:00:00Z"
```

---

## Rollback Plan

### Gate Enforcement Rollback

If gate enforcement causes critical production issues:

#### Immediate Rollback (P0)

| Step | Action | Owner | Time |
|------|--------|-------|------|
| 1 | Disable blocking gates in CI config | DevOps | 5 min |
| 2 | Notify Tech Lead and QA Lead | DevOps | 5 min |
| 3 | Create incident ticket | DevOps | 10 min |
| 4 | Document root cause | Tech Lead | 1 hour |
| 5 | Implement fix | Engineering | TBD |

#### Configuration Rollback

```yaml
# emergency-disable-gates.yml
gate_enforcement:
  enabled: false
  reason: "incident-2026-02-18-001"
  disabled_by: "devops-team"
  disabled_at: "2026-02-18T10:00:00Z"
  expected_resolution: "2026-02-18T18:00:00Z"
```

#### Re-enablement Criteria

Gates MUST NOT be re-enabled until:

- [ ] Root cause identified and documented
- [ ] Fix implemented and tested
- [ ] Tech Lead + QA Lead approval obtained
- [ ] Staged rollout plan created

### Release Rollback

If a blocked release needs emergency promotion:

1. **Emergency Waiver**: Tech Lead creates emergency waiver with incident link
2. **Time-Bounded**: Maximum 24-hour duration
3. **Monitoring**: Enhanced monitoring required during waiver period
4. **Post-Mortem**: Required within 48 hours of waiver activation

---

## Monitoring and Alerting

### Gate Metrics

| Metric | Threshold | Alert |
|--------|-----------|-------|
| Gate execution time | > 5 minutes | Warning |
| Gate failure rate | > 5% | Warning |
| P0 gate failures | > 0 | Critical |
| Missing evidence | > 0 | Critical |
| Expired waivers | > 0 | Warning |

### Dashboards

- **Gate Status Dashboard**: Real-time gate execution status
- **Waiver Tracking**: Active and pending waivers
- **Release Blocker Summary**: Current blocking conditions

---

## Audit Requirements

### Evidence Retention

| Evidence Type | Retention Period | Storage |
|---------------|------------------|---------|
| Gate execution logs | 90 days | CI artifacts |
| Waiver records | 1 year | Audit database |
| Release decisions | 2 years | Audit database |
| Exception approvals | 2 years | Audit database |

### Compliance Checks

| Check | Frequency | Owner |
|-------|-----------|-------|
| Gate evidence completeness | Per release | CI Pipeline |
| Waiver expiry enforcement | Daily | Automated |
| Orphaned waivers | Weekly | QA Team |
| Policy compliance audit | Quarterly | Tech Lead |

---

## References

- `required-gate-interfaces.md` - Interface scope and classification
- `baseline-policy.md` - Baseline capture and maintenance
- `critical-whitelist.yaml` - Authoritative interface catalog
- `openspec/changes/update-api-compatibility-bridge-with-required-gate-policy/` - Spec requirements

---

*Effective Date: 2026-02-18*
*Last Updated: 2026-02-18*
*Version: 1.0.0*
