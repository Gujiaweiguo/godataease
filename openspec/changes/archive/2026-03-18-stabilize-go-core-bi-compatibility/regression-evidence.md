# Regression Evidence

This document records the concrete verification evidence produced during implementation of `stabilize-go-core-bi-compatibility`.

## Core BI baseline and route-contract evidence

### Datasource

- `go test ./internal/transport/http -run 'TestRegisterRoutes_DatasourceCanonicalAndCompatibilityContracts|TestRegisterRoutes_DatasourceValidateSuccessEnvelopeAcrossAliases' -count=1`
- `go test ./internal/service -run 'TestDatasourceService_Validate|TestDatasourceServiceHelpers_PingTCPTimeout' -count=1`
- `go test ./internal/transport/http/middleware -run 'TestDatasourceView_401_Unauthenticated|TestDatasourceView_403_Forbidden' -count=1`

Covered outcomes:
- canonical `/api/ds/*` routes are reachable and return governed envelopes
- compatibility `/api/datasource/*` and `/de2api/datasource/*` routes are reachable and non-404
- datasource validation success/failure semantics are deterministic
- datasource permission denial remains `403`, not `404`

### Dataset

- `go test ./internal/transport/http -run TestRegisterRoutes_DatasetCanonicalAndCompatibilityContracts -count=1`
- `TEST_DB_HOST=172.19.0.2 TEST_DB_PASSWORD=Admin168 TEST_DB_NAME=dataease_test go test -tags=integration ./internal/service -run 'TestDatasetServiceIntegration_(Tree|Fields_WithMetadata|Preview_MissingPhysicalTableReturnsError|PreviewAndPreviewWithPermission|PreviewWithPermission_DeniesUnauthorizedDatasourceDependency)$' -count=1`

Covered outcomes:
- canonical `/api/dataset/*` and compatibility `/api/datasetTree/*`, `/api/datasetData/*`, `/de2api/*` entry paths are reachable and non-404
- dataset tree payload shape is stable
- dataset field metadata is not silently dropped while returning success
- dataset preview missing-table failure is deterministic
- unauthorized datasource dependency in `PreviewWithPermission` returns explicit denied semantics

### Visualization / Dashboard / Big-screen

- `go test ./internal/transport/http -run TestRegisterRoutes_VisualizationCanonicalAndCompatibilityContracts -count=1`
- `go test ./internal/transport/http/handler -run 'TestBuildVisualizationTreeValidation|TestBuildVisualizationTreeContractShape|TestResolveBusiTypes' -count=1`
- `TEST_DB_HOST=172.19.0.2 TEST_DB_PASSWORD=Admin168 TEST_DB_NAME=dataease_test go test -tags=integration ./internal/service -run 'TestVisualizationServiceIntegration_(Detail|Detail_Completeness|Detail_NotFound)$' -count=1`

Covered outcomes:
- visualization tree/detail routes are reachable via canonical and compatibility entries
- malformed tree payloads fail instead of silently succeeding
- dashboard/big-screen detail payloads retain required rendering fields
- missing visualization detail does not degrade into placeholder success

### Permission semantics and compatibility governance

- `go test ./internal/transport/http/middleware -run 'Test(ScreenView_401_Unauthenticated|ScreenView_403_Forbidden|DatasourceView_401_Unauthenticated|DatasourceView_403_Forbidden|DatasetPreviewWithPerm_401_Unauthenticated|DatasetPreviewWithPerm_403_Forbidden|DataVisualizationFindById_401_Unauthenticated|DataVisualizationFindById_403_Forbidden|DashboardDelete_401_Unauthenticated|DashboardDelete_403_Forbidden)$' -count=1`
- `go test ./internal/transport/http/middleware -run 'TestRequireMenuAuth_NoRole|TestRequireMenuAuth_NonAdminDenied|TestRequireMenuAuth_InvalidRoleType' -count=1`
- `go test ./internal/transport/http/handler -run 'TestFrontendCompatHandler_InteractiveTreeUsesAuthorizedMenus|TestPermissionCompatHandler_TargetPermissionEndpointsReturnExplicitNonSuccess' -count=1`
- `go test ./internal/transport/http/handler -run 'TestStatusConsistencyWithWhitelist|TestNoPlaceholderSuccessInCompatibilityBridge' -count=1`
- `./scripts/check-status-drift.sh`

Covered outcomes:
- datasource / dataset / dashboard / screen permission denial remains distinguishable from missing-route behavior
- permission-filtered interactive tree payloads preserve required node fields
- target permission compatibility APIs no longer return placeholder success
- whitelist metadata and observed compatibility status remain aligned for checked endpoints

### Frontend compatibility convergence and smoke evidence

- `npm run test -- --run tests/unit/store/interactive.test.ts`
- `npm run ts:check`
- `npm run e2e:system-smoke`

Covered outcomes:
- interactive store falls back to empty compatibility state when BI tree API rejects
- missing BI families in interactive tree aggregate responses are backfilled to empty trees instead of being left undefined
- `@system-smoke` now exercises dashboard / screen / dataset / datasource entry paths through existing Playwright smoke coverage

## Current evidence scope

The evidence above covers task groups `1.x` through `8.4` for the work completed in this implementation session.
