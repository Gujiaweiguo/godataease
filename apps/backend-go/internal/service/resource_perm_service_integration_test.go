//go:build integration

package service

import (
	"testing"

	"dataease/backend/internal/domain/permission"

	"github.com/stretchr/testify/assert"
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
func (m *mockResourcePermRepoIT) CheckPermissionConsistency() (*permission.PermissionConsistencyResult, error) {
	if m.consistency == nil {
		return &permission.PermissionConsistencyResult{Consistent: true}, nil
	}
	return m.consistency, nil
}
