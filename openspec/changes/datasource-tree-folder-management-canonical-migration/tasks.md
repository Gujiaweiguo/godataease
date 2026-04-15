## 1. Backend canonical datasource tree/folder management routes

- [ ] 1.1 Add canonical datasource handlers for `Move`, `Rename`, and `CreateFolder` in `apps/backend-go/internal/transport/http/handler/datasource_handler.go`, reusing existing tree/folder management behavior and preserving compatibility-safe envelopes.
- [ ] 1.2 Register `POST /api/ds/move`, `POST /api/ds/reName`, and `POST /api/ds/createFolder` in canonical datasource routing without removing `/datasource/move`, `/datasource/reName`, and `/datasource/createFolder` compatibility aliases.
- [ ] 1.3 Add backend regression coverage for canonical tree/folder management routes, including success envelopes and explicit failure semantics for invalid move target, rename conflict, duplicate folder name, or unavailable backend conditions.

## 2. Frontend datasource API cutover for tree/folder management and canonical gaps

- [ ] 2.1 Update `apps/frontend/src/api/datasource.ts` so `move`, `reName`, and `createFolder` switch to canonical `/api/ds/*` routes while keeping wrapper names, request shapes, and response contracts unchanged.
- [ ] 2.2 Update `apps/frontend/src/api/dataset.ts` so `tree` switches from `/datasource/tree` to `/ds/tree` and `getTables` switches from `/datasource/getTables` to `/ds/tables`, both already having existing canonical backend routes.
- [ ] 2.3 Update `apps/frontend/src/api/datasource.ts` so `validate` (POST) switches from `/datasource/validate` to `/ds/validate`, the canonical route already existing in backend.
- [ ] 2.4 Keep non-scoped datasource APIs on compatibility paths in this change and make the migration boundary explicit in tests or notes where needed.
- [ ] 2.5 Add or update frontend datasource API regression tests to assert canonical URLs for the six migrated wrappers (move, reName, createFolder, tree, getTables, validate POST) and verify compatibility-safe wrapper behavior remains unchanged.

## 3. Verification and rollout safety

- [ ] 3.1 Run backend verification for tree/folder management canonical routes (affected Go tests and `make test`) and resolve any canonical/compat contract mismatches.
- [ ] 3.2 Run frontend verification (`npm run lint`, `npm run ts:check`, and affected datasource/dataset API tests) after URL cutover.
- [ ] 3.3 Smoke-test executable datasource tree/folder management flows (move, rename, create folder) and dataset datasource selection flows (tree, getTables, validate) in local/dev environment and record environment limitations or compatibility-only caveats in final notes.
