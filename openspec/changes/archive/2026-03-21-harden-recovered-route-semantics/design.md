## Context

Two recovery batches are already archived:

- `2026-03-21-recover-broken-core-features`
- `2026-03-21-recover-export-audit-operational-flows`

Those changes fixed the primary regressions and closed their release-readiness loops. This change is intentionally narrower: it only targets the remaining hardening items explicitly documented as optional or follow-on work in the archived matrices and evidence docs.

## Goals / Non-Goals

**Goals**
- Tighten dashboard and big-screen detail-path semantics where missing-resource behavior is still strongest at service level instead of frontend-facing boundaries.
- Add route-level smoke for export-center and audit beyond the current unit/handler-heavy coverage.
- Improve confidence without reopening already-closed broad recovery scope.

**Non-Goals**
- Reimplementing the archived recovery batches.
- Broad UI redesign or unrelated cleanup.
- Expanding into new product behavior.

## Decisions

### Decision: Treat all remaining work as hardening, not recovery
These paths are already functional enough to have been archived in earlier changes, so this change should only tighten semantics and smoke confidence.

### Decision: Prefer direct boundary verification over internal refactors
Where possible, use route-level tests, handler-level tests, or route-smoke checks instead of deeper service rewrites.

### Decision: Move datasource runtime redesign into a dedicated design change
The datasource hardening slice established forbidden proof at the permission-middleware boundary, but runtime datasource list semantics require a separate permission/API design change.

- **Why:** datasource list aliases do not carry a stable runtime resource identifier, so further work would be design, not hardening.
- **Follow-up owner:** `design-datasource-list-resource-identity`

### Decision: Keep delivery scope to visualization, export-center, and audit runtime hardening
The runtime hardening completed in this change is therefore limited to visualization-management, export-center-management, and audit-logs. Dataset-management remains excluded unless a new blocker emerges.

## Risks / Trade-offs

- **[Risk] Reopening archived scope by accident** → **Mitigation:** every task must map to a documented residual hardening gap from the archived matrices/evidence.
- **[Risk] Hardening drifts into speculative coverage** → **Mitigation:** prefer route/handler assertions tied to known caller paths.
- **[Risk] Overfitting to test infrastructure** → **Mitigation:** keep fixes at real HTTP boundary or caller-path level where possible.

## Migration Plan

1. Freeze the residual hardening matrix.
2. Add the narrowest missing semantics/smoke checks per capability.
3. Apply only minimal boundary fixes revealed by those checks.
4. Re-run focused verification and then close the hardening batch.
