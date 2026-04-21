//go:build integration

package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"dataease/backend/internal/domain/menu"
	"dataease/backend/internal/domain/permission"
	"dataease/backend/internal/domain/role"
	"dataease/backend/internal/domain/user"
	"dataease/backend/internal/repository"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func openPermissionCompatHandlerTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	host := envOrDefaultPermissionCompat("TEST_DB_HOST", "localhost")
	port := envOrDefaultPermissionCompat("TEST_DB_PORT", "3306")
	dbUser := envOrDefaultPermissionCompat("TEST_DB_USER", "root")
	password := envOrDefaultPermissionCompat("TEST_DB_PASSWORD", "Admin168")
	name := envOrDefaultPermissionCompat("TEST_DB_NAME", "dataease_test")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", dbUser, password, host, port, name)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&role.SysRole{},
		&role.RoleMenu{},
		&menu.CoreMenu{},
		&permission.SysPerm{},
		&permission.SysResource{},
		&permission.SysResourcePerm{},
		&permission.SysRolePerm{},
		&permission.SysUserPerm{},
		&user.SysUser{},
		&user.SysUserRole{},
	))
	require.NoError(t, db.Exec("SET FOREIGN_KEY_CHECKS = 0").Error)
	require.NoError(t, db.Exec("DELETE FROM sys_resource_perm").Error)
	require.NoError(t, db.Exec("DELETE FROM sys_resource").Error)
	require.NoError(t, db.Exec("DELETE FROM sys_user_perm").Error)
	require.NoError(t, db.Exec("DELETE FROM sys_user_role").Error)
	require.NoError(t, db.Exec("DELETE FROM sys_role_perm").Error)
	require.NoError(t, db.Exec("DELETE FROM sys_perm").Error)
	require.NoError(t, db.Exec("DELETE FROM sys_user").Error)
	require.NoError(t, db.Exec("DELETE FROM sys_role_menu").Error)
	require.NoError(t, db.Exec("DELETE FROM core_menu").Error)
	require.NoError(t, db.Exec("DELETE FROM sys_role").Error)
	require.NoError(t, db.Exec("SET FOREIGN_KEY_CHECKS = 1").Error)

	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})

	return db
}

func envOrDefaultPermissionCompat(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func TestPermissionCompatHandlerIntegration_MenuPermissionAndSaveRoundTrip(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := openPermissionCompatHandlerTestDB(t)
	roleRepo := repository.NewRoleRepository(db)
	menuRepo := repository.NewMenuRepository(db)
	roleMenuRepo := repository.NewRoleMenuRepository(db)

	menuSvc := service.NewMenuService(menuRepo)
	roleMenuSvc := service.NewRoleMenuService(roleMenuRepo, roleRepo, menuRepo, nil)
	h := NewPermissionCompatHandler(menuSvc, nil, roleMenuSvc, nil)

	testRole := &role.SysRole{
		RoleName: "Permission Menu Role",
		RoleCode: "permission_menu_role",
		Status:   1,
	}
	require.NoError(t, roleRepo.Create(testRole))

	rootMenu := &menu.CoreMenu{Name: "System", Pid: 0, MenuSort: 1, Path: "/system", Type: 2, Auth: true}
	childMenuA := &menu.CoreMenu{Name: "User", Pid: 0, MenuSort: 2, Path: "/system/user", Type: 2, Auth: true}
	childMenuB := &menu.CoreMenu{Name: "Permission", Pid: 0, MenuSort: 3, Path: "/system/permission", Type: 2, Auth: true}
	require.NoError(t, menuRepo.Create(rootMenu))
	require.NoError(t, menuRepo.Create(childMenuA))
	require.NoError(t, menuRepo.Create(childMenuB))
	require.NoError(t, roleMenuRepo.SaveRoleMenus(testRole.RoleID, []int64{childMenuA.ID}))

	r := gin.New()
	api := r.Group("/api")
	RegisterPermissionCompatRoutes(api, h)

	newQueryReq := func() *http.Request {
		req := httptest.NewRequest(http.MethodPost, "/api/auth/menuPermission", strings.NewReader(fmt.Sprintf(`{"roleId":%d}`, testRole.RoleID)))
		req.Header.Set("Content-Type", "application/json")
		return req
	}

	queryResp := httptest.NewRecorder()
	r.ServeHTTP(queryResp, newQueryReq())
	require.Equal(t, http.StatusOK, queryResp.Code)

	var queryBody struct {
		Code string `json:"code"`
		Data struct {
			MenuTree []map[string]interface{} `json:"menuTree"`
			MenuIDs  []int64                  `json:"menuIds"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(queryResp.Body.Bytes(), &queryBody))
	assert.Equal(t, "000000", queryBody.Code)
	assert.NotEmpty(t, queryBody.Data.MenuTree)
	assert.Equal(t, []int64{childMenuA.ID}, queryBody.Data.MenuIDs)

	saveReq := httptest.NewRequest(http.MethodPost, "/api/auth/saveMenuPer", strings.NewReader(fmt.Sprintf(`{"roleId":%d,"menuIds":[%d,%d]}`, testRole.RoleID, childMenuA.ID, childMenuB.ID)))
	saveReq.Header.Set("Content-Type", "application/json")
	saveResp := httptest.NewRecorder()
	r.ServeHTTP(saveResp, saveReq)
	require.Equal(t, http.StatusOK, saveResp.Code)

	var saveBody map[string]interface{}
	require.NoError(t, json.Unmarshal(saveResp.Body.Bytes(), &saveBody))
	assert.Equal(t, "000000", saveBody["code"])

	queryRespAfterSave := httptest.NewRecorder()
	r.ServeHTTP(queryRespAfterSave, newQueryReq())
	require.Equal(t, http.StatusOK, queryRespAfterSave.Code)
	require.NoError(t, json.Unmarshal(queryRespAfterSave.Body.Bytes(), &queryBody))
	assert.ElementsMatch(t, []int64{childMenuA.ID, childMenuB.ID}, queryBody.Data.MenuIDs)

	invalidSaveReq := httptest.NewRequest(http.MethodPost, "/api/auth/saveMenuPer", strings.NewReader(fmt.Sprintf(`{"roleId":%d,"menuIds":[999999]}`, testRole.RoleID)))
	invalidSaveReq.Header.Set("Content-Type", "application/json")
	invalidSaveResp := httptest.NewRecorder()
	r.ServeHTTP(invalidSaveResp, invalidSaveReq)
	require.Equal(t, http.StatusOK, invalidSaveResp.Code)

	var invalidBody map[string]interface{}
	require.NoError(t, json.Unmarshal(invalidSaveResp.Body.Bytes(), &invalidBody))
	assert.Equal(t, "500000", invalidBody["code"])
	assert.Contains(t, invalidBody["msg"], "menu not found")
}

func TestPermissionCompatHandlerIntegration_MenuTargetPermissionAndSaveRoundTrip(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := openPermissionCompatHandlerTestDB(t)
	roleRepo := repository.NewRoleRepository(db)
	menuRepo := repository.NewMenuRepository(db)
	roleMenuRepo := repository.NewRoleMenuRepository(db)

	menuSvc := service.NewMenuService(menuRepo)
	roleMenuSvc := service.NewRoleMenuService(roleMenuRepo, roleRepo, menuRepo, nil)
	h := NewPermissionCompatHandler(menuSvc, nil, roleMenuSvc, nil)

	testRole := &role.SysRole{RoleName: "Menu Target Role", RoleCode: "menu_target_role", Status: 1}
	require.NoError(t, roleRepo.Create(testRole))

	menuA := &menu.CoreMenu{Name: "User", Pid: 0, MenuSort: 1, Path: "/system/user", Type: 2, Auth: true}
	menuB := &menu.CoreMenu{Name: "Permission", Pid: 0, MenuSort: 2, Path: "/system/permission", Type: 2, Auth: true}
	require.NoError(t, menuRepo.Create(menuA))
	require.NoError(t, menuRepo.Create(menuB))
	require.NoError(t, roleMenuRepo.SaveRoleMenus(testRole.RoleID, []int64{menuA.ID}))

	r := gin.New()
	api := r.Group("/api")
	RegisterPermissionCompatRoutes(api, h)

	queryReq := httptest.NewRequest(http.MethodPost, "/api/auth/menuTargetPermission", strings.NewReader(fmt.Sprintf(`{"roleId":%d}`, testRole.RoleID)))
	queryReq.Header.Set("Content-Type", "application/json")
	queryResp := httptest.NewRecorder()
	r.ServeHTTP(queryResp, queryReq)
	require.Equal(t, http.StatusOK, queryResp.Code)

	var queryBody struct {
		Code string `json:"code"`
		Data struct {
			MenuIDs []int64 `json:"menuIds"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(queryResp.Body.Bytes(), &queryBody))
	assert.Equal(t, "000000", queryBody.Code)
	assert.Equal(t, []int64{menuA.ID}, queryBody.Data.MenuIDs)

	saveReq := httptest.NewRequest(http.MethodPost, "/api/auth/saveMenuTargetPer", strings.NewReader(fmt.Sprintf(`{"roleId":%d,"targetPerms":[{"targetType":"role","targetId":%d,"permIds":[%d,%d]}]}`, testRole.RoleID, testRole.RoleID, menuA.ID, menuB.ID)))
	saveReq.Header.Set("Content-Type", "application/json")
	saveResp := httptest.NewRecorder()
	r.ServeHTTP(saveResp, saveReq)
	require.Equal(t, http.StatusOK, saveResp.Code)

	var saveBody map[string]interface{}
	require.NoError(t, json.Unmarshal(saveResp.Body.Bytes(), &saveBody))
	assert.Equal(t, "000000", saveBody["code"])

	persistedMenuIDs, err := roleMenuRepo.GetMenuIDsByRoleID(testRole.RoleID)
	require.NoError(t, err)
	assert.ElementsMatch(t, []int64{menuA.ID, menuB.ID}, persistedMenuIDs)
}

func TestPermissionCompatHandlerIntegration_BusiTargetPermissionAndSaveRoundTrip(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := openPermissionCompatHandlerTestDB(t)
	roleRepo := repository.NewRoleRepository(db)
	resourceRepo := repository.NewResourcePermissionRepository(db)
	resourceSvc := service.NewResourcePermissionService(resourceRepo, nil)
	h := NewPermissionCompatHandler(nil, nil, nil, resourceSvc)

	testRole := &role.SysRole{
		RoleName: "Permission Resource Role",
		RoleCode: "permission_resource_role",
		Status:   1,
	}
	require.NoError(t, roleRepo.Create(testRole))

	testUser := &user.SysUser{
		Username: "resource_tester",
		Password: "irrelevant",
		NickName: "Resource Tester",
		Status:   1,
		DelFlag:  0,
	}
	require.NoError(t, db.Create(testUser).Error)
	require.NoError(t, db.Create(&user.SysUserRole{UserID: testUser.UserID, RoleID: testRole.RoleID, OrgID: 1}).Error)

	dashboardView := &permission.SysPerm{PermName: "仪表板查看", PermKey: "dashboard:view", PermType: permission.PermTypeData, Status: 1, DelFlag: 0}
	dashboardEdit := &permission.SysPerm{PermName: "仪表板编辑", PermKey: "dashboard:edit", PermType: permission.PermTypeData, Status: 1, DelFlag: 0}
	datasetView := &permission.SysPerm{PermName: "数据集查看", PermKey: "dataset:view", PermType: permission.PermTypeData, Status: 1, DelFlag: 0}
	require.NoError(t, db.Create(dashboardView).Error)
	require.NoError(t, db.Create(dashboardEdit).Error)
	require.NoError(t, db.Create(datasetView).Error)
	require.NoError(t, db.Create(&permission.SysRolePerm{RoleID: testRole.RoleID, PermID: dashboardView.PermID}).Error)
	require.NoError(t, db.Create(&permission.SysRolePerm{RoleID: testRole.RoleID, PermID: datasetView.PermID}).Error)
	require.NoError(t, resourceSvc.RegisterResource(101, "dashboard-101", permission.ResourceTypeDashboard, nil))
	require.NoError(t, resourceSvc.ReplaceResourcePermissions(101, permission.ResourceTypeDashboard, []int64{dashboardView.PermID}))

	r := gin.New()
	api := r.Group("/api")
	RegisterPermissionCompatRoutes(api, h)

	queryReq := httptest.NewRequest(http.MethodPost, "/api/auth/busiTargetPermission", strings.NewReader(`{"id":101,"type":1,"flag":"dashboard"}`))
	queryReq.Header.Set("Content-Type", "application/json")
	queryResp := httptest.NewRecorder()
	r.ServeHTTP(queryResp, queryReq)
	require.Equal(t, http.StatusOK, queryResp.Code)

	var queryBody struct {
		Code string                           `json:"code"`
		Data []*permission.ResourceUserPermVO `json:"data"`
	}
	require.NoError(t, json.Unmarshal(queryResp.Body.Bytes(), &queryBody))
	assert.Equal(t, "000000", queryBody.Code)
	if assert.Len(t, queryBody.Data, 1) {
		assert.Equal(t, "dashboard:view", queryBody.Data[0].PermKey)
		assert.Equal(t, testRole.RoleID, queryBody.Data[0].SourceID)
	}

	saveReq := httptest.NewRequest(http.MethodPost, "/api/auth/saveBusiTargetPer", strings.NewReader(fmt.Sprintf(`{"id":101,"type":1,"flag":"dashboard","targetPerms":[{"targetType":"role","targetId":%d,"permIds":[%d]}]}`, testRole.RoleID, dashboardEdit.PermID)))
	saveReq.Header.Set("Content-Type", "application/json")
	saveResp := httptest.NewRecorder()
	r.ServeHTTP(saveResp, saveReq)
	require.Equal(t, http.StatusOK, saveResp.Code)

	var saveBody map[string]interface{}
	require.NoError(t, json.Unmarshal(saveResp.Body.Bytes(), &saveBody))
	assert.Equal(t, "000000", saveBody["code"])

	queryAfterSaveReq := httptest.NewRequest(http.MethodPost, "/api/auth/busiTargetPermission", strings.NewReader(`{"id":101,"type":1,"flag":"dashboard"}`))
	queryAfterSaveReq.Header.Set("Content-Type", "application/json")
	queryAfterSaveResp := httptest.NewRecorder()
	r.ServeHTTP(queryAfterSaveResp, queryAfterSaveReq)
	require.Equal(t, http.StatusOK, queryAfterSaveResp.Code)
	require.NoError(t, json.Unmarshal(queryAfterSaveResp.Body.Bytes(), &queryBody))
	assert.Equal(t, "000000", queryBody.Code)
	if assert.Len(t, queryBody.Data, 1) {
		assert.Equal(t, "dashboard:edit", queryBody.Data[0].PermKey)
	}

	persistedResourcePerms, exists, err := resourceRepo.GetResourcePermissionIDs(101, permission.ResourceTypeDashboard)
	require.NoError(t, err)
	assert.True(t, exists)
	assert.ElementsMatch(t, []int64{dashboardEdit.PermID}, persistedResourcePerms)

	var persistedRolePerms []permission.SysRolePerm
	require.NoError(t, db.Where("role_id = ?", testRole.RoleID).Find(&persistedRolePerms).Error)
	assert.Len(t, persistedRolePerms, 2)

	persistedPermIDs := make([]int64, 0, len(persistedRolePerms))
	for _, item := range persistedRolePerms {
		persistedPermIDs = append(persistedPermIDs, item.PermID)
	}
	assert.ElementsMatch(t, []int64{dashboardEdit.PermID, datasetView.PermID}, persistedPermIDs)
}

func TestPermissionCompatHandlerIntegration_UserPerspectiveWithResourceIDUsesGovernedResourceState(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testCases := []struct {
		name         string
		resourceType string
		resourceID   int64
		permKey      string
	}{
		{name: "datasource", resourceType: permission.ResourceTypeDatasource, resourceID: 201, permKey: "datasource:view"},
		{name: "dataset", resourceType: permission.ResourceTypeDataset, resourceID: 202, permKey: "dataset:view"},
		{name: "dashboard", resourceType: permission.ResourceTypeDashboard, resourceID: 203, permKey: "dashboard:view"},
		{name: "screen", resourceType: permission.ResourceTypeScreen, resourceID: 204, permKey: "screen:view"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db := openPermissionCompatHandlerTestDB(t)
			roleRepo := repository.NewRoleRepository(db)
			resourceRepo := repository.NewResourcePermissionRepository(db)
			resourceSvc := service.NewResourcePermissionService(resourceRepo, nil)
			h := NewPermissionCompatHandler(nil, nil, nil, resourceSvc)

			testRole := &role.SysRole{RoleName: tc.name + " perspective role", RoleCode: tc.name + "_perspective_role", Status: 1}
			require.NoError(t, roleRepo.Create(testRole))

			testUser := &user.SysUser{Username: tc.name + "_perspective_user", Password: "irrelevant", NickName: tc.name + " perspective user", Status: 1, DelFlag: 0}
			require.NoError(t, db.Create(testUser).Error)
			require.NoError(t, db.Create(&user.SysUserRole{UserID: testUser.UserID, RoleID: testRole.RoleID, OrgID: 1}).Error)

			viewPerm := &permission.SysPerm{PermName: tc.name + "查看", PermKey: tc.permKey, PermType: permission.PermTypeData, Status: 1, DelFlag: 0}
			require.NoError(t, db.Create(viewPerm).Error)
			require.NoError(t, db.Create(&permission.SysRolePerm{RoleID: testRole.RoleID, PermID: viewPerm.PermID}).Error)
			require.NoError(t, resourceSvc.RegisterResource(tc.resourceID, tc.name+"-resource", tc.resourceType, nil))
			require.NoError(t, resourceSvc.ReplaceResourcePermissions(tc.resourceID, tc.resourceType, []int64{viewPerm.PermID}))

			r := gin.New()
			api := r.Group("/api")
			RegisterPermissionCompatRoutes(api, h)

			resourceReq := httptest.NewRequest(http.MethodPost, "/api/auth/busiTargetPermission", strings.NewReader(fmt.Sprintf(`{"id":%d,"type":1,"flag":"%s"}`, tc.resourceID, tc.resourceType)))
			resourceReq.Header.Set("Content-Type", "application/json")
			resourceResp := httptest.NewRecorder()
			r.ServeHTTP(resourceResp, resourceReq)
			require.Equal(t, http.StatusOK, resourceResp.Code)

			var resourceBody struct {
				Code string                           `json:"code"`
				Data []*permission.ResourceUserPermVO `json:"data"`
			}
			require.NoError(t, json.Unmarshal(resourceResp.Body.Bytes(), &resourceBody))
			assert.Equal(t, "000000", resourceBody.Code)
			if assert.Len(t, resourceBody.Data, 1) {
				assert.Equal(t, tc.permKey, resourceBody.Data[0].PermKey)
				assert.Equal(t, testUser.UserID, resourceBody.Data[0].UserID)
			}

			userReq := httptest.NewRequest(http.MethodPost, "/api/auth/userPerspective", strings.NewReader(fmt.Sprintf(`{"userId":%d,"resourceId":%d,"resourceType":"%s"}`, testUser.UserID, tc.resourceID, tc.resourceType)))
			userReq.Header.Set("Content-Type", "application/json")
			userResp := httptest.NewRecorder()
			r.ServeHTTP(userResp, userReq)
			require.Equal(t, http.StatusOK, userResp.Code)

			var userBody struct {
				Code string                           `json:"code"`
				Data []*permission.UserResourcePermVO `json:"data"`
			}
			require.NoError(t, json.Unmarshal(userResp.Body.Bytes(), &userBody))
			assert.Equal(t, "000000", userBody.Code)
			if assert.Len(t, userBody.Data, 1) {
				assert.Equal(t, tc.resourceID, userBody.Data[0].ResourceID)
				assert.Equal(t, tc.resourceType, userBody.Data[0].ResourceType)
				assert.Equal(t, tc.permKey, userBody.Data[0].PermKey)
				assert.Equal(t, "role", userBody.Data[0].SourceType)
			}
		})
	}
}

func TestPermissionCompatHandlerIntegration_UserPerspectiveReturnsFilteredEffectivePerms(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := openPermissionCompatHandlerTestDB(t)
	roleRepo := repository.NewRoleRepository(db)
	resourceRepo := repository.NewResourcePermissionRepository(db)
	resourceSvc := service.NewResourcePermissionService(resourceRepo, nil)
	h := NewPermissionCompatHandler(nil, nil, nil, resourceSvc)

	testRole := &role.SysRole{
		RoleName: "Permission User Perspective Role",
		RoleCode: "permission_user_perspective_role",
		Status:   1,
	}
	require.NoError(t, roleRepo.Create(testRole))

	testUser := &user.SysUser{
		Username: "user_perspective_tester",
		Password: "irrelevant",
		NickName: "User Perspective Tester",
		Status:   1,
		DelFlag:  0,
	}
	require.NoError(t, db.Create(testUser).Error)
	require.NoError(t, db.Create(&user.SysUserRole{UserID: testUser.UserID, RoleID: testRole.RoleID, OrgID: 1}).Error)

	dashboardView := &permission.SysPerm{PermName: "仪表板查看", PermKey: "dashboard:view", PermType: permission.PermTypeData, Status: 1, DelFlag: 0}
	datasetView := &permission.SysPerm{PermName: "数据集查看", PermKey: "dataset:view", PermType: permission.PermTypeData, Status: 1, DelFlag: 0}
	require.NoError(t, db.Create(dashboardView).Error)
	require.NoError(t, db.Create(datasetView).Error)
	require.NoError(t, db.Create(&permission.SysRolePerm{RoleID: testRole.RoleID, PermID: dashboardView.PermID}).Error)
	require.NoError(t, db.Create(&permission.SysRolePerm{RoleID: testRole.RoleID, PermID: datasetView.PermID}).Error)
	require.NoError(t, db.Create(&permission.SysUserPerm{UserID: testUser.UserID, OrgID: int64PtrPermissionCompat(1), PermID: dashboardView.PermID, Status: 1, DelFlag: 0}).Error)

	r := gin.New()
	api := r.Group("/api")
	RegisterPermissionCompatRoutes(api, h)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/userPerspective", strings.NewReader(fmt.Sprintf(`{"userId":%d,"resourceType":"dashboard"}`, testUser.UserID)))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)
	require.Equal(t, http.StatusOK, resp.Code)

	var body struct {
		Code string                           `json:"code"`
		Data []*permission.UserResourcePermVO `json:"data"`
	}
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &body))
	assert.Equal(t, "000000", body.Code)
	if assert.Len(t, body.Data, 2) {
		assert.Equal(t, "dashboard", body.Data[0].ResourceType)
		for _, item := range body.Data {
			assert.Equal(t, "dashboard", item.ResourceType)
			assert.Equal(t, "dashboard:view", item.PermKey)
		}
	}
}

func TestPermissionCompatHandlerIntegration_SaveBusiTargetPerPreservesDirectEffectivePerms(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := openPermissionCompatHandlerTestDB(t)
	roleRepo := repository.NewRoleRepository(db)
	resourceRepo := repository.NewResourcePermissionRepository(db)
	resourceSvc := service.NewResourcePermissionService(resourceRepo, nil)
	h := NewPermissionCompatHandler(nil, nil, nil, resourceSvc)

	testRole := &role.SysRole{RoleName: "Direct Preserve Role", RoleCode: "direct_preserve_role", Status: 1}
	require.NoError(t, roleRepo.Create(testRole))
	testUser := &user.SysUser{Username: "direct_keeper", Password: "irrelevant", NickName: "Direct Keeper", Status: 1, DelFlag: 0}
	require.NoError(t, db.Create(testUser).Error)
	require.NoError(t, db.Create(&user.SysUserRole{UserID: testUser.UserID, RoleID: testRole.RoleID, OrgID: 1}).Error)

	dashboardView := &permission.SysPerm{PermName: "仪表板查看", PermKey: "dashboard:view", PermType: permission.PermTypeData, Status: 1, DelFlag: 0}
	dashboardEdit := &permission.SysPerm{PermName: "仪表板编辑", PermKey: "dashboard:edit", PermType: permission.PermTypeData, Status: 1, DelFlag: 0}
	require.NoError(t, db.Create(dashboardView).Error)
	require.NoError(t, db.Create(dashboardEdit).Error)
	require.NoError(t, db.Create(&permission.SysRolePerm{RoleID: testRole.RoleID, PermID: dashboardView.PermID}).Error)
	require.NoError(t, db.Create(&permission.SysUserPerm{UserID: testUser.UserID, OrgID: int64PtrPermissionCompat(1), PermID: dashboardView.PermID, Status: 1, DelFlag: 0}).Error)
	require.NoError(t, resourceSvc.RegisterResource(101, "dashboard-101", permission.ResourceTypeDashboard, nil))
	require.NoError(t, resourceSvc.ReplaceResourcePermissions(101, permission.ResourceTypeDashboard, []int64{dashboardView.PermID}))

	r := gin.New()
	api := r.Group("/api")
	RegisterPermissionCompatRoutes(api, h)

	saveReq := httptest.NewRequest(http.MethodPost, "/api/auth/saveBusiTargetPer", strings.NewReader(fmt.Sprintf(`{"id":101,"type":1,"flag":"dashboard","targetPerms":[{"targetType":"role","targetId":%d,"permIds":[%d,%d]}]}`, testRole.RoleID, dashboardView.PermID, dashboardEdit.PermID)))
	saveReq.Header.Set("Content-Type", "application/json")
	saveResp := httptest.NewRecorder()
	r.ServeHTTP(saveResp, saveReq)
	require.Equal(t, http.StatusOK, saveResp.Code)

	queryReq := httptest.NewRequest(http.MethodPost, "/api/auth/busiTargetPermission", strings.NewReader(`{"id":101,"type":1,"flag":"dashboard"}`))
	queryReq.Header.Set("Content-Type", "application/json")
	queryResp := httptest.NewRecorder()
	r.ServeHTTP(queryResp, queryReq)
	require.Equal(t, http.StatusOK, queryResp.Code)

	var queryBody struct {
		Code string                           `json:"code"`
		Data []*permission.ResourceUserPermVO `json:"data"`
	}
	require.NoError(t, json.Unmarshal(queryResp.Body.Bytes(), &queryBody))
	assert.Equal(t, "000000", queryBody.Code)
	if assert.Len(t, queryBody.Data, 3) {
		permKeys := make([]string, 0, len(queryBody.Data))
		sourceTypes := make([]string, 0, len(queryBody.Data))
		for _, item := range queryBody.Data {
			permKeys = append(permKeys, item.PermKey)
			sourceTypes = append(sourceTypes, item.SourceType)
		}
		assert.ElementsMatch(t, []string{"dashboard:view", "dashboard:view", "dashboard:edit"}, permKeys)
		assert.ElementsMatch(t, []string{"direct", "role", "role"}, sourceTypes)
	}
}

func TestPermissionCompatHandlerIntegration_SaveBusiTargetPerRejectsNonRoleTargets(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := openPermissionCompatHandlerTestDB(t)
	roleRepo := repository.NewRoleRepository(db)
	resourceRepo := repository.NewResourcePermissionRepository(db)
	resourceSvc := service.NewResourcePermissionService(resourceRepo, nil)
	h := NewPermissionCompatHandler(nil, nil, nil, resourceSvc)

	testRole := &role.SysRole{
		RoleName: "Permission Reject Target Role",
		RoleCode: "permission_reject_target_role",
		Status:   1,
	}
	require.NoError(t, roleRepo.Create(testRole))

	dashboardView := &permission.SysPerm{PermName: "仪表板查看", PermKey: "dashboard:view", PermType: permission.PermTypeData, Status: 1, DelFlag: 0}
	require.NoError(t, db.Create(dashboardView).Error)
	require.NoError(t, db.Create(&permission.SysRolePerm{RoleID: testRole.RoleID, PermID: dashboardView.PermID}).Error)

	r := gin.New()
	api := r.Group("/api")
	RegisterPermissionCompatRoutes(api, h)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/saveBusiTargetPer", strings.NewReader(fmt.Sprintf(`{"id":101,"type":1,"flag":"dashboard","targetPerms":[{"targetType":"direct","targetId":%d,"permIds":[%d]}]}`, testRole.RoleID, dashboardView.PermID)))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)
	require.Equal(t, http.StatusOK, resp.Code)

	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &body))
	assert.Equal(t, "500000", body["code"])
	assert.Contains(t, body["msg"], "only role targets are supported")

	var persistedRolePerms []permission.SysRolePerm
	require.NoError(t, db.Where("role_id = ?", testRole.RoleID).Find(&persistedRolePerms).Error)
	assert.Len(t, persistedRolePerms, 1)
	assert.Equal(t, dashboardView.PermID, persistedRolePerms[0].PermID)
}

func int64PtrPermissionCompat(value int64) *int64 {
	return &value
}
