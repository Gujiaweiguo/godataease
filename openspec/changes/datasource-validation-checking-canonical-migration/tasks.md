## 1. Backend canonical datasource validation/checking routes

- [x] 1.1 Add canonical datasource handlers for `ValidateByID`, `CheckRepeat`, and `CheckAPIDatasource` in `apps/backend-go/internal/transport/http/handler/datasource_handler.go`, reusing existing validation/checking behavior and preserving compatibility-safe envelopes.
- [x] 1.2 Register `GET /api/ds/validate/:id`, `POST /api/ds/checkRepeat`, and `POST /api/ds/checkApiDatasource` in canonical datasource routing without removing `/datasource/validate/:id`, `/datasource/checkRepeat`, and `/datasource/checkApiDatasource` compatibility aliases.
- [x] 1.3 Add backend regression coverage for canonical validation/checking routes, including success envelopes and explicit failure semantics for invalid ID, duplicate name/type, invalid API datasource, or unavailable backend conditions.

## 2. Frontend datasource API cutover for validation/checking

- [x] 2.1 Update `apps/frontend/src/api/datasource.ts` so `validateById`, `checkRepeat`, and `checkApiItem` switch to canonical `/api/ds/*` routes while keeping wrapper names, request shapes, and response contracts unchanged.
- [x] 2.2 Update `apps/frontend/src/views/visualized/data/datasource/form/ApiHttpRequestDraw.vue` to change all `cancelMap` keys from `/datasource/checkApiDatasource` to `/api/ds/checkApiDatasource`.
- [x] 2.3 Keep non-scoped datasource APIs on compatibility paths in this change and make the migration boundary explicit in tests or notes where needed.
- [x] 2.4 Add or update frontend datasource API regression tests to assert canonical URLs for the three migrated wrappers and verify compatibility-safe wrapper behavior remains unchanged.

## 3. Verification and rollout safety

- [x] 3.1 Run backend verification for validation/checking canonical routes (affected Go tests and `make test`) and resolve any canonical/compat contract mismatches.
- [x] 3.2 Run frontend verification (`npm run lint`, `npm run ts:check`, and affected datasource API tests) after URL cutover.
- [x] 3.3 Smoke-test executable datasource validation and checking flows in local/dev environment and record environment limitations or compatibility-only caveats in final notes.
