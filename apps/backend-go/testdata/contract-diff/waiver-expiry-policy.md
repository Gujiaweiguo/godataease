# Waiver Expiry Policy

## Purpose

This policy defines the mandatory expiry rules for gate failure waivers. It establishes duration limits by exception category, automatic expiry triggers, renewal requirements, and enforcement mechanisms that ensure expired waivers cannot continue to unblock releases.

---

## Duration Limits by Category

### Maximum Waiver Duration

Each exception category has a fixed maximum duration. Waivers MUST NOT exceed these limits:

| Category | Max Duration | Extension Allowed | Rationale |
|----------|--------------|-------------------|-----------|
| Critical Hotfix | 24 hours | No | Emergency only, requires immediate resolution |
| Planned Migration | 7 days | Yes (1x) | Known window, limited scope |
| Feature Flag Gate | 7 days | Yes (1x) | Gradual rollout, monitoring required |
| Performance Exception | 3 days | Yes (1x) | Active optimization expected |
| Dependency Delay | 5 days | Yes (1x) | External factor, limited control |
| Test Environment | 3 days | No | Infrastructure issue, needs fix |

### Duration Calculation

```
waiver_duration = min(requested_duration, category_max_duration)

activation_time = approval_complete_timestamp
expiry_time = activation_time + waiver_duration
```

### Duration Enforcement

```
IF requested_duration > category_max_duration:
    REJECT request with error "Duration exceeds category maximum"
    REQUIRE resubmission with valid duration
```

---

## Automatic Expiry Rules

### Time-Based Expiry

All waivers MUST automatically expire when their duration elapses:

| Condition | Expiry Behavior |
|-----------|-----------------|
| `current_time >= expiry_time` | Waiver status -> `expired` |
| `current_time < expiry_time` | Waiver status remains `active` |

#### Time-Based Expiry Logic

```python
def check_time_based_expiry(waiver):
    if waiver.status != "active":
        return waiver.status  # No change for non-active waivers
    
    if current_time() >= waiver.expires_at:
        waiver.status = "expired"
        waiver.expired_at = current_time()
        waiver.expiry_reason = "duration_elapsed"
        log_expiry_event(waiver)
        notify_stakeholders(waiver, "expired")
    
    return waiver.status
```

### Event-Based Expiry

Certain events MUST trigger immediate waiver expiry regardless of time remaining:

| Trigger Event | Expiry Behavior | Notification |
|---------------|-----------------|--------------|
| Underlying gate passes | Immediate expiry | Team notified |
| Gate requirement removed | Immediate expiry | Team notified |
| Interface decommissioned | Immediate expiry | Team notified |
| Category policy change | Immediate expiry | Team notified |
| Emergency revocation | Immediate expiry | Tech Lead notified |

#### Event-Based Expiry Logic

```python
def check_event_based_expiry(waiver, event):
    expiry_triggers = {
        "gate_passed": True,
        "requirement_removed": True,
        "interface_decommissioned": True,
        "policy_change": True,
        "emergency_revocation": True
    }
    
    if event.type in expiry_triggers:
        waiver.status = "expired"
        waiver.expired_at = current_time()
        waiver.expiry_reason = f"event:{event.type}"
        waiver.triggering_event = event.id
        log_expiry_event(waiver)
        notify_stakeholders(waiver, f"expired_due_to_{event.type}")
        return True
    
    return False
```

### Combined Expiry Check

```python
def evaluate_waiver_status(waiver):
    # Check event-based expiry first (higher priority)
    if check_event_based_expiry(waiver, get_pending_events(waiver.interface)):
        return "expired"
    
    # Then check time-based expiry
    if check_time_based_expiry(waiver) == "expired":
        return "expired"
    
    return waiver.status
```

---

## Renewal Process

### Renewal Requirements

Renewal is NOT automatic. All renewals require full re-approval:

| Requirement | Description |
|-------------|-------------|
| New Request | Must submit new waiver request (not extension of existing) |
| Fresh Evidence | Updated business justification required |
| Updated Risk Assessment | Current risk evaluation mandatory |
| Progress Documentation | Evidence of work toward resolution |
| Full Approval Cycle | All category-required approvers must sign off again |

### Renewal Workflow

```
┌──────────────────────┐
│  Expiring Soon       │
│  (24h before expiry) │
└──────────┬───────────┘
           │
           ▼
┌──────────────────────┐     ┌──────────────────┐
│  Renewal Reminder    │────▶│  No Renewal      │
│  Sent to Requester   │     │  Needed          │
└──────────┬───────────┘     └──────────────────┘
           │
           ▼
┌──────────────────────┐
│  Submit New Request  │
│  (with fresh evidence)│
└──────────┬───────────┘
           │
           ▼
┌──────────────────────┐
│  Full Approval Cycle │
│  (same as new waiver)│
└──────────┬───────────┘
           │
           ▼
┌──────────────────────┐
│  New Waiver Active   │
│  (new expiry time)   │
└──────────────────────┘
```

### Renewal Timing Constraints

| Timing | Rule |
|--------|------|
| Earliest renewal submission | 48 hours before expiry |
| Latest renewal submission | Before expiry (expired = new full request) |
| Maximum renewals | 1 per original waiver (category dependent) |

### Renewal Request Schema

```yaml
renewal_request:
  original_waiver_id: "waiver-2026-02-18-001"
  renewal_type: "standard"  # standard | expedited
  
  progress_since_original:
    work_completed:
      - "Query optimization 60% complete"
      - "Test coverage improved to 85%"
    remaining_work:
      - "Complete index restructuring"
      - "Performance validation"
    resolution_eta: "2026-02-20T18:00:00Z"
  
  updated_justification:
    impact_statement: "Continued partial data during optimization"
    urgency_explanation: "Feature launch dependent, fix in progress"
    stakeholder_update: "Product team and customers informed of timeline"
  
  updated_risk_assessment:
    user_impact: "Low - Feature flag limits exposure"
    data_integrity: "Low - Read-only operations"
    system_stability: "Low - Monitoring in place"
    overall_risk_level: "Low"
  
  requested_duration: "72h"
  
  approvals:  # Fresh approvals required
    - role: "tech_lead"
      required: true
      status: "pending"
    - role: "qa_lead"
      required: true
      status: "pending"
```

### Renewal Rejection Criteria

Renewal MUST be rejected if:

| Criterion | Reason |
|-----------|--------|
| No progress documented | Original issue not being addressed |
| Risk level increased | Situation has deteriorated |
| Original waiver violated | Terms of original approval not met |
| Alternative resolution available | No longer need for exception |
| Extension limit reached | Category maximum renewals exceeded |

---

## Expiry Enforcement

### Automatic Block Restoration

When a waiver expires, the release block MUST be automatically restored:

```python
def evaluate_release_permission(gate_result, waiver):
    # Gate passed - no waiver needed
    if gate_result.status == "passed":
        return True
    
    # No waiver exists - block release
    if waiver is None:
        return False
    
    # Check waiver is active and not expired
    if waiver.status != "active":
        return False
    
    if current_time() >= waiver.expires_at:
        # Auto-expire the waiver
        waiver.status = "expired"
        waiver.expired_at = current_time()
        waiver.expiry_reason = "duration_elapsed"
        return False  # Block release
    
    return True  # Valid active waiver
```

### Expiry Notification Flow

```
┌─────────────────────┐
│   Waiver Expires    │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  Update Status to   │
│  'expired'          │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  Block Release      │
│  Pipeline           │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  Notify Stakeholders│
│  - Requester        │
│  - Approvers        │
│  - Release Manager  │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  Log Expiry Event   │
│  in Audit Trail     │
└─────────────────────┘
```

### Expiry Behavior Matrix

| Waiver State | Release Status | Action Required |
|--------------|----------------|-----------------|
| `active` + not expired | Allowed | None |
| `active` + expired | Blocked | Renewal or fix |
| `expired` | Blocked | New waiver or fix |
| `revoked` | Blocked | New waiver or fix |
| `cancelled` | Blocked | New waiver or fix |

### Grace Period Rules

| Category | Grace Period | Notes |
|----------|--------------|-------|
| Critical Hotfix | None | Immediate enforcement |
| Planned Migration | None | Pre-planned, no excuse |
| Feature Flag Gate | 1 hour | Allow graceful feature disable |
| Performance Exception | None | Active issue, immediate |
| Dependency Delay | 2 hours | External coordination buffer |
| Test Environment | None | Infrastructure fix needed |

---

## Expiry Behavior Types

### Invalidation (Default)

By default, expired waivers are invalidated and provide no release permission:

```
expired_waiver:
  status: "expired"
  release_permission: false
  renewal_allowed: true  # Subject to category limits
  historical_reference: true
```

### Re-Review Trigger

For certain high-risk categories, expiry may trigger mandatory re-review:

| Category | Expiry Behavior | Re-Review Required |
|----------|-----------------|-------------------|
| Critical Hotfix | Invalidation | Yes - post-mortem required |
| Planned Migration | Invalidation | No - was pre-planned |
| Feature Flag Gate | Invalidation | Yes - feature status review |
| Performance Exception | Invalidation | Yes - optimization progress check |
| Dependency Delay | Invalidation | No - external factor |
| Test Environment | Invalidation | No - infrastructure issue |

### Re-Review Process

When re-review is triggered:

```yaml
re_review_trigger:
  waiver_id: "waiver-2026-02-18-001"
  triggered_at: "2026-02-19T08:00:00Z"
  trigger_reason: "performance_exception_expiry"
  
  review_requirements:
    - type: "post_mortem"
      due: "48 hours after expiry"
      owner: "Tech Lead"
    - type: "optimization_progress_report"
      due: "24 hours after expiry"
      owner: "API Owner"
    - type: "updated_risk_assessment"
      due: "24 hours after expiry"
      owner: "QA Lead"
  
  blocking_until_review: false  # Does not block new waiver
```

---

## Machine-Readable Schema

### Expiry Status Determination

The following schema enables automated determination of waiver expiry status:

```yaml
waiver_expiry_schema:
  version: "1.0.0"
  
  status_determination:
    algorithm: |
      IF waiver.status NOT IN ["approved", "active"]:
          RETURN waiver.status  # expired, revoked, cancelled, rejected
      
      IF current_timestamp >= waiver.expires_at:
          RETURN "expired"
      
      FOR event IN pending_events:
          IF event.type IN ["gate_passed", "requirement_removed", 
                           "interface_decommissioned", "policy_change"]:
              RETURN "expired"
      
      RETURN "active"
  
  fields:
    waiver_id:
      type: string
      required: true
      description: "Unique waiver identifier"
    
    status:
      type: enum
      values: ["requested", "pending_approval", "approved", "active", 
               "expired", "revoked", "cancelled", "rejected"]
      required: true
    
    category:
      type: enum
      values: ["critical_hotfix", "planned_migration", "feature_flag_gate",
               "performance_exception", "dependency_delay", "test_environment"]
      required: true
    
    activated_at:
      type: timestamp
      format: "ISO 8601"
      required: true
      description: "When waiver became active"
    
    expires_at:
      type: timestamp
      format: "ISO 8601"
      required: true
      description: "Calculated expiry time"
    
    duration_seconds:
      type: integer
      required: true
      description: "Waiver duration in seconds"
    
    max_duration_seconds:
      type: integer
      required: true
      description: "Category maximum duration"
    
    expired_at:
      type: timestamp
      format: "ISO 8601"
      required: false
      description: "Actual expiry timestamp (when status changed to expired)"
    
    expiry_reason:
      type: enum
      values: ["duration_elapsed", "event:gate_passed", 
               "event:requirement_removed", "event:interface_decommissioned",
               "event:policy_change", "event:emergency_revocation"]
      required: false
    
    renewal_count:
      type: integer
      default: 0
      description: "Number of renewals (original = 0)"
    
    max_renewals:
      type: integer
      description: "Maximum allowed renewals for category"
    
    allows_release:
      type: boolean
      computed: true
      description: "Whether this waiver currently unblocks release"
```

### Release Permission Check Schema

```yaml
release_permission_check:
  version: "1.0.0"
  
  algorithm: |
    # Returns true if release is allowed, false otherwise
    FUNCTION check_release_permission(gate_result, waiver):
        IF gate_result.status == "passed":
            RETURN true
        
        IF waiver IS null:
            RETURN false
        
        IF waiver.status != "active":
            RETURN false
        
        IF current_timestamp >= waiver.expires_at:
            RETURN false
        
        RETURN true
  
  output:
    allowed:
      type: boolean
      description: "Whether release can proceed"
    
    blocking_reason:
      type: string
      nullable: true
      description: "Reason if blocked (null if allowed)"
    
    waiver_status:
      type: string
      description: "Current waiver status for logging"
    
    check_timestamp:
      type: timestamp
      description: "When this check was performed"
```

### Example Machine-Readable Output

```json
{
  "waiver_expiry_evaluation": {
    "waiver_id": "waiver-2026-02-18-001",
    "evaluation_timestamp": "2026-02-19T12:00:00Z",
    
    "status": {
      "current": "expired",
      "previous": "active",
      "changed_at": "2026-02-19T08:00:00Z",
      "reason": "duration_elapsed"
    },
    
    "timing": {
      "activated_at": "2026-02-18T08:00:00Z",
      "expires_at": "2026-02-19T08:00:00Z",
      "duration_seconds": 86400,
      "remaining_seconds": 0,
      "overdue_seconds": 14400
    },
    
    "release_permission": {
      "allowed": false,
      "blocking_reason": "waiver_expired",
      "resolution_options": [
        "submit_renewal_request",
        "resolve_gate_failure"
      ]
    },
    
    "renewal": {
      "eligible": true,
      "renewal_count": 0,
      "max_renewals": 1,
      "remaining_renewals": 1,
      "earliest_submission": "2026-02-17T08:00:00Z",
      "deadline_passed": true
    },
    
    "notifications": {
      "expiry_warning_sent": true,
      "expiry_notification_sent": true,
      "recipients": ["dataset-team", "tech-lead@example.com"]
    }
  }
}
```

---

## Monitoring and Compliance

### Expiry Metrics

| Metric | Threshold | Alert |
|--------|-----------|-------|
| Expired waivers still referenced | > 0 | Critical |
| Waivers expiring within 24h | Any | Warning |
| Renewals requested after expiry | > 10% | Warning |
| Average time to renewal | > 4 hours | Info |

### Compliance Checks

| Check | Frequency | Owner |
|-------|-----------|-------|
| Expiry enforcement accuracy | Hourly | System |
| Notification delivery | Per expiry | System |
| Renewal process compliance | Weekly | QA Team |
| Duration limit adherence | Per waiver | System |

---

## Audit Trail

### Expiry Event Logging

All expiry events MUST be logged:

```yaml
expiry_audit_log:
  event_id: "exp-2026-02-19-001"
  waiver_id: "waiver-2026-02-18-001"
  
  event_type: "time_based_expiry"
  timestamp: "2026-02-19T08:00:00Z"
  
  details:
    activated_at: "2026-02-18T08:00:00Z"
    expires_at: "2026-02-19T08:00:00Z"
    category: "performance_exception"
    duration_granted: "24h"
    duration_used: "24h"
  
  actions_taken:
    - action: "status_update"
      from: "active"
      to: "expired"
    - action: "release_block_restored"
      interface: "/api/dataset/list"
    - action: "notification_sent"
      recipients: ["dataset-team", "tech-lead@example.com"]
  
  renewal_status:
    renewal_requested: false
    renewal_eligible: true
    renewal_deadline: "2026-02-19T08:00:00Z"
```

---

## References

- `exception-approval-policy.md` - Exception categories and approval workflow
- `pre-release-gate-enforcement.md` - Gate blocking and waiver governance
- `required-gate-interfaces.md` - Interface scope and ownership
- `critical-whitelist.yaml` - Authoritative interface catalog
- `openspec/changes/update-api-compatibility-bridge-with-required-gate-policy/` - Spec requirements

---

*Effective Date: 2026-02-18*
*Last Updated: 2026-02-18*
*Version: 1.0.0*
