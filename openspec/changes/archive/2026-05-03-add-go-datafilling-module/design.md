## Context

DataFilling (数据填报) is the largest remaining Java xpack module. It provides form-based data entry against user-chosen datasources, with task scheduling, user assignment, Excel import/export, and audit logging. The Java implementation spans 37 API endpoints, 23 DTOs, and a dynamic DDL SPI that generates SQL to create/alter/drop tables in external databases.

The Go backend already has established patterns for domain-driven modules (threshold, template, engine). This design follows those patterns and extends them with a DDL provider interface for cross-database table management.

The module will be delivered in 6 slices. This design covers the full architecture but implementation focuses on Slice 1: form CRUD + domain model.

## Goals / Non-Goals

**Goals:**
- Provide a Go-native DataFilling module with clean domain/repository/service/handler layers
- Support dynamic table creation in user-chosen datasources via a DDL provider interface
- Store form metadata in the main DataEase DB while storing actual form data in the user's datasource
- Maintain API compatibility with the Java endpoints under both `/api/data-filling` and `/de2api/data-filling` paths
- Support folder-based tree organization for forms (pid-based, matching existing template/tree patterns)

**Non-Goals:**
- Supporting datasources other than MySQL in Slice 1 (the DDL interface is extensible, but only MySQL is implemented initially)
- Frontend implementation (handled in Slice 6)
- Migrating existing Java xpack data (no data migration tooling)
- Real-time collaboration or conflict resolution on form editing

## Decisions

### Decision 1: DDLProvider as Go interface (not plugin SPI)

```go
type DDLProvider interface {
    CreateTable(ctx context.Context, db *gorm.DB, tableName string, fields []ExtTableField) error
    AlterTable(ctx context.Context, db *gorm.DB, tableName string, toCreate, toModify, toDrop []ExtTableField) error
    DropTable(ctx context.Context, db *gorm.DB, tableName string) error
    CreateIndexes(ctx context.Context, db *gorm.DB, tableName string, indexes []ExtIndexField) error
    InsertData(ctx context.Context, db *gorm.DB, tableName string, data map[string]any) error
    UpdateData(ctx context.Context, db *gorm.DB, tableName string, data map[string]any, pk string) error
    DeleteDataByIDs(ctx context.Context, db *gorm.DB, tableName string, ids []string) error
    SearchData(ctx context.Context, db *gorm.DB, tableName string, params *DataSearchRequest) ([]map[string]any, int64, error)
    TruncateTable(ctx context.Context, db *gorm.DB, tableName string) error
}
```

**Rationale**: Java uses an abstract class SPI with per-database implementations. In Go, an interface is the natural equivalent. Only MySQL is needed initially. The interface takes a `*gorm.DB` connection to the target datasource, so the provider itself is stateless.

**Alternative considered**: Generating raw SQL strings (like the Java SPI does) and executing them. Rejected because GORM's DB executor already handles connection management, parameter binding, and error wrapping. Passing `*gorm.DB` keeps things simpler.

### Decision 2: Form metadata in main DB, data in user datasource

Form definitions (name, folder, column schema, datasource reference) live in `data_filling_forms` in the main DataEase database. The actual form response data lives in dynamically created tables (e.g., `df_xxxx`) in the user's chosen datasource.

**Rationale**: This matches the Java architecture. Form metadata is lightweight and frequently queried for tree navigation. Form data can be large and belongs with the user's data infrastructure. This separation also means the main DB doesn't grow unbounded with form submissions.

### Decision 3: Datasource connection via existing service

DataFilling needs `*gorm.DB` connections to external datasources to execute DDL/DML. The service will depend on a `DatasourceConnectionProvider` interface that wraps the existing datasource service's connection logic.

```go
type DatasourceConnectionProvider interface {
    GetConnection(ctx context.Context, datasourceID int64) (*gorm.DB, error)
}
```

**Rationale**: Avoids coupling DataFilling to the internals of datasource management. The adapter implementation will reuse whatever connection pooling the existing datasource service provides.

### Decision 4: Tree structure with pid-based hierarchy

Forms are organized in folders using a `pid` (parent ID) field, with `node_type` distinguishing folders (`folder`) from forms (`leaf`). This matches the pattern used in `core_visualization_template` (template tree).

**Rationale**: Consistent with existing tree patterns in the codebase. Simple to query with recursive CTEs or application-level tree building.

### Decision 5: Forms JSON stored as TEXT column

The `forms` field containing `[]ExtTableField` is serialized to JSON and stored in a TEXT column. It's deserialized when needed for DDL operations.

**Rationale**: Matches the Java approach where `forms` is a JSON string. Avoids a separate table for field definitions. The field schema is tightly coupled to the form and rarely queried independently.

### Decision 6: Dual route registration

Routes are registered under both `/api/data-filling` and `/de2api/data-filling` paths. The handler layer is shared; only the route group prefix differs.

**Rationale**: Java served endpoints under both prefixes for backward compatibility. Maintaining both ensures frontend compatibility during migration.

## Data Model

### Main DB Table: `data_filling_forms`

| Column | Type | Description |
|--------|------|-------------|
| id | bigint PK auto-increment | Primary key |
| name | varchar(255) | Form or folder name |
| pid | bigint | Parent folder ID (0 = root) |
| level | int | Depth in tree |
| node_type | varchar(50) | `folder` or `leaf` |
| table_name | varchar(255) | Physical table name in datasource (e.g., `df_xxxx`) |
| datasource_id | bigint | References the datasource where data is stored |
| forms | text | JSON string of `[]ExtTableField` definitions |
| create_index | tinyint(1) | Whether to auto-create indexes |
| table_indexes | text | JSON string of `[]ExtIndexField` index definitions |
| create_by | varchar(255) | Creator username |
| create_time | datetime | Creation timestamp |
| update_by | varchar(255) | Last updater username |
| update_time | datetime | Last update timestamp |

### Domain Types

```go
// ExtTableField defines a dynamic column in the form's physical table.
type ExtTableField struct {
    ID       string               `json:"id"`
    Name     string               `json:"name"`
    Type     BaseType             `json:"type"`     // nvarchar, text, number, decimal, datetime
    Settings ExtTableFieldSetting `json:"settings"` // required, unique, inputType, etc.
}

type ExtTableFieldSetting struct {
    Name      string              `json:"name"`
    Required  bool                `json:"required"`
    Unique    bool                `json:"unique"`
    InputType string              `json:"inputType"` // text, select, date, number, etc.
    Mapping   ExtTableFieldMapping `json:"mapping,omitempty"`
}

type ExtTableFieldMapping struct {
    DatasourceID int64  `json:"datasourceId"`
    TableName    string `json:"tableName"`
    ColumnName   string `json:"columnName"`
}

type BaseType string

const (
    BaseTypeNvarchar BaseType = "nvarchar"
    BaseTypeText     BaseType = "text"
    BaseTypeNumber   BaseType = "number"
    BaseTypeDecimal  BaseType = "decimal"
    BaseTypeDatetime BaseType = "datetime"
)

type ExtIndexField struct {
    Name   string `json:"name"`
    Column string `json:"column"`
}
```

## Architecture

```
HTTP Handler (datafilling_handler.go)
    ↓
Service (datafilling_service.go)
    ├── DataFillingRepository (main DB CRUD + tree)
    ├── DDLProvider (dynamic table management)
    └── DatasourceConnectionProvider (get external DB connections)
```

### Slice 1 File Layout

```
internal/
├── domain/datafilling/
│   └── datafilling.go          # Domain types: DataFillingForm, ExtTableField, etc.
├── repository/
│   └── datafilling_repo.go     # CRUD + tree queries against main DB
├── service/
│   ├── datafilling_service.go  # Form CRUD orchestration
│   └── datafilling_ddl.go      # DDLProvider interface + MySQLDDLProvider
├── transport/http/handler/
│   └── datafilling_handler.go  # HTTP handlers
└── transport/http/
    └── router.go               # Updated with data-filling routes
```

## Risks / Trade-offs

**[SQL injection in dynamic DDL]** → DDL provider uses parameterized queries where possible. Table and column names are validated against alphanumeric + underscore patterns. No user-supplied SQL fragments are ever executed directly.

**[Connection leaks to external datasources]** → DatasourceConnectionProvider is responsible for connection lifecycle. The adapter implementation must use connection pooling with idle timeouts. Each DDL/DML operation gets a connection, uses it, and returns it to the pool.

**[Schema drift between forms JSON and physical table]** → When a form is saved with changed fields, the service compares old vs. new field definitions and generates ALTER TABLE statements. This is a best-effort operation. If the physical table is modified outside DataFilling, the schema may diverge. No automatic reconciliation is planned for Slice 1.

**[Large forms JSON in single column]** → TEXT column supports up to 64KB in MySQL. Forms with hundreds of fields could approach this limit. For Slice 1 this is acceptable. If needed later, migrate to MEDIUMTEXT or LONGTEXT.

**[Tree queries on large form counts]** → pid-based tree building in application code is O(n). For thousands of forms this is fine. For tens of thousands, consider adding a materialized path or nested set. Not needed for initial release.

## Migration Plan

1. Database migration adds `data_filling_forms` table to main DB
2. No data migration from Java (new feature in Go)
3. Frontend switches from Java `/de2api/data-filling` endpoints to Go endpoints when ready
4. Both route prefixes (`/api` and `/de2api`) are served from day one for backward compatibility
5. Rollback: drop the `data_filling_forms` table; remove route registration. No foreign keys to other tables.

## Open Questions

- Should the DDL provider handle database type mapping differences (e.g., MySQL TEXT vs. PostgreSQL TEXT)? Not needed for Slice 1 since only MySQL is supported. The interface allows per-database implementations later.
- Should form deletion cascade to drop the physical table in the datasource? The Java implementation does drop it. Slice 1 will implement this behavior.
- How does the existing datasource service expose connections? Need to verify the adapter implementation during Slice 1 coding.
