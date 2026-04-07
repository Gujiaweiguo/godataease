## 1. Chart-governed middleware contract

- [x] 1.1 Add chart-aware permission middleware support that resolves `chartID -> datasetGroupID`, validates dataset-governed view permission, and seeds governed dataset context for downstream middleware and handlers.
- [x] 1.2 Add or update middleware-focused backend tests covering canonical and compatibility chart request shapes, including successful dataset-context resolution, invalid chart identity, unresolved dataset context, and unauthenticated fail-closed behavior.

## 2. Governed chart route rollout

- [x] 2.1 Wire the approved governed chart runtime routes (`/api/chart/data`, `/chartData/getData`, `/chart/getData`, `/chart/listByDQ/:id/:chartId`) through the new chart-aware permission gate plus `RowPermissionMiddleware()` without broadening enforcement to chart CRUD or unrelated compatibility endpoints.
- [x] 2.2 Update in-scope handlers and compatibility-bridge flows so governed chart runtime requests no longer fall back to permissive execution or synthetic admin identity when authenticated user context is missing.
- [x] 2.3 Add or update handler/service regression tests proving governed chart runtime flows remain dataset-governed, preserve row/column permission behavior, and return explicit denial or error semantics instead of fail-open results.

## 3. Verification and scope control

- [x] 3.1 Run focused backend tests for affected middleware, handler, and service packages and confirm success for allowed, denied, missing-auth, and chart-context-resolution failure scenarios.
- [x] 3.2 Run `make test` in `apps/backend-go` and address any failures caused by this change without broadening scope into unrelated permission work.
- [x] 3.3 Run `TEST_DB_HOST=127.0.0.1 make test-integration` in `apps/backend-go` if the final implementation changes permission-sensitive runtime behavior beyond isolated unit coverage, and record any environment blockers separately if it cannot be completed.
- [x] 3.4 Perform a final route-scope review confirming this change governs only the chart runtime entry points defined in the spec/design and does not expand into chart CRUD, visualization/share flows, export-specific permission work, or whitelist/system-variable implementation.
