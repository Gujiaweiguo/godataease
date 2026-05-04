package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	datafillingdomain "dataease/backend/internal/domain/datafilling"
	datasourcedomain "dataease/backend/internal/domain/datasource"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type dataFillingHandlerResponse struct {
	Code string          `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

func TestDataFillingHandler_Save(t *testing.T) {
	repo := newFakeDataFillingRepo()
	svc := service.NewDataFillingService(repo, &serviceTestDataFillingDatasourceServiceBridge{}, &serviceTestDataFillingDDLBridge{}, &serviceTestCommitLogRepoBridge{}, nil, nil, nil, nil)
	svc.SetDatasourceConnectionProvider(&serviceTestDatasourceConnProviderBridge{})
	h := NewDataFillingHandler(svc)

	resp := performDataFillingHandlerCall(t, func(c *gin.Context) {
		c.Request = httptest.NewRequest(http.MethodPost, "/data-filling/save", bytes.NewReader([]byte(`{"name":"folder-1","nodeType":"folder","pid":0}`)))
		c.Request.Header.Set("Content-Type", "application/json")
		h.Save(c)
	})
	assert.Equal(t, "000000", resp.Body.Code)

	var saved datafillingdomain.DataFillingForm
	require.NoError(t, json.Unmarshal(resp.Body.Data, &saved))
	assert.Equal(t, "folder-1", saved.Name)
}

func TestDataFillingHandler_TreeBadRequest(t *testing.T) {
	svc := service.NewDataFillingService(newFakeDataFillingRepo(), &serviceTestDataFillingDatasourceServiceBridge{}, &serviceTestDataFillingDDLBridge{}, &serviceTestCommitLogRepoBridge{}, nil, nil, nil, nil)
	svc.SetDatasourceConnectionProvider(&serviceTestDatasourceConnProviderBridge{})
	h := NewDataFillingHandler(svc)
	resp := performDataFillingHandlerCall(t, func(c *gin.Context) {
		c.Request = httptest.NewRequest(http.MethodPost, "/data-filling/rename", bytes.NewReader([]byte(`{}`)))
		c.Request.Header.Set("Content-Type", "application/json")
		h.Rename(c)
	})
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "500000", resp.Body.Code)
}

func TestDataFillingHandler_RegisterRoutesIncludesDMLEndpoints(t *testing.T) {
	engine := gin.New()
	svc := service.NewDataFillingService(newFakeDataFillingRepo(), &serviceTestDataFillingDatasourceServiceBridge{}, &serviceTestDataFillingDDLBridge{}, &serviceTestCommitLogRepoBridge{}, nil, nil, nil, nil)
	h := NewDataFillingHandler(svc)
	RegisterDataFillingRoutes(engine, h, nil, nil)

	routes := map[string]bool{}
	for _, route := range engine.Routes() {
		routes[route.Method+" "+route.Path] = true
	}
	assert.True(t, routes["POST /data-filling/form/:formId/tableData"])
	assert.True(t, routes["POST /data-filling/form/:formId/rowData/save"])
	assert.True(t, routes["GET /data-filling/form/:formId/delete/:id"])
	assert.True(t, routes["POST /data-filling/form/:formId/batch-delete"])
	assert.True(t, routes["GET /data-filling/form/:formId/truncate"])
	assert.True(t, routes["POST /data-filling/form/:formId/listColumnData"])
	assert.True(t, routes["POST /data-filling/log/page/:goPage/:pageSize"])
	assert.True(t, routes["POST /data-filling/log/clear"])
	assert.True(t, routes["GET /data-filling/task/info/:taskId"])
	assert.True(t, routes["POST /data-filling/task/save"])
	assert.True(t, routes["POST /data-filling/task/executeNow"])
	assert.True(t, routes["POST /data-filling/form/:formId/task/page/:goPage/:pageSize"])
	assert.True(t, routes["GET /data-filling/form/:formId/task/:id/start"])
	assert.True(t, routes["GET /data-filling/form/:formId/task/:id/stop"])
	assert.True(t, routes["POST /data-filling/form/:formId/task/delete"])
	assert.True(t, routes["POST /data-filling/sub-task/page/:goPage/:pageSize"])
	assert.True(t, routes["POST /data-filling/form/:formId/sub-task/delete"])
	assert.True(t, routes["GET /data-filling/sub-task/:id/users/list/:type"])
}

func performDataFillingHandlerCall(t *testing.T, invoke func(c *gin.Context)) *struct {
	StatusCode int
	Body       dataFillingHandlerResponse
} {
	t.Helper()
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	invoke(c)
	var envelope dataFillingHandlerResponse
	body := recorder.Body.Bytes()
	if len(body) == 0 {
		body, _ = io.ReadAll(recorder.Body)
	}
	require.NoError(t, json.Unmarshal(body, &envelope))
	return &struct {
		StatusCode int
		Body       dataFillingHandlerResponse
	}{StatusCode: recorder.Code, Body: envelope}
}

func newFakeDataFillingRepo() *serviceTestDataFillingRepoBridge {
	return &serviceTestDataFillingRepoBridge{records: map[int64]*datafillingdomain.DataFillingForm{}, nextID: 1}
}

type serviceTestDataFillingRepoBridge struct {
	records map[int64]*datafillingdomain.DataFillingForm
	nextID  int64
}

func (r *serviceTestDataFillingRepoBridge) Create(ctx context.Context, form *datafillingdomain.DataFillingForm) error {
	_ = ctx
	cloned := *form
	if cloned.ID <= 0 {
		cloned.ID = r.nextID
		r.nextID++
	}
	r.records[cloned.ID] = &cloned
	form.ID = cloned.ID
	return nil
}

func (r *serviceTestDataFillingRepoBridge) GetByID(ctx context.Context, id int64) (*datafillingdomain.DataFillingForm, error) {
	_ = ctx
	row := r.records[id]
	if row == nil {
		return nil, gorm.ErrRecordNotFound
	}
	cloned := *row
	return &cloned, nil
}

func (r *serviceTestDataFillingRepoBridge) Update(ctx context.Context, form *datafillingdomain.DataFillingForm) error {
	_ = ctx
	r.records[form.ID] = form
	return nil
}

func (r *serviceTestDataFillingRepoBridge) DeleteByID(ctx context.Context, id int64) error {
	_ = ctx
	delete(r.records, id)
	return nil
}
func (r *serviceTestDataFillingRepoBridge) Rename(ctx context.Context, id int64, name string) error {
	_ = ctx
	if row := r.records[id]; row != nil {
		row.Name = name
	}
	return nil
}
func (r *serviceTestDataFillingRepoBridge) Move(ctx context.Context, id int64, pid int64) error {
	_ = ctx
	if row := r.records[id]; row != nil {
		row.PID = pid
	}
	return nil
}
func (r *serviceTestDataFillingRepoBridge) GetTree(ctx context.Context) ([]*datafillingdomain.DataFillingForm, error) {
	_ = ctx
	rows := make([]*datafillingdomain.DataFillingForm, 0, len(r.records))
	for _, row := range r.records {
		cloned := *row
		rows = append(rows, &cloned)
	}
	return rows, nil
}
func (r *serviceTestDataFillingRepoBridge) GetByPID(ctx context.Context, pid int64) ([]*datafillingdomain.DataFillingForm, error) {
	return r.GetChildren(ctx, pid)
}
func (r *serviceTestDataFillingRepoBridge) GetChildren(ctx context.Context, pid int64) ([]*datafillingdomain.DataFillingForm, error) {
	_ = ctx
	rows := make([]*datafillingdomain.DataFillingForm, 0)
	for _, row := range r.records {
		if row.PID == pid {
			cloned := *row
			rows = append(rows, &cloned)
		}
	}
	return rows, nil
}

type serviceTestDataFillingDatasourceServiceBridge struct{}

func (f *serviceTestDataFillingDatasourceServiceBridge) GetByID(id int64) (*datasourcedomain.CoreDatasource, error) {
	_ = id
	conf := `{"host":"127.0.0.1","port":3306,"dataBase":"demo","username":"u","password":"p"}`
	return &datasourcedomain.CoreDatasource{ID: id, Type: "mysql", Configuration: &conf}, nil
}
func (f *serviceTestDataFillingDatasourceServiceBridge) Tree(req *datasourcedomain.ListRequest) ([]*datasourcedomain.CoreDatasource, error) {
	_ = req
	return []*datasourcedomain.CoreDatasource{}, nil
}
func (f *serviceTestDataFillingDatasourceServiceBridge) GetTables(req *datasourcedomain.TableRequest) ([]datasourcedomain.TableInfo, error) {
	_ = req
	return []datasourcedomain.TableInfo{}, nil
}

type serviceTestDataFillingDDLBridge struct{}

func (f *serviceTestDataFillingDDLBridge) CreateTable(ctx context.Context, db *gorm.DB, tableName string, fields []datafillingdomain.ExtTableField) error {
	_ = ctx
	_ = db
	_ = tableName
	_ = fields
	return nil
}
func (f *serviceTestDataFillingDDLBridge) DropTable(ctx context.Context, db *gorm.DB, tableName string) error {
	_ = ctx
	_ = db
	_ = tableName
	return nil
}

func (f *serviceTestDataFillingDDLBridge) InsertRow(ctx context.Context, db *gorm.DB, tableName string, rowData map[string]interface{}) error {
	_ = ctx
	_ = db
	_ = tableName
	rowData["id"] = "generated-id"
	return nil
}

func (f *serviceTestDataFillingDDLBridge) UpdateRow(ctx context.Context, db *gorm.DB, tableName string, id string, rowData map[string]interface{}) error {
	_ = ctx
	_ = db
	_ = tableName
	_ = id
	_ = rowData
	return nil
}

func (f *serviceTestDataFillingDDLBridge) DeleteRows(ctx context.Context, db *gorm.DB, tableName string, ids []string) error {
	_ = ctx
	_ = db
	_ = tableName
	_ = ids
	return nil
}

func (f *serviceTestDataFillingDDLBridge) SearchRows(ctx context.Context, db *gorm.DB, tableName string, whereClause string, args []interface{}, limit, offset int64) ([]map[string]interface{}, error) {
	_ = ctx
	_ = db
	_ = tableName
	_ = whereClause
	_ = args
	_ = limit
	_ = offset
	return []map[string]interface{}{{"id": "row-1"}}, nil
}

func (f *serviceTestDataFillingDDLBridge) CountRows(ctx context.Context, db *gorm.DB, tableName string, whereClause string, args []interface{}) (int64, error) {
	_ = ctx
	_ = db
	_ = tableName
	_ = whereClause
	_ = args
	return 1, nil
}

func (f *serviceTestDataFillingDDLBridge) TruncateTable(ctx context.Context, db *gorm.DB, tableName string) error {
	_ = ctx
	_ = db
	_ = tableName
	return nil
}

func (f *serviceTestDataFillingDDLBridge) ListColumnData(ctx context.Context, db *gorm.DB, tableName string, columnName string) ([]string, error) {
	_ = ctx
	_ = db
	_ = tableName
	_ = columnName
	return []string{"a", "b"}, nil
}

func (f *serviceTestDataFillingDDLBridge) AddTableColumns(ctx context.Context, db *gorm.DB, tableName string, fields []datafillingdomain.ExtTableField) error {
	_ = ctx
	_ = db
	_ = tableName
	_ = fields
	return nil
}

func (f *serviceTestDataFillingDDLBridge) DropTableColumns(ctx context.Context, db *gorm.DB, tableName string, columnNames []string) error {
	_ = ctx
	_ = db
	_ = tableName
	_ = columnNames
	return nil
}

type serviceTestCommitLogRepoBridge struct{}

func (f *serviceTestCommitLogRepoBridge) Create(ctx context.Context, log *datafillingdomain.DfCommitLog) error {
	_ = ctx
	_ = log
	return nil
}

func (f *serviceTestCommitLogRepoBridge) ListByFormID(ctx context.Context, formID int64, page, pageSize int) ([]*datafillingdomain.DfCommitLog, int64, error) {
	_ = ctx
	_ = formID
	_ = page
	_ = pageSize
	return []*datafillingdomain.DfCommitLog{{ID: 1, FormID: 1}}, 1, nil
}

func (f *serviceTestCommitLogRepoBridge) DeleteByFormID(ctx context.Context, formID int64) error {
	_ = ctx
	_ = formID
	return nil
}

type serviceTestDatasourceConnProviderBridge struct{}

func (f *serviceTestDatasourceConnProviderBridge) GetDatasourceConnection(ctx context.Context, datasourceID int64) (*gorm.DB, error) {
	_ = ctx
	_ = datasourceID
	return &gorm.DB{}, nil
}
