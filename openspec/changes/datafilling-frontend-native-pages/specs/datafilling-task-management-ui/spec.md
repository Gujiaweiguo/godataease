## ADDED Requirements

### Requirement: Task list for a form
The system SHALL display a paginated list of tasks associated with a selected form, using the `getTaskPageList` API.

#### Scenario: View task list
- **WHEN** user navigates to the task management tab for a leaf form
- **THEN** the system calls `getTaskPageList` with the form id, default page, and page size
- **AND** renders each task with name, status, schedule type, last execution status, and next execution time

#### Scenario: Empty task list
- **WHEN** no tasks exist for the selected form
- **THEN** the system displays an empty-state prompt with a "Create Task" action

### Requirement: Task creation and editing
The system SHALL provide a task editor for creating and editing tasks, serializing all fields to the `saveTask` API.

#### Scenario: Create new task
- **WHEN** user clicks "Create Task" for a form
- **THEN** the system opens a task editor form with fields for name, fill type, fit type, fit column, rate type, rate value, start/end time, and recipient configuration
- **AND** the form id is pre-populated from the selected form

#### Scenario: Configure task recipients
- **WHEN** user configures recipient settings in the task editor
- **THEN** the system allows selecting recipient flags (`reciFlagList`), specific users (`uidList`), and roles (`ridList`)
- **AND** the selected values are included in the `saveTask` request

#### Scenario: Configure task scheduling
- **WHEN** user sets rate type (one-time or recurring) and associated time parameters
- **THEN** the system serializes rateType, rateVal, oneTimeType, startTime, endTime, publishRangeTime, and publishRangeTimeType into the save request

#### Scenario: Configure form filter and extension settings
- **WHEN** user sets form extension settings (`formExtSetting`) and form filter settings (`formFilterSetting`)
- **THEN** the system includes these JSON strings in the `saveTask` request

#### Scenario: Edit existing task
- **WHEN** user selects an existing task for editing
- **THEN** the system calls `getTaskInfo` with the task id
- **AND** populates the task editor with the returned task configuration

#### Scenario: Save task
- **WHEN** user submits the task editor form
- **THEN** the system calls `saveTask` with all configured fields
- **AND** refreshes the task list on success

### Requirement: Task lifecycle control
The system SHALL allow starting, stopping, executing immediately, and deleting tasks.

#### Scenario: Start a stopped task
- **WHEN** user clicks "Start" on a stopped task
- **THEN** the system calls `startTask` with the form id and task id
- **AND** the task status updates to running

#### Scenario: Stop a running task
- **WHEN** user clicks "Stop" on a running task
- **THEN** the system calls `stopTask` with the form id and task id
- **AND** the task status updates to stopped

#### Scenario: Execute task immediately
- **WHEN** user clicks "Execute Now" on a task
- **THEN** the system calls `executeNowTask` with the task id
- **AND** the task list reflects the updated last execution status

#### Scenario: Delete tasks in batch
- **WHEN** user selects one or more tasks and clicks "Delete"
- **THEN** the system prompts for confirmation
- **AND** calls `deleteTasks` with the form id and array of task ids
- **AND** the deleted tasks are removed from the list

### Requirement: Sub-task listing and progress tracking
The system SHALL display sub-tasks for a task with execution progress, using the `getSubTaskPageList` API.

#### Scenario: View sub-tasks for a task
- **WHEN** user expands or navigates to sub-tasks for a task
- **THEN** the system calls `getSubTaskPageList` with the task id and pagination parameters
- **AND** renders each sub-task with start time, end time, execution status, total count, unfinished count, and user progress

#### Scenario: Delete sub-tasks
- **WHEN** user selects sub-tasks and triggers delete
- **THEN** the system calls `deleteSubTasks` with the form id and selected sub-task ids
- **AND** the sub-task list refreshes

### Requirement: Sub-task user detail view
The system SHALL display the list of users assigned to a sub-task with their fill status, using the `getSubTaskUsersList` API.

#### Scenario: View user assignments for a sub-task
- **WHEN** user selects a sub-task to view user details
- **THEN** the system calls `getSubTaskUsersList` with the sub-task id and a type filter (finished/unfinished)
- **AND** renders each user assignment with user name, status, finish time, and data id

#### Scenario: Filter users by completion status
- **WHEN** user toggles between "Finished" and "Unfinished" filter tabs
- **THEN** the system calls `getSubTaskUsersList` with the corresponding type parameter
- **AND** the user list updates to show only matching users
