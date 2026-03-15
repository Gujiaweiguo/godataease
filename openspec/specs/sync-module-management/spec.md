# sync-module-management Specification

## Purpose
Define Go backend sync center requirements for sync datasource management, sync task lifecycle, task logs, and summary visibility.

## Requirements

### Requirement: Sync Datasource Management Parity
The Go backend SHALL provide the sync datasource API surface required by the frontend sync center.

#### Scenario: Page sync datasources by source or target role
- **WHEN** a client requests `/sync/datasource/source/pager/:page/:limit` or `/sync/datasource/target/pager/:page/:limit`
- **THEN** the backend returns deterministic paging data for sync datasources matching the requested role and filters

#### Scenario: Manage sync datasource connection definitions
- **WHEN** a client calls `/sync/datasource/save`, `/sync/datasource/update`, `/sync/datasource/get/:id`, `/sync/datasource/delete/:id`, or `/sync/datasource/batchDel`
- **THEN** the backend persists or returns sync datasource configuration data in a contract compatible with the frontend editor flow

#### Scenario: Validate sync datasource metadata and schema
- **WHEN** a client calls `/sync/datasource/validate`, `/sync/datasource/validate/:id`, `/sync/datasource/getSchema`, `/sync/datasource/fields`, `/sync/datasource/table/list/:dsId`, or `/sync/datasource/latestUse/:sourceType`
- **THEN** the backend returns validation, schema, field, table, and latest-use data needed to author sync tasks

### Requirement: Sync Task Lifecycle Management Parity
The Go backend SHALL provide sync task CRUD and lifecycle APIs required by the frontend sync task center.

#### Scenario: Page and inspect sync tasks
- **WHEN** a client calls `/sync/task/pager/:current/:size` or `/sync/task/get/:taskId`
- **THEN** the backend returns persisted sync task data with deterministic paging and task-detail fields

#### Scenario: Create, update, and remove sync tasks
- **WHEN** a client calls `/sync/task/add`, `/sync/task/update`, `/sync/task/remove/:taskId`, or `/sync/task/batch/del`
- **THEN** the backend persists the requested task changes and returns compatibility-safe success or failure semantics

#### Scenario: Execute and control sync task lifecycle
- **WHEN** a client calls `/sync/task/execute/:id`, `/sync/task/start/:id`, or `/sync/task/stop/:id`
- **THEN** the backend triggers or controls the underlying sync orchestration and returns deterministically mapped lifecycle status

#### Scenario: List datasource options for task authoring
- **WHEN** a client calls `/sync/datasource/list/:type`
- **THEN** the backend returns datasource options compatible with the sync task editor for the requested datasource type

### Requirement: Sync Task Log and Summary Visibility
The Go backend SHALL expose sync task logs and sync summary data required by the frontend operational views.

#### Scenario: Page and inspect sync task logs
- **WHEN** a client calls `/sync/task/log/pager/:current/:size` or `/sync/task/log/detail/:logId/:fromLineNum`
- **THEN** the backend returns persisted task log metadata and incremental detail content in a deterministic format

#### Scenario: Manage sync task logs
- **WHEN** a client calls `/sync/task/log/delete/:logId`, `/sync/task/log/clear`, or `/sync/task/log/terminationTask/:logId`
- **THEN** the backend deletes, clears, or terminates task-log-related state with deterministic success semantics

#### Scenario: Return sync summary metrics
- **WHEN** a client calls `/sync/summary/resourceCount` or `/sync/summary/logChartData`
- **THEN** the backend returns summary counters and chart-ready operational data required by the frontend sync dashboard
