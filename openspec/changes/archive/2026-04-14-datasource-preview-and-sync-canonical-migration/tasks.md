## 1. Backend canonical datasource preview/sync routes

- [x] 1.1 Add canonical datasource handlers for `previewData`, `syncApiTable`, and `syncApiDs` in `apps/backend-go/internal/transport/http/handler/datasource_handler.go`, reusing existing service/domain behavior and preserving compatibility-safe envelopes.
- [x] 1.2 Register `POST /api/ds/previewData`, `POST /api/ds/syncApiTable`, and `POST /api/ds/syncApiDs` in canonical datasource routing without removing `/datasource/previewData`, `/datasource/syncApiTable`, and `/datasource/syncApiDs` compatibility aliases.
- [x] 1.3 Add backend regression coverage for canonical preview/sync routes, including success envelopes and explicit failure semantics for invalid datasource or unavailable backend conditions.

## 2. Frontend datasource API cutover for preview/sync

- [x] 2.1 Update `apps/frontend/src/api/datasource.ts` so only `previewData`, `syncApiTable`, and `syncApiDs` switch to canonical `/api/ds/*` routes while keeping wrapper names, request shapes, and response contracts unchanged.
- [x] 2.2 Keep non-scoped datasource APIs (upload, remote file, and other untouched endpoints) on compatibility paths in this change and make the migration boundary explicit in tests/notes.
- [x] 2.3 Add or update frontend datasource API regression tests to assert canonical URLs for the three migrated wrappers and verify compatibility-safe wrapper behavior remains unchanged.

## 3. Verification and rollout safety

- [x] 3.1 Run backend verification for preview/sync canonical routes (affected Go tests and `make test`) and resolve any canonical/compat contract mismatches.
- [x] 3.2 Run frontend verification (`npm run lint`, `npm run ts:check`, and affected datasource API tests) after URL cutover.
- [x] 3.3 Smoke-test executable datasource preview/sync flows in local/dev environment and record environment limitations or compatibility-only caveats in final notes.
