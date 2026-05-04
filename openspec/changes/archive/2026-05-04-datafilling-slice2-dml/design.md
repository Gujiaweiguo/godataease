## Context

DataFilling Slice 1 established the form definition layer: CRUD for form metadata (DataFillingForm), tree structure, datasource listing, and basic DDL (CreateTable/DropTable) via a DDLProvider interface with MySQLDDLProvider implementation. Physical tables get created in the user's external MySQL datasource when forms are saved, and dropped when forms are deleted.

The current `DDLProvider` interface in `internal/service/datafilling_ddl.go` has two methods: `CreateTable` and `DropTable`. The `MySQLDDLProvider` builds parameterized DDL SQL with identifier validation via regex (`^[a-zA-Z0-9_]+$`) and backtick quoting. External datasource connections are opened per-request through `DataFillingService.GetDatasourceConnection()`, which decodes the datasource configuration, builds a MySQL DSN, and returns a fresh `*gorm.DB`.

The handler layer (`internal/transport/http/handler/datafilling_handler.go`) registers routes under `/data-filling` with auth and menu-auth middleware. Routes follow the pattern `group.POST("/path", h.Method)` with `defer recoverServicePanic(c)` for panic recovery.

What's missing: actual data operations on the physical tables. Users can define forms and create tables, but cannot insert rows, search data, update records, delete rows, or track changes. The form update flow also lacks AlterTable support, so changing field definitions after initial creation has no effect on the physical table schema.

## Goals / Non-Goals

**Goals:**

- Add DML methods to the DDLProvider interface: InsertRow, UpdateRow, DeleteRows, SearchRows, CountRows, TruncateTable, ListColumnData
- Add AlterTable DDL methods: AddTableColumns, DropTableColumns for schema evolution when form fields change
- Implement a parameterized WHERE clause builder that supports eq, not_eq, lt, gt, le, ge, null, not_null, and IN operators
- Add HTTP endpoints for paginated table data search, row upsert, single/batch delete, truncate, column data listing
- Add commit log infrastructure: DfCommitLog domain model, repository, and endpoints for pagination and clearing
- Follow existing patterns: GORM models in domain, interface+impl in repository, business logic in service, HTTP handlers in transport

**Non-Goals:**

- Multi-database support beyond MySQL (same scope as Slice 1)
- Excel import/export (planned for Slice 5)
- Task management and scheduling (planned for Slice 3)
- User task assignment and submission (planned for Slice 3)
- Connection pooling for external datasources (current per-request model is sufficient for now)
- Frontend changes (frontend already calls these endpoints from the xpack plugin era)

## Decisions

### D1: Extend DDLProvider rather than create a separate DMLProvider

The DDLProvider interface already manages external datasource table operations. Adding DML methods to it keeps all physical table operations in one place. This avoids splitting the MySQL-specific SQL generation across two provider types and keeps the service layer's dependency list simple.

Alternative considered: A separate `DMLProvider` interface. Rejected because DDL and DML share the same connection management, identifier validation, and SQL quoting logic. Splitting would duplicate these concerns.

The interface will be renamed to `DataOperationProvider` in a future refactor, but for Slice 2 we keep the `DDLProvider` name and add methods to it, to minimize diff size.

### D2: Dynamic WHERE clause builder with parameterized queries

Search parameters come from the frontend as an array of `{term, field, value, values, multiple}` objects. The WHERE clause builder converts these into parameterized SQL:

- `eq` → `WHERE `field` = ?`
- `not_eq` → `WHERE `field` != ?`
- `lt/gt/le/ge` → comparison operators
- `null` → `WHERE `field` IS NULL`
- `not_null` → `WHERE `field` IS NOT NULL`
- `multiple=true` with `values` → `WHERE `field` IN (?, ?, ...)`

All identifiers go through the same `isValidDDLIdentifier()` regex check and backtick quoting. All values are passed as `?` placeholders with args slice. This prevents SQL injection in both column names and values.

### D3: Commit log stored in main DB, not external datasource

Commit logs (DfCommitLog) track who modified what data in which form. These records are stored in the main application database (`dataease_dev`), not in the user's external datasource. Rationale: logs are application-level metadata, they should be queryable without connecting to the external datasource, and they should survive even if the external datasource is unreachable.

The DfCommitLog model fields: id, form_id, data_id, operate (0=delete, 1=insert, 2=update), commit_by (user ID), committer (user name), commit_time (Unix millis), count (affected rows).

### D4: Upsert semantics via PK presence check

The `saveRowData` endpoint implements upsert: if the incoming row data map contains an `id` key with a non-empty value, it performs an UPDATE; otherwise it generates a new UUID and performs an INSERT. This matches the Java contract where `rowData/save` handles both cases.

This is simpler than using MySQL's `INSERT ... ON DUPLICATE KEY UPDATE` because the PK is always a client-provided or generated UUID string, not auto-increment.

### D5: AlterTable via field diff on form update

When a form definition is updated and the `forms` field changes, the service diffs old and new field lists to produce three sets:
- `fieldsToCreate`: fields present in new but not old
- `fieldsToDrop`: fields present in old but not new (and marked `removed`)
- `fieldsToModify`: fields whose mapping changed (rename tracked via `ExtTableField.Settings.Mapping`)

`AddTableColumns` handles both new columns and renamed columns (ADD + drop old). `DropTableColumns` handles removed columns. Type changes are handled as drop+recreate.

## Risks / Trade-offs

**[SQL injection in dynamic WHERE]** → All column identifiers go through `isValidDDLIdentifier()` regex + backtick quoting. All values use `?` parameterized placeholders. The builder never interpolates user input into SQL strings.

**[Connection resource leaks]** → Each DML operation opens a datasource connection via `GetDatasourceConnection()`. The returned `*gorm.DB` uses Go's `database/sql` connection pool under the hood, so individual queries don't leak. However, we should ensure `Close()` is called on the `*gorm.DB` when the operation is done if we're not reusing it across multiple queries in the same request.

**[Large IN clause for batch delete]** → Batch delete with very large ID lists could hit MySQL's `max_allowed_packet` or query planner limits. Mitigation: cap batch size at 500 IDs per query, chunk larger requests.

**[AlterTable on tables with existing data]** → Dropping columns destroys data. The service should warn or require confirmation before destructive ALTER operations. For Slice 2, we follow the Java behavior: the frontend sends the complete new field list, and the backend diff determines what to alter. Data loss in dropped columns is by design (user explicitly removed the field).

**[No transaction support across main DB and external datasource]** → Commit logs are in the main DB while data operations happen on the external datasource. If the data operation succeeds but the log write fails, we have unlogged changes. Acceptable for Slice 2 since commit logs are informational, not audited.
