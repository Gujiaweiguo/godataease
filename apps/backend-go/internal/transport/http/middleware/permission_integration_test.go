package middleware

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"dataease/backend/internal/domain/permission"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type mockResourcePermRepo struct {
	hasPermission bool
	reason        string
}

func (m *mockResourcePermRepo) GetPermByID(permID int64) (*permission.SysPerm, error) {
	return &permission.SysPerm{PermID: permID, PermKey: "view"}, nil
}

func (m *mockResourcePermRepo) GetPermByKey(permKey string) (*permission.SysPerm, error) {
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

func TestDatasetPreviewWithPerm_401_Unauthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockResourcePermRepo{hasPermission: true}
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, nil)
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, nil, nil)

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
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, nil, adminChecker)

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
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, nil, adminChecker)

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
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, nil, adminChecker)

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

func TestDataVisualizationFindById_401_Unauthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockResourcePermRepo{hasPermission: true}
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, nil)
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, nil, nil)

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
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, nil, adminChecker)

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
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, nil, adminChecker)

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
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, nil, adminChecker)

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
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, nil, adminChecker)

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
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, nil, adminChecker)

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
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, nil, adminChecker)

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
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, nil, nil)

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

func TestScreenView_401_Unauthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockResourcePermRepo{hasPermission: true}
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, nil)
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, nil, nil)

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

func TestDatasourceView_401_Unauthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockResourcePermRepo{hasPermission: true}
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, nil)
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, nil, nil)

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

func TestDashboardDelete_401_Unauthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockResourcePermRepo{hasPermission: true}
	resourcePermSvc := service.NewResourcePermissionService(mockRepo, nil)
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, nil, nil)

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
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, nil, adminChecker)

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
	permMiddleware := NewPermissionMiddleware(resourcePermSvc, nil, adminChecker)

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
