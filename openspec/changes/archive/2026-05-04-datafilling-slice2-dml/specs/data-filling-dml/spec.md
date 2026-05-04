## ADDED Requirements

### Requirement: DML Provider Interface
The system SHALL define a DML provider interface that encapsulates all data manipulation SQL generation for external datasource tables.

#### Scenario: DDLProvider interface includes DML methods
- **WHEN** the service layer needs to manipulate data in a physical form table
- **THEN** the DDLProvider interface provides methods: InsertRow, UpdateRow, DeleteRows, SearchRows, CountRows, TruncateTable, ListColumnData, AddTableColumns, DropTableColumns
- **AND** each method accepts a context, a `*gorm.DB` connection to the external datasource, and operation-specific parameters
- **AND** all methods return errors without panicking

#### Scenario: MySQLDDLProvider implements DML methods
- **WHEN** the MySQLDDLProvider is used for DML operations
- **THEN** InsertRow generates `INSERT INTO tableName (columns) VALUES (?, ?, ...)` with parameterized values
- **AND** UpdateRow generates `UPDATE tableName SET col1=?, col2=? WHERE id=?` with parameterized values
- **AND** DeleteRows generates `DELETE FROM tableName WHERE id IN (?, ?, ...)` with parameterized IDs
- **AND** SearchRows generates `SELECT * FROM tableName [WHERE ...] LIMIT ? OFFSET ?` with parameterized values
- **AND** CountRows generates `SELECT COUNT(*) FROM tableName [WHERE ...]` with parameterized values
- **AND** TruncateTable generates `TRUNCATE TABLE tableName`
- **AND** ListColumnData generates `SELECT DISTINCT column FROM tableName WHERE column IS NOT NULL`

### Requirement: Search Parameter Processing
The system SHALL convert search parameter objects into parameterized SQL WHERE clauses.

#### Scenario: Build WHERE clause from search parameters
- **WHEN** the service receives an array of search parameters, each with term, field, value, values, and multiple fields
- **THEN** the system validates each field name against the identifier regex `^[a-zA-Z0-9_]+$`
- **AND** builds a WHERE clause by joining conditions with AND
- **AND** returns both the SQL fragment (with `?` placeholders) and the args slice

#### Scenario: Handle equality operator
- **WHEN** a search parameter has term="eq", field="status", value="active"
- **THEN** the system generates `` `status` = ? `` with args=["active"]

#### Scenario: Handle comparison operators
- **WHEN** a search parameter has term="gt", field="age", value="18"
- **THEN** the system generates `` `age` > ? `` with args=["18"]
- **AND** the same pattern applies for lt (<), le (<=), ge (>=), not_eq (!=)

#### Scenario: Handle null check operators
- **WHEN** a search parameter has term="null", field="email"
- **THEN** the system generates `` `email` IS NULL `` with no args
- **AND** term="not_null" generates `` `email` IS NOT NULL `` with no args

#### Scenario: Handle IN operator for multi-value filter
- **WHEN** a search parameter has multiple=true, field="category", values=["A", "B", "C"]
- **THEN** the system generates `` `category` IN (?, ?, ?) `` with args=["A", "B", "C"]

#### Scenario: Reject invalid field names
- **WHEN** a search parameter contains a field name that does not match the identifier regex
- **THEN** the system returns an error without generating SQL
- **AND** the error message indicates the field name is invalid

### Requirement: Commit Log Domain Model
The system SHALL define a DfCommitLog domain model for tracking data modifications in the main application database.

#### Scenario: DfCommitLog GORM model structure
- **WHEN** the system initializes the commit log infrastructure
- **THEN** the DfCommitLog model maps to a `df_commit_log` table with columns: id (bigint auto-increment PK), form_id (bigint), data_id (varchar(64)), operate (int: 0=delete, 1=insert, 2=update), commit_by (bigint, user ID), committer (varchar(255), user name), commit_time (bigint, Unix millis), count (int, affected rows)
- **AND** the table has an index on form_id for query performance

#### Scenario: Commit log repository interface
- **WHEN** the service layer needs to interact with commit logs
- **THEN** a CommitLogRepository interface provides methods: Create, ListByFormID (with pagination), DeleteByFormID, DeleteAll
- **AND** the implementation uses the main application database (not the external datasource)

#### Scenario: Commit log created on data operation
- **WHEN** any DML operation (insert, update, delete) succeeds on a form's physical table
- **THEN** the system creates a DfCommitLog record in the main DB with the operation type, user identity, timestamp, and affected row count
- **AND** if the log creation fails, the data operation is not rolled back (log failure is logged but non-blocking)

### Requirement: Row Data Request and Response Types
The system SHALL define request and response types for row data operations.

#### Scenario: Table data request with pagination
- **WHEN** the frontend requests table data
- **THEN** the request includes: currentPage (int64, 1-based), pageSize (int64), and an optional searchParameters array of {term, field, value, values, multiple} objects

#### Scenario: Table data response structure
- **WHEN** the system returns table data
- **THEN** the response includes: data ([]map[string]interface{}, array of rows), fields (string, comma-separated column names), total (int64, total matching rows), currentPage (int64), pageSize (int64), key (string, optional form key)

#### Scenario: Row data save request
- **WHEN** the frontend submits row data for save
- **THEN** the request body is `map[string]interface{}` where keys are column names and values are the cell values
- **AND** if the map contains an "id" key with a non-empty value, the operation is an update; otherwise it is an insert with a generated UUID

#### Scenario: Batch delete request
- **WHEN** the frontend submits a batch delete
- **THEN** the request body contains a JSON array of string IDs to delete
