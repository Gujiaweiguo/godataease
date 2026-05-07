package repository

import (
	"testing"

	"dataease/backend/internal/domain/permission"
	"dataease/backend/internal/domain/role"
	"dataease/backend/internal/domain/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResourcePermissionRepository_Gap2DBAndOrgScopedUserPerms(t *testing.T) {
	var nilRepo *ResourcePermissionRepository
	assert.Nil(t, nilRepo.DB())

	repo := setupResourcePermissionRepositoryTest(t)
	require.Same(t, repo.db, repo.DB())

	datasetView := &permission.SysPerm{PermName: "Dataset View", PermKey: "dataset:view", PermType: permission.PermTypeData, Status: permission.StatusEnabled, DelFlag: permission.DelFlagNormal}
	dashboardView := &permission.SysPerm{PermName: "Dashboard View", PermKey: "dashboard:view", PermType: permission.PermTypeData, Status: permission.StatusEnabled, DelFlag: permission.DelFlagNormal}
	require.NoError(t, repo.db.Create(datasetView).Error)
	require.NoError(t, repo.db.Create(dashboardView).Error)

	require.NoError(t, repo.GrantPermToUserInOrg(901, datasetView.PermID, "tester", 9))
	require.NoError(t, repo.GrantPermToUserInOrg(901, dashboardView.PermID, "tester", 10))
	require.NoError(t, repo.db.Create(&role.SysRole{RoleID: 801, RoleName: "Scoped Role", Status: role.StatusEnabled}).Error)
	require.NoError(t, repo.db.Create(&permission.SysRolePerm{RoleID: 801, PermID: datasetView.PermID}).Error)
	require.NoError(t, repo.db.Create(&user.SysUserRole{UserID: 901, RoleID: 801, OrgID: 9}).Error)

	results, err := repo.GetUserResourcesByOrg(901, permission.ResourceTypeDataset, 9)
	require.NoError(t, err)
	require.Len(t, results, 2)
	assert.Equal(t, "dataset:view", results[0].PermKey)
	assert.Equal(t, "direct", results[0].SourceType)
	assert.Equal(t, "dataset:view", results[1].PermKey)
	assert.Equal(t, "role", results[1].SourceType)

	results, err = repo.GetUserResourcesByOrg(901, permission.ResourceTypeDataset, 10)
	require.NoError(t, err)
	assert.Empty(t, results)

	require.NoError(t, repo.RevokePermFromUserInOrg(901, datasetView.PermID, 9))
	hasPerm, err := repo.CheckUserPermission(901, datasetView.PermID)
	require.NoError(t, err)
	assert.False(t, hasPerm)
}

func TestResourcePermissionRepository_Gap2OrgScopedResourceUsersAndHelpers(t *testing.T) {
	repo := setupResourcePermissionRepositoryTest(t)
	datasetView := &permission.SysPerm{PermName: "Dataset View", PermKey: "dataset:view", PermType: permission.PermTypeData, Status: permission.StatusEnabled, DelFlag: permission.DelFlagNormal}
	require.NoError(t, repo.db.Create(datasetView).Error)
	require.NoError(t, repo.db.Create(&user.SysUser{UserID: 1001, Username: "direct-org-user", Status: user.StatusEnabled, DelFlag: user.DelFlagNormal}).Error)
	require.NoError(t, repo.db.Create(&user.SysUser{UserID: 1002, Username: "role-org-user", Status: user.StatusEnabled, DelFlag: user.DelFlagNormal}).Error)
	require.NoError(t, repo.db.Create(&role.SysRole{RoleID: 1003, RoleName: "Org Role", Status: role.StatusEnabled}).Error)
	require.NoError(t, repo.db.Create(&permission.SysUserPerm{UserID: 1001, OrgID: ptrInt64Value(11), PermID: datasetView.PermID, Status: 1, DelFlag: 0}).Error)
	require.NoError(t, repo.db.Create(&permission.SysRolePerm{RoleID: 1003, PermID: datasetView.PermID}).Error)
	require.NoError(t, repo.db.Create(&user.SysUserRole{UserID: 1002, RoleID: 1003, OrgID: 11}).Error)
	require.NoError(t, repo.db.Create(&user.SysUserRole{UserID: 1002, RoleID: 1003, OrgID: 12}).Error)
	require.NoError(t, repo.ReplaceResourcePermissions(501, permission.ResourceTypeDataset, []int64{datasetView.PermID}))

	results, err := repo.GetResourceUsersByOrg(501, permission.ResourceTypeDataset, 11)
	require.NoError(t, err)
	require.Len(t, results, 2)
	assert.Equal(t, int64(1001), results[0].UserID)
	assert.Equal(t, "direct", results[0].SourceType)
	assert.Equal(t, int64(1002), results[1].UserID)
	assert.Equal(t, "role", results[1].SourceType)

	results, err = repo.GetResourceUsersByOrg(501, permission.ResourceTypeDataset, 12)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, int64(1002), results[0].UserID)
	assert.Equal(t, "role", results[0].SourceType)

	assert.Equal(t, int64(501), extractResourceID(resourceRow{ResourceID: scopedResourceID(permission.ResourceTypeDataset, 501), ResourceType: permission.ResourceTypeDataset}))
	assert.Zero(t, extractResourceID(resourceRow{ResourceID: 1, ResourceType: permission.ResourceTypeDataset}))
	assert.Zero(t, extractResourceID(resourceRow{ResourceID: 0, ResourceType: permission.ResourceTypeDataset}))
	assert.Zero(t, extractResourceID(resourceRow{ResourceID: scopedResourceID(permission.ResourceTypeDataset, 5), ResourceType: ""}))
	assert.Equal(t, "", orgScopedSQL("AND sur.org_id = ?", 0))
	assert.Equal(t, "AND sur.org_id = ?", orgScopedSQL("AND sur.org_id = ?", 11))
}
