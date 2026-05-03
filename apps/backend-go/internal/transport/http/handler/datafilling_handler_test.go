package handler

import (
	"bytes"
	"context"
	"encoding/json"
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

func TestDataFillingHandler_SaveAndGet(t *testing.T) {
	engine := gin.New()
	svc := service.NewDataFillingService(newFakeDataFillingRepo(), &serviceTestDataFillingDatasourceServiceBridge{}, &serviceTestDataFillingDDLBridge{})
	h := NewDataFillingHandler(svc)
	RegisterDataFillingRoutes(engine, h, nil, nil)

	body := []byte(`{"name":"folder-1","nodeType":"folder","pid":0}`)
	resp := performDataFillingRequest(t, engine, http.MethodPost, "/data-filling/save", body)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "000000", resp.Body.Code)

	var saved datafillingdomain.DataFillingForm
	require.NoError(t, json.Unmarshal(resp.Body.Data, &saved))
	assert.Equal(t, "folder-1", saved.Name)

	getResp := performDataFillingRequest(t, engine, http.MethodGet, "/data-filling/get/1", nil)
	assert.Equal(t, "000000", getResp.Body.Code)
	var got datafillingdomain.DataFillingForm
	require.NoError(t, json.Unmarshal(getResp.Body.Data, &got))
	assert.Equal(t, saved.ID, got.ID)
}

func TestDataFillingHandler_TreeBadRequest(t *testing.T) {
	engine := gin.New()
	svc := service.NewDataFillingService(newFakeDataFillingRepo(), &serviceTestDataFillingDatasourceServiceBridge{}, &serviceTestDataFillingDDLBridge{})
	h := NewDataFillingHandler(svc)
	RegisterDataFillingRoutes(engine, h, nil, nil)

	resp := performDataFillingRequest(t, engine, http.MethodPost, "/data-filling/rename", []byte(`{}`))
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "500000", resp.Body.Code)
}

func performDataFillingRequest(t *testing.T, engine *gin.Engine, method, path string, body []byte) *struct {
	StatusCode int
	Body       dataFillingHandlerResponse
} {
	t.Helper()
	request := httptest.NewRequest(method, path, bytes.NewReader(body))
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, request)
	var envelope dataFillingHandlerResponse
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &envelope))
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
	return nil, nil
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
