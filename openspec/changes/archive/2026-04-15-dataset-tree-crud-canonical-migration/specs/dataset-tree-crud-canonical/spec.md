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
