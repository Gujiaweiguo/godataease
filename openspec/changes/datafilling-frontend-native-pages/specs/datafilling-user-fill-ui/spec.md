## ADDED Requirements

### Requirement: User task list display
The system SHALL display a paginated list of data-filling tasks assigned to the current user, using the `getUserTaskPageList` API.

#### Scenario: View assigned tasks
- **WHEN** the current user navigates to the DataFilling fill section from the workbranch or dedicated route
- **THEN** the system calls `getUserTaskPageList` with default pagination and empty filter
- **AND** renders each assigned task with task name, form name, start/end time, status, completion progress (finish count / total count), and expired flag

#### Scenario: Filter tasks by name
- **WHEN** user enters a task name search term
- **THEN** the system calls `getUserTaskPageList` with the `taskName` filter parameter
- **AND** only matching tasks are displayed

#### Scenario: Todo count badge
- **WHEN** the workbranch shortcut loads the data-filling tab
- **THEN** the system calls `getUserTaskTodoCount` to retrieve the count of unfinished tasks
- **AND** displays the count as a badge on the data-filling tab or shortcut

### Requirement: User task fill form
The system SHALL present a fill form for the user to view, edit, and submit data rows for an assigned task, using `getUserTaskData` to load the form and data.

#### Scenario: Open fill form for a task
- **WHEN** user clicks on an assigned task from the task list
- **THEN** the system calls `getUserTaskData` with the sub-task id
- **AND** renders the form fields defined in `form` with the existing data rows from `dataIds` and `subInstances`
- **AND** displays the form title from `formTitle`

#### Scenario: Submit filled data
- **WHEN** user fills in or modifies data rows and clicks "Submit"
- **THEN** the system calls `saveUserTaskData` with the sub-task id and the updated data array
- **AND** shows a success confirmation and returns to the task list

#### Scenario: Append data rows
- **WHEN** user adds new rows to the fill form
- **THEN** the system includes the new rows in the data array sent to `saveUserTaskData` or `appendUserTaskData`

#### Scenario: Delete a data row from fill form
- **WHEN** user deletes a data row in the fill form
- **THEN** the system calls `deleteUserTaskData` with the task instance id and data id
- **AND** the row is removed from the form view

### Requirement: Excel append in user fill
The system SHALL allow users to append data via Excel upload during task filling, using the `uploadExcelFile` and `userTaskConfirmUpload` APIs.

#### Scenario: Upload Excel to append data
- **WHEN** user selects an Excel file for upload in the fill form
- **THEN** the system calls `uploadExcelFile` with the form id and file
- **AND** displays a preview of parsed data rows

#### Scenario: Confirm Excel upload in user task
- **WHEN** user confirms the uploaded Excel preview
- **THEN** the system calls `userTaskConfirmUpload` with the sub-task id, form id, and upload id
- **AND** the appended rows appear in the fill form data

### Requirement: Task completion and expiration
The system SHALL display task completion status and respect task expiration.

#### Scenario: Display expired tasks
- **WHEN** the task list contains tasks past their end time
- **THEN** those tasks are visually marked as expired
- **AND** the fill action is disabled for expired tasks

#### Scenario: Mark task as finished
- **WHEN** user completes all required data rows and submits
- **THEN** the task status updates to finished in the task list
- **AND** the task moves to the "Completed" section or filter

#### Scenario: View fill type indicator
- **WHEN** the task has a specific fill type (e.g., single fill vs. multi-fill)
- **THEN** the fill form enforces the corresponding behavior (allowing or preventing multiple submissions based on `fillType`)
