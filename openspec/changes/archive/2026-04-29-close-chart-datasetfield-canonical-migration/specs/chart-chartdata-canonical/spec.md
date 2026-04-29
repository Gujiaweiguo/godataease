## MODIFIED Requirements

### Requirement: ChartData Canonical Route Registration
The system SHALL register chartData endpoints through canonical handler methods on `ChartHandler`, ensuring the 4 chartData endpoints (`getFieldData`, `getDrillFieldData`, `innerExportDetails`, `innerExportDataSetDetails`) are wired through `RegisterChartDataRoutes` rather than exclusively through compat bridge inline closures.

#### Scenario: ChartData getFieldData route uses canonical handler
- **WHEN** a request hits `POST /api/chartData/getFieldData/:fieldId/:fieldType`
- **THEN** the request SHALL be handled by `ChartHandler.GetFieldData`
- **AND** the response SHALL be identical to the current compat bridge behavior

#### Scenario: ChartData getDrillFieldData route uses canonical handler
- **WHEN** a request hits `POST /api/chartData/getDrillFieldData/:fieldId`
- **THEN** the request SHALL be handled by `ChartHandler.GetDrillFieldData`
- **AND** the response SHALL be identical to the current compat bridge behavior

#### Scenario: ChartData innerExportDetails route uses canonical handler
- **WHEN** a request hits `POST /api/chartData/innerExportDetails`
- **THEN** the request SHALL be handled by `ChartHandler.InnerExportDetails`
- **AND** the response SHALL be identical to the current compat bridge behavior

#### Scenario: ChartData innerExportDataSetDetails route uses canonical handler
- **WHEN** a request hits `POST /api/chartData/innerExportDataSetDetails`
- **THEN** the request SHALL be handled by `ChartHandler.InnerExportDataSetDetails`
- **AND** the response SHALL be identical to the current compat bridge behavior

#### Scenario: Compat bridge delegates to canonical handlers
- **WHEN** the compatibility bridge registers chartData routes
- **THEN** the bridge SHALL delegate to the canonical handler methods instead of duplicating service-call logic in inline closures
