## Context

DataFilling slices 1–4 migrated form CRUD, DML operations, task scheduling (cron-based), and user-task workflows to Go. The Go backend uses Gin with a domain-driven architecture: `domain/` for GORM models and DTOs, `repository/` for data access, `service/` for business logic, and `handler/` for HTTP endpoints.

The remaining endpoints — Excel import/export, cross-datasource lookups, and form template retrieval — are defined in the legacy Java `DataFillingApi` interface but have no Go implementation. The `excelize/v2` library (v2.10.1) is already in `go.mod` but not yet imported.

Current architecture constraints:
- DML operations on external datasources go through `DDLProvider` interface (with `MySQLDDLProvider` implementation).
- Form field definitions are stored as JSON in `DataFillingForm.Forms` (a `[]ExtTableField` slice).
- External datasource connections are obtained via `DatasourceConnectionProvider.GetDatasourceConnection()`.
- Routes are registered under both `/api/data-filling/...` and `/de2api/data-filling/...`.

## Goals / Non-Goals

**Goals:**
- Implement Excel template download (generate `.xlsx` from form fields, stream to client).
- Implement Excel upload (accept multipart file, parse with `excelize/v2`, return parsed data for preview).
- Implement confirm upload (insert confirmed rows into form's physical table via DML provider).
- Implement user-task confirm upload (same flow scoped to a SubInstance).
- Implement extra details (cross-datasource lookup for select-type fields with extra columns).
- Implement datasource column options (distinct values from external datasource table).
- Implement template retrieval by SubInstance item ID.
- Implement data export (stream all form data as `.xlsx`).

**Non-Goals:**
- Async export center integration (Java uses ExportCenterManage for background export tasks; this slice uses synchronous streaming).
- Watermark support on exported Excel files.
- Task log message endpoint (`/task/logMsg`).
- Any frontend changes.

## Decisions

### D1: Use `excelize/v2` for all Excel operations
**Rationale**: Already in `go.mod` (v2.10.1). Provides full read/write support for `.xlsx` files with streaming API for large datasets. No viable alternative in the Go ecosystem matches its maturity.
**Alternative considered**: `tealeg/xlsx` — less actively maintained, no streaming support.

### D2: Synchronous streaming for template download and data export
**Rationale**: The legacy Java backend uses `HttpServletResponse` streaming. In Gin, we set response headers (`Content-Type: application/vnd.openxmlformats-officedocument.spreadsheetml.sheet`, `Content-Disposition: attachment`) and write the Excel bytes directly to the response writer. This avoids temp file cleanup complexity.
**Alternative considered**: Async export via background task queue (overkill for this slice; can be added later if file sizes become problematic).

### D3: Two-phase upload flow (upload → confirm)
**Rationale**: Matches the legacy Java pattern. Upload parses the Excel and returns preview data (`DfExcelData` with `formFields` + `dataList`). The frontend shows this preview, then the user confirms to persist rows. This prevents accidental data insertion.
**Implementation**: Upload stores parsed data in memory (session-scoped or via a temp file reference). Confirm reads the parsed data and inserts rows via `DDLProvider.InsertRow`.

### D4: Cross-datasource queries for extra details and column options
**Rationale**: Select-type fields can reference columns in external datasources. The `ExtraDetailsRequest` specifies `optionDatasource`, `optionTable`, `optionColumn`, and optional `extraColumns`. The service opens a connection to the specified datasource and executes a parameterized query.
**SQL pattern**: `SELECT optionColumn, extraColumn1, extraColumn2 FROM optionTable WHERE optionColumn = ?` for extra details; `SELECT DISTINCT optionColumn FROM optionTable ORDER BY optionOrder` for column options.

### D5: Template retrieval returns form JSON directly
**Rationale**: The `GET /template/{itemId}` endpoint looks up the SubInstance by ID, resolves its form ID via the task chain, and returns the form's `Forms` JSON string. This matches the Java behavior where the endpoint returns a raw JSON string.

### D6: New domain types in existing `datafilling.go`
**Rationale**: All DataFilling domain types are in a single file. Adding the new types there maintains consistency.

New types:
- `DfExcelData` — formFields []ExtTableField, dataList []RowDataDatum, id, excelName, path, suffix
- `RowDataDatum` — id string, data map[string]interface{}, insert bool
- `ExtraDetailsRequest` — optionDatasource int64, optionTable, optionColumn string, extraColumns []ExtraColumnItem, value string
- `ExtraDetails` — name string, value interface{}
- `ExtraColumnItem` — fieldName, displayName string
- `DatasourceOptionsRequest` — optionTable, optionColumn, optionOrder string
- `ColumnOption` — name string, value interface{}

## Risks / Trade-offs

- **[Large Excel files]** → Excel upload streaming: `excelize/v2` reads the entire file into memory. For files > 10MB, this could cause memory pressure. Mitigation: document a reasonable file size limit (e.g., 10MB) and enforce it at the handler level via Gin's `MaxMultipartMemory`.
- **[Cross-datasource SQL injection]** → Table/column names in `ExtraDetailsRequest` and `DatasourceOptionsRequest` come from user input. Mitigation: validate all identifiers against regex `^[a-zA-Z0-9_]+$` before use in SQL, consistent with existing `ListColumnData` validation.
- **[Temp file cleanup on upload]** → If confirm is never called after upload, temp files accumulate. Mitigation: use `os.CreateTemp` with a `datafilling-upload-` prefix; implement a cleanup routine or rely on OS temp dir policies.
- **[Synchronous export blocking]** → Large datasets could block the response goroutine. Mitigation: acceptable for now; document that exports are limited to ~100K rows. Async export can be added in a future slice.
