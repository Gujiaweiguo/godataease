# T5 Foundation Policy Lock Package

## Purpose
This document converts the T1-T4 evidence into an executable T5 entry package so the next implementation wave can start without reopening baseline analysis.

## Default Next Wave Order
1. Lock last-role policy.
2. Lock organization delete policy.
3. Implement corresponding backend enforcement.
4. Add targeted backend tests for policy semantics.
5. Add frontend/API regression checks for changed semantics.

## Decision Package A — Last-Role Policy

### Option A1 (Recommended)
- Policy: Block removal of the user's last remaining role and record this as an intentional fork deviation.
- Why:
  - lower destructive risk than implicit user deletion
  - consistent with current implementation tendency
  - easier rollback and auditability
- Required implementation consequences:
  - role-member removal path must detect last-role condition
  - backend returns deterministic validation error instead of silent success
  - user-management and role-management UI must surface the same semantic outcome
- Required verification:
  - removal with multiple roles succeeds and keeps user active
  - removal with last remaining role is rejected deterministically
  - denial is distinguishable from missing-user or generic server error

### Option A2
- Policy: Remove user when the last remaining role is removed.
- Why:
  - closer to a strict reading of the official handbook flow
- Required implementation consequences:
  - role-member removal path must cascade into user deactivation/deletion policy
  - audit trail must record role-removal-driven account deletion
  - downstream org membership and permission relations must remain consistent
- Required verification:
  - last-role removal deletes or deactivates governed user deterministically
  - linked relations are cleaned consistently
  - non-last-role removal still behaves as partial detach only

## Decision Package B — Organization Delete Policy

### Option B1 (Recommended)
- Policy: Keep child-organization precondition, soft-delete eligible leaf organizations, and make resource disposition auditable and traceable.
- Why:
  - matches current implementation direction more closely
  - lower irreversible data-loss risk
  - preserves rollback safety while still meeting deterministic-policy requirement
- Required implementation consequences:
  - deleting org with children remains blocked
  - deleting leaf org triggers governed soft-delete path
  - resource disposition outcome is recorded and queryable
- Required verification:
  - delete with children is rejected with explicit dependency reason
  - delete leaf org succeeds through soft-delete path
  - affected resource disposition is traceable in audit output or status record

### Option B2
- Policy: Synchronous cascade delete of governed resources after child preconditions pass.
- Why:
  - closest to strict “delete org resources together” interpretation
- Required implementation consequences:
  - cascade path must enumerate governed resources deterministically
  - failure handling must be transactional or compensatable
- Required verification:
  - no orphaned governed resource remains after delete
  - partial cascade failure is impossible or explicitly rolled back

### Option B3
- Policy: Async cleanup after org delete request succeeds.
- Why:
  - reduces request latency for large orgs
- Required implementation consequences:
  - delete API must return cleanup job state
  - interim resource visibility rules must remain safe
- Required verification:
  - deleted org does not remain accessible during cleanup lag
  - cleanup completion and failures are observable

## Recommended Combined Path
- Last-role policy: **A1**
- Organization delete policy: **B1**

This combined path minimizes destructive behavior, aligns with the current code tendency, and preserves clear rollback surfaces while still satisfying the requirement that policy be deterministic and testable.

## T5 Immediate Implementation Scope Once Policy Is Confirmed
- backend
  - `apps/backend-go/internal/service/role_service.go`
  - `apps/backend-go/internal/service/org_service.go`
  - related handlers and repositories touched by policy enforcement
- tests
  - add/extend unit tests for role-member removal semantics
  - add/extend unit or integration tests for organization delete semantics
- regression
  - update governed frontend/API expectations only where semantics change

## Blocker Boundary
T5 implementation should not start before the two policy decisions are confirmed because the downstream enforcement and regression expectations differ materially between the available branches.
