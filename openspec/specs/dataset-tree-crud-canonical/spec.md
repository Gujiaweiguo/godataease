## ADDED Requirements

### Requirement: Save Dataset via Canonical Endpoint
The system SHALL provide a canonical endpoint `POST /api/dataset/save` that accepts a dataset write request body, validates required fields (name is required), and delegates to `DatasetService.Save()`.

#### Scenario: Save dataset successfully
- **WHEN** client sends `POST /api/dataset/save` with a valid JSON body containing `name` and dataset fields
- **THEN** the system parses the request via `parseDatasetWriteRequest(c, true)`, calls `service.Save(req)`, and returns the result in a success envelope

#### Scenario: Save with missing required name
- **WHEN** client sends `POST /api/dataset/save` with an empty or missing `name` field
- **THEN** the system returns an error response with code `500000` and message "dataset name is required"

### Requirement: Create Dataset via Canonical Endpoint
The system SHALL provide a canonical endpoint `POST /api/dataset/create` that accepts a dataset write request body, validates required fields (name is required), and delegates to `DatasetService.Create()`.

#### Scenario: Create dataset successfully
- **WHEN** client sends `POST /api/dataset/create` with a valid JSON body containing `name` and dataset fields
- **THEN** the system parses the request via `parseDatasetWriteRequest(c, true)`, calls `service.Create(req)`, and returns the result in a success envelope

#### Scenario: Create with missing required name
- **WHEN** client sends `POST /api/dataset/create` with an empty or missing `name` field
- **THEN** the system returns an error response with code `500000` and message "dataset name is required"

### Requirement: Rename Dataset via Canonical Endpoint
The system SHALL provide a canonical endpoint `POST /api/dataset/rename` that accepts a dataset write request with a valid `id` and `name`, and delegates to `DatasetService.Rename(id, name)`.

#### Scenario: Rename dataset successfully
- **WHEN** client sends `POST /api/dataset/rename` with a valid JSON body containing `id` (> 0) and `name`
- **THEN** the system parses the request via `parseDatasetWriteRequest(c, true)`, validates `id > 0`, calls `service.Rename(req.ID, req.Name)`, and returns the result in a success envelope

#### Scenario: Rename with invalid dataset ID
- **WHEN** client sends `POST /api/dataset/rename` with `id` <= 0 or missing
- **THEN** the system returns an error response with code `500000` and message "Invalid dataset ID"

### Requirement: Move Dataset via Canonical Endpoint
The system SHALL provide a canonical endpoint `POST /api/dataset/move` that accepts a dataset write request with a valid `id` and optional `pid`, and delegates to `DatasetService.Move(id, pid)`.

#### Scenario: Move dataset to a new parent
- **WHEN** client sends `POST /api/dataset/move` with a valid JSON body containing `id` (> 0) and `pid`
- **THEN** the system parses the request via `parseDatasetWriteRequest(c, false)`, validates `id > 0`, extracts PID (defaulting to 0 if nil), calls `service.Move(req.ID, pid)`, and returns the result in a success envelope

#### Scenario: Move with invalid dataset ID
- **WHEN** client sends `POST /api/dataset/move` with `id` <= 0 or missing
- **THEN** the system returns an error response with code `500000` and message "Invalid dataset ID"

### Requirement: Delete Dataset via Canonical Endpoint
The system SHALL provide a canonical endpoint `POST /api/dataset/delete/:id` that accepts a path parameter `id` (int64) and delegates to `DatasetService.Delete(id)`.

#### Scenario: Soft-delete dataset successfully
- **WHEN** client sends `POST /api/dataset/delete/123` where 123 is a valid dataset ID
- **THEN** the system parses the `:id` path parameter as int64, calls `service.Delete(id)`, and returns a success envelope with nil data

#### Scenario: Delete with invalid ID format
- **WHEN** client sends `POST /api/dataset/delete/abc` where the ID is not a valid integer
- **THEN** the system returns an error response with code `500000` and message "Invalid dataset ID"

### Requirement: Permanently Delete Dataset via Canonical Endpoint
The system SHALL provide a canonical endpoint `POST /api/dataset/perDelete/:id` that accepts a path parameter `id` (int64) and delegates to `DatasetService.PerDelete(id)`.

#### Scenario: Permanently delete dataset successfully
- **WHEN** client sends `POST /api/dataset/perDelete/123` where 123 is a valid dataset ID
- **THEN** the system parses the `:id` path parameter as int64, calls `service.PerDelete(id)`, and returns the result in a success envelope

#### Scenario: PerDelete with invalid ID format
- **WHEN** client sends `POST /api/dataset/perDelete/abc` where the ID is not a valid integer
- **THEN** the system returns an error response with code `500000` and message "Invalid dataset ID"

### Requirement: Get Dataset Detail via Canonical Endpoint
The system SHALL provide a canonical endpoint `POST /api/dataset/get/:id` that accepts a path parameter `id` (int64) and delegates to `DatasetService.buildDatasetDetail(id)`.

#### Scenario: Get dataset detail successfully
- **WHEN** client sends `POST /api/dataset/get/123` where 123 is a valid dataset ID
- **THEN** the system parses the `:id` path parameter as int64, calls `service.buildDatasetDetail(id)`, and returns the result in a success envelope

#### Scenario: Get dataset detail with invalid ID
- **WHEN** client sends `POST /api/dataset/get/abc` where the ID is not a valid integer
- **THEN** the system returns an error response with code `500000` and message "Invalid dataset ID"

### Requirement: Get Dataset Details via Canonical Endpoint
The system SHALL provide a canonical endpoint `POST /api/dataset/details/:id` that accepts a path parameter `id` (int64) and delegates to `DatasetService.buildDatasetDetail(id)`.

#### Scenario: Get dataset details successfully
- **WHEN** client sends `POST /api/dataset/details/123` where 123 is a valid dataset ID
- **THEN** the system parses the `:id` path parameter as int64, calls `service.buildDatasetDetail(id)`, and returns the result in a success envelope

#### Scenario: Get dataset details with invalid ID
- **WHEN** client sends `POST /api/dataset/details/abc` where the ID is not a valid integer
- **THEN** the system returns an error response with code `500000` and message "Invalid dataset ID"

### Requirement: Batch Get Dataset Details via Canonical Endpoint
The system SHALL provide a canonical endpoint `POST /api/dataset/dsDetails` that accepts a JSON body with dataset IDs and delegates to `DatasetService.buildDatasetDetail()` for each.

#### Scenario: Batch get dataset details successfully
- **WHEN** client sends `POST /api/dataset/dsDetails` with a valid JSON body containing dataset IDs
- **THEN** the system iterates over the IDs, calls `service.buildDatasetDetail(id)` for each, and returns the results in a success envelope

### Requirement: Get SQL Params via Canonical Endpoint
The system SHALL provide a canonical endpoint `POST /api/dataset/getSqlParams` that delegates to `DatasetService.GetSQLParams()`.

#### Scenario: Get SQL params successfully
- **WHEN** client sends `POST /api/dataset/getSqlParams` with a valid request body
- **THEN** the system calls `service.GetSQLParams()`, and returns the result in a success envelope

### Requirement: Get Dataset Bar Info via Canonical Endpoint
The system SHALL provide a canonical endpoint `GET /api/dataset/barInfo/:id` that accepts a path parameter `id` (int64) and delegates to `DatasetGroupService.GetGroupByID(id)`.

#### Scenario: Get bar info successfully
- **WHEN** client sends `GET /api/dataset/barInfo/123` where 123 is a valid dataset group ID
- **THEN** the system parses the `:id` path parameter as int64, calls `service.GetGroupByID(id)`, and returns the result in a success envelope

#### Scenario: Get bar info with invalid ID
- **WHEN** client sends `GET /api/dataset/barInfo/abc` where the ID is not a valid integer
- **THEN** the system returns an error response with code `500000` and message "Invalid dataset ID"

### Requirement: Get Dataset Total via Canonical Endpoint
The system SHALL provide a canonical endpoint `POST /api/dataset/getDatasetTotal` that delegates to `DatasetService.Preview()` with a limit of 1 to count rows.

#### Scenario: Get dataset total successfully
- **WHEN** client sends `POST /api/dataset/getDatasetTotal` with a valid request body
- **THEN** the system calls `service.Preview()` with limit=1, and returns the total count in a success envelope

### Requirement: Preview SQL via Canonical Endpoint
The system SHALL provide a canonical endpoint `POST /api/dataset/previewSql` that delegates to `DatasetService.PreviewSQLWithUser()`.

#### Scenario: Preview SQL successfully
- **WHEN** client sends `POST /api/dataset/previewSql` with a valid SQL preview request body
- **THEN** the system calls `service.PreviewSQLWithUser()`, and returns the preview results in a success envelope

### Requirement: Get Field Enum Object via Canonical Endpoint
The system SHALL provide a canonical endpoint `POST /api/dataset/enumValueObj` that delegates to `DatasetService.GetFieldEnumObj()`.

#### Scenario: Get field enum object successfully
- **WHEN** client sends `POST /api/dataset/enumValueObj` with a valid request body
- **THEN** the system calls `service.GetFieldEnumObj()`, and returns the enum object in a success envelope

### Requirement: Get Field Enum Datasource via Canonical Endpoint
The system SHALL provide a canonical endpoint `POST /api/dataset/enumValueDs` that delegates to `DatasetService.GetFieldEnumDs()`.

#### Scenario: Get field enum datasource successfully
- **WHEN** client sends `POST /api/dataset/enumValueDs` with a valid request body
- **THEN** the system calls `service.GetFieldEnumDs()`, and returns the enum values in a success envelope

### Requirement: Get Multi-Field Enum via Canonical Endpoint
The system SHALL provide a canonical endpoint `POST /api/dataset/enumValue` that delegates to `DatasetService.GetFieldEnum()`.

#### Scenario: Get multi-field enum successfully
- **WHEN** client sends `POST /api/dataset/enumValue` with a valid request body
- **THEN** the system calls `service.GetFieldEnum()`, and returns the enum values in a success envelope
