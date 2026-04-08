## Context

The current visualization surface is only partially aligned with the governed permission model.

- API-level visualization registration in `apps/backend-go/internal/transport/http/router.go` already protects `findById`, `updateCanvas`, and `deleteLogic` with `CheckDashboardView()` / `CheckDashboardEdit()`, but leaves `saveCanvas` and non-listed write paths outside resource-level enforcement.
- Root legacy visualization routes are still registered through `handler.RegisterVisualizationRoutes(r.engine.Group(""), r.visualHandler)`, which exposes additional write aliases such as `/dataVisualization/save`, `/copy`, `/updateBase`, `/move`, `/updatePublishStatus`, and `/recoverToPublished` without the governed middleware wiring used on selected `/api` routes.
- Visualization requests do not all identify authorization targets the same way. `UpdateRequest`, `MoveRequest`, and `DetailRequest` carry an existing visualization `id`, while `SaveRequest` and `CopyRequest` rely on `pid` plus visualization `type`, and `CopyRequest` also references a source visualization `id`.
- The service layer already distinguishes dashboard vs. big-screen resources through `normalizeVisualizationResourceType()` in `apps/backend-go/internal/service/visualization_service.go`, and new resources already inherit parent permissions on create when a governed parent exists.
- The generic permission middleware can extract `id` and `pid`, but it cannot safely infer which one is the governing authorization target for every visualization write operation, nor can it distinguish dashboard vs. screen semantics from request shape alone.

This change needs to strengthen visualization write-route authorization without broadening into discovery/list flows, redesigning the permission model, or accidentally governing every legacy visualization alias in one sweep.

## Goals / Non-Goals

**Goals:**
- Make in-scope visualization write routes fail closed on missing or insufficient governed authorization instead of relying on Auth-only protection.
- Preserve the existing dashboard/screen resource model by enforcing visualization writes against `dashboard` or `screen` permissions already present in the permission domain.
- Keep the route rollout small and independently mergeable by targeting only write paths whose authorization target can be resolved safely.
- Cover both API and root legacy aliases when they represent the same governed write operation, so legacy write paths do not remain silently weaker than canonical ones.
- Add focused backend verification for authorized, denied, and unresolved-target cases on visualization write routes.

**Non-Goals:**
- Do not redesign visualization CRUD, tree/list behavior, or the broader visualization lifecycle.
- Do not govern read/discovery routes such as `tree`, `list`, `nameCheck`, `checkCanvasChange`, `findDvType`, or `updateCheckVersion` in this change.
- Do not broaden this change into datasource, share, export, or resource-group inheritance workstreams.
- Do not duplicate business logic already owned by `VisualizationService` or reimplement permission inheritance in the handler layer.
- Do not force a new create-permission model for root-level visualization creation without an existing governed parent scope.

## Decisions

### 1. Visualization write authorization remains dashboard/screen-resource-governed

**Decision:** Visualization write routes will authorize against existing `dashboard` and `screen` resource types rather than introducing a new visualization-specific permission type.

**Rationale:** The codebase already encodes visualization governance this way. `normalizeVisualizationResourceType()` maps `dashboard` to `ResourceTypeDashboard` and `dataV` to `ResourceTypeScreen`, and parent-permission inheritance on create already depends on that distinction. Reusing the current resource model keeps route enforcement aligned with the service and repository layer.

**Alternative considered:** Treat all visualization writes as dashboard-only permission checks. Rejected because it would silently mis-govern big-screen (`dataV`) resources and drift from the service-layer resource model already present in tests and backfill logic.

### 2. Add a visualization-aware write gate instead of relying on generic request ID extraction

**Decision:** The route layer should use visualization-aware authorization middleware (or equivalent route-specific resolution helpers) that can determine the governing target for each write operation, rather than reusing generic `extractResourceID()` semantics unchanged.

**Rationale:** Generic extraction collects both `id` and `pid` from request bodies and cannot know whether a write route should authorize against the existing visualization resource, the destination parent resource, or both. Visualization-aware resolution keeps the contract explicit and prevents `saveCanvas`/`copy` flows from accidentally checking the wrong identifier.

**Alternative considered:** Reuse `CheckDashboardEdit()` everywhere and let `extractResourceID()` pick the first body field it finds. Rejected because it is ambiguous for `saveCanvas`/`copy`, ignores screen-vs-dashboard routing, and risks authorizing the wrong target.

### 3. Split write-route enforcement by authorization target type

**Decision:** The design will distinguish two classes of visualization writes:

- **Existing-resource edit operations**: routes whose request identifies the visualization being changed (`updateBase`, `move`, `updatePublishStatus`, `recoverToPublished`, and any already-governed updates such as `updateCanvas` / `deleteLogic`). These should authorize against the existing visualization resource ID.
- **Parent-scoped create operations**: routes whose request creates a new visualization beneath a parent (`saveCanvas`, alias `save`) and can only be governed when a positive `pid` plus visualization type is available. These should authorize against the parent governed resource scope.

`copy` is intentionally treated as a special case because it references both a source visualization and an optional destination parent.

**Rationale:** Existing-resource edits map cleanly to the current permission model. Parent-scoped creation already has a natural governed boundary because the service inherits permissions from the parent. Treating them separately keeps the first rollout explicit and testable.

**Alternative considered:** Force every visualization write route into a single “resource edit” rule keyed only by `id`. Rejected because create paths do not have an existing resource ID and would either fail incorrectly or fall back to ambiguous request parsing.

### 4. Defer `copy` until dual-target authorization semantics are explicit

**Decision:** The first rollout should not treat `copy` as just another single-target edit route. Instead, the design keeps `copy` as an explicit follow-up within the same change unless the spec phase proves a single safe authorization rule.

**Rationale:** `CopyRequest` includes a source visualization `id` and an optional destination `pid`. A safe governed contract likely needs to reason about both the source resource and the target parent scope. Folding that into the same first slice as ordinary edits would increase risk and blur the smallest independently mergeable rollout.

**Alternative considered:** Check only the source visualization edit permission for `copy`. Rejected because it can still allow copying into an unauthorized parent scope. Check only the destination parent. Rejected because it can still expose unauthorized source resources.

### 5. Roll out API and root legacy aliases only for explicitly selected write paths

**Decision:** When a write operation exists on both canonical `/api/dataVisualization/*` and root legacy `/dataVisualization/*` paths, the rollout should align both entry points for that operation. But routes outside the selected write set should remain untouched.

**Rationale:** The chart-route work already established that leaving root legacy aliases weaker than API routes creates an avoidable governance gap. At the same time, root visualization registration currently bundles many unrelated routes together, so the safest approach is route-level or split-registration alignment for only the chosen write aliases.

**Alternative considered:** Wrap the entire root visualization group in governed middleware. Rejected because it would accidentally broaden scope to list/tree/helper routes and make rollback much harder.

### 6. Preserve service-layer behavior and use middleware only as the authorization boundary

**Decision:** Middleware should only resolve authorization targets, check permission, and fail closed. It should not move visualization mutation behavior, parent inheritance, or type normalization out of `VisualizationService`.

**Rationale:** The service layer already owns create/update/copy/move semantics and governed inheritance behavior. Keeping permission checks at the boundary while preserving service logic minimizes diff size and follows the same pattern used in chart-route middleware alignment.

**Alternative considered:** Push authorization and visualization-type resolution deeper into each handler or service method. Rejected because it duplicates route-boundary concerns and makes the rollout harder to verify consistently across aliases.

## Risks / Trade-offs

- **[Risk] Parent-scoped creation without a governed `pid` may not have a safe authorization target** → **Mitigation:** explicitly scope first-rollout creation enforcement to requests with a positive parent ID, and leave rootless creation semantics as an open question/spec decision instead of guessing.
- **[Risk] Root legacy visualization registration bundles many unrelated routes together** → **Mitigation:** use split registration or route-level middleware only for the selected write aliases, and add router contract tests proving discovery routes were not broadened.
- **[Risk] Dashboard and big-screen writes may require different resource-type resolution than the current API-level dashboard-only checks** → **Mitigation:** reuse the existing `normalizeVisualizationResourceType()` semantics and add denial-path tests for both dashboard and `dataV` request shapes.
- **[Risk] `copy` can become a scope trap because it references both source and destination targets** → **Mitigation:** keep `copy` explicitly separated in design/specs and only include it in the first implementation slice if a single, testable governed rule is agreed.

## Migration Plan

1. Add delta specs for `permission-config` and `visualization-management` that define governed authorization expectations for in-scope visualization write routes.
2. Introduce visualization-aware permission-gating support at the route boundary for existing-resource edits and parent-scoped creation flows.
3. Attach the gate to the first selected visualization write routes on `/api` and matching root legacy aliases without broadening into list/tree/helper paths.
4. Add router/handler regression coverage for allowed, denied, and unresolved-target cases, including dashboard vs. `dataV` request shapes where applicable.
5. If regressions appear, revert the new write-route wiring while preserving the authorization tests that expose the gap, then narrow the rollout to a smaller subset of write operations.

## Open Questions

- Should `saveCanvas` without a positive `pid` remain outside the first governed rollout until the product has an explicit create-permission rule?
- Can `copy` be safely included in the first slice, or should it remain a follow-up because it requires both source-resource and destination-parent authorization semantics?
- Should the first implementation slice update only `/api/dataVisualization/saveCanvas` plus the root legacy write aliases, or should the API route surface be expanded first and root legacy parity follow in a second PR?
