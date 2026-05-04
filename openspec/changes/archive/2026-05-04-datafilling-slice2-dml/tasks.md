## 1. Domain Models and Types

- [ ] 1.1 Add DfCommitLog GORM model to `internal/domain/datafilling/` with fields: id (int64 auto-increment PK), form_id (int64), data_id (string/varchar64), operate (int: 0=delete,1=insert,2=update), commit_by (int64), committer (varchar255), commit_time (int64), count (int). Include TableName() returning "df_commit_log". Add GORM index tag on form_id.
- [ ] 1.2 Add request/response types to `internal/domain/datafilling/`: TableDataRequest (currentPage, pageSize, searchParameters array), SearchParam (term, field, value, values, multiple), TableDataResponse (data []map[string]interface{}, fields string, total int64, currentPage int64, pageSize int64, key string).
- [ ] 1.3 Write unit tests for SearchParam validation (valid/invalid field names, supported terms).

## 2. DML Provider Interface and MySQL Implementation

- [ ] 2.1 Extend DDLProvider interface in `internal/service/datafilling_ddl.go` with DML methods: InsertRow(ctx, db, tableName, rowData map[string]interface{}) error, UpdateRow(ctx, db, tableName, rowData map[string]interface{}) error, DeleteRows(ctx, db, tableName, ids []string) error, SearchRows(ctx, db, tableName string, whereClause string, args []interface{}, limit, offset int64) ([]map[string]interface{}, error), CountRows(ctx, db, tableName, whereClause string, args []interface{}) (int64, error), TruncateTable(ctx, db, tableName string) error, ListColumnData(ctx, db, tableName, columnName string) ([]string, error).
- [ ] 2.2 Extend DDLProvider with AlterTable methods: AddTableColumns(ctx, db, tableName string, fields []ExtTableField) error, DropTableColumns(ctx, db, tableName string, columnNames []string) error.
- [ ] 2.3 Implement all new DDLProvider methods on MySQLDDLProvider with parameterized SQL. Use isValidDDLIdentifier for all identifiers. Use `?` placeholders for all values.
- [ ] 2.4 Implement search WHERE clause builder: buildWhereClause(params []SearchParam) (string, []interface{}, error). Support terms: eq, not_eq, lt, gt, le, ge, null, not_null, and IN (when multiple=true). Reject invalid field names.
- [ ] 2.5 Write unit tests for MySQLDDLProvider DML methods using table-driven tests (mock gorm.DB or test SQL strings).
- [ ] 2.6 Write unit tests for buildWhereClause covering all term types, multi-value IN, invalid field rejection, empty params.

## 3. Commit Log Repository

- [ ] 3.1 Define CommitLogRepository interface in `internal/repository/` with methods: Create(ctx, log *DfCommitLog) error, ListByFormID(ctx, formID int64, page, pageSize int) ([]*DfCommitLog, int64, error), DeleteAll(ctx) error.
- [ ] 3.2 Implement CommitLogRepository using the main application database (GORM). Include pagination with offset/limit and total count.
- [ ] 3.3 Write integration tests for CommitLogRepository (CRUD, pagination) using MySQL test database with `//go:build integration` tag.

## 4. Service Layer DML Methods

- [ ] 4.1 Add CommitLogRepository dependency to DataFillingService struct. Update NewDataFillingService constructor.
- [ ] 4.2 Implement SearchTableData(ctx, formID int64, req *TableDataRequest) (*TableDataResponse, error): load form definition, open datasource connection, build WHERE clause from search params, call SearchRows + CountRows, build response with fields list.
- [ ] 4.3 Implement SaveRowData(ctx, formID int64, rowData map[string]interface{}, userID int64, userName string) (*TableDataResponse, error): detect insert vs update by ID presence, generate UUID for new rows, call InsertRow/UpdateRow, write commit log.
- [ ] 4.4 Implement DeleteRowData(ctx, formID int64, rowID string, userID int64, userName string) error: call DeleteRows with single ID, write commit log.
- [ ] 4.5 Implement BatchDeleteRowData(ctx, formID int64, ids []string, userID int64, userName string) error: chunk IDs into batches of 500, call DeleteRows per batch, write single commit log with total count.
- [ ] 4.6 Implement TruncateTableData(ctx, formID int64) error: load form, open datasource, call TruncateTable.
- [ ] 4.7 Implement ListColumnData(ctx, formID int64, columnName string) ([]string, error): load form, validate column name, open datasource, call ListColumnData.
- [ ] 4.8 Implement ListCommitLogs(ctx, page, pageSize int) ([]*DfCommitLog, int64, error) and ClearCommitLogs(ctx) error: delegate to CommitLogRepository.
- [ ] 4.9 Update the existing Update service method to compute field diffs and call AddTableColumns/DropTableColumns when forms field definition changes.
- [ ] 4.10 Write service-level unit tests with mocked DDLProvider and repository for: SearchTableData, SaveRowData (insert path), SaveRowData (update path), DeleteRowData, BatchDeleteRowData, TruncateTableData, ListColumnData.

## 5. HTTP Handlers

- [ ] 5.1 Add handler methods to DataFillingHandler: TableData, SaveRowData, DeleteRowData, BatchDeleteRowData, TruncateTableData, ListColumnData, LogPage, LogClear.
- [ ] 5.2 Each handler follows existing pattern: `defer recoverServicePanic(c)`, parse path params with parseIDParamBadRequest, bind JSON body, call service method, return response via response.Success/Error.
- [ ] 5.3 Register new routes in RegisterDataFillingRoutes: POST /form/:id/tableData, POST /form/:formId/rowData/save, GET /form/:formId/delete/:id, POST /form/:formId/batch-delete, GET /form/:formId/truncate, POST /form/:formId/listColumnData, POST /log/page/:goPage/:pageSize, POST /log/clear.

## 6. Verification

- [ ] 6.1 Run `make test` in apps/backend-go. All existing and new unit tests pass.
- [ ] 6.2 Run `make lint` or `golangci-lint run` in apps/backend-go. No new warnings.
- [ ] 6.3 Manual smoke test: start backend, create a form via API, insert a row, search with pagination, update the row, delete the row, verify commit logs appear.
