## Why

The archived recovery changes restored the primary BI and operational flows, but they intentionally left a bounded set of non-blocking hardening gaps behind. Those gaps are no longer broad breakages; they are narrower semantics and smoke-confidence issues around explicit forbidden behavior, missing-resource handling, and route-level operational coverage.

This change isolates those residual hardening tasks so they can be addressed without reopening the already-archived recovery batches.

## What Changes

- Preserve datasource forbidden semantics as a boundary-proof hardening result, while moving datasource runtime list identity design into a dedicated follow-up design change.
- Harden dashboard and big-screen route/detail semantics so missing-resource and detail-path behavior are validated closer to the frontend-facing boundary.
- Harden export-center download and route-level operational semantics beyond the already-fixed query/retry slice.
- Harden audit page route-level smoke so current query/detail semantics are exercised beyond unit and handler coverage.

## Capabilities

### Modified Capabilities
- `visualization-management`: add post-recovery hardening for detail-path missing-resource and deeper route-smoke coverage
- `export-center-management`: add post-recovery hardening for download-path semantics and route-level smoke
- `audit-logs`: add post-recovery hardening for route-level page-entry/detail smoke

## Impact

- **Frontend**: targeted route-smoke and page-entry verification for dashboard, big-screen, export-center, and audit
- **Backend Go**: narrow auth/detail semantics around visualization and export-center/audit route boundaries; datasource runtime list design is tracked separately in `design-datasource-list-resource-identity`
- **Verification**: hardening-oriented regression evidence layered on top of the already-completed recovery batches
