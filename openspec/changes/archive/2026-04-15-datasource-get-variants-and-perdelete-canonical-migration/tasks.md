## 1. Backend canonical datasource get variants and perDelete routes

- [ ] 1.1 Add canonical datasource handlers for `HidePw`, `Simple`, and `PerDelete` in `apps/backend-go/internal/transport/http/handler/datasource_handler.go`, reusing existing `service.GetByID` (with response transformation for hidePw and simple) and `service.PerDelete`, preserving compatibility-safe envelopes.
- [ ] 1.2 Register `GET /api/ds/hidePw/:id`, `GET /api/ds/simple/:id`, and `POST /api/ds/perDelete/:id` in canonical datasource routing without removing `/datasource/hidePw/:id`, `/datasource/getSimpleDs/:id`, and `/datasource/perDelete/:id` compatibility aliases.
- [ ] 1.3 Add backend regression coverage for canonical get variants and perDelete routes, including success envelopes and explicit failure semantics for non-existent ID, invalid ID format, and unmet deletion precondition scenarios.

## 2. Frontend datasource API cutover for get variants and perDelete

- [ ] 2.1 Update `apps/frontend/src/api/datasource.ts` so `getHidePwById` (~line 169) switches from `/datasource/hidePw/${id}` to `/api/ds/hidePw/${id}` while keeping wrapper name, request shape, and response contract unchanged.
- [ ] 2.2 Update `apps/frontend/src/api/datasource.ts` so `getSimpleDs` (~line 171) switches from `/datasource/getSimpleDs/${id}` to `/api/ds/simple/${id}` while keeping wrapper name, request shape, and response contract unchanged.
- [ ] 2.3 Update `apps/frontend/src/api/datasource.ts` so `perDelete` (~line 105) switches from `/datasource/perDelete/${id}` to `/api/ds/perDelete/${id}` while keeping wrapper name, request shape, and response contract unchanged.
- [ ] 2.4 Keep non-scoped datasource APIs on compatibility paths in this change and make the migration boundary explicit in tests or notes where needed.
- [ ] 2.5 Add or update frontend datasource API regression tests to assert canonical URLs for the three migrated wrappers (getHidePwById, getSimpleDs, perDelete) and verify compatibility-safe wrapper behavior remains unchanged.

## 3. Verification and rollout safety

- [ ] 3.1 Run backend verification for get variants and perDelete canonical routes (affected Go tests and `make test`) and resolve any canonical/compat contract mismatches.
- [ ] 3.2 Run frontend verification (`npm run lint`, `npm run ts:check`, and affected datasource API tests) after URL cutover.
- [ ] 3.3 Smoke-test executable datasource get variants (hidePw, getSimpleDs) and permanent delete flows in local/dev environment and record environment limitations or compatibility-only caveats in final notes.
