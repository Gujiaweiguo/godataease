## ADDED Requirements

### Requirement: Dataset Field List By Dataset Group
The system SHALL provide a canonical endpoint to list dataset fields for a given dataset group, with optional permission filtering and field list flattening.

#### Scenario: List fields with permission
- **WHEN** client sends `POST /api/datasetField/listByDatasetGroup/:datasetId` and the user has view permission on the dataset
- **THEN** the system calls the chart service `ListByDQWithPermission` with dataset ID, chart ID 0, and user ID
- **AND** flattens the resulting chart field list into a flat field array
- **AND** returns the flattened list

#### Scenario: List fields without permission
- **WHEN** client sends `POST /api/datasetField/listByDatasetGroup/:datasetId` and the user does not have view permission
- **THEN** the system calls the chart service `ListByDQ` with dataset ID and chart ID 0
- **AND** flattens the resulting chart field list into a flat field array
- **AND** returns the flattened list

### Requirement: Dataset Field List With Permissions
The system SHALL provide a canonical endpoint to list dataset fields with permission context for a given dataset.

#### Scenario: List fields with permissions via GET
- **WHEN** client sends `GET /api/datasetField/listWithPermissions/:datasetId`
- **THEN** the system calls the chart service `ListByDQWithPermission` or `ListByDQ` based on user permissions
- **AND** flattens the resulting chart field list
- **AND** returns the flattened permission-aware field list

### Requirement: Dataset Field Save
The system SHALL provide a canonical endpoint to save a dataset field (create or update).

#### Scenario: Save a new dataset field
- **WHEN** client sends `POST /api/datasetField/save` with a field object that has no ID
- **THEN** the system creates a new dataset table field record
- **AND** returns the saved field

#### Scenario: Update an existing dataset field
- **WHEN** client sends `POST /api/datasetField/save` with a field object that has a valid ID
- **THEN** the system updates the existing dataset table field record
- **AND** returns the updated field

### Requirement: Dataset Field Get Function
The system SHALL provide a canonical endpoint to retrieve available SQL functions for the calculated field editor.

#### Scenario: Get field functions
- **WHEN** client sends `POST /api/datasetField/getFunction`
- **THEN** the system calls the dataset service field function list method
- **AND** returns a categorized list of SQL functions compatible with the CalcFieldEdit component

### Requirement: Dataset Field Multi-Field Values For Permissions
The system SHALL provide a canonical endpoint to retrieve enumeration values for multiple fields under permission constraints.

#### Scenario: Get multi-field values with permission filtering
- **WHEN** client sends `POST /api/datasetField/multFieldValuesForPermissions` with a field enumeration request body
- **THEN** the system calls the dataset service field enum method with the parsed request
- **AND** returns only field values permitted by the governing field and permission filters

#### Scenario: Empty result under permission filtering
- **WHEN** no field values are visible after applying permission filtering
- **THEN** the system returns an empty successful enumeration result
- **AND** does not misclassify the empty result as a failure

### Requirement: Dataset Field Copilot Fields
The system SHALL provide a canonical endpoint to retrieve dataset fields for the AI copilot feature.

#### Scenario: Get copilot fields for a dataset
- **WHEN** client sends `POST /api/datasetField/copilotFields/:id` for a valid dataset
- **THEN** the system calls the dataset service copilot fields method with the dataset ID and current user ID
- **AND** returns the field list derived from the governed dataset field metadata

### Requirement: Dataset Field List By Datasource IDs
The system SHALL provide a canonical endpoint to batch-query dataset fields by multiple datasource identifiers.

#### Scenario: Query fields by multiple datasource IDs
- **WHEN** client sends `POST /api/datasetField/listByDsIds` with a request body containing an array of datasource IDs
- **THEN** the system calls the dataset service list fields by datasource IDs method
- **AND** returns all fields associated with any of the specified datasources

### Requirement: Frontend Dataset Field API Path Migration
The frontend dataset API module SHALL call canonical `/api/datasetField/*` paths instead of legacy compatibility paths.

#### Scenario: DatasetField API functions use canonical paths
- **WHEN** frontend calls datasetField-related API functions (listByDatasetGroup, listWithPermissions, save, getFunction, multFieldValuesForPermissions, copilotFields, listByDsIds)
- **THEN** the request URLs use `/api/datasetField/*` prefix instead of `/datasetField/*`
