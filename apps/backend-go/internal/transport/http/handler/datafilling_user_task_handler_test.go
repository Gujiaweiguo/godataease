package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	datafillingdomain "dataease/backend/internal/domain/datafilling"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type userTaskHandlerFakeTaskRepo struct {
	records map[int64]*datafillingdomain.DataFillingTask
}

type userTaskHandlerFakeSubTaskRepo struct {
	records map[int64]*datafillingdomain.DataFillingSubTask
}

type userTaskHandlerFakeSubInstanceRepo struct {
	records []*datafillingdomain.DataFillingSubInstance
}

func (r *userTaskHandlerFakeTaskRepo) CreateTask(ctx context.Context, task *datafillingdomain.DataFillingTask) error {
	_ = ctx
	r.records[task.ID] = task
	return nil
}
func (r *userTaskHandlerFakeTaskRepo) UpdateTask(ctx context.Context, task *datafillingdomain.DataFillingTask) error {
	_ = ctx
	r.records[task.ID] = task
	return nil
}
func (r *userTaskHandlerFakeTaskRepo) GetTaskByID(ctx context.Context, taskID int64) (*datafillingdomain.DataFillingTask, error) {
	_ = ctx
	row := r.records[taskID]
	if row == nil {
		return nil, gorm.ErrRecordNotFound
	}
	cloned := *row
	return &cloned, nil
}
func (r *userTaskHandlerFakeTaskRepo) ListTasksByFormID(ctx context.Context, formID int64, page, pageSize int) ([]*datafillingdomain.DataFillingTask, int64, error) {
	_ = ctx
	_ = formID
	_ = page
	_ = pageSize
	return nil, 0, nil
}
func (r *userTaskHandlerFakeTaskRepo) DeleteTasksByIDs(ctx context.Context, taskIDs []int64) error {
	_ = ctx
	_ = taskIDs
	return nil
}
func (r *userTaskHandlerFakeTaskRepo) GetStartedTasks(ctx context.Context) ([]*datafillingdomain.DataFillingTask, error) {
	_ = ctx
	return nil, nil
}

func (r *userTaskHandlerFakeSubTaskRepo) CreateSubTask(ctx context.Context, subTask *datafillingdomain.DataFillingSubTask) error {
	_ = ctx
	r.records[subTask.ID] = subTask
	return nil
}
func (r *userTaskHandlerFakeSubTaskRepo) GetSubTaskByID(ctx context.Context, subTaskID int64) (*datafillingdomain.DataFillingSubTask, error) {
	_ = ctx
	row := r.records[subTaskID]
	if row == nil {
		return nil, gorm.ErrRecordNotFound
	}
	cloned := *row
	return &cloned, nil
}
func (r *userTaskHandlerFakeSubTaskRepo) UpdateSubTaskCounts(ctx context.Context, subTaskID int64, totalCount, unfinishedCount, totalUserCount, unfinishedUserCount int) error {
	_ = ctx
	row := r.records[subTaskID]
	if row == nil {
		return gorm.ErrRecordNotFound
	}
	row.TotalCount = totalCount
	row.UnfinishedCount = unfinishedCount
	row.TotalUserCount = totalUserCount
	row.UnfinishedUserCount = unfinishedUserCount
	return nil
}
func (r *userTaskHandlerFakeSubTaskRepo) ListSubTasksByTaskID(ctx context.Context, taskID int64, page, pageSize int) ([]*datafillingdomain.DataFillingSubTask, int64, error) {
	_ = ctx
	_ = taskID
	_ = page
	_ = pageSize
	return nil, 0, nil
}
func (r *userTaskHandlerFakeSubTaskRepo) DeleteSubTasksByIDs(ctx context.Context, subTaskIDs []int64) error {
	_ = ctx
	_ = subTaskIDs
	return nil
}
func (r *userTaskHandlerFakeSubTaskRepo) ListSubTaskIDsByTaskIDs(ctx context.Context, taskIDs []int64) ([]int64, error) {
	_ = ctx
	_ = taskIDs
	return nil, nil
}
func (r *userTaskHandlerFakeSubTaskRepo) DecrementSubTaskUnfinishedCount(ctx context.Context, subTaskID int64) error {
	_ = ctx
	row := r.records[subTaskID]
	if row == nil {
		return gorm.ErrRecordNotFound
	}
	if row.UnfinishedCount > 0 {
		row.UnfinishedCount--
	}
	if row.UnfinishedUserCount > 0 {
		row.UnfinishedUserCount--
	}
	return nil
}

func (r *userTaskHandlerFakeSubInstanceRepo) BatchCreateSubInstances(ctx context.Context, instances []*datafillingdomain.DataFillingSubInstance) error {
	_ = ctx
	r.records = append(r.records, instances...)
	return nil
}
func (r *userTaskHandlerFakeSubInstanceRepo) DeleteSubInstancesByPID(ctx context.Context, pid int64) error {
	_ = ctx
	_ = pid
	return nil
}
func (r *userTaskHandlerFakeSubInstanceRepo) DeleteSubInstancesByPIDs(ctx context.Context, pids []int64) error {
	_ = ctx
	_ = pids
	return nil
}
func (r *userTaskHandlerFakeSubInstanceRepo) DeleteSubInstancesByTaskIDs(ctx context.Context, taskIDs []int64) error {
	_ = ctx
	_ = taskIDs
	return nil
}
func (r *userTaskHandlerFakeSubInstanceRepo) ListSubInstancesByPID(ctx context.Context, pid int64, statusFilter *int) ([]*datafillingdomain.DataFillingSubInstance, error) {
	_ = ctx
	_ = pid
	_ = statusFilter
	return nil, nil
}
func (r *userTaskHandlerFakeSubInstanceRepo) ListSubInstancesByUID(ctx context.Context, uid int64, page, pageSize int, req *datafillingdomain.UserTaskPageRequest) ([]*datafillingdomain.UserTaskVO, int64, error) {
	_ = ctx
	_ = page
	_ = pageSize
	rows := make([]*datafillingdomain.UserTaskVO, 0)
	for _, row := range r.records {
		if row.UID != uid {
			continue
		}
		if req != nil && req.Type != nil && row.Status != *req.Type {
			continue
		}
		rows = append(rows, &datafillingdomain.UserTaskVO{ID: row.ID, TaskID: row.TaskID, FormID: row.FormID, Status: row.Status})
	}
	return rows, int64(len(rows)), nil
}
func (r *userTaskHandlerFakeSubInstanceRepo) CountOpenSubInstancesByUID(ctx context.Context, uid int64) (int64, error) {
	_ = ctx
	var total int64
	for _, row := range r.records {
		if row.UID == uid && row.Status == datafillingdomain.SubInstanceStatusOpen {
			total++
		}
	}
	return total, nil
}
func (r *userTaskHandlerFakeSubInstanceRepo) GetSubInstanceByID(ctx context.Context, id int64) (*datafillingdomain.DataFillingSubInstance, error) {
	_ = ctx
	for _, row := range r.records {
		if row.ID == id {
			cloned := *row
			return &cloned, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}
func (r *userTaskHandlerFakeSubInstanceRepo) GetSubInstanceByPIDAndUID(ctx context.Context, pid, uid int64) ([]*datafillingdomain.DataFillingSubInstance, error) {
	_ = ctx
	rows := make([]*datafillingdomain.DataFillingSubInstance, 0)
	for _, row := range r.records {
		if row.PID == pid && row.UID == uid {
			cloned := *row
			rows = append(rows, &cloned)
		}
	}
	return rows, nil
}
func (r *userTaskHandlerFakeSubInstanceRepo) UpdateSubInstanceStatus(ctx context.Context, id int64, status int, finishTime int64) error {
	_ = ctx
	for _, row := range r.records {
		if row.ID == id {
			row.Status = status
			row.FinishTime = finishTime
			return nil
		}
	}
	return gorm.ErrRecordNotFound
}

func setupUserTaskHandler(t *testing.T) (*gin.Engine, *UserTaskHandler) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	formRepo := newFakeDataFillingRepo()
	formRepo.records[1] = &datafillingdomain.DataFillingForm{ID: 1, Name: "form", NodeType: datafillingdomain.NodeTypeForm, PhysicalTableName: "df_demo", DatasourceID: 8, Forms: "[]"}
	taskRepo := &userTaskHandlerFakeTaskRepo{records: map[int64]*datafillingdomain.DataFillingTask{1: {ID: 1, FormID: 1, FormExtSetting: "{}", FillType: 0}}}
	subTaskRepo := &userTaskHandlerFakeSubTaskRepo{records: map[int64]*datafillingdomain.DataFillingSubTask{10: {ID: 10, TaskID: 1, UnfinishedCount: 1, UnfinishedUserCount: 1}}}
	subInstanceRepo := &userTaskHandlerFakeSubInstanceRepo{records: []*datafillingdomain.DataFillingSubInstance{{ID: 100, TaskID: 1, PID: 10, UID: 9, FormID: 1, Status: datafillingdomain.SubInstanceStatusOpen}}}
	svc := service.NewDataFillingService(formRepo, &serviceTestDataFillingDatasourceServiceBridge{}, &serviceTestDataFillingDDLBridge{}, &serviceTestCommitLogRepoBridge{}, taskRepo, subTaskRepo, subInstanceRepo, nil)
	svc.SetDatasourceConnectionProvider(&serviceTestDatasourceConnProviderBridge{})
	h := NewUserTaskHandler(svc)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(9))
		c.Set("username", "tester")
		c.Next()
	})
	RegisterDataFillingUserTaskRoutes(r.Group(""), h, nil, nil)
	return r, h
}

func performUserTaskJSONRequest(t *testing.T, r *gin.Engine, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var payload []byte
	switch v := body.(type) {
	case nil:
		payload = nil
	default:
		var err error
		payload, err = json.Marshal(v)
		require.NoError(t, err)
	}
	request, err := http.NewRequest(method, path, bytes.NewReader(payload))
	require.NoError(t, err)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	r.ServeHTTP(recorder, request)
	return recorder
}

func TestUserTaskHandler_RoutesRegistered(t *testing.T) {
	r, _ := setupUserTaskHandler(t)
	routes := map[string]bool{}
	for _, route := range r.Routes() {
		routes[route.Method+" "+route.Path] = true
	}
	assert.True(t, routes["POST /data-filling/user-task/page/:goPage/:pageSize"])
	assert.True(t, routes["POST /data-filling/user-task/todo/count"])
	assert.True(t, routes["GET /data-filling/user-task/list/:id"])
	assert.True(t, routes["POST /data-filling/user-task/saveData/:id"])
	assert.True(t, routes["POST /data-filling/user-task/appendData/:id"])
	assert.True(t, routes["GET /data-filling/user-task/:taskInstanceId/deleteData/:id"])
}

func TestUserTaskHandler_Endpoints(t *testing.T) {
	r, _ := setupUserTaskHandler(t)

	w := performUserTaskJSONRequest(t, r, http.MethodPost, "/data-filling/user-task/page/1/10", map[string]any{})
	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeDatasetResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)

	w = performUserTaskJSONRequest(t, r, http.MethodPost, "/data-filling/user-task/todo/count", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	resp = decodeDatasetResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)

	w = performUserTaskJSONRequest(t, r, http.MethodGet, "/data-filling/user-task/list/10", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	resp = decodeDatasetResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)

	w = performUserTaskJSONRequest(t, r, http.MethodPost, "/data-filling/user-task/saveData/10", map[string]any{"data": []map[string]any{{"name": "alice"}}})
	assert.Equal(t, http.StatusOK, w.Code)
	resp = decodeDatasetResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)

	w = performUserTaskJSONRequest(t, r, http.MethodPost, "/data-filling/user-task/appendData/10", map[string]any{"data": []map[string]any{{"name": "bob"}}})
	assert.Equal(t, http.StatusOK, w.Code)
	resp = decodeDatasetResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)

	w = performUserTaskJSONRequest(t, r, http.MethodGet, "/data-filling/user-task/10/deleteData/row-1", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	resp = decodeDatasetResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)
}

func TestUserTaskHandler_BadRequestAndNotFound(t *testing.T) {
	r, _ := setupUserTaskHandler(t)

	w := performUserTaskJSONRequest(t, r, http.MethodPost, "/data-filling/user-task/page/0/10", map[string]any{})
	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeDatasetResp(t, w.Body.Bytes())
	assert.Equal(t, "10001", resp.Code)

	w = performUserTaskJSONRequest(t, r, http.MethodGet, "/data-filling/user-task/list/999", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	resp = decodeDatasetResp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
}
