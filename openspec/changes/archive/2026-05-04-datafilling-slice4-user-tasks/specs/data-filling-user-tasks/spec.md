## ADDED Requirements

### Requirement: User Task Paginated List
The system SHALL provide a paginated list of tasks assigned to the current user, joined with sub-task and task metadata.

#### Scenario: List assigned tasks with default pagination
- **WHEN** an authenticated user submits a `GET /data-filling/user-task/list/{goPage}/{pageSize}`
- **THEN** the system queries `data_filling_sub_instance` rows where `uid` matches the current user's ID
- **AND** joins each SubInstance with its parent SubTask (via `pid`) and grandparent Task (via `task_id`)
- **AND** returns a paginated result with each item containing: subTaskID, taskID, taskName, formID, startTime, endTime, status (SubInstance status), finishTime
- **AND** results are ordered by sub-task start_time descending
- **AND** the response includes total count, current page, and page size

#### Scenario: List with task type filter
- **WHEN** an authenticated user submits a `GET /data-filling/user-task/list/{goPage}/{pageSize}?type={typeValue}`
- **THEN** the system applies an additional filter where type=0 returns only unfinished SubInstances (status=0) and type=1 returns only finished SubInstances (status=1)
- **AND** returns the same paginated structure

#### Scenario: List with task name keyword filter
- **WHEN** an authenticated user submits a `GET /data-filling/user-task/list/{goPage}/{pageSize}?taskName={keyword}`
- **THEN** the system applies a LIKE filter on the task name
- **AND** returns matching paginated results

#### Scenario: User with no assigned tasks
- **WHEN** an authenticated user requests the task list and has no SubInstance rows
- **THEN** the system returns an empty list with total=0

### Requirement: User Task Todo Count
The system SHALL return the count of unfinished SubInstances for the current user.

#### Scenario: Get todo count
- **WHEN** an authenticated user submits a `GET /data-filling/user-task/todo-count`
- **THEN** the system counts `data_filling_sub_instance` rows where `uid` matches the current user's ID and `status=0` (OPEN)
- **AND** returns the count as an integer

#### Scenario: User with no unfinished tasks
- **WHEN** an authenticated user requests todo count and has no open SubInstances
- **THEN** the system returns 0

### Requirement: User Task Data Retrieval
The system SHALL return a single SubInstance with its associated form structure and any existing fill data.

#### Scenario: Get task data for an assigned sub-task
- **WHEN** an authenticated user submits a `GET /data-filling/user-task/data/{subTaskId}`
- **AND** a SubInstance exists where `pid` matches the subTaskId and `uid` matches the current user's ID
- **THEN** the system returns: SubInstance ID, SubInstance status, finishTime, the form structure (from the parent Task's Form), and any existing fill data from the form table
- **AND** the fill data is fetched from the physical form table using the SubInstance's `data_id` as the primary key filter

#### Scenario: Get task data for unassigned sub-task
- **WHEN** an authenticated user requests data for a subTaskId that has no SubInstance matching the user's UID
- **THEN** the system returns a 403 error

#### Scenario: Get task data for non-existent sub-task
- **WHEN** an authenticated user requests data for a subTaskId that does not exist
- **THEN** the system returns a 404 error

### Requirement: Save User Task Data
The system SHALL allow a filler to save fill data for their SubInstance, transitioning its status to FINISHED.

#### Scenario: Save data for an open SubInstance
- **WHEN** an authenticated user submits a `POST /data-filling/user-task/save` with body containing subTaskId and data payload
- **AND** a SubInstance exists matching the subTaskId and user's UID
- **THEN** the system writes the data payload to the physical form table (insert or update based on whether rows exist for this SubInstance's `data_id`)
- **AND** updates the SubInstance: `status=1` (FINISHED), `finish_time` to current timestamp
- **AND** updates the parent SubTask's `unfinished_count` by decrementing it
- **AND** returns success

#### Scenario: Save data for already finished SubInstance
- **WHEN** an authenticated user submits save for a SubInstance that already has `status=1` (FINISHED)
- **THEN** the system updates the data in the form table
- **AND** updates the SubInstance's `finish_time` to the current timestamp (re-save)
- **AND** does not decrement the SubTask's `unfinished_count` again
- **AND** returns success

#### Scenario: Save data for another user's SubInstance
- **WHEN** an authenticated user submits save for a subTaskId that has no SubInstance matching their UID
- **THEN** the system returns a 403 error

### Requirement: Append User Task Data
The system SHALL allow a filler to append new data rows for their SubInstance, transitioning its status to FINISHED.

#### Scenario: Append data for an open SubInstance
- **WHEN** an authenticated user submits a `POST /data-filling/user-task/append` with body containing subTaskId and data rows array
- **AND** a SubInstance exists matching the subTaskId and user's UID
- **THEN** the system inserts the data rows into the physical form table
- **AND** associates the inserted rows with the SubInstance's `data_id`
- **AND** updates the SubInstance: `status=1` (FINISHED), `finish_time` to current timestamp
- **AND** updates the parent SubTask's `unfinished_count` by decrementing it (if not already FINISHED)
- **AND** returns success

#### Scenario: Append data for unassigned SubInstance
- **WHEN** an authenticated user submits append for a subTaskId with no matching SubInstance for their UID
- **THEN** the system returns a 403 error

### Requirement: Delete User Task Data
The system SHALL allow a filler to delete specific fill data rows from their SubInstance.

#### Scenario: Delete specific data rows
- **WHEN** an authenticated user submits a `POST /data-filling/user-task/delete` with body containing subTaskId and a list of data row IDs
- **AND** a SubInstance exists matching the subTaskId and user's UID
- **THEN** the system deletes the specified rows from the physical form table
- **AND** the deleted rows MUST belong to this SubInstance (verified via `data_id` match)
- **AND** the SubInstance status is NOT changed (remains FINISHED or OPEN)
- **AND** returns success

#### Scenario: Delete rows from unassigned SubInstance
- **WHEN** an authenticated user submits delete for a subTaskId with no matching SubInstance for their UID
- **THEN** the system returns a 403 error

#### Scenario: Delete rows not belonging to the SubInstance
- **WHEN** an authenticated user submits delete with row IDs that do not match the SubInstance's `data_id`
- **THEN** the system skips those rows (no-op for non-matching rows)
- **AND** returns success for any rows that were actually deleted
