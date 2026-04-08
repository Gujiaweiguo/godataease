package middleware

import (
	"encoding/json"
	"errors"
	"net/http/httptest"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"dataease/backend/internal/domain/permission"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type mockResourcePermRepo struct {
	hasPermission         bool
	permIDsByKey          map[string]int64
	resourcePermissionIDs map[string][]int64
}

type mockChartDatasetResolver struct {
	datasetGroupIDs map[int64]int64
	err             error
}

type mockVisualizationTypeResolver struct {
	types map[int64]string
	err   error
}

func (m *mockChartDatasetResolver) GetDatasetGroupIDByChartID(chartID int64) (int64, error) {
	if m.err != nil {
		return 0, m.err
	}
	datasetGroupID, ok := m.datasetGroupIDs[chartID]
	if !ok {
		return 0, errors.New("chart dataset mapping not found")
	}
	return datasetGroupID, nil
}

func (m *mockVisualizationTypeResolver) FindDvType(id int64) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	visualizationType, ok := m.types[id]
	if !ok {
		return "", errors.New("visualization type not found")
	}
	return visualizationType, nil
}

func (m *mockResourcePermRepo) GetPermByID(permID int64) (*permission.SysPerm, error) {
	return &permission.SysPerm{PermID: permID, PermKey: "view"}, nil
}

func (m *mockResourcePermRepo) GetPermByKey(permKey string) (*permission.SysPerm, error) {
	if m.permIDsByKey != nil {
		if permID, ok := m.permIDsByKey[permKey]; ok {
			return &permission.SysPerm{PermID: permID, PermKey: permKey}, nil
		}
	}
	return &permission.SysPerm{PermID: 1, PermKey: permKey}, nil
}

func (m *mockResourcePermRepo) ListPerms(permType string, page, size int) ([]*permission.SysPerm, int64, error) {
	return nil, 0, nil
}

func (m *mockResourcePermRepo) CreatePerm(perm *permission.SysPerm) error {
	return nil
}

func (m *mockResourcePermRepo) UpdatePerm(perm *permission.SysPerm) error {
	return nil
}

func (m *mockResourcePermRepo) DeletePerm(permID int64) error {
	return nil
}

func (m *mockResourcePermRepo) GetUserPerms(userID int64) ([]int64, error) {
	return nil, nil
}

func (m *mockResourcePermRepo) GetRolePerms(roleID int64) ([]int64, error) {
	return nil, nil
}

func (m *mockResourcePermRepo) GetUserRoleIDs(userID int64) ([]int64, error) {
	if m.hasPermission {
		return []int64{1}, nil
	}
	return nil, nil
}

func (m *mockResourcePermRepo) CheckUserPermission(userID, permID int64) (bool, error) {
	return m.hasPermission, nil
}

func (m *mockResourcePermRepo) CheckRolePermission(roleID, permID int64) (bool, error) {
	return m.hasPermission, nil
}

func (m *mockResourcePermRepo) GrantPermToUser(userID, permID int64, createBy string) error {
	return nil
}

func (m *mockResourcePermRepo) RevokePermFromUser(userID, permID int64) error {
	return nil
}

func (m *mockResourcePermRepo) GrantPermToRole(roleID, permID int64) error {
	return nil
}

func (m *mockResourcePermRepo) RevokePermFromRole(roleID, permID int64) error {
	return nil
}

// ========== 双视角接口 mock 实现 ==========
// ========== 双视角接口 mock 实现 ==========
func (m *mockResourcePermRepo) GetUserResources(userID int64, resourceType string) ([]*permission.UserResourcePermVO, error) {
	return []*permission.UserResourcePermVO{}, nil
}

func (m *mockResourcePermRepo) GetResourceUsers(resourceID int64, resourceType string) ([]*permission.ResourceUserPermVO, error) {
	return []*permission.ResourceUserPermVO{}, nil
}

func (m *mockResourcePermRepo) ApplyGroupPermissions(groupID, resourceID int64, resourceType string) error {
	return nil
}

func (m *mockResourcePermRepo) RegisterResource(resourceID int64, resourceName, resourceType string, parentID *int64) error {
	return nil
}

func (m *mockResourcePermRepo) ReplaceResourcePermissions(resourceID int64, resourceType string, permIDs []int64) error {
	return nil
}

func (m *mockResourcePermRepo) GetResourcePermissionIDs(resourceID int64, resourceType string) ([]int64, bool, error) {
	if m.resourcePermissionIDs != nil {
		key := resourceType + ":" + strconv.FormatInt(resourceID, 10)
		permIDs, ok := m.resourcePermissionIDs[key]
		if ok {
			return permIDs, true, nil
		}
	}
	return nil, false, nil
}

func (m *mockResourcePermRepo) CheckPermissionConsistency() (*permission.PermissionConsistencyResult, error) {
	return &permission.PermissionConsistencyResult{Consistent: true}, nil
}

func TestDatasetPreviewWithPerm_401_Unauthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockResourcePermRepo{hasPermission: true}
	adminChecker := NewDefaultAdminChecker([]int64{})
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, exportPermSvc, adminChecker)

	r := gin.New()
	r.POST("/dataset/previewWithPerm", permMiddleware.CheckDatasetView(), func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true})
	})

	req := httptest.NewRequest("POST", "/dataset/previewWithPerm", strings.NewReader(`{"datasetGroupId":123}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 401 {
		t.Errorf("expected 401 for unauthenticated, got %d", w.Code)
	}
}

func TestDatasetPreviewWithPerm_403_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockResourcePermRepo{hasPermission: false}
	adminChecker := NewDefaultAdminChecker([]int64{})
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, exportPermSvc, adminChecker)

	r := gin.New()
	r.POST("/dataset/previewWithPerm", func(c *gin.Context) {
		c.Set("user_id", uint64(100))
		c.Next()
	}, permMiddleware.CheckDatasetView(), func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true})
	})

	req := httptest.NewRequest("POST", "/dataset/previewWithPerm", strings.NewReader(`{"datasetGroupId":456}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 403 {
		t.Errorf("expected 403 for forbidden, got %d", w.Code)
	}
}

func TestDatasetPreviewWithPerm_200_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockResourcePermRepo{hasPermission: true}
	adminChecker := NewDefaultAdminChecker([]int64{})
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, exportPermSvc, adminChecker)

	r := gin.New()
	r.POST("/dataset/previewWithPerm", func(c *gin.Context) {
		c.Set("user_id", uint64(100))
		c.Next()
	}, permMiddleware.CheckDatasetView(), func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true})
	})

	req := httptest.NewRequest("POST", "/dataset/previewWithPerm", strings.NewReader(`{"datasetGroupId":789}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected 200 for authorized, got %d", w.Code)
	}
}

func TestDatasetPreviewWithPerm_200_AdminBypass(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockResourcePermRepo{hasPermission: false}
	adminChecker := NewDefaultAdminChecker([]int64{1})
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, exportPermSvc, adminChecker)

	r := gin.New()
	r.POST("/dataset/previewWithPerm", func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	}, permMiddleware.CheckDatasetView(), func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true})
	})

	req := httptest.NewRequest("POST", "/dataset/previewWithPerm", strings.NewReader(`{"datasetGroupId":999}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected 200 for admin bypass, got %d", w.Code)
	}
}

func TestDatasetPreviewWithPerm_RowPermissionContextEstablished(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockResourcePermRepo{hasPermission: true}
	adminChecker := NewDefaultAdminChecker([]int64{})
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, exportPermSvc, adminChecker)

	r := gin.New()
	r.POST("/dataset/previewWithPerm", func(c *gin.Context) {
		c.Set("user_id", uint64(100))
		c.Next()
	}, permMiddleware.CheckDatasetView(), RowPermissionMiddleware(), func(c *gin.Context) {
		c.JSON(200, gin.H{
			"datasetId":  GetRowPermissionDatasetID(c),
			"datasetIds": GetRowPermissionDatasetIDs(c),
		})
	})

	req := httptest.NewRequest("POST", "/dataset/previewWithPerm", strings.NewReader(`{"datasetGroupId":789}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200 for authorized row-permission request, got %d", w.Code)
	}

	var resp struct {
		DatasetID  int64   `json:"datasetId"`
		DatasetIDs []int64 `json:"datasetIds"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.DatasetID != 789 {
		t.Fatalf("expected row-permission dataset id 789, got %d", resp.DatasetID)
	}
	if !reflect.DeepEqual(resp.DatasetIDs, []int64{789}) {
		t.Fatalf("expected row-permission dataset ids [789], got %#v", resp.DatasetIDs)
	}
}

func TestDatasetPreviewWithPerm_RowPermissionMiddlewareFailsClosedWithoutDatasetID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockResourcePermRepo{hasPermission: false}
	adminChecker := NewDefaultAdminChecker([]int64{1})
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, exportPermSvc, adminChecker)

	r := gin.New()
	r.POST("/dataset/previewWithPerm", func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	}, permMiddleware.CheckDatasetView(), RowPermissionMiddleware(), func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true})
	})

	req := httptest.NewRequest("POST", "/dataset/previewWithPerm", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected HTTP 200 with business error for missing row-permission context, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp["code"] != "10001" {
		t.Fatalf("expected business code 10001 for missing row-permission context, got %#v", resp["code"])
	}
}

func TestDatasetTreeDetailWithPerm_RowPermissionMiddlewareSupportsBatchIDs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockResourcePermRepo{hasPermission: false}
	adminChecker := NewDefaultAdminChecker([]int64{1})
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, exportPermSvc, adminChecker)

	r := gin.New()
	r.POST("/datasetTree/detailWithPerm", func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	}, permMiddleware.CheckDatasetView(), RowPermissionMiddleware(), func(c *gin.Context) {
		c.JSON(200, gin.H{
			"datasetId":  GetRowPermissionDatasetID(c),
			"datasetIds": GetRowPermissionDatasetIDs(c),
		})
	})

	req := httptest.NewRequest("POST", "/datasetTree/detailWithPerm", strings.NewReader(`{"ids":[321,322,321]}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200 for batch row-permission context, got %d", w.Code)
	}

	var resp struct {
		DatasetID  int64   `json:"datasetId"`
		DatasetIDs []int64 `json:"datasetIds"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.DatasetID != 321 {
		t.Fatalf("expected first row-permission dataset id 321, got %d", resp.DatasetID)
	}
	if !reflect.DeepEqual(resp.DatasetIDs, []int64{321, 322}) {
		t.Fatalf("expected row-permission dataset ids [321 322], got %#v", resp.DatasetIDs)
	}
}

func TestDatasetTreeDetailWithPerm_RowPermissionMiddlewareSupportsRawArrayBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockResourcePermRepo{hasPermission: false}
	adminChecker := NewDefaultAdminChecker([]int64{1})
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, exportPermSvc, adminChecker)

	r := gin.New()
	r.POST("/datasetTree/detailWithPerm", func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	}, permMiddleware.CheckDatasetView(), RowPermissionMiddleware(), func(c *gin.Context) {
		c.JSON(200, gin.H{
			"datasetId":  GetRowPermissionDatasetID(c),
			"datasetIds": GetRowPermissionDatasetIDs(c),
		})
	})

	req := httptest.NewRequest("POST", "/datasetTree/detailWithPerm", strings.NewReader(`[901,902,901]`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200 for raw-array row-permission context, got %d", w.Code)
	}

	var resp struct {
		DatasetID  int64   `json:"datasetId"`
		DatasetIDs []int64 `json:"datasetIds"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.DatasetID != 901 {
		t.Fatalf("expected first row-permission dataset id 901, got %d", resp.DatasetID)
	}
	if !reflect.DeepEqual(resp.DatasetIDs, []int64{901, 902}) {
		t.Fatalf("expected row-permission dataset ids [901 902], got %#v", resp.DatasetIDs)
	}
}

func TestChartData_401_Unauthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockResourcePermRepo{hasPermission: true}
	adminChecker := NewDefaultAdminChecker([]int64{})
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, exportPermSvc, adminChecker)
	permMiddleware.SetChartDatasetResolver(&mockChartDatasetResolver{datasetGroupIDs: map[int64]int64{123: 789}})

	r := gin.New()
	r.POST("/chart/data", permMiddleware.CheckChartDataView(), RowPermissionMiddleware(), func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true})
	})

	req := httptest.NewRequest("POST", "/chart/data", strings.NewReader(`{"id":123}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 401 {
		t.Fatalf("expected 401 for unauthenticated chart request, got %d", w.Code)
	}
}

func TestChartData_403_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockResourcePermRepo{hasPermission: false}
	adminChecker := NewDefaultAdminChecker([]int64{})
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, exportPermSvc, adminChecker)
	permMiddleware.SetChartDatasetResolver(&mockChartDatasetResolver{datasetGroupIDs: map[int64]int64{123: 789}})

	r := gin.New()
	r.POST("/chart/data", func(c *gin.Context) {
		c.Set("user_id", uint64(100))
		c.Next()
	}, permMiddleware.CheckChartDataView(), RowPermissionMiddleware(), func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true})
	})

	req := httptest.NewRequest("POST", "/chart/data", strings.NewReader(`{"id":123}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 403 {
		t.Fatalf("expected 403 for forbidden chart request, got %d", w.Code)
	}
}

func TestChartData_200_SuccessWithRowPermissionContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockResourcePermRepo{hasPermission: true}
	adminChecker := NewDefaultAdminChecker([]int64{})
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, exportPermSvc, adminChecker)
	permMiddleware.SetChartDatasetResolver(&mockChartDatasetResolver{datasetGroupIDs: map[int64]int64{123: 789}})

	r := gin.New()
	r.POST("/chart/data", func(c *gin.Context) {
		c.Set("user_id", uint64(100))
		c.Next()
	}, permMiddleware.CheckChartDataView(), RowPermissionMiddleware(), func(c *gin.Context) {
		c.JSON(200, gin.H{
			"datasetId":  GetRowPermissionDatasetID(c),
			"datasetIds": GetRowPermissionDatasetIDs(c),
		})
	})

	req := httptest.NewRequest("POST", "/chart/data", strings.NewReader(`{"id":123}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200 for authorized chart request, got %d", w.Code)
	}

	var resp struct {
		DatasetID  int64   `json:"datasetId"`
		DatasetIDs []int64 `json:"datasetIds"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.DatasetID != 789 {
		t.Fatalf("expected row-permission dataset id 789, got %d", resp.DatasetID)
	}
	if !reflect.DeepEqual(resp.DatasetIDs, []int64{789}) {
		t.Fatalf("expected row-permission dataset ids [789], got %#v", resp.DatasetIDs)
	}
}

func TestChartData_200_AdminBypassStillSeedsDatasetContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockResourcePermRepo{hasPermission: false}
	adminChecker := NewDefaultAdminChecker([]int64{1})
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, exportPermSvc, adminChecker)
	permMiddleware.SetChartDatasetResolver(&mockChartDatasetResolver{datasetGroupIDs: map[int64]int64{123: 789}})

	r := gin.New()
	r.POST("/chart/data", func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	}, permMiddleware.CheckChartDataView(), RowPermissionMiddleware(), func(c *gin.Context) {
		c.JSON(200, gin.H{
			"datasetId":  GetRowPermissionDatasetID(c),
			"datasetIds": GetRowPermissionDatasetIDs(c),
		})
	})

	req := httptest.NewRequest("POST", "/chart/data", strings.NewReader(`{"id":123}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200 for admin chart request, got %d", w.Code)
	}

	var resp struct {
		DatasetID  int64   `json:"datasetId"`
		DatasetIDs []int64 `json:"datasetIds"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.DatasetID != 789 || !reflect.DeepEqual(resp.DatasetIDs, []int64{789}) {
		t.Fatalf("expected admin flow to seed dataset context [789], got id=%d ids=%#v", resp.DatasetID, resp.DatasetIDs)
	}
}

func TestChartData_FailsClosedWithoutChartID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockResourcePermRepo{hasPermission: true}
	adminChecker := NewDefaultAdminChecker([]int64{})
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, exportPermSvc, adminChecker)
	permMiddleware.SetChartDatasetResolver(&mockChartDatasetResolver{datasetGroupIDs: map[int64]int64{123: 789}})

	r := gin.New()
	r.POST("/chart/data", func(c *gin.Context) {
		c.Set("user_id", uint64(100))
		c.Next()
	}, permMiddleware.CheckChartDataView(), RowPermissionMiddleware(), func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true})
	})

	req := httptest.NewRequest("POST", "/chart/data", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected HTTP 200 with business error for missing chart id, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp["code"] != "10001" {
		t.Fatalf("expected business code 10001 for missing chart id, got %#v", resp["code"])
	}
}

func TestChartData_FailsClosedWhenDatasetFieldPretendsToBeChartID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockResourcePermRepo{hasPermission: true}
	adminChecker := NewDefaultAdminChecker([]int64{})
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, exportPermSvc, adminChecker)
	permMiddleware.SetChartDatasetResolver(&mockChartDatasetResolver{datasetGroupIDs: map[int64]int64{123: 789}})

	r := gin.New()
	r.POST("/chart/data", func(c *gin.Context) {
		c.Set("user_id", uint64(100))
		c.Next()
	}, permMiddleware.CheckChartDataView(), RowPermissionMiddleware(), func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true})
	})

	req := httptest.NewRequest("POST", "/chart/data", strings.NewReader(`{"datasetGroupId":789}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected HTTP 200 with business error for missing chart id, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp["code"] != "10001" {
		t.Fatalf("expected business code 10001 when dataset field is used instead of chart id, got %#v", resp["code"])
	}
	if !strings.Contains(resp["msg"].(string), "chart id is required") {
		t.Fatalf("expected chart id required message, got %#v", resp["msg"])
	}
}

func TestChartData_FailsClosedWhenChartDatasetResolutionFails(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockResourcePermRepo{hasPermission: true}
	adminChecker := NewDefaultAdminChecker([]int64{})
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, exportPermSvc, adminChecker)
	permMiddleware.SetChartDatasetResolver(&mockChartDatasetResolver{err: errors.New("chart lookup failed")})

	r := gin.New()
	r.POST("/chart/data", func(c *gin.Context) {
		c.Set("user_id", uint64(100))
		c.Next()
	}, permMiddleware.CheckChartDataView(), RowPermissionMiddleware(), func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true})
	})

	req := httptest.NewRequest("POST", "/chart/data", strings.NewReader(`{"id":123}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected HTTP 200 with business error for chart resolution failure, got %d", w.Code)
	}

	var resp struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Code != "500000" {
		t.Fatalf("expected business code 500000 for chart resolution failure, got %s", resp.Code)
	}
	if !strings.Contains(resp.Msg, "chart lookup failed") {
		t.Fatalf("expected failure message to mention chart lookup failure, got %q", resp.Msg)
	}
}

func TestDataVisualizationFindById_401_Unauthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockResourcePermRepo{hasPermission: true}
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, nil)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, exportPermSvc, nil)

	r := gin.New()
	r.POST("/dataVisualization/findById", permMiddleware.CheckDashboardView(), func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true})
	})

	req := httptest.NewRequest("POST", "/dataVisualization/findById", strings.NewReader(`{"id":123}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 401 {
		t.Errorf("expected 401 for unauthenticated, got %d", w.Code)
	}
}

func TestDataVisualizationFindById_403_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockResourcePermRepo{hasPermission: false}
	adminChecker := NewDefaultAdminChecker([]int64{})
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, exportPermSvc, adminChecker)

	r := gin.New()
	r.POST("/dataVisualization/findById", func(c *gin.Context) {
		c.Set("user_id", uint64(200))
		c.Next()
	}, permMiddleware.CheckDashboardView(), func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true})
	})

	req := httptest.NewRequest("POST", "/dataVisualization/findById", strings.NewReader(`{"id":456}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 403 {
		t.Errorf("expected 403 for forbidden, got %d", w.Code)
	}
}

func TestDataVisualizationFindById_200_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockResourcePermRepo{hasPermission: true}
	adminChecker := NewDefaultAdminChecker([]int64{})
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, exportPermSvc, adminChecker)

	r := gin.New()
	r.POST("/dataVisualization/findById", func(c *gin.Context) {
		c.Set("user_id", uint64(200))
		c.Next()
	}, permMiddleware.CheckDashboardView(), func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true})
	})

	req := httptest.NewRequest("POST", "/dataVisualization/findById", strings.NewReader(`{"id":789}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected 200 for authorized, got %d", w.Code)
	}
}

func TestDataVisualizationFindById_200_AdminBypass(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockResourcePermRepo{hasPermission: false}
	adminChecker := NewDefaultAdminChecker([]int64{1})
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, exportPermSvc, adminChecker)

	r := gin.New()
	r.POST("/dataVisualization/findById", func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	}, permMiddleware.CheckDashboardView(), func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true})
	})

	req := httptest.NewRequest("POST", "/dataVisualization/findById", strings.NewReader(`{"id":999}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected 200 for admin bypass, got %d", w.Code)
	}
}

func TestDashboardEdit_403_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockResourcePermRepo{hasPermission: false}
	adminChecker := NewDefaultAdminChecker([]int64{})
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, exportPermSvc, adminChecker)

	r := gin.New()
	r.POST("/dataVisualization/updateCanvas", func(c *gin.Context) {
		c.Set("user_id", uint64(300))
		c.Next()
	}, permMiddleware.CheckDashboardEdit(), func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true})
	})

	req := httptest.NewRequest("POST", "/dataVisualization/updateCanvas", strings.NewReader(`{"id":111}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 403 {
		t.Errorf("expected 403 for edit forbidden, got %d", w.Code)
	}
}

func TestDashboardEdit_200_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockResourcePermRepo{hasPermission: true}
	adminChecker := NewDefaultAdminChecker([]int64{})
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, exportPermSvc, adminChecker)

	r := gin.New()
	r.POST("/dataVisualization/updateCanvas", func(c *gin.Context) {
		c.Set("user_id", uint64(300))
		c.Next()
	}, permMiddleware.CheckDashboardEdit(), func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true})
	})

	req := httptest.NewRequest("POST", "/dataVisualization/updateCanvas", strings.NewReader(`{"id":222}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected 200 for edit authorized, got %d", w.Code)
	}
}

func TestPermissionDenied_ResponseBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockResourcePermRepo{hasPermission: false}
	adminChecker := NewDefaultAdminChecker([]int64{})
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, exportPermSvc, adminChecker)

	r := gin.New()
	r.POST("/dataset/previewWithPerm", func(c *gin.Context) {
		c.Set("user_id", uint64(100))
		c.Next()
	}, permMiddleware.CheckDatasetView(), func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true})
	})

	req := httptest.NewRequest("POST", "/dataset/previewWithPerm", strings.NewReader(`{"datasetGroupId":456}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if _, ok := resp["msg"]; !ok {
		t.Error("expected 'msg' field in error response")
	}
}

func TestDatasetExport_401_Unauthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockResourcePermRepo{hasPermission: true}
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, nil)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, exportPermSvc, nil)

	r := gin.New()
	r.POST("/export/dataset/:id", permMiddleware.CheckDatasetExport(), func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true})
	})

	req := httptest.NewRequest("POST", "/export/dataset/123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 401 {
		t.Errorf("expected 401 for unauthenticated dataset export, got %d", w.Code)
	}
}

func TestExportTasksDownload_401_Unauthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockResourcePermRepo{hasPermission: true}
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, nil)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, exportPermSvc, nil)

	r := gin.New()
	r.GET("/exportTasks/:id/download", permMiddleware.CheckExportPermission("task"), func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/exportTasks/123/download", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 401 {
		t.Errorf("expected 401 for unauthenticated exportTasks download, got %d", w.Code)
	}
}

func TestExportTasksDownload_403_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockResourcePermRepo{hasPermission: false}
	adminChecker := NewDefaultAdminChecker([]int64{})
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, exportPermSvc, adminChecker)

	r := gin.New()
	r.GET("/exportTasks/:id/download", func(c *gin.Context) {
		c.Set("user_id", uint64(500))
		c.Next()
	}, permMiddleware.CheckExportPermission("task"), func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/exportTasks/456/download", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 403 {
		t.Errorf("expected 403 for forbidden exportTasks download, got %d", w.Code)
	}
}

func TestExportTasksDownload_200_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockResourcePermRepo{hasPermission: true}
	adminChecker := NewDefaultAdminChecker([]int64{})
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, exportPermSvc, adminChecker)

	r := gin.New()
	r.GET("/exportTasks/:id/download", func(c *gin.Context) {
		c.Set("user_id", uint64(500))
		c.Next()
	}, permMiddleware.CheckExportPermission("task"), func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/exportTasks/789/download", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected 200 for authorized exportTasks download, got %d", w.Code)
	}
}

func TestExportTasksDownload_200_AdminBypass(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockResourcePermRepo{hasPermission: false}
	adminChecker := NewDefaultAdminChecker([]int64{1})
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, exportPermSvc, adminChecker)

	r := gin.New()
	r.GET("/exportTasks/:id/download", func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	}, permMiddleware.CheckExportPermission("task"), func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/exportTasks/999/download", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected 200 for admin bypass exportTasks download, got %d", w.Code)
	}
}

func TestScreenView_401_Unauthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockResourcePermRepo{hasPermission: true}
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, nil)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, exportPermSvc, nil)

	r := gin.New()
	r.POST("/screen/:id", permMiddleware.CheckScreenView(), func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true})
	})

	req := httptest.NewRequest("POST", "/screen/123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 401 {
		t.Errorf("expected 401 for unauthenticated screen view, got %d", w.Code)
	}
}

func TestScreenView_403_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockResourcePermRepo{hasPermission: false}
	adminChecker := NewDefaultAdminChecker([]int64{})
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, exportPermSvc, adminChecker)

	r := gin.New()
	r.POST("/screen/:id", func(c *gin.Context) {
		c.Set("user_id", uint64(500))
		c.Next()
	}, permMiddleware.CheckScreenView(), func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true})
	})

	req := httptest.NewRequest("POST", "/screen/123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 403 {
		t.Fatalf("expected 403 for forbidden screen view, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if resp["code"] != "70001" {
		t.Fatalf("expected code 70001, got %#v", resp["code"])
	}
}

func TestDatasourceView_401_Unauthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockResourcePermRepo{hasPermission: true}
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, nil)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, exportPermSvc, nil)

	r := gin.New()
	r.POST("/datasource/:id", permMiddleware.CheckDatasourceView(), func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true})
	})

	req := httptest.NewRequest("POST", "/datasource/123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 401 {
		t.Errorf("expected 401 for unauthenticated datasource view, got %d", w.Code)
	}
}

func TestDatasourceView_403_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockResourcePermRepo{hasPermission: false}
	adminChecker := NewDefaultAdminChecker([]int64{})
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, exportPermSvc, adminChecker)

	r := gin.New()
	r.POST("/datasource/:id", func(c *gin.Context) {
		c.Set("user_id", uint64(300))
		c.Next()
	}, permMiddleware.CheckDatasourceView(), func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true})
	})

	req := httptest.NewRequest("POST", "/datasource/123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 403 {
		t.Fatalf("expected 403 for forbidden datasource view, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if resp["code"] != "70001" {
		t.Fatalf("expected code 70001, got %#v", resp["code"])
	}
}

func TestDatasourceListAliases_403_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockResourcePermRepo{hasPermission: false}
	adminChecker := NewDefaultAdminChecker([]int64{})
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, exportPermSvc, adminChecker)

	r := gin.New()
	for _, path := range []string{"/api/ds/list", "/api/datasource/list", "/de2api/datasource/list"} {
		r.POST(path, func(c *gin.Context) {
			c.Set("user_id", uint64(300))
			c.Next()
		}, permMiddleware.CheckDatasourceView(), func(c *gin.Context) {
			c.JSON(200, gin.H{"success": true})
		})
	}

	for _, path := range []string{"/api/ds/list", "/api/datasource/list", "/de2api/datasource/list"} {
		req := httptest.NewRequest("POST", path, strings.NewReader(`{"id":123}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != 403 {
			t.Fatalf("expected 403 for forbidden datasource list alias %s, got %d", path, w.Code)
		}

		var resp map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal response failed for %s: %v", path, err)
		}
		if resp["code"] != "70001" {
			t.Fatalf("expected code 70001 for %s, got %#v", path, resp["code"])
		}
	}
}

func TestVisualizationView_401_Unauthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockResourcePermRepo{hasPermission: true}
	adminChecker := NewDefaultAdminChecker([]int64{})
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, exportPermSvc, adminChecker)
	permMiddleware.SetVisualizationTypeResolver(&mockVisualizationTypeResolver{types: map[int64]string{123: "dashboard"}})

	r := gin.New()
	r.POST("/dataVisualization/findById", permMiddleware.CheckVisualizationView(), func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true})
	})

	req := httptest.NewRequest("POST", "/dataVisualization/findById", strings.NewReader(`{"id":123}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 401 {
		t.Fatalf("expected 401 for unauthenticated visualization view, got %d", w.Code)
	}
}

func TestVisualizationEdit_403_ForbiddenDashboard(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockResourcePermRepo{hasPermission: false}
	adminChecker := NewDefaultAdminChecker([]int64{})
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, exportPermSvc, adminChecker)
	permMiddleware.SetVisualizationTypeResolver(&mockVisualizationTypeResolver{types: map[int64]string{456: "dashboard"}})

	r := gin.New()
	r.POST("/dataVisualization/updateCanvas", func(c *gin.Context) {
		c.Set("user_id", uint64(300))
		c.Next()
	}, permMiddleware.CheckVisualizationEdit(), func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true})
	})

	req := httptest.NewRequest("POST", "/dataVisualization/updateCanvas", strings.NewReader(`{"id":456}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 403 {
		t.Fatalf("expected 403 for forbidden visualization edit, got %d", w.Code)
	}
}

func TestVisualizationEdit_200_ScreenSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockResourcePermRepo{hasPermission: true}
	adminChecker := NewDefaultAdminChecker([]int64{})
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, exportPermSvc, adminChecker)
	permMiddleware.SetVisualizationTypeResolver(&mockVisualizationTypeResolver{types: map[int64]string{789: "dataV"}})

	r := gin.New()
	r.POST("/dataVisualization/updateCanvas", func(c *gin.Context) {
		c.Set("user_id", uint64(301))
		c.Next()
	}, permMiddleware.CheckVisualizationEdit(), func(c *gin.Context) {
		c.JSON(200, gin.H{
			"resourceType":  c.GetString(ResourceTypeKey),
			"resourceID":    c.MustGet(ResourceIDKey),
			"permissionKey": c.GetString(PermissionKeyKey),
		})
	})

	req := httptest.NewRequest("POST", "/dataVisualization/updateCanvas", strings.NewReader(`{"id":789}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200 for authorized screen visualization edit, got %d", w.Code)
	}

	var resp struct {
		ResourceType  string `json:"resourceType"`
		ResourceID    int64  `json:"resourceID"`
		PermissionKey string `json:"permissionKey"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if resp.ResourceType != permission.ResourceTypeScreen {
		t.Fatalf("expected resource type %s, got %s", permission.ResourceTypeScreen, resp.ResourceType)
	}
	if resp.ResourceID != 789 {
		t.Fatalf("expected resource id 789, got %d", resp.ResourceID)
	}
	if resp.PermissionKey != permission.PermKeyEdit {
		t.Fatalf("expected permission key %s, got %s", permission.PermKeyEdit, resp.PermissionKey)
	}
}

func TestVisualizationEdit_401_RemainingRootRoutesUnauthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	routes := []string{
		"/dataVisualization/updateBase",
		"/dataVisualization/move",
		"/dataVisualization/updatePublishStatus",
		"/dataVisualization/recoverToPublished",
	}

	for _, route := range routes {
		t.Run(route, func(t *testing.T) {
			mockRepo := &mockResourcePermRepo{hasPermission: true}
			adminChecker := NewDefaultAdminChecker([]int64{})
			resourcePermSvc := service.NewResourcePermissionService(mockRepo, adminChecker)
			exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
			permMiddleware := NewPermissionMiddleware(resourcePermSvc, exportPermSvc, adminChecker)
			permMiddleware.SetVisualizationTypeResolver(&mockVisualizationTypeResolver{types: map[int64]string{123: "dashboard"}})

			r := gin.New()
			r.POST(route, permMiddleware.CheckVisualizationEdit(), func(c *gin.Context) {
				c.JSON(200, gin.H{"success": true})
			})

			req := httptest.NewRequest("POST", route, strings.NewReader(`{"id":123}`))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != 401 {
				t.Fatalf("expected 401 for unauthenticated visualization edit on %s, got %d", route, w.Code)
			}
		})
	}
}

func TestVisualizationEdit_200_RemainingRootRoutesDashboardSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	routes := []string{
		"/dataVisualization/updateBase",
		"/dataVisualization/move",
		"/dataVisualization/updatePublishStatus",
		"/dataVisualization/recoverToPublished",
	}

	for _, route := range routes {
		t.Run(route, func(t *testing.T) {
			mockRepo := &mockResourcePermRepo{hasPermission: true}
			adminChecker := NewDefaultAdminChecker([]int64{})
			resourcePermSvc := service.NewResourcePermissionService(mockRepo, adminChecker)
			exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
			permMiddleware := NewPermissionMiddleware(resourcePermSvc, exportPermSvc, adminChecker)
			permMiddleware.SetVisualizationTypeResolver(&mockVisualizationTypeResolver{types: map[int64]string{123: "dashboard"}})

			r := gin.New()
			r.POST(route, func(c *gin.Context) {
				c.Set("user_id", uint64(401))
				c.Next()
			}, permMiddleware.CheckVisualizationEdit(), func(c *gin.Context) {
				c.JSON(200, gin.H{
					"resourceType":  c.GetString(ResourceTypeKey),
					"resourceID":    c.MustGet(ResourceIDKey),
					"permissionKey": c.GetString(PermissionKeyKey),
				})
			})

			req := httptest.NewRequest("POST", route, strings.NewReader(`{"id":123}`))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != 200 {
				t.Fatalf("expected 200 for authorized visualization edit on %s, got %d", route, w.Code)
			}
		})
	}
}

func TestVisualizationEdit_403_RemainingRootRoutesForbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)

	routes := []string{
		"/dataVisualization/updateBase",
		"/dataVisualization/move",
		"/dataVisualization/updatePublishStatus",
		"/dataVisualization/recoverToPublished",
	}

	for _, route := range routes {
		t.Run(route, func(t *testing.T) {
			mockRepo := &mockResourcePermRepo{hasPermission: false}
			adminChecker := NewDefaultAdminChecker([]int64{})
			resourcePermSvc := service.NewResourcePermissionService(mockRepo, adminChecker)
			exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
			permMiddleware := NewPermissionMiddleware(resourcePermSvc, exportPermSvc, adminChecker)
			permMiddleware.SetVisualizationTypeResolver(&mockVisualizationTypeResolver{types: map[int64]string{456: "dashboard"}})

			r := gin.New()
			r.POST(route, func(c *gin.Context) {
				c.Set("user_id", uint64(402))
				c.Next()
			}, permMiddleware.CheckVisualizationEdit(), func(c *gin.Context) {
				c.JSON(200, gin.H{"success": true})
			})

			req := httptest.NewRequest("POST", route, strings.NewReader(`{"id":456}`))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != 403 {
				t.Fatalf("expected 403 for forbidden visualization edit on %s, got %d", route, w.Code)
			}
		})
	}
}

func TestVisualizationEdit_FailsClosedOnResolverErrorsForRemainingRootRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	routes := []string{
		"/dataVisualization/updateBase",
		"/dataVisualization/move",
		"/dataVisualization/updatePublishStatus",
		"/dataVisualization/recoverToPublished",
	}

	for _, route := range routes {
		t.Run(route, func(t *testing.T) {
			mockRepo := &mockResourcePermRepo{hasPermission: true}
			adminChecker := NewDefaultAdminChecker([]int64{})
			resourcePermSvc := service.NewResourcePermissionService(mockRepo, adminChecker)
			exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
			permMiddleware := NewPermissionMiddleware(resourcePermSvc, exportPermSvc, adminChecker)
			permMiddleware.SetVisualizationTypeResolver(&mockVisualizationTypeResolver{err: errors.New("visualization lookup failed")})

			r := gin.New()
			r.POST(route, func(c *gin.Context) {
				c.Set("user_id", uint64(403))
				c.Next()
			}, permMiddleware.CheckVisualizationEdit(), func(c *gin.Context) {
				c.JSON(200, gin.H{"success": true})
			})

			req := httptest.NewRequest("POST", route, strings.NewReader(`{"id":999}`))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != 200 {
				t.Fatalf("expected HTTP 200 with business error for resolver failure on %s, got %d", route, w.Code)
			}

			var resp struct {
				Code string `json:"code"`
				Msg  string `json:"msg"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("unmarshal response failed: %v", err)
			}
			if resp.Code != "500000" {
				t.Fatalf("expected code 500000 for %s, got %s", route, resp.Code)
			}
			if !strings.Contains(resp.Msg, "visualization lookup failed") {
				t.Fatalf("expected resolver failure message for %s, got %q", route, resp.Msg)
			}
		})
	}
}

func TestVisualizationView_FailsClosedWhenResolverErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockResourcePermRepo{hasPermission: true}
	adminChecker := NewDefaultAdminChecker([]int64{})
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, exportPermSvc, adminChecker)
	permMiddleware.SetVisualizationTypeResolver(&mockVisualizationTypeResolver{err: errors.New("visualization lookup failed")})

	r := gin.New()
	r.POST("/dataVisualization/findById", func(c *gin.Context) {
		c.Set("user_id", uint64(302))
		c.Next()
	}, permMiddleware.CheckVisualizationView(), func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true})
	})

	req := httptest.NewRequest("POST", "/dataVisualization/findById", strings.NewReader(`{"id":999}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected HTTP 200 with business error for resolver failure, got %d", w.Code)
	}

	var resp struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if resp.Code != "500000" {
		t.Fatalf("expected code 500000, got %s", resp.Code)
	}
	if !strings.Contains(resp.Msg, "visualization lookup failed") {
		t.Fatalf("expected resolver failure message, got %q", resp.Msg)
	}
}

func TestVisualizationView_FailsClosedWhenVisualizationTypeUnsupported(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockResourcePermRepo{hasPermission: true}
	adminChecker := NewDefaultAdminChecker([]int64{})
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, exportPermSvc, adminChecker)
	permMiddleware.SetVisualizationTypeResolver(&mockVisualizationTypeResolver{types: map[int64]string{1001: ""}})

	r := gin.New()
	r.POST("/dataVisualization/findById", func(c *gin.Context) {
		c.Set("user_id", uint64(303))
		c.Next()
	}, permMiddleware.CheckVisualizationView(), func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true})
	})

	req := httptest.NewRequest("POST", "/dataVisualization/findById", strings.NewReader(`{"id":1001}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected HTTP 200 with business error for unsupported visualization type, got %d", w.Code)
	}

	var resp struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if resp.Code != "500000" {
		t.Fatalf("expected code 500000, got %s", resp.Code)
	}
	if !strings.Contains(resp.Msg, `unsupported visualization type: ""`) {
		t.Fatalf("expected unsupported type message, got %q", resp.Msg)
	}
}

func TestVisualizationParentEdit_401_Unauthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockResourcePermRepo{hasPermission: true}
	adminChecker := NewDefaultAdminChecker([]int64{})
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, exportPermSvc, adminChecker)

	r := gin.New()
	r.POST("/dataVisualization/saveCanvas", permMiddleware.CheckVisualizationParentEdit(), func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true})
	})

	req := httptest.NewRequest("POST", "/dataVisualization/saveCanvas", strings.NewReader(`{"pid":123,"type":"dashboard"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 401 {
		t.Fatalf("expected 401 for unauthenticated visualization parent edit, got %d", w.Code)
	}
}

func TestVisualizationParentEdit_200_DashboardSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockResourcePermRepo{hasPermission: true}
	adminChecker := NewDefaultAdminChecker([]int64{})
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, exportPermSvc, adminChecker)

	r := gin.New()
	r.POST("/dataVisualization/saveCanvas", func(c *gin.Context) {
		c.Set("user_id", uint64(304))
		c.Next()
	}, permMiddleware.CheckVisualizationParentEdit(), func(c *gin.Context) {
		c.JSON(200, gin.H{
			"resourceType":  c.GetString(ResourceTypeKey),
			"resourceID":    c.MustGet(ResourceIDKey),
			"permissionKey": c.GetString(PermissionKeyKey),
		})
	})

	req := httptest.NewRequest("POST", "/dataVisualization/saveCanvas", strings.NewReader(`{"pid":123,"type":"dashboard"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200 for authorized dashboard parent edit, got %d", w.Code)
	}

	var resp struct {
		ResourceType  string `json:"resourceType"`
		ResourceID    int64  `json:"resourceID"`
		PermissionKey string `json:"permissionKey"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if resp.ResourceType != permission.ResourceTypeDashboard {
		t.Fatalf("expected resource type %s, got %s", permission.ResourceTypeDashboard, resp.ResourceType)
	}
	if resp.ResourceID != 123 {
		t.Fatalf("expected resource id 123, got %d", resp.ResourceID)
	}
	if resp.PermissionKey != permission.PermKeyEdit {
		t.Fatalf("expected permission key %s, got %s", permission.PermKeyEdit, resp.PermissionKey)
	}
}

func TestVisualizationParentEdit_403_DashboardForbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockResourcePermRepo{hasPermission: false}
	adminChecker := NewDefaultAdminChecker([]int64{})
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, exportPermSvc, adminChecker)

	r := gin.New()
	r.POST("/dataVisualization/saveCanvas", func(c *gin.Context) {
		c.Set("user_id", uint64(305))
		c.Next()
	}, permMiddleware.CheckVisualizationParentEdit(), func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true})
	})

	req := httptest.NewRequest("POST", "/dataVisualization/saveCanvas", strings.NewReader(`{"pid":456,"type":"dashboard"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 403 {
		t.Fatalf("expected 403 for forbidden dashboard parent edit, got %d", w.Code)
	}
}

func TestVisualizationParentEdit_200_ScreenSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockResourcePermRepo{hasPermission: true}
	adminChecker := NewDefaultAdminChecker([]int64{})
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, exportPermSvc, adminChecker)

	r := gin.New()
	r.POST("/dataVisualization/saveCanvas", func(c *gin.Context) {
		c.Set("user_id", uint64(306))
		c.Next()
	}, permMiddleware.CheckVisualizationParentEdit(), func(c *gin.Context) {
		c.JSON(200, gin.H{
			"resourceType":  c.GetString(ResourceTypeKey),
			"resourceID":    c.MustGet(ResourceIDKey),
			"permissionKey": c.GetString(PermissionKeyKey),
		})
	})

	req := httptest.NewRequest("POST", "/dataVisualization/saveCanvas", strings.NewReader(`{"pid":789,"type":"dataV"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200 for authorized screen parent edit, got %d", w.Code)
	}

	var resp struct {
		ResourceType  string `json:"resourceType"`
		ResourceID    int64  `json:"resourceID"`
		PermissionKey string `json:"permissionKey"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if resp.ResourceType != permission.ResourceTypeScreen {
		t.Fatalf("expected resource type %s, got %s", permission.ResourceTypeScreen, resp.ResourceType)
	}
	if resp.ResourceID != 789 {
		t.Fatalf("expected resource id 789, got %d", resp.ResourceID)
	}
	if resp.PermissionKey != permission.PermKeyEdit {
		t.Fatalf("expected permission key %s, got %s", permission.PermKeyEdit, resp.PermissionKey)
	}
}

func TestVisualizationParentEdit_403_ScreenForbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockResourcePermRepo{hasPermission: false}
	adminChecker := NewDefaultAdminChecker([]int64{})
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, exportPermSvc, adminChecker)

	r := gin.New()
	r.POST("/dataVisualization/saveCanvas", func(c *gin.Context) {
		c.Set("user_id", uint64(307))
		c.Next()
	}, permMiddleware.CheckVisualizationParentEdit(), func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true})
	})

	req := httptest.NewRequest("POST", "/dataVisualization/saveCanvas", strings.NewReader(`{"pid":790,"type":"dataV"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 403 {
		t.Fatalf("expected 403 for forbidden screen parent edit, got %d", w.Code)
	}
}

func TestVisualizationParentEdit_FailsClosedWhenPIDMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockResourcePermRepo{hasPermission: true}
	adminChecker := NewDefaultAdminChecker([]int64{})
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, exportPermSvc, adminChecker)

	r := gin.New()
	r.POST("/dataVisualization/saveCanvas", func(c *gin.Context) {
		c.Set("user_id", uint64(308))
		c.Next()
	}, permMiddleware.CheckVisualizationParentEdit(), func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true})
	})

	req := httptest.NewRequest("POST", "/dataVisualization/saveCanvas", strings.NewReader(`{"type":"dashboard"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected HTTP 200 with business error for missing pid, got %d", w.Code)
	}

	var resp struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if resp.Code != "500000" {
		t.Fatalf("expected code 500000, got %s", resp.Code)
	}
	if !strings.Contains(resp.Msg, "visualization parent id is required") {
		t.Fatalf("expected missing parent id message, got %q", resp.Msg)
	}
}

func TestVisualizationParentEdit_FailsClosedWhenPIDInvalid(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockResourcePermRepo{hasPermission: true}
	adminChecker := NewDefaultAdminChecker([]int64{})
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, exportPermSvc, adminChecker)

	r := gin.New()
	r.POST("/dataVisualization/saveCanvas", func(c *gin.Context) {
		c.Set("user_id", uint64(309))
		c.Next()
	}, permMiddleware.CheckVisualizationParentEdit(), func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true})
	})

	req := httptest.NewRequest("POST", "/dataVisualization/saveCanvas", strings.NewReader(`{"pid":0,"type":"dashboard"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected HTTP 200 with business error for invalid pid, got %d", w.Code)
	}

	var resp struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if resp.Code != "500000" {
		t.Fatalf("expected code 500000, got %s", resp.Code)
	}
	if !strings.Contains(resp.Msg, "visualization parent id is required") {
		t.Fatalf("expected invalid parent id message, got %q", resp.Msg)
	}
}

func TestVisualizationParentEdit_FailsClosedWhenVisualizationTypeUnsupported(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockResourcePermRepo{hasPermission: true}
	adminChecker := NewDefaultAdminChecker([]int64{})
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, exportPermSvc, adminChecker)

	r := gin.New()
	r.POST("/dataVisualization/saveCanvas", func(c *gin.Context) {
		c.Set("user_id", uint64(310))
		c.Next()
	}, permMiddleware.CheckVisualizationParentEdit(), func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true})
	})

	req := httptest.NewRequest("POST", "/dataVisualization/saveCanvas", strings.NewReader(`{"pid":1001,"type":""}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected HTTP 200 with business error for unsupported visualization type, got %d", w.Code)
	}

	var resp struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if resp.Code != "500000" {
		t.Fatalf("expected code 500000, got %s", resp.Code)
	}
	if !strings.Contains(resp.Msg, `unsupported visualization type: ""`) {
		t.Fatalf("expected unsupported type message, got %q", resp.Msg)
	}
}

func TestVisualizationParentEdit_200_AdminBypass(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockResourcePermRepo{hasPermission: false}
	adminChecker := NewDefaultAdminChecker([]int64{1})
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, exportPermSvc, adminChecker)

	r := gin.New()
	r.POST("/dataVisualization/saveCanvas", func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	}, permMiddleware.CheckVisualizationParentEdit(), func(c *gin.Context) {
		c.JSON(200, gin.H{
			"resourceType":  c.GetString(ResourceTypeKey),
			"resourceID":    c.MustGet(ResourceIDKey),
			"permissionKey": c.GetString(PermissionKeyKey),
		})
	})

	req := httptest.NewRequest("POST", "/dataVisualization/saveCanvas", strings.NewReader(`{"pid":1002,"type":"dashboard"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200 for admin bypass visualization parent edit, got %d", w.Code)
	}
}

func TestVisualizationCopy_401_Unauthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockResourcePermRepo{hasPermission: true}
	adminChecker := NewDefaultAdminChecker([]int64{})
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, exportPermSvc, adminChecker)
	permMiddleware.SetVisualizationTypeResolver(&mockVisualizationTypeResolver{types: map[int64]string{11: "dashboard"}})

	r := gin.New()
	r.POST("/dataVisualization/copy", permMiddleware.CheckVisualizationCopy(), func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true})
	})

	req := httptest.NewRequest("POST", "/dataVisualization/copy", strings.NewReader(`{"id":11,"pid":21,"name":"copy","type":"dashboard"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 401 {
		t.Fatalf("expected 401 for unauthenticated visualization copy, got %d", w.Code)
	}
}

func TestVisualizationCopy_200_DashboardSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockResourcePermRepo{
		hasPermission: true,
		permIDsByKey: map[string]int64{
			"dashboard:view": 1,
			"dashboard:edit": 2,
		},
		resourcePermissionIDs: map[string][]int64{
			"dashboard:11": {1},
			"dashboard:21": {2},
		},
	}
	adminChecker := NewDefaultAdminChecker([]int64{})
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, exportPermSvc, adminChecker)
	permMiddleware.SetVisualizationTypeResolver(&mockVisualizationTypeResolver{types: map[int64]string{11: "dashboard"}})

	r := gin.New()
	r.POST("/dataVisualization/copy", func(c *gin.Context) {
		c.Set("user_id", uint64(311))
		c.Next()
	}, permMiddleware.CheckVisualizationCopy(), func(c *gin.Context) {
		c.JSON(200, gin.H{
			"resourceType":  c.GetString(ResourceTypeKey),
			"resourceID":    c.MustGet(ResourceIDKey),
			"permissionKey": c.GetString(PermissionKeyKey),
		})
	})

	req := httptest.NewRequest("POST", "/dataVisualization/copy", strings.NewReader(`{"id":11,"pid":21,"name":"copy","type":"dashboard"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200 for authorized visualization copy, got %d", w.Code)
	}

	var resp struct {
		ResourceType  string `json:"resourceType"`
		ResourceID    int64  `json:"resourceID"`
		PermissionKey string `json:"permissionKey"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if resp.ResourceType != permission.ResourceTypeDashboard {
		t.Fatalf("expected destination resource type %s, got %s", permission.ResourceTypeDashboard, resp.ResourceType)
	}
	if resp.ResourceID != 21 {
		t.Fatalf("expected destination resource id 21, got %d", resp.ResourceID)
	}
	if resp.PermissionKey != permission.PermKeyEdit {
		t.Fatalf("expected permission key %s, got %s", permission.PermKeyEdit, resp.PermissionKey)
	}
}

func TestVisualizationCopy_200_ScreenSuccessFallsBackToSourceType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockResourcePermRepo{
		hasPermission: true,
		permIDsByKey: map[string]int64{
			"screen:view": 1,
			"screen:edit": 2,
		},
		resourcePermissionIDs: map[string][]int64{
			"screen:12": {1},
			"screen:22": {2},
		},
	}
	adminChecker := NewDefaultAdminChecker([]int64{})
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, exportPermSvc, adminChecker)
	permMiddleware.SetVisualizationTypeResolver(&mockVisualizationTypeResolver{types: map[int64]string{12: "dataV"}})

	r := gin.New()
	r.POST("/dataVisualization/copy", func(c *gin.Context) {
		c.Set("user_id", uint64(312))
		c.Next()
	}, permMiddleware.CheckVisualizationCopy(), func(c *gin.Context) {
		c.JSON(200, gin.H{
			"resourceType":  c.GetString(ResourceTypeKey),
			"resourceID":    c.MustGet(ResourceIDKey),
			"permissionKey": c.GetString(PermissionKeyKey),
		})
	})

	req := httptest.NewRequest("POST", "/dataVisualization/copy", strings.NewReader(`{"id":12,"pid":22,"name":"copy"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200 for authorized screen visualization copy, got %d", w.Code)
	}
}

func TestVisualizationCopy_403_SourceForbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockResourcePermRepo{
		hasPermission: true,
		permIDsByKey: map[string]int64{
			"dashboard:view": 1,
			"dashboard:edit": 2,
		},
		resourcePermissionIDs: map[string][]int64{
			"dashboard:13": {},
			"dashboard:23": {2},
		},
	}
	adminChecker := NewDefaultAdminChecker([]int64{})
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, exportPermSvc, adminChecker)
	permMiddleware.SetVisualizationTypeResolver(&mockVisualizationTypeResolver{types: map[int64]string{13: "dashboard"}})

	r := gin.New()
	r.POST("/dataVisualization/copy", func(c *gin.Context) {
		c.Set("user_id", uint64(313))
		c.Next()
	}, permMiddleware.CheckVisualizationCopy(), func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true})
	})

	req := httptest.NewRequest("POST", "/dataVisualization/copy", strings.NewReader(`{"id":13,"pid":23,"name":"copy","type":"dashboard"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 403 {
		t.Fatalf("expected 403 for forbidden visualization copy source, got %d", w.Code)
	}
}

func TestVisualizationCopy_403_DestinationForbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockResourcePermRepo{
		permIDsByKey: map[string]int64{
			"dashboard:view": 1,
			"dashboard:edit": 2,
		},
		resourcePermissionIDs: map[string][]int64{
			"dashboard:14": {1},
			"dashboard:24": {},
		},
	}
	adminChecker := NewDefaultAdminChecker([]int64{})
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, exportPermSvc, adminChecker)
	permMiddleware.SetVisualizationTypeResolver(&mockVisualizationTypeResolver{types: map[int64]string{14: "dashboard"}})

	r := gin.New()
	r.POST("/dataVisualization/copy", func(c *gin.Context) {
		c.Set("user_id", uint64(314))
		c.Next()
	}, permMiddleware.CheckVisualizationCopy(), func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true})
	})

	req := httptest.NewRequest("POST", "/dataVisualization/copy", strings.NewReader(`{"id":14,"pid":24,"name":"copy","type":"dashboard"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 403 {
		t.Fatalf("expected 403 for forbidden visualization copy destination, got %d", w.Code)
	}
}

func TestVisualizationCopy_FailsClosedWhenSourceIDMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockResourcePermRepo{hasPermission: true}
	adminChecker := NewDefaultAdminChecker([]int64{})
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, exportPermSvc, adminChecker)
	permMiddleware.SetVisualizationTypeResolver(&mockVisualizationTypeResolver{types: map[int64]string{15: "dashboard"}})

	r := gin.New()
	r.POST("/dataVisualization/copy", func(c *gin.Context) {
		c.Set("user_id", uint64(315))
		c.Next()
	}, permMiddleware.CheckVisualizationCopy(), func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true})
	})

	req := httptest.NewRequest("POST", "/dataVisualization/copy", strings.NewReader(`{"pid":25,"name":"copy","type":"dashboard"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected HTTP 200 with business error for missing source id, got %d", w.Code)
	}

	var resp struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if resp.Code != "500000" {
		t.Fatalf("expected code 500000, got %s", resp.Code)
	}
	if !strings.Contains(resp.Msg, "copy source id is required") {
		t.Fatalf("expected missing source id message, got %q", resp.Msg)
	}
}

func TestVisualizationCopy_FailsClosedWhenDestinationPIDMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockResourcePermRepo{hasPermission: true}
	adminChecker := NewDefaultAdminChecker([]int64{})
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, exportPermSvc, adminChecker)
	permMiddleware.SetVisualizationTypeResolver(&mockVisualizationTypeResolver{types: map[int64]string{16: "dashboard"}})

	r := gin.New()
	r.POST("/dataVisualization/copy", func(c *gin.Context) {
		c.Set("user_id", uint64(316))
		c.Next()
	}, permMiddleware.CheckVisualizationCopy(), func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true})
	})

	req := httptest.NewRequest("POST", "/dataVisualization/copy", strings.NewReader(`{"id":16,"name":"copy","type":"dashboard"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected HTTP 200 with business error for missing destination pid, got %d", w.Code)
	}

	var resp struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if resp.Code != "500000" {
		t.Fatalf("expected code 500000, got %s", resp.Code)
	}
	if !strings.Contains(resp.Msg, "copy destination parent id is required") {
		t.Fatalf("expected missing destination parent id message, got %q", resp.Msg)
	}
}

func TestVisualizationCopy_FailsClosedWhenResolverErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockResourcePermRepo{hasPermission: true}
	adminChecker := NewDefaultAdminChecker([]int64{})
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, exportPermSvc, adminChecker)
	permMiddleware.SetVisualizationTypeResolver(&mockVisualizationTypeResolver{err: errors.New("visualization lookup failed")})

	r := gin.New()
	r.POST("/dataVisualization/copy", func(c *gin.Context) {
		c.Set("user_id", uint64(317))
		c.Next()
	}, permMiddleware.CheckVisualizationCopy(), func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true})
	})

	req := httptest.NewRequest("POST", "/dataVisualization/copy", strings.NewReader(`{"id":17,"pid":27,"name":"copy"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected HTTP 200 with business error for resolver failure, got %d", w.Code)
	}

	var resp struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if resp.Code != "500000" {
		t.Fatalf("expected code 500000, got %s", resp.Code)
	}
	if !strings.Contains(resp.Msg, "visualization lookup failed") {
		t.Fatalf("expected resolver failure message, got %q", resp.Msg)
	}
}

func TestVisualizationCopy_FailsClosedWhenEffectiveTypeUnsupported(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockResourcePermRepo{hasPermission: true}
	adminChecker := NewDefaultAdminChecker([]int64{})
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, exportPermSvc, adminChecker)
	permMiddleware.SetVisualizationTypeResolver(&mockVisualizationTypeResolver{types: map[int64]string{18: ""}})

	r := gin.New()
	r.POST("/dataVisualization/copy", func(c *gin.Context) {
		c.Set("user_id", uint64(318))
		c.Next()
	}, permMiddleware.CheckVisualizationCopy(), func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true})
	})

	req := httptest.NewRequest("POST", "/dataVisualization/copy", strings.NewReader(`{"id":18,"pid":28,"name":"copy"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected HTTP 200 with business error for unsupported effective type, got %d", w.Code)
	}

	var resp struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if resp.Code != "500000" {
		t.Fatalf("expected code 500000, got %s", resp.Code)
	}
	if !strings.Contains(resp.Msg, `unsupported visualization type: ""`) {
		t.Fatalf("expected unsupported type message, got %q", resp.Msg)
	}
}

func TestVisualizationCopy_200_AdminBypass(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockResourcePermRepo{
		permIDsByKey: map[string]int64{
			"dashboard:view": 1,
			"dashboard:edit": 2,
		},
		resourcePermissionIDs: map[string][]int64{
			"dashboard:19": {},
			"dashboard:29": {},
		},
	}
	adminChecker := NewDefaultAdminChecker([]int64{1})
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, exportPermSvc, adminChecker)
	permMiddleware.SetVisualizationTypeResolver(&mockVisualizationTypeResolver{types: map[int64]string{19: "dashboard"}})

	r := gin.New()
	r.POST("/dataVisualization/copy", func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	}, permMiddleware.CheckVisualizationCopy(), func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true})
	})

	req := httptest.NewRequest("POST", "/dataVisualization/copy", strings.NewReader(`{"id":19,"pid":29,"name":"copy","type":"dashboard"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200 for admin bypass visualization copy, got %d", w.Code)
	}
}

func TestDashboardDelete_401_Unauthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockResourcePermRepo{hasPermission: true}
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, nil)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, exportPermSvc, nil)

	r := gin.New()
	r.DELETE("/dataVisualization/deleteLogic/:id", permMiddleware.CheckDashboardEdit(), func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true})
	})

	req := httptest.NewRequest("DELETE", "/dataVisualization/deleteLogic/123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 401 {
		t.Errorf("expected 401 for unauthenticated delete, got %d", w.Code)
	}
}

func TestDashboardDelete_403_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockResourcePermRepo{hasPermission: false}
	adminChecker := NewDefaultAdminChecker([]int64{})
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, exportPermSvc, adminChecker)

	r := gin.New()
	r.DELETE("/dataVisualization/deleteLogic/:id", func(c *gin.Context) {
		c.Set("user_id", uint64(400))
		c.Next()
	}, permMiddleware.CheckDashboardEdit(), func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true})
	})

	req := httptest.NewRequest("DELETE", "/dataVisualization/deleteLogic/456", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 403 {
		t.Errorf("expected 403 for delete forbidden, got %d", w.Code)
	}
}

func TestDashboardDelete_200_AdminBypass(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockResourcePermRepo{hasPermission: false}
	adminChecker := NewDefaultAdminChecker([]int64{1})
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, exportPermSvc, adminChecker)

	r := gin.New()
	r.DELETE("/dataVisualization/deleteLogic/:id", func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	}, permMiddleware.CheckDashboardEdit(), func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true})
	})

	req := httptest.NewRequest("DELETE", "/dataVisualization/deleteLogic/999", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected 200 for admin delete bypass, got %d", w.Code)
	}
}
