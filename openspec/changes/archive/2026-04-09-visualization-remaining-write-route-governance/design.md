## Context

The previous visualization write-route alignment closed the highest-risk governed gaps for `findById`, `updateCanvas`, `deleteLogic`, `save`, `saveCanvas`, and `copy`, but four root legacy visualization mutation routes remain outside governed middleware wiring.

- `apps/backend-go/internal/transport/http/handler/visualization_handler.go` still registers `/dataVisualization/updateBase`, `/move`, `/updatePublishStatus`, and `/recoverToPublished` without visualization-aware permission middleware.
- These routes all mutate an existing visualization resource rather than creating a new one, so their authorization target is the same class already handled by `CheckVisualizationEdit()`.
- The existing visualization-aware middleware already resolves dashboard vs. big-screen (`screen`) resource type by visualization `id` and already has regression coverage for allowed, denied, missing-auth, and unresolved-target behavior.
- The remaining gap is therefore not a new permission-model problem; it is a rollout-completeness gap on the root legacy mutation surface.

This change should finish that rollout without broadening into API routes, visualization create/copy flows, discovery/list endpoints, or any broader permission-center redesign.

## Goals / Non-Goals

**Goals:**
- Govern the remaining root visualization mutation routes that target an existing visualization resource by `id`.
- Reuse the existing visualization-aware edit authorization path rather than introducing new middleware concepts.
- Add focused backend verification proving those routes now fail closed on missing or insufficient governed authorization.
- Keep the slice independently mergeable and easy to roll back.

**Non-Goals:**
- Do not change API route behavior in this change.
- Do not revisit `save`, `saveCanvas`, `copy`, `findById`, `updateCanvas`, or `deleteLogic`.
- Do not broaden into list/tree/helper routes.
- Do not redesign visualization service behavior or permission inheritance.

## Decisions

### 1. Reuse existing-resource visualization edit gating unchanged

**Decision:** The remaining root write routes will use the already-established visualization-aware existing-resource edit middleware (`CheckVisualizationEdit()` or its exact equivalent) without introducing route-specific authorization logic.

**Rationale:** `updateBase`, `move`, `updatePublishStatus`, and `recoverToPublished` all operate on an existing visualization resource identified by request `id`, which matches the authorization target class already used by `updateCanvas` and `deleteLogic`. Reusing the same middleware keeps semantics consistent and minimizes regression risk.

**Alternative considered:** Add route-specific checks for each of the four handlers. Rejected because it duplicates already-proven target-resolution behavior and would make future verification harder.

### 2. Scope remains root-legacy-only for this slice

**Decision:** This rollout will only touch the root legacy routes still lacking governed write enforcement.

**Rationale:** The remaining gap identified after the previous change is specifically on the root visualization registration surface. Pulling in API paths or adjacent visualization routes would broaden the slice without adding meaningful governance value for this follow-up.

**Alternative considered:** Re-scan and rewire all visualization routes again in one pass. Rejected because it would turn a small follow-up into a redundant wide-scope audit.

### 3. Regression coverage should prove route-scope completion, not just middleware behavior

**Decision:** Tests must verify both middleware semantics and root route wiring for the four selected paths.

**Rationale:** The main risk is not whether the middleware works — that has already been proven — but whether the remaining root routes are still accidentally left on Auth-only access. Route-level regression tests keep that gap explicit and prevent future drift.

**Alternative considered:** Add only middleware unit/integration tests. Rejected because they would not prove that the route registration actually changed.

## Risks / Trade-offs

- **[Risk] A remaining root route may use a request shape that does not safely resolve through existing-resource edit middleware** → **Mitigation:** verify request contracts before implementation and fail closed if a route cannot safely resolve a governed target.
- **[Risk] Root visualization registration still bundles many unrelated handlers together** → **Mitigation:** apply route-level middleware only to the four selected paths and keep router contract tests proving adjacent routes remain untouched.
- **[Risk] The slice may look small enough to skip tests** → **Mitigation:** require focused middleware and router regression coverage even though the logic mostly reuses existing behavior.

## Migration Plan

1. Add delta specs defining governed authorization expectations for the four remaining root visualization mutation routes.
2. Wire those routes through the existing visualization-aware edit permission middleware.
3. Add focused backend tests for unauthenticated, denied, and allowed scenarios, plus route-contract coverage proving the selected root routes are no longer Auth-only.
4. If regressions appear, revert only the new route wiring and tests for this slice without disturbing the already-merged visualization write-route changes.

## Open Questions

- Do all four selected routes accept an existing visualization `id` in a shape already supported by the current visualization edit middleware, or does one of them require a small resolver extension?
