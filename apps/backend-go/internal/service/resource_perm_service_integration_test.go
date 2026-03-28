//go:build integration

package service

import (
	"testing"

	"dataease/backend/internal/domain/permission"
	"dataease/backend/internal/domain/role"
	"dataease/backend/internal/domain/user"
	"dataease/backend/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResourcePermissionService_GetUserPerspective(t *testing.T) {
	repo := &mockResourcePermRepoIT{
		userResources: []*permission.UserResourcePermVO{
			{
				ResourceID:   101,
				ResourceName: "sales-dashboard",
				ResourceType: "dashboard",
				PermKey:      "resource:view",
				PermName:     "查看",
				SourceType:   "direct",
			},
		},
	}
	svc := NewResourcePermissionService(repo, &mockAdminCheckerForPerm{isAdmin: false})

	result, err := svc.GetUserPerspective(1001, "dashboard")
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, int64(101), result[0].ResourceID)
}

func TestResourcePermissionService_GetUserPerspectiveAdmin(t *testing.T) {
	repo := &mockResourcePermRepoIT{}
	svc := NewResourcePermissionService(repo, &mockAdminCheckerForPerm{isAdmin: true})

	result, err := svc.GetUserPerspective(1, "resource")
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "*", result[0].PermKey)
	assert.Equal(t, "admin", result[0].SourceType)
}

func TestResourcePermissionService_GetResourcePerspective(t *testing.T) {
	repo := &mockResourcePermRepoIT{
		resourceUsers: []*permission.ResourceUserPermVO{
			{
				UserID:     2001,
				Username:   "tester",
				PermKey:    "resource:edit",
				PermName:   "编辑",
				SourceType: "role",
			},
		},
	}
	svc := NewResourcePermissionService(repo, nil)

	result, err := svc.GetResourcePerspective(101, "dashboard")
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, int64(2001), result[0].UserID)
}

func TestResourcePermissionService_ApplyGroupPermissionsToResource(t *testing.T) {
	repo := &mockResourcePermRepoIT{}
	svc := NewResourcePermissionService(repo, nil)

	err := svc.ApplyGroupPermissionsToResource(10, 101, "dashboard")
	assert.NoError(t, err)
	assert.True(t, repo.applyCalled)
}

func TestResourcePermissionService_CheckPermissionConsistency(t *testing.T) {
	repo := &mockResourcePermRepoIT{
		consistency: &permission.PermissionConsistencyResult{
			Consistent:      true,
			UserCount:       3,
			ResourceCount:   2,
			Inconsistencies: []*permission.PermissionInconsistencyVO{},
		},
	}
	svc := NewResourcePermissionService(repo, nil)

	result, err := svc.CheckPermissionConsistency()
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Consistent)
}

func TestResourcePermissionService_GetUserPerspective_RepoNil(t *testing.T) {
	svc := NewResourcePermissionService(nil, nil)
	result, err := svc.GetUserPerspective(1001, "dashboard")
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestResourcePermissionService_GetResourcePerspective_RepoNil(t *testing.T) {
	svc := NewResourcePermissionService(nil, nil)
	result, err := svc.GetResourcePerspective(101, "dashboard")
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestResourcePermissionService_ApplyGroupPermissionsToResource_RepoNil(t *testing.T) {
	svc := NewResourcePermissionService(nil, nil)
	err := svc.ApplyGroupPermissionsToResource(10, 101, "dashboard")
	assert.Error(t, err)
}

func TestResourcePermissionService_CheckPermissionConsistency_RepoNil(t *testing.T) {
	svc := NewResourcePermissionService(nil, nil)
	result, err := svc.CheckPermissionConsistency()
	assert.Error(t, err)
	assert.Nil(t, result)
}

type mockAdminCheckerForPerm struct {
	isAdmin bool
}

func (m *mockAdminCheckerForPerm) IsAdmin(userID int64) bool {
	return m.isAdmin
}

type mockResourcePermRepoIT struct {
	userResources []*permission.UserResourcePermVO
	resourceUsers []*permission.ResourceUserPermVO
	consistency   *permission.PermissionConsistencyResult
	applyCalled   bool
}

func (m *mockResourcePermRepoIT) GetPermByID(permID int64) (*permission.SysPerm, error) {
	return nil, nil
}
func (m *mockResourcePermRepoIT) GetPermByKey(permKey string) (*permission.SysPerm, error) {
	return nil, nil
}
func (m *mockResourcePermRepoIT) ListPerms(permType string, page, size int) ([]*permission.SysPerm, int64, error) {
	return nil, 0, nil
}
func (m *mockResourcePermRepoIT) CreatePerm(perm *permission.SysPerm) error { return nil }
func (m *mockResourcePermRepoIT) UpdatePerm(perm *permission.SysPerm) error { return nil }
func (m *mockResourcePermRepoIT) DeletePerm(permID int64) error             { return nil }
func (m *mockResourcePermRepoIT) GetUserPerms(userID int64) ([]int64, error) {
	return nil, nil
}
func (m *mockResourcePermRepoIT) GetRolePerms(roleID int64) ([]int64, error) {
	return nil, nil
}
func (m *mockResourcePermRepoIT) GetUserRoleIDs(userID int64) ([]int64, error) {
	return nil, nil
}
func (m *mockResourcePermRepoIT) CheckUserPermission(userID, permID int64) (bool, error) {
	return false, nil
}
func (m *mockResourcePermRepoIT) CheckRolePermission(roleID, permID int64) (bool, error) {
	return false, nil
}
func (m *mockResourcePermRepoIT) GrantPermToUser(userID, permID int64, createBy string) error {
	return nil
}
func (m *mockResourcePermRepoIT) RevokePermFromUser(userID, permID int64) error { return nil }
func (m *mockResourcePermRepoIT) GrantPermToRole(roleID, permID int64) error    { return nil }
func (m *mockResourcePermRepoIT) RevokePermFromRole(roleID, permID int64) error { return nil }
func (m *mockResourcePermRepoIT) GetUserResources(userID int64, resourceType string) ([]*permission.UserResourcePermVO, error) {
	if m.userResources == nil {
		return []*permission.UserResourcePermVO{}, nil
	}
	return m.userResources, nil
}
func (m *mockResourcePermRepoIT) GetResourceUsers(resourceID int64, resourceType string) ([]*permission.ResourceUserPermVO, error) {
	if m.resourceUsers == nil {
		return []*permission.ResourceUserPermVO{}, nil
	}
	return m.resourceUsers, nil
}
func (m *mockResourcePermRepoIT) ApplyGroupPermissions(groupID, resourceID int64, resourceType string) error {
	m.applyCalled = true
	return nil
}
func (m *mockResourcePermRepoIT) RegisterResource(resourceID int64, resourceName, resourceType string, parentID *int64) error {
	return nil
}
func (m *mockResourcePermRepoIT) ReplaceResourcePermissions(resourceID int64, resourceType string, permIDs []int64) error {
	return nil
}
func (m *mockResourcePermRepoIT) GetResourcePermissionIDs(resourceID int64, resourceType string) ([]int64, bool, error) {
	return nil, false, nil
}
func (m *mockResourcePermRepoIT) CheckPermissionConsistency() (*permission.PermissionConsistencyResult, error) {
	if m.consistency == nil {
		return &permission.PermissionConsistencyResult{Consistent: true}, nil
	}
	return m.consistency, nil
}

func TestResourcePermissionService_GovernedResourcePerspectiveMatchesRuntimeChecks(t *testing.T) {
	testCases := []struct {
		name         string
		resourceType string
		resourceID   int64
	}{
		{name: "datasource", resourceType: permission.ResourceTypeDatasource, resourceID: 1101},
		{name: "dataset", resourceType: permission.ResourceTypeDataset, resourceID: 1102},
		{name: "dashboard", resourceType: permission.ResourceTypeDashboard, resourceID: 1103},
		{name: "screen", resourceType: permission.ResourceTypeScreen, resourceID: 1104},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cleanupTables(&permission.SysResourcePerm{}, &permission.SysResource{}, &permission.SysRolePerm{}, &permission.SysUserPerm{}, &permission.SysPerm{}, &user.SysUser{}, &user.SysUserRole{}, &role.SysRole{})

			repo := repository.NewResourcePermissionRepository(testDB)
			svc := NewResourcePermissionService(repo, nil)

			roleType := "organization"
			testRole := &role.SysRole{RoleName: tc.name + " governed role", RoleCode: tc.name + "_governed_role", RoleType: &roleType, Status: 1}
			require.NoError(t, testDB.Create(testRole).Error)

			testUser := &user.SysUser{Username: tc.name + "_resource_user", Password: "irrelevant", NickName: tc.name + " user", Status: 1, DelFlag: 0}
			require.NoError(t, testDB.Create(testUser).Error)
			require.NoError(t, testDB.Create(&user.SysUserRole{UserID: testUser.UserID, RoleID: testRole.RoleID, OrgID: 1}).Error)

			viewPerm := &permission.SysPerm{PermName: tc.name + " 查看", PermKey: tc.resourceType + ":view", PermType: permission.PermTypeData, Status: 1, DelFlag: 0}
			require.NoError(t, testDB.Create(viewPerm).Error)
			require.NoError(t, testDB.Create(&permission.SysRolePerm{RoleID: testRole.RoleID, PermID: viewPerm.PermID}).Error)

			resourceName := tc.name + "-governed-resource"
			require.NoError(t, svc.RegisterResource(tc.resourceID, resourceName, tc.resourceType, nil))
			require.NoError(t, svc.ReplaceResourcePermissions(tc.resourceID, tc.resourceType, []int64{viewPerm.PermID}))

			resourcePerspective, err := svc.GetResourcePerspective(tc.resourceID, tc.resourceType)
			require.NoError(t, err)
			if assert.Len(t, resourcePerspective, 1) {
				assert.Equal(t, testUser.UserID, resourcePerspective[0].UserID)
				assert.Equal(t, tc.resourceType+":view", resourcePerspective[0].PermKey)
				assert.Equal(t, "role", resourcePerspective[0].SourceType)
			}

			userPerspective, err := svc.GetUserPerspective(testUser.UserID, tc.resourceType)
			require.NoError(t, err)
			if assert.Len(t, userPerspective, 1) {
				assert.Equal(t, tc.resourceType, userPerspective[0].ResourceType)
				assert.Equal(t, tc.resourceType+":view", userPerspective[0].PermKey)
				assert.Equal(t, "role", userPerspective[0].SourceType)
			}

			result := svc.CheckPermission(testUser.UserID, tc.resourceType, tc.resourceID, permission.PermKeyView)
			assert.True(t, result.HasPermission)
			assert.Equal(t, "role_permission", result.Reason)

			denyPerm := &permission.SysPerm{PermName: tc.name + " 编辑", PermKey: tc.resourceType + ":edit", PermType: permission.PermTypeData, Status: 1, DelFlag: 0}
			require.NoError(t, testDB.Create(denyPerm).Error)
			require.NoError(t, testDB.Create(&permission.SysRolePerm{RoleID: testRole.RoleID, PermID: denyPerm.PermID}).Error)

			denyResult := svc.CheckPermission(testUser.UserID, tc.resourceType, tc.resourceID, permission.PermKeyEdit)
			assert.False(t, denyResult.HasPermission)
			assert.Equal(t, "resource_permission_denied", denyResult.Reason)
		})
	}
}
