## 1. Backend canonical datasource file ingest routes

- [x] 1.1 Add canonical datasource handlers for `uploadFile` and `loadRemoteFile` in `apps/backend-go/internal/transport/http/handler/datasource_handler.go`, reusing existing ingest behavior and preserving compatibility-safe envelopes.
- [x] 1.2 Register `POST /api/ds/uploadFile` and `POST /api/ds/loadRemoteFile` in canonical datasource routing without removing `/datasource/uploadFile` and `/datasource/loadRemoteFile` compatibility aliases.
- [x] 1.3 Add backend regression coverage for canonical file-ingest routes, including success envelopes and explicit failure semantics for invalid multipart input, invalid remote source, or unavailable backend conditions.

## 2. Frontend datasource API cutover for file ingest

- [x] 2.1 Update `apps/frontend/src/api/datasource.ts` so only `uploadFile` and `loadRemoteFile` switch to canonical `/api/ds/*` routes while keeping wrapper names, multipart handling, request shapes, and response contracts unchanged.
- [x] 2.2 Keep non-scoped datasource APIs on compatibility paths in this change and make the migration boundary explicit in tests or notes where needed.
- [x] 2.3 Add or update frontend datasource API regression tests to assert canonical URLs for the two migrated wrappers and verify compatibility-safe wrapper behavior remains unchanged.

## 3. Verification and rollout safety

- [x] 3.1 Run backend verification for file-ingest canonical routes (affected Go tests and `make test`) and resolve any canonical/compat contract mismatches.
- [x] 3.2 Run frontend verification (`npm run lint`, `npm run ts:check`, and affected datasource API tests) after URL cutover.
- [x] 3.3 Smoke-test executable datasource file-ingest flows in local/dev environment and record environment limitations or compatibility-only caveats in final notes.
