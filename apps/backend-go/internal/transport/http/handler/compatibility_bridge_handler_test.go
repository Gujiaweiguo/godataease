package handler

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"dataease/backend/internal/domain/chart"
	"dataease/backend/internal/domain/dataset"
	"dataease/backend/internal/domain/datasource"
	exportdomain "dataease/backend/internal/domain/export"
	"dataease/backend/internal/domain/org"
	"dataease/backend/internal/domain/permission"
	"dataease/backend/internal/domain/user"
	"dataease/backend/internal/repository"
	"dataease/backend/internal/service"
	"dataease/backend/internal/transport/http/middleware"
	calcitev1 "dataease/backend/proto/calcite/v1"
	seatunnelv1 "dataease/backend/proto/seatunnel/v1"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type bridgeResp struct {
	Code string        `json:"code"`
	Data []interface{} `json:"data"`
}

type bridgeAnyResp struct {
	Code string                 `json:"code"`
	Data map[string]interface{} `json:"data"`
}

type bridgeCodeResp struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
}

type bridgeFieldListResp struct {
	Code string `json:"code"`
	Data struct {
		DimensionList []map[string]interface{} `json:"dimensionList"`
		QuotaList     []map[string]interface{} `json:"quotaList"`
	} `json:"data"`
}

type mockBridgeResourcePermRepo struct {
	hasPermission         bool
	governedPermissionIDs map[int64][]int64
}

type mockBridgeExportRepo struct {
	created   *exportdomain.ExportTask
	createErr error
}

func (m *mockBridgeExportRepo) Create(task *exportdomain.ExportTask) error {
	m.created = task
	return m.createErr
}

func (m *mockBridgeExportRepo) GetByID(string) (*exportdomain.ExportTask, error) {
	return nil, nil
}

func (m *mockBridgeExportRepo) List(int, int, string) ([]exportdomain.ExportTask, int64, error) {
	return nil, 0, nil
}

func (m *mockBridgeExportRepo) UpdateStatus(string, string) error {
	return nil
}

func (m *mockBridgeExportRepo) Delete(string) error {
	return nil
}

func (m *mockBridgeExportRepo) DeleteBatch([]string) error {
	return nil
}

func (m *mockBridgeExportRepo) DeleteAllByType(string) error {
	return nil
}

func (m *mockBridgeExportRepo) CountByStatus() (map[string]int64, error) {
	return map[string]int64{}, nil
}

func (m *mockBridgeResourcePermRepo) GetPermByID(permID int64) (*permission.SysPerm, error) {
	return &permission.SysPerm{PermID: permID, PermKey: "view"}, nil
}

func (m *mockBridgeResourcePermRepo) GetPermByKey(permKey string) (*permission.SysPerm, error) {
	return &permission.SysPerm{PermID: 1, PermKey: permKey}, nil
}

func (m *mockBridgeResourcePermRepo) ListPerms(permType string, page, size int) ([]*permission.SysPerm, int64, error) {
	return nil, 0, nil
}

func (m *mockBridgeResourcePermRepo) CreatePerm(perm *permission.SysPerm) error { return nil }
func (m *mockBridgeResourcePermRepo) UpdatePerm(perm *permission.SysPerm) error { return nil }
func (m *mockBridgeResourcePermRepo) DeletePerm(permID int64) error             { return nil }
func (m *mockBridgeResourcePermRepo) GetUserPerms(userID int64) ([]int64, error) {
	return nil, nil
}
func (m *mockBridgeResourcePermRepo) GetRolePerms(roleID int64) ([]int64, error) {
	return nil, nil
}
func (m *mockBridgeResourcePermRepo) GetUserRoleIDs(userID int64) ([]int64, error) {
	if m.hasPermission {
		return []int64{1}, nil
	}
	return nil, nil
}
func (m *mockBridgeResourcePermRepo) CheckUserPermission(userID, permID int64) (bool, error) {
	return m.hasPermission, nil
}
func (m *mockBridgeResourcePermRepo) CheckRolePermission(roleID, permID int64) (bool, error) {
	return m.hasPermission, nil
}
func (m *mockBridgeResourcePermRepo) GrantPermToUser(userID, permID int64, createBy string) error {
	return nil
}
func (m *mockBridgeResourcePermRepo) RevokePermFromUser(userID, permID int64) error { return nil }
func (m *mockBridgeResourcePermRepo) GrantPermToRole(roleID, permID int64) error    { return nil }
func (m *mockBridgeResourcePermRepo) RevokePermFromRole(roleID, permID int64) error { return nil }
func (m *mockBridgeResourcePermRepo) GetUserResources(userID int64, resourceType string) ([]*permission.UserResourcePermVO, error) {
	return []*permission.UserResourcePermVO{}, nil
}
func (m *mockBridgeResourcePermRepo) GetResourceUsers(resourceID int64, resourceType string) ([]*permission.ResourceUserPermVO, error) {
	return []*permission.ResourceUserPermVO{}, nil
}
func (m *mockBridgeResourcePermRepo) ApplyGroupPermissions(groupID, resourceID int64, resourceType string) error {
	return nil
}
func (m *mockBridgeResourcePermRepo) RegisterResource(resourceID int64, resourceName, resourceType string, parentID *int64) error {
	return nil
}
func (m *mockBridgeResourcePermRepo) ReplaceResourcePermissions(resourceID int64, resourceType string, permIDs []int64) error {
	return nil
}
func (m *mockBridgeResourcePermRepo) GetResourcePermissionIDs(resourceID int64, resourceType string) ([]int64, bool, error) {
	if m.governedPermissionIDs == nil {
		return nil, false, nil
	}
	permIDs, exists := m.governedPermissionIDs[resourceID]
	if !exists {
		return nil, false, nil
	}
	return permIDs, true, nil
}
func (m *mockBridgeResourcePermRepo) CheckPermissionConsistency() (*permission.PermissionConsistencyResult, error) {
	return &permission.PermissionConsistencyResult{Consistent: true}, nil
}

type mockSeatunnelSyncService struct {
	seatunnelv1.UnimplementedSyncServiceServer
	taskID string
}

type mockBridgeCalciteValidateServer struct {
	calcitev1.UnimplementedCalciteServiceServer
	validateCalls int32
}

func (m *mockSeatunnelSyncService) SubmitTask(context.Context, *seatunnelv1.SubmitTaskRequest) (*seatunnelv1.SubmitTaskResponse, error) {
	return &seatunnelv1.SubmitTaskResponse{TaskId: m.taskID}, nil
}

func (m *mockSeatunnelSyncService) GetTaskStatus(context.Context, *seatunnelv1.GetTaskStatusRequest) (*seatunnelv1.GetTaskStatusResponse, error) {
	return &seatunnelv1.GetTaskStatusResponse{Task: &seatunnelv1.SyncTask{Id: m.taskID, Status: "running", Progress: 50}}, nil
}

func (m *mockSeatunnelSyncService) CancelTask(context.Context, *seatunnelv1.CancelTaskRequest) (*seatunnelv1.CancelTaskResponse, error) {
	return &seatunnelv1.CancelTaskResponse{Success: true}, nil
}

func (m *mockBridgeCalciteValidateServer) ParseSQL(context.Context, *calcitev1.ParseSQLRequest) (*calcitev1.ParseSQLResponse, error) {
	return &calcitev1.ParseSQLResponse{NormalizedSql: "SELECT 1"}, nil
}

func (m *mockBridgeCalciteValidateServer) ValidateSQL(context.Context, *calcitev1.ValidateSQLRequest) (*calcitev1.ValidateSQLResponse, error) {
	atomic.AddInt32(&m.validateCalls, 1)
	return &calcitev1.ValidateSQLResponse{Valid: false, Message: "invalid sql"}, nil
}

func startMockSeatunnelServer(t *testing.T, taskID string) (string, func()) {
	t.Helper()

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}
	grpcServer := grpc.NewServer()
	seatunnelv1.RegisterSyncServiceServer(grpcServer, &mockSeatunnelSyncService{taskID: taskID})

	go func() {
		_ = grpcServer.Serve(lis)
	}()

	cleanup := func() {
		grpcServer.Stop()
		_ = lis.Close()
	}
	return lis.Addr().String(), cleanup
}

func startMockBridgeCalciteServer(t *testing.T, srv calcitev1.CalciteServiceServer) (string, func()) {
	t.Helper()

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}
	grpcServer := grpc.NewServer()
	calcitev1.RegisterCalciteServiceServer(grpcServer, srv)

	go func() {
		_ = grpcServer.Serve(lis)
	}()

	cleanup := func() {
		grpcServer.Stop()
		_ = lis.Close()
	}
	return lis.Addr().String(), cleanup
}

type fakeBridgeChartRepo struct {
	charts             map[int64]*chart.CoreChartView
	dsFields           map[int64][]*dataset.CoreDatasetTableField
	chartFields        map[int64][]*dataset.CoreDatasetTableField
	nextFieldID        int64
	fieldRegistry      map[int64]*dataset.CoreDatasetTableField
	chartDatasetGroups map[int64]int64
}

type bridgeStubPreviewExecutor struct {
	rows []map[string]interface{}
	err  error
}

func (s *bridgeStubPreviewExecutor) PreviewSQL(context.Context, string, int) ([]map[string]interface{}, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.rows, nil
}

func (s *bridgeStubPreviewExecutor) Close() error {
	return nil
}

type mockBridgeChartDatasetResolver struct {
	datasetGroupIDs map[int64]int64
	err             error
}

func (m *mockBridgeChartDatasetResolver) GetDatasetGroupIDByChartID(chartID int64) (int64, error) {
	if m.err != nil {
		return 0, m.err
	}
	datasetGroupID, ok := m.datasetGroupIDs[chartID]
	if !ok {
		return 0, gorm.ErrRecordNotFound
	}
	return datasetGroupID, nil
}

func (r *fakeBridgeChartRepo) GetByID(id int64) (*chart.CoreChartView, error) {
	if r.charts == nil {
		return nil, nil
	}
	v := r.charts[id]
	if v == nil {
		return nil, nil
	}
	clone := *v
	return &clone, nil
}

func (r *fakeBridgeChartRepo) Update(view *chart.CoreChartView) error {
	if r.charts == nil {
		r.charts = make(map[int64]*chart.CoreChartView)
	}
	clone := *view
	r.charts[view.ID] = &clone
	return nil
}

func (r *fakeBridgeChartRepo) QueryRows(chartID int64, limit int) ([]map[string]interface{}, int64, error) {
	return []map[string]interface{}{}, 0, nil
}

func (r *fakeBridgeChartRepo) QueryRowsWithFilter(chartID int64, selectColumns string, whereClause string, whereArgs []interface{}, limit int) ([]map[string]interface{}, int64, error) {
	return r.QueryRows(chartID, limit)
}

func (r *fakeBridgeChartRepo) GetDatasetGroupIDByChartID(chartID int64) (int64, error) {
	if r.chartDatasetGroups == nil {
		return 0, gorm.ErrRecordNotFound
	}
	datasetGroupID, ok := r.chartDatasetGroups[chartID]
	if !ok {
		return 0, gorm.ErrRecordNotFound
	}
	return datasetGroupID, nil
}

func (r *fakeBridgeChartRepo) ListDatasetFieldsByGroup(datasetGroupID int64) ([]*dataset.CoreDatasetTableField, error) {
	if r.dsFields == nil {
		return []*dataset.CoreDatasetTableField{}, nil
	}
	list := r.dsFields[datasetGroupID]
	result := make([]*dataset.CoreDatasetTableField, 0, len(list))
	for _, f := range list {
		cloned := *f
		result = append(result, &cloned)
	}
	return result, nil
}

func (r *fakeBridgeChartRepo) ListDatasetFieldsByChart(chartID int64) ([]*dataset.CoreDatasetTableField, error) {
	if r.chartFields == nil {
		return []*dataset.CoreDatasetTableField{}, nil
	}
	list := r.chartFields[chartID]
	result := make([]*dataset.CoreDatasetTableField, 0, len(list))
	for _, f := range list {
		cloned := *f
		result = append(result, &cloned)
	}
	return result, nil
}

func (r *fakeBridgeChartRepo) GetDatasetFieldByID(id int64) (*dataset.CoreDatasetTableField, error) {
	if r.fieldRegistry == nil {
		return nil, nil
	}
	f := r.fieldRegistry[id]
	if f == nil {
		return nil, nil
	}
	clone := *f
	return &clone, nil
}

func (r *fakeBridgeChartRepo) CountDatasetFieldName(datasetGroupID int64, name string) (int64, error) {
	if r.fieldRegistry == nil {
		return 0, nil
	}
	var count int64
	for _, f := range r.fieldRegistry {
		if f == nil || f.Name == nil {
			continue
		}
		if f.DatasetGroupID == datasetGroupID && strings.EqualFold(*f.Name, name) {
			count++
		}
	}
	return count, nil
}

func (r *fakeBridgeChartRepo) CreateDatasetField(field *dataset.CoreDatasetTableField) error {
	if r.nextFieldID <= 0 {
		r.nextFieldID = 1000
	}
	if field.ID <= 0 {
		field.ID = r.nextFieldID
		r.nextFieldID++
	}
	if r.fieldRegistry == nil {
		r.fieldRegistry = make(map[int64]*dataset.CoreDatasetTableField)
	}
	if r.chartFields == nil {
		r.chartFields = make(map[int64][]*dataset.CoreDatasetTableField)
	}
	clone := *field
	r.fieldRegistry[field.ID] = &clone
	if clone.ChartID != nil {
		r.chartFields[*clone.ChartID] = append(r.chartFields[*clone.ChartID], &clone)
	}
	return nil
}

func (r *fakeBridgeChartRepo) UpdateDatasetFieldNames(id int64, dataeaseName string, fieldShortName string) error {
	if r.fieldRegistry != nil {
		if f, ok := r.fieldRegistry[id]; ok && f != nil {
			f.DataeaseName = &dataeaseName
			f.FieldShortName = &fieldShortName
		}
	}
	if r.chartFields != nil {
		for _, items := range r.chartFields {
			for _, item := range items {
				if item == nil || item.ID != id {
					continue
				}
				item.DataeaseName = &dataeaseName
				item.FieldShortName = &fieldShortName
			}
		}
	}
	return nil
}

func (r *fakeBridgeChartRepo) DeleteDatasetField(id int64) error {
	if r.fieldRegistry != nil {
		delete(r.fieldRegistry, id)
	}
	if r.chartFields != nil {
		for chartID, fields := range r.chartFields {
			filtered := make([]*dataset.CoreDatasetTableField, 0, len(fields))
			for _, item := range fields {
				if item == nil || item.ID == id {
					continue
				}
				filtered = append(filtered, item)
			}
			r.chartFields[chartID] = filtered
		}
	}
	return nil
}

func (r *fakeBridgeChartRepo) DeleteDatasetFieldsByChart(chartID int64) error {
	if r.chartFields != nil {
		for _, item := range r.chartFields[chartID] {
			if item == nil {
				continue
			}
			if r.fieldRegistry != nil {
				delete(r.fieldRegistry, item.ID)
			}
		}
		delete(r.chartFields, chartID)
	}
	return nil
}

func TestChartDataGetFieldDataInvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterCompatibilityBridgeRoutes(r, nil, nil, nil, nil, &ChartHandler{}, nil)

	req := httptest.NewRequest("POST", "/chartData/getFieldData/not-number/xAxis", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	resp := bridgeResp{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if resp.Code != "500000" {
		t.Fatalf("expected code 500000, got %s", resp.Code)
	}
}

func TestChartDataGetFieldDataFallbackEmptyWhenDatasetHandlerNil(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterCompatibilityBridgeRoutes(r, nil, nil, nil, nil, &ChartHandler{}, nil)

	req := httptest.NewRequest("POST", "/chartData/getFieldData/100/xAxis", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	resp := bridgeResp{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if resp.Code != "000000" {
		t.Fatalf("expected code 000000, got %s", resp.Code)
	}
	if len(resp.Data) != 0 {
		t.Fatalf("expected empty data, got %#v", resp.Data)
	}
}

func TestChartDataGetDrillFieldDataFallbackEmptyWhenDatasetHandlerNil(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterCompatibilityBridgeRoutes(r, nil, nil, nil, nil, &ChartHandler{}, nil)

	req := httptest.NewRequest("POST", "/chartData/getDrillFieldData/100", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	resp := bridgeResp{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if resp.Code != "000000" {
		t.Fatalf("expected code 000000, got %s", resp.Code)
	}
	if len(resp.Data) != 0 {
		t.Fatalf("expected empty data, got %#v", resp.Data)
	}
}

func TestChartSaveRouteUpdatesCoreFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &fakeBridgeChartRepo{charts: map[int64]*chart.CoreChartView{}}
	title := "old"
	repo.charts[101] = &chart.CoreChartView{ID: 101, Title: &title}
	chartHandler := NewChartHandler(service.NewChartService(repo))

	r := gin.New()
	RegisterCompatibilityBridgeRoutes(r, nil, nil, nil, nil, chartHandler, nil)

	reqBody := `{"id":101,"title":"new-title","resultMode":"all"}`
	req := httptest.NewRequest("POST", "/chart/save", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	resp := bridgeAnyResp{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if resp.Code != "000000" {
		t.Fatalf("expected code 000000, got %s", resp.Code)
	}
	if resp.Data["title"] != "new-title" {
		t.Fatalf("expected title updated, got %#v", resp.Data["title"])
	}
}

func TestChartListByDQRouteReturnsDimensionAndQuota(t *testing.T) {
	gin.SetMode(gin.TestMode)
	checked := true
	groupD := "d"
	typeD := "VARCHAR"
	nameD := "region"
	originD := "region"
	dataeaseD := "region"
	deTypeD := 0

	repo := &fakeBridgeChartRepo{
		charts:   map[int64]*chart.CoreChartView{},
		dsFields: map[int64][]*dataset.CoreDatasetTableField{},
	}
	repo.dsFields[11] = []*dataset.CoreDatasetTableField{{
		ID:             1,
		DatasetGroupID: 11,
		Name:           &nameD,
		OriginName:     &originD,
		DataeaseName:   &dataeaseD,
		GroupType:      &groupD,
		Type:           &typeD,
		DeType:         &deTypeD,
		Checked:        &checked,
	}}
	chartHandler := NewChartHandler(service.NewChartService(repo))

	r := gin.New()
	RegisterCompatibilityBridgeRoutes(r, nil, nil, nil, nil, chartHandler, nil)

	req := httptest.NewRequest("POST", "/chart/listByDQ/11/9", strings.NewReader(`{"type":"bar"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	resp := bridgeFieldListResp{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if resp.Code != "000000" {
		t.Fatalf("expected code 000000, got %s", resp.Code)
	}
	if len(resp.Data.DimensionList) != 1 {
		t.Fatalf("expected 1 dimension field, got %d", len(resp.Data.DimensionList))
	}
	if len(resp.Data.QuotaList) == 0 {
		t.Fatal("expected quota list contains pseudo count field")
	}
}

func TestApiAliasChartSaveAndListByDQ(t *testing.T) {
	gin.SetMode(gin.TestMode)
	checked := true
	groupD := "d"
	typeD := "VARCHAR"
	nameD := "city"
	originD := "city"
	dataeaseD := "city"
	deTypeD := 0
	title := "origin"

	repo := &fakeBridgeChartRepo{
		charts: map[int64]*chart.CoreChartView{201: {ID: 201, Title: &title}},
		dsFields: map[int64][]*dataset.CoreDatasetTableField{21: {{
			ID:             2,
			DatasetGroupID: 21,
			Name:           &nameD,
			OriginName:     &originD,
			DataeaseName:   &dataeaseD,
			GroupType:      &groupD,
			Type:           &typeD,
			DeType:         &deTypeD,
			Checked:        &checked,
		}}},
	}
	chartHandler := NewChartHandler(service.NewChartService(repo))

	r := gin.New()
	api := r.Group("/api")
	RegisterCompatibilityBridgeRoutes(api, nil, nil, nil, nil, chartHandler, nil)

	saveReq := httptest.NewRequest("POST", "/api/chart/save", strings.NewReader(`{"id":201,"title":"alias-title"}`))
	saveReq.Header.Set("Content-Type", "application/json")
	saveW := httptest.NewRecorder()
	r.ServeHTTP(saveW, saveReq)
	if saveW.Code != 200 {
		t.Fatalf("expected status 200, got %d", saveW.Code)
	}
	saveResp := bridgeAnyResp{}
	if err := json.Unmarshal(saveW.Body.Bytes(), &saveResp); err != nil {
		t.Fatalf("unmarshal save response failed: %v", err)
	}
	if saveResp.Code != "000000" {
		t.Fatalf("expected save code 000000, got %s", saveResp.Code)
	}
	if saveResp.Data["title"] != "alias-title" {
		t.Fatalf("expected alias save title updated, got %#v", saveResp.Data["title"])
	}

	listReq := httptest.NewRequest("POST", "/api/chart/listByDQ/21/201", strings.NewReader(`{"type":"bar"}`))
	listReq.Header.Set("Content-Type", "application/json")
	listW := httptest.NewRecorder()
	r.ServeHTTP(listW, listReq)
	if listW.Code != 200 {
		t.Fatalf("expected status 200, got %d", listW.Code)
	}
	listResp := bridgeFieldListResp{}
	if err := json.Unmarshal(listW.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("unmarshal list response failed: %v", err)
	}
	if listResp.Code != "000000" {
		t.Fatalf("expected list code 000000, got %s", listResp.Code)
	}
	if len(listResp.Data.DimensionList) != 1 {
		t.Fatalf("expected alias list dimension size 1, got %d", len(listResp.Data.DimensionList))
	}
}

func TestDatasetFieldListWithPermissions_FiltersDisabledAndMarksMasked(t *testing.T) {
	gin.SetMode(gin.TestMode)
	checked := true
	groupD := "d"
	groupQ := "q"
	varcharType := "VARCHAR"
	intType := "INT"
	region := "region"
	amount := "amount"
	deTypeD := 0
	deTypeQ := 2

	repo := &fakeBridgeChartRepo{
		dsFields: map[int64][]*dataset.CoreDatasetTableField{11: {
			{ID: 1, DatasetGroupID: 11, Name: &region, OriginName: &region, DataeaseName: &region, GroupType: &groupD, Type: &varcharType, DeType: &deTypeD, Checked: &checked},
			{ID: 2, DatasetGroupID: 11, Name: &amount, OriginName: &amount, DataeaseName: &amount, GroupType: &groupQ, Type: &intType, DeType: &deTypeQ, Checked: &checked},
		}},
	}
	chartService := service.NewChartService(repo)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	if err = db.AutoMigrate(&permission.DataPermColumn{}); err != nil {
		t.Fatalf("migrate permission columns failed: %v", err)
	}
	columnRepo := repository.NewColumnPermissionRepository(db)
	if err = columnRepo.Create(&permission.DataPermColumn{DatasetID: 11, DatasetGroupID: 11, FieldName: "amount", PermType: permission.PermTypeDisable, Status: 1}); err != nil {
		t.Fatalf("create disable permission failed: %v", err)
	}
	if err = columnRepo.Create(&permission.DataPermColumn{DatasetID: 11, DatasetGroupID: 11, FieldName: "region", PermType: permission.PermTypeMask, Status: 1}); err != nil {
		t.Fatalf("create mask permission failed: %v", err)
	}
	chartService.SetColumnPermissionService(service.NewColumnPermissionService(columnRepo))
	chartHandler := NewChartHandler(chartService)

	r := gin.New()
	api := r.Group("/api")
	api.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(42))
		c.Next()
	})
	RegisterCompatibilityBridgeRoutes(api, nil, nil, nil, nil, chartHandler, nil)

	req := httptest.NewRequest("GET", "/api/datasetField/listWithPermissions/11", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	var resp struct {
		Code string             `json:"code"`
		Data []chart.ChartField `json:"data"`
	}
	if err = json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if resp.Code != "000000" {
		t.Fatalf("expected code 000000, got %s", resp.Code)
	}
	if len(resp.Data) != 2 {
		t.Fatalf("expected region plus count field, got %#v", resp.Data)
	}
	if resp.Data[0].OriginName != "region" || !resp.Data[0].Desensitized {
		t.Fatalf("expected region field to be present and desensitized, got %#v", resp.Data[0])
	}
	if resp.Data[1].ID != -1 {
		t.Fatalf("expected count pseudo field to remain, got %#v", resp.Data[1])
	}
}

func TestChartCopyAndDeleteFieldRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	name := "sales"
	origin := "sales"
	dataeaseName := "sales"
	groupType := "q"
	typeName := "DECIMAL"
	deType := 3
	checked := true

	repo := &fakeBridgeChartRepo{
		charts:        map[int64]*chart.CoreChartView{},
		chartFields:   map[int64][]*dataset.CoreDatasetTableField{},
		fieldRegistry: map[int64]*dataset.CoreDatasetTableField{},
		nextFieldID:   2000,
	}
	repo.fieldRegistry[10] = &dataset.CoreDatasetTableField{
		ID:             10,
		DatasetGroupID: 30,
		Name:           &name,
		OriginName:     &origin,
		DataeaseName:   &dataeaseName,
		GroupType:      &groupType,
		Type:           &typeName,
		DeType:         &deType,
		Checked:        &checked,
	}
	chartHandler := NewChartHandler(service.NewChartService(repo))

	r := gin.New()
	RegisterCompatibilityBridgeRoutes(r, nil, nil, nil, nil, chartHandler, nil)

	copyReq := httptest.NewRequest("POST", "/chart/copyField/10/300", strings.NewReader("{}"))
	copyReq.Header.Set("Content-Type", "application/json")
	copyW := httptest.NewRecorder()
	r.ServeHTTP(copyW, copyReq)
	if copyW.Code != 200 {
		t.Fatalf("expected status 200, got %d", copyW.Code)
	}
	copyResp := bridgeCodeResp{}
	if err := json.Unmarshal(copyW.Body.Bytes(), &copyResp); err != nil {
		t.Fatalf("unmarshal copy response failed: %v", err)
	}
	if copyResp.Code != "000000" {
		t.Fatalf("expected copy code 000000, got %s", copyResp.Code)
	}
	if len(repo.chartFields[300]) != 1 {
		t.Fatalf("expected 1 copied field for chart 300, got %d", len(repo.chartFields[300]))
	}
	copiedID := repo.chartFields[300][0].ID

	delReq := httptest.NewRequest("POST", "/chart/deleteField/"+strconv.FormatInt(copiedID, 10), strings.NewReader("{}"))
	delReq.Header.Set("Content-Type", "application/json")
	delW := httptest.NewRecorder()
	r.ServeHTTP(delW, delReq)
	if delW.Code != 200 {
		t.Fatalf("expected status 200, got %d", delW.Code)
	}
	delResp := bridgeCodeResp{}
	if err := json.Unmarshal(delW.Body.Bytes(), &delResp); err != nil {
		t.Fatalf("unmarshal delete response failed: %v", err)
	}
	if delResp.Code != "000000" {
		t.Fatalf("expected delete code 000000, got %s", delResp.Code)
	}
	if len(repo.chartFields[300]) != 0 {
		t.Fatalf("expected copied field deleted, remaining %d", len(repo.chartFields[300]))
	}
}

func TestApiAliasChartCopyAndDeleteFieldByChart(t *testing.T) {
	gin.SetMode(gin.TestMode)
	name := "profit"
	origin := "profit"
	dataeaseName := "profit"
	groupType := "q"
	typeName := "DECIMAL"
	deType := 3
	checked := true

	repo := &fakeBridgeChartRepo{
		charts:        map[int64]*chart.CoreChartView{},
		chartFields:   map[int64][]*dataset.CoreDatasetTableField{},
		fieldRegistry: map[int64]*dataset.CoreDatasetTableField{},
		nextFieldID:   3000,
	}
	repo.fieldRegistry[11] = &dataset.CoreDatasetTableField{
		ID:             11,
		DatasetGroupID: 31,
		Name:           &name,
		OriginName:     &origin,
		DataeaseName:   &dataeaseName,
		GroupType:      &groupType,
		Type:           &typeName,
		DeType:         &deType,
		Checked:        &checked,
	}
	chartHandler := NewChartHandler(service.NewChartService(repo))

	r := gin.New()
	api := r.Group("/api")
	RegisterCompatibilityBridgeRoutes(api, nil, nil, nil, nil, chartHandler, nil)

	copyReq := httptest.NewRequest("POST", "/api/chart/copyField/11/400", strings.NewReader("{}"))
	copyReq.Header.Set("Content-Type", "application/json")
	copyW := httptest.NewRecorder()
	r.ServeHTTP(copyW, copyReq)
	if copyW.Code != 200 {
		t.Fatalf("expected status 200, got %d", copyW.Code)
	}
	copyResp := bridgeCodeResp{}
	if err := json.Unmarshal(copyW.Body.Bytes(), &copyResp); err != nil {
		t.Fatalf("unmarshal copy response failed: %v", err)
	}
	if copyResp.Code != "000000" {
		t.Fatalf("expected copy code 000000, got %s", copyResp.Code)
	}
	if len(repo.chartFields[400]) != 1 {
		t.Fatalf("expected alias copied field count 1, got %d", len(repo.chartFields[400]))
	}

	delByChartReq := httptest.NewRequest("POST", "/api/chart/deleteFieldByChart/400", strings.NewReader("{}"))
	delByChartReq.Header.Set("Content-Type", "application/json")
	delByChartW := httptest.NewRecorder()
	r.ServeHTTP(delByChartW, delByChartReq)
	if delByChartW.Code != 200 {
		t.Fatalf("expected status 200, got %d", delByChartW.Code)
	}
	delByChartResp := bridgeCodeResp{}
	if err := json.Unmarshal(delByChartW.Body.Bytes(), &delByChartResp); err != nil {
		t.Fatalf("unmarshal delete-by-chart response failed: %v", err)
	}
	if delByChartResp.Code != "000000" {
		t.Fatalf("expected delete-by-chart code 000000, got %s", delByChartResp.Code)
	}
	if _, ok := repo.chartFields[400]; ok {
		t.Fatal("expected chart fields deleted by chart id")
	}
}

func TestDatasetFieldAliasRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	name := "profit"
	origin := "profit"
	dataeaseName := "profit"
	groupType := "q"
	typeName := "DECIMAL"
	deType := 3
	checked := true

	repo := &fakeBridgeChartRepo{
		charts:        map[int64]*chart.CoreChartView{},
		dsFields:      map[int64][]*dataset.CoreDatasetTableField{},
		chartFields:   map[int64][]*dataset.CoreDatasetTableField{},
		fieldRegistry: map[int64]*dataset.CoreDatasetTableField{},
		nextFieldID:   5000,
	}
	datasetDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, datasetDB.AutoMigrate(&dataset.CoreDatasetGroup{}, &dataset.CoreDatasetTable{}, &dataset.CoreDatasetTableField{}))
	datasetRepo := repository.NewDatasetRepository(datasetDB)
	datasetService := service.NewDatasetService(datasetRepo)
	datasetHandler := NewDatasetHandler(datasetService)
	field := &dataset.CoreDatasetTableField{
		ID:             21,
		DatasetGroupID: 41,
		ChartID:        int64PtrBridge(601),
		Name:           &name,
		OriginName:     &origin,
		DataeaseName:   &dataeaseName,
		GroupType:      &groupType,
		Type:           &typeName,
		DeType:         &deType,
		Checked:        &checked,
	}
	repo.fieldRegistry[21] = field
	repo.dsFields[41] = []*dataset.CoreDatasetTableField{field}
	require.NoError(t, datasetDB.Create(&dataset.CoreDatasetTableField{
		ID:             21,
		DatasetGroupID: 41,
		ChartID:        int64PtrBridge(601),
		Name:           &name,
		OriginName:     &origin,
		DataeaseName:   &dataeaseName,
		GroupType:      &groupType,
		Type:           &typeName,
		DeType:         &deType,
		Checked:        &checked,
	}).Error)
	chartHandler := NewChartHandler(service.NewChartService(repo))

	r := gin.New()
	RegisterCompatibilityBridgeRoutes(r, nil, nil, nil, datasetHandler, chartHandler, nil)

	listReq := httptest.NewRequest("POST", "/datasetField/listByDatasetGroup/41", strings.NewReader("{}"))
	listReq.Header.Set("Content-Type", "application/json")
	listW := httptest.NewRecorder()
	r.ServeHTTP(listW, listReq)
	if listW.Code != 200 {
		t.Fatalf("expected status 200, got %d", listW.Code)
	}
	var listResp struct {
		Code string             `json:"code"`
		Data []chart.ChartField `json:"data"`
	}
	if err := json.Unmarshal(listW.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("unmarshal dataset field list failed: %v", err)
	}
	if listResp.Code != "000000" {
		t.Fatalf("expected list code 000000, got %s", listResp.Code)
	}
	if len(listResp.Data) != 2 {
		t.Fatalf("expected flattened field count 2, got %d", len(listResp.Data))
	}

	delReq := httptest.NewRequest("POST", "/datasetField/delete/21", strings.NewReader("{}"))
	delReq.Header.Set("Content-Type", "application/json")
	delW := httptest.NewRecorder()
	r.ServeHTTP(delW, delReq)
	if delW.Code != 200 {
		t.Fatalf("expected delete status 200, got %d", delW.Code)
	}
	delResp := bridgeCodeResp{}
	if err := json.Unmarshal(delW.Body.Bytes(), &delResp); err != nil {
		t.Fatalf("unmarshal dataset field delete failed: %v", err)
	}
	if delResp.Code != "000000" {
		t.Fatalf("expected delete code 000000, got %s", delResp.Code)
	}

	delByChartReq := httptest.NewRequest("POST", "/datasetField/deleteByChartId/601", strings.NewReader("{}"))
	delByChartReq.Header.Set("Content-Type", "application/json")
	delByChartW := httptest.NewRecorder()
	r.ServeHTTP(delByChartW, delByChartReq)
	if delByChartW.Code != 200 {
		t.Fatalf("expected deleteByChart status 200, got %d", delByChartW.Code)
	}
	delByChartResp := bridgeCodeResp{}
	if err := json.Unmarshal(delByChartW.Body.Bytes(), &delByChartResp); err != nil {
		t.Fatalf("unmarshal dataset field deleteByChart failed: %v", err)
	}
	if delByChartResp.Code != "000000" {
		t.Fatalf("expected deleteByChart code 000000, got %s", delByChartResp.Code)
	}
}

func int64PtrBridge(v int64) *int64 { return &v }

func strPtrBridge(v string) *string { return &v }

func setupStage3DatasetFieldRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()

	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&dataset.CoreDatasetGroup{}, &dataset.CoreDatasetTable{}, &dataset.CoreDatasetTableField{}, &user.SysUser{}))

	datasetRepo := repository.NewDatasetRepository(db)
	datasetService := service.NewDatasetService(datasetRepo)
	datasetService.SetUserRepository(repository.NewUserRepository(db))
	datasetHandler := NewDatasetHandler(datasetService)

	chartRepo := &fakeBridgeChartRepo{
		charts:        map[int64]*chart.CoreChartView{},
		dsFields:      map[int64][]*dataset.CoreDatasetTableField{},
		chartFields:   map[int64][]*dataset.CoreDatasetTableField{},
		fieldRegistry: map[int64]*dataset.CoreDatasetTableField{},
		nextFieldID:   9000,
	}
	chartHandler := NewChartHandler(service.NewChartService(chartRepo))

	r := gin.New()
	RegisterCompatibilityBridgeRoutes(r, nil, nil, nil, datasetHandler, chartHandler, nil)
	return r, db
}

func seedBridgeUser(t *testing.T, db *gorm.DB, id int64, username string, nickname string) {
	t.Helper()
	require.NoError(t, db.Create(&user.SysUser{UserID: id, Username: username, NickName: nickname, Status: user.StatusEnabled, DelFlag: user.DelFlagNormal}).Error)
}

func seedBridgeDatasetGroup(t *testing.T, db *gorm.DB, group *dataset.CoreDatasetGroup) {
	t.Helper()
	require.NoError(t, db.Create(group).Error)
}

func seedBridgeDatasetField(t *testing.T, db *gorm.DB, field *dataset.CoreDatasetTableField) {
	t.Helper()
	require.NoError(t, db.Create(field).Error)
}

func TestOldPathDatasourceList(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterCompatibilityBridgeRoutes(r, nil, nil, nil, nil, nil, nil)

	req := httptest.NewRequest("POST", "/datasource/list", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 404 {
		t.Logf("Route /datasource/list registered, status: %d", w.Code)
	}
}

func TestOldPathDatasetTree(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterCompatibilityBridgeRoutes(r, nil, nil, nil, nil, nil, nil)

	req := httptest.NewRequest("POST", "/datasetTree/tree", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 404 {
		t.Logf("Route /datasetTree/tree registered, status: %d", w.Code)
	}
}

func TestApiAliasDatasourceList(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	api := r.Group("/api")
	RegisterCompatibilityBridgeRoutes(api, nil, nil, nil, nil, nil, nil)

	req := httptest.NewRequest("POST", "/api/datasource/list", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 404 {
		t.Logf("Route /api/datasource/list registered, status: %d", w.Code)
	}
}

func TestApiAliasDatasetTree(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	api := r.Group("/api")
	RegisterCompatibilityBridgeRoutes(api, nil, nil, nil, nil, nil, nil)

	req := httptest.NewRequest("POST", "/api/datasetTree/tree", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 404 {
		t.Logf("Route /api/datasetTree/tree registered, status: %d", w.Code)
	}
}

func TestPaginationResponseFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterCompatibilityBridgeRoutes(r, nil, nil, nil, nil, nil, nil)

	req := httptest.NewRequest("POST", "/chart/listByDQ/1/1", strings.NewReader(`{"type":"bar"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code == 200 {
		var resp map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err == nil {
			if code, ok := resp["code"]; ok {
				if _, isString := code.(string); !isString {
					t.Errorf("expected code field to be string, got %T", code)
				}
			}
		}
	}
}

func TestErrorResponseFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &fakeBridgeChartRepo{charts: map[int64]*chart.CoreChartView{}}
	chartHandler := NewChartHandler(service.NewChartService(repo))

	r := gin.New()
	RegisterCompatibilityBridgeRoutes(r, nil, nil, nil, nil, chartHandler, nil)

	req := httptest.NewRequest("POST", "/chartData/getFieldData/invalid/xAxis", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	resp := bridgeCodeResp{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}

	if resp.Code != "500000" {
		t.Fatalf("expected error code '500000', got '%s'", resp.Code)
	}
}

func TestOldPathChartSaveWithNilHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterCompatibilityBridgeRoutes(r, nil, nil, nil, nil, nil, nil)

	req := httptest.NewRequest("POST", "/chart/save", strings.NewReader(`{"id":1}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 404 {
		t.Logf("Route /chart/save registered, status: %d", w.Code)
	}
}

func TestApiAliasChartDataGetData(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	api := r.Group("/api")
	RegisterCompatibilityBridgeRoutes(api, nil, nil, nil, nil, nil, nil)

	req := httptest.NewRequest("POST", "/api/chartData/getData", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 404 {
		t.Logf("Route /api/chartData/getData registered, status: %d", w.Code)
	}
}

func TestDatasetTreeExportDatasetRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("queues export task when dataEaseBi is false", func(t *testing.T) {
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		require.NoError(t, err)
		require.NoError(t, db.AutoMigrate(&dataset.CoreDatasetGroup{}, &dataset.CoreDatasetTable{}, &dataset.CoreDatasetTableField{}))

		repo := repository.NewDatasetRepository(db)
		datasetSvc := service.NewDatasetService(repo)
		exportRepo := &mockBridgeExportRepo{}
		datasetSvc.SetExportRepository(exportRepo)
		datasetHandler := NewDatasetHandler(datasetSvc)

		rootPID := int64(0)
		nodeType := dataset.NodeTypeDataset
		require.NoError(t, db.Create(&dataset.CoreDatasetGroup{ID: 501, Name: "Bridge Dataset", PID: &rootPID, NodeType: &nodeType}).Error)

		r := gin.New()
		RegisterCompatibilityBridgeRoutes(r, nil, nil, nil, datasetHandler, nil, nil)

		req := httptest.NewRequest("POST", "/datasetTree/exportDataset", strings.NewReader(`{"id":501,"viewName":"Bridge Export"}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		require.Equal(t, 200, w.Code)
		resp := bridgeAnyResp{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "000000", resp.Code)
		assert.Equal(t, "PENDING", resp.Data["status"])
		assert.Equal(t, "dataset", resp.Data["exportFromType"])
		require.NotNil(t, exportRepo.created)
		assert.Equal(t, int64(501), exportRepo.created.ExportFrom)
		assert.Equal(t, int64(0), exportRepo.created.UserID)
	})

	t.Run("resolves compat dataset ids before creating export task", func(t *testing.T) {
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		require.NoError(t, err)
		require.NoError(t, db.AutoMigrate(&dataset.CoreDatasetGroup{}, &dataset.CoreDatasetTable{}, &dataset.CoreDatasetTableField{}))

		repo := repository.NewDatasetRepository(db)
		datasetSvc := service.NewDatasetService(repo)
		exportRepo := &mockBridgeExportRepo{}
		datasetSvc.SetExportRepository(exportRepo)
		datasetHandler := NewDatasetHandler(datasetSvc)

		rootPID := int64(0)
		nodeType := dataset.NodeTypeDataset
		require.NoError(t, db.Create(&dataset.CoreDatasetGroup{ID: 101, Name: "Compat Target", PID: &rootPID, NodeType: &nodeType}).Error)

		r := gin.New()
		RegisterCompatibilityBridgeRoutes(r, nil, nil, nil, datasetHandler, nil, nil)

		req := httptest.NewRequest("POST", "/datasetTree/exportDataset", strings.NewReader(`{"id":200,"viewName":"Compat Target Export"}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		require.Equal(t, 200, w.Code)
		resp := bridgeAnyResp{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "000000", resp.Code)
		assert.Equal(t, float64(101), resp.Data["exportFrom"])
		require.NotNil(t, exportRepo.created)
		assert.Equal(t, int64(101), exportRepo.created.ExportFrom)
	})

	t.Run("keeps inline download behavior when dataEaseBi is true", func(t *testing.T) {
		datasetHandler := NewDatasetHandler(service.NewDatasetService(nil))
		chartHandler := NewChartHandler(nil)

		r := gin.New()
		RegisterCompatibilityBridgeRoutes(r, nil, nil, nil, datasetHandler, chartHandler, nil)

		reqBody := `{"id":1,"viewName":"Inline Export","dataEaseBi":true,"header":["col1"],"details":[["v1"]]}`
		req := httptest.NewRequest("POST", "/datasetTree/exportDataset", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		require.Equal(t, 200, w.Code)
		assert.Contains(t, w.Header().Get("Content-Type"), "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		assert.NotEmpty(t, w.Body.Bytes())
	})
}

func TestDatasetTreeBarInfoRoute_ReturnsAuditFields(t *testing.T) {
	r, db := setupStage3DatasetFieldRouter(t)
	seedBridgeUser(t, db, 101, "alice_login", "Alice")
	seedBridgeUser(t, db, 102, "bob", "")

	nodeType := dataset.NodeTypeDataset
	seedBridgeDatasetGroup(t, db, &dataset.CoreDatasetGroup{
		ID:             1001,
		Name:           "orders",
		NodeType:       &nodeType,
		CreateBy:       "101",
		CreateTime:     1711111111,
		UpdateBy:       "102",
		LastUpdateTime: 1712222222,
	})

	req := httptest.NewRequest(http.MethodGet, "/datasetTree/barInfo/1001", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp struct {
		Code string          `json:"code"`
		Data dataset.BarInfo `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "000000", resp.Code)
	assert.Equal(t, int64(1001), resp.Data.ID)
	assert.Equal(t, "orders", resp.Data.Name)
	assert.Equal(t, dataset.NodeTypeDataset, resp.Data.NodeType)
	assert.Equal(t, "101", resp.Data.CreateBy)
	assert.Equal(t, "102", resp.Data.UpdateBy)
	assert.Equal(t, int64(1711111111), resp.Data.CreateTime)
	assert.Equal(t, int64(1712222222), resp.Data.LastUpdateTime)
	assert.Equal(t, "Alice", resp.Data.Creator)
	assert.Equal(t, "bob", resp.Data.Updater)
	assert.False(t, resp.Data.IsCross)
}

func TestDatasetTreeBarInfoRoute_InvalidID(t *testing.T) {
	r, _ := setupStage3DatasetFieldRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/datasetTree/barInfo/abc", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	resp := bridgeCodeResp{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "500000", resp.Code)
	assert.Contains(t, resp.Msg, "Invalid dataset ID")
}

func TestDatasetTreeBarInfoRoute_NotFound(t *testing.T) {
	r, _ := setupStage3DatasetFieldRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/datasetTree/barInfo/99999", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	resp := bridgeCodeResp{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "500000", resp.Code)
	assert.Contains(t, resp.Msg, "Failed to get dataset")
}

func TestDatasetFieldSaveRoute_CreateAndValidation(t *testing.T) {
	r, db := setupStage3DatasetFieldRouter(t)
	seedBridgeDatasetGroup(t, db, &dataset.CoreDatasetGroup{ID: 2001, Name: "dataset-save"})

	t.Run("create field", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/datasetField/save", strings.NewReader(`{"name":"order_amount","datasetGroupId":2001,"type":"int","originName":"amount"}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		var resp struct {
			Code string                        `json:"code"`
			Data dataset.CoreDatasetTableField `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "000000", resp.Code)
		require.NotNil(t, resp.Data.Name)
		require.NotNil(t, resp.Data.Type)
		assert.Equal(t, "order_amount", *resp.Data.Name)
		assert.Equal(t, "int", *resp.Data.Type)
		assert.Equal(t, int64(2001), resp.Data.DatasetGroupID)

		var stored []dataset.CoreDatasetTableField
		require.NoError(t, db.Where("dataset_group_id = ?", 2001).Find(&stored).Error)
		require.Len(t, stored, 1)
		require.NotNil(t, stored[0].Name)
		assert.Equal(t, "order_amount", *stored[0].Name)
	})

	t.Run("validation errors", func(t *testing.T) {
		tests := []struct {
			name       string
			body       string
			wantSubstr string
		}{
			{name: "missing name", body: `{"datasetGroupId":2001,"type":"int"}`, wantSubstr: "field name is required"},
			{name: "missing dataset group", body: `{"name":"field1","type":"int"}`, wantSubstr: "datasetGroupId is required"},
			{name: "missing type", body: `{"name":"field1","datasetGroupId":2001}`, wantSubstr: "field type is required"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				req := httptest.NewRequest(http.MethodPost, "/datasetField/save", strings.NewReader(tt.body))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)

				require.Equal(t, http.StatusOK, w.Code)
				resp := bridgeCodeResp{}
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "500000", resp.Code)
				assert.Contains(t, resp.Msg, tt.wantSubstr)
			})
		}
	})
}

func TestDatasetFieldGetFunctionRoute_ReturnsFunctionCategories(t *testing.T) {
	r, _ := setupStage3DatasetFieldRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/datasetField/getFunction", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp struct {
		Code string                     `json:"code"`
		Data []service.FunctionCategory `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "000000", resp.Code)
	require.Len(t, resp.Data, 5)
	assert.Equal(t, "聚合函数", resp.Data[0].Name)
	assert.Equal(t, "日期函数", resp.Data[1].Name)
	assert.Equal(t, "字符串函数", resp.Data[2].Name)
	assert.Equal(t, "数学函数", resp.Data[3].Name)
	assert.Equal(t, "条件函数", resp.Data[4].Name)
	for _, category := range resp.Data {
		assert.NotEmpty(t, category.Functions)
	}
}

func TestDatasetFieldListByDsIdsRoute_ReturnsMatchedFields(t *testing.T) {
	r, db := setupStage3DatasetFieldRouter(t)
	dsAID := int64(11)
	dsBID := int64(22)
	seedBridgeDatasetField(t, db, &dataset.CoreDatasetTableField{ID: 4001, DatasourceID: &dsAID, DatasetGroupID: 1001, Name: strPtrBridge("field_a1"), Type: strPtrBridge("string")})
	seedBridgeDatasetField(t, db, &dataset.CoreDatasetTableField{ID: 4002, DatasourceID: &dsAID, DatasetGroupID: 1001, Name: strPtrBridge("field_a2"), Type: strPtrBridge("int")})
	seedBridgeDatasetField(t, db, &dataset.CoreDatasetTableField{ID: 4003, DatasourceID: &dsBID, DatasetGroupID: 1002, Name: strPtrBridge("field_b1"), Type: strPtrBridge("string")})

	t.Run("matched datasource ids", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/datasetField/listByDsIds", strings.NewReader(`{"dsIds":[11]}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		var resp struct {
			Code string                          `json:"code"`
			Data []dataset.CoreDatasetTableField `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "000000", resp.Code)
		names := make([]string, 0, len(resp.Data))
		for _, field := range resp.Data {
			if field.Name != nil {
				names = append(names, *field.Name)
			}
		}
		assert.ElementsMatch(t, []string{"field_a1", "field_a2"}, names)
		assert.NotContains(t, names, "field_b1")
	})

	t.Run("empty input returns empty list", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/datasetField/listByDsIds", strings.NewReader(`{"dsIds":[]}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		var resp struct {
			Code string                          `json:"code"`
			Data []dataset.CoreDatasetTableField `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "000000", resp.Code)
		assert.Len(t, resp.Data, 0)
	})
}

func TestDatasetFieldMultFieldValuesForPermissionsRoute_ReturnsEnumValues(t *testing.T) {
	r, db := setupStage3DatasetFieldRouter(t)
	rootPID := int64(0)
	nodeType := dataset.NodeTypeDataset
	tableName := "bridge_enum_values"
	originName := "status"
	fieldName := "status"
	deType := 0

	seedBridgeDatasetGroup(t, db, &dataset.CoreDatasetGroup{ID: 2001, Name: "Enum Dataset", PID: &rootPID, NodeType: &nodeType})
	require.NoError(t, db.Create(&dataset.CoreDatasetTable{ID: 2101, DatasetGroupID: 2001, PhysicalTable: &tableName}).Error)
	seedBridgeDatasetField(t, db, &dataset.CoreDatasetTableField{ID: 2201, DatasetTableID: int64PtrBridge(2101), DatasetGroupID: 2001, OriginName: &originName, Name: &fieldName, DeType: &deType})
	require.NoError(t, db.Exec("CREATE TABLE bridge_enum_values (status TEXT)").Error)
	require.NoError(t, db.Exec("INSERT INTO bridge_enum_values (status) VALUES ('A'), ('B'), ('A')").Error)

	req := httptest.NewRequest(http.MethodPost, "/datasetField/multFieldValuesForPermissions", strings.NewReader(`{"fieldIds":[2201]}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp struct {
		Code string   `json:"code"`
		Data []string `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "000000", resp.Code)
	assert.Equal(t, []string{"A", "B"}, resp.Data)
}

func TestDatasetFieldMultFieldValuesForPermissionsRoute_NormalizesFieldIDs(t *testing.T) {
	r, db := setupStage3DatasetFieldRouter(t)
	rootPID := int64(0)
	nodeType := dataset.NodeTypeDataset
	tableName := "bridge_enum_values_normalized"
	originName := "region"
	fieldName := "region"
	deType := 0

	seedBridgeDatasetGroup(t, db, &dataset.CoreDatasetGroup{ID: 2002, Name: "Enum Dataset 2", PID: &rootPID, NodeType: &nodeType})
	require.NoError(t, db.Create(&dataset.CoreDatasetTable{ID: 2102, DatasetGroupID: 2002, PhysicalTable: &tableName}).Error)
	seedBridgeDatasetField(t, db, &dataset.CoreDatasetTableField{ID: 2202, DatasetTableID: int64PtrBridge(2102), DatasetGroupID: 2002, OriginName: &originName, Name: &fieldName, DeType: &deType})
	require.NoError(t, db.Exec("CREATE TABLE bridge_enum_values_normalized (region TEXT)").Error)
	require.NoError(t, db.Exec("INSERT INTO bridge_enum_values_normalized (region) VALUES ('North'), ('South')").Error)

	req := httptest.NewRequest(http.MethodPost, "/datasetField/multFieldValuesForPermissions", strings.NewReader(`{"fieldIds":[0,2202,2202,-1,999999]}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp struct {
		Code string   `json:"code"`
		Data []string `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "000000", resp.Code)
	assert.Equal(t, []string{"North", "South"}, resp.Data)
}

func TestDatasetFieldMultFieldValuesForPermissionsRoute_InvalidJSON(t *testing.T) {
	r, _ := setupStage3DatasetFieldRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/datasetField/multFieldValuesForPermissions", strings.NewReader(`{"fieldIds":[`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	resp := bridgeCodeResp{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "500000", resp.Code)
	assert.Contains(t, resp.Msg, "Invalid request")
}

func TestDatasourceSyncRouteReturnsErrorWhenSeatunnelUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	dsHandler := NewDatasourceHandler(service.NewDatasourceService(nil))
	RegisterCompatibilityBridgeRoutes(r, nil, nil, dsHandler, nil, nil, nil)

	req := httptest.NewRequest("POST", "/datasource/syncApiDs", strings.NewReader(`{"datasourceId":"1"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	resp := bridgeCodeResp{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if resp.Code != "500000" {
		t.Fatalf("expected code 500000 when seatunnel unavailable, got %s", resp.Code)
	}
}

func TestDatasourceSyncRouteReturnsSuccessWhenSeatunnelAvailable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	addr, cleanup := startMockSeatunnelServer(t, "task-123")
	defer cleanup()

	r := gin.New()
	dsService := service.NewDatasourceService(nil)
	dsService.SetSeatunnelConfig(addr, 2*time.Second, 0)
	dsHandler := NewDatasourceHandler(dsService)
	RegisterCompatibilityBridgeRoutes(r, nil, nil, dsHandler, nil, nil, nil)

	req := httptest.NewRequest("POST", "/datasource/syncApiDs", strings.NewReader(`{"datasourceId":"1","name":"sync-job"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	resp := bridgeAnyResp{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if resp.Code != "000000" {
		t.Fatalf("expected code 000000, got %s", resp.Code)
	}
	if resp.Data["taskId"] != "task-123" {
		t.Fatalf("expected taskId task-123, got %#v", resp.Data["taskId"])
	}
	if resp.Data["status"] != "running" {
		t.Fatalf("expected status running, got %#v", resp.Data["status"])
	}
	if resp.Data["syncType"] != "datasource" {
		t.Fatalf("expected syncType datasource, got %#v", resp.Data["syncType"])
	}
	if resp.Data["datasourceId"] != float64(1) {
		t.Fatalf("expected datasourceId 1, got %#v", resp.Data["datasourceId"])
	}
}

func TestDatasourceSyncTableRouteReturnsSuccessWhenSeatunnelAvailable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	addr, cleanup := startMockSeatunnelServer(t, "task-456")
	defer cleanup()

	r := gin.New()
	dsService := service.NewDatasourceService(nil)
	dsService.SetSeatunnelConfig(addr, 2*time.Second, 0)
	dsHandler := NewDatasourceHandler(dsService)
	RegisterCompatibilityBridgeRoutes(r, nil, nil, dsHandler, nil, nil, nil)

	req := httptest.NewRequest("POST", "/datasource/syncApiTable", strings.NewReader(`{"datasourceId":"2","name":"sync-table-job","tableName":"orders"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	resp := bridgeAnyResp{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if resp.Code != "000000" {
		t.Fatalf("expected code 000000, got %s", resp.Code)
	}
	if resp.Data["taskId"] != "task-456" {
		t.Fatalf("expected taskId task-456, got %#v", resp.Data["taskId"])
	}
	if resp.Data["status"] != "running" {
		t.Fatalf("expected status running, got %#v", resp.Data["status"])
	}
	if resp.Data["syncType"] != "table" {
		t.Fatalf("expected syncType table, got %#v", resp.Data["syncType"])
	}
	if resp.Data["datasourceId"] != float64(2) {
		t.Fatalf("expected datasourceId 2, got %#v", resp.Data["datasourceId"])
	}
}

func TestDatasourceListSyncRecordReturnsErrorWithoutRepository(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	dsHandler := NewDatasourceHandler(service.NewDatasourceService(nil))
	RegisterCompatibilityBridgeRoutes(r, nil, nil, dsHandler, nil, nil, nil)

	req := httptest.NewRequest("POST", "/datasource/listSyncRecord/1/1/10", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	resp := bridgeCodeResp{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if resp.Code != "500000" {
		t.Fatalf("expected code 500000 when repository unavailable, got %s", resp.Code)
	}
}

func TestDatasourceDeleteRoutes_PostAndGetShareSemantics(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&datasource.CoreDatasource{}))

	repo := repository.NewDatasourceRepository(db)
	dsService := service.NewDatasourceService(repo)
	dsHandler := NewDatasourceHandler(dsService)

	rootPID := int64(0)
	delFlag := 0
	require.NoError(t, db.Create(&datasource.CoreDatasource{ID: 601, PID: &rootPID, Name: "Folder", Type: datasource.TypeFolder, DelFlag: &delFlag}).Error)
	require.NoError(t, db.Create(&datasource.CoreDatasource{ID: 602, PID: int64PtrBridge(601), Name: "DS-POST", Type: "API", DelFlag: &delFlag}).Error)
	require.NoError(t, db.Create(&datasource.CoreDatasource{ID: 603, PID: int64PtrBridge(601), Name: "DS-GET", Type: "API", DelFlag: &delFlag}).Error)

	r := gin.New()
	RegisterCompatibilityBridgeRoutes(r, nil, nil, dsHandler, nil, nil, nil)

	postReq := httptest.NewRequest("POST", "/datasource/delete/602", strings.NewReader("{}"))
	postReq.Header.Set("Content-Type", "application/json")
	postW := httptest.NewRecorder()
	r.ServeHTTP(postW, postReq)
	require.Equal(t, 200, postW.Code)
	postResp := bridgeCodeResp{}
	require.NoError(t, json.Unmarshal(postW.Body.Bytes(), &postResp))
	assert.Equal(t, "000000", postResp.Code)

	getReq := httptest.NewRequest("GET", "/datasource/delete/603", nil)
	getW := httptest.NewRecorder()
	r.ServeHTTP(getW, getReq)
	require.Equal(t, 200, getW.Code)
	getResp := bridgeCodeResp{}
	require.NoError(t, json.Unmarshal(getW.Body.Bytes(), &getResp))
	assert.Equal(t, "000000", getResp.Code)

	var postDeleted datasource.CoreDatasource
	require.NoError(t, db.First(&postDeleted, 602).Error)
	require.NotNil(t, postDeleted.DelFlag)
	assert.Equal(t, 1, *postDeleted.DelFlag)

	var getDeleted datasource.CoreDatasource
	require.NoError(t, db.First(&getDeleted, 603).Error)
	require.NotNil(t, getDeleted.DelFlag)
	assert.Equal(t, 1, *getDeleted.DelFlag)

	invalidPostReq := httptest.NewRequest("POST", "/datasource/delete/not-a-number", strings.NewReader("{}"))
	invalidPostReq.Header.Set("Content-Type", "application/json")
	invalidPostW := httptest.NewRecorder()
	r.ServeHTTP(invalidPostW, invalidPostReq)
	require.Equal(t, 200, invalidPostW.Code)
	invalidPostResp := bridgeCodeResp{}
	require.NoError(t, json.Unmarshal(invalidPostW.Body.Bytes(), &invalidPostResp))
	assert.Equal(t, "500000", invalidPostResp.Code)

	invalidGetReq := httptest.NewRequest("GET", "/datasource/delete/not-a-number", nil)
	invalidGetW := httptest.NewRecorder()
	r.ServeHTTP(invalidGetW, invalidGetReq)
	require.Equal(t, 200, invalidGetW.Code)
	invalidGetResp := bridgeCodeResp{}
	require.NoError(t, json.Unmarshal(invalidGetW.Body.Bytes(), &invalidGetResp))
	assert.Equal(t, "500000", invalidGetResp.Code)
}

func TestDatasetPreviewSQLRouteUsesCalciteValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	calciteMock := &mockBridgeCalciteValidateServer{}
	addr, cleanup := startMockBridgeCalciteServer(t, calciteMock)
	defer cleanup()

	datasetService := service.NewDatasetService(nil)
	datasetService.SetCalciteConfig(addr, 2*time.Second, 0)
	datasetHandler := NewDatasetHandler(datasetService)

	r := gin.New()
	RegisterCompatibilityBridgeRoutes(r, nil, nil, nil, datasetHandler, nil, nil)

	req := httptest.NewRequest("POST", "/datasetData/previewSql", strings.NewReader(`{"sql":"SELECT 1"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if resp["code"] != "500000" {
		t.Fatalf("expected code 500000 when calcite validation fails, got %#v", resp["code"])
	}
	if atomic.LoadInt32(&calciteMock.validateCalls) == 0 {
		t.Fatal("expected calcite validate to be called")
	}
}

func TestApiAliasDatasetPreviewSQLRouteUsesCalciteValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	calciteMock := &mockBridgeCalciteValidateServer{}
	addr, cleanup := startMockBridgeCalciteServer(t, calciteMock)
	defer cleanup()

	datasetService := service.NewDatasetService(nil)
	datasetService.SetCalciteConfig(addr, 2*time.Second, 0)
	datasetHandler := NewDatasetHandler(datasetService)

	r := gin.New()
	api := r.Group("/api")
	RegisterCompatibilityBridgeRoutes(api, nil, nil, nil, datasetHandler, nil, nil)

	req := httptest.NewRequest("POST", "/api/datasetData/previewSql", strings.NewReader(`{"sql":"SELECT 1"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if resp["code"] != "500000" {
		t.Fatalf("expected code 500000 when calcite validation fails, got %#v", resp["code"])
	}
	if atomic.LoadInt32(&calciteMock.validateCalls) == 0 {
		t.Fatal("expected calcite validate to be called")
	}
}

func TestDatasetPreviewSQLRouteReturnsExplicitUnsupportedForExternalDatasource(t *testing.T) {
	gin.SetMode(gin.TestMode)
	datasetService := service.NewDatasetService(nil)
	datasetHandler := NewDatasetHandler(datasetService)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(9))
		c.Next()
	})
	RegisterCompatibilityBridgeRoutes(r, nil, nil, nil, datasetHandler, nil, nil)

	sql := base64.StdEncoding.EncodeToString([]byte("SELECT 1"))
	req := httptest.NewRequest("POST", "/datasetData/previewSql", strings.NewReader(`{"sql":"`+sql+`","datasourceId":99,"sqlVariableDetails":"[{\"variableName\":\"region\"}]"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, 200, w.Code)
	resp := bridgeCodeResp{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "500000", resp.Code)
	assert.Contains(t, resp.Msg, "external datasource SQL preview is not supported yet")
}

func TestApiAliasDatasetPreviewSQLRouteReturnsExplicitUnsupportedForExternalDatasource(t *testing.T) {
	gin.SetMode(gin.TestMode)
	datasetService := service.NewDatasetService(nil)
	datasetHandler := NewDatasetHandler(datasetService)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(9))
		c.Next()
	})
	api := r.Group("/api")
	RegisterCompatibilityBridgeRoutes(api, nil, nil, nil, datasetHandler, nil, nil)

	sql := base64.StdEncoding.EncodeToString([]byte("SELECT 1"))
	req := httptest.NewRequest("POST", "/api/datasetData/previewSql", strings.NewReader(`{"sql":"`+sql+`","datasourceId":99}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, 200, w.Code)
	resp := bridgeCodeResp{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "500000", resp.Code)
	assert.Contains(t, resp.Msg, "external datasource SQL preview is not supported yet")
}

func TestDatasetPreviewSQLRouteRoutesMySQLDatasourcePreview(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&dataset.CoreDatasetGroup{}, &dataset.CoreDatasetTable{}, &dataset.CoreDatasetTableField{}, &datasource.CoreDatasource{}))

	datasetService := service.NewDatasetService(repository.NewDatasetRepository(db))
	datasetService.SetDatasourceRepository(repository.NewDatasourceRepository(db))
	configBytes := base64.StdEncoding.EncodeToString([]byte(`{"host":"mysql.local","port":3306,"dataBase":"analytics","username":"root","password":"secret"}`))
	require.NoError(t, db.Create(&datasource.CoreDatasource{ID: 66, Name: "mysql-ds", Type: "mysql", Configuration: &configBytes}).Error)
	datasetService.SetPreviewExecutorFactory(func(ds *datasource.CoreDatasource, cfg *datasource.ConnectionConfig) (service.PreviewExecutor, error) {
		assert.Equal(t, int64(66), ds.ID)
		assert.Equal(t, "root", cfg.Username)
		return &bridgeStubPreviewExecutor{rows: []map[string]interface{}{{"name": "alice"}}}, nil
	})
	datasetHandler := NewDatasetHandler(datasetService)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(9))
		c.Next()
	})
	RegisterCompatibilityBridgeRoutes(r, nil, nil, nil, datasetHandler, nil, nil)

	sql := base64.StdEncoding.EncodeToString([]byte("SELECT name FROM orders"))
	req := httptest.NewRequest("POST", "/datasetData/previewSql", strings.NewReader(`{"sql":"`+sql+`","datasourceId":66,"sqlVariableDetails":"[{\"variableName\":\"region\"}]"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, 200, w.Code)
	var resp struct {
		Code string `json:"code"`
		Data struct {
			Data dataset.SQLPreviewData `json:"data"`
			SQL  string                 `json:"sql"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "000000", resp.Code)
	require.Len(t, resp.Data.Data.Data, 1)
	assert.Equal(t, "alice", resp.Data.Data.Data[0]["name"])
	assert.NotEmpty(t, resp.Data.SQL)
}

func TestDatasetPreviewSQLRouteReturnsPermissionDeniedForDirectPreview(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&dataset.CoreDatasetGroup{}, &dataset.CoreDatasetTable{}, &dataset.CoreDatasetTableField{}, &datasource.CoreDatasource{}))

	datasetService := service.NewDatasetService(repository.NewDatasetRepository(db))
	datasetService.SetDatasourceRepository(repository.NewDatasourceRepository(db))
	datasetService.SetResourcePermissionService(service.NewResourcePermissionService(&mockBridgeResourcePermRepo{hasPermission: false}, nil))
	configBytes := base64.StdEncoding.EncodeToString([]byte(`{"host":"mysql.local","port":3306,"dataBase":"analytics","username":"root","password":"secret"}`))
	require.NoError(t, db.Create(&datasource.CoreDatasource{ID: 67, Name: "mysql-ds", Type: "mysql", Configuration: &configBytes}).Error)
	datasetHandler := NewDatasetHandler(datasetService)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(9))
		c.Next()
	})
	RegisterCompatibilityBridgeRoutes(r, nil, nil, nil, datasetHandler, nil, nil)

	sql := base64.StdEncoding.EncodeToString([]byte("SELECT 1"))
	req := httptest.NewRequest("POST", "/datasetData/previewSql", strings.NewReader(`{"sql":"`+sql+`","datasourceId":67,"sqlVariableDetails":"[{\"variableName\":\"tenant\"}]"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, 200, w.Code)
	resp := bridgeCodeResp{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "500000", resp.Code)
	assert.Contains(t, resp.Msg, "insufficient datasource permissions")
}

func TestDatasetPreviewSQLRouteReturnsTimeoutAndTooLargeErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setup := func(t *testing.T, execErr error, datasourceID int64) *gin.Engine {
		t.Helper()
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		require.NoError(t, err)
		require.NoError(t, db.AutoMigrate(&dataset.CoreDatasetGroup{}, &dataset.CoreDatasetTable{}, &dataset.CoreDatasetTableField{}, &datasource.CoreDatasource{}))

		datasetService := service.NewDatasetService(repository.NewDatasetRepository(db))
		datasetService.SetDatasourceRepository(repository.NewDatasourceRepository(db))
		configBytes := base64.StdEncoding.EncodeToString([]byte(`{"host":"mysql.local","port":3306,"dataBase":"analytics","username":"root","password":"secret"}`))
		require.NoError(t, db.Create(&datasource.CoreDatasource{ID: datasourceID, Name: "mysql-ds", Type: "mysql", Configuration: &configBytes}).Error)
		datasetService.SetPreviewExecutorFactory(func(ds *datasource.CoreDatasource, cfg *datasource.ConnectionConfig) (service.PreviewExecutor, error) {
			return &bridgeStubPreviewExecutor{err: execErr}, nil
		})
		datasetHandler := NewDatasetHandler(datasetService)

		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set("user_id", uint64(9))
			c.Next()
		})
		RegisterCompatibilityBridgeRoutes(r, nil, nil, nil, datasetHandler, nil, nil)
		return r
	}

	t.Run("timeout", func(t *testing.T) {
		r := setup(t, service.ErrPreviewSQLTimeout, 68)
		sql := base64.StdEncoding.EncodeToString([]byte("SELECT 1"))
		req := httptest.NewRequest("POST", "/datasetData/previewSql", strings.NewReader(`{"sql":"`+sql+`","datasourceId":68,"sqlVariableDetails":"[{\"variableName\":\"timeout\"}]"}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		resp := bridgeCodeResp{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "500000", resp.Code)
		assert.Contains(t, resp.Msg, "preview query timed out")
	})

	t.Run("result too large", func(t *testing.T) {
		r := setup(t, service.ErrPreviewSQLResultTooLarge, 69)
		sql := base64.StdEncoding.EncodeToString([]byte("SELECT 1"))
		req := httptest.NewRequest("POST", "/datasetData/previewSql", strings.NewReader(`{"sql":"`+sql+`","datasourceId":69,"sqlVariableDetails":"[{\"variableName\":\"size\"}]"}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		resp := bridgeCodeResp{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "500000", resp.Code)
		assert.Contains(t, resp.Msg, "preview result is too large")
	})
}

func TestDatasetPreviewSQLRouteAcceptsSQLVariableDetailsForLocalPreview(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&dataset.CoreDatasetGroup{}, &dataset.CoreDatasetTable{}, &dataset.CoreDatasetTableField{}, &datasource.CoreDatasource{}))
	require.NoError(t, db.Exec("CREATE TABLE preview_sql_local_variables (name TEXT)").Error)
	require.NoError(t, db.Exec("INSERT INTO preview_sql_local_variables (name) VALUES ('alice')").Error)

	datasetService := service.NewDatasetService(repository.NewDatasetRepository(db))
	datasetHandler := NewDatasetHandler(datasetService)

	r := gin.New()
	RegisterCompatibilityBridgeRoutes(r, nil, nil, nil, datasetHandler, nil, nil)

	sql := base64.StdEncoding.EncodeToString([]byte("SELECT name FROM preview_sql_local_variables"))
	req := httptest.NewRequest("POST", "/datasetData/previewSql", strings.NewReader(`{"sql":"`+sql+`","sqlVariableDetails":"[{\"variableName\":\"city\",\"defaultValue\":\"shanghai\"}]"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, 200, w.Code)
	var resp struct {
		Code string `json:"code"`
		Data struct {
			Data dataset.SQLPreviewData `json:"data"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "000000", resp.Code)
	require.Len(t, resp.Data.Data.Data, 1)
	assert.Equal(t, "alice", resp.Data.Data.Data[0]["name"])
}

func TestCompatibilityBridge_DatasetDetailWithPerm_401_Unauthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockBridgeResourcePermRepo{hasPermission: true}
	adminChecker := middleware.NewDefaultAdminChecker([]int64{})
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	permMiddleware := middleware.NewPermissionMiddleware(resourcePermSvc, exportPermSvc, adminChecker)

	r := gin.New()
	api := r.Group("/api")
	RegisterCompatibilityBridgeRoutes(api, nil, nil, nil, &DatasetHandler{}, nil, permMiddleware)

	req := httptest.NewRequest("POST", "/api/datasetTree/detailWithPerm", strings.NewReader(`{"datasetGroupId":123}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 401 {
		t.Fatalf("expected 401 for unauthenticated detailWithPerm, got %d", w.Code)
	}
}

func TestCompatibilityBridge_DatasetDetailWithPerm_403_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockBridgeResourcePermRepo{hasPermission: false}
	adminChecker := middleware.NewDefaultAdminChecker([]int64{})
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	permMiddleware := middleware.NewPermissionMiddleware(resourcePermSvc, exportPermSvc, adminChecker)

	r := gin.New()
	api := r.Group("/api")
	api.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(100))
		c.Next()
	})
	RegisterCompatibilityBridgeRoutes(api, nil, nil, nil, &DatasetHandler{}, nil, permMiddleware)

	req := httptest.NewRequest("POST", "/api/datasetTree/detailWithPerm", strings.NewReader(`{"datasetGroupId":123}`))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 403 {
		t.Fatalf("expected 403 for forbidden detailWithPerm, got %d", w.Code)
	}
}

func TestCompatibilityBridge_DatasetDetailWithPerm_400_MissingDatasetContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockBridgeResourcePermRepo{hasPermission: false}
	adminChecker := middleware.NewDefaultAdminChecker([]int64{1})
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	permMiddleware := middleware.NewPermissionMiddleware(resourcePermSvc, exportPermSvc, adminChecker)

	r := gin.New()
	api := r.Group("/api")
	api.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	})
	RegisterCompatibilityBridgeRoutes(api, nil, nil, nil, &DatasetHandler{}, nil, permMiddleware)

	req := httptest.NewRequest("POST", "/api/datasetTree/detailWithPerm", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected HTTP 200 with business error for missing row-permission dataset context, got %d", w.Code)
	}

	var resp bridgeCodeResp
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if resp.Code != "10001" {
		t.Fatalf("expected business code 10001 for missing row-permission dataset context, got %s", resp.Code)
	}
}

func TestCompatibilityBridge_DatasetDetailWithPerm_UsesPermissionAwareFieldsAndPreview(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := setupPermissionAwareDatasetDetailRouter(t)

	req := httptest.NewRequest("POST", "/api/datasetTree/detailWithPerm", strings.NewReader(`[11]`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	assertPermissionAwareDatasetDetailResponse(t, w.Body.Bytes())
}

func TestCompatibilityBridge_ChartDataGetData_401_Unauthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockBridgeResourcePermRepo{hasPermission: true}
	adminChecker := middleware.NewDefaultAdminChecker([]int64{})
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	permMiddleware := middleware.NewPermissionMiddleware(resourcePermSvc, exportPermSvc, adminChecker)
	permMiddleware.SetChartDatasetResolver(&mockBridgeChartDatasetResolver{datasetGroupIDs: map[int64]int64{101: 11}})

	repo := &fakeBridgeChartRepo{chartDatasetGroups: map[int64]int64{101: 11}}
	chartHandler := NewChartHandler(service.NewChartService(repo))

	r := gin.New()
	api := r.Group("/api")
	RegisterCompatibilityBridgeRoutes(api, nil, nil, nil, nil, chartHandler, permMiddleware)

	req := httptest.NewRequest("POST", "/api/chartData/getData", strings.NewReader(`{"id":101}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 401 {
		t.Fatalf("expected 401 for unauthenticated chartData/getData, got %d", w.Code)
	}
}

func TestCompatibilityBridge_ChartGetData_403_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockBridgeResourcePermRepo{hasPermission: false}
	adminChecker := middleware.NewDefaultAdminChecker([]int64{})
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	permMiddleware := middleware.NewPermissionMiddleware(resourcePermSvc, exportPermSvc, adminChecker)
	permMiddleware.SetChartDatasetResolver(&mockBridgeChartDatasetResolver{datasetGroupIDs: map[int64]int64{101: 11}})

	repo := &fakeBridgeChartRepo{chartDatasetGroups: map[int64]int64{101: 11}}
	chartHandler := NewChartHandler(service.NewChartService(repo))

	r := gin.New()
	api := r.Group("/api")
	api.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(100))
		c.Next()
	})
	RegisterCompatibilityBridgeRoutes(api, nil, nil, nil, nil, chartHandler, permMiddleware)

	req := httptest.NewRequest("POST", "/api/chart/getData", strings.NewReader(`{"id":101}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 403 {
		t.Fatalf("expected 403 for forbidden chart/getData, got %d", w.Code)
	}
}

func TestCompatibilityBridge_ChartDataGetData_400_WhenDatasetFieldPretendsToBeChartID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockBridgeResourcePermRepo{hasPermission: true}
	adminChecker := middleware.NewDefaultAdminChecker([]int64{})
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	permMiddleware := middleware.NewPermissionMiddleware(resourcePermSvc, exportPermSvc, adminChecker)
	permMiddleware.SetChartDatasetResolver(&mockBridgeChartDatasetResolver{datasetGroupIDs: map[int64]int64{101: 11}})

	repo := &fakeBridgeChartRepo{chartDatasetGroups: map[int64]int64{101: 11}}
	chartHandler := NewChartHandler(service.NewChartService(repo))

	r := gin.New()
	api := r.Group("/api")
	api.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(100))
		c.Next()
	})
	RegisterCompatibilityBridgeRoutes(api, nil, nil, nil, nil, chartHandler, permMiddleware)

	req := httptest.NewRequest("POST", "/api/chartData/getData", strings.NewReader(`{"datasetGroupId":11}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected HTTP 200 with business error for missing chart id, got %d", w.Code)
	}

	var resp bridgeCodeResp
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if resp.Code != "10001" {
		t.Fatalf("expected code 10001 when dataset field pretends to be chart id, got %s", resp.Code)
	}
}

func TestCompatibilityBridge_ChartListByDQ_401_Unauthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockBridgeResourcePermRepo{hasPermission: true}
	adminChecker := middleware.NewDefaultAdminChecker([]int64{})
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	permMiddleware := middleware.NewPermissionMiddleware(resourcePermSvc, exportPermSvc, adminChecker)

	repo := &fakeBridgeChartRepo{}
	chartHandler := NewChartHandler(service.NewChartService(repo))

	r := gin.New()
	api := r.Group("/api")
	RegisterCompatibilityBridgeRoutes(api, nil, nil, nil, nil, chartHandler, permMiddleware)

	req := httptest.NewRequest("POST", "/api/chart/listByDQ/11/9", strings.NewReader(`{"type":"bar"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 401 {
		t.Fatalf("expected 401 for unauthenticated chart/listByDQ, got %d", w.Code)
	}
}

func TestCompatibilityBridge_ChartListByDQ_UsesPermissionAwareFieldsWhenGoverned(t *testing.T) {
	gin.SetMode(gin.TestMode)

	checked := true
	groupD := "d"
	groupQ := "q"
	varcharType := "VARCHAR"
	intType := "INT"
	region := "region"
	amount := "amount"
	deTypeD := 0
	deTypeQ := 2

	repo := &fakeBridgeChartRepo{
		dsFields: map[int64][]*dataset.CoreDatasetTableField{11: {
			{ID: 1, DatasetGroupID: 11, Name: &region, OriginName: &region, DataeaseName: &region, GroupType: &groupD, Type: &varcharType, DeType: &deTypeD, Checked: &checked},
			{ID: 2, DatasetGroupID: 11, Name: &amount, OriginName: &amount, DataeaseName: &amount, GroupType: &groupQ, Type: &intType, DeType: &deTypeQ, Checked: &checked},
		}},
	}
	chartService := service.NewChartService(repo)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	if err = db.AutoMigrate(&permission.DataPermColumn{}); err != nil {
		t.Fatalf("migrate permission columns failed: %v", err)
	}
	columnRepo := repository.NewColumnPermissionRepository(db)
	if err = columnRepo.Create(&permission.DataPermColumn{DatasetID: 11, DatasetGroupID: 11, FieldName: "amount", PermType: permission.PermTypeDisable, Status: 1}); err != nil {
		t.Fatalf("create disable permission failed: %v", err)
	}
	if err = columnRepo.Create(&permission.DataPermColumn{DatasetID: 11, DatasetGroupID: 11, FieldName: "region", PermType: permission.PermTypeMask, Status: 1}); err != nil {
		t.Fatalf("create mask permission failed: %v", err)
	}
	chartService.SetColumnPermissionService(service.NewColumnPermissionService(columnRepo))
	chartHandler := NewChartHandler(chartService)

	mockRepo := &mockBridgeResourcePermRepo{hasPermission: true}
	adminChecker := middleware.NewDefaultAdminChecker([]int64{})
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	permMiddleware := middleware.NewPermissionMiddleware(resourcePermSvc, exportPermSvc, adminChecker)

	r := gin.New()
	api := r.Group("/api")
	api.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(42))
		c.Next()
	})
	RegisterCompatibilityBridgeRoutes(api, nil, nil, nil, nil, chartHandler, permMiddleware)

	req := httptest.NewRequest("POST", "/api/chart/listByDQ/11/9", strings.NewReader(`{"type":"bar"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	resp := bridgeFieldListResp{}
	if err = json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if resp.Code != "000000" {
		t.Fatalf("expected code 000000, got %s", resp.Code)
	}
	if len(resp.Data.DimensionList) != 1 {
		t.Fatalf("expected masked dimension field to remain, got %d", len(resp.Data.DimensionList))
	}
	if resp.Data.DimensionList[0]["originName"] != "region" {
		t.Fatalf("expected region field to remain, got %#v", resp.Data.DimensionList[0])
	}
	if desensitized, ok := resp.Data.DimensionList[0]["desensitized"].(bool); !ok || !desensitized {
		t.Fatalf("expected region field to be marked desensitized, got %#v", resp.Data.DimensionList[0]["desensitized"])
	}
	if len(resp.Data.QuotaList) != 1 || int(resp.Data.QuotaList[0]["id"].(float64)) != -1 {
		t.Fatalf("expected only count pseudo field in quota list, got %#v", resp.Data.QuotaList)
	}
}

func TestCompatibilityBridge_DatasetDetailWithPerm_403_WhenBatchContainsForbiddenDataset(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockBridgeResourcePermRepo{
		hasPermission: true,
		governedPermissionIDs: map[int64][]int64{
			11: {1},
			12: {},
		},
	}
	adminChecker := middleware.NewDefaultAdminChecker([]int64{})
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	permMiddleware := middleware.NewPermissionMiddleware(resourcePermSvc, exportPermSvc, adminChecker)

	r := gin.New()
	api := r.Group("/api")
	api.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(42))
		c.Next()
	})
	RegisterCompatibilityBridgeRoutes(api, nil, nil, nil, &DatasetHandler{}, nil, permMiddleware)

	req := httptest.NewRequest("POST", "/api/datasetTree/detailWithPerm", strings.NewReader(`[11,12]`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 403 {
		t.Fatalf("expected 403 when batch contains forbidden dataset, got %d", w.Code)
	}
}

func setupPermissionAwareDatasetDetailRouter(t *testing.T) *gin.Engine {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	seedPermissionAwareDatasetDetailFixture(t, db)

	datasetRepo := repository.NewDatasetRepository(db)
	rowPermRepo := repository.NewRowPermissionRepository(db)
	columnPermRepo := repository.NewColumnPermissionRepository(db)
	rowPermSvc := service.NewRowPermissionService(rowPermRepo, columnPermRepo, nil, middleware.NewDefaultAdminChecker([]int64{}))
	rowPermSvc.SetDatasetFieldResolver(datasetRepo)
	datasetSvc := service.NewDatasetServiceWithPermission(datasetRepo, rowPermSvc)
	datasetHandler := NewDatasetHandler(datasetSvc)

	mockRepo := &mockBridgeResourcePermRepo{hasPermission: true}
	adminChecker := middleware.NewDefaultAdminChecker([]int64{})
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	permMiddleware := middleware.NewPermissionMiddleware(resourcePermSvc, exportPermSvc, adminChecker)

	r := gin.New()
	api := r.Group("/api")
	api.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(42))
		c.Next()
	})
	RegisterCompatibilityBridgeRoutes(api, nil, nil, nil, datasetHandler, nil, permMiddleware)

	return r
}

func seedPermissionAwareDatasetDetailFixture(t *testing.T, db *gorm.DB) {
	t.Helper()

	if err := db.AutoMigrate(&dataset.CoreDatasetGroup{}, &dataset.CoreDatasetTable{}, &dataset.CoreDatasetTableField{}, &permission.DataPermRow{}, &permission.DataPermColumn{}); err != nil {
		t.Fatalf("migrate dataset detail tables failed: %v", err)
	}

	previewTable := "dataset_detail_perm"
	createDatasetDetailFixtureRecords(t, db, previewTable)
	createDatasetDetailFixtureFields(t, db)
	createDatasetDetailFixturePermissions(t, db)
}

func createDatasetDetailFixtureRecords(t *testing.T, db *gorm.DB, previewTable string) {
	t.Helper()

	if err := db.Create(&dataset.CoreDatasetGroup{ID: 11, Name: "Sales"}).Error; err != nil {
		t.Fatalf("create dataset group failed: %v", err)
	}
	if err := db.Create(&dataset.CoreDatasetTable{ID: 111, DatasetGroupID: 11, PhysicalTable: &previewTable}).Error; err != nil {
		t.Fatalf("create dataset table failed: %v", err)
	}
	if err := db.Exec("CREATE TABLE dataset_detail_perm (region TEXT, amount INTEGER)").Error; err != nil {
		t.Fatalf("create preview table failed: %v", err)
	}
	if err := db.Exec("INSERT INTO dataset_detail_perm (region, amount) VALUES ('north', 10)").Error; err != nil {
		t.Fatalf("insert preview row failed: %v", err)
	}
}

func createDatasetDetailFixtureFields(t *testing.T, db *gorm.DB) {
	t.Helper()

	checked := true
	groupD := "d"
	groupQ := "q"
	varcharType := "VARCHAR"
	intType := "INT"
	region := "region"
	amount := "amount"
	deTypeD := 0
	deTypeQ := 2

	fieldFixtures := []*dataset.CoreDatasetTableField{
		{ID: 1, DatasetGroupID: 11, Name: &region, OriginName: &region, DataeaseName: &region, GroupType: &groupD, Type: &varcharType, DeType: &deTypeD, Checked: &checked},
		{ID: 2, DatasetGroupID: 11, Name: &amount, OriginName: &amount, DataeaseName: &amount, GroupType: &groupQ, Type: &intType, DeType: &deTypeQ, Checked: &checked},
	}
	for _, field := range fieldFixtures {
		if err := db.Create(field).Error; err != nil {
			t.Fatalf("create dataset field failed: %v", err)
		}
	}
}

func createDatasetDetailFixturePermissions(t *testing.T, db *gorm.DB) {
	t.Helper()

	permissionFixtures := []*permission.DataPermColumn{
		{DatasetID: 11, DatasetGroupID: 11, FieldName: "amount", PermType: permission.PermTypeDisable, Status: 1},
		{DatasetID: 11, DatasetGroupID: 11, FieldName: "region", PermType: permission.PermTypeMask, Status: 1},
	}
	for _, perm := range permissionFixtures {
		if err := db.Create(perm).Error; err != nil {
			t.Fatalf("create permission fixture failed: %v", err)
		}
	}
}

func assertPermissionAwareDatasetDetailResponse(t *testing.T, body []byte) {
	t.Helper()

	var resp struct {
		Code string                   `json:"code"`
		Data []map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if resp.Code != "000000" {
		t.Fatalf("expected success code 000000, got %s", resp.Code)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected single dataset detail, got %#v", resp.Data)
	}

	detail := resp.Data[0]
	fields, ok := detail["fields"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected fields object in detail, got %#v", detail["fields"])
	}
	assertPermissionAwareDatasetFields(t, fields)
	assertPermissionAwareDatasetPreview(t, detail)
}

func assertPermissionAwareDatasetFields(t *testing.T, fields map[string]interface{}) {
	t.Helper()

	dimensions, ok := fields["dimensionList"].([]interface{})
	if !ok || len(dimensions) != 1 {
		t.Fatalf("expected one dimension field, got %#v", fields["dimensionList"])
	}
	quotas, ok := fields["quotaList"].([]interface{})
	if !ok || len(quotas) != 1 {
		t.Fatalf("expected one quota field (count pseudo field), got %#v", fields["quotaList"])
	}

	dimension, ok := dimensions[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected dimension field object, got %#v", dimensions[0])
	}
	if dimension["originName"] != "region" {
		t.Fatalf("expected region dimension field, got %#v", dimension)
	}
	if desensitized, ok := dimension["desensitized"].(bool); !ok || !desensitized {
		t.Fatalf("expected region dimension field to be desensitized, got %#v", dimension)
	}

	quota, ok := quotas[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected quota field object, got %#v", quotas[0])
	}
	if quota["id"] != float64(-1) {
		t.Fatalf("expected count pseudo field to remain after permission filtering, got %#v", quota)
	}
}

func assertPermissionAwareDatasetPreview(t *testing.T, detail map[string]interface{}) {
	t.Helper()

	dataSection, ok := detail["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected preview data section, got %#v", detail["data"])
	}
	rows, ok := dataSection["data"].([]interface{})
	if !ok || len(rows) != 1 {
		t.Fatalf("expected one preview row, got %#v", dataSection["data"])
	}
	row, ok := rows[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected row object, got %#v", rows[0])
	}
	if _, exists := row["amount"]; exists {
		t.Fatalf("expected disabled amount column to be removed, got %#v", row)
	}
	if row["region"] == "north" {
		t.Fatalf("expected masked region value in preview row, got %#v", row)
	}
}

func TestCompatibilityBridge_UserOrgOption_UsesUserOptions_WhenOrgHandlerPresent(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userHandler, orgHandler := setupBridgeUserOrgHandlers(t)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("org_id", int64(1))
		c.Next()
	})
	RegisterCompatibilityBridgeRoutes(r, userHandler, orgHandler, nil, nil, nil, nil)

	req := httptest.NewRequest("GET", "/user/org/option", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	assertUserOrgOptionReturnsUserShape(t, w.Body.Bytes())
}

func TestCompatibilityBridge_UserOrgOption_UsesUserOptions_WhenOrgHandlerMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userHandler, _ := setupBridgeUserOrgHandlers(t)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("org_id", int64(1))
		c.Next()
	})
	RegisterCompatibilityBridgeRoutes(r, userHandler, nil, nil, nil, nil, nil)

	req := httptest.NewRequest("GET", "/user/org/option", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	assertUserOrgOptionReturnsUserShape(t, w.Body.Bytes())
}

func setupBridgeUserOrgHandlers(t *testing.T) (*UserHandler, *OrgHandler) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	if err = db.AutoMigrate(&user.SysUser{}, &user.SysUserRole{}, &org.SysOrg{}); err != nil {
		t.Fatalf("migrate bridge user-org tables failed: %v", err)
	}

	if err = db.Create(&org.SysOrg{OrgID: 1, OrgName: "Org A", ParentID: 0, Level: 1, Status: 1, DelFlag: 0}).Error; err != nil {
		t.Fatalf("seed org failed: %v", err)
	}
	alice := "Alice"
	if err = db.Create(&user.SysUser{UserID: 101, Username: "alice", NickName: alice, Status: 1, DelFlag: 0}).Error; err != nil {
		t.Fatalf("seed user failed: %v", err)
	}
	if err = db.Create(&user.SysUserRole{UserID: 101, RoleID: 1, OrgID: 1}).Error; err != nil {
		t.Fatalf("seed user-org relation failed: %v", err)
	}

	userRepo := repository.NewUserRepository(db)
	userRoleRepo := repository.NewUserRoleRepository(db)
	userSvc := service.NewUserService(userRepo, userRoleRepo, nil)
	userHandler := NewUserHandler(userSvc, service.NewUserImportService(userSvc))

	orgRepo := repository.NewOrgRepository(db)
	orgSvc := service.NewOrgService(orgRepo, nil, userRepo, nil)
	orgHandler := NewOrgHandler(orgSvc)

	return userHandler, orgHandler
}

func assertUserOrgOptionReturnsUserShape(t *testing.T, body []byte) {
	t.Helper()

	var resp struct {
		Code string                   `json:"code"`
		Data []map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("unmarshal /user/org/option response failed: %v", err)
	}
	if resp.Code != "000000" {
		t.Fatalf("expected success code 000000, got %s", resp.Code)
	}
	if len(resp.Data) == 0 {
		t.Fatalf("expected non-empty user option list, got %#v", resp.Data)
	}
	if _, ok := resp.Data[0]["username"]; !ok {
		t.Fatalf("expected user option payload to include username, got %#v", resp.Data[0])
	}
	if _, exists := resp.Data[0]["orgId"]; exists {
		t.Fatalf("expected /user/org/option to avoid org-list payload shape, got %#v", resp.Data[0])
	}
}
