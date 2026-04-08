## 1. Remaining root visualization route rollout

- [x] 1.1 Wire `/dataVisualization/updateBase`, `/dataVisualization/move`, `/dataVisualization/updatePublishStatus`, and `/dataVisualization/recoverToPublished` through the existing visualization-aware governed edit middleware without broadening adjacent visualization list, tree, helper, create, or copy routes.
- [x] 1.2 Confirm the selected routes safely resolve their governed visualization target by existing resource `id`, and fail closed rather than falling back to Auth-only behavior if a selected route cannot establish that target.

## 2. Focused regression coverage

- [x] 2.1 Add backend middleware and/or handler tests covering allowed, denied, missing-auth, and unresolved-target behavior for the newly governed root visualization mutation routes.
- [x] 2.2 Add router contract coverage proving the four selected root legacy routes now require governed authorization while adjacent visualization routes remain outside this slice.

## 3. Verification and scope control

- [x] 3.1 Run focused backend tests for affected visualization middleware, handler, and router packages and confirm success for allowed, denied, missing-auth, and unresolved-target scenarios.
- [x] 3.2 Run `make test` in `apps/backend-go` and fix any failures caused by this change without broadening scope into unrelated governance work.
- [x] 3.3 Run `TEST_DB_HOST=127.0.0.1 make test-integration` in `apps/backend-go` if the final route wiring changes permission-sensitive runtime behavior beyond isolated unit coverage, and record any environment blockers separately if it cannot be completed.
- [x] 3.4 Perform a final route-scope review confirming this slice only governs the four selected root visualization mutation routes and does not implicitly expand into API visualization routes, create/copy flows, or discovery/list/helper endpoints.
