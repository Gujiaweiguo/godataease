## 1. Visualization-aware authorization target resolution

- [x] 1.1 Add visualization-aware permission-gating support that resolves existing-resource edit targets by visualization `id`, preserves dashboard vs. `screen` (`dataV`) resource typing, and fails closed when the governed target cannot be determined safely.
- [x] 1.2 Add focused middleware and route-level backend tests covering authorized, denied, invalid-target, and missing-auth cases for visualization write requests that target an existing visualization resource.

## 2. Existing-resource write route rollout

- [x] 2.1 Wire the first existing-resource visualization routes through visualization-aware governed authorization on both canonical and matching root legacy aliases for `findById`, `updateCanvas`, and `deleteLogic`, without broadening into list/tree/helper routes or later write-route slices.
- [x] 2.2 Add or update handler/router regression tests proving those in-scope routes and the root `deleteLogic/:id/:busiFlag` alias now return explicit governed auth-required or permission-error semantics instead of relying on Auth-only access.

## 3. Parent-scoped create rollout and copy scope decision

- [x] 3.1 Add parent-scoped authorization support for `saveCanvas` and its legacy-compatible `save` alias when a positive governed `pid` is present, and fail closed when the route cannot establish a safe governed parent target.
- [x] 3.2 Add backend tests covering authorized parent-scoped creation, denied parent-scoped creation, and unresolved-parent failure semantics for dashboard and `dataV` request shapes.
- [x] 3.3 Resolve the governed authorization contract for `copy`: either implement a safe dual-target rule with focused tests, or explicitly narrow the change scope by updating the artifacts before implementation proceeds past this point.

## 4. Verification and scope control

- [x] 4.1 Run focused backend tests for affected middleware, router, and visualization handler/service packages and confirm success for allowed, denied, missing-auth, and unresolved-target scenarios.
- [x] 4.2 Run `make test` in `apps/backend-go` and fix any failures caused by this change without broadening scope into unrelated governance work.
- [x] 4.3 Run `TEST_DB_HOST=127.0.0.1 make test-integration` in `apps/backend-go` if the final write-route rollout changes permission-sensitive runtime behavior beyond isolated unit coverage, and record any environment blockers separately if it cannot be completed.
- [x] 4.4 Perform a final route-scope review confirming this change governs only the selected visualization write routes and does not implicitly expand into visualization discovery/list flows, datasource/share/export work, or unrelated permission-center gaps.
