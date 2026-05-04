## ADDED Requirements

### Requirement: Task Definition CRUD
The system SHALL provide create, read, update, and delete operations for data filling task definitions.

#### Scenario: Create a new task
- **WHEN** an authenticated user submits a `POST /data-filling/task/save` with a task body that has no `id` field
- **THEN** the system creates a new `data_filling_task` row with the provided form_id, name, recipient configuration (reci_flag_list, uid_list, rid_list), schedule configuration (fill_type, fit_type, fit_column, rate_type, rate_val, one_time_type), time range (start_time, end_time), publish range settings (publish_range_time, publish_range_time_type), and form settings (form_ext_setting, form_filter_setting)
- **AND** the system sets `status=0` (stopped), `create_by` and `create_time` from the current user and timestamp
- **AND** the system returns the new task ID as a `int64`

#### Scenario: Update an existing task
- **WHEN** an authenticated user submits a `POST /data-filling/task/save` with a task body that includes an `id` field
- **THEN** the system updates the existing `data_filling_task` row with the provided fields
- **AND** sets `update_by` and `update_time` from the current user and timestamp
- **AND** if the task has `status=1` (started), the system stops the running cron job, updates the row, and restarts the cron job with the new schedule
- **AND** returns the task ID

#### Scenario: Get task detail
- **WHEN** an authenticated user submits a `GET /data-filling/task/info/{taskId}`
- **THEN** the system returns the full `DataFillingTask` row including all fields
- **AND** if the task does not exist, returns a 404 error

#### Scenario: List tasks with pagination
- **WHEN** an authenticated user submits a `POST /data-filling/form/{formId}/task/page/{goPage}/{pageSize}`
- **THEN** the system returns a paginated list of tasks filtered by `form_id`
- **AND** each task includes: id, name, status, rate_type, rate_val, start_time, end_time, last_exec_status, last_exec_time, next_exec_time, create_by, create_time, update_by, update_time
- **AND** the response includes total count

#### Scenario: Batch delete tasks
- **WHEN** an authenticated user submits a `POST /data-filling/form/{formId}/task/delete` with a JSON body containing a list of task IDs
- **THEN** the system stops any running cron jobs for the specified tasks
- **AND** deletes the `data_filling_task` rows
- **AND** cascade deletes all associated `data_filling_sub_task` and `data_filling_sub_instance` rows
- **AND** returns success

### Requirement: Task Scheduling Lifecycle
The system SHALL manage the lifecycle of scheduled tasks with start, stop, and immediate execution.

#### Scenario: Start a task
- **WHEN** an authenticated user submits a `GET /data-filling/form/{formId}/task/{id}/start`
- **THEN** the system validates the task exists and has `status=0` (stopped)
- **AND** computes the next execution time based on rate_type and rate_val
- **AND** registers a cron job in the in-memory scheduler
- **AND** updates the task row: `status=1`, `next_exec_time` to the computed next fire time
- **AND** returns success

#### Scenario: Stop a task
- **WHEN** an authenticated user submits a `GET /data-filling/form/{formId}/task/{id}/stop`
- **THEN** the system removes the cron job from the in-memory scheduler
- **AND** updates the task row: `status=0`
- **AND** returns success

#### Scenario: Execute a task immediately
- **WHEN** an authenticated user submits a `POST /data-filling/task/executeNow` with a task ID
- **THEN** the system triggers the sub-task generation flow immediately regardless of the task's status or schedule
- **AND** creates a SubTask, resolves recipients, and creates SubInstances as described in the Sub-task Generation requirement
- **AND** returns success

#### Scenario: Start a non-existent task
- **WHEN** an authenticated user submits a start request for a task ID that does not exist
- **THEN** the system returns a 404 error

#### Scenario: Start an already started task
- **WHEN** an authenticated user submits a start request for a task with `status=1`
- **THEN** the system returns an error indicating the task is already running

#### Scenario: Stop a non-existent task
- **WHEN** an authenticated user submits a stop request for a task ID that does not exist
- **THEN** the system returns a 404 error

### Requirement: Sub-task Generation on Schedule Fire
The system SHALL generate sub-tasks and sub-instances when a scheduled task fires.

#### Scenario: Cron fire creates sub-task and sub-instances
- **WHEN** a registered cron job fires for a task with `status=1`
- **THEN** the system creates a `data_filling_sub_task` row with: `task_id`, `start_time` (current time), `end_time` (computed from task's `publish_range_time` and `publish_range_time_type`), `exec_status=0`, `status=1`, `fill_type` from the task
- **AND** the system resolves the recipient user list by combining users from `uid_list` and expanding all users belonging to roles in `rid_list`
- **AND** deduplicates the user list
- **AND** for each resolved user, creates a `data_filling_sub_instance` row with: `task_id`, `pid` (sub-task ID), `uid` (user ID), `form_id` from the task, `status=0` (OPEN)
- **AND** updates the sub-task's `total_count`, `unfinished_count` (both set to instance count), `total_user_count`, `unfinished_user_count` (both set to deduplicated user count)
- **AND** updates the task's `last_exec_time` to current time, `last_exec_status=0` (success), and `next_exec_time` to the next computed fire time
- **AND** the sub-task creation and all sub-instance inserts execute within a single database transaction

#### Scenario: Cron fire with no recipients
- **WHEN** a cron job fires and the resolved recipient list is empty (both uid_list and rid_list are empty or resolve to no users)
- **THEN** the system creates the sub-task row with `total_count=0`, `unfinished_count=0`, `total_user_count=0`, `unfinished_user_count=0`
- **AND** no sub-instances are created
- **AND** the task's `last_exec_status` is set to 0 (success)

#### Scenario: Cron fire with large recipient list
- **WHEN** a cron job fires and the resolved recipient list exceeds 500 users
- **THEN** sub-instance inserts are batched into chunks of 500
- **AND** all chunks execute within the same database transaction

### Requirement: Sub-task Management
The system SHALL provide read and delete operations for sub-tasks and sub-instances.

#### Scenario: List sub-tasks with pagination
- **WHEN** an authenticated user submits a `POST /data-filling/sub-task/page/{goPage}/{pageSize}` with a task_id filter
- **THEN** the system returns a paginated list of `data_filling_sub_task` rows filtered by `task_id`
- **AND** each sub-task includes: id, task_id, start_time, end_time, exec_status, status, total_count, unfinished_count, total_user_count, unfinished_user_count, fill_type
- **AND** the response includes total count

#### Scenario: Batch delete sub-tasks
- **WHEN** an authenticated user submits a `POST /data-filling/form/{formId}/sub-task/delete` with a JSON body containing a list of sub-task IDs
- **THEN** the system deletes the specified `data_filling_sub_task` rows
- **AND** cascade deletes all associated `data_filling_sub_instance` rows (where `pid` matches the deleted sub-task IDs)
- **AND** returns success

#### Scenario: List sub-task users
- **WHEN** an authenticated user submits a `GET /data-filling/sub-task/{id}/users/list/{type}` where type is "all", "finished", or "unfinished"
- **THEN** the system queries `data_filling_sub_instance` rows for the given sub-task ID
- **AND** filters by status: type="finished" returns status=1, type="unfinished" returns status=0, type="all" returns both
- **AND** returns a list of users with their uid, finish_time, status, and data_id

### Requirement: Task Domain Models
The system SHALL define GORM domain models for task, sub-task, and sub-instance tables.

#### Scenario: DataFillingTask model structure
- **WHEN** the system initializes the data filling infrastructure
- **THEN** the `DataFillingTask` model maps to a `data_filling_task` table with columns: id (bigint auto-increment PK), form_id (bigint, indexed), name (varchar 255), reci_flag_list (json), uid_list (json), rid_list (json), fill_type (int), fit_type (int), fit_column (varchar 255), rate_type (int), rate_val (varchar 255), one_time_type (int), start_time (bigint), end_time (bigint), publish_range_time (int), publish_range_time_type (int), status (int: 0=stopped, 1=started), last_exec_status (int), last_exec_time (bigint), next_exec_time (bigint), create_by (bigint), create_time (bigint), update_by (bigint), update_time (bigint), form_ext_setting (text/json), form_filter_setting (text/json)
- **AND** the table has an index on form_id

#### Scenario: DataFillingSubTask model structure
- **WHEN** the system initializes the data filling infrastructure
- **THEN** the `DataFillingSubTask` model maps to a `data_filling_sub_task` table with columns: id (bigint auto-increment PK), task_id (bigint, indexed), start_time (bigint), end_time (bigint), exec_status (int), status (int: 0=expired, 1=active), total_count (int), unfinished_count (int), total_user_count (int), unfinished_user_count (int), fill_type (int)
- **AND** the table has an index on task_id

#### Scenario: DataFillingSubInstance model structure
- **WHEN** the system initializes the data filling infrastructure
- **THEN** the `DataFillingSubInstance` model maps to a `data_filling_sub_instance` table with columns: id (bigint auto-increment PK), task_id (bigint), pid (bigint, FK to sub_task), uid (bigint), form_id (bigint), data_id (varchar 64), finish_time (bigint), status (int: 0=OPEN, 1=FINISHED)
- **AND** the table has a composite index on (task_id, pid) and an index on uid

### Requirement: Cron Scheduler Service
The system SHALL implement a scheduler service using robfig/cron/v3 that manages task execution lifecycle.

#### Scenario: Scheduler initialization on startup
- **WHEN** the application starts
- **THEN** the scheduler service creates a new `cron.Cron` instance
- **AND** queries all `data_filling_task` rows with `status=1` (started)
- **AND** registers a cron job for each started task based on its rate_type and rate_val
- **AND** for each task whose `next_exec_time` has already passed, triggers immediate execution

#### Scenario: Register cron job for rateType=0 (cron expression)
- **WHEN** the scheduler registers a task with `rate_type=0`
- **THEN** it passes `rate_val` as the cron expression to `cron.AddFunc`
- **AND** stores the returned `cron.EntryID` in the internal map keyed by task ID

#### Scenario: Register cron job for rateType=1 (daily interval)
- **WHEN** the scheduler registers a task with `rate_type=1`
- **THEN** it computes the next fire time as `start_time + N * 86400` where N is the integer value of `rate_val`
- **AND** schedules a one-shot function at the computed time
- **AND** after execution, re-computes and re-registers the next occurrence

#### Scenario: Register cron job for rateType=2 (weekly)
- **WHEN** the scheduler registers a task with `rate_type=2`
- **THEN** it parses `rate_val` as a JSON array of day-of-week integers (0=Sunday through 6=Saturday)
- **AND** registers a daily cron job that checks if the current day is in the array before executing

#### Scenario: Register cron job for rateType=3 (monthly)
- **WHEN** the scheduler registers a task with `rate_type=3`
- **THEN** it parses `rate_val` as a JSON array of day-of-month integers (1-31)
- **AND** registers a daily cron job that checks if the current day is in the array before executing

#### Scenario: Scheduler graceful shutdown
- **WHEN** the application shuts down
- **THEN** the scheduler calls `cron.Stop()` to stop all registered jobs
- **AND** waits for any currently running jobs to complete
