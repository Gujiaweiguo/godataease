## Why

Seventeen routes for chart, chartData, and datasetField operations currently live as inline handlers inside the compatibility bridge (`compat_bridge.go`). Each route extracts parameters, calls a service method, and returns a response inline, bypassing the canonical handler pattern used by the rest of the Go backend. This duplication makes the bridge file a growing maintenance burden, obscures ownership of business logic, and prevents the frontend from adopting canonical `/api/chart/*`, `/api/chartData/*`, `/api/datasetField/*` paths.

Migrating these 17 routes to canonical handler methods on `ChartHandler` and `DatasetHandler` consolidates routing logic, enables direct frontend consumption of canonical paths, and lets us remove inline bridge code once all clients switch over.

## What Changes

- Add 10 canonical handler methods to `ChartHandler` (6 chart routes, 4 chartData routes) in `chart_handler.go`
- Add 7 canonical handler methods to `DatasetHandler` (7 datasetField routes) in `dataset_handler.go`
- Extend `ChartHandler` with a `datasetService *service.DatasetService` field to support chartData routes that call dataset service methods
- Extend `DatasetHandler` with a `chartService *service.ChartService` field to support datasetField routes that call chart service methods
- Register all 17 canonical routes under `/api/chart/`, `/api/chartData/`, `/api/datasetField/` in `router.go`
- Update constructor functions to inject cross-dependency service references
- Update frontend `chart.ts` to call canonical `/api/chart/*` and `/api/chartData/*` paths (12 functions)
- Update frontend `dataset.ts` to call canonical `/api/datasetField/*` paths (9 functions)
- Compatibility bridge routes remain as backward-compatible aliases; no **BREAKING** changes

## Capabilities

### New Capabilities
- `chart-chartdata-canonical`: Canonical handler methods for 6 chart routes (checkSameDataSet, save, listByDQ, copyField, deleteField, deleteFieldByChart) and 4 chartData routes (getFieldData, getDrillFieldData, innerExportDetails, innerExportDataSetDetails), plus frontend path migration for these endpoints
- `datasetfield-canonical`: Canonical handler methods for 7 datasetField routes (listByDatasetGroup, listWithPermissions, save, getFunction, multFieldValuesForPermissions, copilotFields, listByDsIds), plus frontend path migration for these endpoints

### Modified Capabilities
- `api-compatibility-bridge`: Compatibility bridge retains the 17 routes as aliases pointing to the new canonical handlers, ensuring backward compatibility while the frontend migration completes

## Impact

- **Backend handlers**: `chart_handler.go` gains 10 methods + 1 service field; `dataset_handler.go` gains 7 methods + 1 service field
- **Router**: `router.go` constructor signatures change to accept cross-dependency services; 17 new canonical route registrations added
- **Frontend API**: 21 API functions across `chart.ts` and `dataset.ts` switch from compatibility paths (`/chart/*`, `/chartData/*`, `/datasetField/*`) to canonical paths (`/api/chart/*`, `/api/chartData/*`, `/api/datasetField/*`)
- **Rollback strategy**: Revert frontend API paths to compatibility routes; canonical routes and bridge aliases coexist safely, so no backend rollback is required for frontend-only changes
