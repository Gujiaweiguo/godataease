## MODIFIED Requirements

### Requirement: Sub-task Management
The system SHALL provide read and delete operations for sub-tasks and sub-instances, and SHALL track sub-instance status transitions driven by filler actions.

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

#### Scenario: SubInstance status transitions to FINISHED on filler save
- **WHEN** a filler saves or appends data via the user-task endpoints
- **THEN** the SubInstance transitions from `status=0` (OPEN) to `status=1` (FINISHED)
- **AND** the SubInstance's `finish_time` is set to the current timestamp
- **AND** the parent SubTask's `unfinished_count` is decremented (only on the first transition, not on subsequent re-saves)

#### Scenario: SubInstance status remains FINISHED on data deletion
- **WHEN** a filler deletes specific data rows via the user-task delete endpoint
- **THEN** the SubInstance status remains unchanged (FINISHED stays FINISHED)
- **AND** no status transition or count update occurs
