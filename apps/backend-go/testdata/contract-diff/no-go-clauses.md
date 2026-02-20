# No-Go Clauses Policy

## Purpose

This policy defines the mandatory No-Go clauses that MUST block release when bypass conditions are detected. It establishes clear bypass detection criteria, the No-Go blocking condition checklist, enforcement rules that any bypass hit triggers blocking, alignment with release gates, and audit trail requirements for bypass attempts.

---

## Bypass Detection Conditions

### Definition of Bypass

A bypass is any attempt to circumvent established release gate requirements without proper authorization. Bypasses include unapproved exceptions, expired waivers, missing evidence, and unauthorized state transitions.

### Bypass Condition Categories

| Category | Condition | Detection Method | Severity |
|----------|-----------|------------------|----------|
| 未审批豁免 (Unapproved Exception) | Exception activated without all required approvals | Approval chain verification | Critical |
| 超期豁免 (Expired Waiver) | Waiver used beyond expiry time | Time-based expiry check | Critical |
| 缺证据发布 (Missing Evidence Release) | Release attempted without gate evidence | Evidence completeness check | Critical |
| 权限绕过 (Authority Bypass) | Unauthorized role acting as approver | Role verification | Critical |
| 状态篡改 (State Tampering) | Manual status manipulation outside workflow | State transition audit | Critical |
| 重复利用 (Reuse Exploit) | Expired/revoked waiver identifier reused | Waiver ID tracking | High |

### Bypass Detection Logic

```python
def detect_bypass(release_context):
    bypasses = []
    
    # 1. 未审批豁免检测 (Unapproved Exception Detection)
    for exception in release_context.exceptions:
        if exception.status in ["active", "approved"]:
            missing_approvers = get_missing_approvers(exception)
            if missing_approvers:
                bypasses.append({
                    "type": "unapproved_exception",
                    "exception_id": exception.id,
                    "missing_approvers": missing_approvers,
                    "severity": "critical"
                })
    
    # 2. 超期豁免检测 (Expired Waiver Detection)
    for waiver in release_context.waivers:
        if waiver.status == "active":
            if current_time() >= waiver.expires_at:
                bypasses.append({
                    "type": "expired_waiver",
                    "waiver_id": waiver.id,
                    "expired_at": waiver.expires_at,
                    "severity": "critical"
                })
    
    # 3. 缺证据发布检测 (Missing Evidence Detection)
    if not has_complete_gate_evidence(release_context):
        bypasses.append({
            "type": "missing_evidence",
            "required_fields": get_missing_evidence_fields(release_context),
            "severity": "critical"
        })
    
    return bypasses
```

### Bypass Detection Examples

#### Example 1: 未审批豁免 (Unapproved Exception)

```yaml
bypass_detected:
  type: "unapproved_exception"
  exception_id: "exc-2026-02-18-001"
  
  detection:
    method: "approval_chain_verification"
    timestamp: "2026-02-18T10:00:00Z"
    
  details:
    required_approvers:
      - role: "tech_lead"
        status: "approved"
      - role: "qa_lead"
        status: "pending"  # Missing approval
    actual_status: "active"  # Should be pending_qa_lead
    
  bypass_attempt:
    attempted_by: "system_user"
    attempted_action: "activate_exception"
    
  blocking_result:
    blocked: true
    reason: "QA Lead approval required but not obtained"
```

#### Example 2: 超期豁免 (Expired Waiver)

```yaml
bypass_detected:
  type: "expired_waiver"
  waiver_id: "waiver-2026-02-15-001"
  
  detection:
    method: "time_based_expiry_check"
    timestamp: "2026-02-18T10:00:00Z"
    
  details:
    activated_at: "2026-02-15T10:00:00Z"
    expires_at: "2026-02-17T10:00:00Z"
    current_time: "2026-02-18T10:00:00Z"
    overdue_duration: "24h"
    
  bypass_attempt:
    attempted_by: "release_pipeline"
    attempted_action: "use_waiver_for_release"
    
  blocking_result:
    blocked: true
    reason: "Waiver expired 24 hours ago"
```

#### Example 3: 缺证据发布 (Missing Evidence Release)

```yaml
bypass_detected:
  type: "missing_evidence"
  release_id: "rel-2026-02-18-001"
  
  detection:
    method: "evidence_completeness_check"
    timestamp: "2026-02-18T10:00:00Z"
    
  details:
    required_fields:
      - field: "result_status"
        status: "present"
      - field: "evaluated_scope"
        status: "missing"  # Missing field
      - field: "execution_timestamp"
        status: "present"
      - field: "interface_results"
        status: "incomplete"  # Partial data
    
  bypass_attempt:
    attempted_by: "release_manager"
    attempted_action: "approve_release"
    
  blocking_result:
    blocked: true
    reason: "Gate evidence incomplete: evaluated_scope missing, interface_results incomplete"
```

---

## No-Go Blocking Conditions Checklist

### Mandatory Blocking Conditions

The following conditions MUST block release to publish stage. ANY single condition being true triggers immediate blocking.

| # | Condition | 中文描述 | Blocking Level | Override |
|---|-----------|----------|----------------|----------|
| 1 | Any P0 interface fails compatibility check | 任意 P0 级别接口兼容性检查失败 | Immediate | None |
| 2 | Any P0 interface not evaluated | 任意 P0 级别接口未被评估 | Immediate | None |
| 3 | Gate evidence missing or incomplete | Gate 证据缺失或不完整 | Immediate | None |
| 4 | Active waiver has expired | 有效豁免已过期 | Immediate | None |
| 5 | Exception activated without all approvals | 豁免在未获得所有审批时激活 | Immediate | None |
| 6 | Unauthorized approver in approval chain | 审批链中存在未授权审批人 | Immediate | None |
| 7 | Manual state transition outside workflow | 工作流外的手动状态转换 | Immediate | None |
| 8 | Expired/revoked waiver identifier reused | 过期/撤销的豁免标识被复用 | Immediate | None |
| 9 | P1 interface fails for > 24 hours (grace period exceeded) | P1 级别接口失败超过 24 小时 | Delayed | Tech Lead |
| 10 | P1 interface not evaluated (grace period exceeded) | P1 级别接口未被评估超过宽限期 | Delayed | Tech Lead |

### No-Go Condition Evaluation Logic

```python
def evaluate_no_go_conditions(release_context):
    blocking_conditions = []
    
    # === CRITICAL: P0 Interface Conditions ===
    
    # Condition 1: P0 interface failure
    p0_failures = [i for i in release_context.interfaces 
                   if i.priority == "P0" and i.gate_status == "failed"]
    if p0_failures:
        blocking_conditions.append({
            "condition_id": "NO-GO-001",
            "description": "P0 interface compatibility check failed",
            "affected_interfaces": [i.path for i in p0_failures],
            "blocking": True,
            "override_allowed": False
        })
    
    # Condition 2: P0 interface not evaluated
    p0_not_evaluated = [i for i in release_context.interfaces 
                        if i.priority == "P0" and i.gate_status == "not_evaluated"]
    if p0_not_evaluated:
        blocking_conditions.append({
            "condition_id": "NO-GO-002",
            "description": "P0 interface not evaluated",
            "affected_interfaces": [i.path for i in p0_not_evaluated],
            "blocking": True,
            "override_allowed": False
        })
    
    # === CRITICAL: Evidence Conditions ===
    
    # Condition 3: Missing or incomplete gate evidence
    missing_fields = validate_gate_evidence(release_context.gate_evidence)
    if missing_fields:
        blocking_conditions.append({
            "condition_id": "NO-GO-003",
            "description": "Gate evidence missing or incomplete",
            "missing_fields": missing_fields,
            "blocking": True,
            "override_allowed": False
        })
    
    # === CRITICAL: Waiver Conditions ===
    
    # Condition 4: Expired waiver
    expired_waivers = [w for w in release_context.active_waivers 
                       if current_time() >= w.expires_at]
    if expired_waivers:
        blocking_conditions.append({
            "condition_id": "NO-GO-004",
            "description": "Active waiver has expired",
            "expired_waivers": [{"id": w.id, "expired_at": w.expires_at} 
                               for w in expired_waivers],
            "blocking": True,
            "override_allowed": False
        })
    
    # === CRITICAL: Exception Conditions ===
    
    # Condition 5: Unapproved exception activation
    unapproved_active = [e for e in release_context.exceptions 
                         if e.status == "active" and not all_approvers_signed(e)]
    if unapproved_active:
        blocking_conditions.append({
            "condition_id": "NO-GO-005",
            "description": "Exception activated without all approvals",
            "unapproved_exceptions": [{"id": e.id, "missing_approvers": get_missing(e)} 
                                     for e in unapproved_active],
            "blocking": True,
            "override_allowed": False
        })
    
    # Condition 6: Unauthorized approver
    unauthorized_approvers = find_unauthorized_approvers(release_context.exceptions)
    if unauthorized_approvers:
        blocking_conditions.append({
            "condition_id": "NO-GO-006",
            "description": "Unauthorized approver in approval chain",
            "unauthorized": unauthorized_approvers,
            "blocking": True,
            "override_allowed": False
        })
    
    # === CRITICAL: State Integrity Conditions ===
    
    # Condition 7: Manual state transition
    invalid_transitions = detect_manual_state_transitions(release_context)
    if invalid_transitions:
        blocking_conditions.append({
            "condition_id": "NO-GO-007",
            "description": "Manual state transition outside workflow",
            "invalid_transitions": invalid_transitions,
            "blocking": True,
            "override_allowed": False
        })
    
    # Condition 8: Waiver ID reuse
    reused_ids = detect_waiver_id_reuse(release_context)
    if reused_ids:
        blocking_conditions.append({
            "condition_id": "NO-GO-008",
            "description": "Expired/revoked waiver identifier reused",
            "reused_ids": reused_ids,
            "blocking": True,
            "override_allowed": False
        })
    
    # === HIGH: P1 Interface Conditions (with grace period) ===
    
    # Condition 9: P1 failure beyond grace period
    p1_failures_beyond_grace = [i for i in release_context.interfaces 
                                if i.priority == "P1" 
                                and i.gate_status == "failed"
                                and i.failure_duration > timedelta(hours=24)]
    if p1_failures_beyond_grace:
        blocking_conditions.append({
            "condition_id": "NO-GO-009",
            "description": "P1 interface failed beyond grace period",
            "affected_interfaces": [{"path": i.path, "duration": str(i.failure_duration)} 
                                   for i in p1_failures_beyond_grace],
            "blocking": True,
            "override_allowed": True,  # Tech Lead approval
            "override_requirements": "Tech Lead approval required"
        })
    
    # Condition 10: P1 not evaluated beyond grace period
    p1_not_evaluated_beyond_grace = [i for i in release_context.interfaces 
                                     if i.priority == "P1" 
                                     and i.gate_status == "not_evaluated"
                                     and i.not_evaluated_duration > timedelta(hours=24)]
    if p1_not_evaluated_beyond_grace:
        blocking_conditions.append({
            "condition_id": "NO-GO-010",
            "description": "P1 interface not evaluated beyond grace period",
            "affected_interfaces": [{"path": i.path, "duration": str(i.not_evaluated_duration)} 
                                   for i in p1_not_evaluated_beyond_grace],
            "blocking": True,
            "override_allowed": True,  # Tech Lead approval
            "override_requirements": "Tech Lead approval required"
        })
    
    return blocking_conditions
```

---

## Enforcement Rules

### Any Bypass Condition Hit Triggers Block

**MANDATORY**: If ANY bypass condition is detected, the release MUST be blocked immediately.

```python
def enforce_no_go_policy(release_context):
    # Step 1: Detect all bypass conditions
    bypasses = detect_bypass(release_context)
    
    # Step 2: Evaluate all No-Go conditions
    blocking_conditions = evaluate_no_go_conditions(release_context)
    
    # Step 3: Any bypass or blocking condition triggers block
    if bypasses or blocking_conditions:
        return {
            "release_allowed": False,
            "blocking_reason": "no_go_condition_detected",
            "bypasses_detected": bypasses,
            "blocking_conditions": blocking_conditions,
            "action_required": "Resolve all blocking conditions before release"
        }
    
    # Step 4: No blocking conditions - release allowed
    return {
        "release_allowed": True,
        "blocking_reason": None,
        "bypasses_detected": [],
        "blocking_conditions": []
    }
```

### Blocking Enforcement Matrix

| Bypass Type | Blocking Behavior | Notification | Resolution Path |
|-------------|-------------------|--------------|-----------------|
| Unapproved Exception | Immediate block | Tech Lead, QA Lead | Complete approval chain |
| Expired Waiver | Immediate block | Waiver owner | Renew waiver or fix gate |
| Missing Evidence | Immediate block | Release Manager | Generate complete evidence |
| Authority Bypass | Immediate block | Tech Lead, Security | Verify approver authority |
| State Tampering | Immediate block | Tech Lead, Security | Audit and correct state |
| Reuse Exploit | Immediate block | Tech Lead, Security | Generate new waiver ID |

### Hard Block vs Soft Block

| Block Type | Conditions | Override | Time Limit |
|------------|------------|----------|------------|
| **Hard Block** | P0 failures, unapproved exceptions, expired waivers, missing evidence | None | Until resolved |
| **Soft Block** | P1 failures beyond grace period | Tech Lead approval | 24h approval window |

### Enforcement Flow

```
┌─────────────────────────┐
│  Release Request        │
└───────────┬─────────────┘
            │
            ▼
┌─────────────────────────┐
│  Bypass Detection       │
│  - Unapproved exception │
│  - Expired waiver       │
│  - Missing evidence     │
│  - Authority bypass     │
│  - State tampering      │
│  - Reuse exploit        │
└───────────┬─────────────┘
            │
            ▼
┌─────────────────────────┐     ┌──────────────────┐
│  Any Bypass Detected?   │────▶│  HARD BLOCK      │
└───────────┬─────────────┘ Yes │  Log & Notify    │
            │ No                 └──────────────────┘
            ▼
┌─────────────────────────┐
│  No-Go Condition Check  │
│  - P0 interface status  │
│  - Evidence completeness│
│  - Waiver validity      │
│  - Approval completeness│
└───────────┬─────────────┘
            │
            ▼
┌─────────────────────────┐     ┌──────────────────┐
│  Any Condition Met?     │────▶│  BLOCK RELEASE   │
└───────────┬─────────────┘ Yes │  (Hard or Soft)  │
            │ No                 └──────────────────┘
            ▼
┌─────────────────────────┐
│  Release Allowed        │
│  Log & Proceed          │
└─────────────────────────┘
```

---

## Alignment with Release Gates

### No-Go Clauses Match Gate Behavior

No-Go clauses MUST be consistent with pre-release gate enforcement behavior. The following table shows the alignment:

| Pre-Release Gate Rule | No-Go Clause | Consistency Check |
|-----------------------|--------------|-------------------|
| P0 interface failure blocks release | NO-GO-001: P0 interface compatibility check failed | ✓ Aligned |
| P0 interface not evaluated blocks release | NO-GO-002: P0 interface not evaluated | ✓ Aligned |
| Missing gate evidence blocks release | NO-GO-003: Gate evidence missing or incomplete | ✓ Aligned |
| Expired waiver does not unblock | NO-GO-004: Active waiver has expired | ✓ Aligned |
| Exception must be approved before activation | NO-GO-005: Exception activated without all approvals | ✓ Aligned |
| Manual override denied without approved waiver | NO-GO-007: Manual state transition outside workflow | ✓ Aligned |
| P1 grace period before blocking | NO-GO-009/010: 24-hour grace for P1 conditions | ✓ Aligned |

### Consistency Validation

The system MUST validate that No-Go clauses remain consistent with release gate enforcement:

```python
def validate_no_go_gate_alignment():
    inconsistencies = []
    
    # Check P0 failure alignment
    gate_blocks_p0_failure = GATE_CONFIG["p0_failure_blocks"]
    no_go_blocks_p0_failure = NO_GO_CLAUSES["NO-GO-001"]["blocking"]
    if gate_blocks_p0_failure != no_go_blocks_p0_failure:
        inconsistencies.append({
            "gate_rule": "p0_failure_blocks",
            "no_go_clause": "NO-GO-001",
            "expected": gate_blocks_p0_failure,
            "actual": no_go_blocks_p0_failure
        })
    
    # Check P0 not evaluated alignment
    gate_blocks_p0_skip = GATE_CONFIG["p0_skip_blocks"]
    no_go_blocks_p0_skip = NO_GO_CLAUSES["NO-GO-002"]["blocking"]
    if gate_blocks_p0_skip != no_go_blocks_p0_skip:
        inconsistencies.append({
            "gate_rule": "p0_skip_blocks",
            "no_go_clause": "NO-GO-002",
            "expected": gate_blocks_p0_skip,
            "actual": no_go_blocks_p0_skip
        })
    
    # Check missing evidence alignment
    gate_requires_evidence = GATE_CONFIG["evidence_required"]
    no_go_blocks_missing_evidence = NO_GO_CLAUSES["NO-GO-003"]["blocking"]
    if gate_requires_evidence != no_go_blocks_missing_evidence:
        inconsistencies.append({
            "gate_rule": "evidence_required",
            "no_go_clause": "NO-GO-003",
            "expected": gate_requires_evidence,
            "actual": no_go_blocks_missing_evidence
        })
    
    if inconsistencies:
        raise PolicyAlignmentError(
            f"No-Go clauses not aligned with release gates: {inconsistencies}"
        )
    
    return True
```

### Cross-Reference Mapping

```yaml
alignment_mapping:
  gate_policy: "pre-release-gate-enforcement.md"
  no_go_policy: "no-go-clauses.md"
  
  rules:
    - gate_section: "Blocking Conditions"
      gate_rule: "P0 interface fails compatibility check"
      no_go_clause: "NO-GO-001"
      alignment_status: "consistent"
      
    - gate_section: "Blocking Conditions"
      gate_rule: "P0 interface not evaluated"
      no_go_clause: "NO-GO-002"
      alignment_status: "consistent"
      
    - gate_section: "Evidence Requirements"
      gate_rule: "Missing gate evidence blocks release"
      no_go_clause: "NO-GO-003"
      alignment_status: "consistent"
      
    - gate_section: "Exception Handling"
      gate_rule: "Waiver must be approved before activation"
      no_go_clause: "NO-GO-005"
      alignment_status: "consistent"
      
    - waiver_policy: "waiver-expiry-policy.md"
      waiver_rule: "Expired waiver does not unblock release"
      no_go_clause: "NO-GO-004"
      alignment_status: "consistent"
      
    - exception_policy: "exception-approval-policy.md"
      exception_rule: "Exception inactive until all approvals"
      no_go_clause: "NO-GO-005"
      alignment_status: "consistent"
```

---

## Audit Trail for Bypass Attempts

### Mandatory Logging Requirements

All bypass attempts MUST be logged with the following information:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `attempt_id` | string | Yes | Unique identifier for the bypass attempt |
| `timestamp` | ISO 8601 | Yes | When the bypass attempt occurred |
| `bypass_type` | enum | Yes | Type of bypass detected |
| `attempted_by` | string | Yes | User or system that attempted the bypass |
| `target_resource` | string | Yes | Release, exception, or waiver being targeted |
| `detection_method` | string | Yes | How the bypass was detected |
| `blocking_result` | enum | Yes | `blocked` or `allowed` (with justification) |
| `resolution_action` | string | No | Action taken to resolve if blocked |

### Audit Log Schema

```yaml
bypass_audit_log:
  attempt_id: "bypass-2026-02-18-001"
  timestamp: "2026-02-18T10:30:00Z"
  
  bypass_type: "unapproved_exception"
  
  attempted_by:
    type: "user"  # user | system | pipeline
    identifier: "release-manager@example.com"
    ip_address: "192.168.1.100"
    session_id: "sess-abc123"
  
  target_resource:
    type: "release"
    id: "rel-2026-02-18-001"
    exception_id: "exc-2026-02-18-001"
  
  detection:
    method: "approval_chain_verification"
    triggered_by: "release_pipeline_check"
    detection_timestamp: "2026-02-18T10:30:00.123Z"
  
  details:
    expected_state: "pending_qa_lead"
    actual_state: "active"
    missing_approvals:
      - role: "qa_lead"
        status: "pending"
  
  blocking_result:
    status: "blocked"
    reason: "Exception activated without QA Lead approval"
    blocked_at: "2026-02-18T10:30:00.456Z"
    blocked_by: "no_go_policy_engine"
  
  notifications_sent:
    - channel: "#release-alerts"
      message: "Bypass attempt blocked: Unapproved exception exc-2026-02-18-001"
      timestamp: "2026-02-18T10:30:01Z"
    - channel: "tech-lead@example.com"
      message: "Bypass attempt detected and blocked"
      timestamp: "2026-02-18T10:30:01Z"
  
  resolution_action:
    required: true
    actions:
      - "Complete QA Lead approval"
      - "Verify exception state transition"
    status: "pending"
    assigned_to: "release-manager@example.com"
```

### Audit Event Categories

| Event Category | Description | Retention | Alert Level |
|----------------|-------------|-----------|-------------|
| `bypass_detected` | Bypass attempt identified and blocked | 2 years | Critical |
| `bypass_blocked` | Bypass successfully blocked | 2 years | Warning |
| `bypass_suspected` | Potential bypass under investigation | 2 years | Info |
| `policy_violation` | General policy violation detected | 2 years | Critical |
| `unauthorized_access` | Unauthorized access to approval system | 2 years | Critical |

### Audit Trail Query Interface

```yaml
audit_query_schema:
  supported_filters:
    - field: "timestamp"
      operators: ["equals", "range", "before", "after"]
    - field: "bypass_type"
      operators: ["equals", "in"]
    - field: "attempted_by"
      operators: ["equals", "contains"]
    - field: "blocking_result"
      operators: ["equals", "in"]
    - field: "target_resource.type"
      operators: ["equals", "in"]
  
  example_queries:
    - description: "All bypass attempts in last 24 hours"
      filter:
        timestamp:
          operator: "after"
          value: "-24h"
    
    - description: "All blocked unapproved exception attempts"
      filter:
        bypass_type: "unapproved_exception"
        blocking_result: "blocked"
    
    - description: "All attempts by specific user"
      filter:
        attempted_by.identifier: "user@example.com"
```

### Compliance Reporting

| Report | Frequency | Recipients | Content |
|--------|-----------|------------|---------|
| Daily Bypass Summary | Daily | Tech Lead, Security | All bypass attempts in past 24h |
| Weekly Compliance Report | Weekly | Tech Lead, QA Lead, Security | Bypass trends, policy adherence |
| Monthly Security Review | Monthly | Engineering Director | Full audit review, incident summary |
| Quarterly Audit | Quarterly | Compliance Team | Complete audit trail export |

### Tamper-Evident Logging

All audit logs MUST be tamper-evident:

```yaml
audit_integrity:
  hash_chain:
    algorithm: "SHA-256"
    method: "each_log_entry_hashed_with_previous"
  
  signature:
    algorithm: "RS256"
    key_rotation: "quarterly"
  
  storage:
    type: "append_only"
    replication: 3
    retention_years: 2
  
  verification:
    frequency: "daily"
    alert_on_tamper: true
    tamper_response: "immediate_security_alert"
```

---

## Integration Points

### CI/CD Pipeline Integration

```yaml
# Pipeline checkpoint for No-Go evaluation
no_go_checkpoint:
  stage: "pre-release"
  
  evaluation_sequence:
    - step: "bypass_detection"
      blocking: true
      timeout: "30s"
    
    - step: "no_go_condition_check"
      blocking: true
      timeout: "60s"
    
    - step: "gate_alignment_validation"
      blocking: true
      timeout: "30s"
    
    - step: "audit_log_verification"
      blocking: false
      timeout: "30s"
  
  on_failure:
    action: "block_release"
    notify:
      - "#release-alerts"
      - "tech-lead@example.com"
    generate_report: true
  
  outputs:
    - "no-go-evaluation-report.json"
    - "bypass-audit-log.yaml"
```

### Exception Management Integration

```python
def integrate_with_exception_management():
    # Before any exception state change
    def on_exception_state_change(exception, new_state):
        # Check if this would trigger a bypass
        if new_state == "active":
            if not all_approvers_signed(exception):
                log_bypass_attempt(
                    type="unapproved_exception",
                    target=exception.id,
                    blocked=True
                )
                raise BypassBlockedException(
                    "Cannot activate exception without all approvals"
                )
        
        # Log all state transitions
        audit_log.record({
            "event": "exception_state_change",
            "exception_id": exception.id,
            "from_state": exception.status,
            "to_state": new_state,
            "timestamp": current_time()
        })
```

### Waiver Management Integration

```python
def integrate_with_waiver_management():
    # Before using waiver for release
    def on_waiver_use(waiver, release_context):
        # Check if waiver is expired
        if current_time() >= waiver.expires_at:
            log_bypass_attempt(
                type="expired_waiver",
                target=waiver.id,
                blocked=True,
                details={"expired_at": waiver.expires_at}
            )
            raise BypassBlockedException(
                "Cannot use expired waiver for release"
            )
        
        # Verify waiver ID has not been reused
        if is_reused_waiver_id(waiver.id):
            log_bypass_attempt(
                type="reuse_exploit",
                target=waiver.id,
                blocked=True
            )
            raise BypassBlockedException(
                "Waiver ID reuse detected"
            )
```

---

## Monitoring and Alerting

### Real-Time Monitoring

| Metric | Alert Threshold | Response |
|--------|-----------------|----------|
| Bypass attempts per hour | > 3 | Critical alert to Security |
| Blocked releases due to No-Go | Any | Info notification |
| Failed bypass attempts | Any | Warning to Tech Lead |
| Successful bypasses (should be 0) | > 0 | Critical alert + immediate investigation |
| Audit log integrity failures | > 0 | Critical alert + tamper investigation |

### Dashboard Requirements

- **Bypass Attempt Dashboard**: Real-time view of bypass attempts
- **No-Go Condition Status**: Current state of all No-Go conditions
- **Audit Trail Viewer**: Searchable interface for audit logs
- **Policy Alignment Status**: Consistency check between No-Go and gates

---

## References

- `pre-release-gate-enforcement.md` - Gate blocking and waiver governance
- `exception-approval-policy.md` - Exception categories and approval workflow
- `waiver-expiry-policy.md` - Waiver duration limits and expiry enforcement
- `required-gate-interfaces.md` - Interface scope and ownership
- `critical-whitelist.yaml` - Authoritative interface catalog

---

*Effective Date: 2026-02-18*
*Last Updated: 2026-02-18*
*Version: 1.0.0*
