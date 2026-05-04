## Context

DataFilling slices 1-3 established the form management, DML operations, and admin task management layers. The existing codebase has:

- **Domain types** in `internal/domain/datafilling/datafilling.go`: `DataFillingTask`, `DataFillingSubTask`, `DataFillingSubInstance` with constants for status values (`SubInstanceStatusOpen=0`, `SubInstanceStatusFinished=1`)
- **Repository** in `internal/repository/datafilling_repo.go`: CRUD for forms, tasks, sub-tasks, and sub-instances
- **Service** in `internal/service/datafilling_service.go` and `datafilling_ddl.go`: business logic for admin operations
- **Handler** in `internal/transport/http/handler/datafilling_handler.go`: Gin route registration

The filler (user) perspective is missing. Users assigned to tasks need endpoints to view their work, open forms, and submit data.

## Goals / Non-Goals

**Goals:**
- Expose 6 REST endpoints for filler-role users to interact with their assigned tasks
- Join SubInstance + SubTask + Task data to give users a meaningful task list view
- Handle SubInstance status transitions (OPEN to FINISHED) on save/append
- Enforce UID-level access control so users only see their own SubInstances

**Non-Goals:**
- Admin task management endpoints (already done in slice 3)
- Form schema CRUD or DDL operations (slice 1)
- Generic DML read/write on form tables (slice 2)
- Frontend UI for user task views (future slice)
- Email/notification integration for task assignment

## Decisions

### 1. Separate handler file for user-task endpoints

**Decision**: Create `datafilling_user_task_handler.go` alongside the existing `datafilling_handler.go`.

**Rationale**: The existing handler handles admin endpoints. User-task endpoints have distinct authorization (filler vs admin) and distinct route prefixes (`/user-task/` vs `/task/`). A separate file keeps concerns clean without requiring structural changes.

**Alternative**: Add methods to existing handler. Rejected because the existing handler is already substantial and the user-task domain is self-contained.

### 2. Repository methods use GORM joins for aggregated user task view

**Decision**: `ListSubInstancesByUID` will join `data_filling_sub_instance` with `data_filling_sub_task` and `data_filling_task` to produce the `UserTaskVO` projection in a single query.

**Rationale**: The user task list needs task name, date range, and status from both SubTask and Task tables. A join query is more efficient than N+1 lookups. GORM's `Joins` and `Select` support this well.

**Alternative**: Multiple queries with in-memory assembly. Rejected for performance reasons on paginated lists.

### 3. Status transition on save/append only

**Decision**: SubInstance transitions from OPEN to FINISHED only when `SaveUserTaskData` or `AppendUserTaskData` is called. Deleting data rows does not revert status.

**Rationale**: Matches the Java original behavior. A filler marking work as done via save/append is the completion signal. Deleting specific rows is a data correction, not a status change.

### 4. Authorization via UID match on SubInstance

**Decision**: Each endpoint extracts `userID` from the Gin context (set by auth middleware) and filters queries by `WHERE uid = ?`. For single-instance endpoints, verify ownership before returning data.

**Rationale**: Simple, consistent with the existing auth pattern. SubInstance rows have a `uid` column that ties them to specific users.

## Risks / Trade-offs

- **[Performance on large task lists]** Users with many assigned tasks across many sub-tasks could produce large join result sets. Mitigation: enforce pagination with reasonable page size limits (max 100).
- **[Race condition on concurrent saves]** Two fillers editing the same SubInstance simultaneously. Mitigation: SubInstance is per-UID (each user gets their own), so this is unlikely. If it occurs, last-write-wins is acceptable.
- **[SubInstance without SubTask]** Orphaned sub-instances if a sub-task is deleted while a filler has an open instance. Mitigation: use INNER JOIN so orphaned instances are excluded from listings.
