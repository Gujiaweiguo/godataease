## Why

The completed `recover-broken-core-features` batch intentionally stabilized only the BI-critical P0 families: datasource, dataset, dashboard, and big-screen. Export-center and audit were explicitly held out as operational P1 work so the first batch could close with bounded verification.

Now that the BI batch is complete, export task management and audit query/detail flows need their own bounded recovery change so they can be implemented and verified without reopening the already-closed P0 scope.

## What Changes

- Recover export-center query, retry, download, and page-init paths as a dedicated operational follow-up batch.
- Recover audit page reachability, filter-query, and detail-read flows as a dedicated operational follow-up batch.
- Preserve deterministic failure semantics so authorization, route/resource, and business-query failures remain distinguishable.
- Add targeted regression and smoke coverage before this operational follow-up batch is considered complete.

## Capabilities

### Modified Capabilities
- `export-center-management`: tighten export-center query, retry, download, and failure-semantics expectations for operational recovery
- `audit-logs`: tighten audit page reachability, query initialization, detail-read, and failure-semantics expectations for operational recovery

## Impact

- **Frontend**: export-center and audit pages, route-entry and page-init logic, related API callers
- **Backend Go**: export-center and audit handlers/services/repositories plus any compatibility or middleware wiring required for recovered flows
- **Verification**: targeted regression tests, smoke coverage, and release-readiness evidence for the operational follow-up batch
