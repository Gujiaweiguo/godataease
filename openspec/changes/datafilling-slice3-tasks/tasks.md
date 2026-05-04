## 1. Domain Models and Auto-Migration

- [ ] 1.1 Add `DataFillingTask` struct to `internal/domain/datafilling/datafilling.go` with all fields per spec (id, form_id, name, reci_flag_list, uid_list, rid_list, fill_type, fit_type, fit_column, rate_type, rate_val, one_time_type, start_time, end_time, publish_range_time, publish_range_time_type, status, last_exec_status, last_exec_time, next_exec_time, create_by, create_time, update_by, update_time, form_ext_setting, form_filter_setting). Add `TableName()` returning `"data_filling_task"`. Add index on `form_id`.
- [ ] 1.2 Add `DataFillingSubTask` struct to the same file with all fields per spec (id, task_id, start_time, end_time, exec_status, status, total_count, unfinished_count, total_user_count, unfinished_user_count, fill_type). Add `TableName()` returning `"data_filling_sub_task"`. Add index on `task_id`.
- [ ] 1.3 Add `DataFillingSubInstance` struct to the same file with all fields per spec (id, task_id, pid, uid, form_id, data_id, finish_time, status). Add `TableName()` returning `"data_filling_sub_instance"`. Add composite index on `(task_id, pid)` and index on `uid`.
- [ ] 1.4 Add request/response types: `TaskSaveRequest`, `TaskInfoVO`, `TaskPageRequest`, `TaskPageResponse`, `SubTaskPageRequest`, `SubTaskPageResponse`, `SubTaskUsersRequest`, `SubTaskUserItem`, `BatchDeleteTaskRequest`, `BatchDeleteSubTaskRequest`, `ExecuteNowRequest`.
- [ ] 1.5 Add status constants: `TaskStatusStopped=0`, `TaskStatusStarted=1`, `SubTaskStatusExpired=0`, `SubTaskStatusActive=1`, `SubInstanceStatusOpen=0`, `SubInstanceStatusFinished=1`.
- [ ] 1.6 Register the three new models in the auto-migration list (wherever `DataFillingForm` and `DfCommitLog` are registered for `AutoMigrate`). Verify tables are created on startup.

## 2. Task Repository

- [ ] 2.1 Create `internal/repository/task_repo.go` with a `TaskRepository` struct and constructor `NewTaskRepository(db *gorm.DB)`.
- [ ] 2.2 Implement `CreateTask(ctx, task *datafilling.DataFillingTask) error` — inserts a new task row.
- [ ] 2.3 Implement `UpdateTask(ctx, task *datafilling.DataFillingTask) error` — updates by ID.
- [ ] 2.4 Implement `GetTaskByID(ctx, taskID int64) (*datafilling.DataFillingTask, error)` — returns task or gorm ErrRecordNotFound.
- [ ] 2.5 Implement `ListTasksByFormID(ctx, formID int64, page, pageSize int) ([]*datafilling.DataFillingTask, int64, error)` — paginated query with total count.
- [ ] 2.6 Implement `DeleteTasksByIDs(ctx, taskIDs []int64) error` — batch delete by IDs.
- [ ] 2.7 Implement `GetStartedTasks(ctx) ([]*datafilling.DataFillingTask, error)` — returns all tasks with status=1 (for scheduler init).

## 3. Sub-task and Sub-instance Repositories

- [ ] 3.1 Add `SubTaskRepository` to `internal/repository/task_repo.go` (or a new file) with constructor.
- [ ] 3.2 Implement `CreateSubTask(ctx, subTask *datafilling.DataFillingSubTask) error`.
- [ ] 3.3 Implement `UpdateSubTaskCounts(ctx, subTaskID int64, totalCount, unfinishedCount, totalUserCount, unfinishedUserCount int) error`.
- [ ] 3.4 Implement `ListSubTasksByTaskID(ctx, taskID int64, page, pageSize int) ([]*datafilling.DataFillingSubTask, int64, error)` — paginated.
- [ ] 3.5 Implement `DeleteSubTasksByIDs(ctx, subTaskIDs []int64) error` — batch delete.
- [ ] 3.6 Add `SubInstanceRepository` with constructor.
- [ ] 3.7 Implement `BatchCreateSubInstances(ctx, instances []*datafilling.DataFillingSubInstance) error` — batch insert in chunks of 500.
- [ ] 3.8 Implement `DeleteSubInstancesByPID(ctx, pid int64) error` — delete by sub-task ID (for cascade).
- [ ] 3.9 Implement `DeleteSubInstancesByPIDs(ctx, pids []int64) error` — batch delete by sub-task IDs.
- [ ] 3.10 Implement `ListSubInstancesByPID(ctx, pid int64, statusFilter *int) ([]*datafilling.DataFillingSubInstance, error)` — list users for a sub-task.

## 4. Cron Scheduler Service

- [ ] 4.1 Add `robfig/cron/v3` dependency to `go.mod` (`go get github.com/robfig/cron/v3`).
- [ ] 4.2 Create `internal/service/datafilling_scheduler.go` with a `DataFillingScheduler` struct holding a `*cron.Cron` instance, a `map[int64]cron.EntryID` for tracking registered jobs, and references to task repo and the service itself.
- [ ] 4.3 Implement `NewDataFillingScheduler(taskRepo, subTaskRepo, subInstanceRepo, userRepo, roleRepo)` constructor that creates a `cron.New(cron.WithSeconds())` instance.
- [ ] 4.4 Implement `Start(ctx)` method: starts the cron runner, calls `loadAndRegisterStartedTasks`.
- [ ] 4.5 Implement `loadAndRegisterStartedTasks(ctx)` — queries `GetStartedTasks`, registers a cron job for each, checks `next_exec_time` for overdue tasks and fires them immediately.
- [ ] 4.6 Implement `RegisterTask(ctx, taskID)` — reads task from DB, parses rate_type/rate_val, calls `cron.AddFunc` (or schedules one-shot), stores EntryID in the map.
- [ ] 4.7 Implement `UnregisterTask(taskID)` — removes the cron entry by EntryID, deletes from the map.
- [ ] 4.8 Implement `onFire(ctx, taskID)` — the cron callback: creates SubTask, resolves recipients (query users by uid_list + expand rid_list roles), deduplicates, batch-creates SubInstances, updates counts, updates task's last_exec_time/last_exec_status/next_exec_time.
- [ ] 4.9 Implement `resolveRecipients(ctx, uidList, ridList)` helper — combines direct user IDs with role-expanded user IDs, deduplicates.
- [ ] 4.10 Implement `computeNextExecTime(task)` helper — returns next fire time based on rate_type and rate_val.
- [ ] 4.11 Implement `Stop()` — calls `cron.Stop()` for graceful shutdown.

## 5. Task Service Methods

- [ ] 5.1 Add task-related dependencies to `DataFillingService` struct: `taskRepo`, `subTaskRepo`, `subInstanceRepo`, `scheduler`.
- [ ] 5.2 Implement `SaveTask(ctx, req *datafilling.TaskSaveRequest) (int64, error)` — creates or updates task. If updating a started task, stop cron, update, re-register.
- [ ] 5.3 Implement `GetTaskInfo(ctx, taskID int64) (*datafilling.TaskInfoVO, error)` — returns task detail or error.
- [ ] 5.4 Implement `StartTask(ctx, formID, taskID int64) error` — validates task exists + status=0, registers cron, updates status=1 and next_exec_time.
- [ ] 5.5 Implement `StopTask(ctx, formID, taskID int64) error` — unregisters cron, updates status=0.
- [ ] 5.6 Implement `ExecuteNowTask(ctx, taskID int64) error` — calls scheduler.onFire directly.
- [ ] 5.7 Implement `TaskPageList(ctx, formID int64, page, pageSize int) (*datafilling.TaskPageResponse, error)`.
- [ ] 5.8 Implement `DeleteTasks(ctx, formID int64, taskIDs []int64) error` — stops running tasks, deletes sub-instances by task IDs, deletes sub-tasks by task IDs, deletes tasks.
- [ ] 5.9 Implement `SubTaskPageList(ctx, taskID int64, page, pageSize int) (*datafilling.SubTaskPageResponse, error)`.
- [ ] 5.10 Implement `DeleteSubTasks(ctx, formID int64, subTaskIDs []int64) error` — deletes sub-instances by PIDs, deletes sub-tasks.
- [ ] 5.11 Implement `SubTaskUsersList(ctx, subTaskID int64, listType string) ([]*datafilling.SubTaskUserItem, error)` — maps listType to status filter, queries sub-instances.

## 6. HTTP Handlers and Routes

- [ ] 6.1 Add `GetTaskInfo` handler to `datafilling_handler.go` — binds taskID from URL param, calls service, returns JSON.
- [ ] 6.2 Add `SaveTask` handler — binds JSON body to `TaskSaveRequest`, calls service, returns task ID.
- [ ] 6.3 Add `StartTask` handler — binds formID and id from URL params, calls service.
- [ ] 6.4 Add `StopTask` handler — binds formID and id from URL params, calls service.
- [ ] 6.5 Add `ExecuteNowTask` handler — binds JSON body to `ExecuteNowRequest`, calls service.
- [ ] 6.6 Add `TaskPageList` handler — binds formID, goPage, pageSize from URL params, calls service.
- [ ] 6.7 Add `DeleteTasks` handler — binds formID from URL, JSON body for task IDs, calls service.
- [ ] 6.8 Add `SubTaskPageList` handler — binds goPage, pageSize from URL, taskID from JSON body, calls service.
- [ ] 6.9 Add `DeleteSubTasks` handler — binds formID from URL, JSON body for sub-task IDs, calls service.
- [ ] 6.10 Add `SubTaskUsersList` handler — binds subTaskID and type from URL params, calls service.
- [ ] 6.11 Register all 10 routes in the route registration function under the `/data-filling` group.

## 7. Wiring and Initialization

- [ ] 7.1 Wire `TaskRepository`, `SubTaskRepository`, `SubInstanceRepository` in the DI container (or manual wiring in `main.go`/`wire.go`).
- [ ] 7.2 Create `DataFillingScheduler` instance and inject it into `DataFillingService`.
- [ ] 7.3 Call `scheduler.Start(ctx)` during application startup (after DB connection is established and repos are ready).
- [ ] 7.4 Call `scheduler.Stop()` during application shutdown (in the graceful shutdown handler).

## 8. Tests

- [ ] 8.1 Unit tests for `TaskRepository`: create, update, get by ID, list by form ID, delete. Use integration test suite with MySQL.
- [ ] 8.2 Unit tests for `SubTaskRepository` and `SubInstanceRepository`: create, list, delete, cascade delete.
- [ ] 8.3 Unit tests for `DataFillingService` task methods: SaveTask (create + update), StartTask, StopTask, ExecuteNowTask, DeleteTasks. Mock repos.
- [ ] 8.4 Unit tests for `DataFillingScheduler`: register/unregister task, onFire creates correct sub-task + instances, computeNextExecTime for each rateType.
- [ ] 8.5 Handler tests for task endpoints: verify route registration, request binding, response format, error cases (404, already started).
- [ ] 8.6 Integration test: full lifecycle — create task, start task, verify cron fires and creates sub-task/instances, stop task, delete task.

## 9. Verification

- [ ] 9.1 Run `make test` and verify all tests pass.
- [ ] 9.2 Run `golangci-lint run` and fix any lint issues in new code.
- [ ] 9.3 Start the backend locally, verify the three new tables are auto-migrated.
- [ ] 9.4 Test all 10 endpoints via curl or Swagger UI.
