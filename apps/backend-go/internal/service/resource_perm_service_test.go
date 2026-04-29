package service

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"dataease/backend/internal/domain/permission"
)

type mockResourcePermAdminChecker struct {
	adminUserIDs map[int64]bool
}

func (m *mockResourcePermAdminChecker) IsAdmin(userID int64) bool {
	return m.adminUserIDs[userID]
}

type mockResourcePermRepo struct {
	permByID         map[int64]*permission.SysPerm
	permByIDErr      error
	userPerms        map[int64][]int64
	userPermsErr     error
	rolePerms        map[int64][]int64
	rolePermsErr     error
	userRoles        map[int64][]int64
	userRoleErr      error
	permKeys         map[string]*permission.SysPerm
	userPermOk       map[int64]map[int64]bool
	rolePermOk       map[int64]map[int64]bool
	rolePermCheckErr error
	resourcePerms    map[string][]int64
	userResources    []*permission.UserResourcePermVO
	resourceUsers    []*permission.ResourceUserPermVO
	applyGroupErr    error
	resourcePermErr  error
	registerErr      error
	replaceErr       error
	userResourcesErr error
	resourceUsersErr error
	consistencyErr   error
	consistencyResult *permission.PermissionConsistencyResult
	applyGroupCalls  int
	registerCalls    int
	replaceCalls     int
	lastParentID     *int64
}

func newMockResourcePermRepo() *mockResourcePermRepo {
	return &mockResourcePermRepo{
		permByID:      make(map[int64]*permission.SysPerm),
		userPerms:     make(map[int64][]int64),
		rolePerms:     make(map[int64][]int64),
		userRoles:     make(map[int64][]int64),
		permKeys:      make(map[string]*permission.SysPerm),
		userPermOk:    make(map[int64]map[int64]bool),
		rolePermOk:    make(map[int64]map[int64]bool),
		resourcePerms: make(map[string][]int64),
	}
}

func (m *mockResourcePermRepo) GetPermByID(permID int64) (*permission.SysPerm, error) {
	if m.permByIDErr != nil {
		return nil, m.permByIDErr
	}
	if perm, ok := m.permByID[permID]; ok {
		return perm, nil
	}
	return nil, nil
}

func (m *mockResourcePermRepo) GetPermByKey(permKey string) (*permission.SysPerm, error) {
	if perm, ok := m.permKeys[permKey]; ok {
		return perm, nil
	}
	if _, suffix, found := strings.Cut(permKey, ":"); found {
		if perm, ok := m.permKeys[suffix]; ok {
			return perm, nil
		}
	}
	if strings.Contains(permKey, ":") {
		return nil, fmt.Errorf("permission not found")
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
	if m.userPermsErr != nil {
		return nil, m.userPermsErr
	}
	return m.userPerms[userID], nil
}

func (m *mockResourcePermRepo) GetRolePerms(roleID int64) ([]int64, error) {
	if m.rolePermsErr != nil {
		return nil, m.rolePermsErr
	}
	return m.rolePerms[roleID], nil
}

func (m *mockResourcePermRepo) GetUserRoleIDs(userID int64) ([]int64, error) {
	if m.userRoleErr != nil {
		return nil, m.userRoleErr
	}
	return m.userRoles[userID], nil
}

func (m *mockResourcePermRepo) CheckUserPermission(userID, permID int64) (bool, error) {
	if userPerms, ok := m.userPermOk[userID]; ok {
		return userPerms[permID], nil
	}
	return false, nil
}

func (m *mockResourcePermRepo) CheckRolePermission(roleID, permID int64) (bool, error) {
	if m.rolePermCheckErr != nil {
		return false, m.rolePermCheckErr
	}
	if rolePerms, ok := m.rolePermOk[roleID]; ok {
		return rolePerms[permID], nil
	}
	return false, nil
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

// ========== 双视角接口实现 ==========
func (m *mockResourcePermRepo) GetUserResources(userID int64, resourceType string) ([]*permission.UserResourcePermVO, error) {
	if m.userResourcesErr != nil {
		return nil, m.userResourcesErr
	}
	if m.userResources != nil {
		return m.userResources, nil
	}
	return []*permission.UserResourcePermVO{}, nil
}

func (m *mockResourcePermRepo) GetResourceUsers(resourceID int64, resourceType string) ([]*permission.ResourceUserPermVO, error) {
	if m.resourceUsersErr != nil {
		return nil, m.resourceUsersErr
	}
	if m.resourceUsers != nil {
		return m.resourceUsers, nil
	}
	return []*permission.ResourceUserPermVO{}, nil
}

func (m *mockResourcePermRepo) ApplyGroupPermissions(groupID, resourceID int64, resourceType string) error {
	m.applyGroupCalls++
	if m.applyGroupErr != nil {
		return m.applyGroupErr
	}
	return nil
}

func (m *mockResourcePermRepo) RegisterResource(resourceID int64, resourceName, resourceType string, parentID *int64) error {
	m.registerCalls++
	m.lastParentID = parentID
	if m.registerErr != nil {
		return m.registerErr
	}
	return nil
}

func (m *mockResourcePermRepo) InheritParentResourcePermissions(parentID, resourceID int64, resourceName, resourceType string) error {
	return nil
}

func (m *mockResourcePermRepo) ReplaceResourcePermissions(resourceID int64, resourceType string, permIDs []int64) error {
	m.replaceCalls++
	if m.replaceErr != nil {
		return m.replaceErr
	}
	m.resourcePerms[resourceTypeKey(resourceType, resourceID)] = append([]int64{}, permIDs...)
	return nil
}

func (m *mockResourcePermRepo) GetResourcePermissionIDs(resourceID int64, resourceType string) ([]int64, bool, error) {
	if m.resourcePermErr != nil {
		return nil, false, m.resourcePermErr
	}
	permIDs, ok := m.resourcePerms[resourceTypeKey(resourceType, resourceID)]
	if !ok {
		return nil, false, nil
	}
	return append([]int64{}, permIDs...), true, nil
}

func (m *mockResourcePermRepo) CheckPermissionConsistency() (*permission.PermissionConsistencyResult, error) {
	if m.consistencyErr != nil {
		return nil, m.consistencyErr
	}
	if m.consistencyResult != nil {
		return m.consistencyResult, nil
	}
	return &permission.PermissionConsistencyResult{Consistent: true}, nil
}

func resourceTypeKey(resourceType string, resourceID int64) string {
	return resourceType + ":" + strconv.FormatInt(resourceID, 10)
}

func TestCheckPermission_AdminBypass(t *testing.T) {
	mockRepo := newMockResourcePermRepo()
	mockChecker := &mockResourcePermAdminChecker{adminUserIDs: map[int64]bool{1: true}}
	svc := NewResourcePermissionService(mockRepo, mockChecker)

	result := svc.CheckPermission(1, permission.ResourceTypeDashboard, 100, permission.PermKeyView)
	if !result.HasPermission {
		t.Errorf("Admin should have permission, got %v", result.HasPermission)
	}
	if result.Reason != "admin" {
		t.Errorf("Expected reason 'admin', got %s", result.Reason)
	}
}

func TestCheckPermission_UserPermission(t *testing.T) {
	mockRepo := newMockResourcePermRepo()
	mockRepo.userPermOk[2] = map[int64]bool{1: true}
	mockRepo.permKeys[permission.PermKeyView] = &permission.SysPerm{PermID: 1, PermKey: permission.PermKeyView}
	mockChecker := &mockResourcePermAdminChecker{adminUserIDs: map[int64]bool{}}
	svc := NewResourcePermissionService(mockRepo, mockChecker)

	result := svc.CheckPermission(2, permission.ResourceTypeDashboard, 100, permission.PermKeyView)
	if !result.HasPermission {
		t.Errorf("User with direct permission should have access, got %v", result.HasPermission)
	}
	if result.Reason != "user_permission" {
		t.Errorf("Expected reason 'user_permission', got %s", result.Reason)
	}
}

func TestCheckPermission_RolePermission(t *testing.T) {
	mockRepo := newMockResourcePermRepo()
	mockRepo.userRoles[3] = []int64{10}
	mockRepo.rolePermOk[10] = map[int64]bool{1: true}
	mockRepo.permKeys[permission.PermKeyEdit] = &permission.SysPerm{PermID: 1, PermKey: permission.PermKeyEdit}
	mockChecker := &mockResourcePermAdminChecker{adminUserIDs: map[int64]bool{}}
	svc := NewResourcePermissionService(mockRepo, mockChecker)

	result := svc.CheckPermission(3, permission.ResourceTypeDataset, 200, permission.PermKeyEdit)
	if !result.HasPermission {
		t.Errorf("User with role permission should have access, got %v", result.HasPermission)
	}
	if result.Reason != "role_permission" {
		t.Errorf("Expected reason 'role_permission', got %s", result.Reason)
	}
}

func TestCheckPermission_NoPermission(t *testing.T) {
	mockRepo := newMockResourcePermRepo()
	mockChecker := &mockResourcePermAdminChecker{adminUserIDs: map[int64]bool{}}
	svc := NewResourcePermissionService(mockRepo, mockChecker)

	result := svc.CheckPermission(4, permission.ResourceTypeDashboard, 100, permission.PermKeyExport)
	if result.HasPermission {
		t.Errorf("User without permission should not have access, got %v", result.HasPermission)
	}
	if result.Reason != "no_roles" {
		t.Errorf("Expected reason 'no_roles', got %s", result.Reason)
	}
}

func TestCheckPermission_ResourceScopedPermissionDenied(t *testing.T) {
	mockRepo := newMockResourcePermRepo()
	mockRepo.userPermOk[9] = map[int64]bool{1: true}
	mockRepo.permKeys[permission.ResourceTypeDashboard+":"+permission.PermKeyView] = &permission.SysPerm{PermID: 1, PermKey: permission.ResourceTypeDashboard + ":" + permission.PermKeyView}
	mockRepo.resourcePerms[resourceTypeKey(permission.ResourceTypeDashboard, 700)] = []int64{2}
	mockChecker := &mockResourcePermAdminChecker{adminUserIDs: map[int64]bool{}}
	svc := NewResourcePermissionService(mockRepo, mockChecker)

	result := svc.CheckPermission(9, permission.ResourceTypeDashboard, 700, permission.PermKeyView)
	if result.HasPermission {
		t.Fatalf("expected governed resource permission denial, got allowed")
	}
	if result.Reason != "resource_permission_denied" {
		t.Fatalf("expected resource_permission_denied, got %s", result.Reason)
	}
}

func TestCheckPermission_ResourceScopedPermissionAllowed(t *testing.T) {
	mockRepo := newMockResourcePermRepo()
	mockRepo.userPermOk[10] = map[int64]bool{1: true}
	mockRepo.permKeys[permission.ResourceTypeDashboard+":"+permission.PermKeyView] = &permission.SysPerm{PermID: 1, PermKey: permission.ResourceTypeDashboard + ":" + permission.PermKeyView}
	mockRepo.resourcePerms[resourceTypeKey(permission.ResourceTypeDashboard, 701)] = []int64{1}
	mockChecker := &mockResourcePermAdminChecker{adminUserIDs: map[int64]bool{}}
	svc := NewResourcePermissionService(mockRepo, mockChecker)

	result := svc.CheckPermission(10, permission.ResourceTypeDashboard, 701, permission.PermKeyView)
	if !result.HasPermission {
		t.Fatalf("expected governed resource permission allowed, got denied: %s", result.Reason)
	}
	if result.Reason != "user_permission" {
		t.Fatalf("expected user_permission, got %s", result.Reason)
	}
}

func TestCheckViewPermission(t *testing.T) {
	mockRepo := newMockResourcePermRepo()
	mockRepo.userPermOk[5] = map[int64]bool{1: true}
	mockRepo.permKeys[permission.PermKeyView] = &permission.SysPerm{PermID: 1, PermKey: permission.PermKeyView}
	mockChecker := &mockResourcePermAdminChecker{adminUserIDs: map[int64]bool{}}
	svc := NewResourcePermissionService(mockRepo, mockChecker)

	if !svc.CheckViewPermission(5, permission.ResourceTypeDatasource, 300) {
		t.Error("CheckViewPermission should return true")
	}
}

func TestCheckEditPermission(t *testing.T) {
	mockRepo := newMockResourcePermRepo()
	mockRepo.userPermOk[6] = map[int64]bool{2: true}
	mockRepo.permKeys[permission.PermKeyEdit] = &permission.SysPerm{PermID: 2, PermKey: permission.PermKeyEdit}
	mockChecker := &mockResourcePermAdminChecker{adminUserIDs: map[int64]bool{}}
	svc := NewResourcePermissionService(mockRepo, mockChecker)

	if !svc.CheckEditPermission(6, permission.ResourceTypeDataset, 400) {
		t.Error("CheckEditPermission should return true")
	}
}

func TestCheckExportPermission(t *testing.T) {
	mockRepo := newMockResourcePermRepo()
	mockRepo.userPermOk[7] = map[int64]bool{3: true}
	mockRepo.permKeys[permission.PermKeyExport] = &permission.SysPerm{PermID: 3, PermKey: permission.PermKeyExport}
	mockChecker := &mockResourcePermAdminChecker{adminUserIDs: map[int64]bool{}}
	svc := NewResourcePermissionService(mockRepo, mockChecker)

	if !svc.CheckExportPermission(7, permission.ResourceTypeDashboard, 500) {
		t.Error("CheckExportPermission should return true")
	}
}

func TestCheckManagePermission(t *testing.T) {
	mockRepo := newMockResourcePermRepo()
	mockRepo.userPermOk[8] = map[int64]bool{4: true}
	mockRepo.permKeys[permission.PermKeyManage] = &permission.SysPerm{PermID: 4, PermKey: permission.PermKeyManage}
	mockChecker := &mockResourcePermAdminChecker{adminUserIDs: map[int64]bool{}}
	svc := NewResourcePermissionService(mockRepo, mockChecker)

	if !svc.CheckManagePermission(8, permission.ResourceTypeScreen, 600) {
		t.Error("CheckManagePermission should return true")
	}
}

func TestCheckPermissionConsistency_ConsistentState(t *testing.T) {
	mockRepo := newMockResourcePermRepo()
	expected := &permission.PermissionConsistencyResult{
		Consistent:      true,
		UserCount:       3,
		ResourceCount:   5,
		Inconsistencies: []*permission.PermissionInconsistencyVO{},
	}
	mockRepo.consistencyResult = expected
	svc := NewResourcePermissionService(mockRepo, &mockResourcePermAdminChecker{adminUserIDs: map[int64]bool{}})

	result, err := svc.CheckPermissionConsistency()
	if err != nil {
		t.Fatalf("CheckPermissionConsistency failed: %v", err)
	}
	if result != expected {
		t.Fatalf("expected delegated result pointer to pass through")
	}
	if !result.Consistent {
		t.Fatalf("expected consistent result")
	}
	if result.UserCount != 3 || result.ResourceCount != 5 {
		t.Fatalf("unexpected counts: users=%d resources=%d", result.UserCount, result.ResourceCount)
	}
	if len(result.Inconsistencies) != 0 {
		t.Fatalf("expected no inconsistencies, got %d", len(result.Inconsistencies))
	}
}

func TestCheckPermissionConsistency_DivergentState(t *testing.T) {
	mockRepo := newMockResourcePermRepo()
	expectedInconsistencies := []*permission.PermissionInconsistencyVO{{
		UserID:       42,
		ResourceID:   1001,
		ResourceType: permission.ResourceTypeDashboard,
		UserView:     "granted",
		ResourceView: "missing",
		Description:  "user 42 has dashboard:view in user view but resource view is missing",
	}}
	expected := &permission.PermissionConsistencyResult{
		Consistent:      false,
		UserCount:       8,
		ResourceCount:   2,
		Inconsistencies: expectedInconsistencies,
	}
	mockRepo.consistencyResult = expected
	svc := NewResourcePermissionService(mockRepo, &mockResourcePermAdminChecker{adminUserIDs: map[int64]bool{}})

	result, err := svc.CheckPermissionConsistency()
	if err != nil {
		t.Fatalf("CheckPermissionConsistency failed: %v", err)
	}
	if result != expected {
		t.Fatalf("expected delegated result pointer to pass through")
	}
	if result.Consistent {
		t.Fatalf("expected inconsistent result")
	}
	if len(result.Inconsistencies) != 1 {
		t.Fatalf("expected 1 inconsistency, got %d", len(result.Inconsistencies))
	}
	if result.Inconsistencies[0] != expectedInconsistencies[0] {
		t.Fatalf("expected inconsistency details to pass through")
	}
}

func TestCheckPermissionConsistency_EmptySystem(t *testing.T) {
	mockRepo := newMockResourcePermRepo()
	expected := &permission.PermissionConsistencyResult{
		Consistent:      true,
		UserCount:       0,
		ResourceCount:   0,
		Inconsistencies: []*permission.PermissionInconsistencyVO{},
	}
	mockRepo.consistencyResult = expected
	svc := NewResourcePermissionService(mockRepo, &mockResourcePermAdminChecker{adminUserIDs: map[int64]bool{}})

	result, err := svc.CheckPermissionConsistency()
	if err != nil {
		t.Fatalf("CheckPermissionConsistency failed: %v", err)
	}
	if result != expected {
		t.Fatalf("expected delegated result pointer to pass through")
	}
	if !result.Consistent || result.UserCount != 0 || result.ResourceCount != 0 {
		t.Fatalf("unexpected empty system result: %+v", result)
	}
	if len(result.Inconsistencies) != 0 {
		t.Fatalf("expected no inconsistencies, got %d", len(result.Inconsistencies))
	}
}

func TestCheckPermissionConsistency_SkipsLargeUserBase(t *testing.T) {
	mockRepo := newMockResourcePermRepo()
	expectedInconsistencies := []*permission.PermissionInconsistencyVO{{
		Description: "permission consistency check skipped: active user count 10001 exceeds limit 10000",
	}}
	expected := &permission.PermissionConsistencyResult{
		Consistent:      true,
		UserCount:       10001,
		ResourceCount:   0,
		Inconsistencies: expectedInconsistencies,
	}
	mockRepo.consistencyResult = expected
	svc := NewResourcePermissionService(mockRepo, &mockResourcePermAdminChecker{adminUserIDs: map[int64]bool{}})

	result, err := svc.CheckPermissionConsistency()
	if err != nil {
		t.Fatalf("CheckPermissionConsistency failed: %v", err)
	}
	if result != expected {
		t.Fatalf("expected delegated result pointer to pass through")
	}
	if !result.Consistent {
		t.Fatalf("expected skip result to remain consistent")
	}
	if result.UserCount != 10001 {
		t.Fatalf("expected skipped user count 10001, got %d", result.UserCount)
	}
	if len(result.Inconsistencies) != 1 || result.Inconsistencies[0].Description != expectedInconsistencies[0].Description {
		t.Fatalf("expected skip inconsistency description to pass through")
	}
}

func TestResourcePermissionService_DelegateMethods(t *testing.T) {
	mockRepo := newMockResourcePermRepo()
	mockRepo.userPerms[11] = []int64{1, 2}
	mockRepo.rolePerms[12] = []int64{3, 4}
	svc := NewResourcePermissionService(mockRepo, &mockResourcePermAdminChecker{adminUserIDs: map[int64]bool{}})

	userPerms, err := svc.GetUserPermissionIDs(11)
	if err != nil {
		t.Fatalf("GetUserPermissionIDs failed: %v", err)
	}
	if len(userPerms) != 2 {
		t.Fatalf("expected 2 user perms, got %d", len(userPerms))
	}

	rolePerms, err := svc.GetRolePermissionIDs(12)
	if err != nil {
		t.Fatalf("GetRolePermissionIDs failed: %v", err)
	}
	if len(rolePerms) != 2 {
		t.Fatalf("expected 2 role perms, got %d", len(rolePerms))
	}

	if err = svc.GrantPermissionToUser(1, 2, "tester"); err != nil {
		t.Fatalf("GrantPermissionToUser failed: %v", err)
	}
	if err = svc.RevokePermissionFromUser(1, 2); err != nil {
		t.Fatalf("RevokePermissionFromUser failed: %v", err)
	}
	if err = svc.GrantPermissionToRole(3, 4); err != nil {
		t.Fatalf("GrantPermissionToRole failed: %v", err)
	}
	if err = svc.RevokePermissionFromRole(3, 4); err != nil {
		t.Fatalf("RevokePermissionFromRole failed: %v", err)
	}
}

func TestTryInheritParentResourcePermissions_ParentGoverned(t *testing.T) {
	mockRepo := newMockResourcePermRepo()
	mockRepo.resourcePerms[resourceTypeKey(permission.ResourceTypeDatasource, 10)] = []int64{11, 12}
	svc := NewResourcePermissionService(mockRepo, nil)

	inherited, err := svc.TryInheritParentResourcePermissions(10, 20, "Child Datasource", permission.ResourceTypeDatasource)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !inherited {
		t.Fatalf("expected governed inheritance to be applied")
	}
	permIDs, exists, err := mockRepo.GetResourcePermissionIDs(20, permission.ResourceTypeDatasource)
	if err != nil {
		t.Fatalf("expected no lookup error, got %v", err)
	}
	if !exists {
		t.Fatalf("expected child resource permissions to exist after inheritance")
	}
	if len(permIDs) != 2 || permIDs[0] != 11 || permIDs[1] != 12 {
		t.Fatalf("expected inherited perm ids [11 12], got %v", permIDs)
	}
}

func TestTryInheritParentResourcePermissions_ParentNotGoverned(t *testing.T) {
	mockRepo := newMockResourcePermRepo()
	svc := NewResourcePermissionService(mockRepo, nil)

	inherited, err := svc.TryInheritParentResourcePermissions(10, 20, "Child Datasource", permission.ResourceTypeDatasource)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if inherited {
		t.Fatalf("expected ungoverned parent to skip inheritance")
	}
	permIDs, exists, err := mockRepo.GetResourcePermissionIDs(20, permission.ResourceTypeDatasource)
	if err != nil {
		t.Fatalf("expected no lookup error, got %v", err)
	}
	if exists || len(permIDs) != 0 {
		t.Fatalf("expected no child resource permissions for ungoverned parent, got exists=%v permIDs=%v", exists, permIDs)
	}
}

func TestCheckPermission_ResourcePermissionLookupFailed(t *testing.T) {
	mockRepo := newMockResourcePermRepo()
	mockRepo.resourcePermErr = errors.New("lookup failed")
	mockRepo.permKeys[permission.ResourceTypeDashboard+":"+permission.PermKeyView] = &permission.SysPerm{PermID: 1, PermKey: permission.ResourceTypeDashboard + ":" + permission.PermKeyView}
	svc := NewResourcePermissionService(mockRepo, &mockResourcePermAdminChecker{adminUserIDs: map[int64]bool{}})

	result := svc.CheckPermission(1, permission.ResourceTypeDashboard, 99, permission.PermKeyView)
	if result.HasPermission {
		t.Fatalf("expected permission denial when resource lookup fails")
	}
	if result.Reason != "resource_permission_lookup_failed" {
		t.Fatalf("expected reason resource_permission_lookup_failed, got %s", result.Reason)
	}
}

func TestCheckPermission_ExplicitEmptyResourcePermissionsDenied(t *testing.T) {
	mockRepo := newMockResourcePermRepo()
	mockRepo.permKeys[permission.ResourceTypeDashboard+":"+permission.PermKeyView] = &permission.SysPerm{PermID: 1, PermKey: permission.ResourceTypeDashboard + ":" + permission.PermKeyView}
	mockRepo.resourcePerms[resourceTypeKey(permission.ResourceTypeDashboard, 1001)] = []int64{}
	svc := NewResourcePermissionService(mockRepo, &mockResourcePermAdminChecker{adminUserIDs: map[int64]bool{}})

	result := svc.CheckPermission(1, permission.ResourceTypeDashboard, 1001, permission.PermKeyView)
	if result.HasPermission {
		t.Fatalf("expected permission denial when resource is explicitly governed with empty perms")
	}
	if result.Reason != "resource_permission_denied" {
		t.Fatalf("expected resource_permission_denied, got %s", result.Reason)
	}
}

func TestGetUserPerspective_GuardsAndAdmin(t *testing.T) {
	t.Run("nil repo returns error", func(t *testing.T) {
		svc := NewResourcePermissionService(nil, nil)
		_, err := svc.GetUserPerspective(1, permission.ResourceTypeDashboard)
		if err == nil || err.Error() != "repository not initialized" {
			t.Fatalf("expected repository not initialized error, got %v", err)
		}
	})

	t.Run("admin returns wildcard permission", func(t *testing.T) {
		svc := NewResourcePermissionService(newMockResourcePermRepo(), &mockResourcePermAdminChecker{adminUserIDs: map[int64]bool{7: true}})
		items, err := svc.GetUserPerspective(7, permission.ResourceTypeDashboard)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(items) != 1 || items[0].PermKey != "*" || items[0].SourceType != "admin" {
			t.Fatalf("unexpected admin result: %+v", items)
		}
	})
}

func TestGetResourcePerspective_AndConsistency_Guards(t *testing.T) {
	t.Run("nil repo returns error", func(t *testing.T) {
		svc := NewResourcePermissionService(nil, nil)
		if _, err := svc.GetResourcePerspective(1, permission.ResourceTypeDashboard); err == nil || err.Error() != "repository not initialized" {
			t.Fatalf("expected repository not initialized error, got %v", err)
		}
		if _, err := svc.CheckPermissionConsistency(); err == nil || err.Error() != "repository not initialized" {
			t.Fatalf("expected repository not initialized error, got %v", err)
		}
	})

	t.Run("delegates repository consistency result", func(t *testing.T) {
		wantErr := errors.New("consistency failed")
		mockRepo := newMockResourcePermRepo()
		mockRepo.consistencyErr = wantErr
		svc := NewResourcePermissionService(mockRepo, nil)
		_, err := svc.CheckPermissionConsistency()
		if !errors.Is(err, wantErr) {
			t.Fatalf("expected consistency error %v, got %v", wantErr, err)
		}
	})
}

func TestTryInheritParentResourcePermissions_GuardsAndErrors(t *testing.T) {
	t.Run("nil repo returns error", func(t *testing.T) {
		svc := NewResourcePermissionService(nil, nil)
		inherited, err := svc.TryInheritParentResourcePermissions(1, 2, "child", permission.ResourceTypeDataset)
		if inherited {
			t.Fatalf("expected inherited=false when repo missing")
		}
		if err == nil || err.Error() != "repository not initialized" {
			t.Fatalf("expected repository not initialized error, got %v", err)
		}
	})

	t.Run("invalid input returns false without repo writes", func(t *testing.T) {
		mockRepo := newMockResourcePermRepo()
		svc := NewResourcePermissionService(mockRepo, nil)
		inherited, err := svc.TryInheritParentResourcePermissions(0, 2, "child", permission.ResourceTypeDataset)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if inherited {
			t.Fatalf("expected invalid input to skip inheritance")
		}
		if mockRepo.registerCalls != 0 || mockRepo.replaceCalls != 0 {
			t.Fatalf("expected no repo writes for invalid input, got register=%d replace=%d", mockRepo.registerCalls, mockRepo.replaceCalls)
		}
	})

	t.Run("register error is returned", func(t *testing.T) {
		wantErr := errors.New("register failed")
		mockRepo := newMockResourcePermRepo()
		mockRepo.resourcePerms[resourceTypeKey(permission.ResourceTypeDataset, 10)] = []int64{5}
		mockRepo.registerErr = wantErr
		svc := NewResourcePermissionService(mockRepo, nil)
		_, err := svc.TryInheritParentResourcePermissions(10, 20, "child", permission.ResourceTypeDataset)
		if !errors.Is(err, wantErr) {
			t.Fatalf("expected register error %v, got %v", wantErr, err)
		}
	})

	t.Run("replace error is returned", func(t *testing.T) {
		wantErr := errors.New("replace failed")
		mockRepo := newMockResourcePermRepo()
		mockRepo.resourcePerms[resourceTypeKey(permission.ResourceTypeDataset, 10)] = []int64{5}
		mockRepo.replaceErr = wantErr
		svc := NewResourcePermissionService(mockRepo, nil)
		_, err := svc.TryInheritParentResourcePermissions(10, 20, "child", permission.ResourceTypeDataset)
		if !errors.Is(err, wantErr) {
			t.Fatalf("expected replace error %v, got %v", wantErr, err)
		}
	})
}

func TestResolvePermission_FallsBackToRawKey(t *testing.T) {
	mockRepo := newMockResourcePermRepo()
	mockRepo.permKeys[permission.PermKeyManage] = &permission.SysPerm{PermID: 44, PermKey: permission.PermKeyManage}
	svc := NewResourcePermissionService(mockRepo, nil)

	perm, err := svc.ResolvePermission(permission.ResourceTypeDatasource, permission.PermKeyManage)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if perm == nil || perm.PermID != 44 {
		t.Fatalf("expected fallback raw permission, got %+v", perm)
	}
}

func TestResourcePermissionService_ThinWrappersAndGuards(t *testing.T) {
	t.Run("apply group permissions delegates", func(t *testing.T) {
		mockRepo := newMockResourcePermRepo()
		svc := NewResourcePermissionService(mockRepo, nil)
		if err := svc.ApplyGroupPermissionsToResource(1, 2, permission.ResourceTypeDataset); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if mockRepo.applyGroupCalls != 1 {
			t.Fatalf("expected apply group to be called once, got %d", mockRepo.applyGroupCalls)
		}
	})

	t.Run("register and replace guard nil repo", func(t *testing.T) {
		svc := NewResourcePermissionService(nil, nil)
		if err := svc.RegisterResource(1, "name", permission.ResourceTypeDataset, nil); err == nil || err.Error() != "repository not initialized" {
			t.Fatalf("expected repository not initialized from RegisterResource, got %v", err)
		}
		if err := svc.ReplaceResourcePermissions(1, permission.ResourceTypeDataset, []int64{1}); err == nil || err.Error() != "repository not initialized" {
			t.Fatalf("expected repository not initialized from ReplaceResourcePermissions, got %v", err)
		}
	})

	t.Run("inherit wrapper forwards success", func(t *testing.T) {
		mockRepo := newMockResourcePermRepo()
		mockRepo.resourcePerms[resourceTypeKey(permission.ResourceTypeDataset, 11)] = []int64{2, 3}
		svc := NewResourcePermissionService(mockRepo, nil)
		if err := svc.InheritParentResourcePermissions(11, 22, "child", permission.ResourceTypeDataset); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		permIDs, exists, err := mockRepo.GetResourcePermissionIDs(22, permission.ResourceTypeDataset)
		if err != nil || !exists || len(permIDs) != 2 {
			t.Fatalf("expected inherited permissions on child, got exists=%v permIDs=%v err=%v", exists, permIDs, err)
		}
	})

	t.Run("register and replace delegate errors", func(t *testing.T) {
		registerErr := errors.New("register wrapper failed")
		replaceErr := errors.New("replace wrapper failed")
		mockRepo := newMockResourcePermRepo()
		mockRepo.registerErr = registerErr
		svc := NewResourcePermissionService(mockRepo, nil)
		if err := svc.RegisterResource(1, "name", permission.ResourceTypeDataset, nil); !errors.Is(err, registerErr) {
			t.Fatalf("expected register error %v, got %v", registerErr, err)
		}

		mockRepo = newMockResourcePermRepo()
		mockRepo.replaceErr = replaceErr
		svc = NewResourcePermissionService(mockRepo, nil)
		if err := svc.ReplaceResourcePermissions(1, permission.ResourceTypeDataset, []int64{1}); !errors.Is(err, replaceErr) {
			t.Fatalf("expected replace error %v, got %v", replaceErr, err)
		}
	})
}

func TestGetUserPerspective_RepoErrorAndDelegation(t *testing.T) {
	t.Run("non admin repo error", func(t *testing.T) {
		wantErr := errors.New("user perspective failed")
		mockRepo := newMockResourcePermRepo()
		mockRepo.userResourcesErr = wantErr
		svc := NewResourcePermissionService(mockRepo, &mockResourcePermAdminChecker{adminUserIDs: map[int64]bool{}})

		items, err := svc.GetUserPerspective(3, permission.ResourceTypeDashboard)
		if !errors.Is(err, wantErr) {
			t.Fatalf("expected user perspective error %v, got %v", wantErr, err)
		}
		if items != nil {
			t.Fatalf("expected nil items, got %+v", items)
		}
	})

	t.Run("non admin delegates success", func(t *testing.T) {
		mockRepo := newMockResourcePermRepo()
		mockRepo.userResources = []*permission.UserResourcePermVO{{PermKey: permission.PermKeyView, PermName: "查看", SourceType: "user"}}
		svc := NewResourcePermissionService(mockRepo, &mockResourcePermAdminChecker{adminUserIDs: map[int64]bool{}})

		items, err := svc.GetUserPerspective(4, permission.ResourceTypeDataset)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(items) != 1 || items[0].PermKey != permission.PermKeyView || items[0].SourceType != "user" {
			t.Fatalf("unexpected user perspective items: %+v", items)
		}
	})
}

func TestGetResourcePerspective_DelegationAndError(t *testing.T) {
	t.Run("delegates success", func(t *testing.T) {
		mockRepo := newMockResourcePermRepo()
		mockRepo.resourceUsers = []*permission.ResourceUserPermVO{{UserID: 8, Username: "alice", SourceType: "role"}}
		svc := NewResourcePermissionService(mockRepo, nil)

		items, err := svc.GetResourcePerspective(1, permission.ResourceTypeDashboard)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(items) != 1 || items[0].UserID != 8 || items[0].SourceType != "role" {
			t.Fatalf("unexpected resource perspective items: %+v", items)
		}
	})

	t.Run("repo error", func(t *testing.T) {
		wantErr := errors.New("resource perspective failed")
		mockRepo := newMockResourcePermRepo()
		mockRepo.resourceUsersErr = wantErr
		svc := NewResourcePermissionService(mockRepo, nil)

		items, err := svc.GetResourcePerspective(1, permission.ResourceTypeDashboard)
		if !errors.Is(err, wantErr) {
			t.Fatalf("expected resource perspective error %v, got %v", wantErr, err)
		}
		if items != nil {
			t.Fatalf("expected nil items, got %+v", items)
		}
	})
}

func TestResourcePermissionService_WrapperErrorsAndDelegation(t *testing.T) {
	t.Run("apply group permissions nil repo and error", func(t *testing.T) {
		svc := NewResourcePermissionService(nil, nil)
		if err := svc.ApplyGroupPermissionsToResource(1, 2, permission.ResourceTypeDataset); err == nil || err.Error() != "repository not initialized" {
			t.Fatalf("expected repository not initialized, got %v", err)
		}

		wantErr := errors.New("apply group failed")
		mockRepo := newMockResourcePermRepo()
		mockRepo.applyGroupErr = wantErr
		svc = NewResourcePermissionService(mockRepo, nil)
		if err := svc.ApplyGroupPermissionsToResource(1, 2, permission.ResourceTypeDataset); !errors.Is(err, wantErr) {
			t.Fatalf("expected apply group error %v, got %v", wantErr, err)
		}
	})

	t.Run("register delegates parent id and replace delegates perm ids", func(t *testing.T) {
		mockRepo := newMockResourcePermRepo()
		svc := NewResourcePermissionService(mockRepo, nil)
		parentID := int64(55)
		permIDs := []int64{7, 8}

		if err := svc.RegisterResource(2, "child", permission.ResourceTypeDataset, &parentID); err != nil {
			t.Fatalf("expected no register error, got %v", err)
		}
		if mockRepo.lastParentID == nil || *mockRepo.lastParentID != parentID {
			t.Fatalf("expected delegated parent id %d, got %+v", parentID, mockRepo.lastParentID)
		}

		if err := svc.ReplaceResourcePermissions(2, permission.ResourceTypeDataset, permIDs); err != nil {
			t.Fatalf("expected no replace error, got %v", err)
		}
		stored, exists, err := mockRepo.GetResourcePermissionIDs(2, permission.ResourceTypeDataset)
		if err != nil || !exists || len(stored) != 2 || stored[0] != 7 || stored[1] != 8 {
			t.Fatalf("expected delegated perm ids [7 8], got exists=%v stored=%v err=%v", exists, stored, err)
		}
	})
}

func TestResolvePermission_AndConsistency_AdditionalBranches(t *testing.T) {
	t.Run("resolve permission nil repo", func(t *testing.T) {
		svc := NewResourcePermissionService(nil, nil)
		perm, err := svc.ResolvePermission(permission.ResourceTypeDataset, permission.PermKeyView)
		if err == nil || err.Error() != "repository not initialized" {
			t.Fatalf("expected repository not initialized, got perm=%+v err=%v", perm, err)
		}
	})

	t.Run("resolve permission not found", func(t *testing.T) {
		mockRepo := newMockResourcePermRepo()
		svc := NewResourcePermissionService(mockRepo, nil)
		perm, err := svc.ResolvePermission(permission.ResourceTypeDataset, permission.ResourceTypeDataset+":"+permission.PermKeyView)
		if err == nil || !strings.Contains(err.Error(), "permission "+permission.ResourceTypeDataset+":"+permission.PermKeyView+" not found") {
			t.Fatalf("expected resolve not found error, got perm=%+v err=%v", perm, err)
		}
	})

	t.Run("consistency success", func(t *testing.T) {
		mockRepo := newMockResourcePermRepo()
		svc := NewResourcePermissionService(mockRepo, nil)
		result, err := svc.CheckPermissionConsistency()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if result == nil || !result.Consistent {
			t.Fatalf("expected consistent result, got %+v", result)
		}
	})
}

func TestCheckPermission_AdditionalFailureBranches(t *testing.T) {
	t.Run("role permission error falls through to denied", func(t *testing.T) {
		mockRepo := newMockResourcePermRepo()
		mockRepo.permKeys[permission.PermKeyView] = &permission.SysPerm{PermID: 1, PermKey: permission.PermKeyView}
		mockRepo.userRoles[13] = []int64{9}
		mockRepo.rolePermCheckErr = errors.New("role permission check failed")
		svc := NewResourcePermissionService(mockRepo, &mockResourcePermAdminChecker{adminUserIDs: map[int64]bool{}})

		result := svc.CheckPermission(13, permission.ResourceTypeDashboard, 100, permission.PermKeyView)
		if result.HasPermission || result.Reason != "permission_denied" {
			t.Fatalf("expected permission_denied after role perm check error, got %+v", result)
		}
	})

	t.Run("user permission miss then role permission denied", func(t *testing.T) {
		mockRepo := newMockResourcePermRepo()
		mockRepo.permKeys[permission.PermKeyEdit] = &permission.SysPerm{PermID: 2, PermKey: permission.PermKeyEdit}
		mockRepo.userRoles[14] = []int64{10}
		mockRepo.rolePermOk[10] = map[int64]bool{2: false}
		svc := NewResourcePermissionService(mockRepo, &mockResourcePermAdminChecker{adminUserIDs: map[int64]bool{}})

		result := svc.CheckPermission(14, permission.ResourceTypeDataset, 200, permission.PermKeyEdit)
		if result.HasPermission || result.Reason != "permission_denied" {
			t.Fatalf("expected permission_denied after user/role miss, got %+v", result)
		}
	})

	t.Run("no roles after role lookup error", func(t *testing.T) {
		mockRepo := newMockResourcePermRepo()
		mockRepo.permKeys[permission.PermKeyView] = &permission.SysPerm{PermID: 1, PermKey: permission.PermKeyView}
		mockRepo.userRoleErr = errors.New("role lookup failed")
		svc := NewResourcePermissionService(mockRepo, &mockResourcePermAdminChecker{adminUserIDs: map[int64]bool{}})

		result := svc.CheckPermission(13, permission.ResourceTypeDashboard, 100, permission.PermKeyView)
		if result.HasPermission || result.Reason != "no_roles" {
			t.Fatalf("expected no_roles after role lookup error, got %+v", result)
		}
	})

	t.Run("permission not found", func(t *testing.T) {
		mockRepo := newMockResourcePermRepo()
		svc := NewResourcePermissionService(mockRepo, &mockResourcePermAdminChecker{adminUserIDs: map[int64]bool{}})

		result := svc.CheckPermission(14, permission.ResourceTypeDashboard, 100, permission.ResourceTypeDashboard+":missing")
		if result.HasPermission || result.Reason != "permission_not_found" {
			t.Fatalf("expected permission_not_found, got %+v", result)
		}
	})
}

func TestResourcePermissionService_AdditionalDelegations(t *testing.T) {
	t.Run("get permission by id delegates", func(t *testing.T) {
		mockRepo := newMockResourcePermRepo()
		mockRepo.permByID[77] = &permission.SysPerm{PermID: 77, PermKey: permission.PermKeyManage}
		svc := NewResourcePermissionService(mockRepo, nil)

		perm, err := svc.GetPermissionByID(77)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if perm == nil || perm.PermID != 77 || perm.PermKey != permission.PermKeyManage {
			t.Fatalf("unexpected permission result: %+v", perm)
		}
	})

	t.Run("get user permission ids repo error", func(t *testing.T) {
		wantErr := errors.New("user perm query failed")
		mockRepo := newMockResourcePermRepo()
		mockRepo.userPermsErr = wantErr
		svc := NewResourcePermissionService(mockRepo, nil)

		ids, err := svc.GetUserPermissionIDs(1)
		if !errors.Is(err, wantErr) {
			t.Fatalf("expected user perm error %v, got %v", wantErr, err)
		}
		if ids != nil {
			t.Fatalf("expected nil ids, got %v", ids)
		}
	})

	t.Run("get role permission ids repo error", func(t *testing.T) {
		wantErr := errors.New("role perm query failed")
		mockRepo := newMockResourcePermRepo()
		mockRepo.rolePermsErr = wantErr
		svc := NewResourcePermissionService(mockRepo, nil)

		ids, err := svc.GetRolePermissionIDs(2)
		if !errors.Is(err, wantErr) {
			t.Fatalf("expected role perm error %v, got %v", wantErr, err)
		}
		if ids != nil {
			t.Fatalf("expected nil ids, got %v", ids)
		}
	})
}
