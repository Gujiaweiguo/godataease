package handler

import (
	"dataease/backend/internal/domain/menu"
	"dataease/backend/internal/domain/permission"
	"dataease/backend/internal/domain/role"
	"dataease/backend/internal/repository"
	"dataease/backend/internal/service"
	"encoding/json"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type mockTargetResourcePermRepo struct {
	resourceUsers         []*permission.ResourceUserPermVO
	userResources         []*permission.UserResourcePermVO
	rolePerms             map[int64][]int64
	permByID              map[int64]*permission.SysPerm
	userRoles             []userRoleEdge
	replacedResourcePerms []int64
}

type userRoleEdge struct {
	userID   int64
	username string
	nickName string
	roleID   int64
	roleName string
}

func (m *mockTargetResourcePermRepo) GetPermByID(permID int64) (*permission.SysPerm, error) {
	if m.permByID == nil {
		return nil, nil
	}
	return m.permByID[permID], nil
}
func (m *mockTargetResourcePermRepo) GetPermByKey(permKey string) (*permission.SysPerm, error) {
	return nil, nil
}
func (m *mockTargetResourcePermRepo) ListPerms(permType string, page, size int) ([]*permission.SysPerm, int64, error) {
	return nil, 0, nil
}
func (m *mockTargetResourcePermRepo) CreatePerm(perm *permission.SysPerm) error { return nil }
func (m *mockTargetResourcePermRepo) UpdatePerm(perm *permission.SysPerm) error { return nil }
func (m *mockTargetResourcePermRepo) DeletePerm(permID int64) error             { return nil }
func (m *mockTargetResourcePermRepo) GetUserPerms(userID int64) ([]int64, error) {
	return nil, nil
}
func (m *mockTargetResourcePermRepo) GetRolePerms(roleID int64) ([]int64, error) {
	if m.rolePerms == nil {
		return nil, nil
	}
	return append([]int64{}, m.rolePerms[roleID]...), nil
}
func (m *mockTargetResourcePermRepo) GetUserRoleIDs(userID int64) ([]int64, error) {
	return nil, nil
}
func (m *mockTargetResourcePermRepo) CheckUserPermission(userID, permID int64) (bool, error) {
	return false, nil
}
func (m *mockTargetResourcePermRepo) CheckRolePermission(roleID, permID int64) (bool, error) {
	return false, nil
}
func (m *mockTargetResourcePermRepo) GrantPermToUser(userID, permID int64, createBy string) error {
	return nil
}
func (m *mockTargetResourcePermRepo) RevokePermFromUser(userID, permID int64) error { return nil }
func (m *mockTargetResourcePermRepo) GetUserResources(userID int64, resourceType string) ([]*permission.UserResourcePermVO, error) {
	if m.userResources == nil {
		return []*permission.UserResourcePermVO{}, nil
	}
	return m.userResources, nil
}
func (m *mockTargetResourcePermRepo) GetResourceUsers(resourceID int64, resourceType string) ([]*permission.ResourceUserPermVO, error) {
	if m.resourceUsers == nil {
		return []*permission.ResourceUserPermVO{}, nil
	}
	return m.resourceUsers, nil
}
func (m *mockTargetResourcePermRepo) ApplyGroupPermissions(groupID, resourceID int64, resourceType string) error {
	return nil
}
func (m *mockTargetResourcePermRepo) RegisterResource(resourceID int64, resourceName, resourceType string, parentID *int64) error {
	return nil
}
func (m *mockTargetResourcePermRepo) ReplaceResourcePermissions(resourceID int64, resourceType string, permIDs []int64) error {
	m.replacedResourcePerms = append([]int64{}, permIDs...)
	return nil
}
func (m *mockTargetResourcePermRepo) GetResourcePermissionIDs(resourceID int64, resourceType string) ([]int64, bool, error) {
	if len(m.replacedResourcePerms) == 0 {
		return nil, false, nil
	}
	return append([]int64{}, m.replacedResourcePerms...), true, nil
}
func (m *mockTargetResourcePermRepo) CheckPermissionConsistency() (*permission.PermissionConsistencyResult, error) {
	return &permission.PermissionConsistencyResult{Consistent: true}, nil
}

func TestPermissionCompatHandler_BusiTargetPermissionReturnsResourcePerspective(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &mockTargetResourcePermRepo{
		resourceUsers: []*permission.ResourceUserPermVO{{
			UserID:     2001,
			Username:   "tester",
			NickName:   "Tester",
			PermKey:    "resource:view",
			PermName:   "查看",
			SourceType: "role",
			SourceID:   10,
			SourceName: "普通角色",
		}},
	}
	h := &PermissionCompatHandler{
		resourcePermService: service.NewResourcePermissionService(repo, nil),
	}

	r := gin.New()
	api := r.Group("/api")
	RegisterPermissionCompatRoutes(api, h)

	req := httptest.NewRequest("POST", "/api/auth/busiTargetPermission", strings.NewReader(`{"id":101,"type":1,"flag":"dashboard"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp struct {
		Code string                           `json:"code"`
		Data []*permission.ResourceUserPermVO `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if resp.Code != "000000" {
		t.Fatalf("expected code 000000, got %s", resp.Code)
	}
	if len(resp.Data) != 1 || resp.Data[0].UserID != 2001 {
		t.Fatalf("expected resource perspective payload, got %#v", resp.Data)
	}
}

func TestPermissionCompatHandler_UserPerspectiveReturnsUserPerspectivePayload(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &mockTargetResourcePermRepo{
		userResources: []*permission.UserResourcePermVO{{
			PermKey:      "dashboard:view",
			PermName:     "查看",
			SourceType:   "role",
			SourceID:     10,
			SourceName:   "普通角色",
			ResourceType: "dashboard",
		}},
	}
	h := &PermissionCompatHandler{
		resourcePermService: service.NewResourcePermissionService(repo, nil),
	}

	r := gin.New()
	api := r.Group("/api")
	RegisterPermissionCompatRoutes(api, h)

	req := httptest.NewRequest("POST", "/api/auth/userPerspective", strings.NewReader(`{"userId":2001,"resourceType":"dashboard"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp struct {
		Code string                           `json:"code"`
		Data []*permission.UserResourcePermVO `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if resp.Code != "000000" {
		t.Fatalf("expected code 000000, got %s", resp.Code)
	}
	if len(resp.Data) != 1 || resp.Data[0].PermKey != "dashboard:view" {
		t.Fatalf("expected user perspective payload, got %#v", resp.Data)
	}
}

func TestPermissionCompatHandler_UserPerspectiveFiltersByResourceWhenProvided(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &mockTargetResourcePermRepo{
		resourceUsers: []*permission.ResourceUserPermVO{
			{UserID: 2001, Username: "tester", PermKey: "dashboard:view", PermName: "查看", SourceType: "role", SourceID: 10, SourceName: "普通角色"},
			{UserID: 2002, Username: "other", PermKey: "dashboard:edit", PermName: "编辑", SourceType: "role", SourceID: 11, SourceName: "其他角色"},
		},
	}
	h := &PermissionCompatHandler{resourcePermService: service.NewResourcePermissionService(repo, nil)}

	r := gin.New()
	api := r.Group("/api")
	RegisterPermissionCompatRoutes(api, h)

	req := httptest.NewRequest("POST", "/api/auth/userPerspective", strings.NewReader(`{"userId":2001,"resourceId":101,"resourceType":"dashboard"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp struct {
		Code string                           `json:"code"`
		Data []*permission.UserResourcePermVO `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if len(resp.Data) != 1 || resp.Data[0].SourceID != 10 {
		t.Fatalf("expected single filtered user perspective row, got %#v", resp.Data)
	}
}

func TestPermissionCompatHandler_SaveBusiPerEchoesInResourcePerspective(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &mockTargetResourcePermRepo{
		rolePerms: map[int64][]int64{},
		permByID: map[int64]*permission.SysPerm{
			3: {PermID: 3, PermKey: "dashboard:view", PermName: "仪表板查看"},
			4: {PermID: 4, PermKey: "dashboard:edit", PermName: "仪表板编辑"},
		},
		userRoles: []userRoleEdge{{
			userID:   2001,
			username: "tester",
			nickName: "Tester",
			roleID:   10,
			roleName: "普通角色",
		}},
	}
	h := &PermissionCompatHandler{
		resourcePermService: service.NewResourcePermissionService(repo, nil),
	}

	r := gin.New()
	api := r.Group("/api")
	RegisterPermissionCompatRoutes(api, h)

	saveReq := httptest.NewRequest("POST", "/api/auth/saveBusiPer", strings.NewReader(`{"roleId":10,"permIds":[3,4]}`))
	saveReq.Header.Set("Content-Type", "application/json")
	saveResp := httptest.NewRecorder()
	r.ServeHTTP(saveResp, saveReq)
	if saveResp.Code != 200 {
		t.Fatalf("expected save status 200, got %d", saveResp.Code)
	}

	queryReq := httptest.NewRequest("POST", "/api/auth/busiTargetPermission", strings.NewReader(`{"id":101,"type":1,"flag":"dashboard"}`))
	queryReq.Header.Set("Content-Type", "application/json")
	queryResp := httptest.NewRecorder()
	r.ServeHTTP(queryResp, queryReq)
	if queryResp.Code != 200 {
		t.Fatalf("expected query status 200, got %d", queryResp.Code)
	}

	var resp struct {
		Code string                           `json:"code"`
		Data []*permission.ResourceUserPermVO `json:"data"`
	}
	if err := json.Unmarshal(queryResp.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal query response failed: %v", err)
	}
	if resp.Code != "000000" {
		t.Fatalf("expected code 000000, got %s", resp.Code)
	}
	if len(resp.Data) != 2 {
		t.Fatalf("expected 2 resource-perspective entries, got %#v", resp.Data)
	}
	if resp.Data[0].SourceType != "role" || resp.Data[0].UserID != 2001 {
		t.Fatalf("expected role-derived user entry, got %#v", resp.Data[0])
	}
}

func (m *mockTargetResourcePermRepo) GrantPermToRole(roleID, permID int64) error {
	if m.rolePerms == nil {
		m.rolePerms = map[int64][]int64{}
	}
	for _, existing := range m.rolePerms[roleID] {
		if existing == permID {
			return nil
		}
	}
	m.rolePerms[roleID] = append(m.rolePerms[roleID], permID)
	m.resourceUsers = m.buildResourceUsers("dashboard")
	return nil
}

func (m *mockTargetResourcePermRepo) RevokePermFromRole(roleID, permID int64) error {
	if m.rolePerms == nil {
		return nil
	}
	filtered := make([]int64, 0, len(m.rolePerms[roleID]))
	for _, existing := range m.rolePerms[roleID] {
		if existing != permID {
			filtered = append(filtered, existing)
		}
	}
	m.rolePerms[roleID] = filtered
	m.resourceUsers = m.buildResourceUsers("dashboard")
	return nil
}

func (m *mockTargetResourcePermRepo) buildResourceUsers(resourceType string) []*permission.ResourceUserPermVO {
	prefix := resourceType + ":"
	results := make([]*permission.ResourceUserPermVO, 0)
	for _, edge := range m.userRoles {
		for _, permID := range m.rolePerms[edge.roleID] {
			perm := m.permByID[permID]
			if perm == nil || !strings.HasPrefix(perm.PermKey, prefix) {
				continue
			}
			results = append(results, &permission.ResourceUserPermVO{
				UserID:     edge.userID,
				Username:   edge.username,
				NickName:   edge.nickName,
				PermKey:    perm.PermKey,
				PermName:   perm.PermName,
				SourceType: "role",
				SourceID:   edge.roleID,
				SourceName: edge.roleName,
			})
		}
	}
	return results
}

func TestPermissionCompatHandler_MenuTargetPermissionAndSaveRoundTrip(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&role.SysRole{}, &role.RoleMenu{}, &menu.CoreMenu{}))
	roleRepo := repository.NewRoleRepository(db)
	menuRepo := repository.NewMenuRepository(db)
	roleMenuRepo := repository.NewRoleMenuRepository(db)
	menuSvc := service.NewMenuService(menuRepo)
	roleMenuSvc := service.NewRoleMenuService(roleMenuRepo, roleRepo, menuRepo)
	h := NewPermissionCompatHandler(menuSvc, nil, roleMenuSvc, nil)

	testRole := &role.SysRole{RoleName: "menu-target-role", RoleCode: "menu-target-role", Status: 1}
	require.NoError(t, roleRepo.Create(testRole))
	menuA := &menu.CoreMenu{Name: "User", Path: "/system/user", Type: 2, Auth: true, MenuSort: 1}
	menuB := &menu.CoreMenu{Name: "Permission", Path: "/system/permission", Type: 2, Auth: true, MenuSort: 2}
	require.NoError(t, menuRepo.Create(menuA))
	require.NoError(t, menuRepo.Create(menuB))
	require.NoError(t, roleMenuRepo.SaveRoleMenus(testRole.RoleID, []int64{menuA.ID}))

	r := gin.New()
	api := r.Group("/api")
	RegisterPermissionCompatRoutes(api, h)

	queryReq := httptest.NewRequest("POST", "/api/auth/menuTargetPermission", strings.NewReader(`{"roleId":`+strconv.FormatInt(testRole.RoleID, 10)+`}`))
	queryReq.Header.Set("Content-Type", "application/json")
	queryResp := httptest.NewRecorder()
	r.ServeHTTP(queryResp, queryReq)
	require.Equal(t, 200, queryResp.Code)
	var queryBody struct {
		Code string `json:"code"`
		Data struct {
			MenuIDs []int64 `json:"menuIds"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(queryResp.Body.Bytes(), &queryBody))
	require.Equal(t, "000000", queryBody.Code)
	require.Equal(t, []int64{menuA.ID}, queryBody.Data.MenuIDs)

	saveReq := httptest.NewRequest("POST", "/api/auth/saveMenuTargetPer", strings.NewReader(`{"roleId":`+strconv.FormatInt(testRole.RoleID, 10)+`,"targetPerms":[{"targetType":"role","targetId":`+strconv.FormatInt(testRole.RoleID, 10)+`,"permIds":[`+strconv.FormatInt(menuA.ID, 10)+`,`+strconv.FormatInt(menuB.ID, 10)+`]}]}`))
	saveReq.Header.Set("Content-Type", "application/json")
	saveResp := httptest.NewRecorder()
	r.ServeHTTP(saveResp, saveReq)
	require.Equal(t, 200, saveResp.Code)
	var saveBody map[string]interface{}
	require.NoError(t, json.Unmarshal(saveResp.Body.Bytes(), &saveBody))
	require.Equal(t, "000000", saveBody["code"])

	queryAfterReq := httptest.NewRequest("POST", "/api/auth/menuTargetPermission", strings.NewReader(`{"roleId":`+strconv.FormatInt(testRole.RoleID, 10)+`}`))
	queryAfterReq.Header.Set("Content-Type", "application/json")
	queryAfterResp := httptest.NewRecorder()
	r.ServeHTTP(queryAfterResp, queryAfterReq)
	require.Equal(t, 200, queryAfterResp.Code)
	require.NoError(t, json.Unmarshal(queryAfterResp.Body.Bytes(), &queryBody))
	require.ElementsMatch(t, []int64{menuA.ID, menuB.ID}, queryBody.Data.MenuIDs)
}

func TestPermissionCompatHandler_SaveBusiTargetPerUpdatesOnlyRolePermsForResourceType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &mockTargetResourcePermRepo{
		rolePerms: map[int64][]int64{
			10: {3, 5},
		},
		permByID: map[int64]*permission.SysPerm{
			3: {PermID: 3, PermKey: "dashboard:view", PermName: "仪表板查看"},
			4: {PermID: 4, PermKey: "dashboard:edit", PermName: "仪表板编辑"},
			5: {PermID: 5, PermKey: "dataset:view", PermName: "数据集查看"},
		},
		userRoles: []userRoleEdge{{
			userID:   2001,
			username: "tester",
			nickName: "Tester",
			roleID:   10,
			roleName: "普通角色",
		}},
	}
	repo.resourceUsers = repo.buildResourceUsers("dashboard")
	h := &PermissionCompatHandler{
		resourcePermService: service.NewResourcePermissionService(repo, nil),
	}

	r := gin.New()
	api := r.Group("/api")
	RegisterPermissionCompatRoutes(api, h)

	saveReq := httptest.NewRequest("POST", "/api/auth/saveBusiTargetPer", strings.NewReader(`{"id":101,"type":1,"flag":"dashboard","targetPerms":[{"targetType":"role","targetId":10,"permIds":[4]}]}`))
	saveReq.Header.Set("Content-Type", "application/json")
	saveResp := httptest.NewRecorder()
	r.ServeHTTP(saveResp, saveReq)
	if saveResp.Code != 200 {
		t.Fatalf("expected save status 200, got %d", saveResp.Code)
	}

	var saveBody map[string]interface{}
	if err := json.Unmarshal(saveResp.Body.Bytes(), &saveBody); err != nil {
		t.Fatalf("unmarshal save response failed: %v", err)
	}
	if saveBody["code"] != "000000" {
		t.Fatalf("expected save code 000000, got %#v", saveBody["code"])
	}

	rolePerms := repo.rolePerms[10]
	if len(rolePerms) != 2 {
		t.Fatalf("expected dataset permission to stay and dashboard permission to be replaced, got %#v", rolePerms)
	}

	hasDatasetView := false
	hasDashboardEdit := false
	for _, permID := range rolePerms {
		if permID == 5 {
			hasDatasetView = true
		}
		if permID == 4 {
			hasDashboardEdit = true
		}
	}
	if !hasDatasetView || !hasDashboardEdit {
		t.Fatalf("expected dataset permission to remain and dashboard permission to switch, got %#v", rolePerms)
	}

	queryReq := httptest.NewRequest("POST", "/api/auth/busiTargetPermission", strings.NewReader(`{"id":101,"type":1,"flag":"dashboard"}`))
	queryReq.Header.Set("Content-Type", "application/json")
	queryResp := httptest.NewRecorder()
	r.ServeHTTP(queryResp, queryReq)
	if queryResp.Code != 200 {
		t.Fatalf("expected query status 200, got %d", queryResp.Code)
	}

	var resp struct {
		Code string                           `json:"code"`
		Data []*permission.ResourceUserPermVO `json:"data"`
	}
	if err := json.Unmarshal(queryResp.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal query response failed: %v", err)
	}
	if resp.Code != "000000" {
		t.Fatalf("expected code 000000, got %s", resp.Code)
	}
	if len(resp.Data) != 1 || resp.Data[0].PermKey != "dashboard:edit" {
		t.Fatalf("expected dashboard role entry to refresh after save, got %#v", resp.Data)
	}
}

func TestPermissionCompatHandler_SaveBusiTargetPerRejectsNonRoleTargets(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &mockTargetResourcePermRepo{
		permByID: map[int64]*permission.SysPerm{
			3: {PermID: 3, PermKey: "dashboard:view", PermName: "仪表板查看"},
		},
	}
	h := &PermissionCompatHandler{
		resourcePermService: service.NewResourcePermissionService(repo, nil),
	}

	r := gin.New()
	api := r.Group("/api")
	RegisterPermissionCompatRoutes(api, h)

	req := httptest.NewRequest("POST", "/api/auth/saveBusiTargetPer", strings.NewReader(`{"id":101,"type":1,"flag":"dashboard","targetPerms":[{"targetType":"direct","targetId":2001,"permIds":[3]}]}`))
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
		t.Fatalf("expected code 500000, got %#v", resp["code"])
	}
}

func TestPermissionCompatHandler_SaveBusiTargetPerReplacesResourcePermissionUnion(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &mockTargetResourcePermRepo{
		rolePerms: map[int64][]int64{},
		permByID: map[int64]*permission.SysPerm{
			3: {PermID: 3, PermKey: "dashboard:view", PermName: "仪表板查看"},
			4: {PermID: 4, PermKey: "dashboard:edit", PermName: "仪表板编辑"},
			6: {PermID: 6, PermKey: "dashboard:manage", PermName: "仪表板管理"},
		},
	}
	h := &PermissionCompatHandler{resourcePermService: service.NewResourcePermissionService(repo, nil)}

	r := gin.New()
	api := r.Group("/api")
	RegisterPermissionCompatRoutes(api, h)

	req := httptest.NewRequest("POST", "/api/auth/saveBusiTargetPer", strings.NewReader(`{"id":101,"flag":"dashboard","targetPerms":[{"targetType":"role","targetId":10,"permIds":[3,4]},{"targetType":"role","targetId":11,"permIds":[4,6]}]}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected status 200, got %d body=%s", w.Code, w.Body.String())
	}
	if len(repo.replacedResourcePerms) != 3 {
		t.Fatalf("expected dashboard-only union to be persisted, got %#v", repo.replacedResourcePerms)
	}
	if !(containsPerm(repo.replacedResourcePerms, 3) && containsPerm(repo.replacedResourcePerms, 4) && containsPerm(repo.replacedResourcePerms, 6)) {
		t.Fatalf("expected resource perms [3 4 6], got %#v", repo.replacedResourcePerms)
	}
}

func containsPerm(items []int64, target int64) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func TestPermissionCompatHandler_BusiTargetPermissionRequiresIDAndFlag(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &PermissionCompatHandler{}

	r := gin.New()
	api := r.Group("/api")
	RegisterPermissionCompatRoutes(api, h)

	req := httptest.NewRequest("POST", "/api/auth/busiTargetPermission", strings.NewReader(`{"type":1}`))
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
		t.Fatalf("expected code 500000, got %#v", resp["code"])
	}
}
