package handler

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"testing"

	"dataease/backend/internal/domain/chart"
	"dataease/backend/internal/domain/permission"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type fakeRowPermissionHandlerStore struct {
	items       []*permission.DataPermRow
	targetItems map[string][]*permission.DataPermRow
}

func (f *fakeRowPermissionHandlerStore) PagerByDatasetID(datasetID int64, page, size int) ([]*permission.DataPermRow, int64, error) {
	return f.items, int64(len(f.items)), nil
}

func (f *fakeRowPermissionHandlerStore) PagerByDatasetIDAndTarget(datasetID int64, targetType string, targetID int64, page, size int) ([]*permission.DataPermRow, int64, error) {
	if f.targetItems == nil {
		return []*permission.DataPermRow{}, 0, nil
	}
	key := fmt.Sprintf("%d:%s:%d", datasetID, targetType, targetID)
	items := f.targetItems[key]
	return items, int64(len(items)), nil
}

func (f *fakeRowPermissionHandlerStore) GetByID(id int64) (*permission.DataPermRow, error) {
	return nil, nil
}
func (f *fakeRowPermissionHandlerStore) Create(perm *permission.DataPermRow) error { return nil }
func (f *fakeRowPermissionHandlerStore) Update(perm *permission.DataPermRow) error { return nil }
func (f *fakeRowPermissionHandlerStore) Delete(id int64) error                     { return nil }

type fakeColumnPermissionHandlerStore struct{}

func (f *fakeColumnPermissionHandlerStore) PagerByDatasetID(datasetID int64, page, size int) ([]*permission.DataPermColumn, int64, error) {
	return []*permission.DataPermColumn{}, 0, nil
}
func (f *fakeColumnPermissionHandlerStore) GetByID(id int64) (*permission.DataPermColumn, error) {
	return nil, nil
}
func (f *fakeColumnPermissionHandlerStore) Create(perm *permission.DataPermColumn) error { return nil }
func (f *fakeColumnPermissionHandlerStore) Update(perm *permission.DataPermColumn) error { return nil }
func (f *fakeColumnPermissionHandlerStore) Delete(id int64) error                        { return nil }

type fakeDataPermissionFieldProvider struct{}

func (f *fakeDataPermissionFieldProvider) ListByDQ(datasetGroupID int64, chartID int64) (*chart.ChartFieldListResponse, error) {
	return &chart.ChartFieldListResponse{
		DimensionList: []chart.ChartField{{ID: 11, OriginName: "region"}},
	}, nil
}

func TestDataPermissionHandler_RowPermissionPagerByTarget(t *testing.T) {
	gin.SetMode(gin.TestMode)

	expr := `{"logic":"OR","items":[{"type":"item","fieldId":11,"filterType":"logic","term":"eq","value":"east"}]}`

	rowStore := &fakeRowPermissionHandlerStore{targetItems: map[string][]*permission.DataPermRow{
		"9:role:7": {{
			ID:             1,
			DatasetID:      9,
			AuthTargetType: permission.AuthTargetTypeRole,
			AuthTargetID:   7,
			ExpressionTree: expr,
			Status:         1,
		}},
	}}
	h := NewDataPermissionHandler(service.NewDataPermissionAdminService(rowStore, &fakeColumnPermissionHandlerStore{}, &fakeDataPermissionFieldProvider{}, nil))

	r := gin.New()
	api := r.Group("/api")
	RegisterDataPermissionRoutes(api, h)

	req := httptest.NewRequest("GET", "/api/dataset/rowPermissions/pagerByTarget/9/role/7/1/10", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp struct {
		Code string `json:"code"`
		Data struct {
			List []service.RowPermissionForm `json:"list"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if resp.Code != "000000" {
		t.Fatalf("expected code 000000, got %s", resp.Code)
	}
	if len(resp.Data.List) != 1 || resp.Data.List[0].TargetID != 7 {
		t.Fatalf("expected one role-target row permission, got %#v", resp.Data.List)
	}
}

func TestDataPermissionHandler_RowPermissionPagerByTargetRejectsInvalidTargetID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataPermissionHandler(service.NewDataPermissionAdminService(&fakeRowPermissionHandlerStore{}, &fakeColumnPermissionHandlerStore{}, &fakeDataPermissionFieldProvider{}, nil))

	r := gin.New()
	api := r.Group("/api")
	RegisterDataPermissionRoutes(api, h)

	req := httptest.NewRequest("GET", "/api/dataset/rowPermissions/pagerByTarget/9/role/invalid/1/10", nil)
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
		t.Fatalf("expected code 500000, got %#v", resp["code"])
	}
}

func TestDataPermissionHandler_RowPermissionPagerByTargetRejectsUnsupportedTargetType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataPermissionHandler(service.NewDataPermissionAdminService(&fakeRowPermissionHandlerStore{}, &fakeColumnPermissionHandlerStore{}, &fakeDataPermissionFieldProvider{}, nil))

	r := gin.New()
	api := r.Group("/api")
	RegisterDataPermissionRoutes(api, h)

	req := httptest.NewRequest("GET", "/api/dataset/rowPermissions/pagerByTarget/9/dept/7/1/10", nil)
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
		t.Fatalf("expected code 500000, got %#v", resp["code"])
	}
	if msg, _ := resp["msg"].(string); msg == "" || msg == "Failed: " {
		t.Fatalf("expected informative unsupported target type message, got %#v", resp["msg"])
	}
}

func TestDataPermissionHandler_RowPermissionPagerByTargetRejectsSysParamsTargetType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataPermissionHandler(service.NewDataPermissionAdminService(&fakeRowPermissionHandlerStore{}, &fakeColumnPermissionHandlerStore{}, &fakeDataPermissionFieldProvider{}, nil))

	r := gin.New()
	api := r.Group("/api")
	RegisterDataPermissionRoutes(api, h)

	req := httptest.NewRequest("GET", "/api/dataset/rowPermissions/pagerByTarget/9/sysParams/7/1/10", nil)
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
		t.Fatalf("expected code 500000, got %#v", resp["code"])
	}
	if msg, _ := resp["msg"].(string); msg != "Failed: [DEFERRED_DIMENSION_SYS_PARAMS] system-variable permission assignment is not supported in the current permission center; use system variable management for variable definitions" {
		t.Fatalf("unexpected sysParams message: %#v", resp["msg"])
	}
}
