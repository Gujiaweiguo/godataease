package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"dataease/backend/internal/domain/menu"
	"dataease/backend/internal/domain/permission"
	"dataease/backend/internal/domain/role"
	"dataease/backend/internal/repository"
	"dataease/backend/internal/service"
	"dataease/backend/internal/transport/http/middleware"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type mockDatasourcePermRepo struct {
	hasPermission bool
}

func (m *mockDatasourcePermRepo) GetPermByID(permID int64) (*permission.SysPerm, error) {
	return &permission.SysPerm{PermID: permID, PermKey: permission.PermKeyView}, nil
}
func (m *mockDatasourcePermRepo) GetPermByKey(permKey string) (*permission.SysPerm, error) {
	return &permission.SysPerm{PermID: 1, PermKey: permKey}, nil
}
func (m *mockDatasourcePermRepo) ListPerms(string, int, int) ([]*permission.SysPerm, int64, error) {
	return nil, 0, nil
}
func (m *mockDatasourcePermRepo) CreatePerm(*permission.SysPerm) error { return nil }
func (m *mockDatasourcePermRepo) UpdatePerm(*permission.SysPerm) error { return nil }
func (m *mockDatasourcePermRepo) DeletePerm(int64) error               { return nil }
func (m *mockDatasourcePermRepo) GetUserPerms(int64) ([]int64, error)  { return nil, nil }
func (m *mockDatasourcePermRepo) GetRolePerms(int64) ([]int64, error)  { return nil, nil }
func (m *mockDatasourcePermRepo) GetUserRoleIDs(int64) ([]int64, error) {
	if m.hasPermission {
		return []int64{2}, nil
	}
	return nil, nil
}
func (m *mockDatasourcePermRepo) CheckUserPermission(int64, int64) (bool, error) {
	return m.hasPermission, nil
}
func (m *mockDatasourcePermRepo) CheckRolePermission(int64, int64) (bool, error) {
	return m.hasPermission, nil
}
func (m *mockDatasourcePermRepo) GrantPermToUser(int64, int64, string) error { return nil }
func (m *mockDatasourcePermRepo) RevokePermFromUser(int64, int64) error       { return nil }
func (m *mockDatasourcePermRepo) GrantPermToRole(int64, int64) error          { return nil }
func (m *mockDatasourcePermRepo) RevokePermFromRole(int64, int64) error       { return nil }
func (m *mockDatasourcePermRepo) GetUserResources(int64, string) ([]*permission.UserResourcePermVO, error) {
	return []*permission.UserResourcePermVO{}, nil
}
func (m *mockDatasourcePermRepo) GetResourceUsers(int64, string) ([]*permission.ResourceUserPermVO, error) {
	return []*permission.ResourceUserPermVO{}, nil
}
func (m *mockDatasourcePermRepo) ApplyGroupPermissions(int64, int64, string) error { return nil }
func (m *mockDatasourcePermRepo) RegisterResource(int64, string, string, *int64) error {
	return nil
}
func (m *mockDatasourcePermRepo) ReplaceResourcePermissions(int64, string, []int64) error { return nil }
func (m *mockDatasourcePermRepo) GetResourcePermissionIDs(int64, string) ([]int64, bool, error) {
	return nil, false, nil
}
func (m *mockDatasourcePermRepo) CheckPermissionConsistency() (*permission.PermissionConsistencyResult, error) {
	return &permission.PermissionConsistencyResult{Consistent: true}, nil
}

func createMenuAuthMiddlewareForTests(t *testing.T) *middleware.MenuAuthMiddleware {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&role.SysRole{}, &role.RoleMenu{}, &menu.CoreMenu{}))
	roleRepo := repository.NewRoleRepository(db)
	roleMenuRepo := repository.NewRoleMenuRepository(db)
	menuRepo := repository.NewMenuRepository(db)
	require.NoError(t, roleRepo.Create(&role.SysRole{RoleID: 2, RoleName: "role-2", RoleCode: "role-2", Status: role.StatusEnabled}))
	require.NoError(t, roleRepo.Create(&role.SysRole{RoleID: 3, RoleName: "role-3", RoleCode: "role-3", Status: role.StatusEnabled}))
	auditMenu := &menu.CoreMenu{ID: 201, Name: "Audit", Path: auditMenuPath, Type: 2, Auth: true, MenuSort: 1}
	datasourceMenu := &menu.CoreMenu{ID: 202, Name: "Datasource", Path: datasourceMenuPath, Type: 2, Auth: true, MenuSort: 2}
	require.NoError(t, menuRepo.Create(auditMenu))
	require.NoError(t, menuRepo.Create(datasourceMenu))
	require.NoError(t, roleMenuRepo.SaveRoleMenus(2, []int64{auditMenu.ID, datasourceMenu.ID}))
	return middleware.NewMenuAuthMiddleware(
		service.NewRoleMenuService(roleMenuRepo, roleRepo, menuRepo, nil),
		service.NewMenuServiceWithRoleFilter(menuRepo, roleMenuRepo),
	)
}

func createDatasourcePermissionMiddleware(hasPermission bool) *middleware.PermissionMiddleware {
	repo := &mockDatasourcePermRepo{hasPermission: hasPermission}
	adminChecker := middleware.NewDefaultAdminChecker([]int64{1})
	resourcePermSvc := service.NewResourcePermissionService(repo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	return middleware.NewPermissionMiddleware(resourcePermSvc, exportPermSvc, adminChecker)
}

func installAuthContext(r *gin.Engine, roleIDs []int64, userID uint64) {
	r.Use(func(c *gin.Context) {
		c.Set("role_ids", roleIDs)
		c.Set("user_id", userID)
		c.Next()
	})
}

func TestRegisterAuditRoutes_RequiresMenuAuthorizationForExportRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuditHandler(nil)
	menuAuth := createMenuAuthMiddlewareForTests(t)

	t.Run("authorized role can hit export route", func(t *testing.T) {
		r := gin.New()
		installAuthContext(r, []int64{2}, 2)
		RegisterAuditRoutes(r.Group(""), h, menuAuth)

		req := httptest.NewRequest(http.MethodPost, "/audit/export", strings.NewReader("{"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"code":"10001"`)
	})

	t.Run("unauthorized role gets forbidden", func(t *testing.T) {
		r := gin.New()
		installAuthContext(r, []int64{3}, 3)
		RegisterAuditRoutes(r.Group(""), h, menuAuth)

		req := httptest.NewRequest(http.MethodGet, "/audit/download?path=/etc/passwd&format=csv", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		var body bridgeCodeResp
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		assert.Equal(t, "70001", body.Code)
	})

	t.Run("admin bypasses menu auth", func(t *testing.T) {
		r := gin.New()
		installAuthContext(r, []int64{1}, 1)
		RegisterAuditRoutes(r.Group(""), h, menuAuth)

		req := httptest.NewRequest(http.MethodPost, "/audit/export", strings.NewReader("{"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"code":"10001"`)
	})
}

func TestDatasourceValidationRoutes_RequireAuthorization(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDatasourceHandler(service.NewDatasourceService(nil))
	menuAuth := createMenuAuthMiddlewareForTests(t)

	t.Run("canonical post validate requires datasource menu auth", func(t *testing.T) {
		r := gin.New()
		installAuthContext(r, []int64{3}, 3)
		RegisterDatasourceRoutes(r.Group("/api"), h, createDatasourcePermissionMiddleware(true), menuAuth)

		req := httptest.NewRequest(http.MethodPost, "/api/ds/validate", strings.NewReader("{"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("canonical get validate by id uses datasource permission", func(t *testing.T) {
		r := gin.New()
		installAuthContext(r, []int64{2}, 2)
		RegisterDatasourceRoutes(r.Group("/api"), h, createDatasourcePermissionMiddleware(true), menuAuth)

		req := httptest.NewRequest(http.MethodGet, "/api/ds/validate/1", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"code":"000000"`)
		assert.Contains(t, w.Body.String(), `"status":"Error"`)
	})

	t.Run("canonical get validate by id denies missing datasource permission", func(t *testing.T) {
		r := gin.New()
		installAuthContext(r, []int64{2}, 2)
		RegisterDatasourceRoutes(r.Group("/api"), h, createDatasourcePermissionMiddleware(false), menuAuth)

		req := httptest.NewRequest(http.MethodGet, "/api/ds/validate/1", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("compatibility aliases mirror canonical auth", func(t *testing.T) {
		r := gin.New()
		installAuthContext(r, []int64{3}, 3)
		api := r.Group("/api")
		de2api := r.Group("/de2api")
		perm := createDatasourcePermissionMiddleware(false)
		RegisterCompatibilityBridgeRoutes(api, nil, nil, h, nil, nil, perm, menuAuth)
		RegisterCompatibilityBridgeRoutes(de2api, nil, nil, h, nil, nil, perm, menuAuth)

		postReq := httptest.NewRequest(http.MethodPost, "/api/datasource/validate", strings.NewReader("{"))
		postReq.Header.Set("Content-Type", "application/json")
		postResp := httptest.NewRecorder()
		r.ServeHTTP(postResp, postReq)
		assert.Equal(t, http.StatusForbidden, postResp.Code)

		getReq := httptest.NewRequest(http.MethodGet, "/de2api/datasource/validate/1", nil)
		getResp := httptest.NewRecorder()
		r.ServeHTTP(getResp, getReq)
		assert.Equal(t, http.StatusForbidden, getResp.Code)
	})
}
