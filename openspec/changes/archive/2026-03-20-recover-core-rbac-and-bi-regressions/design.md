## Context

Current repository evidence points to a recovery problem, not a blank-slate implementation problem. The backend still wires auth, user, role, organization, menu, permission, datasource, dataset, and visualization handlers in `apps/backend-go/internal/transport/http/router.go`. The stronger suspicion is that the frontend’s login bootstrap and authorization-driven route chain is making whole feature domains appear missing when they are actually unreachable.

The highest-risk path is now concentrated in:

- `apps/frontend/src/views/login/index.vue`
- `apps/frontend/src/store/modules/user.ts`
- `apps/frontend/src/permission.ts`
- `apps/frontend/src/store/modules/permission.ts`
- `apps/frontend/src/router/index.ts`

This change exists to recover that chain systematically before touching deeper feature modules. If the total gate remains broken, any module-level repair work will be misleading and unstable.

## Goals / Non-Goals

**Goals**
- Restore reachability of core RBAC and BI feature domains.
- Distinguish perceived feature loss from route generation loss, authorization misclassification, API mismatch, and real implementation gaps.
- Re-establish regression coverage for login → menu → route → page-init → feature-access flows.

**Non-Goals**
- Re-implementing all modules from scratch.
- Refactoring unrelated runtime/UI improvements unless they are proven to block feature recovery.
- Archiving or reworking prior stabilization changes; this change builds on them.

## Decisions

### Decision: Recover the total gate before module workflows
The repair order will start with login bootstrap, current-user initialization, authorized menu retrieval, dynamic route generation, and route validation.

- **Why:** these control whether all downstream features appear present at all.
- **Rejected alternative:** start by fixing individual pages. That risks masking the real gate failure and spreading duplicate work across modules.

### Decision: Use a feature-recovery matrix as the source of truth
Each affected domain must be classified before repairs begin.

- **Why:** “功能丢了” can mean multiple failure modes that require very different fixes.
- **Rejected alternative:** jump straight to implementation from anecdotal symptoms.

### Decision: Recover RBAC administration before BI page workflows
Once routes and menus are healthy, the next layer is user/role/org/menu/permission administration.

- **Why:** those features both validate the authorization chain and provide the administrative controls needed for BI reachability.

### Decision: Treat BI recovery as entry-chain recovery first
Datasource, dataset, dashboard, and big-screen repair should start from menu visibility, route entry, and resource-tree / initialization APIs, not from deep page internals.

- **Why:** current evidence suggests BI handlers still exist; the likely loss is discovery or access path failure.

## Risks / Trade-offs

- **Misdiagnosing unreachable as unimplemented** → mitigated by the recovery matrix.
- **Fixing frontend symptoms while backend contract drift remains** → mitigated by parallel API/path verification in each phase.
- **Broad regression scope causing slow iteration** → mitigated by strict phased ordering and feature-family checkpoints.

## Recovery Phases

### Phase 0 — Recovery matrix
- Build a domain-by-domain matrix for the 9 reported feature families.
- For each, classify: route missing, menu missing, unauthorized misclassified, API mismatch, page init failure, or real backend gap.

### Phase 1 — Total gate recovery
- Login success path
- current-user bootstrap
- authorized menu fetch
- dynamic route generation
- `pathValid()` and unauthorized-vs-404 semantics

### Phase 2 — RBAC administration recovery
- User page workflows
- Role page workflows
- Organization workflows
- Menu administration
- Permission administration and compatibility views

### Phase 3 — BI entry-chain recovery
- Datasource
- Dataset
- Dashboard
- Big-screen

### Phase 4 — Regression gate hardening
- Automated route/access checks
- targeted smoke coverage
- contract and permission verification

## Open Questions

- Whether some current frontend runtime fallback changes should be narrowed after the total gate is healthy.
- Whether target permission compatibility APIs need full functional restoration or only explicit UI fallback to recover the admin pages safely.
