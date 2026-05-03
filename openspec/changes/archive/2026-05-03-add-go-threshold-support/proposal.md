## Why

The Go backend already carries migrated threshold tables and generated GORM models, but it has no threshold repository, service, handler, or routes. That leaves legacy threshold workflows unavailable in Go even though the visualization migration work already depends on threshold-related persistence and users can still lose or block threshold-driven behavior when moving off the Java backend.

## What Changes

- Add first-phase Go threshold support for threshold definition CRUD, threshold listing, chart-linked lookup, switch enable/disable, and delete compatibility flows.
- Add a Go threshold preview/matching engine that can evaluate stored threshold rules against chart data without requiring the full legacy notification pipeline.
- Add threshold instance/history listing support backed by the existing `xpack_threshold_instance` table so threshold execution results can be inspected once the first slice is active.
- Keep scheduled dispatch, external notification channels, and report/data-filling work out of scope for this slice.

## Capabilities

### New Capabilities
- `threshold-management`: Manage threshold definitions, preview threshold matching results, and query threshold instance history in the Go backend.

### Modified Capabilities
- `visualization-management`: Extend visualization lifecycle compatibility so chart-linked threshold data remains a governed part of visualization behavior in Go.

## Impact

- Affected Go code: `apps/backend-go/internal/repository`, `internal/service`, `internal/transport/http/handler`, `internal/domain/auto`, and related visualization integration points.
- Affected APIs: new Go-side threshold endpoints aligned with legacy `/threshold/*` behavior for the first supported slice.
- Affected persistence: existing `xpack_threshold_info`, `xpack_threshold_instance`, and visualization-related threshold linkage records.
- Dependencies: chart/visualization data access, existing generated threshold models, and compatibility expectations from legacy Java threshold workflows.
- Breaking changes: none intended; this slice adds missing Go functionality rather than changing existing public Go behavior.
- Rollback: disable or remove new threshold route registration and service wiring while leaving existing visualization and audit behavior unchanged.
