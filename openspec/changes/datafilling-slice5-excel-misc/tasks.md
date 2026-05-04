## 1. Domain Types

- [ ] 1.1 Add new domain types to `apps/backend-go/internal/domain/datafilling/datafilling.go`: `DfExcelData`, `RowDataDatum`, `ExtraDetailsRequest`, `ExtraDetails`, `ExtraColumnItem`, `DatasourceOptionsRequest`, `ColumnOption`
- [ ] 1.2 Add upload session store (in-memory map with mutex) to track parsed Excel data between upload and confirm calls

## 2. Excel Template Generation

- [ ] 2.1 Implement `ExcelTemplateDownload` service method: load form, extract field definitions, create `excelize.File` with header row from field names, stream to Gin response writer with proper Content-Type and Content-Disposition headers
- [ ] 2.2 Add unit tests for Excel template generation (verify header row content, column count, error on non-existent form)

## 3. Excel Upload and Parsing

- [ ] 3.1 Implement `ExcelUpload` service method: accept `*multipart.FileHeader`, open with `excelize`, parse rows against form fields, build `DfExcelData` response, store in session map with UUID key, enforce 10MB size limit
- [ ] 3.2 Add unit tests for Excel upload parsing (valid file, column mapping, empty file, size limit)

## 4. Confirm Upload

- [ ] 4.1 Implement `ConfirmUpload` service method: retrieve parsed data from session map by UUID, iterate `dataList` and insert each row via `DDLProvider.InsertRow` with generated UUID, create commit log, cleanup temp file and session entry
- [ ] 4.2 Implement `UserTaskConfirmUpload` service method: same insert flow scoped to SubInstance, verify user ownership, update SubInstance status to FINISHED, decrement SubTask unfinished count
- [ ] 4.3 Add unit tests for confirm upload (happy path, empty data, non-existent session, unauthorized user-task access)

## 5. Extra Details and Datasource Options

- [ ] 5.1 Implement `ExtraDetails` service method: validate identifiers, open external datasource connection, build parameterized query with extra columns, return `[]ExtraDetails`
- [ ] 5.2 Implement `ListDatasourceOptions` service method: validate identifiers, open external datasource connection, execute `SELECT DISTINCT` query, return `[]ColumnOption` capped at 1000 results
- [ ] 5.3 Add unit tests for identifier validation, query construction, and error handling (invalid identifiers, unreachable datasource)

## 6. Template and Export Endpoints

- [ ] 6.1 Implement `GetTemplateByUserTaskItem` service method: lookup SubInstance by ID, resolve form, return `Forms` JSON string
- [ ] 6.2 Implement `ExportFormData` service method: load form, query all rows with optional search filters, create `excelize.File` with headers and data rows, stream to response writer
- [ ] 6.3 Add unit tests for template retrieval (valid item, non-existent item) and export (empty table, with filters)

## 7. Handler and Route Registration

- [ ] 7.1 Add 7 handler methods to `DataFillingHandler`: `ExcelTemplate`, `ExcelUpload`, `ConfirmUpload`, `ExtraDetails`, `ListDatasourceOptions`, `GetTemplateByUserTaskItem`, `ExportFormData`
- [ ] 7.2 Add 1 handler method to `UserTaskHandler`: `UserTaskConfirmUpload`
- [ ] 7.3 Register 8 new routes in `router.go` under both `/api/data-filling/...` and `/de2api/data-filling/...` groups
- [ ] 7.4 Verify all 8 endpoints return correct status codes for happy path and error cases via manual curl or automated test
