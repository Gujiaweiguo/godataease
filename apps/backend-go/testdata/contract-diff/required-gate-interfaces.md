# Required Gate Interfaces Policy

## Purpose

This policy defines the required-gate interface set that MUST pass compatibility checks before any release. It establishes the scope of interfaces, ownership attribution, change approval process, and audit requirements for maintaining release integrity during the Java-to-Go migration.

---

## Interface Scope Definition

### Scope Classification

Required-gate interfaces are classified into three tiers based on criticality:

| Tier | Priority | Blocking Level | Description |
|------|----------|----------------|-------------|
| Critical | P0 | critical | Core business APIs that block all deployments immediately |
| High | P1 | high | Important APIs that block merge within 24 hours if unresolved |
| Normal | P2 | normal | Standard APIs that flag for review but do not auto-block |

### Inclusion Criteria

An interface MUST be included in the required-gate set when:

1. **Business Critical**: The API is essential for core user workflows (auth, data, export)
2. **High Traffic**: The API handles significant request volume in production
3. **Integration Point**: The API is consumed by external systems or plugins
4. **Security Sensitive**: The API handles authentication, authorization, or data access control

### Exclusion Criteria

An interface MAY be excluded from required-gate scope when:

1. Deprecated and scheduled for removal
2. Internal-only with no external consumers
3. Experimental or behind feature flags
4. Documented as non-guaranteed in API contract

### Scope Source

The authoritative source for required-gate interfaces is:

```
backend-go/testdata/contract-diff/critical-whitelist.yaml
```

Key sections:
- `criticalApis`: P0 interfaces with critical blocking level
- `highPriorityApis`: P1 interfaces with high blocking level
- `nativeGoRoutes`: Go-only routes (reference, not subject to Java comparison)

---

## Source Attribution Rules

### Ownership Model

Every required-gate interface MUST have a designated owner:

| Role | Responsibility |
|------|----------------|
| API Owner | Primary accountability for interface correctness, contract compliance, and change requests |
| Tech Lead | Architecture oversight, security review, approval authority |
| QA Lead | Test coverage validation, baseline verification, release sign-off |

### Team Attribution

Interfaces are attributed to domain teams:

| Team | Scope |
|------|-------|
| auth-team | Authentication, logout, session management |
| datasource-team | Data source CRUD, connection, validation |
| dataset-team | Dataset tree, field metadata, data preview |
| chart-team | Chart data retrieval, visualization |
| template-team | Template management, marketplace |
| export-team | Export tasks, download |
| user-team | User CRUD, permissions |
| share-team | Share creation, validation |
| org-team | Organization structure, hierarchy |

### Attribution Requirements

Each interface entry MUST include:

```yaml
- path: "/api/endpoint"
  method: "POST"
  owner: "team-name"        # Required: responsible team
  priority: "P0"            # Required: P0, P1, or P2
  blockingLevel: "critical" # Required: critical, high, or normal
  javaStatus: "exists"      # Required: exists, partial, n/a
  goStatus: "full"          # Required: full, partial, stub, missing
  notes: "Context"          # Optional: additional information
```

---

## Change Approval Process

### Scope Change Types

| Change Type | Description | Required Approvers |
|-------------|-------------|-------------------|
| Add Interface | Include new API in required-gate scope | API Owner + Tech Lead |
| Remove Interface | Exclude API from required-gate scope | Tech Lead + QA Lead |
| Modify Priority | Change P0/P1/P2 classification | API Owner + Tech Lead |
| Modify Blocking Level | Change critical/high/normal | Tech Lead |
| Update Ownership | Transfer to different team | Current Owner + New Owner + Tech Lead |

### Approval Workflow

1. **Submit Request**: Open PR with changes to `critical-whitelist.yaml`
2. **Document Rationale**: Include justification in PR description
3. **Link Evidence**: Reference related issues, specs, or incident reports
4. **Required Approvals**: Obtain signatures per change type matrix
5. **Merge**: Only after all required approvals obtained

### Emergency Changes

For urgent production issues:

1. **Hotfix Path**: Tech Lead may approve emergency changes verbally
2. **Documentation**: Must create retroactive PR within 24 hours
3. **Review**: Full approval process required within 48 hours
4. **Audit Trail**: Incident must be linked to change record

---

## Audit Requirements

### Traceability

Every scope change MUST be traceable:

| Element | Requirement |
|---------|-------------|
| Change ID | Unique identifier linked to PR or issue |
| Timestamp | ISO 8601 format recording when change occurred |
| Actor | Identity of requester and approvers |
| Rationale | Documented reason for change |
| Evidence | Links to supporting documentation |

### Versioning

Scope changes follow semantic versioning:

- **Format**: `v{MAJOR}.{MINOR}.{PATCH}`
- **MAJOR**: Breaking changes to scope definition
- **MINOR**: Additions or removals of interfaces
- **PATCH**: Corrections, clarifications, metadata updates

Version metadata stored in `critical-whitelist.yaml`:

```yaml
metadata:
  version: "1.0.0"
  changeId: "change-identifier"
  lastUpdated: "2026-02-18"
  sourceMatrix: "path/to/compatibility-matrix.md"
```

### Audit Log

All scope changes are recorded in Git history with:

- Commit message containing change ID
- Signed commits for approval chain
- Linked PR with approval comments

### Compliance Checks

| Check | Frequency | Owner |
|-------|-----------|-------|
| Orphan interfaces (no owner) | Weekly | CI Pipeline |
| Stale interfaces (deprecated > 30 days) | Monthly | QA Team |
| Ownership validation | Quarterly | Tech Lead |
| Full scope review | Quarterly | All Stakeholders |

---

## Release Blocking Behavior

### Gate Failure Handling

When required-gate interfaces fail compatibility checks:

| Blocking Level | Behavior |
|----------------|----------|
| critical | CI fails immediately, no override without approved waiver |
| high | CI warns, fails after 24-hour grace period |
| normal | CI logs warning, allows merge with acknowledgment |

### Failure Evidence

Gate failures MUST provide actionable evidence:

1. **Interface Identifier**: Path and method of failing interface
2. **Expected vs Actual**: Contract diff showing discrepancy
3. **Baseline Reference**: Link to expected response in baselines/
4. **Remediation Steps**: Suggested actions to resolve

### Waiver Governance

Temporary waivers for gate failures:

| Requirement | Enforcement |
|-------------|-------------|
| Approval Required | Waiver inactive until Tech Lead + QA Lead approve |
| Time-Bounded | Maximum 7-day duration, must specify expiry |
| Automatic Expiry | System enforces expiry, blocks release if expired |
| Renewal | Requires full re-approval, not automatic extension |

---

## Maintenance Schedule

| Activity | Frequency | Owner | Deliverable |
|----------|-----------|-------|-------------|
| Scope integrity check | Weekly | CI Pipeline | Automated report |
| Orphan interface cleanup | Monthly | QA Team | PR with removals |
| Ownership audit | Quarterly | Tech Lead | Updated attributions |
| Full policy review | Quarterly | All Stakeholders | Updated policy doc |
| Baseline alignment | Quarterly | QA Lead | Verified baselines/ |

### Quarterly Review Checklist

- [ ] Verify all interfaces have valid ownership
- [ ] Remove interfaces for deprecated endpoints
- [ ] Add interfaces for new critical features
- [ ] Validate priority and blocking level accuracy
- [ ] Confirm threshold values remain appropriate
- [ ] Update documentation for any process changes

---

## References

- `baseline-policy.md` - Baseline capture and maintenance
- `whitelist-governance.md` - Whitelist entry management
- `critical-whitelist.yaml` - Authoritative interface catalog
- `openspec/changes/update-api-compatibility-bridge-with-required-gate-policy/` - Spec requirements

---

*Effective Date: 2026-02-18*
*Last Updated: 2026-02-18*
*Version: 1.0.0*
