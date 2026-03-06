package service

import (
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
	userPerms  map[int64][]int64
	rolePerms  map[int64][]int64
	userRoles  map[int64][]int64
	permKeys   map[string]*permission.SysPerm
	userPermOk map[int64]map[int64]bool
	rolePermOk map[int64]map[int64]bool
}

func newMockResourcePermRepo() *mockResourcePermRepo {
	return &mockResourcePermRepo{
		userPerms:  make(map[int64][]int64),
		rolePerms:  make(map[int64][]int64),
		userRoles:  make(map[int64][]int64),
		permKeys:   make(map[string]*permission.SysPerm),
		userPermOk: make(map[int64]map[int64]bool),
		rolePermOk: make(map[int64]map[int64]bool),
	}
}

func (m *mockResourcePermRepo) GetPermByID(permID int64) (*permission.SysPerm, error) {
	return nil, nil
}

func (m *mockResourcePermRepo) GetPermByKey(permKey string) (*permission.SysPerm, error) {
	if perm, ok := m.permKeys[permKey]; ok {
		return perm, nil
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
	return m.userPerms[userID], nil
}

func (m *mockResourcePermRepo) GetRolePerms(roleID int64) ([]int64, error) {
	return m.rolePerms[roleID], nil
}

func (m *mockResourcePermRepo) GetUserRoleIDs(userID int64) ([]int64, error) {
	return m.userRoles[userID], nil
}

func (m *mockResourcePermRepo) CheckUserPermission(userID, permID int64) (bool, error) {
	if userPerms, ok := m.userPermOk[userID]; ok {
		return userPerms[permID], nil
	}
	return false, nil
}

func (m *mockResourcePermRepo) CheckRolePermission(roleID, permID int64) (bool, error) {
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
	return []*permission.UserResourcePermVO{}, nil
}

func (m *mockResourcePermRepo) GetResourceUsers(resourceID int64, resourceType string) ([]*permission.ResourceUserPermVO, error) {
	return []*permission.ResourceUserPermVO{}, nil
}

func (m *mockResourcePermRepo) ApplyGroupPermissions(groupID, resourceID int64, resourceType string) error {
	return nil
}

func (m *mockResourcePermRepo) CheckPermissionConsistency() (*permission.PermissionConsistencyResult, error) {
	return &permission.PermissionConsistencyResult{Consistent: true}, nil
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
