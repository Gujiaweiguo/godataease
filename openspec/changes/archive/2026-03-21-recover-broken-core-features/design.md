## Context

This change starts after the system-management/RBAC/menu recovery line has been isolated into its own closure scope. The remaining problem is broader but still bounded: multiple non-system-management user flows appear broken across BI and operational modules, and the failures may come from route loss, API drift, page-init failures, compatibility mismatches, or real implementation gaps.

The goal here is not to redesign product behavior. It is to make broken flows visible, recover them in bounded batches, and add enough regression evidence that the same failures do not keep reappearing under different symptoms.

## Goals / Non-Goals

**Goals:**
- Define a separate stabilization lane for broken non-system-management flows.
- Recover datasource, dataset, visualization, export-center, and audit critical paths in explicit batches.
- Make failure semantics deterministic so operators can distinguish route loss, authorization failure, compatibility drift, and business failure.
- Require targeted regression and smoke verification before each recovery batch is considered complete.

**Non-Goals:**
- Reopening system-management, RBAC, or menu information-architecture scope.
- Introducing brand-new product capabilities.
- Rewriting every historical compatibility path in one pass.
- Treating all low-value UI glitches as part of the first stabilization batch.

## Decisions

### Decision: Keep broad broken-feature recovery separate from RBAC/menu recovery
This change explicitly excludes the system-management/RBAC/menu closure line.

- **Why:** that line has a single acceptance surface, while broad broken-feature recovery has multiple unrelated acceptance surfaces.
- **Alternative considered:** merge everything into one stabilization change. Rejected because it would erase completion boundaries and slow verification.

### Decision: Recover by module batch, not by ad hoc bug list
Implementation work driven by this change must proceed in bounded batches with clear ownership and regression evidence.

- **Why:** scattered bugfixing makes it difficult to tell whether a module is actually stable.
- **Alternative considered:** take bugs in user-reported order only. Rejected because symptoms often hide shared route/init/contract issues.

### Decision: Evidence-first before repair
Each recovery batch should begin with a broken-feature inventory and failing or explicit verification targets.

- **Why:** the current problem is ambiguity about what is broken versus what is merely unreachable or mismatched.
- **Alternative considered:** fix pages opportunistically and add tests later. Rejected because it allows drift to reappear without a stable boundary.

### Decision: Use additive spec deltas
This change adds stabilization requirements to existing capability specs instead of rewriting baseline requirements.

- **Why:** the base capabilities already exist; the missing piece is reliable recovery and governed verification.
- **Alternative considered:** rewrite base specs directly. Rejected because it would blur baseline capability requirements with stabilization requirements.

## Risks / Trade-offs

- **[Risk] Scope expands into all remaining product issues** → **Mitigation:** keep only the named capabilities in scope and recover them batch by batch.
- **[Risk] Teams confuse route or contract issues with true implementation gaps** → **Mitigation:** require issue classification in the broken-feature inventory before implementation work starts.
- **[Risk] Fixes land without enough regression evidence** → **Mitigation:** make targeted smoke and regression coverage part of the task list rather than optional follow-up.

## Migration Plan

1. Build a broken-feature matrix for the in-scope modules.
2. Define P0 and P1 recovery batches from that matrix.
3. Add or update failing verification for each batch.
4. Recover backend/frontend contract, route, and page-init behavior module by module.
5. Run targeted verification and record any remaining gaps separately from recovered flows.

## Open Questions

- Whether export-center and audit should remain in the first stabilization batch or move to the second batch after datasource/dataset/visualization are stable.
- Whether some in-scope flows require additional compatibility governance beyond existing route and envelope checks.
