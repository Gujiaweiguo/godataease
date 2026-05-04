## ADDED Requirements

### Requirement: Extra Details for Cross-Datasource Lookups
The system SHALL query an external datasource table to retrieve extra column values for select-type fields.

#### Scenario: Retrieve extra details by value
- **WHEN** an authenticated user submits a `POST /data-filling/form/extraDetails` with an `ExtraDetailsRequest` body containing optionDatasource, optionTable, optionColumn, extraColumns, and value
- **THEN** the system validates optionTable and optionColumn against the identifier regex `^[a-zA-Z0-9_]+$`
- **AND** validates each extraColumns fieldName against the same regex
- **AND** opens a connection to the specified external datasource
- **AND** executes a parameterized query: `SELECT optionColumn, extraCol1, extraCol2, ... FROM optionTable WHERE optionColumn = ?` with the provided value
- **AND** returns an array of `ExtraDetails` objects, each containing `name` (the extra column's displayName) and `value` (the queried value)
- **AND** if the query returns no rows, returns an empty array

#### Scenario: Extra details with invalid table name
- **WHEN** the optionTable contains characters not matching the identifier regex
- **THEN** the system returns an error without executing any query

#### Scenario: Extra details with no extra columns
- **WHEN** the extraColumns list is empty
- **THEN** the system returns an empty array (nothing to look up beyond the primary column)

#### Scenario: Extra details for unreachable datasource
- **WHEN** the specified datasource connection cannot be established
- **THEN** the system returns an error with connection failure details

### Requirement: Datasource Column Options
The system SHALL list distinct values from a specified external datasource table and column.

#### Scenario: List column options from external datasource
- **WHEN** an authenticated user submits a `POST /data-filling/form/{optionDatasource}/options` with a `DatasourceOptionsRequest` body containing optionTable, optionColumn, and optional optionOrder
- **THEN** the system validates optionTable and optionColumn against the identifier regex `^[a-zA-Z0-9_]+$`
- **AND** opens a connection to the specified external datasource
- **AND** executes a parameterized query: `SELECT DISTINCT optionColumn FROM optionTable [ORDER BY optionOrder]` where ORDER BY is applied if optionOrder is non-empty
- **AND** returns an array of `ColumnOption` objects, each containing `name` and `value` (both set to the column value)
- **AND** limits results to a maximum of 1000 distinct values

#### Scenario: Column options with ordering
- **WHEN** optionOrder is a non-empty string matching the identifier regex
- **THEN** the system appends `ORDER BY optionOrder` to the query

#### Scenario: Column options with invalid identifiers
- **WHEN** optionTable or optionColumn does not match the identifier regex
- **THEN** the system returns an error without executing any query

### Requirement: Form Template by User Task Item
The system SHALL return the form's JSON template configuration for a given SubInstance item.

#### Scenario: Get template for valid SubInstance item
- **WHEN** an authenticated user submits a `GET /data-filling/template/{itemId}`
- **THEN** the system looks up the SubInstance by `itemId`
- **AND** resolves the form ID via the SubInstance's `form_id` field
- **AND** loads the form and returns its `Forms` field (the JSON string of `[]ExtTableField`)
- **AND** returns the string as the response body with `Content-Type: application/json`

#### Scenario: Get template for non-existent item
- **WHEN** the itemId does not match any SubInstance
- **THEN** the system returns a 404 error
