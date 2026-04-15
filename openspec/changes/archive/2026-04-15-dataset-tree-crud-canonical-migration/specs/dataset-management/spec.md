## MODIFIED Requirements

### Requirement: Dataset API Path Configuration
The frontend dataset API module SHALL call canonical `/dataset/*` paths instead of legacy `/datasetTree/*` paths for all 6 CRUD operations.

This updates the following API call URLs:
- `/datasetTree/save` → `/dataset/save`
- `/datasetTree/create` → `/dataset/create`
- `/datasetTree/rename` → `/dataset/rename`
- `/datasetTree/move` → `/dataset/move`
- `/datasetTree/delete/${id}` → `/dataset/delete/${id}`
- `/datasetTree/perDelete/${id}` → `/dataset/perDelete/${id}`

#### Scenario: Save dataset uses canonical path
- **WHEN** frontend calls the save dataset function
- **THEN** the request is sent to `POST /dataset/save` (Vite proxies to `/api/dataset/save`)

#### Scenario: Create dataset uses canonical path
- **WHEN** frontend calls the create dataset function
- **THEN** the request is sent to `POST /dataset/create` (Vite proxies to `/api/dataset/create`)

#### Scenario: Rename dataset uses canonical path
- **WHEN** frontend calls the rename dataset function
- **THEN** the request is sent to `POST /dataset/rename` (Vite proxies to `/api/dataset/rename`)

#### Scenario: Move dataset uses canonical path
- **WHEN** frontend calls the move dataset function
- **THEN** the request is sent to `POST /dataset/move` (Vite proxies to `/api/dataset/move`)

#### Scenario: Delete dataset uses canonical path
- **WHEN** frontend calls the delete dataset function with id 123
- **THEN** the request is sent to `POST /dataset/delete/123` (Vite proxies to `/api/dataset/delete/123`)

#### Scenario: Permanently delete dataset uses canonical path
- **WHEN** frontend calls the permanent delete dataset function with id 456
- **THEN** the request is sent to `POST /dataset/perDelete/456` (Vite proxies to `/api/dataset/perDelete/456`)
