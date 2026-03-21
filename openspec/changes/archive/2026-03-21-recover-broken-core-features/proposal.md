## Why

After the system-management, RBAC, and menu-recovery scope is isolated, multiple broken feature flows still remain across non-system-management modules. These issues span BI entry chains and operational pages, so they need a dedicated stabilization change instead of being mixed into the RBAC/menu recovery line.

## What Changes

- Build a module-based broken-feature inventory for non-system-management critical flows and classify each issue as route/access regression, API mismatch, page-init failure, state-sync failure, or real implementation gap.
- Recover critical broken flows in controlled batches, with datasource, dataset, and visualization in the first execution batch while later operational follow-up work is tracked separately.
- Define deterministic failure semantics so unauthenticated, forbidden, missing-route/resource, and explicit non-success outcomes for unsupported, dependency, and business failures are distinguishable during recovery.
- Add targeted regression and smoke gates for each recovered batch before broadening scope.

## P0 / P1 batching intent

The first execution batch for this change is intentionally limited to the P0 BI critical-path families:

- datasource
- dataset
- dashboard
- big-screen

This P0 batch is limited to governed reachability, initialization, contract, discovery, and state-consumption recovery for those BI flows. It does not include new product capabilities, broad UI redesign, or low-value cosmetic cleanup.

The following operational families were explicitly held outside the first batch and moved into a separate follow-up change after the P0 BI paths stabilized:

- export-center
- audit

Any newly discovered issue must remain outside P0 unless it is classified, has bounded verification cost, and is added to the governed P0 cut line explicitly.

Real implementation gaps may only enter P0 when they directly block a governed BI critical path and remain bounded enough to verify and close inside the first batch.

## Capabilities

### New Capabilities

### Modified Capabilities
- `datasource-management`: tighten datasource entry, page-init, and recovery expectations for broken critical flows
- `dataset-management`: tighten dataset entry, initialization, and failure-semantics expectations for broken critical flows
- `visualization-management`: tighten dashboard and big-screen entry-chain stability and recovery expectations

## Impact

- **Frontend**: datasource, dataset, dashboard, big-screen, export-center, and audit pages; related API wrappers; route-entry and page-init logic.
- **Backend Go**: datasource, dataset, visualization, export, and audit handlers/services/repositories plus any compatibility or middleware wiring needed by the recovered flows.
- **Verification**: broken-feature inventory, targeted regression tests, smoke coverage, and release-readiness evidence for each recovery batch, with the first batch centered on datasource, dataset, dashboard, and big-screen recovery.
