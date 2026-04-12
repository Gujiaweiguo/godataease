## ADDED Requirements

### Requirement: Dataset BarInfo Must Return Complete Audit Data
The system SHALL return complete audit metadata (creator, creation time, updater, update time) when querying dataset barInfo through the compatibility route.

#### Scenario: BarInfo returns real audit fields
- **WHEN** a client calls `GET /datasetTree/barInfo/{id}` for a valid dataset group
- **THEN** the system MUST return `createBy`, `createTime`, `updateBy`, `lastUpdateTime` populated from the `core_dataset_group` database record
- **AND** the `creator` field MUST contain the resolved user name (not the raw user ID)

#### Scenario: BarInfo resolves user names from user IDs
- **WHEN** the system populates `creator` and `updater` fields in barInfo response
- **THEN** it MUST resolve `create_by` user ID to a displayable user name via the user service
- **AND** if the user cannot be found, the system MUST fall back to the raw user ID value

#### Scenario: BarInfo returns meaningful timestamps
- **WHEN** a dataset group has non-zero `create_time` or `last_update_time` values in the database
- **THEN** the barInfo response MUST include those exact timestamp values
- **AND** the system MUST NOT return hardcoded zero values for these fields

### Requirement: Dataset Field Save Must Be Executable
The system SHALL provide an executable dataset field save capability supporting both field creation and field update through the compatibility route.

#### Scenario: Save a new dataset field
- **WHEN** a client calls `POST /datasetField/save` with a field object that has no `id` (or `id=0`)
- **THEN** the system MUST create a new `core_dataset_table_field` record
- **AND** the response MUST include the newly created field with its assigned `id`

#### Scenario: Update an existing dataset field
- **WHEN** a client calls `POST /datasetField/save` with a field object that has a valid `id`
- **THEN** the system MUST update the existing `core_dataset_table_field` record
- **AND** the response MUST include the updated field with all modified attributes

#### Scenario: Save field validates required attributes
- **WHEN** a client calls `POST /datasetField/save` with missing required attributes (name, datasetGroupId, or type)
- **THEN** the system MUST return an explicit validation failure
- **AND** the failure MUST NOT result in a partial or corrupted field record

#### Scenario: Save field for calculated field (extField=2)
- **WHEN** a client saves a calculated field with `extField=2` and a `params` expression
- **THEN** the system MUST persist the expression and mark the field as calculated
- **AND** the field MUST be queryable through existing field list endpoints

### Requirement: Dataset Field Function List Must Be Available
The system SHALL provide a list of available SQL functions for the calculated field editor.

#### Scenario: Query available field functions
- **WHEN** a client calls `POST /datasetField/getFunction`
- **THEN** the system MUST return a categorized list of SQL functions
- **AND** the response MUST be compatible with the frontend CalcFieldEdit.vue component expectations

### Requirement: Dataset Field List By Datasource IDs Must Be Available
The system SHALL provide batch field query capability by multiple datasource identifiers.

#### Scenario: Query fields by multiple datasource IDs
- **WHEN** a client calls `POST /datasetField/listByDsIds` with an array of datasource IDs
- **THEN** the system MUST return fields associated with any of the specified datasources
- **AND** the response MUST include field metadata compatible with the standard `Field` contract
