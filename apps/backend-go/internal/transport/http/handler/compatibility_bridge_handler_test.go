package handler

import (
	"context"
	"encoding/json"
	"net"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"dataease/backend/internal/domain/chart"
	"dataease/backend/internal/domain/dataset"
	"dataease/backend/internal/domain/permission"
	"dataease/backend/internal/repository"
	"dataease/backend/internal/service"
	"dataease/backend/internal/transport/http/middleware"
	calcitev1 "dataease/backend/proto/calcite/v1"
	seatunnelv1 "dataease/backend/proto/seatunnel/v1"

	"github.com/gin-gonic/gin"
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
	field := &dataset.CoreDatasetTableField{
		ID:             21,
		DatasetGroupID: 41,
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
	chartHandler := NewChartHandler(service.NewChartService(repo))

	r := gin.New()
	RegisterCompatibilityBridgeRoutes(r, nil, nil, nil, nil, chartHandler, nil)

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
