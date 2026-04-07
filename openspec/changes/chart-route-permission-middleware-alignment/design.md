## Context

The current permission-governance work already made part of chart runtime behavior permission-aware at the service layer, but the HTTP entry points are still inconsistent.

- Canonical chart routes are registered under `RegisterChartRoutes()` and currently expose `/chart/query` and `/chart/data`, with `ChartHandler.Data()` choosing between `QueryDataWithPermission()` and permissive `QueryData()` based only on whether `GetUserID(c)` is non-zero.
- Compatibility routes under `RegisterCompatibilityBridgeRoutes()` expose additional chart runtime paths such as `/chartData/getData`, `/chart/getData`, and `/chart/listByDQ/:id/:chartId`, and some of those flows still rely on compatibility helpers that fall back to synthetic admin identity when request context is missing.
- The permission middleware stack today is dataset-oriented. `PermissionMiddleware.CheckDatasetView()` checks dataset resource permission directly from request/context IDs, and `RowPermissionMiddleware()` is already fail-closed but expects dataset IDs to be present before it runs.
- The chart permission model is still dataset-governed. There is no separate `chart` resource type in the permission domain; `ChartService.QueryDataWithPermission()` already resolves `chartID -> datasetGroupID` and applies row/column rules against dataset-governed permission semantics.

This change needs to make chart runtime entry points explicit governed paths at the HTTP boundary without inventing a new resource model, duplicating row/column rule logic, or accidentally broadening scope into general chart CRUD.

## Goals / Non-Goals

**Goals:**
- Make governed chart runtime routes fail closed at the HTTP boundary instead of silently falling back to permissive execution.
- Keep chart authorization dataset-governed by resolving chart requests to `datasetGroupID` before permission-aware execution continues.
- Reuse the existing permission middleware and `RowPermissionMiddleware()` context contract where possible instead of creating a parallel chart-only enforcement stack.
- Remove synthetic admin fallback from in-scope compatibility chart permission flows.
- Preserve the existing service-layer permission logic as the source of truth for row/column enforcement.

**Non-Goals:**
- Do not introduce a new `chart` resource type or redesign the permission model.
- Do not broaden this change to chart CRUD, visualization routes, share/export flows, or unrelated compatibility APIs.
- Do not move row/column SQL compilation out of `ChartService`, `RowPermissionService`, or `ColumnPermissionService`.
- Do not globally rewrite every compatibility helper; only chart-governed permission flows are in scope.
- Do not treat `/chart/query` metadata retrieval as part of the first governed runtime rollout unless later specs require it.

## Decisions

### 1. Chart runtime permission remains dataset-governed, not chart-resource-governed

**Decision:** The design will continue to authorize chart runtime requests against dataset view permission, using the dataset group behind the chart as the governed resource identity.

**Rationale:** The existing permission domain exposes dataset/dashboard/screen/datasource resource types, but not a dedicated chart resource type. `ChartService.QueryDataWithPermission()` already resolves the backing dataset group and applies row/column rules there, so the safest alignment is to make the HTTP boundary explicit about the same dataset-governed contract.

**Alternative considered:** Introduce a new chart resource type and route chart requests through separate chart permission records. Rejected because it changes the permission model, expands the migration surface, and is unnecessary for the current governed runtime semantics.

### 2. Add a chart-specific dataset-resolution gate instead of overloading generic request-ID extraction

**Decision:** Introduce a chart runtime permission gate in the middleware layer that resolves `chartID -> datasetGroupID`, checks dataset view permission for that dataset group, and stores the resolved governed dataset identity into existing Gin context keys before the handler runs.

**Rationale:** Generic `CheckDatasetView()` assumes the request already carries dataset IDs in route params or JSON. Canonical chart data requests only send chart `id`, and compatibility chart routes often carry a mix of chart IDs and dataset IDs in path params. A dedicated chart-aware gate keeps the resolution logic explicit and lets downstream code reuse `DatasetIDKey`, `ResourceIDKey`, and row-permission context handoff instead of depending on fragile payload heuristics.

**Alternative considered:** Extend `extractResourceID()` with chart-specific guessing logic. Rejected because it would make dataset-oriented middleware implicitly depend on chart payload shapes and increase ambiguity across unrelated routes.

### 3. Reuse `RowPermissionMiddleware()` after dataset resolution instead of duplicating row-permission logic for charts

**Decision:** Once chart middleware has resolved and stored the governed dataset identity, the route stack should reuse `RowPermissionMiddleware()` so chart runtime requests participate in the same fail-closed row-permission context establishment as dataset permission-aware routes.

**Rationale:** `RowPermissionMiddleware()` already enforces authenticated-user presence and requires a dataset identity before execution continues. If chart middleware seeds the dataset context first, the existing row-permission middleware can remain the normalization layer and the chart service can remain the only place that compiles row/column rules into runtime query behavior.

**Alternative considered:** Build chart-specific row-permission middleware that resolves filters itself. Rejected because it duplicates existing semantics and risks drift from the already-correct service-layer enforcement.

### 4. Scope the first rollout to explicit permission-aware chart runtime entry points

**Decision:** The first rollout should only cover chart runtime routes that already have permission-aware service branches or represent governed data access: canonical `/api/chart/data`, compatibility `/chartData/getData`, compatibility `/chart/getData`, and compatibility `/chart/listByDQ/:id/:chartId`.

**Rationale:** These routes already express runtime data access semantics and either call `QueryDataWithPermission()` or `ListByDQWithPermission()` when a user context exists. Making them explicit governed routes closes the fail-open gap without turning chart CRUD or metadata endpoints into a larger authorization migration.

**Alternative considered:** Apply the new gate to every `/chart/*` and `/chartData/*` endpoint. Rejected because many of those handlers are CRUD, export, or helper flows with different resource semantics and would make this change too broad to merge independently.

### 5. Compatibility chart flows must stop synthesizing admin identity for governed authorization

**Decision:** In-scope compatibility chart permission flows must use the authenticated request context as-is and return explicit unauthorized / permission-denied behavior when that context is absent, rather than defaulting `getCurrentUserID()` or `getCurrentUsername()` to admin values.

**Rationale:** A synthetic admin fallback defeats the purpose of governed runtime enforcement and creates the exact fail-open behavior this change is meant to remove. Narrowing the fix to chart-governed flows keeps the change small while eliminating the highest-risk compatibility behavior.

**Alternative considered:** Keep admin fallback in compat handlers for backward compatibility. Rejected because it preserves a silent privilege-escalation path and leaves canonical and compatibility routes with materially different security semantics.

## Risks / Trade-offs

- **[Risk] Chart middleware may resolve the wrong dataset group for a chart request** → **Mitigation:** use the same chart-to-dataset lookup contract already required by `QueryDataWithPermission()` and add focused handler/middleware tests for chart ID resolution and error paths.
- **[Risk] Reusing `RowPermissionMiddleware()` could expose assumptions that only held for dataset routes** → **Mitigation:** seed dataset context before row middleware runs and add coverage for canonical and compatibility chart request shapes.
- **[Risk] Some legacy clients may begin receiving explicit authorization failures where they previously succeeded** → **Mitigation:** keep rollout scoped to permission-aware runtime routes, mark the behavior as breaking in proposal/specs, and preserve service-level enforcement so rollback is narrow.
- **[Risk] Touching compatibility helpers may accidentally affect unrelated compat endpoints** → **Mitigation:** avoid global helper rewrites and only route in-scope chart permission handlers through the new governed path.

## Migration Plan

1. Add a `permission-config` delta spec that defines governed chart runtime routes as dataset-governed middleware-protected entry points.
2. Introduce chart-aware middleware support that can resolve `chartID -> datasetGroupID`, check dataset view permission, and seed governed dataset context for downstream middleware.
3. Attach the new gate plus `RowPermissionMiddleware()` to the in-scope canonical and compatibility chart runtime routes.
4. Remove synthetic admin fallback from the in-scope compatibility chart permission flows and normalize missing-context behavior to fail closed.
5. Add focused backend tests for canonical and compatibility chart runtime success, denial, invalid chart ID, lookup failure, and missing-auth/context cases.
6. If regressions appear, roll back the new route wiring and compatibility fallback changes while keeping the tests that expose the gap, then narrow the governed route set before retrying.

## Open Questions

- Should `/datasetField/listByDatasetGroup/:datasetId` and `/datasetField/listWithPermissions/:datasetId` remain out of scope for this slice even though they also branch into permission-aware chart field listing?
- Is there any chart-related export route that should share the same governed dataset gate now, or should export remain a separate follow-up because it uses export-specific permission semantics?
- Should the chart-to-dataset resolver live inside `PermissionMiddleware` directly, or should it depend on a narrow service/repository interface to keep transport-layer wiring simpler in tests?
