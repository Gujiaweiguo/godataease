# Exception Approval Policy

## Purpose

This policy defines the mandatory approval process for exceptions to pre-release gate requirements. It establishes the conditions under which exceptions may be requested, the approval chain governance, evidence requirements, and enforcement rules that ensure unapproved exceptions cannot bypass release controls.

---

## Exception Conditions

### When Exceptions May Be Requested

An exception request MAY be submitted when all of the following conditions are met:

1. **Gate Failure Documented**: The required-gate interface failure has been identified and logged
2. **Immediate Fix Not Feasible**: Resolution cannot be completed within the release window
3. **Business Justification Exists**: A valid business reason for proceeding despite the failure
4. **Risk Acceptable**: The team has assessed and accepted the associated risks
5. **Mitigation Planned**: Concrete steps are defined to reduce impact during the exception period

### When Exceptions MUST NOT Be Requested

Exceptions MUST NOT be requested when:

1. **No Business Urgency**: The release can be delayed without significant impact
2. **Fix Is Available**: A resolution exists and can be applied within reasonable time
3. **Previous Exception Active**: An existing exception covers the same interface/failure
4. **Critical Security Issue**: The failure involves a security vulnerability
5. **Regulatory Non-Compliance**: Proceeding would violate compliance requirements

### Exception Categories

| Category | Max Duration | Description | Typical Use Case |
|----------|--------------|-------------|------------------|
| Critical Hotfix | 24 hours | Emergency production fix | Incident response, security patch |
| Planned Migration | 7 days | Known migration window | Java-to-Go transition, infrastructure change |
| Feature Flag Gate | 7 days | Controlled rollout behind flag | Gradual feature enablement |
| Performance Exception | 3 days | Performance optimization in progress | Latency tuning, scaling work |
| Dependency Delay | 5 days | External dependency unavailable | Third-party API, vendor issue |
| Test Environment | 3 days | Test infrastructure limitation | CI/CD issue, test data problem |

---

## Approval Chain

### Approval Roles and Responsibilities

| Role | Authority | Responsibility |
|------|-----------|----------------|
| API Owner | Initiator | Submits exception request, provides technical context |
| Tech Lead | Technical Approval | Validates risk assessment, approves mitigation plan |
| Release Manager | Release Approval | Confirms release timing, coordinates deployment |
| QA Lead | Quality Approval | Verifies test coverage, accepts quality risk |

### Required Approvers by Category

| Exception Category | Required Approvers | Approval Order |
|--------------------|-------------------|----------------|
| Critical Hotfix | Tech Lead | 1. Tech Lead |
| Planned Migration | Tech Lead → QA Lead → Release Manager | 1. Tech Lead, 2. QA Lead, 3. Release Manager |
| Feature Flag Gate | Tech Lead → QA Lead | 1. Tech Lead, 2. QA Lead |
| Performance Exception | Tech Lead → QA Lead | 1. Tech Lead, 2. QA Lead |
| Dependency Delay | Tech Lead → Release Manager | 1. Tech Lead, 2. Release Manager |
| Test Environment | Tech Lead → QA Lead | 1. Tech Lead, 2. QA Lead |

### Approval Workflow

```
┌─────────────────┐
│  API Owner      │
│  Submit Request │
└────────┬────────┘
         │
         ▼
┌─────────────────┐     ┌──────────────┐
│  Tech Lead      │────▶│  Rejected    │
│  Review         │     └──────────────┘
└────────┬────────┘
         │ Approved
         ▼
┌─────────────────┐     ┌──────────────┐
│  QA Lead        │────▶│  Rejected    │
│  Review         │     │  (if req'd)  │
└────────┬────────┘     └──────────────┘
         │ Approved
         ▼
┌─────────────────┐     ┌──────────────┐
│ Release Manager │────▶│  Rejected    │
│  Review         │     │  (if req'd)  │
└────────┬────────┘     └──────────────┘
         │ Approved
         ▼
┌─────────────────┐
│  Exception      │
│  ACTIVATED      │
└─────────────────┘
```

### Approval States

| State | Description | Release Impact |
|-------|-------------|----------------|
| `requested` | Initial submission, pending review | Block maintained |
| `pending_tech_lead` | Awaiting Tech Lead approval | Block maintained |
| `pending_qa_lead` | Tech Lead approved, awaiting QA Lead | Block maintained |
| `pending_release_manager` | QA Lead approved, awaiting Release Manager | Block maintained |
| `approved` | All required approvals obtained | Block maintained |
| `active` | Exception activated, release unblocked | Release allowed |
| `rejected` | Request denied by any approver | Block maintained |
| `cancelled` | Request withdrawn by submitter | Block maintained |
| `expired` | Duration elapsed | Block restored |

---

## Evidence Requirements

### Mandatory Evidence Fields

Every exception request MUST include:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `request_id` | string | Yes | Unique identifier for the exception request |
| `interface_path` | string | Yes | API path that requires exception |
| `interface_method` | string | Yes | HTTP method of the interface |
| `gate_failure_id` | string | Yes | Reference to the gate failure record |
| `category` | enum | Yes | Exception category from defined list |
| `requested_duration` | duration | Yes | Requested exception duration |
| `business_justification` | text | Yes | Why this exception is needed |
| `risk_assessment` | text | Yes | Analysis of potential impact |
| `mitigation_plan` | text | Yes | Steps to reduce risk during exception |
| `rollback_plan` | text | Yes | Steps to revert if issues occur |
| `requested_by` | string | Yes | API Owner or team submitting request |
| `created_at` | timestamp | Yes | Request submission time |

### Business Justification Requirements

The business justification MUST include:

1. **Impact Statement**: What business function is affected
2. **Urgency Explanation**: Why this cannot wait for a fix
3. **Stakeholder Awareness**: Confirmation that affected stakeholders are informed
4. **Alternative Consideration**: What alternatives were evaluated and why rejected

### Risk Assessment Requirements

The risk assessment MUST include:

| Risk Dimension | Required Content |
|----------------|------------------|
| **User Impact** | Affected user segments and expected impact level |
| **Data Integrity** | Risk to data consistency or accuracy |
| **System Stability** | Risk to overall system reliability |
| **Integration Risk** | Impact on dependent systems or APIs |
| **Compliance Risk** | Any regulatory or policy implications |

Risk levels:
- **Low**: Minimal user impact, easily reversible
- **Medium**: Some user impact, reversible with effort
- **High**: Significant user impact, difficult to reverse
- **Critical**: Widespread impact, may require emergency rollback

### Mitigation Plan Requirements

The mitigation plan MUST include:

1. **Monitoring**: Specific metrics and alerts to track during exception
2. **Communication**: User communication plan if impact is visible
3. **Escalation Path**: Who to contact if issues escalate
4. **Resolution Timeline**: When the underlying issue will be fixed
5. **Progress Checkpoints**: Scheduled reviews during exception period

### Evidence Schema

```yaml
exception_request:
  request_id: "exc-2026-02-18-001"
  interface:
    path: "/api/dataset/list"
    method: "GET"
  gate_failure_id: "gf-2026-02-18-042"
  category: "performance_exception"
  requested_duration: "72h"
  
  business_justification:
    impact_statement: "Dataset list API returns partial data due to query optimization in progress"
    urgency_explanation: "Dashboard refresh feature requires this API; delaying release blocks feature launch"
    stakeholder_awareness: "Product team and key customers notified"
    alternatives_considered: "Feature flag considered but requires UI changes not ready for this release"
  
  risk_assessment:
    user_impact: "Medium - Users may see incomplete dataset lists"
    data_integrity: "Low - No data modification involved"
    system_stability: "Low - API is read-only"
    integration_risk: "Low - No external integrations affected"
    compliance_risk: "None"
    overall_risk_level: "Medium"
  
  mitigation_plan:
    monitoring:
      - metric: "api/dataset/list/error_rate"
        threshold: "> 5%"
        alert_channel: "#ops-alerts"
      - metric: "api/dataset/list/latency_p99"
        threshold: "> 2000ms"
        alert_channel: "#ops-alerts"
    communication: "Status page updated with known issue"
    escalation_path:
      primary: "dataset-team-oncall"
      secondary: "tech-lead"
    resolution_timeline: "2026-02-20T18:00:00Z"
    progress_checkpoints:
      - "2026-02-19T12:00:00Z - Mid-point review"
  
  rollback_plan:
    trigger: "Error rate > 10% or latency > 5000ms"
    steps:
      - "Disable feature flag 'optimized-dataset-list'"
      - "Verify API returns to previous behavior"
      - "Notify stakeholders of rollback"
    estimated_time: "15 minutes"
  
  approvals:
    - role: "tech_lead"
      required: true
      status: "pending"
    - role: "qa_lead"
      required: true
      status: "pending"
  
  requested_by: "dataset-team"
  created_at: "2026-02-18T08:00:00Z"
```

---

## Enforcement Rules

### Waiver Inactive Until Approval Complete

**CRITICAL**: Exceptions remain inactive until ALL required approvers have signed off.

```
exception_status transitions:
  requested -> pending_tech_lead -> pending_qa_lead -> pending_release_manager -> approved -> active -> expired
                                     │                    │                      │
                                  rejected            rejected               cancelled

Activation requirement:
  IF all_required_approvers_signed == true:
      SET exception_status = "active"
      SET activated_at = current_time
      SET expires_at = activated_at + requested_duration
  ELSE:
      exception_status REMAINS "pending_*" or "approved"
      RELEASE REMAINS BLOCKED
      waiver_does_not_unblock_release
```

### Unapproved Exception Does Not Unblock Release

The system MUST enforce the following logic:

```python
def can_release_proceed(gate_result, exception):
    if gate_result.status == "passed":
        return True
    
    if exception is None:
        return False  # No exception exists
    
    if exception.status != "active":
        return False  # Exception not activated
    
    if exception.expires_at <= current_time():
        return False  # Exception expired
    
    return True  # Valid active exception
```

### Approval Completeness Check

Before activation, the system MUST verify:

| Check | Enforcement |
|-------|-------------|
| All required roles listed | System validates approver list against category requirements |
| Each required role has approval | System checks each required approver has `status: approved` |
| No missing approvals | System rejects activation if any required approval is missing |
| Approval timestamps valid | System records approval timestamp for each approver |

### Activation Blocking

The following conditions MUST prevent exception activation:

```
BLOCK activation IF:
  - any_required_approver.status != "approved"
  - business_justification is empty or placeholder
  - risk_assessment is empty or placeholder
  - mitigation_plan is empty or placeholder
  - requested_duration exceeds category maximum
  - duplicate_active_exception_exists_for_interface
```

---

## Anti-Bypass Rules

### Cannot Skip Required Approvers

**MANDATORY**: The approval chain MUST NOT allow skipping required approvers.

| Scenario | Allowed? | Reason |
|----------|----------|--------|
| Tech Lead approves without API Owner request | No | Request must originate from API Owner |
| QA Lead approves before Tech Lead | No | Sequential order must be respected |
| Release Manager approves without prior approvals | No | All prior approvals required |
| Single approver for multi-approver category | No | All listed approvers required |
| Self-approval by request submitter | No | Submitter cannot approve own request |

### Role Delegation Rules

| Original Role | Can Delegate To | Conditions |
|---------------|-----------------|------------|
| Tech Lead | Senior Engineer | Must be documented, Tech Lead accountable |
| QA Lead | Senior QA Engineer | Must be documented, QA Lead accountable |
| Release Manager | DevOps Lead | Must be documented, Release Manager accountable |
| API Owner | Team Member | Must be documented, API Owner accountable |

Delegation requirements:
1. Must be documented in exception request
2. Delegated approver has equivalent authority
3. Original role remains accountable
4. Delegation valid for single exception only

### Emergency Override Restrictions

Emergency overrides MUST follow these rules:

| Override Type | Who Can Authorize | Requirements |
|---------------|-------------------|--------------|
| Expedited Review | Tech Lead | Still requires all approvals, faster turnaround |
| Role Substitution | Engineering Director | Must document reason, post-hoc review required |
| Category Change | Tech Lead + QA Lead | Joint approval, documented justification |

**NO OVERRIDE EXISTS** for:
- Skipping required approvers entirely
- Extending exception beyond category maximum
- Activating exception without evidence
- Bypassing risk assessment

### Audit Trail Requirements

All approval actions MUST be logged:

```yaml
audit_log:
  - action: "submitted"
    actor: "dataset-team"
    timestamp: "2026-02-18T08:00:00Z"
    details: "Initial request submitted"
  
  - action: "approved"
    actor: "tech-lead@example.com"
    role: "tech_lead"
    timestamp: "2026-02-18T08:30:00Z"
    rationale: "Risk acceptable, mitigation plan adequate"
  
  - action: "rejected"
    actor: "qa-lead@example.com"
    role: "qa_lead"
    timestamp: "2026-02-18T09:00:00Z"
    rationale: "Insufficient test coverage verification"
```

---

## Exception Lifecycle

### State Transitions

```
                    ┌──────────────┐
                    │  Requested   │
                    └──────┬───────┘
                           │
              ┌────────────┼────────────┐
              ▼            ▼            ▼
       ┌────────────┐ ┌────────────┐ ┌────────────┐
       │  Pending   │ │  Rejected  │ │ Cancelled  │
       │ Tech Lead  │ └────────────┘ └────────────┘
       └─────┬──────┘
             │ Approved
             ▼
       ┌────────────┐     ┌────────────┐
       │  Pending   │────▶│  Rejected  │
       │  QA Lead   │     └────────────┘
       └─────┬──────┘
             │ Approved (if required)
             ▼
       ┌────────────┐     ┌────────────┐
       │  Pending   │────▶│  Rejected  │
       │  Release   │     └────────────┘
       │  Manager   │
       └─────┬──────┘
             │ Approved (if required)
             ▼
       ┌────────────┐
       │  Approved  │
       │ (inactive) │
       └─────┬──────┘
             │ System activates
             ▼
       ┌────────────┐
       │   Active   │──────▶ ┌────────────┐
       │            │        │  Expired   │
       └────────────┘        └────────────┘
```

### Automatic Actions

| Trigger | Action | Owner |
|---------|--------|-------|
| All approvals complete | Activate exception | System |
| Exception expires | Deactivate, block release | System |
| Underlying issue resolved | Prompt cancellation | System |
| 7 days before expiry | Send renewal reminder | System |
| 24 hours before expiry | Send urgent reminder | System |

### Manual Actions

| Action | Who Can Perform | Conditions |
|--------|-----------------|------------|
| Cancel exception | API Owner, Tech Lead | Before or during active state |
| Extend exception | Tech Lead + QA Lead | Requires new approval cycle |
| Early resolution | API Owner | Document resolution, close exception |
| Emergency revocation | Tech Lead, Release Manager | Immediate deactivation |

---

## Rollback Plan

### Exception Revocation Rollback

If an active exception needs to be revoked:

#### Immediate Revocation (Emergency)

| Step | Action | Owner | Time |
|------|--------|-------|------|
| 1 | Notify Tech Lead and Release Manager | Any Stakeholder | Immediate |
| 2 | Set exception status to `revoked` | Tech Lead | 5 min |
| 3 | Block release pipeline | System | Automatic |
| 4 | Notify affected team | Tech Lead | 10 min |
| 5 | Document revocation reason | Tech Lead | 30 min |

#### Revocation Schema

```yaml
revocation:
  exception_id: "exc-2026-02-18-001"
  revoked_at: "2026-02-18T15:00:00Z"
  revoked_by: "tech-lead@example.com"
  reason: "Issue escalated, risk level increased to critical"
  notification_sent:
    - channel: "#release-alerts"
      timestamp: "2026-02-18T15:01:00Z"
    - channel: "dataset-team"
      timestamp: "2026-02-18T15:01:00Z"
```

### Re-approval After Revocation

If exception is needed after revocation:

1. **New Request Required**: Cannot reinstate revoked exception
2. **Fresh Evidence**: Must provide updated justification
3. **Full Approval Cycle**: All approvers must review again
4. **Post-Mortem**: Required if revocation was due to incident

---

## Monitoring and Compliance

### Exception Metrics

| Metric | Threshold | Alert |
|--------|-----------|-------|
| Active exceptions | > 5 | Warning |
| Expired exceptions | Any | Warning |
| Exceptions per interface | > 2 per month | Warning |
| Average exception duration | > 5 days | Warning |
| Revoked exceptions | Any | Info |
| Rejected requests | > 30% of requests | Warning |

### Compliance Checks

| Check | Frequency | Owner |
|-------|-----------|-------|
| Active exception validity | Daily | System |
| Approver authority verification | Per approval | System |
| Evidence completeness | Per request | System |
| Category compliance | Weekly | QA Team |
| Policy adherence | Monthly | Tech Lead |

### Dashboard Requirements

- **Active Exceptions**: Current active exceptions with expiry times
- **Pending Approvals**: Requests awaiting approval
- **Expiring Soon**: Exceptions expiring within 24 hours
- **Historical View**: All exceptions in retention period

---

## References

- `pre-release-gate-enforcement.md` - Gate blocking and waiver governance
- `required-gate-interfaces.md` - Interface scope and ownership
- `critical-whitelist.yaml` - Authoritative interface catalog
- `openspec/changes/update-api-compatibility-bridge-with-required-gate-policy/` - Spec requirements

---

*Effective Date: 2026-02-18*
*Last Updated: 2026-02-18*
*Version: 1.0.0*
