//go:build integration

package service

import (
	"testing"

	"dataease/backend/internal/domain/role"
	"dataease/backend/internal/domain/user"
	"dataease/backend/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create RoleService with all dependencies
func newTestRoleService(t *testing.T) *RoleService {
	cleanupTables(&role.SysRole{})

	repo := repository.NewRoleRepository(testDB)
	userRepo := repository.NewUserRepository(testDB)
	userRoleRepo := repository.NewUserRoleRepository(testDB)

	return NewRoleService(repo, userRepo, userRoleRepo)
}

func TestRoleServiceIntegration_Create(t *testing.T) {
	svc := newTestRoleService(t)
	repo := repository.NewRoleRepository(testDB)

	desc := "Test role description"
	req := &role.RoleCreator{
		Name: "TestRole",
		Desc: &desc,
	}

	id, err := svc.CreateRole(req, "tester")
	assert.NoError(t, err)
	assert.Greater(t, id, int64(0))

	// Verify created
	found, err := repo.GetByID(id)
	assert.NoError(t, err)
	assert.Equal(t, "TestRole", found.RoleName)
	assert.Contains(t, found.RoleCode, "role_")
	assert.Equal(t, role.StatusEnabled, found.Status)
	assert.NotNil(t, found.DataScope)
	assert.Equal(t, role.DataScopeSelf, *found.DataScope)
	assert.NotNil(t, found.CreateBy)
	assert.Equal(t, "tester", *found.CreateBy)
}

func TestRoleServiceIntegration_Edit(t *testing.T) {
	svc := newTestRoleService(t)

	// Create role
	initialDesc := "Initial description"
	id, _ := svc.CreateRole(&role.RoleCreator{Name: "ToEdit", Desc: &initialDesc}, "creator")

	// Edit
	newDesc := "Updated description"
	err := svc.EditRole(&role.RoleEditor{
		ID:   id,
		Name: "EditedRole",
		Desc: &newDesc,
	}, "editor")
	assert.NoError(t, err)

	// Verify
	repo := repository.NewRoleRepository(testDB)
	updated, err := repo.GetByID(id)
	assert.NoError(t, err)
	assert.Equal(t, "EditedRole", updated.RoleName)
	assert.NotNil(t, updated.RoleDesc)
	assert.Equal(t, newDesc, *updated.RoleDesc)
	assert.NotNil(t, updated.UpdateBy)
	assert.Equal(t, "editor", *updated.UpdateBy)
	assert.NotNil(t, updated.UpdateTime)
}

func TestRoleServiceIntegration_EditNotFound(t *testing.T) {
	svc := newTestRoleService(t)

	err := svc.EditRole(&role.RoleEditor{ID: 9999, Name: "NotFound"}, "editor")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "role not found")
}

func TestRoleServiceIntegration_Delete(t *testing.T) {
	svc := newTestRoleService(t)

	// Create role
	id, _ := svc.CreateRole(&role.RoleCreator{Name: "ToDelete"}, "tester")

	// Delete
	err := svc.DeleteRole(id)
	assert.NoError(t, err)

	// Verify deleted
	repo := repository.NewRoleRepository(testDB)
	_, err = repo.GetByID(id)
	assert.Error(t, err)
}

func TestRoleServiceIntegration_GetRoleByID(t *testing.T) {
	svc := newTestRoleService(t)

	// Create role
	id, _ := svc.CreateRole(&role.RoleCreator{
		Name: "DetailRole",
		Desc: pStr("Detailed role"),
	}, "tester")

	// Get by ID
	vo, err := svc.GetRoleByID(id)
	assert.NoError(t, err)
	assert.Equal(t, id, vo.ID)
	assert.Equal(t, "DetailRole", vo.Name)
	assert.NotNil(t, vo.Desc)
	assert.Equal(t, "Detailed role", *vo.Desc)
}

func TestRoleServiceIntegration_GetRoleByIDNotFound(t *testing.T) {
	svc := newTestRoleService(t)

	_, err := svc.GetRoleByID(9999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "role not found")
}

func TestRoleServiceIntegration_QueryRoles(t *testing.T) {
	svc := newTestRoleService(t)
	repo := repository.NewRoleRepository(testDB)

	// Create multiple roles
	zeroParent := int64(0)
	oneParent := int64(1)

	repo.Create(&role.SysRole{RoleName: "Admin", RoleCode: "admin", Status: role.StatusEnabled, ParentID: &zeroParent})
	repo.Create(&role.SysRole{RoleName: "User", RoleCode: "user", Status: role.StatusEnabled, ParentID: &oneParent})
	repo.Create(&role.SysRole{RoleName: "Disabled", RoleCode: "disabled", Status: role.StatusDisabled})

	// Query with keyword
	keyword := "Admin"
	result, err := svc.QueryRoles(&role.RoleQueryRequest{Keyword: &keyword})
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "Admin", result[0].Name)
	assert.True(t, result[0].Root) // parent_id = 0

	// Query without keyword (should return at least 2 enabled roles)
	allResult, err := svc.QueryRoles(&role.RoleQueryRequest{})
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(allResult), 2)

	// Query with non-matching keyword
	emptyResult, err := svc.QueryRoles(&role.RoleQueryRequest{Keyword: pStr("NonExist")})
	assert.NoError(t, err)
	assert.Len(t, emptyResult, 0)
}

func TestRoleServiceIntegration_QueryRolesPage(t *testing.T) {
	svc := newTestRoleService(t)
	repo := repository.NewRoleRepository(testDB)

	systemType := "system"
	customType := "custom"
	zeroParent := int64(0)

	require.NoError(t, repo.Create(&role.SysRole{RoleName: "Admin Root", RoleCode: "admin-root", RoleType: &systemType, Status: role.StatusEnabled, ParentID: &zeroParent}))
	require.NoError(t, repo.Create(&role.SysRole{RoleName: "Admin Custom", RoleCode: "admin-custom", RoleType: &customType, Status: role.StatusEnabled}))
	require.NoError(t, repo.Create(&role.SysRole{RoleName: "Viewer", RoleCode: "viewer", RoleType: &systemType, Status: role.StatusEnabled}))

	keyword := "Admin"
	result, err := svc.QueryRolesPage(&role.RolePageRequest{Keyword: &keyword, Current: 1, Size: 10})
	require.NoError(t, err)
	assert.Equal(t, int64(2), result.Total)
	assert.Len(t, result.List, 2)

	result, err = svc.QueryRolesPage(&role.RolePageRequest{RoleType: &systemType, Current: 1, Size: 10})
	require.NoError(t, err)
	assert.Equal(t, int64(2), result.Total)
	assert.Len(t, result.List, 2)
	for _, item := range result.List {
		require.NotNil(t, item.RoleType)
		assert.Equal(t, systemType, *item.RoleType)
	}
}

func pStr(v string) *string {
	return &v
}

// Test MountUsers
func TestRoleServiceIntegration_MountUsers(t *testing.T) {
	svc := newTestRoleService(t)
	userRepo := repository.NewUserRepository(testDB)
	roleRepo := repository.NewRoleRepository(testDB)

	// Create a test user
	testUser := &user.SysUser{
		Username: "mount_test_user",
		NickName: "Mount Test User",
		Email:    pStr("mount@test.com"),
		Status:   1,
	}
	err := userRepo.Create(testUser)
	assert.NoError(t, err)

	// Create a test role
	roleID, err := svc.CreateRole(&role.RoleCreator{Name: "MountTestRole"}, "tester")
	assert.NoError(t, err)

	// Mount users
	req := &role.MountUserRequest{
		Rid:   roleID,
		OrgId: 1,
		Uids:  []int64{testUser.UserID},
	}
	err = svc.MountUsers(req)
	assert.NoError(t, err)

	// Verify user is bound to role
	roleIDs, err := roleRepo.GetUserRoleIDs(testUser.UserID)
	assert.NoError(t, err)
	assert.Contains(t, roleIDs, roleID)
}

// Test MountExternalUser
func TestRoleServiceIntegration_MountExternalUser(t *testing.T) {
	svc := newTestRoleService(t)
	userRepo := repository.NewUserRepository(testDB)
	roleRepo := repository.NewRoleRepository(testDB)

	// Create a test user
	testUser := &user.SysUser{
		Username: "external_mount_user",
		NickName: "External Mount User",
		Email:    pStr("external@test.com"),
		Status:   1,
	}
	err := userRepo.Create(testUser)
	assert.NoError(t, err)

	// Create a test role
	roleID, err := svc.CreateRole(&role.RoleCreator{Name: "ExternalMountRole"}, "tester")
	assert.NoError(t, err)

	// Mount external user
	req := &role.MountExternalUserRequest{
		Rid: roleID,
		Uid: testUser.UserID,
	}
	err = svc.MountExternalUser(req, 2)
	assert.NoError(t, err)

	// Verify user is bound to role
	roleIDs, err := roleRepo.GetUserRoleIDs(testUser.UserID)
	assert.NoError(t, err)
	assert.Contains(t, roleIDs, roleID)
}

// Test UnmountUser
func TestRoleServiceIntegration_UnmountUser(t *testing.T) {
	svc := newTestRoleService(t)
	userRepo := repository.NewUserRepository(testDB)
	roleRepo := repository.NewRoleRepository(testDB)

	// Create a test user
	testUser := &user.SysUser{
		Username: "unmount_test_user",
		NickName: "Unmount Test User",
		Email:    pStr("unmount@test.com"),
		Status:   1,
	}
	err := userRepo.Create(testUser)
	assert.NoError(t, err)

	// Create two roles so user can be unmounted from one
	roleID1, err := svc.CreateRole(&role.RoleCreator{Name: "UnmountRole1"}, "tester")
	assert.NoError(t, err)
	roleID2, err := svc.CreateRole(&role.RoleCreator{Name: "UnmountRole2"}, "tester")
	assert.NoError(t, err)

	// Mount user to both roles
	svc.MountUsers(&role.MountUserRequest{Rid: roleID1, OrgId: 1, Uids: []int64{testUser.UserID}})
	svc.MountUsers(&role.MountUserRequest{Rid: roleID2, OrgId: 1, Uids: []int64{testUser.UserID}})

	// Unmount from one role
	req := &role.UnmountUserRequest{
		Rid: roleID1,
		Uid: testUser.UserID,
	}
	err = svc.UnmountUser(req)
	assert.NoError(t, err)

	// Verify user still has one role
	roleIDs, err := roleRepo.GetUserRoleIDs(testUser.UserID)
	assert.NoError(t, err)
	assert.NotContains(t, roleIDs, roleID1)
	assert.Contains(t, roleIDs, roleID2)
}

// Test UnmountUser with last role (should fail)
func TestRoleServiceIntegration_UnmountUserLastRole(t *testing.T) {
	svc := newTestRoleService(t)
	userRepo := repository.NewUserRepository(testDB)

	// Create a test user
	testUser := &user.SysUser{
		Username: "last_role_user",
		NickName: "Last Role User",
		Email:    pStr("last@test.com"),
		Status:   1,
	}
	err := userRepo.Create(testUser)
	assert.NoError(t, err)

	// Create one role
	roleID, err := svc.CreateRole(&role.RoleCreator{Name: "OnlyRole"}, "tester")
	assert.NoError(t, err)

	// Mount user to only role
	svc.MountUsers(&role.MountUserRequest{Rid: roleID, OrgId: 1, Uids: []int64{testUser.UserID}})

	// Try to unmount (should fail)
	req := &role.UnmountUserRequest{
		Rid: roleID,
		Uid: testUser.UserID,
	}
	err = svc.UnmountUser(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot remove user's last role")
}

// Test BeforeUnmountInfo
func TestRoleServiceIntegration_BeforeUnmountInfo(t *testing.T) {
	svc := newTestRoleService(t)
	userRepo := repository.NewUserRepository(testDB)

	// Create a test user
	testUser := &user.SysUser{
		Username: "before_unmount_user",
		NickName: "Before Unmount User",
		Email:    pStr("before@test.com"),
		Status:   1,
	}
	err := userRepo.Create(testUser)
	assert.NoError(t, err)

	// Create two roles
	roleID1, _ := svc.CreateRole(&role.RoleCreator{Name: "BeforeRole1"}, "tester")
	roleID2, _ := svc.CreateRole(&role.RoleCreator{Name: "BeforeRole2"}, "tester")

	// Mount user to both roles
	svc.MountUsers(&role.MountUserRequest{Rid: roleID1, OrgId: 1, Uids: []int64{testUser.UserID}})
	svc.MountUsers(&role.MountUserRequest{Rid: roleID2, OrgId: 1, Uids: []int64{testUser.UserID}})

	// Check before unmount info
	req := &role.UnmountUserRequest{
		Rid: roleID1,
		Uid: testUser.UserID,
	}
	count, err := svc.BeforeUnmountInfo(req)
	assert.NoError(t, err)
	assert.Equal(t, 2, count)
}

// Test SearchExternalUser
func TestRoleServiceIntegration_SearchExternalUser(t *testing.T) {
	svc := newTestRoleService(t)
	userRepo := repository.NewUserRepository(testDB)

	// Create test users
	testUser1 := &user.SysUser{
		Username: "search_ext_user1",
		NickName: "Search External User1",
		Email:    pStr("search1@test.com"),
		Status:   1,
	}
	testUser2 := &user.SysUser{
		Username: "search_ext_user2",
		NickName: "Search External User2",
		Email:    pStr("search2@test.com"),
		Status:   1,
	}
	userRepo.Create(testUser1)
	userRepo.Create(testUser2)

	// Search with keyword
	result, err := svc.SearchExternalUser("search_ext", 999)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(result), 1)
}

// Test OptionForUser
func TestRoleServiceIntegration_OptionForUser(t *testing.T) {
	svc := newTestRoleService(t)

	// Create some roles
	svc.CreateRole(&role.RoleCreator{Name: "OptionRole1"}, "tester")
	svc.CreateRole(&role.RoleCreator{Name: "OptionRole2"}, "tester")

	// Get option for user
	req := &role.RoleRequest{
		Keyword: pStr("Option"),
	}
	result, err := svc.OptionForUser(req, 1)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(result), 1)
}

// Test SelectedForUser
func TestRoleServiceIntegration_SelectedForUser(t *testing.T) {
	svc := newTestRoleService(t)
	userRepo := repository.NewUserRepository(testDB)

	// Create a test user
	testUser := &user.SysUser{
		Username: "selected_user",
		NickName: "Selected User",
		Email:    pStr("selected@test.com"),
		Status:   1,
	}
	err := userRepo.Create(testUser)
	assert.NoError(t, err)

	// Create a role
	roleID, err := svc.CreateRole(&role.RoleCreator{Name: "SelectedRole"}, "tester")
	assert.NoError(t, err)

	// Mount user to role
	svc.MountUsers(&role.MountUserRequest{Rid: roleID, OrgId: 1, Uids: []int64{testUser.UserID}})

	// Get selected roles for user
	req := &role.RoleRequest{
		Uid: &testUser.UserID,
	}
	result, err := svc.SelectedForUser(req)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(result), 1)

	// Find the role we created
	found := false
	for _, r := range result {
		if r.ID == roleID {
			found = true
			break
		}
	}
	assert.True(t, found, "Created role should be in selected roles")
}

// Test SelectedForUser without uid (should fail)
func TestRoleServiceIntegration_SelectedForUserNoUid(t *testing.T) {
	svc := newTestRoleService(t)

	req := &role.RoleRequest{}
	_, err := svc.SelectedForUser(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "uid is required")
}

// Test SelectedForUser with no roles
func TestRoleServiceIntegration_SelectedForUserNoRoles(t *testing.T) {
	svc := newTestRoleService(t)
	userRepo := repository.NewUserRepository(testDB)

	// Create a test user without any roles
	testUser := &user.SysUser{
		Username: "no_roles_user",
		NickName: "No Roles User",
		Email:    pStr("noroles@test.com"),
		Status:   1,
	}
	err := userRepo.Create(testUser)
	assert.NoError(t, err)

	// Get selected roles for user
	req := &role.RoleRequest{
		Uid: &testUser.UserID,
	}
	result, err := svc.SelectedForUser(req)
	assert.NoError(t, err)
	assert.Len(t, result, 0)
}
