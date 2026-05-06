package service

import (
	"errors"
	"testing"
	"time"

	"dataease/backend/internal/domain/permission"
	"dataease/backend/internal/domain/role"
	"dataease/backend/internal/domain/user"
	"dataease/backend/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupRoleServiceTest(t *testing.T) (*RoleService, *repository.RoleRepository) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&role.SysRole{}, &user.SysUserRole{}, &user.SysUser{}, &permission.SysRolePerm{}))

	repo := repository.NewRoleRepository(db)
	svc := NewRoleService(repo, nil, nil, nil)
	svc.SetResourcePermissionRepository(repository.NewResourcePermissionRepository(db))
	return svc, repo
}

func setupRoleServiceWithReposTest(t *testing.T) (*RoleService, *repository.RoleRepository, *repository.UserRepository, *repository.UserRoleRepository) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&role.SysRole{}, &user.SysUserRole{}, &user.SysUser{}, &permission.SysRolePerm{}))

	roleRepo := repository.NewRoleRepository(db)
	userRepo := repository.NewUserRepository(db)
	userRoleRepo := repository.NewUserRoleRepository(db)
	svc := NewRoleService(roleRepo, userRepo, userRoleRepo, nil)
	svc.SetResourcePermissionRepository(repository.NewResourcePermissionRepository(db))
	return svc, roleRepo, userRepo, userRoleRepo
}

func setupRoleServiceWithReposAndDBTest(t *testing.T) (*RoleService, *repository.RoleRepository, *repository.UserRepository, *repository.UserRoleRepository, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&role.SysRole{}, &user.SysUserRole{}, &user.SysUser{}, &permission.SysRolePerm{}))

	roleRepo := repository.NewRoleRepository(db)
	userRepo := repository.NewUserRepository(db)
	userRoleRepo := repository.NewUserRoleRepository(db)
	svc := NewRoleService(roleRepo, userRepo, userRoleRepo, nil)
	svc.SetResourcePermissionRepository(repository.NewResourcePermissionRepository(db))
	return svc, roleRepo, userRepo, userRoleRepo, db
}

func setupRoleServiceWithoutUserRoleTableTest(t *testing.T) (*RoleService, *repository.RoleRepository, *repository.UserRepository, *repository.UserRoleRepository) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&role.SysRole{}, &user.SysUser{}, &permission.SysRolePerm{}))

	roleRepo := repository.NewRoleRepository(db)
	userRepo := repository.NewUserRepository(db)
	userRoleRepo := repository.NewUserRoleRepository(db)
	svc := NewRoleService(roleRepo, userRepo, userRoleRepo, nil)
	svc.SetResourcePermissionRepository(repository.NewResourcePermissionRepository(db))
	return svc, roleRepo, userRepo, userRoleRepo
}

func setupRoleServiceWithoutRoleTableTest(t *testing.T) (*RoleService, *repository.RoleRepository, *repository.UserRepository, *repository.UserRoleRepository) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&user.SysUserRole{}, &user.SysUser{}, &permission.SysRolePerm{}))

	roleRepo := repository.NewRoleRepository(db)
	userRepo := repository.NewUserRepository(db)
	userRoleRepo := repository.NewUserRoleRepository(db)
	svc := NewRoleService(roleRepo, userRepo, userRoleRepo, nil)
	svc.SetResourcePermissionRepository(repository.NewResourcePermissionRepository(db))
	return svc, roleRepo, userRepo, userRoleRepo
}

func setupClosedRoleServiceTest(t *testing.T) (*RoleService, *repository.RoleRepository) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&role.SysRole{}, &user.SysUserRole{}, &user.SysUser{}, &permission.SysRolePerm{}))
	repo := repository.NewRoleRepository(db)
	svc := NewRoleService(repo, nil, nil, nil)
	svc.SetResourcePermissionRepository(repository.NewResourcePermissionRepository(db))
	sqlDB, err := db.DB()
	require.NoError(t, err)
	require.NoError(t, sqlDB.Close())
	return svc, repo
}

func setupClosedRoleServiceWithReposTest(t *testing.T) (*RoleService, *repository.RoleRepository, *repository.UserRepository, *repository.UserRoleRepository) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&role.SysRole{}, &user.SysUserRole{}, &user.SysUser{}, &permission.SysRolePerm{}))
	roleRepo := repository.NewRoleRepository(db)
	userRepo := repository.NewUserRepository(db)
	userRoleRepo := repository.NewUserRoleRepository(db)
	svc := NewRoleService(roleRepo, userRepo, userRoleRepo, nil)
	svc.SetResourcePermissionRepository(repository.NewResourcePermissionRepository(db))
	sqlDB, err := db.DB()
	require.NoError(t, err)
	require.NoError(t, sqlDB.Close())
	return svc, roleRepo, userRepo, userRoleRepo
}

func seedRole(t *testing.T, repo *repository.RoleRepository, name, code string, roleType *string, createTime time.Time) {
	t.Helper()

	require.NoError(t, repo.Create(&role.SysRole{
		RoleName:   name,
		RoleCode:   code,
		RoleType:   roleType,
		Status:     role.StatusEnabled,
		CreateTime: &createTime,
	}))
}

func seedUser(t *testing.T, repo *repository.UserRepository, username string, email *string) *user.SysUser {
	t.Helper()

	u := &user.SysUser{
		Username: username,
		NickName: username,
		Email:    email,
		Status:   user.StatusEnabled,
		DelFlag:  user.DelFlagNormal,
	}
	require.NoError(t, repo.Create(u))
	return u
}

func TestRoleService_QueryRolesPage_DefaultPaging(t *testing.T) {
	svc, repo := setupRoleServiceTest(t)
	systemType := "system"
	customType := "custom"

	seedRole(t, repo, "Admin", "admin", &systemType, time.Unix(300, 0))
	seedRole(t, repo, "Viewer", "viewer", &customType, time.Unix(200, 0))
	seedRole(t, repo, "Operator", "operator", &systemType, time.Unix(100, 0))

	result, err := svc.QueryRolesPage(&role.RolePageRequest{Current: 0, Size: 0})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, int64(3), result.Total)
	assert.Equal(t, 1, result.Current)
	assert.Equal(t, 10, result.Size)
	assert.Len(t, result.List, 3)
	assert.ElementsMatch(t, []string{"Admin", "Viewer", "Operator"}, []string{result.List[0].Name, result.List[1].Name, result.List[2].Name})
	for _, item := range result.List {
		require.NotNil(t, item.RoleType)
	}
}

func TestRoleService_QueryRolesPage_WithKeywordAndRoleType(t *testing.T) {
	svc, repo := setupRoleServiceTest(t)
	systemType := "system"
	customType := "custom"

	seedRole(t, repo, "Admin Root", "admin-root", &systemType, time.Unix(300, 0))
	seedRole(t, repo, "Admin Custom", "admin-custom", &customType, time.Unix(200, 0))
	seedRole(t, repo, "Viewer", "viewer", &systemType, time.Unix(100, 0))

	keyword := "Admin"
	result, err := svc.QueryRolesPage(&role.RolePageRequest{Keyword: &keyword, Current: 1, Size: 10})
	require.NoError(t, err)
	assert.Equal(t, int64(2), result.Total)
	assert.Len(t, result.List, 2)

	result, err = svc.QueryRolesPage(&role.RolePageRequest{Keyword: &keyword, RoleType: &systemType, Current: 1, Size: 10})
	require.NoError(t, err)
	assert.Equal(t, int64(1), result.Total)
	assert.Len(t, result.List, 1)
	assert.Equal(t, "Admin Root", result.List[0].Name)
	require.NotNil(t, result.List[0].RoleType)
	assert.Equal(t, systemType, *result.List[0].RoleType)
}

func TestRoleService_QueryRolesPage_OutOfRangePage(t *testing.T) {
	svc, repo := setupRoleServiceTest(t)
	systemType := "system"

	seedRole(t, repo, "Admin", "admin", &systemType, time.Unix(300, 0))
	seedRole(t, repo, "Viewer", "viewer", &systemType, time.Unix(200, 0))

	result, err := svc.QueryRolesPage(&role.RolePageRequest{Current: 3, Size: 1})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, int64(2), result.Total)
	assert.Equal(t, 3, result.Current)
	assert.Equal(t, 1, result.Size)
	assert.Len(t, result.List, 0)
}

func TestRoleService_CreateRole_AllowsBuiltInRootParent(t *testing.T) {
	svc, repo := setupRoleServiceTest(t)
	roleType := role.RoleTypeSystem
	rootParent := int64(0)
	seed := &role.SysRole{RoleName: "System Root", RoleCode: "system-root", RoleType: &roleType, Status: role.StatusEnabled, ParentID: &rootParent}
	require.NoError(t, repo.Create(seed))

	createdID, err := svc.CreateRole(&role.RoleCreator{Name: "Child Custom", ParentID: &seed.RoleID}, "tester", 1)
	require.NoError(t, err)
	created, err := repo.GetByID(createdID)
	require.NoError(t, err)
	require.NotNil(t, created.ParentID)
	assert.Equal(t, seed.RoleID, *created.ParentID)
	require.NotNil(t, created.RoleType)
	assert.Equal(t, role.RoleTypeCustom, *created.RoleType)
}

func TestRoleService_CreateRole_RejectsCustomParent(t *testing.T) {
	svc, repo := setupRoleServiceTest(t)
	customType := role.RoleTypeCustom
	rootParent := int64(0)
	seed := &role.SysRole{RoleName: "Custom Parent", RoleCode: "custom-parent", RoleType: &customType, Status: role.StatusEnabled, ParentID: &rootParent}
	require.NoError(t, repo.Create(seed))

	_, err := svc.CreateRole(&role.RoleCreator{Name: "Invalid Child", ParentID: &seed.RoleID}, "tester", 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "custom role cannot be used as parent role")
}

func TestRoleService_CreateRole_RequiresName(t *testing.T) {
	svc, _ := setupRoleServiceTest(t)

	createdID, err := svc.CreateRole(&role.RoleCreator{}, "tester", 1)
	require.Error(t, err)
	assert.Zero(t, createdID)
	assert.Contains(t, err.Error(), "role name is required")
}

func TestRoleService_CreateRole_UsesFallbackFieldsAndStatus(t *testing.T) {
	svc, repo := setupRoleServiceTest(t)
	desc := "desc from fallback"
	customStatus := 2

	createdID, err := svc.CreateRole(&role.RoleCreator{Name: "Fallback Name", Desc: &desc, RoleKey: "fallback-key", Status: &customStatus}, "tester", 1)
	require.NoError(t, err)

	created, err := repo.GetByID(createdID)
	require.NoError(t, err)
	assert.Equal(t, "Fallback Name", created.RoleName)
	assert.Equal(t, "fallback-key", created.RoleCode)
	require.NotNil(t, created.RoleDesc)
	assert.Equal(t, desc, *created.RoleDesc)
	assert.Equal(t, customStatus, created.Status)
	require.NotNil(t, created.CreateBy)
	assert.Equal(t, "tester", *created.CreateBy)
	require.NotNil(t, created.DataScope)
	assert.Equal(t, role.DataScopeSelf, *created.DataScope)
}

func TestRoleService_CreateRole_RepoError(t *testing.T) {
	svc, _ := setupClosedRoleServiceTest(t)

	createdID, err := svc.CreateRole(&role.RoleCreator{Name: "Broken Create"}, "tester", 1)
	require.Error(t, err)
	assert.Zero(t, createdID)
	assert.Contains(t, err.Error(), "failed to create role")
}

func TestRoleService_EditRole_RejectsNonRootParent(t *testing.T) {
	svc, repo := setupRoleServiceTest(t)
	rootParent := int64(0)
	builtInType := role.RoleTypeOrganization
	builtInParent := &role.SysRole{RoleName: "BuiltIn Parent", RoleCode: "builtin-parent", RoleType: &builtInType, Status: role.StatusEnabled, ParentID: &rootParent}
	require.NoError(t, repo.Create(builtInParent))
	childOfBuiltIn := &role.SysRole{RoleName: "Custom Child", RoleCode: "custom-child", Status: role.StatusEnabled, ParentID: &builtInParent.RoleID}
	require.NoError(t, repo.Create(childOfBuiltIn))
	targetID, err := svc.CreateRole(&role.RoleCreator{Name: "Editable Role"}, "tester", 1)
	require.NoError(t, err)

	err = svc.EditRole(&role.RoleEditor{ID: targetID, ParentID: &childOfBuiltIn.RoleID}, "editor", 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parent role must be a built-in root role")
}

func TestRoleService_EditRole_RequiresID(t *testing.T) {
	svc, _ := setupRoleServiceTest(t)

	err := svc.EditRole(&role.RoleEditor{}, "editor", 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "role id is required")
}

func TestRoleService_EditRole_UsesRoleIDAndUpdatesFields(t *testing.T) {
	svc, repo := setupRoleServiceTest(t)
	parentType := role.RoleTypeOrganization
	rootParent := int64(0)
	parent := &role.SysRole{RoleName: "BuiltIn Parent", RoleCode: "builtin-parent", RoleType: &parentType, Status: role.StatusEnabled, ParentID: &rootParent}
	require.NoError(t, repo.Create(parent))

	initialDesc := "old desc"
	target := &role.SysRole{RoleName: "Old Name", RoleCode: "old-name", RoleDesc: &initialDesc, Status: role.StatusEnabled}
	require.NoError(t, repo.Create(target))

	newDesc := "new desc"
	disabled := role.StatusDisabled
	err := svc.EditRole(&role.RoleEditor{RoleID: target.RoleID, Name: "New Name", Desc: &newDesc, Status: &disabled, ParentID: &parent.RoleID}, "editor", 1)
	require.NoError(t, err)

	updated, err := repo.GetByID(target.RoleID)
	require.NoError(t, err)
	assert.Equal(t, "New Name", updated.RoleName)
	require.NotNil(t, updated.RoleDesc)
	assert.Equal(t, newDesc, *updated.RoleDesc)
	assert.Equal(t, disabled, updated.Status)
	require.NotNil(t, updated.ParentID)
	assert.Equal(t, parent.RoleID, *updated.ParentID)
	require.NotNil(t, updated.RoleType)
	assert.Equal(t, role.RoleTypeCustom, *updated.RoleType)
	require.NotNil(t, updated.UpdateBy)
	assert.Equal(t, "editor", *updated.UpdateBy)
	require.NotNil(t, updated.UpdateTime)
}

func TestRoleService_DeleteRole_Success(t *testing.T) {
	svc, repo := setupRoleServiceTest(t)
	seedRole(t, repo, "ToDelete", "to-delete", nil, time.Now())

	roles, _ := repo.Query("")
	require.Len(t, roles, 1)
	roleID := roles[0].RoleID

	err := svc.DeleteRole(roleID, 1)
	require.NoError(t, err)

	_, err = repo.GetByID(roleID)
	require.Error(t, err)
}

func TestRoleService_DeleteRole_NotFound(t *testing.T) {
	svc, _ := setupRoleServiceTest(t)

	err := svc.DeleteRole(99999, 1)
	require.NoError(t, err)
}

func TestRoleService_DeleteRole_BuiltInBlocked(t *testing.T) {
	svc, repo := setupRoleServiceTest(t)
	systemType := role.RoleTypeSystem
	seedRole(t, repo, "Admin", "admin", &systemType, time.Now())

	roles, _ := repo.Query("")
	require.Len(t, roles, 1)

	err := svc.DeleteRole(roles[0].RoleID, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot delete built-in role")
}

func TestRoleService_RejectsInvalidOrgContext(t *testing.T) {
	svc, repo, userRepo, _ := setupRoleServiceWithReposTest(t)
	seedRole(t, repo, "ScopedRole", "scoped-role", nil, time.Now())
	roles, err := repo.Query("")
	require.NoError(t, err)
	require.NotEmpty(t, roles)

	_, err = svc.CreateRole(&role.RoleCreator{Name: "NoOrg"}, "tester", 0)
	require.ErrorIs(t, err, ErrInvalidOrgContext)

	err = svc.EditRole(&role.RoleEditor{RoleID: roles[0].RoleID, Name: "Rename"}, "tester", 0)
	require.ErrorIs(t, err, ErrInvalidOrgContext)

	err = svc.DeleteRole(roles[0].RoleID, 0)
	require.ErrorIs(t, err, ErrInvalidOrgContext)

	u := seedUser(t, userRepo, "scoped-user", nil)
	err = svc.MountUsers(&role.MountUserRequest{Rid: roles[0].RoleID, Uids: []int64{u.UserID}, OrgId: 0})
	require.ErrorIs(t, err, ErrInvalidOrgContext)

	err = svc.UnmountUser(&role.UnmountUserRequest{Rid: roles[0].RoleID, Uid: u.UserID, OrgId: 0})
	require.ErrorIs(t, err, ErrInvalidOrgContext)

	_, err = svc.BeforeUnmountInfo(&role.UnmountUserRequest{Rid: roles[0].RoleID, Uid: u.UserID, OrgId: 0})
	require.ErrorIs(t, err, ErrInvalidOrgContext)
}

func TestRoleService_DeleteRole_CustomRoleAllowed(t *testing.T) {
	svc, repo := setupRoleServiceTest(t)
	customType := role.RoleTypeCustom
	seedRole(t, repo, "Custom", "custom", &customType, time.Now())

	roles, _ := repo.Query("")
	require.Len(t, roles, 1)
	roleID := roles[0].RoleID

	err := svc.DeleteRole(roleID, 1)
	require.NoError(t, err)

	_, err = repo.GetByID(roleID)
	require.Error(t, err)
}

func TestRoleService_DeleteRole_RepoError(t *testing.T) {
	svc, _ := setupClosedRoleServiceTest(t)

	err := svc.DeleteRole(1, 1)
	require.NoError(t, err) // GetByID fails on closed DB → treated as "not found"
}

func TestRoleService_GetRoleByID_Success(t *testing.T) {
	svc, repo := setupRoleServiceTest(t)
	seedRole(t, repo, "TestRole", "test-role", nil, time.Now())

	roles, _ := repo.Query("")
	roleID := roles[0].RoleID

	result, err := svc.GetRoleByID(roleID)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "TestRole", result.Name)
	assert.Equal(t, "test-role", result.Code)
}

func TestRoleService_GetRoleByID_NotFound(t *testing.T) {
	svc, _ := setupRoleServiceTest(t)

	result, err := svc.GetRoleByID(99999)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "role not found")
	assert.Nil(t, result)
}

func TestRoleService_QueryRoles_Success(t *testing.T) {
	svc, repo := setupRoleServiceTest(t)
	systemType := "system"
	seedRole(t, repo, "Admin", "admin", &systemType, time.Now())
	seedRole(t, repo, "Viewer", "viewer", nil, time.Now())

	keyword := "Admin"
	result, err := svc.QueryRoles(&role.RoleQueryRequest{Keyword: &keyword})
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "Admin", result[0].Name)
}

func TestRoleService_QueryRoles_NoKeyword(t *testing.T) {
	svc, repo := setupRoleServiceTest(t)
	systemType := "system"
	seedRole(t, repo, "Admin", "admin", &systemType, time.Now())
	seedRole(t, repo, "Viewer", "viewer", nil, time.Now())

	result, err := svc.QueryRoles(&role.RoleQueryRequest{})
	require.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestRoleService_QueryRoles_RootFlag(t *testing.T) {
	svc, repo := setupRoleServiceTest(t)
	rootParent := int64(0)
	nonRootParent := int64(1)

	rootRole := &role.SysRole{RoleName: "Root", RoleCode: "root", Status: role.StatusEnabled, ParentID: &rootParent}
	require.NoError(t, repo.Create(rootRole))
	childRole := &role.SysRole{RoleName: "Child", RoleCode: "child", Status: role.StatusEnabled, ParentID: &nonRootParent}
	require.NoError(t, repo.Create(childRole))

	result, err := svc.QueryRoles(&role.RoleQueryRequest{})
	require.NoError(t, err)
	require.Len(t, result, 2)

	var rootFound, childFound bool
	for _, r := range result {
		if r.Name == "Root" {
			rootFound = true
			assert.True(t, r.Root)
		}
		if r.Name == "Child" {
			childFound = true
			assert.False(t, r.Root)
		}
	}
	assert.True(t, rootFound)
	assert.True(t, childFound)
}

func TestRoleService_QueryRoles_RepoError(t *testing.T) {
	svc, _ := setupClosedRoleServiceTest(t)

	result, err := svc.QueryRoles(&role.RoleQueryRequest{})
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to query roles")
}

func TestRoleService_MountUsers_NilRepo(t *testing.T) {
	svc, _ := setupRoleServiceTest(t)

	err := svc.MountUsers(&role.MountUserRequest{Rid: 1, Uids: []int64{1, 2}, OrgId: 1})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "userRoleRepo not initialized")
}

func TestRoleService_MountExternalUser_NilRepo(t *testing.T) {
	svc, _ := setupRoleServiceTest(t)

	err := svc.MountExternalUser(&role.MountExternalUserRequest{Uid: 1, Rid: 1}, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "userRoleRepo not initialized")
}

func TestRoleService_MountExternalUser_InvalidOrgID(t *testing.T) {
	svc, _, _, _ := setupRoleServiceWithReposTest(t)

	err := svc.MountExternalUser(&role.MountExternalUserRequest{Uid: 1, Rid: 1}, 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "org id is required")
}

func TestRoleService_MountUsers_Success(t *testing.T) {
	svc, _, _, userRoleRepo := setupRoleServiceWithReposTest(t)

	err := svc.MountUsers(&role.MountUserRequest{Rid: 9, Uids: []int64{11, 12}, OrgId: 3})
	require.NoError(t, err)

	roles11, err := userRoleRepo.GetByUserID(11)
	require.NoError(t, err)
	require.Len(t, roles11, 1)
	assert.Equal(t, int64(9), roles11[0].RoleID)
	assert.Equal(t, int64(3), roles11[0].OrgID)

	roles12, err := userRoleRepo.GetByUserID(12)
	require.NoError(t, err)
	require.Len(t, roles12, 1)
	assert.Equal(t, int64(9), roles12[0].RoleID)
	assert.Equal(t, int64(3), roles12[0].OrgID)
}

func TestRoleService_MountUsers_DeduplicatesExistingBinding(t *testing.T) {
	svc, _, _, userRoleRepo := setupRoleServiceWithReposTest(t)

	err := svc.MountUsers(&role.MountUserRequest{Rid: 9, Uids: []int64{11}, OrgId: 3})
	require.NoError(t, err)
	err = svc.MountUsers(&role.MountUserRequest{Rid: 9, Uids: []int64{11}, OrgId: 3})
	require.NoError(t, err)

	roles11, err := userRoleRepo.GetByUserID(11)
	require.NoError(t, err)
	require.Len(t, roles11, 1)
}

func TestRoleService_MountUsers_CreateIfMissingError(t *testing.T) {
	svc, _, _, _ := setupRoleServiceWithoutUserRoleTableTest(t)

	err := svc.MountUsers(&role.MountUserRequest{Rid: 9, Uids: []int64{11}, OrgId: 3})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to bind user 11 to role")
}

func TestRoleService_MountExternalUser_UserAlreadyInOrg(t *testing.T) {
	svc, _, _, userRoleRepo := setupRoleServiceWithReposTest(t)
	require.NoError(t, userRoleRepo.Create(&user.SysUserRole{UserID: 21, RoleID: 1, OrgID: 7}))

	err := svc.MountExternalUser(&role.MountExternalUserRequest{Uid: 21, Rid: 8}, 7)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "user already belongs to target organization")
}

func TestRoleService_MountExternalUser_Success(t *testing.T) {
	svc, _, _, userRoleRepo := setupRoleServiceWithReposTest(t)

	err := svc.MountExternalUser(&role.MountExternalUserRequest{Uid: 22, Rid: 8}, 7)
	require.NoError(t, err)

	exists, err := userRoleRepo.Exists(22, 8, 7)
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestRoleService_MountExternalUser_IsUserInOrgError(t *testing.T) {
	svc, _, _, _ := setupRoleServiceWithoutUserRoleTableTest(t)

	err := svc.MountExternalUser(&role.MountExternalUserRequest{Uid: 22, Rid: 8}, 7)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to validate external user organization")
}

func TestRoleService_SearchExternalUser_EmptyKeyword(t *testing.T) {
	svc, _ := setupRoleServiceTest(t)

	result, err := svc.SearchExternalUser("   ", 1)
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestRoleService_SearchExternalUser_InvalidOrgID(t *testing.T) {
	svc, _ := setupRoleServiceTest(t)

	result, err := svc.SearchExternalUser("test", 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "org id is required")
	assert.Nil(t, result)
}

func TestRoleService_SearchExternalUser_NilUserRepo(t *testing.T) {
	svc, _ := setupRoleServiceTest(t)

	result, err := svc.SearchExternalUser("test", 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "userRepo not initialized")
	assert.Nil(t, result)
}

func TestRoleService_SearchExternalUser_Success(t *testing.T) {
	svc, _, userRepo, userRoleRepo := setupRoleServiceWithReposTest(t)
	allowedEmail := "allowed@example.com"
	excludedEmail := "excluded@example.com"
	allowedUser := seedUser(t, userRepo, "allowed", &allowedEmail)
	excludedUser := seedUser(t, userRepo, "excluded", &excludedEmail)
	require.NoError(t, userRoleRepo.Create(&user.SysUserRole{UserID: excludedUser.UserID, RoleID: 1, OrgID: 9}))

	result, err := svc.SearchExternalUser("allowed", 9)
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, allowedUser.UserID, result[0].Uid)
	assert.Equal(t, "allowed", result[0].Account)
	assert.Equal(t, "allowed", result[0].Name)
	require.NotNil(t, result[0].Email)
	assert.Equal(t, allowedEmail, *result[0].Email)
}

func TestRoleService_SelectedForUser_MissingUID(t *testing.T) {
	svc, _ := setupRoleServiceTest(t)

	result, err := svc.SelectedForUser(&role.RoleRequest{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "uid is required")
	assert.Nil(t, result)
}

func TestRoleService_OptionForUser_Success(t *testing.T) {
	svc, repo := setupRoleServiceTest(t)
	orgType := role.RoleTypeOrganization
	seedRole(t, repo, "Org Admin", "org-admin", &orgType, time.Unix(100, 0))

	result, err := svc.OptionForUser(&role.RoleRequest{}, 1)
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, "Org Admin", result[0].Name)
}

func TestRoleService_OptionForUser_FiltersByKeywordAndAssignedRole(t *testing.T) {
	svc, repo, _, _ := setupRoleServiceWithReposTest(t)
	orgType := role.RoleTypeOrganization
	orgRole := &role.SysRole{RoleName: "Org Admin", RoleCode: "org-admin", RoleType: &orgType, Status: role.StatusEnabled}
	assignedRole := &role.SysRole{RoleName: "Assigned Analyst", RoleCode: "assigned-analyst", Status: role.StatusEnabled}
	otherAssignedRole := &role.SysRole{RoleName: "Assigned Viewer", RoleCode: "assigned-viewer", Status: role.StatusEnabled}
	require.NoError(t, repo.Create(orgRole))
	require.NoError(t, repo.Create(assignedRole))
	require.NoError(t, repo.Create(otherAssignedRole))
	require.NoError(t, repo.BindUserRole(51, assignedRole.RoleID, 7))
	require.NoError(t, repo.BindUserRole(52, otherAssignedRole.RoleID, 7))
	keyword := "Analyst"

	result, err := svc.OptionForUser(&role.RoleRequest{Keyword: &keyword}, 7)
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, "Assigned Analyst", result[0].Name)
}

func TestRoleService_UnmountUser_RejectsLastRole(t *testing.T) {
	svc, repo, _, _ := setupRoleServiceWithReposTest(t)
	roleA := &role.SysRole{RoleName: "Role A", RoleCode: "role-a", Status: role.StatusEnabled}
	require.NoError(t, repo.Create(roleA))
	require.NoError(t, repo.BindUserRole(31, roleA.RoleID, 5))

	err := svc.UnmountUser(&role.UnmountUserRequest{Uid: 31, Rid: roleA.RoleID, OrgId: 5})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrLastRoleRemovalBlocked)
	assert.Equal(t, "cannot remove user's last role", err.Error())
}

func TestRoleService_UnmountUser_Success(t *testing.T) {
	svc, repo, _, _ := setupRoleServiceWithReposTest(t)
	roleA := &role.SysRole{RoleName: "Role A", RoleCode: "role-a", Status: role.StatusEnabled}
	roleB := &role.SysRole{RoleName: "Role B", RoleCode: "role-b", Status: role.StatusEnabled}
	require.NoError(t, repo.Create(roleA))
	require.NoError(t, repo.Create(roleB))
	require.NoError(t, repo.BindUserRole(32, roleA.RoleID, 5))
	require.NoError(t, repo.BindUserRole(32, roleB.RoleID, 5))

	err := svc.UnmountUser(&role.UnmountUserRequest{Uid: 32, Rid: roleA.RoleID, OrgId: 5})
	require.NoError(t, err)

	roleIDs, err := repo.GetUserRoleIDs(32)
	require.NoError(t, err)
	assert.Equal(t, []int64{roleB.RoleID}, roleIDs)
}

func TestRoleService_BeforeUnmountInfo_Success(t *testing.T) {
	svc, repo, _, _ := setupRoleServiceWithReposTest(t)
	roleA := &role.SysRole{RoleName: "Role A", RoleCode: "role-a", Status: role.StatusEnabled}
	roleB := &role.SysRole{RoleName: "Role B", RoleCode: "role-b", Status: role.StatusEnabled}
	require.NoError(t, repo.Create(roleA))
	require.NoError(t, repo.Create(roleB))
	require.NoError(t, repo.BindUserRole(33, roleA.RoleID, 5))
	require.NoError(t, repo.BindUserRole(33, roleB.RoleID, 5))

	count, err := svc.BeforeUnmountInfo(&role.UnmountUserRequest{Uid: 33, Rid: roleA.RoleID, OrgId: 5})
	require.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestRoleService_UnmountUser_CountUserRolesError(t *testing.T) {
	svc, _, _, _ := setupClosedRoleServiceWithReposTest(t)

	err := svc.UnmountUser(&role.UnmountUserRequest{Uid: 32, Rid: 1, OrgId: 5})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to check user role count")
	assert.False(t, errors.Is(err, ErrLastRoleRemovalBlocked))
}

func TestRoleService_BeforeUnmountInfo_CountError(t *testing.T) {
	svc, _, _, _ := setupClosedRoleServiceWithReposTest(t)

	count, err := svc.BeforeUnmountInfo(&role.UnmountUserRequest{Uid: 33, Rid: 1, OrgId: 5})
	require.Error(t, err)
	assert.Zero(t, count)
	assert.Contains(t, err.Error(), "failed to count user roles")
}

func TestRoleService_SelectedForUser_Empty(t *testing.T) {
	svc, _, _, _ := setupRoleServiceWithReposTest(t)
	uid := int64(41)

	result, err := svc.SelectedForUser(&role.RoleRequest{Uid: &uid})
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestRoleService_SelectedForUser_Success(t *testing.T) {
	svc, repo, _, _ := setupRoleServiceWithReposTest(t)
	orgType := role.RoleTypeOrganization
	roleA := &role.SysRole{RoleName: "Role A", RoleCode: "role-a", RoleType: &orgType, Status: role.StatusEnabled}
	roleB := &role.SysRole{RoleName: "Role B", RoleCode: "role-b", Status: role.StatusEnabled}
	require.NoError(t, repo.Create(roleA))
	require.NoError(t, repo.Create(roleB))
	require.NoError(t, repo.BindUserRole(42, roleA.RoleID, 5))
	require.NoError(t, repo.BindUserRole(42, roleB.RoleID, 5))
	uid := int64(42)

	result, err := svc.SelectedForUser(&role.RoleRequest{Uid: &uid})
	require.NoError(t, err)
	require.Len(t, result, 2)
	assert.ElementsMatch(t, []string{"Role A", "Role B"}, []string{result[0].Name, result[1].Name})
}

func TestRoleService_SelectedForUser_GetUserRoleIDsError(t *testing.T) {
	svc, _, _, _ := setupClosedRoleServiceWithReposTest(t)
	uid := int64(43)

	result, err := svc.SelectedForUser(&role.RoleRequest{Uid: &uid})
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get user role IDs")
}

func TestRoleService_SelectedForUser_GetRolesByIDsError(t *testing.T) {
	svc, repo, _, _ := setupRoleServiceWithoutRoleTableTest(t)
	require.NoError(t, repo.BindUserRole(44, 8, 5))
	uid := int64(44)

	result, err := svc.SelectedForUser(&role.RoleRequest{Uid: &uid})
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get roles by IDs")
}

func TestRoleService_CreateRoleWithInheritance_RequiresName(t *testing.T) {
	svc, _ := setupRoleServiceTest(t)

	createdID, err := svc.CreateRoleWithInheritance(&role.RoleCreator{}, nil, "tester")
	require.Error(t, err)
	assert.Zero(t, createdID)
	assert.Contains(t, err.Error(), "role name is required")
}

func TestRoleService_CreateRoleWithInheritance_Success(t *testing.T) {
	svc, repo := setupRoleServiceTest(t)
	parentType := role.RoleTypeSystem
	rootParent := int64(0)
	parent := &role.SysRole{RoleName: "Parent", RoleCode: "parent", RoleType: &parentType, Status: role.StatusEnabled, ParentID: &rootParent}
	require.NoError(t, repo.Create(parent))

	createdID, err := svc.CreateRoleWithInheritance(&role.RoleCreator{Name: "Child", RoleKey: "child-key"}, &parent.RoleID, "tester")
	require.NoError(t, err)

	created, err := repo.GetByID(createdID)
	require.NoError(t, err)
	assert.Equal(t, "Child", created.RoleName)
	assert.Equal(t, "child-key", created.RoleCode)
	require.NotNil(t, created.ParentID)
	assert.Equal(t, parent.RoleID, *created.ParentID)
}

func TestRoleService_CreateRoleWithInheritance_RejectsDisabledParent(t *testing.T) {
	svc, repo := setupRoleServiceTest(t)
	parentType := role.RoleTypeSystem
	rootParent := int64(0)
	disabledStatus := 2
	parent := &role.SysRole{RoleName: "Parent", RoleCode: "parent", RoleType: &parentType, Status: disabledStatus, ParentID: &rootParent}
	require.NoError(t, repo.Create(parent))

	createdID, err := svc.CreateRoleWithInheritance(&role.RoleCreator{Name: "Child"}, &parent.RoleID, "tester")
	require.Error(t, err)
	assert.Zero(t, createdID)
	assert.Contains(t, err.Error(), "parent role is disabled")
}

func TestRoleService_CreateRoleWithInheritance_UsesFallbackDesc(t *testing.T) {
	svc, repo := setupRoleServiceTest(t)
	desc := "inherited role description"

	createdID, err := svc.CreateRoleWithInheritance(&role.RoleCreator{Name: "Child With Desc", Desc: &desc, RoleKey: "child-with-desc"}, nil, "tester")
	require.NoError(t, err)

	created, err := repo.GetByID(createdID)
	require.NoError(t, err)
	assert.Equal(t, "Child With Desc", created.RoleName)
	require.NotNil(t, created.RoleDesc)
	assert.Equal(t, desc, *created.RoleDesc)
}

func TestRoleService_OptionForUser_RepoError(t *testing.T) {
	svc, _ := setupClosedRoleServiceTest(t)

	result, err := svc.OptionForUser(&role.RoleRequest{}, 1)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to query roles")
}

func TestRoleService_SearchExternalUser_RepoError(t *testing.T) {
	svc, _, _, _ := setupClosedRoleServiceWithReposTest(t)

	result, err := svc.SearchExternalUser("test", 1)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to search external users")
}

func TestRoleService_UnmountUser_UnbindError(t *testing.T) {
	svc, repo, _, _, db := setupRoleServiceWithReposAndDBTest(t)
	roleA := &role.SysRole{RoleName: "Role A", RoleCode: "role-a", Status: role.StatusEnabled}
	roleB := &role.SysRole{RoleName: "Role B", RoleCode: "role-b", Status: role.StatusEnabled}
	require.NoError(t, repo.Create(roleA))
	require.NoError(t, repo.Create(roleB))
	require.NoError(t, repo.BindUserRole(99, roleA.RoleID, 5))
	require.NoError(t, repo.BindUserRole(99, roleB.RoleID, 5))
	require.NoError(t, db.Exec(`
		CREATE TRIGGER sys_user_role_delete_fail
		BEFORE DELETE ON sys_user_role
		BEGIN
			SELECT RAISE(FAIL, 'forced unbind failure');
		END;
	`).Error)

	err := svc.UnmountUser(&role.UnmountUserRequest{Uid: 99, Rid: roleA.RoleID, OrgId: 5})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unbind user from role")
	count, countErr := repo.CountUserRoles(99)
	require.NoError(t, countErr)
	assert.Equal(t, int64(2), count)
}

func TestRoleService_BeforeUnmountInfo_SingleRole(t *testing.T) {
	svc, repo, _, _ := setupRoleServiceWithReposTest(t)
	roleA := &role.SysRole{RoleName: "Single Role", RoleCode: "single-role", Status: role.StatusEnabled}
	require.NoError(t, repo.Create(roleA))
	require.NoError(t, repo.BindUserRole(88, roleA.RoleID, 5))

	count, err := svc.BeforeUnmountInfo(&role.UnmountUserRequest{Uid: 88, Rid: roleA.RoleID, OrgId: 5})
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestRoleService_CreateRoleWithInheritance_RepoError(t *testing.T) {
	svc, _ := setupClosedRoleServiceTest(t)

	createdID, err := svc.CreateRoleWithInheritance(&role.RoleCreator{Name: "Child"}, nil, "tester")
	require.Error(t, err)
	assert.Zero(t, createdID)
	assert.Contains(t, err.Error(), "failed to create role")
}

func TestRoleService_ValidatePermissionInheritance_RootRole(t *testing.T) {
	svc, repo := setupRoleServiceTest(t)
	rootParent := int64(0)
	rle := &role.SysRole{RoleName: "Root", RoleCode: "root", Status: role.StatusEnabled, ParentID: &rootParent}
	require.NoError(t, repo.Create(rle))

	err := svc.ValidatePermissionInheritance(rle.RoleID, []int64{1, 2, 3})
	require.NoError(t, err)
}

func TestRoleService_ValidatePermissionInheritance_WithParent(t *testing.T) {
	svc, repo, _, _, db := setupRoleServiceWithReposAndDBTest(t)
	parentType := role.RoleTypeSystem
	rootParent := int64(0)
	parent := &role.SysRole{RoleName: "Parent", RoleCode: "parent", RoleType: &parentType, Status: role.StatusEnabled, ParentID: &rootParent}
	require.NoError(t, repo.Create(parent))
	child := &role.SysRole{RoleName: "Child", RoleCode: "child", Status: role.StatusEnabled, ParentID: &parent.RoleID}
	require.NoError(t, repo.Create(child))
	require.NoError(t, db.Create(&permission.SysRolePerm{RoleID: parent.RoleID, PermID: 5}).Error)
	require.NoError(t, db.Create(&permission.SysRolePerm{RoleID: parent.RoleID, PermID: 6}).Error)

	err := svc.ValidatePermissionInheritance(child.RoleID, []int64{5, 6})
	require.NoError(t, err)
}

func TestRoleService_ValidatePermissionInheritance_RejectsPermissionOutsideParentScope(t *testing.T) {
	svc, repo, _, _, db := setupRoleServiceWithReposAndDBTest(t)
	parentType := role.RoleTypeSystem
	rootParent := int64(0)
	parent := &role.SysRole{RoleName: "Parent", RoleCode: "parent", RoleType: &parentType, Status: role.StatusEnabled, ParentID: &rootParent}
	require.NoError(t, repo.Create(parent))
	child := &role.SysRole{RoleName: "Child", RoleCode: "child", Status: role.StatusEnabled, ParentID: &parent.RoleID}
	require.NoError(t, repo.Create(child))
	require.NoError(t, db.Create(&permission.SysRolePerm{RoleID: parent.RoleID, PermID: 5}).Error)

	err := svc.ValidatePermissionInheritance(child.RoleID, []int64{5, 6})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "permission inheritance violation")
}

func TestRoleService_ValidatePermissionInheritance_RoleNotFound(t *testing.T) {
	svc, _ := setupRoleServiceTest(t)

	err := svc.ValidatePermissionInheritance(99999, []int64{1})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "role not found")
}

func TestEditRole_BuiltInSystemRoleProtected(t *testing.T) {
	svc, repo := setupRoleServiceTest(t)
	systemType := role.RoleTypeSystem
	seedRole(t, repo, "SystemAdmin", "system-admin", &systemType, time.Now())

	roles, err := repo.Query("")
	require.NoError(t, err)
	require.Len(t, roles, 1)

	err = svc.EditRole(&role.RoleEditor{RoleID: roles[0].RoleID, Name: "HackedName"}, "attacker", 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot edit built-in system role")
}
