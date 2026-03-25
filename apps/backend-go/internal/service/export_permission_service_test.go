package service

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"dataease/backend/internal/domain/permission"
)

type mockExportResourcePermRepo struct {
	userPermOk map[int64]map[int64]bool
	rolePermOk map[int64]map[int64]bool
	userRoles  map[int64][]int64
	permKeys   map[string]*permission.SysPerm
}

func newMockExportResourcePermRepo() *mockExportResourcePermRepo {
	return &mockExportResourcePermRepo{
		userPermOk: make(map[int64]map[int64]bool),
		rolePermOk: make(map[int64]map[int64]bool),
		userRoles:  make(map[int64][]int64),
		permKeys:   make(map[string]*permission.SysPerm),
	}
}

func (m *mockExportResourcePermRepo) GetPermByID(permID int64) (*permission.SysPerm, error) {
	return nil, nil
}

func (m *mockExportResourcePermRepo) GetPermByKey(permKey string) (*permission.SysPerm, error) {
	if perm, ok := m.permKeys[permKey]; ok {
		return perm, nil
	}
	if idx := strings.Index(permKey, ":"); idx >= 0 {
		if perm, ok := m.permKeys[permKey[idx+1:]]; ok {
			return perm, nil
		}
	}
	if strings.Contains(permKey, ":") {
		return nil, fmt.Errorf("permission not found")
	}
	return &permission.SysPerm{PermID: 1, PermKey: permKey}, nil
}

func (m *mockExportResourcePermRepo) ListPerms(permType string, page, size int) ([]*permission.SysPerm, int64, error) {
	return nil, 0, nil
}

func (m *mockExportResourcePermRepo) CreatePerm(perm *permission.SysPerm) error {
	return nil
}

func (m *mockExportResourcePermRepo) UpdatePerm(perm *permission.SysPerm) error {
	return nil
}

func (m *mockExportResourcePermRepo) DeletePerm(permID int64) error {
	return nil
}

func (m *mockExportResourcePermRepo) GetUserPerms(userID int64) ([]int64, error) {
	return nil, nil
}

func (m *mockExportResourcePermRepo) GetRolePerms(roleID int64) ([]int64, error) {
	return nil, nil
}

func (m *mockExportResourcePermRepo) GetUserRoleIDs(userID int64) ([]int64, error) {
	return m.userRoles[userID], nil
}

func (m *mockExportResourcePermRepo) CheckUserPermission(userID, permID int64) (bool, error) {
	if userPerms, ok := m.userPermOk[userID]; ok {
		return userPerms[permID], nil
	}
	return false, nil
}

func (m *mockExportResourcePermRepo) CheckRolePermission(roleID, permID int64) (bool, error) {
	if rolePerms, ok := m.rolePermOk[roleID]; ok {
		return rolePerms[permID], nil
	}
	return false, nil
}

func (m *mockExportResourcePermRepo) GrantPermToUser(userID, permID int64, createBy string) error {
	return nil
}

func (m *mockExportResourcePermRepo) RevokePermFromUser(userID, permID int64) error {
	return nil
}

func (m *mockExportResourcePermRepo) GrantPermToRole(roleID, permID int64) error {
	return nil
}

func (m *mockExportResourcePermRepo) RevokePermFromRole(roleID, permID int64) error {
	return nil
}

// 双视角接口 mock
// 双视角接口 mock
func (m *mockExportResourcePermRepo) GetUserResources(userID int64, resourceType string) ([]*permission.UserResourcePermVO, error) {
	return []*permission.UserResourcePermVO{}, nil
}

func (m *mockExportResourcePermRepo) GetResourceUsers(resourceID int64, resourceType string) ([]*permission.ResourceUserPermVO, error) {
	return []*permission.ResourceUserPermVO{}, nil
}

func (m *mockExportResourcePermRepo) ApplyGroupPermissions(groupID, resourceID int64, resourceType string) error {
	return nil
}

func (m *mockExportResourcePermRepo) RegisterResource(resourceID int64, resourceName, resourceType string, parentID *int64) error {
	return nil
}

func (m *mockExportResourcePermRepo) ReplaceResourcePermissions(resourceID int64, resourceType string, permIDs []int64) error {
	return nil
}

func (m *mockExportResourcePermRepo) GetResourcePermissionIDs(resourceID int64, resourceType string) ([]int64, bool, error) {
	return nil, false, nil
}

func (m *mockExportResourcePermRepo) CheckPermissionConsistency() (*permission.PermissionConsistencyResult, error) {
	return &permission.PermissionConsistencyResult{Consistent: true}, nil
}

func TestExportPermissionService_CheckDashboardExport(t *testing.T) {
	mockRepo := newMockExportResourcePermRepo()
	mockRepo.userPermOk[1] = map[int64]bool{1: true}
	mockRepo.permKeys[permission.PermKeyExport] = &permission.SysPerm{PermID: 1, PermKey: permission.PermKeyExport}
	mockChecker := &mockResourcePermAdminChecker{adminUserIDs: map[int64]bool{}}
	resourceSvc := NewResourcePermissionService(mockRepo, mockChecker)
	exportSvc := NewExportPermissionService(resourceSvc, nil)
	ctx := context.Background()

	result := exportSvc.CheckDashboardExport(ctx, 1, 100)
	if !result.HasPermission {
		t.Errorf("User with export permission should be able to export, got %v", result.HasPermission)
	}

	result = exportSvc.CheckDashboardExport(ctx, 2, 100)
	if result.HasPermission {
		t.Error("User without export permission should not be able to export")
	}
}

func TestExportPermissionService_CheckDatasetExport(t *testing.T) {
	mockRepo := newMockExportResourcePermRepo()
	mockRepo.userPermOk[1] = map[int64]bool{1: true}
	mockRepo.permKeys[permission.PermKeyExport] = &permission.SysPerm{PermID: 1, PermKey: permission.PermKeyExport}
	mockChecker := &mockResourcePermAdminChecker{adminUserIDs: map[int64]bool{}}
	resourceSvc := NewResourcePermissionService(mockRepo, mockChecker)
	exportSvc := NewExportPermissionService(resourceSvc, nil)
	ctx := context.Background()

	result := exportSvc.CheckDatasetExport(ctx, 1, 200)
	if !result.HasPermission {
		t.Errorf("User with export permission should be able to export dataset, got %v", result.HasPermission)
	}
}

func TestExportPermissionService_CanExportImage(t *testing.T) {
	mockRepo := newMockExportResourcePermRepo()
	mockRepo.userPermOk[1] = map[int64]bool{1: true}
	mockRepo.permKeys[permission.PermKeyExport] = &permission.SysPerm{PermID: 1, PermKey: permission.PermKeyExport}
	mockChecker := &mockResourcePermAdminChecker{adminUserIDs: map[int64]bool{}}
	resourceSvc := NewResourcePermissionService(mockRepo, mockChecker)
	exportSvc := NewExportPermissionService(resourceSvc, nil)
	ctx := context.Background()

	canExport := exportSvc.CanExportImage(ctx, 1, permission.ResourceTypeDashboard, 100)
	if !canExport {
		t.Error("User with export permission should be able to export image")
	}
}

func TestExportPermissionService_CanExportPDF(t *testing.T) {
	mockRepo := newMockExportResourcePermRepo()
	mockRepo.userPermOk[1] = map[int64]bool{1: true}
	mockRepo.permKeys[permission.PermKeyExport] = &permission.SysPerm{PermID: 1, PermKey: permission.PermKeyExport}
	mockChecker := &mockResourcePermAdminChecker{adminUserIDs: map[int64]bool{}}
	resourceSvc := NewResourcePermissionService(mockRepo, mockChecker)
	exportSvc := NewExportPermissionService(resourceSvc, nil)
	ctx := context.Background()

	canExport := exportSvc.CanExportPDF(ctx, 1, permission.ResourceTypeDashboard, 100)
	if !canExport {
		t.Error("User with export permission should be able to export PDF")
	}
}

func TestExportPermissionService_CanExportExcel(t *testing.T) {
	mockRepo := newMockExportResourcePermRepo()
	mockRepo.userPermOk[1] = map[int64]bool{1: true}
	mockRepo.permKeys[permission.PermKeyExport] = &permission.SysPerm{PermID: 1, PermKey: permission.PermKeyExport}
	mockChecker := &mockResourcePermAdminChecker{adminUserIDs: map[int64]bool{}}
	resourceSvc := NewResourcePermissionService(mockRepo, mockChecker)
	exportSvc := NewExportPermissionService(resourceSvc, nil)
	ctx := context.Background()

	canExport := exportSvc.CanExportExcel(ctx, 1, permission.ResourceTypeDataset, 100)
	if !canExport {
		t.Error("User with export permission should be able to export Excel")
	}
}

func TestExportPermissionService_CanExportTemplate(t *testing.T) {
	mockRepo := newMockExportResourcePermRepo()
	mockRepo.userPermOk[1] = map[int64]bool{2: true}
	mockRepo.permKeys[permission.PermKeyManage] = &permission.SysPerm{PermID: 2, PermKey: permission.PermKeyManage}
	mockChecker := &mockResourcePermAdminChecker{adminUserIDs: map[int64]bool{}}
	resourceSvc := NewResourcePermissionService(mockRepo, mockChecker)
	exportSvc := NewExportPermissionService(resourceSvc, nil)
	ctx := context.Background()

	canExport := exportSvc.CanExportTemplate(ctx, 1, permission.ResourceTypeDashboard, 100)
	if !canExport {
		t.Error("User with manage permission should be able to export template")
	}
}

func TestExportPermissionService_CheckExportPermission(t *testing.T) {
	mockRepo := newMockExportResourcePermRepo()
	mockRepo.userPermOk[1] = map[int64]bool{1: true}
	mockRepo.permKeys[permission.PermKeyExport] = &permission.SysPerm{PermID: 1, PermKey: permission.PermKeyExport}
	mockChecker := &mockResourcePermAdminChecker{adminUserIDs: map[int64]bool{}}
	resourceSvc := NewResourcePermissionService(mockRepo, mockChecker)
	exportSvc := NewExportPermissionService(resourceSvc, nil)
	ctx := context.Background()

	req := &ExportCheckRequest{
		UserID:       1,
		ResourceType: permission.ResourceTypeDashboard,
		ResourceID:   100,
		ExportType:   ExportTypeImage,
	}

	result := exportSvc.CheckExportPermission(ctx, req)
	if !result.CanExport {
		t.Error("Should be able to export image")
	}
	if result.RequiredPerm != permission.PermKeyExport {
		t.Errorf("Expected required perm 'export', got '%s'", result.RequiredPerm)
	}

	req.ExportType = ExportTypeTemplate
	mockRepo.userPermOk[1] = map[int64]bool{2: true}
	mockRepo.permKeys[permission.PermKeyManage] = &permission.SysPerm{PermID: 2, PermKey: permission.PermKeyManage}

	result = exportSvc.CheckExportPermission(ctx, req)
	if result.RequiredPerm != permission.PermKeyManage {
		t.Errorf("Expected required perm 'manage' for template export, got '%s'", result.RequiredPerm)
	}
}

func TestExportPermissionService_CheckBatchExportPermission(t *testing.T) {
	mockRepo := newMockExportResourcePermRepo()
	mockRepo.userPermOk[1] = map[int64]bool{1: true}
	mockRepo.permKeys[permission.PermKeyExport] = &permission.SysPerm{PermID: 1, PermKey: permission.PermKeyExport}
	mockChecker := &mockResourcePermAdminChecker{adminUserIDs: map[int64]bool{}}
	resourceSvc := NewResourcePermissionService(mockRepo, mockChecker)
	exportSvc := NewExportPermissionService(resourceSvc, nil)
	ctx := context.Background()

	resourceIDs := []int64{100, 200, 300}
	results := exportSvc.CheckBatchExportPermission(ctx, 1, permission.ResourceTypeDashboard, resourceIDs)

	for _, id := range resourceIDs {
		if !results[id] {
			t.Errorf("Should be able to export resource %d", id)
		}
	}
}

func TestExportPermissionService_FilterExportableResources(t *testing.T) {
	mockRepo := newMockExportResourcePermRepo()
	mockRepo.userPermOk[1] = map[int64]bool{1: true}
	mockRepo.permKeys[permission.PermKeyExport] = &permission.SysPerm{PermID: 1, PermKey: permission.PermKeyExport}
	mockChecker := &mockResourcePermAdminChecker{adminUserIDs: map[int64]bool{}}
	resourceSvc := NewResourcePermissionService(mockRepo, mockChecker)
	exportSvc := NewExportPermissionService(resourceSvc, nil)
	ctx := context.Background()

	resourceIDs := []int64{100, 200, 300}
	exportable := exportSvc.FilterExportableResources(ctx, 1, permission.ResourceTypeDashboard, resourceIDs)

	if len(exportable) != 3 {
		t.Errorf("Expected 3 exportable resources, got %d", len(exportable))
	}
}

func TestExportPermissionService_AdminBypass(t *testing.T) {
	mockRepo := newMockExportResourcePermRepo()
	mockChecker := &mockResourcePermAdminChecker{adminUserIDs: map[int64]bool{1: true}}
	resourceSvc := NewResourcePermissionService(mockRepo, mockChecker)
	exportSvc := NewExportPermissionService(resourceSvc, nil)
	ctx := context.Background()

	result := exportSvc.CheckDashboardExport(ctx, 1, 100)
	if !result.HasPermission {
		t.Error("Admin should always be able to export")
	}
	if result.Reason != "admin" {
		t.Errorf("Expected reason 'admin', got '%s'", result.Reason)
	}
}

func TestExportPermissionService_CheckScreenAndDatasourceExport(t *testing.T) {
	mockRepo := newMockExportResourcePermRepo()
	mockRepo.userPermOk[1] = map[int64]bool{1: true}
	mockRepo.permKeys[permission.PermKeyExport] = &permission.SysPerm{PermID: 1, PermKey: permission.PermKeyExport}
	mockChecker := &mockResourcePermAdminChecker{adminUserIDs: map[int64]bool{}}
	resourceSvc := NewResourcePermissionService(mockRepo, mockChecker)
	exportSvc := NewExportPermissionService(resourceSvc, nil)
	ctx := context.Background()

	screen := exportSvc.CheckScreenExport(ctx, 1, 101)
	if !screen.HasPermission {
		t.Error("expected screen export permission")
	}

	datasource := exportSvc.CheckDatasourceExport(ctx, 1, 201)
	if !datasource.HasPermission {
		t.Error("expected datasource export permission")
	}
}
