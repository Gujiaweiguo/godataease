## 1. Backend canonical datasource core CRUD routes

- [x] 1.1 Add canonical datasource tree, detail, save, update, and delete handlers in `apps/backend-go/internal/transport/http/handler/datasource_handler.go` by reusing existing `DatasourceService` methods.
- [x] 1.2 Register canonical `/api/ds/tree`, `/api/ds/:id`, `/api/ds/save`, `/api/ds/update`, and `/api/ds/delete/:id` routes in `apps/backend-go/internal/transport/http/router.go` without removing compatibility `/datasource/*` routes.
- [x] 1.3 Add backend handler regression coverage proving the canonical routes return compatibility-safe envelopes and explicit error semantics for missing/invalid requests.

## 2. Frontend datasource API cutover

- [x] 2.1 Update `apps/frontend/src/api/datasource.ts` so the five core CRUD/tree callers use canonical `/api/ds/*` routes while preserving existing wrapper names and response-shape expectations.
- [x] 2.2 Keep all non-core datasource routes (`getTables`, `getSchema`, `previewData`, `syncApi*`, upload, etc.) on compatibility paths in this change and document the boundary in tests or notes where needed.
- [x] 2.3 Add or update frontend datasource API regression tests to assert the canonical URLs for tree/detail/save/update/delete and verify compatibility-safe wrapper behavior.

## 3. Verification and rollout safety

- [x] 3.1 Run backend verification for datasource canonical core CRUD (`go test` on affected handler packages and `make test`) and resolve any canonical/compat envelope mismatches.
- [x] 3.2 Run frontend verification (`npm run lint`, `npm run ts:check`, and affected datasource API tests) after the canonical URL cutover.
- [x] 3.3 Manually smoke-test datasource page core flows against the canonical cutover (tree load, detail read, create/update/delete paths that are executable in the local/dev environment) and record any compatibility-only limitations in final notes.
