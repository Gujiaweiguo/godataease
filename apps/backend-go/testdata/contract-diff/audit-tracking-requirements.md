# Audit Tracking Requirements

## Purpose

This policy defines the mandatory audit requirements for exception approvals and gate failures. It establishes the minimum audit fields, evidence archival rules, query capabilities, and traceability requirements that ensure any exception can be traced back to its complete approval chain and effectiveness records, supporting post-release accountability and review.

---

## Mandatory Audit Fields

### Core Audit Fields

Every exception and waiver record MUST include the following mandatory fields for audit compliance:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `interface_path` | string | Yes | API path subject to exception (e.g., `/api/dataset/list`) |
| `interface_method` | string | Yes | HTTP method of the interface (GET, POST, PUT, DELETE) |
| `baseline_version` | string | Yes | Version of baseline used for comparison at time of exception |
| `gate_version` | string | Yes | Version of gate execution engine used |
| `approver_chain` | array | Yes | Ordered list of approvers with timestamps and rationale |
| `approval_status` | enum | Yes | Current status: `requested`, `pending`, `approved`, `active`, `rejected`, `expired`, `revoked` |
| `business_justification` | text | Yes | Documented reason for exception request |
| `requested_duration` | duration | Yes | Requested exception duration in hours/days |
| `actual_duration` | duration | No | Actual duration used (populated on expiry) |
| `valid_from` | timestamp | Yes | When exception becomes/became active (ISO 8601) |
| `valid_until` | timestamp | Yes | When exception expires/expired (ISO 8601) |
| `created_at` | timestamp | Yes | Request submission timestamp |
| `created_by` | string | Yes | Identity of request submitter |

### Approver Chain Fields

Each entry in the `approver_chain` MUST include:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `role` | enum | Yes | Approver role: `tech_lead`, `qa_lead`, `release_manager` |
| `actor` | string | Yes | Identity of approver (email or username) |
| `status` | enum | Yes | Approval status: `pending`, `approved`, `rejected` |
| `timestamp` | timestamp | Yes | When action was taken |
| `rationale` | text | Yes | Approval/rejection reason documented by approver |
| `delegated_from` | string | No | Original role if approval was delegated |

### Audit Field Schema

```yaml
exception_audit_record:
  # Identification
  exception_id: "exc-2026-02-18-001"
  request_id: "req-2026-02-18-042"
  
  # Interface and Version
  interface:
    path: "/api/dataset/list"
    method: "GET"
    priority: "P0"
  
  version_context:
    baseline_version: "baseline-v1.2.0"
    gate_version: "gate-v2.0.0"
    comparison_snapshot: "snapshot-2026-02-18-001"
  
  # Approval Chain
  approver_chain:
    - role: "tech_lead"
      actor: "tech-lead@example.com"
      status: "approved"
      timestamp: "2026-02-18T08:30:00Z"
      rationale: "Risk acceptable, mitigation plan adequate"
      delegated_from: null
    
    - role: "qa_lead"
      actor: "qa-lead@example.com"
      status: "approved"
      timestamp: "2026-02-18T09:00:00Z"
      rationale: "Test coverage verified, quality risk accepted"
      delegated_from: null
  
  # Justification
  business_justification: |
    Dataset list API returns partial data during query optimization.
    Feature launch depends on this API; delaying release blocks critical feature.
  
  # Validity Period
  validity:
    requested_duration: "72h"
    actual_duration: "48h"
    valid_from: "2026-02-18T09:00:00Z"
    valid_until: "2026-02-21T09:00:00Z"
  
  # Status
  approval_status: "active"
  created_at: "2026-02-18T08:00:00Z"
  created_by: "dataset-team"
  
  # Linked Records
  linked_records:
    gate_failure_id: "gf-2026-02-18-042"
    incident_id: null
    renewal_of: null
    renewed_by: null
```

---

## Evidence Archival Requirements

### What to Archive

All evidence related to exception approvals MUST be archived for compliance and traceability:

| Evidence Type | Description | Retention Period |
|---------------|-------------|------------------|
| Exception Request | Original request with all fields | 2 years |
| Approval Records | All approver chain entries with rationale | 2 years |
| Gate Failure Evidence | Result status, evaluated scope, timestamp | 2 years |
| Baseline Comparison | Diff report that triggered the exception | 2 years |
| Mitigation Evidence | Monitoring configs, escalation plans | 2 years |
| Expiry/Revocation Records | End-of-life documentation | 2 years |
| Renewal Chain | All renewal requests linked to original | 2 years |

### Where to Archive

| Storage Location | Content | Access Level |
|------------------|---------|--------------|
| Audit Database | Structured exception records, approval chain | Read-only after creation |
| Object Storage | Full evidence packages, diff reports | Immutable, versioned |
| Log Aggregation | Real-time audit events, state transitions | Searchable, time-indexed |
| Backup Archive | Point-in-time snapshots | Disaster recovery only |

### Retention Enforcement

```
retention_policy:
  default_retention: "2 years"
  extended_retention_triggers:
    - incident_linked: "+1 year"
    - regulatory_hold: "indefinite"
    - legal_hold: "indefinite"
  
  deletion_rules:
    - condition: "retention_elapsed AND no_holds"
      action: "soft_delete"
      grace_period: "30 days"
    - condition: "grace_period_elapsed"
      action: "permanent_delete"
```

### Archive Schema

```yaml
evidence_archive:
  archive_id: "arch-2026-02-18-001"
  exception_id: "exc-2026-02-18-001"
  
  archived_at: "2026-02-18T09:00:00Z"
  archived_by: "system"
  
  components:
    - type: "exception_request"
      location: "s3://audit-bucket/exceptions/2026/02/exc-2026-02-18-001/request.json"
      checksum: "sha256:abc123..."
      size_bytes: 4096
    
    - type: "approval_chain"
      location: "s3://audit-bucket/exceptions/2026/02/exc-2026-02-18-001/approvals.json"
      checksum: "sha256:def456..."
      size_bytes: 2048
    
    - type: "gate_failure_evidence"
      location: "s3://audit-bucket/exceptions/2026/02/exc-2026-02-18-001/gate-failure.json"
      checksum: "sha256:ghi789..."
      size_bytes: 8192
    
    - type: "baseline_diff"
      location: "s3://audit-bucket/exceptions/2026/02/exc-2026-02-18-001/diff-report.html"
      checksum: "sha256:jkl012..."
      size_bytes: 32768
  
  retention:
    expires_at: "2028-02-18T09:00:00Z"
    holds: []
```

---

## Query Requirements

### Queryable Dimensions

Audit records MUST support querying by the following dimensions:

| Dimension | Query Type | Example |
|-----------|------------|---------|
| Interface Path | Exact match, wildcard | `/api/dataset/*` |
| Interface Method | Exact match | `GET`, `POST` |
| Approval Status | Exact match, list | `active`, `expired` |
| Approver | Exact match | `tech-lead@example.com` |
| Date Range | From/To | `2026-02-01` to `2026-02-28` |
| Baseline Version | Exact match | `baseline-v1.2.0` |
| Category | Exact match, list | `performance_exception` |
| Created By | Exact match | `dataset-team` |

### Query Access Levels

| Role | Query Access | Restrictions |
|------|--------------|--------------|
| API Owner | Own team's exceptions | Cannot view other teams |
| Tech Lead | All exceptions (read) | Full access |
| QA Lead | All exceptions (read) | Full access |
| Release Manager | All exceptions (read/write) | Full access |
| Auditor | All exceptions (read-only) | No modification access |
| Compliance Officer | All exceptions (read-only) | Export allowed |

### Query API Specification

```yaml
query_api:
  endpoint: "/api/v1/audit/exceptions/query"
  method: "POST"
  
  request_schema:
    filters:
      interface_path:
        type: "string | array"
        description: "Exact match or wildcard pattern"
      
      interface_method:
        type: "string | array"
        enum: ["GET", "POST", "PUT", "DELETE", "PATCH"]
      
      approval_status:
        type: "string | array"
        enum: ["requested", "pending", "approved", "active", "rejected", "expired", "revoked"]
      
      approver:
        type: "string"
        description: "Filter by approver identity"
      
      date_range:
        from: "ISO 8601 timestamp"
        to: "ISO 8601 timestamp"
      
      baseline_version:
        type: "string"
      
      category:
        type: "string | array"
        enum: ["critical_hotfix", "planned_migration", "feature_flag_gate", 
               "performance_exception", "dependency_delay", "test_environment"]
    
    pagination:
      page: "integer (default: 1)"
      page_size: "integer (default: 50, max: 200)"
    
    sort:
      field: "string (default: created_at)"
      order: "asc | desc (default: desc)"
    
    include:
      - "approval_chain"
      - "linked_records"
      - "evidence_references"
  
  response_schema:
    total_count: "integer"
    page: "integer"
    page_size: "integer"
    results: "array[exception_audit_record]"
```

### Example Queries

```yaml
# Query 1: All active exceptions for a specific interface
query:
  filters:
    interface_path: "/api/dataset/list"
    approval_status: "active"

# Query 2: Exceptions approved by a specific approver in date range
query:
  filters:
    approver: "tech-lead@example.com"
    date_range:
      from: "2026-02-01T00:00:00Z"
      to: "2026-02-28T23:59:59Z"

# Query 3: All expired performance exceptions
query:
  filters:
    category: "performance_exception"
    approval_status: "expired"
  sort:
    field: "valid_until"
    order: "desc"
```

---

## Traceability Matrix

### Exception to Approval Chain Traceability

Every exception MUST be traceable to its complete approval chain:

```
┌─────────────────┐
│   Exception     │
│   Record        │
│  (exc-001)      │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Approval Chain │
│  ┌─────────────┐│
│  │ Tech Lead   ││──▶ rationale + timestamp
│  │ (approved)  ││
│  └─────────────┘│
│  ┌─────────────┐│
│  │ QA Lead     ││──▶ rationale + timestamp
│  │ (approved)  ││
│  └─────────────┘│
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Effectiveness  │
│  Records        │
│  ┌─────────────┐│
│  │ Activated   ││──▶ valid_from timestamp
│  └─────────────┘│
│  ┌─────────────┐│
│  │ Expired     ││──▶ valid_until timestamp
│  └─────────────┘│
└─────────────────┘
```

### Traceability Requirements

| From | To | Required Link | Purpose |
|------|-----|---------------|---------|
| Exception | Gate Failure | `gate_failure_id` | Identify triggering failure |
| Exception | Approval Chain | `approver_chain[]` | Verify complete approval |
| Exception | Baseline | `baseline_version` | Context of exception |
| Exception | Renewal Chain | `renewal_of`, `renewed_by` | Track exception lineage |
| Exception | Incident | `incident_id` | Link to related incidents |
| Approval | Approver | `actor` | Accountability |
| Approval | Rationale | `rationale` | Decision justification |

### Traceability Query Support

```python
def trace_exception_to_approval_chain(exception_id):
    """
    Returns complete traceability from exception to all approvals.
    """
    exception = get_exception(exception_id)
    
    return {
        "exception_id": exception.id,
        "interface": exception.interface,
        "status": exception.approval_status,
        
        "approval_trace": [
            {
                "sequence": i + 1,
                "role": approval.role,
                "actor": approval.actor,
                "status": approval.status,
                "timestamp": approval.timestamp,
                "rationale": approval.rationale,
                "delegated_from": approval.delegated_from
            }
            for i, approval in enumerate(exception.approver_chain)
        ],
        
        "effectiveness_trace": {
            "valid_from": exception.valid_from,
            "valid_until": exception.valid_until,
            "actual_duration": exception.actual_duration,
            "current_status": compute_current_status(exception)
        },
        
        "linked_records": {
            "gate_failure": get_gate_failure(exception.gate_failure_id),
            "baseline": get_baseline(exception.baseline_version),
            "renewal_chain": get_renewal_chain(exception_id)
        }
    }
```

### Traceability Matrix Schema

```yaml
traceability_matrix:
  version: "1.0.0"
  
  exception_to_approvals:
    required_links:
      - source: "exception.approver_chain"
        target: "approval_record"
        cardinality: "1:N"
        validation: "all_required_approvers_present"
    
    integrity_checks:
      - check: "approval_sequence_valid"
        rule: "approvals follow category-defined order"
      - check: "no_missing_approvers"
        rule: "all category-required approvers present"
      - check: "timestamps_consistent"
        rule: "approval timestamps >= request timestamp"
  
  exception_to_effectiveness:
    required_links:
      - source: "exception.valid_from"
        target: "activation_record"
        cardinality: "1:1"
      - source: "exception.valid_until"
        target: "expiry_record"
        cardinality: "1:1"
    
    integrity_checks:
      - check: "valid_from_after_approval"
        rule: "valid_from >= last_approval_timestamp"
      - check: "valid_until_after_valid_from"
        rule: "valid_until > valid_from"
      - check: "duration_within_limits"
        rule: "actual_duration <= category_max_duration"
```

---

## Post-Release Review Support

### Review Data Requirements

Audit records MUST support post-release review and accountability with the following data:

| Review Type | Required Data | Purpose |
|-------------|---------------|---------|
| Incident Post-Mortem | Exception + approval chain + linked incident | Root cause analysis |
| Compliance Audit | All exception records in period + evidence | Regulatory compliance |
| Performance Review | Exceptions by category, duration, resolution time | Process improvement |
| Accountability Review | Approver decisions, rationale, outcomes | Decision quality |

### Post-Release Review Queries

```yaml
# Incident-linked exceptions for post-mortem
review_query:
  name: "incident_post_mortem"
  filters:
    incident_id: "INC-2026-02-18-001"
  include:
    - "full_approval_chain"
    - "gate_failure_evidence"
    - "mitigation_plan"
    - "outcome_assessment"

# Approver decision quality review
review_query:
  name: "approver_accountability"
  filters:
    approver: "tech-lead@example.com"
    date_range:
      from: "2026-01-01T00:00:00Z"
      to: "2026-03-31T23:59:59Z"
  include:
    - "approval_decisions"
    - "rationale_texts"
    - "exception_outcomes"  # Did exception lead to incident?
```

### Review Output Format

```yaml
post_release_review:
  review_id: "review-2026-02-20-001"
  review_type: "incident_post_mortem"
  
  exception_summary:
    exception_id: "exc-2026-02-18-001"
    interface: "/api/dataset/list"
    category: "performance_exception"
    duration: "48h"
  
  approval_chain_review:
    - role: "tech_lead"
      actor: "tech-lead@example.com"
      rationale: "Risk acceptable, mitigation plan adequate"
      decision_quality: "appropriate"  # appropriate | questionable | problematic
      notes: "Correctly assessed low risk for read-only API"
    
    - role: "qa_lead"
      actor: "qa-lead@example.com"
      rationale: "Test coverage verified, quality risk accepted"
      decision_quality: "appropriate"
      notes: "Verified test coverage before approval"
  
  outcome_assessment:
    led_to_incident: false
    exception_justified: true
    resolution_timely: true
    process_followed: true
  
  lessons_learned:
    - "Performance exception properly scoped to read-only operation"
    - "Mitigation plan with monitoring thresholds was effective"
  
  recommendations:
    - "Consider adding automated test coverage check as pre-condition"
```

### Accountability Report Schema

```yaml
accountability_report:
  report_id: "acc-2026-Q1-001"
  reporting_period:
    from: "2026-01-01T00:00:00Z"
    to: "2026-03-31T23:59:59Z"
  
  summary:
    total_exceptions: 45
    by_status:
      active: 5
      expired: 35
      revoked: 3
      rejected: 2
    by_category:
      performance_exception: 20
      planned_migration: 15
      feature_flag_gate: 10
  
  approver_statistics:
    - approver: "tech-lead@example.com"
      role: "tech_lead"
      decisions:
        approved: 40
        rejected: 5
      average_decision_time: "2.5 hours"
      exceptions_led_to_incident: 2
      incident_rate: "5%"
    
    - approver: "qa-lead@example.com"
      role: "qa_lead"
      decisions:
        approved: 35
        rejected: 10
      average_decision_time: "3.2 hours"
      exceptions_led_to_incident: 1
      incident_rate: "2.9%"
  
  process_metrics:
    average_approval_cycle_time: "4.8 hours"
    average_exception_duration: "3.2 days"
    renewal_rate: "15%"
    compliance_rate: "100%"  # All had required approvals
  
  recommendations:
    - "Tech Lead decision time within SLA"
    - "Consider additional QA review for performance exceptions"
    - "Renewal rate suggests need for longer initial durations in some cases"
```

---

## Audit Event Logging

### Mandatory Audit Events

All state changes and significant actions MUST be logged:

| Event Type | Trigger | Required Fields |
|------------|---------|-----------------|
| `exception_requested` | New request submitted | exception_id, interface, requested_by, timestamp |
| `approval_granted` | Approver approves | exception_id, approver, role, rationale, timestamp |
| `approval_rejected` | Approver rejects | exception_id, approver, role, rationale, timestamp |
| `exception_activated` | All approvals complete | exception_id, valid_from, valid_until, timestamp |
| `exception_expired` | Duration elapsed | exception_id, expiry_reason, timestamp |
| `exception_revoked` | Manual revocation | exception_id, revoked_by, reason, timestamp |
| `renewal_requested` | Renewal submitted | original_id, new_request_id, timestamp |
| `evidence_archived` | Archive created | exception_id, archive_location, timestamp |

### Audit Log Schema

```yaml
audit_log_entry:
  event_id: "evt-2026-02-18-001"
  event_type: "approval_granted"
  timestamp: "2026-02-18T08:30:00Z"
  
  context:
    exception_id: "exc-2026-02-18-001"
    interface_path: "/api/dataset/list"
    interface_method: "GET"
  
  actor:
    identity: "tech-lead@example.com"
    role: "tech_lead"
    ip_address: "10.0.1.100"
  
  action:
    type: "approve"
    rationale: "Risk acceptable, mitigation plan adequate"
    previous_status: "pending_tech_lead"
    new_status: "pending_qa_lead"
  
  metadata:
    user_agent: "Mozilla/5.0..."
    session_id: "sess-abc123"
    request_id: "req-xyz789"
```

---

## Compliance and Enforcement

### Compliance Checks

| Check | Frequency | Enforcement |
|-------|-----------|-------------|
| All required audit fields present | Per record creation | Block creation if missing |
| Approver chain complete | Per activation | Block activation if incomplete |
| Evidence archived | Per expiry | Alert if not archived within 24h |
| Retention policy enforced | Daily | Auto-delete per policy |
| Query access controlled | Per query | Deny unauthorized access |

### Non-Compliance Handling

```yaml
non_compliance_actions:
  missing_audit_field:
    severity: "critical"
    action: "block_record_creation"
    remediation: "require all mandatory fields"
  
  incomplete_approval_chain:
    severity: "critical"
    action: "block_exception_activation"
    remediation: "require all category approvers"
  
  unauthorized_query_access:
    severity: "high"
    action: "deny_access_log_attempt"
    remediation: "review_access_policy"
  
  retention_violation:
    severity: "medium"
    action: "alert_compliance_team"
    remediation: "extend_retention_or_obtain_waiver"
```

---

## References

- `exception-approval-policy.md` - Exception categories and approval workflow
- `waiver-expiry-policy.md` - Waiver expiry and renewal rules
- `pre-release-gate-enforcement.md` - Gate blocking and evidence requirements
- `required-gate-interfaces.md` - Interface scope and ownership
- `critical-whitelist.yaml` - Authoritative interface catalog

---

*Effective Date: 2026-02-18*
*Last Updated: 2026-02-18*
*Version: 1.0.0*
