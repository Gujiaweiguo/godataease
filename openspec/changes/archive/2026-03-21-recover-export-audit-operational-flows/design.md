## Context

This change is the direct follow-up to `recover-broken-core-features`, which closed the datasource, dataset, and visualization P0 batch while explicitly deferring export-center and audit to a later operational batch.

The goal here is still stabilization, not product redesign. Export-center and audit should be recovered with the same evidence-first discipline used in the completed BI batch, but without reopening already-closed BI recovery scope.

## Goals / Non-Goals

**Goals:**
- Recover export-center list, retry, download, and page-init behavior in a bounded batch.
- Recover audit page reachability, filter-query initialization, and detail-read behavior in a bounded batch.
- Preserve deterministic failure semantics for authorization, route/resource, and business-query outcomes.
- Require targeted regression and smoke verification before the batch is considered complete.

**Non-Goals:**
- Reopening datasource, dataset, dashboard, or big-screen recovery scope.
- Introducing new export or audit product features.
- Broad UI redesign or unrelated operational cleanup.

## Decisions

### Decision: Keep export-center and audit together as an operational follow-up batch
Export-center and audit are both operational surfaces deferred from the first BI batch, so they stay together here.

- **Why:** they were explicitly held in P1 and can now be recovered without mixing with BI stabilization.
- **Alternative considered:** reopen the prior change and extend it. Rejected because the prior batch is already complete and release-ready.

### Decision: Reuse evidence-first recovery workflow
This batch should begin from explicit verification targets and close with recorded operational evidence.

- **Why:** the same ambiguity about route loss, authorization failure, and business failure exists here.
- **Alternative considered:** implement first and test later. Rejected because it weakens closure boundaries.

## Risks / Trade-offs

- **[Risk] Operational scope drifts back into BI recovery** → **Mitigation:** keep datasource/dataset/visualization explicitly out of this follow-up change.
- **[Risk] Export and audit failures are normalized into generic empty states** → **Mitigation:** make explicit non-success semantics part of the task list.
- **[Risk] Compatibility drift is missed because entry paths seem reachable** → **Mitigation:** require route-entry, page-init, and targeted smoke verification.

## Migration Plan

1. Freeze export-center and audit evidence baselines.
2. Define failing or missing verification for export-center and audit flows.
3. Recover export-center route/contract/page-init gaps.
4. Recover audit route/contract/page-init gaps.
5. Record closeout evidence before widening scope again.

## Open Questions

- Whether export-center needs separate compatibility-path governance beyond the route/init recovery already identified.
- Whether audit detail-read should be validated through dedicated smoke or remain covered by targeted route/query regression only.
