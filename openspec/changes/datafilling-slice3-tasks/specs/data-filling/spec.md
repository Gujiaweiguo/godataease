## ADDED Requirements

### Requirement: Data Filling Task Request and Response Types
The system SHALL define request and response types for task and sub-task operations.

#### Scenario: Task save request structure
- **WHEN** the frontend submits a task save request
- **THEN** the request includes: id (int64, optional for create), form_id (int64), name (string), reci_flag_list (json array), uid_list (json array of int64), rid_list (json array of int64), fill_type (int), fit_type (int), fit_column (string), rate_type (int), rate_val (string), one_time_type (int), start_time (int64), end_time (int64), publish_range_time (int), publish_range_time_type (int), form_ext_setting (string/json), form_filter_setting (string/json)

#### Scenario: Task list response structure
- **WHEN** the system returns a paginated task list
- **THEN** each task includes: id, name, form_id, status, rate_type, rate_val, start_time, end_time, last_exec_status, last_exec_time, next_exec_time, create_by, create_time, update_by, update_time
- **AND** the response includes total count, current page, and page size

#### Scenario: Sub-task list response structure
- **WHEN** the system returns a paginated sub-task list
- **THEN** each sub-task includes: id, task_id, start_time, end_time, exec_status, status, total_count, unfinished_count, total_user_count, unfinished_user_count, fill_type
- **AND** the response includes total count, current page, and page size

#### Scenario: Sub-task user list response structure
- **WHEN** the system returns users for a sub-task
- **THEN** each user entry includes: uid (int64), finish_time (int64), status (int: 0=OPEN, 1=FINISHED), data_id (string)

### Requirement: Data Filling Task HTTP Routes
The system SHALL register task-related HTTP routes under the `/data-filling` group.

#### Scenario: Task routes registered on startup
- **WHEN** the application starts and registers data filling routes
- **THEN** the following routes are added to the Gin engine:
  - `GET /data-filling/task/info/:taskId` → GetTaskInfo handler
  - `POST /data-filling/task/save` → SaveTask handler
  - `POST /data-filling/task/executeNow` → ExecuteNowTask handler
  - `POST /data-filling/form/:formId/task/page/:goPage/:pageSize` → TaskPageList handler
  - `GET /data-filling/form/:formId/task/:id/stop` → StopTask handler
  - `GET /data-filling/form/:formId/task/:id/start` → StartTask handler
  - `POST /data-filling/form/:formId/task/delete` → DeleteTasks handler
  - `POST /data-filling/sub-task/page/:goPage/:pageSize` → SubTaskPageList handler
  - `POST /data-filling/form/:formId/sub-task/delete` → DeleteSubTasks handler
  - `GET /data-filling/sub-task/:id/users/list/:type` → SubTaskUsersList handler
- **AND** all routes require authentication
- **AND** all routes use the existing data filling middleware chain
