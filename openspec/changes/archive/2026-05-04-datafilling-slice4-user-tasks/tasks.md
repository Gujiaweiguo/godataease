## 1. Domain Types

- [ ] 1.1 Add `UserTaskPageRequest` struct with fields: Type (int, optional filter: 0=unfinished, 1=finished), TaskName (string, optional LIKE filter) to `internal/domain/datafilling/datafilling.go`
- [ ] 1.2 Add `UserTaskVO` struct with fields: SubTaskID, TaskID, TaskName, FormID, StartTime, EndTime, Status, FinishTime (projected from SubInstance + SubTask + Task join) to `internal/domain/datafilling/datafilling.go`
- [ ] 1.3 Add `UserTaskData` struct with fields: InstanceID, Status, FinishTime, FormData (form structure JSON string), FillData ([]map[string]interface{}) to `internal/domain/datafilling/datafilling.go`
- [ ] 1.4 Add `UserTaskSaveRequest` struct with fields: SubTaskID (int64), Data ([]map[string]interface{}) to `internal/domain/datafilling/datafilling.go`
- [ ] 1.5 Add `UserTaskDeleteRequest` struct with fields: SubTaskID (int64), DataIDs ([]string) to `internal/domain/datafilling/datafilling.go`

## 2. Repository Layer

- [ ] 2.1 Implement `ListSubInstancesByUID(db *gorm.DB, uid int64, page, pageSize int, req *UserTaskPageRequest) ([]*UserTaskVO, int64, error)` with JOIN on sub_instance + sub_task + task, applying type and taskName filters, ordered by sub_task.start_time DESC, with pagination
- [ ] 2.2 Implement `CountOpenSubInstancesByUID(db *gorm.DB, uid int64) (int64, error)` counting sub_instance rows where uid=? and status=0
- [ ] 2.3 Implement `GetSubInstanceByID(db *gorm.DB, id int64) (*DataFillingSubInstance, error)` fetching a single sub_instance by primary key
- [ ] 2.4 Implement `GetSubInstanceByPIDAndUID(db *gorm.DB, pid, uid int64) (*DataFillingSubInstance, error)` fetching a sub_instance by sub_task ID (pid) and user ID
- [ ] 2.5 Implement `UpdateSubInstanceStatus(db *gorm.DB, id int64, status int, finishTime int64) error` updating status and finish_time
- [ ] 2.6 Implement `DecrementSubTaskUnfinishedCount(db *gorm.DB, subTaskID int64) error` atomically decrementing unfinished_count where it is > 0
- [ ] 2.7 Write integration tests for `ListSubInstancesByUID` (with type filter, taskName filter, pagination) and `CountOpenSubInstancesByUID`

## 3. Service Layer

- [ ] 3.1 Implement `UserTaskPageList(ctx context.Context, userID int64, page, pageSize int, req *UserTaskPageRequest) ([]*UserTaskVO, int64, error)` calling the repository list method
- [ ] 3.2 Implement `UserTaskTodoCount(ctx context.Context, userID int64) (int64, error)` calling the repository count method
- [ ] 3.3 Implement `GetUserTaskData(ctx context.Context, userID, subTaskID int64) (*UserTaskData, error)` looking up the SubInstance by pid+uid, fetching the form structure from the form table, and loading any existing fill data
- [ ] 3.4 Implement `SaveUserTaskData(ctx context.Context, userID, subTaskID int64, data []map[string]interface{}) error` verifying ownership, upserting data into the physical form table, transitioning SubInstance to FINISHED, and decrementing unfinished_count
- [ ] 3.5 Implement `AppendUserTaskData(ctx context.Context, userID, subTaskID int64, data []map[string]interface{}) error` verifying ownership, inserting new rows, transitioning SubInstance to FINISHED, and decrementing unfinished_count
- [ ] 3.6 Implement `DeleteUserTaskData(ctx context.Context, userID, subTaskID int64, dataIDs []string) error` verifying ownership, deleting matching rows from the physical form table (scoped to data_id)
- [ ] 3.7 Write unit tests for `SaveUserTaskData` (OPEN to FINISHED transition, already FINISHED re-save, wrong UID rejection), `AppendUserTaskData` (append + transition, wrong UID), and `DeleteUserTaskData` (correct scoping, wrong UID)

## 4. Handler and Routes

- [ ] 4.1 Create `internal/transport/http/handler/datafilling_user_task_handler.go` with a `UserTaskHandler` struct holding a reference to the datafilling service
- [ ] 4.2 Implement `UserTaskList` handler: parse path params (goPage, pageSize) and query params (type, taskName), extract userID from Gin context, call `UserTaskPageList`, return JSON response
- [ ] 4.3 Implement `UserTaskTodoCount` handler: extract userID, call `UserTaskTodoCount`, return count
- [ ] 4.4 Implement `UserTaskData` handler: parse path param (subTaskId), extract userID, call `GetUserTaskData`, return JSON response
- [ ] 4.5 Implement `UserTaskSave` handler: bind JSON body to `UserTaskSaveRequest`, extract userID, call `SaveUserTaskData`, return success
- [ ] 4.6 Implement `UserTaskAppend` handler: bind JSON body to `UserTaskSaveRequest`, extract userID, call `AppendUserTaskData`, return success
- [ ] 4.7 Implement `UserTaskDelete` handler: bind JSON body to `UserTaskDeleteRequest`, extract userID, call `DeleteUserTaskData`, return success
- [ ] 4.8 Register 6 routes under `/data-filling/user-task/` in the router setup: GET `/list/:goPage/:pageSize`, GET `/todo-count`, GET `/data/:subTaskId`, POST `/save`, POST `/append`, POST `/delete`
- [ ] 4.9 Write handler unit tests using httptest for all 6 endpoints covering success cases and 403/404 error cases
