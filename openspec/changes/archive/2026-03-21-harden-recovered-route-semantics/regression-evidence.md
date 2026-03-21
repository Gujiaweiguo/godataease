# Regression Evidence

This document records concrete hardening evidence for `harden-recovered-route-semantics`.

## Datasource hardening evidence

### Forbidden alias semantics

- 2026-03-21: `cd apps/backend-go && go test -v -run 'TestDatasource(ListAliases_403_Forbidden|View_403_Forbidden|View_401_Unauthenticated)$' ./internal/transport/http/middleware` → PASS.

Covered outcomes:
- Datasource list alias paths can now be directly exercised under the same permission middleware semantics as datasource view permission checks.
- Explicit forbidden semantics (`403`, code `70001`) are frozen for `/api/ds/list`, `/api/datasource/list`, and `/de2api/datasource/list` when an authenticated caller lacks datasource view permission.
- Existing unauthenticated datasource view proof remains intact alongside the new forbidden alias proof.

Current proof limits:
- This slice freezes the boundary semantics through permission middleware coverage, not by changing the recovered runtime datasource list route wiring itself.
- A runtime repair under task `2.2` is currently blocked because datasource list aliases do not expose a stable resource identifier; making live list routes return resource-bound forbidden semantics would require a broader API/permission design change rather than a narrow hardening patch.

## Visualization hardening evidence

### Dashboard detail missing-resource boundary

- 2026-03-21: `cd apps/backend-go && TEST_DB_HOST=172.19.0.2 TEST_DB_PASSWORD=Admin168 TEST_DB_NAME=dataease_test go test -tags=integration -v -run '^TestVisualizationHandler_FindByID_NotFound$' ./internal/transport/http/handler -count=1` → PASS.

Covered outcomes:
- `POST /dataVisualization/findById` now preserves an explicit not-found envelope at the handler boundary when the requested visualization does not exist.
- Dashboard detail missing-resource semantics are no longer only implied by service-level `gorm.ErrRecordNotFound` handling.

Current proof limits:
- This slice hardens dashboard detail missing-resource behavior only.

### Big-screen edit-entry detail semantics

- 2026-03-21: `cd apps/frontend && npm run e2e -- e2e/recovery/core-reachability.spec.ts -g 'should open big-screen edit view with consumable detail payload'` → PASS.

Covered outcomes:
- A real big-screen leaf ID can be resolved from `/api/dataVisualization/tree` with `busiFlag: 'screen'`.
- `/#/dvCanvas?dvId=<id>` reaches the edit canvas and consumes a successful `/dataVisualization/findById` payload.
- The edit-entry path does not degrade into 401/404 or `Invalid tree payload` runtime failures in the current environment.

Current proof limits:
- This slice hardens big-screen edit-entry and detail payload consumption.
- Deeper edit interactions beyond initial canvas load remain outside the current hardening task.

## Operational route-smoke hardening evidence

### Export-center download boundary semantics

- 2026-03-21: `cd apps/backend-go && go test -v -run 'TestGenerateDownloadURI_(UnauthenticatedUser|TaskNotFound|Dataset_NoPermission)$' ./internal/transport/http/handler` → PASS.

Covered outcomes:
- `GenerateDownloadURI` preserves explicit unauthenticated semantics for callers without auth context.
- Missing export tasks preserve explicit not-found semantics at the route boundary.
- Dataset-backed export download URI generation preserves explicit forbidden semantics when the caller lacks export permission.

Current proof limits:
- This slice hardens boundary-level download URI semantics only.
- A browser-level or route-level export-center smoke beyond handler tests remains open under later hardening work.

### Audit page-entry/detail route smoke

- 2026-03-21: `cd apps/frontend && npm run e2e -- e2e/recovery/core-reachability.spec.ts -g 'should keep audit page entry and detail flow reachable without auth or not-found fallback'` → PASS.

Covered outcomes:
- `/#/audit` remains reachable after login without degrading into 401/404 or page-error states.
- The audit page renders its table shell and exposes at least one detail action in the current environment.
- Clicking the first detail action does not degrade the route into auth or not-found fallback behavior.

Current proof limits:
- This slice hardens route-level page-entry/detail reachability only.
- It does not yet assert deeper audit export or multi-filter interaction semantics.

## Hardening closeout verification

### Focused frontend verification

- 2026-03-21: `cd apps/frontend && npm run e2e -- e2e/recovery/core-reachability.spec.ts -g 'should open big-screen edit view with consumable detail payload|should keep audit page entry and detail flow reachable without auth or not-found fallback'` → PASS.

Covered outcomes:
- The newly-added big-screen edit-entry smoke remains green.
- The newly-added audit page-entry/detail route-smoke remains green.

### Focused backend verification

- 2026-03-21: `cd apps/backend-go && go test -v -run 'TestDatasource(ListAliases_403_Forbidden|View_403_Forbidden|View_401_Unauthenticated)$' ./internal/transport/http/middleware && TEST_DB_HOST=172.19.0.2 TEST_DB_PASSWORD=Admin168 TEST_DB_NAME=dataease_test go test -tags=integration -v -run '^TestVisualizationHandler_FindByID_NotFound$' ./internal/transport/http/handler -count=1 && go test -v -run 'TestGenerateDownloadURI_(UnauthenticatedUser|TaskNotFound|Dataset_NoPermission)$' ./internal/transport/http/handler` → PASS.

Covered outcomes:
- Datasource forbidden hardening remains green.
- Dashboard detail missing-resource hardening remains green.
- Export-center download boundary semantics remain green.

Current proof limits:
- This closeout confirms no regression across the touched hardening slices only; it is not intended as a full repo-wide release gate.

### Operational boundary repair conclusion

- 2026-03-21: no additional export-center or audit runtime repair was required under task `4.3`, because the new hardening checks for download-path semantics and audit page-entry/detail route smoke did not reveal any fresh boundary mismatch beyond what was already fixed in the archived recovery work.
