## 1. Domain Types & Data Model

- [ ] 1.1 Create `internal/domain/datafilling/datafilling.go` with domain types: `DataFillingForm` (GORM model with TableName() returning `data_filling_forms`), `ExtTableField`, `ExtTableFieldSetting`, `ExtTableFieldMapping`, `BaseType` enum (nvarchar/text/number/decimal/datetime), `ExtIndexField`, `CreateFormRequest`, `UpdateFormRequest`, `TreeNode`, `TreeResponse`
- [ ] 1.2 Create database migration file for `data_filling_forms` table (id, name, pid, level, node_type, table_name, datasource_id, forms TEXT, create_index, table_indexes TEXT, create_by, create_time, update_by, update_time)
- [ ] 1.3 Add unit tests for domain type JSON serialization/deserialization (ExtTableField round-trip, BaseType string constants)

## 2. Repository Layer

- [ ] 2.1 Create `internal/repository/datafilling_repo.go` with `DataFillingRepository` struct and `DataFillingRepositoryInterface` (Create, GetByID, Update, DeleteByIDs, Rename, Move, GetTree, GetByPid)
- [ ] 2.2 Implement tree query: fetch all records ordered by level then node_type (folders first), build tree in application code
- [ ] 2.3 Implement move operation: update pid and recalculate level
- [ ] 2.4 Add integration tests for CRUD operations (create folder, create form, get, update, rename, move, delete, tree query) using MySQL test database

## 3. DDL Provider Interface

- [ ] 3.1 Create `internal/service/datafilling_ddl.go` with `DDLProvider` interface (CreateTable, DropTable) and `DatasourceConnectionProvider` interface (GetConnection)
- [ ] 3.2 Implement `MySQLDDLProvider` with CreateTable: map BaseType to MySQL column types, include auto-increment `id` primary key, handle VARCHAR/TEXT/BIGINT/DECIMAL/DATETIME
- [ ] 3.3 Implement `MySQLDDLProvider` with DropTable
- [ ] 3.4 Add unit tests for DDL SQL generation (verify correct SQL for each BaseType, verify primary key inclusion)

## 4. Service Layer

- [ ] 4.1 Create `internal/service/datafilling_service.go` with `DataFillingService` struct (depends on DataFillingRepository, DDLProvider, DatasourceConnectionProvider)
- [ ] 4.2 Implement form CRUD methods: Save (create form + call DDL CreateTable), Get, Update, Delete (call DDL DropTable for leaf nodes), Rename, Move
- [ ] 4.3 Implement folder-specific Save (no DDL operations, just metadata)
- [ ] 4.4 Implement Tree method (delegate to repository, return TreeResponse)
- [ ] 4.5 Implement datasource list methods: ListDatasourceList, ListDatasourceListAll (delegate to existing datasource service), GetBuiltInTables
- [ ] 4.6 Add unit tests for service layer with mock repository and DDL provider (verify DDL is called on form save/delete, verify tree building logic, verify move validation)

## 5. HTTP Handler & Routes

- [ ] 5.1 Create `internal/transport/http/handler/datafilling_handler.go` with `DataFillingHandler` struct and handler methods for all Slice 1 endpoints
- [ ] 5.2 Implement handlers: Save, Get, Update, Delete, Rename, Move, Tree, ListDatasourceList, ListDatasourceListAll, GetBuiltInTables
- [ ] 5.3 Create `RegisterDataFillingRoutes` function registering under both `/data-filling` (for `/api` prefix) and `/data-filling` (for `/de2api` prefix) with auth middleware
- [ ] 5.4 Wire DataFillingService, DataFillingRepository, DataFillingHandler in dependency injection (update router.go or app setup)
- [ ] 5.5 Add handler-level tests for request parsing, validation, and response format (verify JSON response envelope matches Java contract)

## 6. Integration Verification

- [ ] 6.1 Run `make test` and ensure all existing tests pass (no regressions)
- [ ] 6.2 Run `golangci-lint run` on new files and fix any issues
- [ ] 6.3 Verify route registration by starting the server and checking `/data-filling` endpoints return expected status codes (401 without auth)

---

## Future Slices (not implemented in this change)

### Slice 2: Data CRUD + Full DDL Provider
- Implement AlterTable in MySQLDDLProvider (add/modify/drop columns)
- Implement CreateIndexes in MySQLDDLProvider
- Implement DML methods: InsertData, UpdateData, DeleteDataByIDs, SearchData, TruncateTable
- Implement form data endpoints: tableData, saveRowData, deleteRowData, batchDelete, truncate
- Implement column option endpoints: listColumnData, extraDetails
- Add search/filter support via DataSearchRequest

### Slice 3: Task Management
- Create task domain types and `data_filling_tasks` table
- Implement task CRUD repository and service
- Integrate Go cron scheduler for task execution
- Implement startTask, stopTask, executeNow endpoints
- Implement sub-task management (subTaskPager, batchDeleteSubTask)

### Slice 4: User Task & Fill
- Create user task domain types and `data_filling_user_tasks` table
- Implement user task listing (listUserTask, countUserTodoList)
- Implement user form submission (saveFormRowData, appendFormRowData, userTaskDeleteRowData)
- Implement sub-task user listing (listSubTaskUser)
- Implement template retrieval (getTemplateByUserTaskItemId)

### Slice 5: Excel Import/Export + Commit Log
- Integrate Excel parsing library (e.g., excelize)
- Implement excelUpload endpoint (parse to DfExcelData)
- Implement confirmUpload endpoint (batch insert parsed data)
- Implement excelTemplate endpoint (generate template with headers)
- Implement innerExport endpoint (query data, generate Excel file)
- Create commit log domain types and `data_filling_commit_log` table
- Implement logPager and clearLog endpoints

### Slice 6: Frontend Integration + Polish
- Create or update `api/dataFilling.ts` frontend API module
- Wire frontend components to Go backend endpoints
- Add permission integration tests
- End-to-end verification of complete data filling workflow
