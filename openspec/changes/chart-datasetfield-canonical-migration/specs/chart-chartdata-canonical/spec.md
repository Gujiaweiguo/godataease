## ADDED Requirements

### Requirement: Chart Check Same Dataset
The system SHALL provide a canonical endpoint to check whether two charts share the same dataset.

#### Scenario: Check same dataset between two charts
- **WHEN** client sends `GET /api/chart/checkSameDataSet/:viewIdSource/:viewIdTarget`
- **THEN** the system queries chart definitions for both view IDs
- **AND** returns a boolean comparison result indicating whether they reference the same dataset

### Requirement: Chart Save From Map
The system SHALL provide a canonical endpoint to save chart configuration from a map payload.

#### Scenario: Save chart configuration
- **WHEN** client sends `POST /api/chart/save` with a chart configuration map in the request body
- **THEN** the system calls the chart service save method with the parsed map
- **AND** returns the saved chart result in `code/data/msg` format

### Requirement: Chart List By DQ
The system SHALL provide a canonical endpoint to list chart fields by dataset query with optional permission filtering.

#### Scenario: List chart fields with permission
- **WHEN** client sends `POST /api/chart/listByDQ/:id/:chartId` and the user has view permission on the dataset
- **THEN** the system calls `ListByDQWithPermission` with the dataset ID, chart ID, and user ID
- **AND** returns the chart field list

#### Scenario: List chart fields without permission
- **WHEN** client sends `POST /api/chart/listByDQ/:id/:chartId` and the user does not have view permission on the dataset
- **THEN** the system calls `ListByDQ` with the dataset ID and chart ID
- **AND** returns the chart field list

### Requirement: Chart Copy Field
The system SHALL provide a canonical endpoint to copy a chart field.

#### Scenario: Copy a chart field
- **WHEN** client sends `POST /api/chart/copyField/:id/:chartId`
- **THEN** the system calls the chart service copy field method with the field ID and chart ID
- **AND** returns the copied field result

### Requirement: Chart Delete Field
The system SHALL provide a canonical endpoint to delete a single chart field.

#### Scenario: Delete a chart field by ID
- **WHEN** client sends `POST /api/chart/deleteField/:id`
- **THEN** the system calls the chart service delete field method with the field ID
- **AND** returns success or error response

### Requirement: Chart Delete Field By Chart
The system SHALL provide a canonical endpoint to delete all chart fields for a given chart.

#### Scenario: Delete all fields for a chart
- **WHEN** client sends `POST /api/chart/deleteFieldByChart/:chartId`
- **THEN** the system calls the chart service delete field by chart method with the chart ID
- **AND** returns success or error response

### Requirement: Chart Data Get Field Data
The system SHALL provide a canonical endpoint to retrieve field enumeration data for a chart field.

#### Scenario: Get field enum data
- **WHEN** client sends `POST /api/chartData/getFieldData/:fieldId/:fieldType`
- **THEN** the system calls the dataset service field enum method with the field ID
- **AND** returns the enumeration values for that field

### Requirement: Chart Data Get Drill Field Data
The system SHALL provide a canonical endpoint to retrieve drill-down field enumeration data.

#### Scenario: Get drill field enum data
- **WHEN** client sends `POST /api/chartData/getDrillFieldData/:fieldId`
- **THEN** the system calls the dataset service field enum by datasource method with the field ID
- **AND** returns the enumeration values for that drill field

### Requirement: Chart Data Inner Export Details
The system SHALL provide a canonical endpoint to export chart detail data.

#### Scenario: Export chart detail data
- **WHEN** client sends `POST /api/chartData/innerExportDetails` with an export request body
- **THEN** the system calls the chart export service inner export details method
- **AND** generates a filename and returns the export result with the filename header

### Requirement: Chart Data Inner Export Dataset Details
The system SHALL provide a canonical endpoint to export dataset detail data.

#### Scenario: Export dataset detail data
- **WHEN** client sends `POST /api/chartData/innerExportDataSetDetails` with an export request body
- **THEN** the system calls the chart export service inner export details method
- **AND** generates a filename and returns the export result with the filename header

### Requirement: Frontend Chart API Path Migration
The frontend chart API module SHALL call canonical `/api/chart/*` and `/api/chartData/*` paths instead of legacy compatibility paths.

#### Scenario: Chart API functions use canonical paths
- **WHEN** frontend calls chart-related API functions (checkSameDataSet, save, listByDQ, copyField, deleteField, deleteFieldByChart)
- **THEN** the request URLs use `/api/chart/*` prefix instead of `/chart/*`

#### Scenario: ChartData API functions use canonical paths
- **WHEN** frontend calls chartData-related API functions (getFieldData, getDrillFieldData, innerExportDetails, innerExportDataSetDetails)
- **THEN** the request URLs use `/api/chartData/*` prefix instead of `/chartData/*`
