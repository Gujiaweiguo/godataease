## Why

DataFilling forms (Slices 1-2) let users define schemas and submit data manually. Real workflows need scheduled task definitions that automatically create sub-tasks on a cadence, assign them to specific users or roles, and track completion. Without this, every data collection round must be coordinated manually, which doesn't scale past a handful of forms.

## What Changes

- Add three new database tables: `data_filling_task` (task template/schedule), `data_filling_sub_task` (execution run per trigger), and `data_filling_sub_instance` (row-level user assignment within a sub-task).
- Add task CRUD: create, update, delete, list (paginated), and detail endpoints.
- Add task lifecycle: start (registers cron job), stop (removes cron job), execute-now (immediate trigger).
- Add sub-task management: list sub-tasks, delete sub-tasks, list assigned users per sub-task.
- Add a Go-native cron scheduler (robfig/cron/v3) that replaces the Java Quartz approach. On each fire, the scheduler creates a SubTask, resolves recipients, and generates SubInstances.
- Extend the existing `datafilling` domain model, service, and HTTP handler with task-related types and methods.
- Add new files: `datafilling_scheduler.go` (cron service), `task_repo.go` (repository).

## Capabilities

### New Capabilities
- `data-filling-tasks`: Task definition CRUD, scheduling lifecycle (start/stop/execute-now), cron-based sub-task generation, sub-task and sub-instance management.

### Modified Capabilities
- `data-filling`: Add task/sub-task/sub-instance domain models and request/response types to the existing datafilling module. Extend the HTTP handler with task-related routes.

## Impact

- **Database**: Three new tables auto-migrated on startup. No schema changes to existing tables.
- **Backend code**: New domain structs in `internal/domain/datafilling/`, new repository in `internal/repository/task_repo.go`, new scheduler in `internal/service/datafilling_scheduler.go`, extended handler routes.
- **API**: 11 new endpoints under `/data-filling`. All new, no existing endpoints modified.
- **Dependencies**: Adds `robfig/cron/v3` to `go.mod`.
- **Rollback**: Remove the three tables, revert route registration, remove the new files. No data loss in existing tables.
