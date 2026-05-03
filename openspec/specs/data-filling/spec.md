## ADDED Requirements

### Requirement: Data Filling Form Definition CRUD
The system SHALL provide create, read, update, delete, rename, and move operations for data filling form definitions, including folder nodes for tree organization.

#### Scenario: Create a new form definition
- **WHEN** an authenticated user submits a `POST /data-filling/save` request with a valid payload containing name, pid, nodeType=`leaf`, datasourceId, forms (field definitions array as JSON), and optional createIndex flag
- **THEN** the system persists a new `data_filling_forms` record in the main DB with the provided fields
- **AND** generates a unique table name (e.g., `df_` + UUID fragment) for the physical table
- **AND** if `createIndex` is true and `tableIndexes` is provided, records the index definitions
- **AND** returns the created form definition with its generated ID

#### Scenario: Create a folder node
- **WHEN** an authenticated user submits a `POST /data-filling/save` request with nodeType=`folder`, name, and pid
- **THEN** the system persists a new `data_filling_forms` record with node_type=`folder`
- **AND** the record has no table_name, datasource_id, or forms content
- **AND** returns the created folder with its generated ID

#### Scenario: Get a form definition by ID
- **WHEN** an authenticated user submits a `POST /data-filling/get/{id}` request for an existing form ID
- **THEN** the system returns the full form definition including id, name, pid, nodeType, tableName, datasourceId, forms JSON, createIndex, and tableIndexes
- **AND** the response matches the Java DataFillingDTO contract

#### Scenario: Get a non-existent form
- **WHEN** a get request targets a form ID that does not exist
- **THEN** the system returns an explicit non-success response indicating resource not found

#### Scenario: Update a form definition
- **WHEN** an authenticated user submits a `POST /data-filling/update` with a valid payload containing an existing form ID and updated fields
- **THEN** the system updates the matching `data_filling_forms` record with the provided fields
- **AND** if the forms field changed, the system prepares ALTER TABLE operations for the physical table (Slice 2 implements actual execution)
- **AND** returns the updated form definition

#### Scenario: Rename a form or folder
- **WHEN** an authenticated user submits a `POST /data-filling/rename` with id and new name
- **THEN** the system updates the name field of the matching record
- **AND** returns success

#### Scenario: Move a form or folder to a new parent
- **WHEN** an authenticated user submits a `POST /data-filling/move` with id and new pid
- **THEN** the system updates the pid field and recalculates the level field of the matching record
- **AND** returns success

#### Scenario: Move to a non-existent parent
- **WHEN** a move request specifies a pid that does not exist or references a leaf node (not a folder)
- **THEN** the system returns an explicit non-success response

#### Scenario: Delete form definitions by IDs
- **WHEN** an authenticated user submits a `POST /data-filling/delete` with a list of form IDs
- **THEN** the system removes the matching `data_filling_forms` records
- **AND** for leaf nodes, the system drops the corresponding physical table from the datasource
- **AND** returns success

#### Scenario: Delete with empty ID list
- **WHEN** a delete request has an empty ID list
- **THEN** the system returns success without modifying any records

### Requirement: Data Filling Tree Structure
The system SHALL return the form/folder hierarchy as a tree structure, matching the existing BusiNodeVO pattern used in other modules.

#### Scenario: Get full tree
- **WHEN** an authenticated user submits a `POST /data-filling/tree` request
- **THEN** the system returns all form and folder records organized by pid into a tree structure
- **AND** each node contains id, name, pid, nodeType, and leaf count where applicable
- **AND** folders appear before leaf nodes at each level

#### Scenario: Tree with empty database
- **WHEN** a tree request is made and no form or folder records exist
- **THEN** the system returns an empty list

### Requirement: Data Filling Datasource Integration
The system SHALL provide endpoints for listing available datasources and their tables for form configuration.

#### Scenario: List available datasources
- **WHEN** an authenticated user submits a `POST /data-filling/listDatasourceList`
- **THEN** the system returns a list of datasources the user has access to, each containing id, name, and type
- **AND** the response matches the Java datasource list contract

#### Scenario: List all datasources (admin)
- **WHEN** an authenticated user submits a `POST /data-filling/listDatasourceListAll`
- **THEN** the system returns all datasources in the system regardless of user permissions (for admin use)
- **AND** the response matches the Java datasource list contract

#### Scenario: List built-in tables for a datasource
- **WHEN** an authenticated user submits a `POST /data-filling/getBuiltInTables/{datasourceId}`
- **THEN** the system connects to the specified datasource and returns a list of existing table names
- **AND** returns an explicit non-success response if the datasource connection fails

### Requirement: Data Filling Form Data Table Management
The system SHALL create, alter, and drop physical tables in the user's chosen datasource based on form definitions.

#### Scenario: Create physical table on form save
- **WHEN** a new leaf form is saved with datasourceId and forms field definitions
- **THEN** the system uses the DDLProvider to create a physical table in the specified datasource
- **AND** the table columns map from ExtTableField definitions: nvarchar→VARCHAR(255), text→TEXT, number→BIGINT, decimal→DECIMAL(18,6), datetime→DATETIME
- **AND** an auto-increment primary key column `id` is included
- **AND** if createIndex is true, the specified indexes are created

#### Scenario: Drop physical table on form delete
- **WHEN** a leaf form is deleted
- **THEN** the system uses the DDLProvider to drop the corresponding physical table from the datasource
- **AND** all data in the physical table is permanently lost

#### Scenario: DDL operation on unreachable datasource
- **WHEN** a DDL operation cannot connect to the specified datasource
- **THEN** the system returns an explicit non-success response with the connection error details
- **AND** the form metadata is still persisted (or deleted) in the main DB with a warning status

### Requirement: Data Filling Form Data CRUD
The system SHALL provide data read, insert, update, and delete operations against the physical form tables.

#### Scenario: Read form data with pagination
- **WHEN** an authenticated user submits a `POST /data-filling/tableData/{formId}/{goPage}/{pageSize}` with optional search parameters
- **THEN** the system queries the physical table in the datasource and returns paginated rows
- **AND** each row is returned as a map of column name to value
- **AND** the response includes total row count

#### Scenario: Insert a new row
- **WHEN** an authenticated user submits a `POST /data-filling/saveRowData/{formId}` with column data
- **THEN** the system inserts a new row into the physical table using the DDLProvider
- **AND** returns the inserted row ID

#### Scenario: Delete rows by IDs
- **WHEN** an authenticated user submits a `POST /data-filling/deleteRowData/{formId}` with a list of row IDs
- **THEN** the system deletes the matching rows from the physical table
- **AND** returns success

#### Scenario: Batch delete rows
- **WHEN** an authenticated user submits a `POST /data-filling/batchDelete/{formId}` with a list of row IDs
- **THEN** the system deletes all matching rows from the physical table in a single operation
- **AND** returns success

#### Scenario: Truncate table data
- **WHEN** an authenticated user submits a `POST /data-filling/truncate/{formId}`
- **THEN** the system removes all rows from the physical table without dropping the table structure
- **AND** returns success

### Requirement: Data Filling Column Options
The system SHALL support querying column option values for select-type fields.

#### Scenario: List column data for select fields
- **WHEN** an authenticated user submits a `POST /data-filling/listColumnData/{formId}` with column name
- **THEN** the system queries the physical table for distinct values in the specified column
- **AND** returns a list of unique values

#### Scenario: Get extra details for mapped fields
- **WHEN** an authenticated user submits a `POST /data-filling/extraDetails` with field mapping information
- **THEN** the system queries the mapped datasource table for the referenced column values
- **AND** returns the available options

### Requirement: Data Filling Task Management
The system SHALL provide task definition CRUD, scheduling, and lifecycle management for data filling assignments.

#### Scenario: Create a task
- **WHEN** an authenticated user submits a `POST /data-filling/saveTask` with task definition (formId, name, startTime, endTime, rateType, rateValue, userLists)
- **THEN** the system persists a task record linked to the form
- **AND** returns the created task with its ID

#### Scenario: Get task info
- **WHEN** an authenticated user submits a `POST /data-filling/info/{taskId}`
- **THEN** the system returns the full task definition including schedule and assigned users

#### Scenario: List tasks with pagination
- **WHEN** an authenticated user submits a `POST /data-filling/taskPager/{goPage}/{pageSize}`
- **THEN** the system returns paginated task definitions with total count

#### Scenario: Start a task
- **WHEN** an authenticated user submits a `POST /data-filling/startTask/{taskId}`
- **THEN** the system activates the task schedule and begins generating sub-tasks for assigned users
- **AND** returns success

#### Scenario: Stop a task
- **WHEN** an authenticated user submits a `POST /data-filling/stopTask/{taskId}`
- **THEN** the system deactivates the task schedule and stops generating new sub-tasks
- **AND** returns success

#### Scenario: Execute task immediately
- **WHEN** an authenticated user submits a `POST /data-filling/executeNow/{taskId}`
- **THEN** the system triggers immediate task execution, generating sub-tasks for all assigned users
- **AND** returns success

#### Scenario: Delete tasks
- **WHEN** an authenticated user submits a `POST /data-filling/batchDeleteTask` with a list of task IDs
- **THEN** the system removes the matching task records and associated sub-tasks
- **AND** returns success

#### Scenario: List sub-tasks with pagination
- **WHEN** an authenticated user submits a `POST /data-filling/subTaskPager/{taskId}/{goPage}/{pageSize}`
- **THEN** the system returns paginated sub-task records for the specified task

#### Scenario: Delete sub-tasks
- **WHEN** an authenticated user submits a `POST /data-filling/batchDeleteSubTask` with a list of sub-task IDs
- **THEN** the system removes the matching sub-task records
- **AND** returns success

### Requirement: Data Filling User Task
The system SHALL provide user-facing endpoints for listing assigned tasks and submitting form data.

#### Scenario: List user tasks
- **WHEN** an authenticated user submits a `POST /data-filling/listUserTask/{goPage}/{pageSize}`
- **THEN** the system returns paginated tasks assigned to the current user with status information

#### Scenario: Count user todo tasks
- **WHEN** an authenticated user submits a `POST /data-filling/countUserTodoList`
- **THEN** the system returns the count of pending tasks assigned to the current user

#### Scenario: Submit form data for a user task
- **WHEN** an authenticated user submits a `POST /data-filling/saveFormRowData` with task item ID and form data
- **THEN** the system inserts or updates a row in the physical table
- **AND** marks the task item as completed for the user

#### Scenario: Append form data for a user task
- **WHEN** an authenticated user submits a `POST /data-filling/appendFormRowData` with task item ID and form data
- **THEN** the system inserts a new row (not update) in the physical table
- **AND** records the submission against the task item

#### Scenario: Delete user's own submitted data
- **WHEN** an authenticated user submits a `POST /data-filling/userTaskDeleteRowData` with task item ID and row ID
- **THEN** the system deletes the specified row from the physical table
- **AND** returns success

#### Scenario: List users in a sub-task
- **WHEN** an authenticated user submits a `POST /data-filling/listSubTaskUser/{subTaskId}/{goPage}/{pageSize}`
- **THEN** the system returns paginated list of users assigned to the sub-task with completion status

#### Scenario: Get form template for user task
- **WHEN** an authenticated user submits a `POST /data-filling/getTemplateByUserTaskItemId/{itemId}`
- **THEN** the system returns the form definition associated with the task item for the user to fill

### Requirement: Data Filling Commit Log
The system SHALL track all data modifications with an operation log.

#### Scenario: List commit logs with pagination
- **WHEN** an authenticated user submits a `POST /data-filling/logPager/{formId}/{goPage}/{pageSize}`
- **THEN** the system returns paginated operation log entries including operator, operation type, timestamp, and affected data summary
- **AND** the response includes total count

#### Scenario: Clear commit logs
- **WHEN** an authenticated user submits a `POST /data-filling/clearLog/{formId}`
- **THEN** the system removes all log entries for the specified form
- **AND** returns success

### Requirement: Data Filling Excel Integration
The system SHALL support Excel file upload for data import, template download, and data export.

#### Scenario: Upload Excel file
- **WHEN** an authenticated user submits a `POST /data-filling/excelUpload/{formId}` with a multipart Excel file
- **THEN** the system parses the Excel file and maps columns to form fields
- **AND** returns parsed data (DfExcelData) for user confirmation without persisting

#### Scenario: Confirm Excel import
- **WHEN** an authenticated user submits a `POST /data-filling/confirmUpload` with the parsed Excel data and form ID
- **THEN** the system inserts all validated rows into the physical table
- **AND** returns import summary (success count, failure count)

#### Scenario: Download Excel template
- **WHEN** an authenticated user submits a `POST /data-filling/excelTemplate/{formId}`
- **THEN** the system generates an Excel template with column headers matching the form field definitions
- **AND** returns the file as a downloadable attachment

#### Scenario: Export form data to Excel
- **WHEN** an authenticated user submits a `POST /data-filling/innerExport/{formId}`
- **THEN** the system queries all data from the physical table and generates an Excel file
- **AND** returns the file as a downloadable attachment
