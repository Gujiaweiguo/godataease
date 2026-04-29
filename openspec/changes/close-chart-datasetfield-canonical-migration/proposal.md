## Why

The archived change `2026-04-16-chart-datasetfield-canonical-migration` left 24 tasks unchecked. Code analysis confirms ~20/24 are implemented, but **Task 5.2 (RegisterChartDataRoutes) is genuinely missing**: the 4 chartData canonical handler methods (`GetFieldData`, `GetDrillFieldData`, `InnerExportDetails`, `InnerExportDataSetDetails`) exist on `ChartHandler` but have no dedicated route registration. They are served exclusively through compat bridge inline closures that duplicate the canonical logic. Additionally, 16 of the 17 new canonical handler methods lack unit test coverage.

## What Changes

- Create `RegisterChartDataRoutes` function to wire the 4 chartData canonical handler methods to `/api/chartData/*` routes
- Refactor compat bridge chartData inline closures to delegate to canonical handlers
- Add unit tests for the 9 untested chart canonical methods and 7 untested datasetField canonical methods
- Add contract-diff baseline fixtures for the 17 new endpoints

## Capabilities

### New Capabilities

_(none)_

### Modified Capabilities

- `chart-chartdata-canonical`: Completes the canonical route registration gap (Task 5.2 from the original migration)
- `datasetfield-canonical`: Adds test coverage for existing canonical handler methods

## Impact

- **Files modified**: `router.go` (new RegisterChartDataRoutes call), `compatibility_bridge_chart_routes.go` (refactor inline closures to delegate), new/updated test files
- **APIs**: No new endpoints — the 4 chartData endpoints are already accessible via compat bridge. This change adds canonical registration and removes code duplication.
- **Database**: No schema changes
- **Risk**: Low — purely additive route registration + test coverage. Compat behavior unchanged.
