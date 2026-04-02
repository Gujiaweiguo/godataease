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

// Helper function to create RoleService with all dependencies
func newTestRoleService(t *testing.T) *RoleService {
	cleanupTables(&role.SysRole{}, &user.SysUser{}, &user.SysUserRole{}, &permission.SysRolePerm{})

	repo := repository.NewRoleRepository(testDB)
	userRepo := repository.NewUserRepository(testDB)
	userRoleRepo := repository.NewUserRoleRepository(testDB)

	svc := NewRoleService(repo, userRepo, userRoleRepo)
	svc.SetResourcePermissionRepository(repository.NewResourcePermissionRepository(testDB))
	return svc
}

func TestRoleServiceIntegration_Create(t *testing.T) {
	svc := newTestRoleService(t)
	repo := repository.NewRoleRepository(testDB)

	desc := "Test role description"
	req := &role.RoleCreator{
		Name: "TestRole",
		Desc: &desc,
	}

	id, err := svc.CreateRole(req, "tester", 0)
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
	id, _ := svc.CreateRole(&role.RoleCreator{Name: "ToEdit", Desc: &initialDesc}, "creator", 0)

	// Edit
	newDesc := "Updated description"
	err := svc.EditRole(&role.RoleEditor{
		ID:   id,
		Name: "EditedRole",
		Desc: &newDesc,
	}, "editor", 0)
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

	err := svc.EditRole(&role.RoleEditor{ID: 9999, Name: "NotFound"}, "editor", 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "role not found")
}

func TestRoleServiceIntegration_CreateRole_WithParent(t *testing.T) {
	svc := newTestRoleService(t)
	repo := repository.NewRoleRepository(testDB)

	parentID, err := svc.CreateRole(&role.RoleCreator{Name: "RootParentRole"}, "tester", 0)
	require.NoError(t, err)

	childID, err := svc.CreateRole(&role.RoleCreator{Name: "ChildCustomRole", ParentID: &parentID}, "tester", 0)
	require.NoError(t, err)

	child, err := repo.GetByID(childID)
	require.NoError(t, err)
	require.NotNil(t, child.ParentID)
	assert.Equal(t, parentID, *child.ParentID)
	if child.RoleType != nil {
		assert.Equal(t, role.RoleTypeCustom, *child.RoleType)
	}
}

func TestRoleServiceIntegration_EditRole_WithParent(t *testing.T) {
	svc := newTestRoleService(t)
	repo := repository.NewRoleRepository(testDB)

	parentID, err := svc.CreateRole(&role.RoleCreator{Name: "EditParentRole"}, "tester", 0)
	require.NoError(t, err)
	childID, err := svc.CreateRole(&role.RoleCreator{Name: "EditChildRole"}, "tester", 0)
	require.NoError(t, err)

	err = svc.EditRole(&role.RoleEditor{ID: childID, Name: "EditChildRole", ParentID: &parentID}, "tester", 0)
	require.NoError(t, err)

	child, err := repo.GetByID(childID)
	require.NoError(t, err)
	require.NotNil(t, child.ParentID)
	assert.Equal(t, parentID, *child.ParentID)
	if child.RoleType != nil {
		assert.Equal(t, role.RoleTypeCustom, *child.RoleType)
	}
}

func TestRoleServiceIntegration_EditRole_ClearParent(t *testing.T) {
	svc := newTestRoleService(t)
	repo := repository.NewRoleRepository(testDB)

	parentID, err := svc.CreateRole(&role.RoleCreator{Name: "ClearParentRoot"}, "tester", 0)
	require.NoError(t, err)
	childID, err := svc.CreateRole(&role.RoleCreator{Name: "ClearParentChild", ParentID: &parentID}, "tester", 0)
	require.NoError(t, err)

	zeroParent := int64(0)
	err = svc.EditRole(&role.RoleEditor{ID: childID, Name: "ClearParentChild", ParentID: &zeroParent}, "tester", 0)
	require.NoError(t, err)

	child, err := repo.GetByID(childID)
	require.NoError(t, err)
	require.NotNil(t, child.ParentID)
	assert.Equal(t, int64(0), *child.ParentID)
}

func TestRoleServiceIntegration_Delete(t *testing.T) {
	svc := newTestRoleService(t)

	// Create role
	id, _ := svc.CreateRole(&role.RoleCreator{Name: "ToDelete"}, "tester", 0)

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
	}, "tester", 0)

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
	roleID, err := svc.CreateRole(&role.RoleCreator{Name: "MountTestRole"}, "tester", 0)
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

func TestRoleServiceIntegration_MountUsers_Idempotent(t *testing.T) {
	svc := newTestRoleService(t)
	userRepo := repository.NewUserRepository(testDB)
	userRoleRepo := repository.NewUserRoleRepository(testDB)

	testUser := &user.SysUser{Username: "mount_idempotent_user", NickName: "Mount Idempotent User", Status: 1}
	require.NoError(t, userRepo.Create(testUser))
	roleID, err := svc.CreateRole(&role.RoleCreator{Name: "MountIdempotentRole"}, "tester", 0)
	require.NoError(t, err)

	req := &role.MountUserRequest{Rid: roleID, OrgId: 7, Uids: []int64{testUser.UserID}}
	require.NoError(t, svc.MountUsers(req))
	require.NoError(t, svc.MountUsers(req))

	bindings, err := userRoleRepo.GetByUserID(testUser.UserID)
	require.NoError(t, err)
	require.Len(t, bindings, 1)
	assert.Equal(t, int64(7), bindings[0].OrgID)
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
	roleID, err := svc.CreateRole(&role.RoleCreator{Name: "ExternalMountRole"}, "tester", 0)
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

func TestRoleServiceIntegration_MountExternalUser_RejectsUserAlreadyInTargetOrg(t *testing.T) {
	svc := newTestRoleService(t)
	userRepo := repository.NewUserRepository(testDB)
	userRoleRepo := repository.NewUserRoleRepository(testDB)

	testUser := &user.SysUser{Username: "external_existing_org_user", NickName: "External Existing Org User", Status: 1}
	require.NoError(t, userRepo.Create(testUser))
	roleID, err := svc.CreateRole(&role.RoleCreator{Name: "ExistingOrgRole"}, "tester", 0)
	require.NoError(t, err)
	require.NoError(t, userRoleRepo.Create(&user.SysUserRole{UserID: testUser.UserID, RoleID: roleID, OrgID: 2}))

	err = svc.MountExternalUser(&role.MountExternalUserRequest{Rid: roleID, Uid: testUser.UserID}, 2)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "user already belongs to target organization")
}

func TestRoleServiceIntegration_MountExternalUser_RequiresOrgID(t *testing.T) {
	svc := newTestRoleService(t)
	userRepo := repository.NewUserRepository(testDB)

	testUser := &user.SysUser{Username: "external_missing_org_user", NickName: "External Missing Org User", Status: 1}
	require.NoError(t, userRepo.Create(testUser))
	roleID, err := svc.CreateRole(&role.RoleCreator{Name: "MissingOrgRole"}, "tester", 0)
	require.NoError(t, err)

	err = svc.MountExternalUser(&role.MountExternalUserRequest{Rid: roleID, Uid: testUser.UserID}, 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "org id is required")
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
	roleID1, err := svc.CreateRole(&role.RoleCreator{Name: "UnmountRole1"}, "tester", 0)
	assert.NoError(t, err)
	roleID2, err := svc.CreateRole(&role.RoleCreator{Name: "UnmountRole2"}, "tester", 0)
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
	roleID, err := svc.CreateRole(&role.RoleCreator{Name: "OnlyRole"}, "tester", 0)
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
	roleID1, _ := svc.CreateRole(&role.RoleCreator{Name: "BeforeRole1"}, "tester", 0)
	roleID2, _ := svc.CreateRole(&role.RoleCreator{Name: "BeforeRole2"}, "tester", 0)

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

func TestRoleServiceIntegration_SearchExternalUser_UsesExactKeyword(t *testing.T) {
	svc := newTestRoleService(t)
	userRepo := repository.NewUserRepository(testDB)
	userRoleRepo := repository.NewUserRoleRepository(testDB)
	roleRepo := repository.NewRoleRepository(testDB)

	insideUser := &user.SysUser{Username: "inside-user", NickName: "Inside User", Email: pStr("inside@example.com"), Status: 1}
	outsideUser := &user.SysUser{Username: "outside-user", NickName: "Outside User", Email: pStr("outside@example.com"), Status: 1}
	require.NoError(t, userRepo.Create(insideUser))
	require.NoError(t, userRepo.Create(outsideUser))
	roleID, err := svc.CreateRole(&role.RoleCreator{Name: "SearchRole"}, "tester", 0)
	require.NoError(t, err)
	require.NoError(t, userRoleRepo.Create(&user.SysUserRole{UserID: insideUser.UserID, RoleID: roleID, OrgID: 4}))
	require.NoError(t, roleRepo.Delete(roleID))

	result, err := svc.SearchExternalUser("outside-user", 4)
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, outsideUser.UserID, result[0].Uid)

	result, err = svc.SearchExternalUser("outside", 4)
	require.NoError(t, err)
	assert.Len(t, result, 0)
}

func TestRoleServiceIntegration_SearchExternalUser_BlankKeywordReturnsEmpty(t *testing.T) {
	svc := newTestRoleService(t)

	result, err := svc.SearchExternalUser("   ", 4)
	require.NoError(t, err)
	assert.Len(t, result, 0)
}

func TestRoleServiceIntegration_SearchExternalUser_RequiresOrgID(t *testing.T) {
	svc := newTestRoleService(t)

	_, err := svc.SearchExternalUser("outside-user", 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "org id is required")
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

	result, err := svc.SearchExternalUser("search_ext_user1", 999)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, testUser1.UserID, result[0].Uid)
}

// Test OptionForUser
func TestRoleServiceIntegration_OptionForUser(t *testing.T) {
	svc := newTestRoleService(t)
	userRepo := repository.NewUserRepository(testDB)

	// Create some roles
	roleID1, err := svc.CreateRole(&role.RoleCreator{Name: "OptionRole1"}, "tester", 0)
	require.NoError(t, err)
	_, err = svc.CreateRole(&role.RoleCreator{Name: "OptionRole2"}, "tester", 0)
	require.NoError(t, err)

	testUser := &user.SysUser{
		Username: "option_user",
		NickName: "Option User",
		Status:   1,
	}
	require.NoError(t, userRepo.Create(testUser))
	require.NoError(t, svc.MountUsers(&role.MountUserRequest{Rid: roleID1, OrgId: 1, Uids: []int64{testUser.UserID}}))

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
	roleID, err := svc.CreateRole(&role.RoleCreator{Name: "SelectedRole"}, "tester", 0)
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

func TestRoleServiceIntegration_CreateRoleWithInheritance(t *testing.T) {
	svc := newTestRoleService(t)
	repo := repository.NewRoleRepository(testDB)

	parentID, err := svc.CreateRole(&role.RoleCreator{Name: "ParentRole"}, "tester", 0)
	require.NoError(t, err)

	childID, err := svc.CreateRoleWithInheritance(&role.RoleCreator{Name: "ChildRole"}, &parentID, "tester")
	require.NoError(t, err)

	child, err := repo.GetByID(childID)
	require.NoError(t, err)
	require.NotNil(t, child.ParentID)
	assert.Equal(t, parentID, *child.ParentID)
	assert.Equal(t, "ChildRole", child.RoleName)
}

func TestRoleServiceIntegration_CreateRoleWithInheritance_UsesRoleKeyAndDesc(t *testing.T) {
	svc := newTestRoleService(t)
	repo := repository.NewRoleRepository(testDB)

	desc := "Inherited role description"
	childID, err := svc.CreateRoleWithInheritance(&role.RoleCreator{RoleName: "InheritedRole", RoleKey: "inherited_role_key", RoleDesc: &desc}, nil, "tester")
	require.NoError(t, err)

	child, err := repo.GetByID(childID)
	require.NoError(t, err)
	assert.Equal(t, "InheritedRole", child.RoleName)
	assert.Equal(t, "inherited_role_key", child.RoleCode)
	if assert.NotNil(t, child.RoleDesc) {
		assert.Equal(t, desc, *child.RoleDesc)
	}
}

func TestRoleServiceIntegration_ValidatePermissionInheritance_WithParent(t *testing.T) {
	svc := newTestRoleService(t)
	resourcePermRepo := repository.NewResourcePermissionRepository(testDB)

	parentID, err := svc.CreateRole(&role.RoleCreator{Name: "InheritanceParent"}, "tester", 0)
	require.NoError(t, err)

	childID, err := svc.CreateRoleWithInheritance(&role.RoleCreator{Name: "InheritanceChild"}, &parentID, "tester")
	require.NoError(t, err)
	require.NoError(t, resourcePermRepo.GrantPermToRole(parentID, 1))
	require.NoError(t, resourcePermRepo.GrantPermToRole(parentID, 2))
	require.NoError(t, resourcePermRepo.GrantPermToRole(parentID, 3))

	require.NoError(t, svc.ValidatePermissionInheritance(childID, []int64{1, 2, 3}))
}

func TestRoleServiceIntegration_ValidatePermissionInheritance_RejectsExtraPermission(t *testing.T) {
	svc := newTestRoleService(t)
	resourcePermRepo := repository.NewResourcePermissionRepository(testDB)

	parentID, err := svc.CreateRole(&role.RoleCreator{Name: "InheritanceParentLimited"}, "tester", 0)
	require.NoError(t, err)
	childID, err := svc.CreateRoleWithInheritance(&role.RoleCreator{Name: "InheritanceChildLimited"}, &parentID, "tester")
	require.NoError(t, err)
	require.NoError(t, resourcePermRepo.GrantPermToRole(parentID, 1))

	err = svc.ValidatePermissionInheritance(childID, []int64{1, 2})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "permission inheritance violation")
}

func TestRoleServiceIntegration_CreateRoleWithInheritance_ParentNotFound(t *testing.T) {
	svc := newTestRoleService(t)

	missingParentID := int64(999999)
	_, err := svc.CreateRoleWithInheritance(&role.RoleCreator{Name: "BrokenChild"}, &missingParentID, "tester")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parent role not found")
}

func TestRoleServiceIntegration_CreateRoleWithInheritance_NameRequired(t *testing.T) {
	svc := newTestRoleService(t)

	_, err := svc.CreateRoleWithInheritance(&role.RoleCreator{}, nil, "tester")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "role name is required")
}

func TestRoleServiceIntegration_CreateRoleWithInheritance_ParentDisabled(t *testing.T) {
	svc := newTestRoleService(t)
	repo := repository.NewRoleRepository(testDB)

	parentID, err := svc.CreateRole(&role.RoleCreator{Name: "DisabledParent"}, "tester", 0)
	require.NoError(t, err)
	parent, err := repo.GetByID(parentID)
	require.NoError(t, err)
	parent.Status = role.StatusDisabled
	require.NoError(t, repo.Update(parent))

	_, err = svc.CreateRoleWithInheritance(&role.RoleCreator{Name: "ChildFromDisabledParent"}, &parent.RoleID, "tester")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parent role is disabled")
}

func TestRoleServiceIntegration_CreateRoleWithInheritance_CustomParentRejected(t *testing.T) {
	svc := newTestRoleService(t)

	rootParentID, err := svc.CreateRole(&role.RoleCreator{Name: "RootParentForCustom"}, "tester", 0)
	require.NoError(t, err)
	customParentID, err := svc.CreateRole(&role.RoleCreator{Name: "CustomParent", ParentID: &rootParentID}, "tester", 0)
	require.NoError(t, err)

	_, err = svc.CreateRoleWithInheritance(&role.RoleCreator{Name: "ChildFromCustomParent"}, &customParentID, "tester")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parent role must be a built-in root role")
}

func TestRoleServiceIntegration_EditRole_SystemRoleProtected(t *testing.T) {
	svc := newTestRoleService(t)
	repo := repository.NewRoleRepository(testDB)

	systemType := role.RoleTypeSystem
	require.NoError(t, repo.Create(&role.SysRole{RoleName: "System Admin", RoleCode: "system-admin", RoleType: &systemType, Status: role.StatusEnabled}))

	roles, err := repo.Query("")
	require.NoError(t, err)
	require.Len(t, roles, 1)
	systemRoleID := roles[0].RoleID

	err = svc.EditRole(&role.RoleEditor{RoleID: systemRoleID, Name: "Hacked Admin"}, "attacker", 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot edit built-in system role")
}

func TestRoleServiceIntegration_ValidatePermissionInheritance_NoParent(t *testing.T) {
	svc := newTestRoleService(t)

	roleID, err := svc.CreateRole(&role.RoleCreator{Name: "NoParentRole"}, "tester", 0)
	require.NoError(t, err)

	require.NoError(t, svc.ValidatePermissionInheritance(roleID, []int64{1}))
}

func TestRoleServiceIntegration_ValidatePermissionInheritance_RoleNotFound(t *testing.T) {
	svc := newTestRoleService(t)

	err := svc.ValidatePermissionInheritance(999999, []int64{1})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "role not found")
}
