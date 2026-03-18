# Permission Semantic Regression Matrix

This matrix captures the repository-verified permission semantics for the core BI flow families during the `stabilize-go-core-bi-compatibility` change.

## Semantic targets

| Condition | Expected status | Expected code | Notes |
|---|---:|---|---|
| Unauthenticated | `401` | `20001` | Must not degrade into `403` or `404` |
| Authenticated but unauthorized | `403` | `70001` | Must remain distinguishable from missing resource |
| Missing route / resource | route-specific | route-specific | Not the same as permission denial |
| Deterministic unavailable / unsupported | explicit non-success | explicit non-success | Must not return placeholder success |

## Repository-verified BI flow coverage

| Flow family | Route example | Unauthenticated evidence | Forbidden evidence | Notes |
|---|---|---|---|---|
| Datasource | `/datasource/:id`, `/api/ds/validate`, `/api/datasource/validate` | `TestDatasourceView_401_Unauthenticated`; datasource route contract tests | `TestDatasourceView_403_Forbidden` | canonical + compatibility validation routes stay non-404 and use governed envelopes |
| Dataset | `/dataset/previewWithPerm` | `TestDatasetPreviewWithPerm_401_Unauthenticated` | `TestDatasetPreviewWithPerm_403_Forbidden`; datasource dependency denial integration test | datasource dependency denial now returns explicit forbidden semantics through handler/service path |
| Dashboard | `/dataVisualization/findById`, `/dataVisualization/deleteLogic/:id`, `/dataVisualization/updateCanvas` | `TestDataVisualizationFindById_401_Unauthenticated`; `TestDashboardDelete_401_Unauthenticated` | `TestDataVisualizationFindById_403_Forbidden`; `TestDashboardDelete_403_Forbidden`; `TestDashboardEdit_403_Forbidden` | canonical and `/de2api` route-entry tests verify these paths do not degrade into `404` |
| Big-screen | `/screen/:id` | `TestScreenView_401_Unauthenticated` | `TestScreenView_403_Forbidden` | big-screen route family now has both unauthenticated and forbidden regression coverage |

## Verified supporting evidence

- `apps/backend-go/internal/transport/http/middleware/permission_integration_test.go`
- `apps/backend-go/internal/transport/http/middleware/menu_auth_test.go`
- `apps/backend-go/internal/transport/http/router_test.go`
- `apps/backend-go/internal/service/dataset_service_integration_test.go`

## Current interpretation

1. Permission denial in core BI flows is currently governed as `403`, not generic `404`.
2. Compatibility route entry for visualization detail remains authentication-gated (`401`) rather than missing-route (`404`).
3. Dataset operations that depend on unauthorized datasources now fail explicitly instead of drifting into generic service errors or placeholder success.
