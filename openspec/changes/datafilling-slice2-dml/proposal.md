## Why

DataFilling Slice 1 established form definition CRUD and basic DDL (CreateTable/DropTable), but the physical form tables have no data operations yet. Users cannot insert, search, update, or delete rows in their form tables. Slice 2 adds the core DML layer — the actual data manipulation operations that make DataFilling functional: row-level CRUD, paginated search with filters, and AlterTable for schema evolution when form fields change.

## What Changes

- **DML Provider interface**: Extend the DDLProvider with data manipulation methods (InsertRow, UpdateRow, DeleteRows, SearchRows, CountRows, TruncateTable, ListColumnData)
- **Dynamic WHERE clause builder**: Build parameterized SQL WHERE clauses from search parameters (eq, not_eq, lt, gt, le, ge, null, not_null, IN)
- **AlterTable DDL**: AddTableColumns, DropTableColumns for when form field definitions change after initial creation
- **Row data CRUD endpoints**: HTTP handlers for tableData (search), saveRowData (upsert), deleteRowData, batchDeleteRowData, truncateRowData
- **Column data endpoint**: List distinct values for select-type fields (listColumnData)
- **Commit log infrastructure**: Domain model and repository for tracking data modifications (logPager, clearLog)
- **Service layer**: DataFillingService methods for all DML operations with external datasource connection handling

## Capabilities

### New Capabilities
- `data-filling-dml`: Data manipulation operations for DataFilling form tables — insert, update, delete, search, truncate, column data listing, and commit logging

### Modified Capabilities
- `data-filling`: Extend existing spec with DML endpoints (tableData, saveRowData, deleteRowData, batchDeleteRowData, truncateRowData, listColumnData) and AlterTable DDL

## Impact

- **Backend Go**: New files in service (DML provider), handler (new endpoints), repository (commit log), domain (commit log model)
- **API contract**: New POST endpoints under `/data-filling/form/{id}/...` and `/data-filling/log/...`
- **No frontend changes**: Frontend already has DataFilling views calling these endpoints (from xpack plugin era)
- **No breaking changes**: All new endpoints are additive
- **Dependencies**: Reuses existing `GetDatasourceConnection()` pattern and `MySQLDDLProvider` infrastructure from Slice 1
