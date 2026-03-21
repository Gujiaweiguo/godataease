package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"dataease/backend/internal/domain/export"
	"dataease/backend/internal/domain/permission"
	"dataease/backend/internal/repository"
	"dataease/backend/internal/service"
	"dataease/backend/internal/transport/http/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type mockExportRepo struct {
	tasks map[string]*export.ExportTask
}

func newMockExportRepo() *mockExportRepo {
	return &mockExportRepo{
		tasks: make(map[string]*export.ExportTask),
	}
}

func (m *mockExportRepo) AddTask(task *export.ExportTask) {
	m.tasks[task.ID] = task
}

func (m *mockExportRepo) Create(task *export.ExportTask) error {
	m.tasks[task.ID] = task
	return nil
}

func (m *mockExportRepo) GetByID(id string) (*export.ExportTask, error) {
	task, ok := m.tasks[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return task, nil
}

func (m *mockExportRepo) List(page, pageSize int, status string) ([]export.ExportTask, int64, error) {
	var result []export.ExportTask
	for _, task := range m.tasks {
		if status == "" || status == "all" || task.ExportStatus == status {
			result = append(result, *task)
		}
	}
	return result, int64(len(result)), nil
}

func (m *mockExportRepo) UpdateStatus(id string, status string) error {
	if task, ok := m.tasks[id]; ok {
		task.ExportStatus = status
	}
	return nil
}

func (m *mockExportRepo) Delete(id string) error {
	delete(m.tasks, id)
	return nil
}

func (m *mockExportRepo) DeleteBatch(ids []string) error {
	for _, id := range ids {
		delete(m.tasks, id)
	}
	return nil
}

func (m *mockExportRepo) DeleteAllByType(exportFromType string) error {
	for id, task := range m.tasks {
		if exportFromType == "" || exportFromType == "all" || task.ExportFromType == exportFromType {
			delete(m.tasks, id)
		}
	}
	return nil
}

func (m *mockExportRepo) CountByStatus() (map[string]int64, error) {
	counts := make(map[string]int64)
	for _, task := range m.tasks {
		counts[task.ExportStatus]++
	}
	return counts, nil
}

var _ repository.ExportRepositoryInterface = (*mockExportRepo)(nil)

func TestRegisterExportRoutes_ExportCenterAliases(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := newMockExportRepo()
	repo.AddTask(&export.ExportTask{ID: "task-1", UserID: 1, ExportStatus: "FAILED", ExportFromType: "dataset"})
	h := NewExportHandler(service.NewExportService(repo), nil, nil)
	r := gin.New()
	RegisterExportRoutes(r, h)

	tests := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{name: "records alias", method: http.MethodPost, path: "/exportCenter/exportTasks/records", body: `{}`},
		{name: "pager alias", method: http.MethodPost, path: "/exportCenter/exportTasks/all/1/10", body: `{}`},
		{name: "delete alias", method: http.MethodGet, path: "/exportCenter/delete/task-1"},
		{name: "batch delete alias", method: http.MethodPost, path: "/exportCenter/delete", body: `{"ids":["task-1"]}`},
		{name: "delete all alias", method: http.MethodPost, path: "/exportCenter/deleteAll/all", body: `{}`},
		{name: "retry alias", method: http.MethodPost, path: "/exportCenter/retry/task-1", body: `{}`},
		{name: "export limit alias", method: http.MethodPost, path: "/exportCenter/exportLimit", body: `{}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			if tt.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				t.Fatalf("expected 200 for %s %s, got %d", tt.method, tt.path, w.Code)
			}
		})
	}
}

type mockResourcePermRepoForExport struct {
	hasPermission bool
	userRoles     map[int64][]int64
}

func (m *mockResourcePermRepoForExport) GetPermByID(permID int64) (*permission.SysPerm, error) {
	return &permission.SysPerm{PermID: permID, PermKey: "export"}, nil
}

func (m *mockResourcePermRepoForExport) GetPermByKey(permKey string) (*permission.SysPerm, error) {
	return &permission.SysPerm{PermID: 1, PermKey: permKey}, nil
}

func (m *mockResourcePermRepoForExport) ListPerms(permType string, page, size int) ([]*permission.SysPerm, int64, error) {
	return nil, 0, nil
}

func (m *mockResourcePermRepoForExport) CreatePerm(perm *permission.SysPerm) error {
	return nil
}

func (m *mockResourcePermRepoForExport) UpdatePerm(perm *permission.SysPerm) error {
	return nil
}

func (m *mockResourcePermRepoForExport) DeletePerm(permID int64) error {
	return nil
}

func (m *mockResourcePermRepoForExport) GetUserPerms(userID int64) ([]int64, error) {
	return nil, nil
}

func (m *mockResourcePermRepoForExport) GetRolePerms(roleID int64) ([]int64, error) {
	return nil, nil
}

func (m *mockResourcePermRepoForExport) GetUserRoleIDs(userID int64) ([]int64, error) {
	if m.userRoles != nil {
		return m.userRoles[userID], nil
	}
	if m.hasPermission {
		return []int64{1}, nil
	}
	return nil, nil
}

func (m *mockResourcePermRepoForExport) CheckUserPermission(userID, permID int64) (bool, error) {
	return m.hasPermission, nil
}

func (m *mockResourcePermRepoForExport) CheckRolePermission(roleID, permID int64) (bool, error) {
	return m.hasPermission, nil
}

func (m *mockResourcePermRepoForExport) GrantPermToUser(userID, permID int64, createBy string) error {
	return nil
}

func (m *mockResourcePermRepoForExport) RevokePermFromUser(userID, permID int64) error {
	return nil
}

func (m *mockResourcePermRepoForExport) GrantPermToRole(roleID, permID int64) error {
	return nil
}

func (m *mockResourcePermRepoForExport) RevokePermFromRole(roleID, permID int64) error {
	return nil
}

// ========== 双视角接口 mock 实现 ==========
// ========== 双视角接口 mock 实现 ==========
func (m *mockResourcePermRepoForExport) GetUserResources(userID int64, resourceType string) ([]*permission.UserResourcePermVO, error) {
	return []*permission.UserResourcePermVO{}, nil
}

func (m *mockResourcePermRepoForExport) GetResourceUsers(resourceID int64, resourceType string) ([]*permission.ResourceUserPermVO, error) {
	return []*permission.ResourceUserPermVO{}, nil
}

func (m *mockResourcePermRepoForExport) ApplyGroupPermissions(groupID, resourceID int64, resourceType string) error {
	return nil
}

func (m *mockResourcePermRepoForExport) CheckPermissionConsistency() (*permission.PermissionConsistencyResult, error) {
	return &permission.PermissionConsistencyResult{Consistent: true}, nil
}

func TestExportDownload_TaskToResourceMapping(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := newMockExportRepo()
	mockRepo.AddTask(&export.ExportTask{
		ID:             "task-dataset-999",
		UserID:         100,
		FileName:       "dataset_export.xlsx",
		ExportFrom:     999,
		ExportFromType: permission.ResourceTypeDataset,
		ExportStatus:   "SUCCESS",
	})

	permRepo := &mockResourcePermRepoForExport{hasPermission: true}
	adminChecker := middleware.NewDefaultAdminChecker([]int64{})
	resourcePermSvc := service.NewResourcePermissionService(permRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)

	exportSvc := service.NewExportService(mockRepo)
	handler := NewExportHandler(exportSvc, exportPermSvc, adminChecker)

	r := gin.New()
	r.GET("/exportTasks/download/:id", func(c *gin.Context) {
		c.Set("user_id", uint64(100))
		c.Set("role", "user")
		c.Next()
	}, handler.Download)

	req := httptest.NewRequest("GET", "/exportTasks/download/task-dataset-999", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected 200 for successful export with permission, got %d, body: %s", w.Code, w.Body.String())
	}
}

func TestExportDownload_Dataset_NoPermission(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := newMockExportRepo()
	mockRepo.AddTask(&export.ExportTask{
		ID:             "task-dataset-denied",
		UserID:         100,
		FileName:       "dataset_export.xlsx",
		ExportFrom:     888,
		ExportFromType: permission.ResourceTypeDataset,
		ExportStatus:   "SUCCESS",
	})

	permRepo := &mockResourcePermRepoForExport{hasPermission: false}
	adminChecker := middleware.NewDefaultAdminChecker([]int64{})
	resourcePermSvc := service.NewResourcePermissionService(permRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)

	exportSvc := service.NewExportService(mockRepo)
	handler := NewExportHandler(exportSvc, exportPermSvc, adminChecker)

	r := gin.New()
	r.GET("/exportTasks/download/:id", func(c *gin.Context) {
		c.Set("user_id", uint64(100))
		c.Set("role", "user")
		c.Next()
	}, handler.Download)

	req := httptest.NewRequest("GET", "/exportTasks/download/task-dataset-denied", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected HTTP 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if resp["code"] != "403001" {
		t.Errorf("expected code 403001 for denied export permission, got %v", resp["code"])
	}

}

func TestExportDownload_Dashboard_AdminBypass(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := newMockExportRepo()
	mockRepo.AddTask(&export.ExportTask{
		ID:             "task-dashboard-admin",
		UserID:         200,
		FileName:       "dashboard_export.pdf",
		ExportFrom:     777,
		ExportFromType: permission.ResourceTypeDashboard,
		ExportStatus:   "SUCCESS",
	})

	permRepo := &mockResourcePermRepoForExport{hasPermission: false}
	adminChecker := middleware.NewDefaultAdminChecker([]int64{1})
	resourcePermSvc := service.NewResourcePermissionService(permRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)

	exportSvc := service.NewExportService(mockRepo)
	handler := NewExportHandler(exportSvc, exportPermSvc, adminChecker)

	r := gin.New()
	r.GET("/exportTasks/download/:id", func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Set("role", "admin")
		c.Next()
	}, handler.Download)

	req := httptest.NewRequest("GET", "/exportTasks/download/task-dashboard-admin", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected 200 for admin bypass, got %d", w.Code)
	}
}

func TestExportDownload_TaskNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := newMockExportRepo()

	permRepo := &mockResourcePermRepoForExport{hasPermission: true}
	adminChecker := middleware.NewDefaultAdminChecker([]int64{})
	resourcePermSvc := service.NewResourcePermissionService(permRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)

	exportSvc := service.NewExportService(mockRepo)
	handler := NewExportHandler(exportSvc, exportPermSvc, adminChecker)

	r := gin.New()
	r.GET("/exportTasks/download/:id", func(c *gin.Context) {
		c.Set("user_id", uint64(100))
		c.Set("role", "user")
		c.Next()
	}, handler.Download)

	req := httptest.NewRequest("GET", "/exportTasks/download/nonexistent", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected HTTP 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if resp["code"] != "404001" {
		t.Errorf("expected code 404001 for task not found, got %v", resp["code"])
	}

}

func TestExportDownload_UnauthorizedUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := newMockExportRepo()
	mockRepo.AddTask(&export.ExportTask{
		ID:             "task-other-user",
		UserID:         999,
		FileName:       "other_user_export.xlsx",
		ExportFrom:     555,
		ExportFromType: permission.ResourceTypeDataset,
		ExportStatus:   "SUCCESS",
	})

	permRepo := &mockResourcePermRepoForExport{hasPermission: true}
	adminChecker := middleware.NewDefaultAdminChecker([]int64{})
	resourcePermSvc := service.NewResourcePermissionService(permRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)

	exportSvc := service.NewExportService(mockRepo)
	handler := NewExportHandler(exportSvc, exportPermSvc, adminChecker)

	r := gin.New()
	r.GET("/exportTasks/download/:id", func(c *gin.Context) {
		c.Set("user_id", uint64(100))
		c.Set("role", "user")
		c.Next()
	}, handler.Download)

	req := httptest.NewRequest("GET", "/exportTasks/download/task-other-user", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected HTTP 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if resp["code"] != "403001" {
		t.Errorf("expected code 403001 for unauthorized task access, got %v", resp["code"])
	}

}

func TestExportDownload_ScreenResource(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := newMockExportRepo()
	mockRepo.AddTask(&export.ExportTask{
		ID:             "task-screen-333",
		UserID:         100,
		FileName:       "screen_export.png",
		ExportFrom:     333,
		ExportFromType: permission.ResourceTypeScreen,
		ExportStatus:   "SUCCESS",
	})

	permRepo := &mockResourcePermRepoForExport{hasPermission: true}
	adminChecker := middleware.NewDefaultAdminChecker([]int64{})
	resourcePermSvc := service.NewResourcePermissionService(permRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)

	exportSvc := service.NewExportService(mockRepo)
	handler := NewExportHandler(exportSvc, exportPermSvc, adminChecker)

	r := gin.New()
	r.GET("/exportTasks/download/:id", func(c *gin.Context) {
		c.Set("user_id", uint64(100))
		c.Set("role", "user")
		c.Next()
	}, handler.Download)

	req := httptest.NewRequest("GET", "/exportTasks/download/task-screen-333", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected 200 for screen export with permission, got %d", w.Code)
	}
}

func TestExportDownload_DatasourceResource(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := newMockExportRepo()
	mockRepo.AddTask(&export.ExportTask{
		ID:             "task-datasource-444",
		UserID:         100,
		FileName:       "datasource_export.xlsx",
		ExportFrom:     444,
		ExportFromType: permission.ResourceTypeDatasource,
		ExportStatus:   "SUCCESS",
	})

	permRepo := &mockResourcePermRepoForExport{hasPermission: true}
	adminChecker := middleware.NewDefaultAdminChecker([]int64{})
	resourcePermSvc := service.NewResourcePermissionService(permRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)

	exportSvc := service.NewExportService(mockRepo)
	handler := NewExportHandler(exportSvc, exportPermSvc, adminChecker)

	r := gin.New()
	r.GET("/exportTasks/download/:id", func(c *gin.Context) {
		c.Set("user_id", uint64(100))
		c.Set("role", "user")
		c.Next()
	}, handler.Download)

	req := httptest.NewRequest("GET", "/exportTasks/download/task-datasource-444", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected 200 for datasource export with permission, got %d", w.Code)
	}
}

func TestGenerateDownloadURI_TaskToResourceMapping(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := newMockExportRepo()
	mockRepo.AddTask(&export.ExportTask{
		ID:             "task-uri-test",
		UserID:         100,
		FileName:       "test_export.pdf",
		ExportFrom:     111,
		ExportFromType: permission.ResourceTypeDashboard,
		ExportStatus:   "SUCCESS",
	})

	permRepo := &mockResourcePermRepoForExport{hasPermission: true}
	adminChecker := middleware.NewDefaultAdminChecker([]int64{})
	resourcePermSvc := service.NewResourcePermissionService(permRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)

	exportSvc := service.NewExportService(mockRepo)
	handler := NewExportHandler(exportSvc, exportPermSvc, adminChecker)

	r := gin.New()
	r.GET("/exportTasks/generateDownloadUri/:id", func(c *gin.Context) {
		c.Set("user_id", uint64(100))
		c.Set("role", "user")
		c.Next()
	}, handler.GenerateDownloadURI)

	req := httptest.NewRequest("GET", "/exportTasks/generateDownloadUri/task-uri-test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected 200 for generateDownloadUri with permission, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "downloads") {
		t.Errorf("expected response to contain download URI, got: %s", w.Body.String())
	}
}

func TestGenerateDownloadURI_UnauthenticatedUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := newMockExportRepo()
	mockRepo.AddTask(&export.ExportTask{
		ID:             "task-uri-unauth",
		UserID:         100,
		FileName:       "dataset_export.xlsx",
		ExportFrom:     555,
		ExportFromType: permission.ResourceTypeDataset,
		ExportStatus:   "SUCCESS",
	})

	permRepo := &mockResourcePermRepoForExport{hasPermission: true}
	adminChecker := middleware.NewDefaultAdminChecker([]int64{})
	resourcePermSvc := service.NewResourcePermissionService(permRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)

	exportSvc := service.NewExportService(mockRepo)
	handler := NewExportHandler(exportSvc, exportPermSvc, adminChecker)

	r := gin.New()
	r.GET("/exportTasks/generateDownloadUri/:id", handler.GenerateDownloadURI)

	req := httptest.NewRequest("GET", "/exportTasks/generateDownloadUri/task-uri-unauth", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected HTTP 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if resp["code"] != "403001" {
		t.Fatalf("expected code 403001 for unauthenticated generateDownloadUri, got %#v", resp["code"])
	}
}

func TestGenerateDownloadURI_TaskNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := newMockExportRepo()
	permRepo := &mockResourcePermRepoForExport{hasPermission: true}
	adminChecker := middleware.NewDefaultAdminChecker([]int64{})
	resourcePermSvc := service.NewResourcePermissionService(permRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)

	exportSvc := service.NewExportService(mockRepo)
	handler := NewExportHandler(exportSvc, exportPermSvc, adminChecker)

	r := gin.New()
	r.GET("/exportTasks/generateDownloadUri/:id", func(c *gin.Context) {
		c.Set("user_id", uint64(100))
		c.Set("role", "user")
		c.Next()
	}, handler.GenerateDownloadURI)

	req := httptest.NewRequest("GET", "/exportTasks/generateDownloadUri/nonexistent", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected HTTP 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if resp["code"] != "404001" {
		t.Fatalf("expected code 404001 for missing download URI task, got %#v", resp["code"])
	}
}

func TestGenerateDownloadURI_Dataset_NoPermission(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := newMockExportRepo()
	mockRepo.AddTask(&export.ExportTask{
		ID:             "task-uri-denied",
		UserID:         100,
		FileName:       "dataset_export.xlsx",
		ExportFrom:     888,
		ExportFromType: permission.ResourceTypeDataset,
		ExportStatus:   "SUCCESS",
	})

	permRepo := &mockResourcePermRepoForExport{hasPermission: false}
	adminChecker := middleware.NewDefaultAdminChecker([]int64{})
	resourcePermSvc := service.NewResourcePermissionService(permRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)

	exportSvc := service.NewExportService(mockRepo)
	handler := NewExportHandler(exportSvc, exportPermSvc, adminChecker)

	r := gin.New()
	r.GET("/exportTasks/generateDownloadUri/:id", func(c *gin.Context) {
		c.Set("user_id", uint64(100))
		c.Set("role", "user")
		c.Next()
	}, handler.GenerateDownloadURI)

	req := httptest.NewRequest("GET", "/exportTasks/generateDownloadUri/task-uri-denied", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected HTTP 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if resp["code"] != "403001" {
		t.Fatalf("expected code 403001 for denied generateDownloadUri, got %#v", resp["code"])
	}
}

func TestExportDownload_TaskWithNoResource(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := newMockExportRepo()
	mockRepo.AddTask(&export.ExportTask{
		ID:             "task-no-resource",
		UserID:         100,
		FileName:       "standalone_export.xlsx",
		ExportFrom:     0,
		ExportFromType: "",
		ExportStatus:   "SUCCESS",
	})

	permRepo := &mockResourcePermRepoForExport{hasPermission: false}
	adminChecker := middleware.NewDefaultAdminChecker([]int64{})
	resourcePermSvc := service.NewResourcePermissionService(permRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)

	exportSvc := service.NewExportService(mockRepo)
	handler := NewExportHandler(exportSvc, exportPermSvc, adminChecker)

	r := gin.New()
	r.GET("/exportTasks/download/:id", func(c *gin.Context) {
		c.Set("user_id", uint64(100))
		c.Set("role", "user")
		c.Next()
	}, handler.Download)

	req := httptest.NewRequest("GET", "/exportTasks/download/task-no-resource", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected 200 for task without resource, got %d", w.Code)
	}
}

func TestExportDownload_UnknownResourceType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := newMockExportRepo()
	mockRepo.AddTask(&export.ExportTask{
		ID:             "task-unknown-type",
		UserID:         100,
		FileName:       "unknown_export.xlsx",
		ExportFrom:     222,
		ExportFromType: "unknown_type",
		ExportStatus:   "SUCCESS",
	})

	permRepo := &mockResourcePermRepoForExport{hasPermission: false}
	adminChecker := middleware.NewDefaultAdminChecker([]int64{})
	resourcePermSvc := service.NewResourcePermissionService(permRepo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)

	exportSvc := service.NewExportService(mockRepo)
	handler := NewExportHandler(exportSvc, exportPermSvc, adminChecker)

	r := gin.New()
	r.GET("/exportTasks/download/:id", func(c *gin.Context) {
		c.Set("user_id", uint64(100))
		c.Set("role", "user")
		c.Next()
	}, handler.Download)

	req := httptest.NewRequest("GET", "/exportTasks/download/task-unknown-type", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected 200 for unknown resource type, got %d", w.Code)
	}
}
