## ADDED Requirements

### Requirement: Excel Template Download
The system SHALL generate and stream an Excel template file based on a form's field definitions.

#### Scenario: Download template for a valid form
- **WHEN** an authenticated user submits a `GET /data-filling/form/{formId}/excelTemplate`
- **THEN** the system loads the form by ID and extracts the `Forms` field definitions
- **AND** creates an `.xlsx` workbook with one sheet named after the form
- **AND** the first row contains column headers derived from each `ExtTableField`'s `settings.name` (or `settings.mapping.columnName` if name is empty)
- **AND** the response has headers: `Content-Type: application/vnd.openxmlformats-officedocument.spreadsheetml.sheet`, `Content-Disposition: attachment; filename="<formName>.xlsx"`
- **AND** the Excel file is streamed directly to the response writer

#### Scenario: Download template for non-existent form
- **WHEN** an authenticated user submits a `GET /data-filling/form/{formId}/excelTemplate` for a form ID that does not exist
- **THEN** the system returns a 404 error

#### Scenario: Download template for folder node
- **WHEN** an authenticated user submits a `GET /data-filling/form/{formId}/excelTemplate` for a form that is a folder node (nodeType != leaf)
- **THEN** the system returns an error indicating that only leaf forms have templates

### Requirement: Excel Upload
The system SHALL accept a multipart Excel file upload, parse it against the form's field definitions, and return the parsed data for preview.

#### Scenario: Upload valid Excel file
- **WHEN** an authenticated user submits a `POST /data-filling/form/{formId}/uploadFile` with a multipart file field named `file`
- **THEN** the system opens the uploaded `.xlsx` file using `excelize/v2`
- **AND** reads the first sheet and maps each row's column values to the form's `ExtTableField` definitions (matched by column header name to `settings.name` or `settings.mapping.columnName`)
- **AND** returns a `DfExcelData` response containing: `formFields` (the form's field definitions), `dataList` (an array of `RowDataDatum` where each has `data` as a map of column name to value and `insert` set to true), `id` (a generated UUID for this upload session), `excelName` (original filename), `path` (temp file path), `suffix` (file extension)
- **AND** the parsed file is stored in the system temp directory with a `datafilling-upload-` prefix

#### Scenario: Upload non-Excel file
- **WHEN** an authenticated user submits a file that is not a valid `.xlsx` file
- **THEN** the system returns an error indicating the file format is not supported

#### Scenario: Upload file exceeding size limit
- **WHEN** an authenticated user submits a file larger than 10MB
- **THEN** the system returns an error indicating the file exceeds the maximum allowed size

#### Scenario: Upload with no file attached
- **WHEN** an authenticated user submits the request without a `file` multipart field
- **THEN** the system returns an error indicating no file was provided

### Requirement: Confirm Upload
The system SHALL insert confirmed Excel data rows into the form's physical table.

#### Scenario: Confirm and insert all rows
- **WHEN** an authenticated user submits a `POST /data-filling/form/{formId}/confirmUpload` with a JSON body containing the upload `id`
- **THEN** the system retrieves the previously parsed Excel data by the upload session ID
- **AND** for each `RowDataDatum` in the data list, generates a new UUID primary key and inserts the row into the physical form table via `DDLProvider.InsertRow`
- **AND** creates a commit log entry with operate=1 (insert) and count of inserted rows
- **AND** cleans up the temp file
- **AND** returns success

#### Scenario: Confirm with empty data list
- **WHEN** the confirmed data has an empty `dataList`
- **THEN** the system performs no inserts and returns success
- **AND** cleans up the temp file

#### Scenario: Confirm with non-existent upload session
- **WHEN** the upload session ID does not match any stored upload
- **THEN** the system returns an error indicating the upload session was not found

### Requirement: User Task Confirm Upload
The system SHALL allow a user to confirm Excel data upload within a user-task SubInstance context.

#### Scenario: Confirm upload for assigned SubInstance
- **WHEN** an authenticated user submits a `POST /data-filling/user-task/appendData/{id}/form/{formId}/confirmUpload` with a JSON body containing the upload `id`
- **AND** a SubInstance exists where `pid` matches the `{id}` and `uid` matches the current user's ID
- **THEN** the system inserts the confirmed rows into the physical form table
- **AND** associates inserted rows with the SubInstance's `data_id`
- **AND** updates the SubInstance: `status=1` (FINISHED), `finish_time` to current timestamp
- **AND** updates the parent SubTask's `unfinished_count` by decrementing it (if not already FINISHED)
- **AND** creates a commit log entry
- **AND** returns success

#### Scenario: Confirm upload for unassigned SubInstance
- **WHEN** an authenticated user submits confirm for a SubInstance not assigned to them
- **THEN** the system returns a 403 error

### Requirement: Data Export
The system SHALL export all form table data as an Excel file and stream it to the client.

#### Scenario: Export form data
- **WHEN** an authenticated user submits a `POST /data-filling/innerExport/{isDataEaseBi}/{formId}` with optional search parameters in the request body
- **THEN** the system loads the form and its field definitions
- **AND** queries all rows from the physical form table (with optional search filters applied)
- **AND** creates an `.xlsx` workbook with column headers from field definitions
- **AND** populates rows with the query results
- **AND** streams the file to the client with headers: `Content-Type: application/vnd.openxmlformats-officedocument.spreadsheetml.sheet`, `Content-Disposition: attachment; filename="<formName>.xlsx"`
- **AND** returns the Excel file as the response body

#### Scenario: Export form with no data
- **WHEN** the form's physical table is empty
- **THEN** the system generates an `.xlsx` file with only the header row
- **AND** streams the file to the client

#### Scenario: Export non-existent form
- **WHEN** the form ID does not exist
- **THEN** the system returns a 404 error

#### Scenario: Export with search filters
- **WHEN** the request body includes search parameters
- **THEN** the system applies the same search filter logic as `tableData` (WHERE clause generation)
- **AND** only matching rows are included in the export
