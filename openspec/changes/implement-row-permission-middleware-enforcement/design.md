## Context

The current permission-governance work already corrected row/column permission semantics in the core service-layer runtime entry points:

- `DatasetService.PreviewWithPermission` builds row-permission WHERE clauses and column-permission masking before preview results are returned
- `ChartService.QueryDataWithPermission` and `ChartService.ListByDQWithPermission` apply row/column permission logic and now fail closed on permission lookup errors

However, the HTTP middleware that is supposed to represent governed row-permission enforcement is still a no-op stub:

- `apps/backend-go/internal/transport/http/middleware/permission.go` defines `RowPermissionMiddleware()` but only logs a warning and calls `c.Next()`
- router wiring already treats `/dataset/previewWithPerm` and other governed endpoints as permission-aware entry points, but only dataset view permission is enforced there today (`apps/backend-go/internal/transport/http/router.go`)

This creates a split-brain model: service-layer permission logic is real, while the middleware boundary that should normalize enforcement and failure semantics is still declarative only. The next change needs to close that gap without duplicating SQL-building logic or regressing the already-correct service enforcement paths.

## Goals / Non-Goals

**Goals:**
- Turn `RowPermissionMiddleware()` into real runtime enforcement infrastructure instead of a warning-only placeholder.
- Define a clear split of responsibility between middleware and existing dataset/chart service permission logic.
- Ensure middleware-protected routes fail closed when row-permission context cannot be resolved.
- Normalize denial/error semantics at the HTTP boundary so governed runtime routes do not silently bypass row-permission enforcement.

**Non-Goals:**
- Do not re-implement SQL rule compilation already handled by `RowPermissionService`, `ColumnPermissionService`, `DatasetService`, or `ChartService`.
- Do not broaden into whitelist persistence, system-variable support, or broader P2 permission-center expansion.
- Do not redesign the general permission middleware stack for unrelated menu/resource/dashboard authorization.
- Do not change non-governed preview/query routes that are intentionally outside the permission-aware runtime surface.

## Decisions

### 1. Middleware establishes and validates governed row-permission context; services remain the SQL enforcement source of truth

**Decision:** `RowPermissionMiddleware()` should not build SQL filters itself. Instead, it should validate that the request is entering a governed row-permission runtime path, resolve required context (dataset/group identity, authenticated user presence, and any prerequisites needed for permission evaluation), and fail closed when that context cannot be established. Service-layer code remains the only place that compiles and applies row/column rules to query behavior.

**Rationale:** The row/column services already contain the correct logic for select-column shaping, WHERE clause building, disabled columns, and masking. Duplicating that logic in middleware would create semantic drift and make future rule changes unsafe.

**Alternative considered:** Move SQL condition generation into middleware. Rejected because middleware is too early in the request lifecycle and should not own repository/query construction details.

### 2. Middleware-enforced routes must be explicit and scoped

**Decision:** Only the governed runtime entry points that are meant to honor row-permission semantics should adopt real `RowPermissionMiddleware()` wiring. The first-class targets are the permission-aware dataset/chart routes already modeled as governed runtime paths, not every route under `/dataset` or `/dataVisualization`.

**Rationale:** The router already distinguishes permission-aware routes like `/dataset/previewWithPerm`. The safest path is to attach real middleware to explicitly governed runtime entry points instead of inferring that all preview/query routes should become permission-aware.

**Alternative considered:** Apply row-permission middleware broadly to all dataset/chart routes. Rejected because that risks changing non-governed routes and expanding scope beyond the documented permission-config contract.

### 3. Middleware must fail closed on missing context or permission lookup preparation errors

**Decision:** If middleware cannot determine the governed dataset/group context, authenticated user, or any prerequisite needed to safely evaluate row-permission semantics, it must terminate the request with explicit permission/error semantics rather than continuing to a service path that might behave permissively.

**Rationale:** The recent service-layer fixes intentionally moved runtime permission behavior to fail-closed semantics. Leaving middleware fail-open would undermine those corrections and make the HTTP boundary less trustworthy than the service code behind it.

**Alternative considered:** Keep middleware observational and let services decide everything. Rejected because the stub itself is the documented gap this change is meant to close.

### 4. Middleware should communicate through request context, not custom response shaping logic

**Decision:** Middleware should store any resolved dataset/group identity or row-permission metadata in Gin context keys (expanding or reusing keys like `RowPermissionDatasetIDKey` / `RowPermissionFilterKey`) and allow handlers/services to consume them consistently.

**Rationale:** The file already declares row-permission context keys, implying the intended architecture was context handoff rather than direct response generation. Making that explicit keeps handlers/services loosely coupled to middleware while avoiding duplicated parsing.

**Alternative considered:** Have middleware fully short-circuit with transformed business responses. Rejected because handlers currently own response serialization and should continue to do so.

### 5. Existing service-layer runtime permission behavior remains the compatibility fallback during rollout

**Decision:** The rollout should preserve current service-layer enforcement while middleware is introduced. Middleware becomes an additional guardrail and context-establishment layer, not an immediate replacement for service checks.

**Rationale:** `PreviewWithPermission`, `QueryDataWithPermission`, and `ListByDQWithPermission` are already validated and covered by tests. Keeping them authoritative during rollout reduces regression risk and makes rollback straightforward.

**Alternative considered:** Remove service-layer permission logic once middleware is live. Rejected because it would enlarge the change and remove the most battle-tested enforcement path.

## Risks / Trade-offs

- **[Risk] Double enforcement or inconsistent semantics between middleware and services** → **Mitigation:** keep middleware limited to context validation/establishment and fail-closed gating, while services remain the rule application source of truth.
- **[Risk] Wrong dataset/group resolution at middleware level** → **Mitigation:** scope middleware only to routes whose identifiers and runtime contract are explicit, and add tests covering route-specific context extraction.
- **[Risk] Over-broad route attachment changes behavior for non-governed endpoints** → **Mitigation:** document and test the exact governed routes that adopt the middleware in this change.
- **[Risk] Previously pass-through clients now see deny/error responses** → **Mitigation:** mark this change as breaking in the proposal, preserve service-layer fallback during rollout, and keep denial/error semantics explicit.

## Migration Plan

1. Modify the `permission-config` delta spec to describe middleware/runtime enforcement more explicitly.
2. Implement real `RowPermissionMiddleware()` behavior for the smallest set of governed runtime routes.
3. Wire middleware context into the relevant handlers/services without moving SQL rule generation out of the services.
4. Add backend tests for middleware success, denial, and fail-closed error propagation.
5. Roll back by removing new middleware wiring and restoring the warning-only middleware if regressions appear, while keeping service-layer permission enforcement intact.

## Open Questions

- Which exact governed routes beyond `/dataset/previewWithPerm` should be part of the first rollout wave versus follow-up expansion?
- Should middleware resolve dataset-group identity directly from route payloads, or should handlers normalize and expose that context before middleware runs?
- Is there any compatibility bridge route that should share the same row-permission middleware behavior in the first rollout, or should compatibility routes remain service-enforced only until a later change?
