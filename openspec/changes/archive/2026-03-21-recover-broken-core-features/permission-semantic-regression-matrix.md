# Permission Semantic Regression Matrix

This matrix freezes the governed failure-semantics expectations for the P0 BI recovery batch of `recover-broken-core-features`.

It is intentionally limited to:
- datasource
- dataset
- dashboard
- big-screen

Export-center and audit remain outside this matrix while they are still held in P1.

## Semantic targets

| Condition | Expected semantic | Must not degrade into | Notes |
|---|---|---|---|
| Unauthenticated | explicit unauthenticated non-success | forbidden, missing route/resource, or placeholder success | authentication absence must remain diagnosable |
| Authenticated but unauthorized | explicit forbidden non-success | missing route/resource or placeholder success | denial must remain distinguishable from absence |
| Missing route / missing resource | explicit missing-route or missing-resource behavior | forbidden or generic placeholder success | route/resource absence must not be normalized into authorization semantics |
| Unsupported / dependency / business / bounded real-gap failure | explicit non-success | silent empty success, misleading blank-success state, or generic success with missing data | unsupported or true-gap outcomes must remain distinguishable from recoverable regressions |

## P0 BI semantic matrix

| Flow family | Governed flow slice | Unauthenticated expectation | Forbidden expectation | Missing route/resource expectation | Explicit non-success expectation | Current verification surface | Missing semantic verification target |
|---|---|---|---|---|---|---|---|
| Datasource | datasource entry, initialization, and critical browse flows | explicit unauthenticated outcome for governed datasource entry paths | explicit forbidden outcome for governed datasource access | remains distinguishable from permission denial | datasource init, dependency, unsupported, bounded real-gap, and validation failures remain explicit | Direct unauthenticated datasource alias proof now exists in `apps/backend-go/internal/transport/http/router_test.go` (`TestRegisterRoutes_DatasourceListAliasesRequireAuthentication`) for `/api/ds/list`, `/api/datasource/list`, and `/de2api/datasource/list`, alongside explicit authenticated store-unavailable proof in `TestRegisterRoutes_DatasourceListAliasesReturnExplicitErrorAfterAuthenticationWhenStoreUnavailable`. Datasource view permission middleware evidence still exists in `apps/backend-go/internal/transport/http/middleware/permission_integration_test.go` (`TestDatasourceView_401_Unauthenticated`, `TestDatasourceView_403_Forbidden`), and explicit business and validation non-success remains covered in `apps/backend-go/internal/service/datasource_service_test.go` and `apps/backend-go/internal/service/datasource_service_integration_test.go`. | Add direct forbidden coverage for datasource list aliases and bind the current frontend datasource callers to those authenticated alias paths so forbidden, missing-resource, and validation-failure behavior are frozen at the actual `/api/ds/*` and compatibility route boundary. |
| Dataset | dataset entry, initialization, browse, field, and preview flows | explicit unauthenticated outcome for governed dataset entry paths | explicit forbidden outcome for dataset access and protected dependencies | remains distinguishable from authorization failure | dataset dependency, preview, business, unsupported, and bounded real-gap failures remain explicit | Direct unauthenticated alias proof now exists in `apps/backend-go/internal/transport/http/router_test.go` (`TestRegisterRoutes_DatasetAliasesRequireAuthentication`) for canonical, compatibility, and `de2api` dataset tree / fields / preview routes. Direct frontend-facing protected-path proof now also exists in `apps/backend-go/internal/transport/http/handler/compatibility_bridge_handler_test.go` (`TestCompatibilityBridge_DatasetDetailWithPerm_401_Unauthenticated`, `TestCompatibilityBridge_DatasetDetailWithPerm_403_Forbidden`) for `/api/datasetTree/detailWithPerm`. Protected preview-path 401/403 proof still exists in `apps/backend-go/internal/transport/http/middleware/permission_integration_test.go` (`TestDatasetPreviewWithPerm_401_Unauthenticated`, `TestDatasetPreviewWithPerm_403_Forbidden`) and in `apps/backend-go/internal/transport/http/handler/dataset_handler.go`, where `PreviewWithPermission()` maps unauthenticated access to `Unauthorized` and datasource-permission dependency failure to `Forbidden`. Explicit dependency and business non-success remains verified in `apps/backend-go/internal/service/dataset_service.go`, `apps/backend-go/internal/service/dataset_service_test.go`, and `apps/backend-go/internal/service/dataset_service_integration_test.go`, including missing physical table and unauthorized datasource dependency failures. | Next semantic tightening is optional rather than blocking: broaden direct forbidden coverage beyond `detailWithPerm` if later dataset caller paths start depending on more protected alias families. |
| Dashboard | dashboard entry-chain and discovery/detail flows | explicit unauthenticated outcome for governed dashboard entry | explicit forbidden outcome for governed dashboard access | remains distinguishable from missing resource or missing route | discovery, detail, unsupported, bounded real-gap, and state-bootstrap failures remain explicit | Direct 401/403 proof exists for dashboard detail and edit/delete permission paths in `apps/backend-go/internal/transport/http/middleware/permission_integration_test.go` (`TestDataVisualizationFindById_401_Unauthenticated`, `TestDataVisualizationFindById_403_Forbidden`, `TestDashboardEdit_403_Forbidden`, `TestDashboardDelete_401_Unauthenticated`, `TestDashboardDelete_403_Forbidden`). Route-contract proof for `/api/dataVisualization/findById` and compatibility aliases exists in `apps/backend-go/internal/transport/http/router_test.go`. Explicit unsupported / invalid business input proof exists in `apps/backend-go/internal/transport/http/handler/visualization_handler_test.go` (`TestResolveBusiTypes`, `TestBuildVisualizationTreeValidation`), while not-found detail/update behavior is currently proven at service level in `apps/backend-go/internal/service/visualization_service_integration_test.go`. | Add handler-level semantic checks that prove missing-resource responses remain distinguishable from permission denial for dashboard detail paths, and add caller-facing verification for discovery/detail payload-consumption and bootstrap failures rather than relying on service-only not-found tests plus route smoke. |
| Big-screen | big-screen entry-chain and discovery/detail flows | explicit unauthenticated outcome for governed big-screen entry | explicit forbidden outcome for governed big-screen access | remains distinguishable from missing resource or missing route | discovery, detail, unsupported, bounded real-gap, and rendering-bootstrap failures remain explicit | Direct 401/403 proof currently exists for screen view in `apps/backend-go/internal/transport/http/middleware/permission_integration_test.go` (`TestScreenView_401_Unauthenticated`, `TestScreenView_403_Forbidden`). Additional authorization filtering proof for big-screen discovery scope exists in `apps/backend-go/internal/transport/http/handler/frontend_compat_handler_test.go` (`TestFrontendCompatHandler_InteractiveTreeFiltersUnauthorizedVisualizationScopes`, `TestFrontendCompatHandler_InteractiveTreeReturnsRealDataVNodes`). Explicit unsupported / invalid business input proof is shared with dashboard through `apps/backend-go/internal/transport/http/handler/visualization_handler_test.go`, but direct big-screen missing-resource semantics are not yet frozen. | Add direct semantic checks for big-screen edit/detail and missing-resource behavior, plus route-contract and payload-consumption verification for the shared visualization caller chain so `screen` / `dataV` semantics are not inferred only from dashboard tests and interactiveTree filtering. |

## Freeze checklist

Each semantic row is not considered frozen until all of the following are captured:

- governed flow slice
- expected unauthenticated behavior
- expected forbidden behavior
- expected missing route/resource behavior
- expected explicit non-success behavior for unsupported, dependency, business, and bounded real-gap failures
- current verification surface
- missing verification target

## Current interpretation

This matrix is the semantic contract for the P0 BI batch. It exists to prevent permission denial, missing-route behavior, and unsupported, dependency, business, or bounded real-gap failures from collapsing into the same observed symptom during recovery.
