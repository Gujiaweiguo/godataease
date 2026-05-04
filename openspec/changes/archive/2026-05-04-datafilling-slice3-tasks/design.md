## Context

Slices 1-2 delivered form tree CRUD, DDL/DML against external datasources, and commit logging. The Go backend follows a domain/repository/service/handler layering. Forms are stored in `data_filling_forms`, commit logs in `df_commit_log`. The DataFilling module lives under `internal/domain/datafilling/`, `internal/service/datafilling_service.go`, and `internal/transport/http/handler/datafilling_handler.go`.

The Java legacy used Quartz for scheduling. The Go backend has no scheduler yet. This slice adds task scheduling as the first cron-based feature in the Go stack.

## Goals / Non-Goals

**Goals:**
- Define task templates with schedule configurations (cron, daily, weekly, monthly).
- Start/stop tasks to register and deregister cron jobs at runtime.
- On cron fire: create a SubTask, resolve recipients (users + roles), and create SubInstances.
- Provide CRUD and pagination for tasks and sub-tasks.
- Support immediate execution (executeNow) for testing and ad-hoc runs.

**Non-Goals:**
- User-facing task operations (filling out assigned instances, marking complete). That is Slice 4.
- Frontend UI for task management.
- Distributed locking or multi-instance coordination (single-instance deployment for now).
- Email/notification on task assignment (future slice).
- Editing the schedule of a running task (stop then re-create).

## Decisions

### D1: robfig/cron/v3 instead of Quartz

**Choice**: Use `github.com/robfig/cron/v3` for scheduling.

**Rationale**: The Java legacy used Quartz, which is JVM-specific. robfig/cron is the de-facto Go cron library. It supports standard cron expressions, is lightweight, has no external dependencies beyond the standard library, and handles time zones correctly. For a single-instance Go service, it's sufficient.

**Alternative considered**: A Redis-based distributed scheduler (e.g., using Redis SCAN + sorted sets). Overkill for current single-instance deployment. Can be added later if horizontal scaling is needed.

### D2: In-memory cron registry with DB as source of truth

**Choice**: Keep a singleton `*cron.Cron` instance in the scheduler service. On startup, load all tasks with `status=1` (started) from the database and register their cron jobs. On stop, remove the entry ID from the in-memory cron and set `status=0` in the DB.

**Rationale**: The DB is the durable record. The in-memory cron is just the runtime trigger mechanism. If the process restarts, it re-registers from DB state. This avoids drift between DB and runtime.

**Key implication**: Each `DataFillingTask` row gets an associated `cron.EntryID` stored in a `map[int64]cron.EntryID` (taskID → entryID) in the scheduler. This map is not persisted.

### D3: Sub-task generation flow on cron fire

**Choice**: When a cron job fires:
1. Create a `DataFillingSubTask` row with the task's `fill_type`, compute `start_time`/`end_time` from the task's `publish_range_time` and `publish_range_time_type`.
2. Resolve recipients: query users by `uid_list` and expand roles from `rid_list` into user IDs. Deduplicate.
3. For each resolved user, if `fit_type` indicates per-user assignment, create a `DataFillingSubInstance` row.
4. Update the sub-task's `total_count`, `unfinished_count`, `total_user_count`, `unfinished_user_count`.
5. Update the task's `last_exec_time`, `last_exec_status`, and compute `next_exec_time`.

**Rationale**: Mirrors the Java implementation's behavior. The sub-instance represents a single user's assignment within a sub-task run. The `data_id` field on SubInstance starts empty and gets filled when the user completes their assignment (Slice 4).

### D4: rateType handling

| rateType | rate_val | Meaning |
|----------|----------|---------|
| 0 | cron expression | Standard cron (e.g. `0 9 * * 1-5`) |
| 1 | day count interval | Every N days |
| 2 | day-of-week list (JSON array) | Weekly on specified days |
| 3 | day-of-month list (JSON array) | Monthly on specified days |

For rateType=0, pass `rate_val` directly to robfig/cron. For rateType=1/2/3, compute the next run time in Go and register as a one-shot delayed function that re-registers itself.

### D5: New files, minimal changes to existing files

**New files**:
- `internal/service/datafilling_scheduler.go` — cron lifecycle + fire handler
- `internal/repository/task_repo.go` — task, sub-task, sub-instance queries

**Extended files**:
- `internal/domain/datafilling/datafilling.go` — add 3 domain structs + request/response types
- `internal/service/datafilling_service.go` — add task CRUD methods
- `internal/transport/http/handler/datafilling_handler.go` — add task routes

### D6: Transaction boundaries

- Task save (create/update): single transaction for the task row. Sub-tasks and sub-instances are created in separate transactions per fire event to avoid holding long locks.
- Batch delete sub-tasks: wrap in a transaction, cascade delete associated sub-instances.
- Sub-task generation on fire: one transaction for creating the sub-task + all its sub-instances.

## Risks / Trade-offs

- **[Single-instance scheduler]** → No distributed locking. If two instances run, duplicate sub-tasks could be created. Mitigation: document single-instance constraint. Future: add Redis-based distributed lock.
- **[Cron drift on restart]** → If the process restarts, cron jobs are re-registered from DB. Any fire events that were due during downtime are missed. Mitigation: on startup, check if `next_exec_time` has passed for any started task and fire immediately.
- **[Role expansion query cost]** → Expanding `rid_list` into user IDs requires a join query. Mitigation: cache role memberships in Redis with a short TTL if performance becomes an issue.
- **[Large recipient lists]** → A task assigned to thousands of users creates thousands of sub-instance rows in one transaction. Mitigation: batch insert in chunks of 500.
