## 1. Backend canonical datasource table exploration routes

- [x] 1.1 Add canonical datasource table exploration handlers for `tables`, `tableStatus`, `tableField`, and `schema` in `apps/backend-go/internal/transport/http/handler/datasource_handler.go` by reusing the existing `DatasourceService` capabilities and preserving compatibility-safe response envelopes.
- [x] 1.2 Register `POST /api/ds/tables`, `POST /api/ds/tableStatus`, `POST /api/ds/tableField`, and `POST /api/ds/schema` in `apps/backend-go/internal/transport/http/router.go` without removing the existing `/datasource/getTables`, `/datasource/getTableStatus`, `/datasource/getTableField`, and `/datasource/getSchema` compatibility routes.
- [x] 1.3 Add backend regression coverage proving the canonical table exploration routes preserve expected success envelopes and explicit invalid datasource / invalid table / missing status evidence semantics.

## 2. Frontend datasource API cutover

- [x] 2.1 Update `apps/frontend/src/api/datasource.ts` so only the four table exploration wrappers (`getTables`, `getTableStatus`, `getTableField`, `getSchema`) switch to canonical `/api/ds/*` routes while keeping wrapper names, request shapes, and return contracts unchanged.
- [x] 2.2 Keep non-scoped datasource APIs (`previewData`, `syncApi*`, upload, remote file, and other untouched routes) on compatibility paths in this change and make the migration boundary explicit in code or tests where needed.
- [x] 2.3 Add or update frontend datasource API regression tests to assert the canonical URLs for the four migrated wrappers and verify compatibility-safe wrapper behavior remains unchanged.

## 3. Verification and rollout safety

- [x] 3.1 Run backend verification for datasource table exploration canonical routes, including affected Go handler/router tests and `make test`, and resolve any canonical versus compatibility contract mismatches.
- [x] 3.2 Run frontend verification for the datasource API cutover, including `npm run lint`, `npm run ts:check`, and the affected datasource API tests.
- [x] 3.3 Smoke-test datasource page table exploration flows against the canonical cutover (table list load, schema fetch, field inspection, and executable status checks) and record any environment-specific limitations or compatibility-only caveats in the final notes.
