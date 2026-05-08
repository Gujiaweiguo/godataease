package repository

import (
	"testing"

	"dataease/backend/internal/domain/permission"
	"dataease/backend/internal/domain/role"
	"dataease/backend/internal/domain/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// round6Setup creates an in-memory SQLite DB with all relevant tables migrated
// and returns a ready-to-use ResourcePermissionRepository.
func round6Setup(t *testing.T) *ResourcePermissionRepository {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&permission.SysPerm{},
		&permission.SysResource{},
		&permission.SysResourcePerm{},
		&permission.SysUserPerm{},
		&permission.SysRolePerm{},
		&user.SysUser{},
		&user.SysUserRole{},
		&role.SysRole{},
	))
	return NewResourcePermissionRepository(db)
}

// round6CreatePerm is a helper that creates a SysPerm and returns it (with ID populated).
func round6CreatePerm(t *testing.T, repo *ResourcePermissionRepository, name, key, permType string) *permission.SysPerm {
	t.Helper()
	p := &permission.SysPerm{
		PermName: name,
		PermKey:  key,
		PermType: permType,
		Status:   permission.StatusEnabled,
		DelFlag:  permission.DelFlagNormal,
	}
	require.NoError(t, repo.CreatePerm(p))
	require.Positive(t, p.PermID)
	return p
}

// round6CreateUser is a helper that creates a SysUser row directly.
func round6CreateUser(t *testing.T, repo *ResourcePermissionRepository, userID int64, username string) {
	t.Helper()
	require.NoError(t, repo.DB().Create(&user.SysUser{
		UserID:   userID,
		Username: username,
		Status:   user.StatusEnabled,
		DelFlag:  user.DelFlagNormal,
	}).Error)
}

// round6CreateRole is a helper that creates a SysRole row directly.
func round6CreateRole(t *testing.T, repo *ResourcePermissionRepository, roleID int64, name string) {
	t.Helper()
	require.NoError(t, repo.DB().Create(&role.SysRole{
		RoleID:   roleID,
		RoleName: name,
		Status:   role.StatusEnabled,
	}).Error)
}

// ==================== Constructor & DB ====================

func TestRound6Perm_NewResourcePermissionRepository(t *testing.T) {
	repo := round6Setup(t)
	require.NotNil(t, repo)
	require.NotNil(t, repo.db)
}

func TestRound6Perm_DB(t *testing.T) {
	// nil receiver returns nil
	var nilRepo *ResourcePermissionRepository
	assert.Nil(t, nilRepo.DB())

	// normal case returns underlying db
	repo := round6Setup(t)
	require.Same(t, repo.db, repo.DB())
}

// ==================== Perm CRUD ====================

func TestRound6Perm_GetPermByID(t *testing.T) {
	repo := round6Setup(t)

	// Not found
	_, err := repo.GetPermByID(99999)
	require.Error(t, err)

	p := round6CreatePerm(t, repo, "View Dashboard", "dashboard:view", permission.PermTypeData)
	got, err := repo.GetPermByID(p.PermID)
	require.NoError(t, err)
	assert.Equal(t, p.PermKey, got.PermKey)
	assert.Equal(t, p.PermName, got.PermName)
}

func TestRound6Perm_GetPermByKey(t *testing.T) {
	repo := round6Setup(t)

	// Not found
	_, err := repo.GetPermByKey("nonexistent:key")
	require.Error(t, err)

	p := round6CreatePerm(t, repo, "Edit Dataset", "dataset:edit", permission.PermTypeData)
	got, err := repo.GetPermByKey("dataset:edit")
	require.NoError(t, err)
	assert.Equal(t, p.PermID, got.PermID)
}

func TestRound6Perm_ListPerms(t *testing.T) {
	repo := round6Setup(t)

	// Empty list
	perms, total, err := repo.ListPerms("", 1, 10)
	require.NoError(t, err)
	assert.Zero(t, total)
	assert.Empty(t, perms)

	// Create 3 perms of different types
	round6CreatePerm(t, repo, "DV", "dashboard:view", permission.PermTypeData)
	round6CreatePerm(t, repo, "MV", "menu:view", permission.PermTypeMenu)
	round6CreatePerm(t, repo, "DE", "dataset:edit", permission.PermTypeData)

	// Filter by type
	perms, total, err = repo.ListPerms(permission.PermTypeData, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	require.Len(t, perms, 2)

	// All types with empty filter
	perms, total, err = repo.ListPerms("", 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
	require.Len(t, perms, 3)

	// Pagination: page 1 size 2
	perms, total, err = repo.ListPerms("", 1, 2)
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
	require.Len(t, perms, 2)

	// Pagination: page 2 size 2
	perms, total, err = repo.ListPerms("", 2, 2)
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
	require.Len(t, perms, 1)
}

func TestRound6Perm_CreatePerm(t *testing.T) {
	repo := round6Setup(t)
	desc := "perm description"
	p := &permission.SysPerm{
		PermName: "Export Screen",
		PermKey:  "screen:export",
		PermType: permission.PermTypeData,
		PermDesc: &desc,
		Status:   permission.StatusEnabled,
		DelFlag:  permission.DelFlagNormal,
	}
	require.NoError(t, repo.CreatePerm(p))
	assert.Positive(t, p.PermID)

	// Verify persisted
	got, err := repo.GetPermByID(p.PermID)
	require.NoError(t, err)
	assert.Equal(t, "screen:export", got.PermKey)
	require.NotNil(t, got.PermDesc)
	assert.Equal(t, desc, *got.PermDesc)
}

func TestRound6Perm_UpdatePerm(t *testing.T) {
	repo := round6Setup(t)
	p := round6CreatePerm(t, repo, "Original", "dashboard:view", permission.PermTypeData)

	p.PermName = "Updated Dashboard View"
	newDesc := "updated desc"
	p.PermDesc = &newDesc
	require.NoError(t, repo.UpdatePerm(p))

	got, err := repo.GetPermByID(p.PermID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Dashboard View", got.PermName)
	require.NotNil(t, got.PermDesc)
	assert.Equal(t, newDesc, *got.PermDesc)
}

func TestRound6Perm_DeletePerm(t *testing.T) {
	repo := round6Setup(t)
	p := round6CreatePerm(t, repo, "Temp Perm", "temp:delete", permission.PermTypeButton)

	require.NoError(t, repo.DeletePerm(p.PermID))

	// Soft delete: GetPermByID should fail (filters del_flag=0)
	_, err := repo.GetPermByID(p.PermID)
	require.Error(t, err)

	// ListPerms should not include it
	perms, total, err := repo.ListPerms("", 1, 10)
	require.NoError(t, err)
	assert.Zero(t, total)
	assert.Empty(t, perms)
}

// ==================== User/Role Permission Assignments ====================

func TestRound6Perm_GetUserPerms(t *testing.T) {
	repo := round6Setup(t)

	// Empty for nonexistent user
	permIDs, err := repo.GetUserPerms(9999)
	require.NoError(t, err)
	assert.Empty(t, permIDs)

	// Grant two perms
	require.NoError(t, repo.GrantPermToUser(1001, 2001, "admin"))
	require.NoError(t, repo.GrantPermToUser(1001, 2002, "admin"))

	permIDs, err = repo.GetUserPerms(1001)
	require.NoError(t, err)
	assert.ElementsMatch(t, []int64{2001, 2002}, permIDs)
}

func TestRound6Perm_GetRolePerms(t *testing.T) {
	repo := round6Setup(t)

	// Empty for nonexistent role
	permIDs, err := repo.GetRolePerms(9999)
	require.NoError(t, err)
	assert.Empty(t, permIDs)

	require.NoError(t, repo.GrantPermToRole(3001, 4001))
	require.NoError(t, repo.GrantPermToRole(3001, 4002))

	permIDs, err = repo.GetRolePerms(3001)
	require.NoError(t, err)
	assert.ElementsMatch(t, []int64{4001, 4002}, permIDs)
}

func TestRound6Perm_GetUserRoleIDs(t *testing.T) {
	repo := round6Setup(t)

	// Empty for nonexistent user
	roleIDs, err := repo.GetUserRoleIDs(9999)
	require.NoError(t, err)
	assert.Empty(t, roleIDs)

	require.NoError(t, repo.DB().Create(&user.SysUserRole{UserID: 5001, RoleID: 6001, OrgID: 1}).Error)
	require.NoError(t, repo.DB().Create(&user.SysUserRole{UserID: 5001, RoleID: 6002, OrgID: 1}).Error)

	roleIDs, err = repo.GetUserRoleIDs(5001)
	require.NoError(t, err)
	assert.ElementsMatch(t, []int64{6001, 6002}, roleIDs)
}

func TestRound6Perm_CheckUserPermission(t *testing.T) {
	repo := round6Setup(t)

	// Not granted
	has, err := repo.CheckUserPermission(7001, 8001)
	require.NoError(t, err)
	assert.False(t, has)

	require.NoError(t, repo.GrantPermToUser(7001, 8001, "tester"))

	has, err = repo.CheckUserPermission(7001, 8001)
	require.NoError(t, err)
	assert.True(t, has)

	has, err = repo.CheckUserPermission(7001, 9999)
	require.NoError(t, err)
	assert.False(t, has)
}

func TestRound6Perm_CheckRolePermission(t *testing.T) {
	repo := round6Setup(t)

	// Not granted
	has, err := repo.CheckRolePermission(9001, 1001)
	require.NoError(t, err)
	assert.False(t, has)

	require.NoError(t, repo.GrantPermToRole(9001, 1001))

	has, err = repo.CheckRolePermission(9001, 1001)
	require.NoError(t, err)
	assert.True(t, has)

	has, err = repo.CheckRolePermission(9001, 9999)
	require.NoError(t, err)
	assert.False(t, has)
}

func TestRound6Perm_GrantPermToUser(t *testing.T) {
	repo := round6Setup(t)

	require.NoError(t, repo.GrantPermToUser(1101, 1201, "creator"))

	has, err := repo.CheckUserPermission(1101, 1201)
	require.NoError(t, err)
	assert.True(t, has)
}

func TestRound6Perm_GrantPermToUserInOrg(t *testing.T) {
	repo := round6Setup(t)

	// orgID=0 should not set org pointer
	require.NoError(t, repo.GrantPermToUserInOrg(1301, 1401, "admin", 0))
	has, err := repo.CheckUserPermission(1301, 1401)
	require.NoError(t, err)
	assert.True(t, has)

	// orgID>0 sets the org scope
	require.NoError(t, repo.GrantPermToUserInOrg(1301, 1402, "admin", 5))
	has, err = repo.CheckUserPermission(1301, 1402)
	require.NoError(t, err)
	assert.True(t, has)
}

func TestRound6Perm_RevokePermFromUser(t *testing.T) {
	repo := round6Setup(t)

	require.NoError(t, repo.GrantPermToUser(1501, 1601, "admin"))
	has, err := repo.CheckUserPermission(1501, 1601)
	require.NoError(t, err)
	assert.True(t, has)

	require.NoError(t, repo.RevokePermFromUser(1501, 1601))
	has, err = repo.CheckUserPermission(1501, 1601)
	require.NoError(t, err)
	assert.False(t, has)
}

func TestRound6Perm_RevokePermFromUserInOrg(t *testing.T) {
	repo := round6Setup(t)

	require.NoError(t, repo.GrantPermToUserInOrg(1701, 1801, "admin", 7))
	has, err := repo.CheckUserPermission(1701, 1801)
	require.NoError(t, err)
	assert.True(t, has)

	require.NoError(t, repo.RevokePermFromUserInOrg(1701, 1801, 7))
	has, err = repo.CheckUserPermission(1701, 1801)
	require.NoError(t, err)
	assert.False(t, has)
}

func TestRound6Perm_GrantPermToRole(t *testing.T) {
	repo := round6Setup(t)

	require.NoError(t, repo.GrantPermToRole(2001, 2101))

	has, err := repo.CheckRolePermission(2001, 2101)
	require.NoError(t, err)
	assert.True(t, has)
}

func TestRound6Perm_RevokePermFromRole(t *testing.T) {
	repo := round6Setup(t)

	require.NoError(t, repo.GrantPermToRole(2201, 2301))
	has, err := repo.CheckRolePermission(2201, 2301)
	require.NoError(t, err)
	assert.True(t, has)

	require.NoError(t, repo.RevokePermFromRole(2201, 2301))
	has, err = repo.CheckRolePermission(2201, 2301)
	require.NoError(t, err)
	assert.False(t, has)

	// Revoking non-existent is not an error
	require.NoError(t, repo.RevokePermFromRole(2201, 9999))
}

// ==================== GetUserResourcesByOrg / GetResourceUsersByOrg ====================

func TestRound6Perm_GetUserResourcesByOrg(t *testing.T) {
	repo := round6Setup(t)

	p := round6CreatePerm(t, repo, "Dataset View", "dataset:view", permission.PermTypeData)

	// Grant directly in org 10
	require.NoError(t, repo.GrantPermToUserInOrg(2401, p.PermID, "admin", 10))

	// Grant via role in org 10
	round6CreateRole(t, repo, 2501, "Analyst")
	require.NoError(t, repo.GrantPermToRole(2501, p.PermID))
	require.NoError(t, repo.DB().Create(&user.SysUserRole{UserID: 2401, RoleID: 2501, OrgID: 10}).Error)

	results, err := repo.GetUserResourcesByOrg(2401, permission.ResourceTypeDataset, 10)
	require.NoError(t, err)
	require.Len(t, results, 2)
	assert.Equal(t, "direct", results[0].SourceType)
	assert.Equal(t, "role", results[1].SourceType)
	assert.Equal(t, "dataset:view", results[0].PermKey)
	assert.Equal(t, "dataset:view", results[1].PermKey)

	// Different org should return empty
	results, err = repo.GetUserResourcesByOrg(2401, permission.ResourceTypeDataset, 99)
	require.NoError(t, err)
	assert.Empty(t, results)

	// orgID=0 falls through to GetUserResources (global)
	results, err = repo.GetUserResourcesByOrg(2401, permission.ResourceTypeDataset, 0)
	require.NoError(t, err)
	require.Len(t, results, 2)
}

func TestRound6Perm_GetResourceUsersByOrg(t *testing.T) {
	repo := round6Setup(t)

	p := round6CreatePerm(t, repo, "Dataset Edit", "dataset:edit", permission.PermTypeData)
	round6CreateUser(t, repo, 2601, "direct-org-user")
	round6CreateRole(t, repo, 2701, "Org Role")

	require.NoError(t, repo.DB().Create(&permission.SysUserPerm{
		UserID: 2601, OrgID: ptrInt64Value(20), PermID: p.PermID, Status: 1, DelFlag: 0,
	}).Error)
	require.NoError(t, repo.GrantPermToRole(2701, p.PermID))
	require.NoError(t, repo.DB().Create(&user.SysUserRole{UserID: 2601, RoleID: 2701, OrgID: 20}).Error)
	require.NoError(t, repo.ReplaceResourcePermissions(301, permission.ResourceTypeDataset, []int64{p.PermID}))

	results, err := repo.GetResourceUsersByOrg(301, permission.ResourceTypeDataset, 20)
	require.NoError(t, err)
	// direct user + role user
	require.Len(t, results, 2)

	// Different org: only role user matches through user_role but org mismatch for direct perm
	results, err = repo.GetResourceUsersByOrg(301, permission.ResourceTypeDataset, 99)
	require.NoError(t, err)
	assert.Empty(t, results)

	// orgID=0 falls through to GetResourceUsers (global)
	results, err = repo.GetResourceUsersByOrg(301, permission.ResourceTypeDataset, 0)
	require.NoError(t, err)
	require.Len(t, results, 2)
}

// ==================== GetUserResources / GetResourceUsers ====================

func TestRound6Perm_GetUserResources(t *testing.T) {
	repo := round6Setup(t)

	dsView := round6CreatePerm(t, repo, "DS View", "dataset:view", permission.PermTypeData)
	dsEdit := round6CreatePerm(t, repo, "DS Edit", "dataset:edit", permission.PermTypeData)
	dbView := round6CreatePerm(t, repo, "DB View", "dashboard:view", permission.PermTypeData)

	round6CreateUser(t, repo, 2801, "test-user")
	round6CreateRole(t, repo, 2802, "Dataset Role")

	// Direct: dataset:view + dashboard:view
	require.NoError(t, repo.DB().Create(&permission.SysUserPerm{UserID: 2801, PermID: dsView.PermID, Status: 1, DelFlag: 0}).Error)
	require.NoError(t, repo.DB().Create(&permission.SysUserPerm{UserID: 2801, PermID: dbView.PermID, Status: 1, DelFlag: 0}).Error)
	// Via role: dataset:edit
	require.NoError(t, repo.GrantPermToRole(2802, dsEdit.PermID))
	require.NoError(t, repo.DB().Create(&user.SysUserRole{UserID: 2801, RoleID: 2802, OrgID: 1}).Error)

	results, err := repo.GetUserResources(2801, permission.ResourceTypeDataset)
	require.NoError(t, err)
	require.Len(t, results, 2)
	assert.Equal(t, "dataset:view", results[0].PermKey)
	assert.Equal(t, "direct", results[0].SourceType)
	assert.Equal(t, "dataset:edit", results[1].PermKey)
	assert.Equal(t, "role", results[1].SourceType)
	assert.Equal(t, int64(2802), results[1].SourceID)
	assert.Equal(t, "Dataset Role", results[1].SourceName)
}

func TestRound6Perm_GetResourceUsers(t *testing.T) {
	repo := round6Setup(t)

	dsView := round6CreatePerm(t, repo, "DS View2", "dataset:view2", permission.PermTypeData)
	dsEdit := round6CreatePerm(t, repo, "DS Edit2", "dataset:edit2", permission.PermTypeData)

	nick := "DirectNick"
	round6CreateUser(t, repo, 2901, "direct-user")
	require.NoError(t, repo.DB().Model(&user.SysUser{}).Where("user_id = ?", 2901).Update("nick_name", nick).Error)
	round6CreateUser(t, repo, 2902, "role-user")
	round6CreateRole(t, repo, 2903, "Resource Role")

	require.NoError(t, repo.DB().Create(&permission.SysUserPerm{UserID: 2901, PermID: dsView.PermID, Status: 1, DelFlag: 0}).Error)
	require.NoError(t, repo.GrantPermToRole(2903, dsEdit.PermID))
	require.NoError(t, repo.DB().Create(&user.SysUserRole{UserID: 2902, RoleID: 2903, OrgID: 1}).Error)
	require.NoError(t, repo.ReplaceResourcePermissions(401, permission.ResourceTypeDataset, []int64{dsView.PermID, dsEdit.PermID}))

	results, err := repo.GetResourceUsers(401, permission.ResourceTypeDataset)
	require.NoError(t, err)
	require.Len(t, results, 2)
	assert.Equal(t, int64(2901), results[0].UserID)
	assert.Equal(t, "direct", results[0].SourceType)
	assert.Equal(t, nick, results[0].NickName)
	assert.Equal(t, int64(2902), results[1].UserID)
	assert.Equal(t, "role", results[1].SourceType)
	assert.Equal(t, int64(2903), results[1].SourceID)
	assert.Equal(t, "Resource Role", results[1].SourceName)
}

// ==================== CheckPermissionConsistency / ByOrg ====================

func TestRound6Perm_CheckPermissionConsistency(t *testing.T) {
	repo := round6Setup(t)

	// Empty DB is consistent
	result, err := repo.CheckPermissionConsistency()
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Consistent)
	assert.Zero(t, result.UserCount)
	assert.Zero(t, result.ResourceCount)
	assert.Empty(t, result.Inconsistencies)
}

func TestRound6Perm_CheckPermissionConsistency_DetectsInconsistency(t *testing.T) {
	repo := round6Setup(t)

	// Create a perm with dataset key, grant to user, but resource has no such perm
	dsView := round6CreatePerm(t, repo, "DS View3", "dataset:view3", permission.PermTypeData)
	round6CreateUser(t, repo, 3101, "inconsistent-user")
	require.NoError(t, repo.DB().Create(&permission.SysUserPerm{
		UserID: 3101, PermID: dsView.PermID, Status: 1, DelFlag: 0,
	}).Error)
	// No resource governed => user has perm in user view but resource view is empty => inconsistency

	result, err := repo.CheckPermissionConsistency()
	require.NoError(t, err)
	assert.False(t, result.Consistent)
	assert.NotEmpty(t, result.Inconsistencies)
}

func TestRound6Perm_CheckPermissionConsistencyByOrg(t *testing.T) {
	repo := round6Setup(t)

	// Empty, any org is consistent
	result, err := repo.CheckPermissionConsistencyByOrg(1)
	require.NoError(t, err)
	assert.True(t, result.Consistent)

	// orgID=0 falls through to CheckPermissionConsistency
	result, err = repo.CheckPermissionConsistencyByOrg(0)
	require.NoError(t, err)
	assert.True(t, result.Consistent)
}

func TestRound6Perm_CheckPermissionConsistencyByOrg_Scoped(t *testing.T) {
	repo := round6Setup(t)

	dsView := round6CreatePerm(t, repo, "DS Org View", "dataset:orgview", permission.PermTypeData)
	round6CreateUser(t, repo, 3201, "org-user-1")
	round6CreateUser(t, repo, 3202, "org-user-2")
	require.NoError(t, repo.DB().Create(&permission.SysUserPerm{
		UserID: 3201, OrgID: ptrInt64Value(30), PermID: dsView.PermID, Status: 1, DelFlag: 0,
	}).Error)
	require.NoError(t, repo.DB().Create(&permission.SysUserPerm{
		UserID: 3202, OrgID: ptrInt64Value(40), PermID: dsView.PermID, Status: 1, DelFlag: 0,
	}).Error)
	require.NoError(t, repo.DB().Create(&user.SysUserRole{UserID: 3201, RoleID: 3301, OrgID: 30}).Error)
	require.NoError(t, repo.DB().Create(&user.SysUserRole{UserID: 3202, RoleID: 3302, OrgID: 40}).Error)
	require.NoError(t, repo.ReplaceResourcePermissions(501, permission.ResourceTypeDataset, []int64{dsView.PermID}))

	// Org 30 should be consistent (user 3201 has perm, resource 501 has perm)
	result, err := repo.CheckPermissionConsistencyByOrg(30)
	require.NoError(t, err)
	assert.True(t, result.Consistent)
	assert.Equal(t, 1, result.UserCount)
	assert.Equal(t, 1, result.ResourceCount)
}

// ==================== Pure helpers ====================

func TestRound6Perm_ParseInconsistency(t *testing.T) {
	// Valid key
	inc := parseInconsistency("42:dashboard:view", "granted", "missing", resourceRow{}, "user %d has %s in user view but resource view is missing")
	require.NotNil(t, inc)
	assert.Equal(t, int64(42), inc.UserID)
	assert.Equal(t, "dashboard", inc.ResourceType)
	assert.Equal(t, "granted", inc.UserView)
	assert.Equal(t, "missing", inc.ResourceView)

	// Invalid key (no colon)
	inc = parseInconsistency("badkey", "a", "b", resourceRow{}, "desc")
	assert.Nil(t, inc)

	// Non-numeric user ID
	inc = parseInconsistency("abc:key", "a", "b", resourceRow{}, "desc")
	assert.Nil(t, inc)
}

func TestRound6Perm_ExtractResourceID(t *testing.T) {
	// Known scoped ID for dataset resource 42
	scoped := scopedResourceID(permission.ResourceTypeDataset, 42)
	assert.Equal(t, int64(42), extractResourceID(resourceRow{ResourceID: scoped, ResourceType: permission.ResourceTypeDataset}))

	// Zero resource ID
	assert.Zero(t, extractResourceID(resourceRow{ResourceID: 0, ResourceType: permission.ResourceTypeDataset}))

	// Empty resource type
	assert.Zero(t, extractResourceID(resourceRow{ResourceID: 100, ResourceType: ""}))

	// Wrong type: extracts non-zero because base offset differs
	dashboardScoped := scopedResourceID(permission.ResourceTypeDashboard, 42)
	gotID := extractResourceID(resourceRow{ResourceID: dashboardScoped, ResourceType: permission.ResourceTypeDataset})
	assert.NotZero(t, gotID)
	assert.NotEqual(t, int64(42), gotID)
}

func TestRound6Perm_OrgScopedSQL(t *testing.T) {
	assert.Equal(t, "", orgScopedSQL("AND org_id = ?", 0))
	assert.Equal(t, "", orgScopedSQL("AND org_id = ?", -1))
	assert.Equal(t, "AND org_id = ?", orgScopedSQL("AND org_id = ?", 5))
}

func TestRound6Perm_ResourcePermKeyPrefix(t *testing.T) {
	assert.Equal(t, "dashboard:", resourcePermKeyPrefix(permission.ResourceTypeDashboard))
	assert.Equal(t, "screen:", resourcePermKeyPrefix(permission.ResourceTypeScreen))
	assert.Equal(t, "dataset:", resourcePermKeyPrefix(permission.ResourceTypeDataset))
	assert.Equal(t, "datasource:", resourcePermKeyPrefix(permission.ResourceTypeDatasource))
	assert.Equal(t, "custom:", resourcePermKeyPrefix("custom"))
}

func TestRound6Perm_NormalizeParentID(t *testing.T) {
	// nil stays nil
	assert.Nil(t, normalizeParentID(permission.ResourceTypeDashboard, nil))

	// zero becomes zero pointer
	zero := int64(0)
	got := normalizeParentID(permission.ResourceTypeDashboard, &zero)
	require.NotNil(t, got)
	assert.Equal(t, int64(0), *got)

	// positive becomes scoped
	parentID := int64(15)
	got = normalizeParentID(permission.ResourceTypeDataset, &parentID)
	require.NotNil(t, got)
	assert.Equal(t, scopedResourceID(permission.ResourceTypeDataset, 15), *got)
}

func TestRound6Perm_DedupeInt64(t *testing.T) {
	assert.Equal(t, []int64{5, 3, 2}, dedupeInt64([]int64{0, -1, 5, 5, 3, 5, 3, 2}))
	assert.Empty(t, dedupeInt64(nil))
	assert.Empty(t, dedupeInt64([]int64{0, -1, 0}))
}

func TestRound6Perm_ScopedResourceID(t *testing.T) {
	const stride int64 = 1_000_000_000_000
	assert.Equal(t, stride+1, scopedResourceID(permission.ResourceTypeDatasource, 1))
	assert.Equal(t, 2*stride+10, scopedResourceID(permission.ResourceTypeDataset, 10))
	assert.Equal(t, 3*stride+99, scopedResourceID(permission.ResourceTypeDashboard, 99))
	assert.Equal(t, 4*stride+5, scopedResourceID(permission.ResourceTypeScreen, 5))
	assert.Equal(t, 9*stride+7, scopedResourceID("custom", 7))
}
