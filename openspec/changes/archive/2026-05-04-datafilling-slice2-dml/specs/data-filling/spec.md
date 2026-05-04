## MODIFIED Requirements

### Requirement: Data Filling Form Data Table Management
The system SHALL create, alter, and drop physical tables in the user's chosen datasource based on form definitions.

#### Scenario: Create physical table on form save
- **WHEN** a new leaf form is saved with datasourceId and forms field definitions
- **THEN** the system uses the DDLProvider to create a physical table in the specified datasource
- **AND** the table columns map from ExtTableField definitions: nvarchar→VARCHAR(255), text→TEXT, number→BIGINT, decimal→DECIMAL(18,6), datetime→DATETIME
- **AND** a primary key column `id` of type VARCHAR(64) is included
- **AND** if createIndex is true, the specified indexes are created
- **AND** the system returns an error if the DDL operation fails and rolls back the form metadata creation

#### Scenario: Drop physical table on form delete
- **WHEN** a leaf form is deleted
- **THEN** the system uses the DDLProvider to drop the corresponding physical table from the datasource
- **AND** all data in the physical table is permanently lost

#### Scenario: DDL operation on unreachable datasource
- **WHEN** a DDL operation cannot connect to the specified datasource
- **THEN** the system returns an explicit non-success response with the connection error details

#### Scenario: Alter table on form field changes
- **WHEN** a leaf form is updated and the forms field definition has changed from the previous version
- **THEN** the system computes a diff between old and new field lists to determine columns to add, drop, or rename
- **AND** the system calls AddTableColumns on the DDLProvider for new columns and renamed columns
- **AND** the system calls DropTableColumns on the DDLProvider for removed columns
- **AND** column renames are detected via ExtTableField.Settings.Mapping tracking the old column name
- **AND** the system returns an error if any ALTER operation fails without partially modifying the table

#### Scenario: Alter table with no field changes
- **WHEN** a leaf form is updated and the forms field definition has not changed
- **THEN** the system skips all ALTER TABLE operations
- **AND** only the form metadata is updated

### Requirement: Data Filling Form Data CRUD
The system SHALL provide data read, insert, update, and delete operations against the physical form tables.

#### Scenario: Read form data with pagination
- **WHEN** an authenticated user submits a `POST /data-filling/form/{id}/tableData` with pagination parameters (currentPage, pageSize) and optional search parameters array
- **THEN** the system opens a connection to the form's datasource and queries the physical table
- **AND** each row is returned as a map of column name to value (`map[string]interface{}`)
- **AND** the response includes total row count, currentPage, pageSize, a fields string (column names comma-separated), and a data array
- **AND** the response matches the DataFillFormTableDataResponse structure: {data, fields, total, currentPage, pageSize, key}

#### Scenario: Read form data with search filters
- **WHEN** the tableData request includes search parameters with term, field, and value
- **THEN** the system builds a WHERE clause from the search parameters using parameterized queries
- **AND** supported terms are: eq, not_eq, lt, gt, le, ge, null, not_null
- **AND** when term is null or not_null, no value is required
- **AND** paginated results are filtered by the generated WHERE clause

#### Scenario: Read form data with multi-value filter
- **WHEN** the tableData request includes a search parameter with multiple=true and a values array
- **THEN** the system builds an IN clause: `WHERE field IN (?, ?, ...)`
- **AND** all values in the array are bound as separate parameters

#### Scenario: Upsert row data (insert)
- **WHEN** an authenticated user submits a `POST /data-filling/form/{formId}/rowData/save` with column data that does not include an `id` field
- **THEN** the system generates a new UUID as the primary key
- **AND** inserts a new row into the physical table using parameterized SQL
- **AND** creates a commit log entry with operate=1 (insert)
- **AND** returns the DataFillFormTableDataResponse containing the new row

#### Scenario: Upsert row data (update)
- **WHEN** an authenticated user submits a `POST /data-filling/form/{formId}/rowData/save` with column data that includes an `id` field
- **THEN** the system updates the existing row with matching primary key using parameterized SQL
- **AND** creates a commit log entry with operate=2 (update)
- **AND** returns the DataFillFormTableDataResponse containing the updated row

#### Scenario: Delete a single row by ID
- **WHEN** an authenticated user submits a `GET /data-filling/form/{formId}/delete/{id}`
- **THEN** the system deletes the row with the specified primary key from the physical table
- **AND** creates a commit log entry with operate=0 (delete)
- **AND** returns success

#### Scenario: Batch delete rows by IDs
- **WHEN** an authenticated user submits a `POST /data-filling/form/{formId}/batch-delete` with a JSON body containing a list of row IDs
- **THEN** the system deletes all rows with matching primary keys using an IN clause
- **AND** creates a single commit log entry with operate=0 (delete) and count of affected rows
- **AND** if the ID list exceeds 500 entries, the system chunks the deletion into batches of 500
- **AND** returns success

#### Scenario: Truncate table data
- **WHEN** an authenticated user submits a `GET /data-filling/form/{formId}/truncate`
- **THEN** the system executes TRUNCATE TABLE on the physical table
- **AND** all data is removed but the table structure remains
- **AND** returns success

### Requirement: Data Filling Column Options
The system SHALL support querying column option values for select-type fields.

#### Scenario: List column data for select fields
- **WHEN** an authenticated user submits a `POST /data-filling/form/{formId}/listColumnData` with a column name in the request body
- **THEN** the system queries the physical table for distinct non-null values in the specified column
- **AND** returns a list of unique values
- **AND** column name is validated against the identifier regex before use in SQL

### Requirement: Data Filling Commit Log
The system SHALL track all data modifications with an operation log stored in the main application database.

#### Scenario: List commit logs with pagination
- **WHEN** an authenticated user submits a `POST /data-filling/log/page/{goPage}/{pageSize}` with optional form ID filter
- **THEN** the system returns paginated operation log entries from the main DB
- **AND** each entry includes: id, formId, dataId, operate (0=delete, 1=insert, 2=update), commitBy (user ID), committer (user name), commitTime (Unix millis), count (affected rows)
- **AND** the response includes total count

#### Scenario: Clear commit logs
- **WHEN** an authenticated user submits a `POST /data-filling/log/clear`
- **THEN** the system removes all log entries from the main DB
- **AND** returns success
