## Why

The DataFilling module migration (slices 1-3) established form CRUD, DML operations (fill data read/write), and admin task management. Users who are assigned tasks have no way to view their pending work, inspect assigned forms, or submit fill data from their perspective. This slice adds the 6 user-facing task endpoints so fillers can interact with tasks assigned to them.

## What Changes

- Add 6 new REST endpoints under `/data-filling/user-task/` for the filler role
- `GET /user-task/list`: paginated list of tasks assigned to the current user, with filters by type and task name
- `GET /user-task/todo-count`: count of open (unfinished) sub-instances for the current user
- `GET /user-task/data/{subTaskId}`: retrieve a single sub-instance with its form structure and any existing fill data
- `POST /user-task/save`: save (insert or update) fill data for a sub-instance, marking it FINISHED
- `POST /user-task/append`: append new fill data rows to a sub-instance, marking it FINISHED
- `POST /user-task/delete`: delete specific fill data rows from a sub-instance
- New domain types: `UserTaskPageRequest`, `UserTaskVO`, `UserTaskData`, `SubInstanceItem`
- New repository methods: `ListSubInstancesByUID`, `CountOpenSubInstancesByUID`, `GetSubInstanceByID`, `UpdateSubInstanceStatus`
- SubInstance status transitions: OPEN to FINISHED when filler saves or appends data

## Capabilities

### New Capabilities
- `data-filling-user-tasks`: User-facing task endpoints for viewing assigned tasks, retrieving form data, and submitting fill data

### Modified Capabilities
- `data-filling-tasks`: Extends task lifecycle with user-side status transitions (SubInstance OPEN to FINISHED)

## Impact

- **Backend**: New handler, service, repository, and domain files under the data-filling module
- **API**: 6 new REST endpoints under `/data-filling/user-task/`
- **Database**: Reads from existing `data_fill_sub_task` and `data_fill_sub_instance` tables (no schema changes)
- **Authorization**: Endpoints require authenticated user; UID-level access control ensures users only see their own assigned SubInstances
- **Rollback**: Remove the 6 routes and delete the new handler/service/repository files; no database migration to reverse
