# Regression Evidence

This document freezes the repo-backed verification baseline for `recover-broken-core-features` and is intended to accumulate concrete execution evidence during implementation.

This first pass freezes the **repo-backed evidence baseline** for the P0 BI batch: exact commands that can be executed, the existing test or smoke targets behind them, and the outcomes those targets are currently designed to prove.

Unless a future update explicitly records execution results, the entries below should be read as **verified evidence sources and runnable commands**, not as claims that every command has already been executed in this session.

The first recorded batch is limited to the P0 BI recovery families:
- datasource
- dataset
- dashboard
- big-screen

Export-center and audit evidence should remain outside this document until the change moves into the P1 batch.

## Evidence recording rules

For each verification item below, record:

- the exact command that should be executed and later annotated with result status
- the exact test, suite, or smoke target that was exercised
- the covered outcome proven by that execution
- whether the evidence applies to canonical paths, compatibility paths, or both

Do not use this document to narrate code changes. Keep it focused on evidence sources, runnable proof, and later execution results.

## Datasource evidence

### Backend route and contract evidence

- `cd apps/backend-go && go test -v -run 'TestRegisterRoutes_DatasourceCanonicalAndCompatibilityContracts|TestRegisterRoutes_DatasourceValidateSuccessEnvelopeAcrossAliases' ./internal/transport/http`
- `cd apps/backend-go && go test -v -run 'TestRegisterRoutes_Datasource.*' ./internal/transport/http`

Evidence sources:
- `apps/backend-go/internal/transport/http/router_test.go`

Covered outcomes:
- datasource governed entry paths are reachable
- datasource initialization routes do not degrade into missing-route behavior
- datasource success and non-success envelopes remain explicit
- datasource failures do not collapse into silent empty success

Recorded execution results:
- 2026-03-21: `cd apps/backend-go && go test -v -run 'TestRegisterRoutes_Datasource.*' ./internal/transport/http` → PASS
- 2026-03-21: datasource router slice now includes `TestRegisterRoutes_DatasourceListReturnsExplicitErrorEnvelopeAcrossAliasesWhenStoreUnavailable`, proving `/api/ds/list`, `/api/datasource/list`, and `/de2api/datasource/list` return an explicit `500000` envelope instead of crashing when the datasource store is unavailable in router-test initialization.

### Backend failure-semantic evidence

- `cd apps/backend-go && go test -v -run 'TestDatasourceView_401_Unauthenticated|TestDatasourceView_403_Forbidden' ./internal/transport/http/middleware`
- `cd apps/backend-go && TEST_DB_HOST=172.19.0.2 TEST_DB_PASSWORD=Admin168 TEST_DB_NAME=dataease_test go test -tags=integration -v -run 'TestDatasourceService_(Save|Update|GetByID)' ./internal/service`

Evidence sources:
- `apps/backend-go/internal/transport/http/middleware/permission_integration_test.go`
- `apps/backend-go/internal/service/datasource_service_integration_test.go`
- `apps/backend-go/internal/pkg/response/response.go`

Covered outcomes:
- unauthenticated datasource access remains distinguishable from forbidden access
- forbidden datasource access remains distinguishable from missing route or missing resource
- datasource validation, duplicate-name, and not-found failures remain explicit at service level

Current proof limits:
- datasource alias coverage is now strongest on unauthenticated 401 and store-unavailable explicit non-success; direct forbidden coverage on datasource list aliases still remains weaker than unauthenticated coverage

Recorded execution results:
- 2026-03-21: `cd apps/backend-go && go test -v -run 'TestRegisterRoutes_DatasourceListAliasesRequireAuthentication|TestRegisterRoutes_DatasourceListAliasesReturnExplicitErrorAfterAuthenticationWhenStoreUnavailable' ./internal/transport/http` → PASS as part of the datasource router slice.
- 2026-03-21: live probes after rebuilding the backend and restarting the dev container stack returned `401 + 20001` for `/api/ds/list`, `/api/datasource/list`, and `/de2api/datasource/list` without an authorization header.
- 2026-03-21: the datasource backend route slice now protects canonical, compatibility, and `de2api` list aliases with JWT auth and preserves explicit `500000` envelopes for authenticated store-unavailable failures.

### Frontend affected evidence

- `cd apps/frontend && npm run test:affected:datasource`

Evidence sources:
- `apps/frontend/package.json`
- `apps/frontend/tests/unit/datasource/`

Covered outcomes:
- datasource-focused frontend unit coverage exists as the nearest repo-backed affected suite
- datasource API and tree-state related assertions can be captured through the datasource-specific Vitest target

Current proof limits:
- this command is a datasource-focused frontend suite, but the detailed assertion surface should be recorded per execution run rather than assumed here

### Smoke evidence

- `cd apps/frontend && npm run e2e -- e2e/datasource/datasource.spec.ts`
- `cd apps/frontend && npm run e2e -- e2e/recovery/core-reachability.spec.ts`

Evidence sources:
- `apps/frontend/e2e/datasource/datasource.spec.ts`
- `apps/frontend/e2e/recovery/core-reachability.spec.ts`

Covered outcomes:
- datasource user-facing entry and critical browse flow reach usable initialized state
- datasource route reachability and edit-path smoke can be exercised against a running backend

Current proof limits:
- `e2e/recovery/core-reachability.spec.ts` has not been re-executed in this slice

Recorded execution results:
- 2026-03-21: started local dev environment with `./scripts/dev.sh start` and confirmed `http://localhost:8080/health` returned `{"service":"dataease-backend","status":"ok"}`.
- 2026-03-21: `cd apps/frontend && npm run e2e -- e2e/datasource/datasource.spec.ts -g 'SYS-SMK-005'` → PASS after tightening datasource entry/init assertions and stabilizing `detectDatasourcePageState()` to read page body text instead of unstable locator union counts.
- 2026-03-21: `cd apps/frontend && npm run e2e -- e2e/datasource/datasource.spec.ts -g 'SYS-SMK-006'` → PASS, confirming the datasource page keeps the create-entry smoke visible after successful initialization.
- 2026-03-21: the passing `SYS-SMK-005` run now proves `/#/module-datasource` reaches datasource page chrome and rejects login-form, API-error, and workbench fallback states as acceptable initialization outcomes.
- 2026-03-21: together, `SYS-SMK-005` and `SYS-SMK-006` now provide live datasource entry/init and create-entry smoke for the datasource P0 lane.

## Dataset evidence

### Backend route and contract evidence

- `cd apps/backend-go && go test -v -run 'TestRegisterRoutes_DatasetCanonicalAndCompatibilityContracts' ./internal/transport/http`

Evidence sources:
- `apps/backend-go/internal/transport/http/router_test.go`

Covered outcomes:
- dataset governed entry paths are reachable
- dataset browse, field, and preview routes remain non-404 where governed
- dataset response contracts remain consumable by the frontend path that triggered the flow

Recorded execution results:
- 2026-03-21: `cd apps/backend-go && go test -v -run 'TestRegisterRoutes_Dataset.*' ./internal/transport/http` → PASS.
- 2026-03-21: dataset router coverage now includes `TestRegisterRoutes_DatasetAliasesRequireAuthentication` and `TestRegisterRoutes_DatasetAliasesReturnExplicitErrorAfterAuthenticationWhenStoreUnavailable`, freezing canonical, compatibility, and `de2api` dataset tree / fields / preview route behavior.

### Backend failure-semantic evidence

- `cd apps/backend-go && go test -v -run 'TestDatasetPreviewWithPerm_401_Unauthenticated|TestDatasetPreviewWithPerm_403_Forbidden|TestPermissionDenied_ResponseBody' ./internal/transport/http/middleware`
- `cd apps/backend-go && TEST_DB_HOST=172.19.0.2 TEST_DB_PASSWORD=Admin168 TEST_DB_NAME=dataease_test go test -tags=integration -v -run 'TestDatasetServiceIntegration_(Preview_MissingPhysicalTableReturnsError|PreviewWithPermission_DeniesUnauthorizedDatasourceDependency|GetGroupByID_NotFound|Save_UpdateNotFound|Create_DestinationNotFound)' ./internal/service`

Evidence sources:
- `apps/backend-go/internal/transport/http/middleware/permission_integration_test.go`
- `apps/backend-go/internal/transport/http/handler/dataset_handler.go`
- `apps/backend-go/internal/service/dataset_service.go`
- `apps/backend-go/internal/service/dataset_service_integration_test.go`
- `apps/backend-go/internal/service/dataset_service_test.go`

Covered outcomes:
- dataset authorization failure remains explicit
- dataset dependency failure remains explicit
- dataset missing route or missing resource remains distinguishable
- dataset business execution failure remains explicit instead of placeholder success

Current proof limits:
- direct forbidden coverage is strongest on `previewWithPerm` plus the real frontend-facing `detailWithPerm` path; wider forbidden coverage for every dataset alias family is now an optional hardening step rather than a blocker for this P0 batch

Recorded execution results:
- 2026-03-21: after rebuilding the backend and restarting the dev container stack, live probes returned `401 + 20001` for `/api/datasetTree/tree`, `/api/datasetData/tableField`, and `/de2api/datasetData/previewData` without an authorization header.
- 2026-03-21: authenticated dataset alias tests now preserve explicit non-success envelopes for store-unavailable or invalid-request conditions instead of silent success or panic.
- 2026-03-21: `go test -v -run 'TestCompatibilityBridge_DatasetDetailWithPerm_' ./internal/transport/http/handler` → PASS, proving `/api/datasetTree/detailWithPerm` now returns 401 when unauthenticated and 403 when the caller lacks dataset view permission.
- 2026-03-21: after rebuilding the backend and restarting the dev container stack, a live probe to `/api/datasetTree/detailWithPerm` without authorization returned `401 + 20001`.

### Frontend affected evidence

- `cd apps/frontend && npm run test:core`

Evidence sources:
- `apps/frontend/package.json`
- `apps/frontend/tests/unit/`
- `apps/frontend/tests/integration/`
- `apps/frontend/src/tests/unit/`

Covered outcomes:
- the current repo contains a broader frontend core suite that can be used as the nearest existing command for dataset-related caller and store behavior

Current proof limits:
- no dedicated dataset-only frontend affected command was verified in this pass, so dataset-specific frontend proof should be recorded from actual run output rather than assumed from the broad core suite

### Smoke evidence

- `cd apps/frontend && npm run e2e -- e2e/recovery/core-reachability.spec.ts`

Evidence sources:
- `apps/frontend/e2e/recovery/core-reachability.spec.ts`

Covered outcomes:
- dataset user-facing entry route remains reachable after login

Current proof limits:
- dataset smoke proof now includes a post-semantic-change stability rerun, but it still represents happy-path browse-to-preview coverage rather than exhaustive failure-state UI coverage

Recorded execution results:
- 2026-03-21: `cd apps/frontend && npm run e2e -- e2e/recovery/core-reachability.spec.ts -g 'should open dataset edit view for an existing dataset'` → PASS.
- 2026-03-21: the new dataset edit-view smoke now proves `/api/datasetTree/tree` can yield an existing leaf dataset ID and that `/#/module-dataset?id=<datasetId>` reaches dataset page chrome plus visible dataset detail content without 401/404 or page errors.
- 2026-03-21: `cd apps/frontend && npm run e2e -- e2e/recovery/core-reachability.spec.ts -g 'should render dataset preview content for an existing dataset'` → PASS.
- 2026-03-21: the new dataset preview smoke proves the governed dataset browse path can resolve a leaf ID via `/api/datasetTree/tree`, open `/#/module-dataset?id=<datasetId>`, render visible dataset detail content, and load preview field headers plus either preview rows or explicit empty-preview state without 401/404 or page errors.
- 2026-03-21: after the dataset auth hardening changes, rerunning `should open dataset edit view for an existing dataset` and `should render dataset preview content for an existing dataset` still passed, confirming the logged-in dataset flow remains healthy.

## Visualization evidence

### Backend route and contract evidence

- `cd apps/backend-go && go test -v -run 'TestRegisterRoutes_RegistersVisualizationCompatibilityRoutes|TestRegisterRoutes_VisualizationCanonicalAndCompatibilityContracts' ./internal/transport/http`
- `cd apps/backend-go && TEST_DB_HOST=172.19.0.2 TEST_DB_PASSWORD=Admin168 TEST_DB_NAME=dataease_test go test -tags=integration -v -run 'TestVisualizationServiceIntegration_(Detail|Detail_NotFound|Update_NotFound|List|InteractiveTree)' ./internal/service`

Evidence sources:
- `apps/backend-go/internal/transport/http/router_test.go`
- `apps/backend-go/internal/service/visualization_service_integration_test.go`
- `apps/backend-go/internal/transport/http/handler/visualization_handler_test.go`

Covered outcomes:
- dashboard and big-screen governed entry paths are reachable
- visualization tree, detail, and discovery payloads remain structurally consumable
- visualization routes do not degrade into missing-route behavior where governed

Recorded execution results:
- 2026-03-21: the enhanced dashboard edit smoke in `apps/frontend/e2e/recovery/core-reachability.spec.ts` now requires at least one successful `/dataVisualization/findById` response while opening `/#/dashboard?resourceId=<id>`, proving the discovered dashboard detail payload is consumable by the frontend path.
- 2026-03-21: rerunning `cd apps/frontend && npm run e2e -- e2e/recovery/core-reachability.spec.ts -g 'should open dashboard edit view without invalid tree payload error|should open big-screen preview view with consumable discovery layout'` → PASS after the `findById` response assertions were added.
- 2026-03-21: `cd apps/backend-go && go test -v -run 'TestRegisterRoutes_Visualization.*' ./internal/transport/http` → PASS after adding direct route coverage for `/de2api/dataVisualization/findById`.
- 2026-03-21: after rebuilding the backend and restarting the dev container stack, a live probe to `/de2api/dataVisualization/findById` without authorization returned a single `401 + 20001` JSON envelope instead of the previous concatenated unauthorized/not-found response.

Current proof limits:
- dashboard not-found behavior is still strongest at service level, and big-screen edit/detail missing-resource proof remains less direct than preview/discovery proof; these are follow-on hardening items rather than blockers for the current P0 slice

### Backend failure-semantic evidence

- `cd apps/backend-go && go test -v -run 'TestDataVisualizationFindById_401_Unauthenticated|TestDataVisualizationFindById_403_Forbidden|TestDashboardEdit_403_Forbidden|TestDashboardDelete_401_Unauthenticated|TestDashboardDelete_403_Forbidden|TestScreenView_401_Unauthenticated|TestScreenView_403_Forbidden' ./internal/transport/http/middleware`
- `cd apps/backend-go && go test -v -run 'TestResolveBusiTypes|TestBuildVisualizationTreeValidation|TestFrontendCompatHandler_InteractiveTreeFiltersUnauthorizedVisualizationScopes|TestFrontendCompatHandler_InteractiveTreeReturnsRealDataVNodes' ./internal/transport/http/handler`

Evidence sources:
- `apps/backend-go/internal/transport/http/middleware/permission_integration_test.go`
- `apps/backend-go/internal/transport/http/handler/visualization_handler_test.go`
- `apps/backend-go/internal/transport/http/handler/frontend_compat_handler_test.go`

Covered outcomes:
- dashboard and big-screen unauthenticated and forbidden behavior remain distinguishable
- governed visualization failures remain explicit instead of placeholder success

Current proof limits:
- big-screen semantic proof is still strongest for screen-view authorization and interactiveTree filtering; broader edit/detail semantic hardening remains future work, not a blocker for the current P0 closure

### Frontend affected evidence

- `cd apps/frontend && npm run test:core`

Evidence sources:
- `apps/frontend/package.json`
- `apps/frontend/tests/unit/`
- `apps/frontend/tests/integration/`
- `apps/frontend/src/tests/unit/`

Covered outcomes:
- the current repo contains a broad frontend core suite that can serve as the nearest existing frontend verification command for shared visualization callers and stores

Current proof limits:
- no dedicated dashboard-only or big-screen-only frontend affected unit command was verified in this pass, but the current lane already has direct Playwright execution evidence for dashboard and big-screen entry/discovery flows

### Smoke evidence

- `cd apps/frontend && npm run e2e -- e2e/recovery/core-reachability.spec.ts`

Evidence sources:
- `apps/frontend/e2e/recovery/core-reachability.spec.ts`

Covered outcomes:
- dashboard reaches usable initialized state through the governed entry chain
- big-screen reaches usable initialized state through the governed entry chain

Current proof limits:
- current repo-backed smoke proof now covers dashboard detail payload consumption plus dashboard/big-screen discovery layout consumption; deeper big-screen editor-only behavior remains future hardening rather than a blocker for the current P0 slice

Recorded execution results:
- 2026-03-21: `cd apps/frontend && npm run e2e -- e2e/recovery/core-reachability.spec.ts -g 'should open dashboard edit view without invalid tree payload error'` → PASS.
- 2026-03-21: `cd apps/frontend && npm run e2e -- e2e/recovery/core-reachability.spec.ts -g 'should open dashboard preview view with consumable discovery layout'` → PASS.
- 2026-03-21: `cd apps/frontend && npm run e2e -- e2e/recovery/core-reachability.spec.ts -g 'should open big-screen preview view with consumable discovery layout'` → PASS.
- 2026-03-21: the new dashboard preview smoke proves `/#/dashboardPreview` reaches visible discovery layout (`.dv-preview`, `.resource-area`, `.preview-area`) without 401/404 and without `Invalid tree payload` page errors.
- 2026-03-21: the new big-screen preview smoke proves `/#/previewShow` reaches visible discovery layout (`.dv-preview`, `.resource-area`, `.preview-area`) without 401/404 and without page errors.
- 2026-03-21: after strengthening the dashboard and big-screen smokes with payload-consumption assertions, both targeted visualization smokes still passed, providing a stable smoke baseline for the current visualization P0 slice.
- 2026-03-21: after the de2api detail-route repair, rerunning the targeted visualization smokes for dashboard edit, dashboard preview, and big-screen preview still passed.

## Cross-module P0 verification

### Frontend minimum verification

- `cd apps/frontend && npm run lint`
- `cd apps/frontend && npm run ts:check`
- `cd apps/frontend && npm run test:core`

Covered outcomes:
- repo-backed frontend quality gates and core test suite exist for the P0 BI scope

Current proof limits:
- `npm run test:core` itself was not re-executed in this closeout pass; the recovered-scope Playwright smoke suite and the frontend quality gates were executed directly instead

Recorded execution results:
- 2026-03-21: `cd apps/frontend && npm run lint` → PASS.
- 2026-03-21: `cd apps/frontend && npm run ts:check` → PASS.
- 2026-03-21: `cd apps/frontend && npm run e2e -- --workers=1 e2e/datasource/datasource.spec.ts -g 'SYS-SMK-005|SYS-SMK-006'` → PASS after stabilizing datasource login setup.
- 2026-03-21: `cd apps/frontend && npm run e2e -- e2e/recovery/core-reachability.spec.ts -g 'should open dataset edit view for an existing dataset|should render dataset preview content for an existing dataset|should open dashboard edit view without invalid tree payload error|should open dashboard preview view with consumable discovery layout|should open big-screen preview view with consumable discovery layout'` → PASS.

### Backend minimum verification

- `cd apps/backend-go && make test`
- `cd apps/backend-go && TEST_DB_HOST=172.19.0.2 TEST_DB_PASSWORD=Admin168 TEST_DB_NAME=dataease_test make test-integration`
- `cd apps/backend-go && make drift-check`

Covered outcomes:
- repo-backed backend unit, integration, and drift-check entrypoints exist for the P0 BI scope

Current proof limits:
- full integration closeout is now green for the current environment and compatibility surface after fixture/schema cleanup and cleanup-helper hardening

Recorded execution results:
- 2026-03-21: `cd apps/backend-go && make test` → PASS.
- 2026-03-21: an earlier `make test-integration` failure exposed fixture/schema gaps around `dataset_group_id` and cleanup helper robustness; these were repaired in the permission domain fixtures and both repository/service integration cleanup helpers.
- 2026-03-21: `cd apps/backend-go && TEST_DB_HOST=172.19.0.2 TEST_DB_PASSWORD=Admin168 TEST_DB_NAME=dataease_test make test-integration` → PASS after the integration-fixture fixes.
- 2026-03-21: `cd apps/backend-go && make drift-check` → PASS (`compared endpoints: 22`, `no status drift detected`).

### Manual or end-to-end smoke verification

- `cd apps/frontend && npm run e2e -- e2e/recovery/core-reachability.spec.ts`
- `cd apps/frontend && npm run e2e -- e2e/datasource/datasource.spec.ts`

Covered outcomes:
- repo-backed Playwright smoke targets exist for datasource reachability and for shared BI route reachability across datasource, dataset, dashboard, and big-screen

Current proof limits:
- current repo-backed manual/e2e smoke now covers datasource entry/create, dataset edit/preview, dashboard edit/preview, and big-screen preview; deeper big-screen editor-only behavior remains outside this closeout pass

Recorded execution results:
- 2026-03-21: datasource-specific smoke was executed successfully via `npm run e2e -- e2e/datasource/datasource.spec.ts -g 'SYS-SMK-005'` and later via `npm run e2e -- --workers=1 e2e/datasource/datasource.spec.ts -g 'SYS-SMK-005|SYS-SMK-006'` against the local dev container stack on `http://localhost:8080`.
- 2026-03-21: recovered-scope BI smoke was executed successfully via `npm run e2e -- e2e/recovery/core-reachability.spec.ts -g 'should open dataset edit view for an existing dataset|should render dataset preview content for an existing dataset|should open dashboard edit view without invalid tree payload error|should open dashboard preview view with consumable discovery layout|should open big-screen preview view with consumable discovery layout'`.

## Remaining gaps recorded after evidence review

- Datasource unauthenticated alias coverage is now frozen on canonical, compatibility, and `de2api` list routes, but direct forbidden coverage for datasource list aliases is still weaker than unauthenticated coverage.
- Dashboard missing-resource semantics are still stronger at service level than at caller-facing handler boundary.
- Big-screen evidence is still weaker than dashboard evidence for edit/detail and missing-resource behavior.
- Big-screen frontend smoke currently proves entry reachability more strongly than payload-consumption or rendering-bootstrap semantics.

Rules:
- list only remaining real implementation gaps or explicitly blocked items
- do not mix recovered regressions back into this section
- do not pull export-center or audit work into this document while they remain P1

## Current evidence scope

This document currently covers the repo-backed evidence baseline for the P0 batch and should be updated incrementally with concrete execution results as each datasource, dataset, and visualization recovery lane gains passing evidence.
