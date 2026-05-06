package repository

import (
	"reflect"
	"testing"

	"dataease/backend/internal/domain/permission"
	"dataease/backend/internal/domain/role"
	"dataease/backend/internal/domain/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupResourcePermissionRepositoryTest(t *testing.T) *ResourcePermissionRepository {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&permission.SysPerm{}, &permission.SysResource{}, &permission.SysResourcePerm{}, &permission.SysUserPerm{}, &permission.SysRolePerm{}, &user.SysUser{}, &user.SysUserRole{}, &role.SysRole{}))
	return NewResourcePermissionRepository(db)
}

func TestNewResourcePermissionRepository(t *testing.T) {
	repo := setupResourcePermissionRepositoryTest(t)
	require.NotNil(t, repo)
}

func TestResourcePermKeyPrefix(t *testing.T) {
	tests := []struct {
		name         string
		resourceType string
		want         string
	}{
		{name: "dashboard", resourceType: permission.ResourceTypeDashboard, want: "dashboard:"},
		{name: "screen", resourceType: permission.ResourceTypeScreen, want: "screen:"},
		{name: "dataset", resourceType: permission.ResourceTypeDataset, want: "dataset:"},
		{name: "datasource", resourceType: permission.ResourceTypeDatasource, want: "datasource:"},
		{name: "fallback", resourceType: "custom", want: "custom:"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resourcePermKeyPrefix(tt.resourceType); got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestNormalizeParentID(t *testing.T) {
	t.Run("nil parent stays nil", func(t *testing.T) {
		if got := normalizeParentID(permission.ResourceTypeDashboard, nil); got != nil {
			t.Fatalf("expected nil parent, got %v", *got)
		}
	})

	t.Run("non-positive parent becomes zero", func(t *testing.T) {
		parentID := int64(0)
		got := normalizeParentID(permission.ResourceTypeDashboard, &parentID)
		if got == nil || *got != 0 {
			t.Fatalf("expected zero parent, got %v", got)
		}
	})

	t.Run("positive parent is scoped", func(t *testing.T) {
		parentID := int64(12)
		got := normalizeParentID(permission.ResourceTypeDataset, &parentID)
		want := scopedResourceID(permission.ResourceTypeDataset, 12)
		if got == nil || *got != want {
			t.Fatalf("expected scoped parent %d, got %v", want, got)
		}
	})
}

func TestDedupeInt64(t *testing.T) {
	got := dedupeInt64([]int64{0, -1, 5, 5, 3, 5, 3, 2})
	want := []int64{5, 3, 2}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestScopedResourceID(t *testing.T) {
	const stride int64 = 1_000_000_000_000
	tests := []struct {
		name         string
		resourceType string
		resourceID   int64
		want         int64
	}{
		{name: "datasource", resourceType: permission.ResourceTypeDatasource, resourceID: 9, want: stride + 9},
		{name: "dataset", resourceType: permission.ResourceTypeDataset, resourceID: 9, want: 2*stride + 9},
		{name: "dashboard", resourceType: permission.ResourceTypeDashboard, resourceID: 9, want: 3*stride + 9},
		{name: "screen", resourceType: permission.ResourceTypeScreen, resourceID: 9, want: 4*stride + 9},
		{name: "fallback", resourceType: "custom", resourceID: 9, want: 9*stride + 9},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := scopedResourceID(tt.resourceType, tt.resourceID); got != tt.want {
				t.Fatalf("expected %d, got %d", tt.want, got)
			}
		})
	}
}

func TestRegisterResource(t *testing.T) {
	t.Run("requires resource id and type", func(t *testing.T) {
		repo := setupResourcePermissionRepositoryTest(t)
		err := repo.RegisterResource(0, "", permission.ResourceTypeDataset, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "resource id and type are required")
	})

	t.Run("creates fallback name and scoped parent", func(t *testing.T) {
		repo := setupResourcePermissionRepositoryTest(t)
		parentID := int64(5)
		err := repo.RegisterResource(11, "  ", permission.ResourceTypeDataset, &parentID)
		require.NoError(t, err)

		permIDs, exists, err := repo.GetResourcePermissionIDs(11, permission.ResourceTypeDataset)
		require.NoError(t, err)
		assert.True(t, exists)
		assert.Empty(t, permIDs)
	})

	t.Run("updates existing name and normalized parent", func(t *testing.T) {
		repo := setupResourcePermissionRepositoryTest(t)
		err := repo.RegisterResource(12, "Old Name", permission.ResourceTypeDashboard, nil)
		require.NoError(t, err)

		zeroParent := int64(0)
		err = repo.RegisterResource(12, "New Name", permission.ResourceTypeDashboard, &zeroParent)
		require.NoError(t, err)

		var resource permission.SysResource
		err = repo.db.Where("resource_id = ? AND resource_type = ?", scopedResourceID(permission.ResourceTypeDashboard, 12), permission.ResourceTypeDashboard).First(&resource).Error
		require.NoError(t, err)
		assert.Equal(t, "New Name", resource.ResourceName)
		require.NotNil(t, resource.ParentID)
		assert.Equal(t, int64(0), *resource.ParentID)
	})
}

func TestReplaceResourcePermissions(t *testing.T) {
	repo := setupResourcePermissionRepositoryTest(t)

	err := repo.ReplaceResourcePermissions(21, permission.ResourceTypeDatasource, []int64{7, 7, 0, -1, 5})
	require.NoError(t, err)

	permIDs, exists, err := repo.GetResourcePermissionIDs(21, permission.ResourceTypeDatasource)
	require.NoError(t, err)
	assert.True(t, exists)
	assert.ElementsMatch(t, []int64{7, 5}, permIDs)

	err = repo.ReplaceResourcePermissions(21, permission.ResourceTypeDatasource, nil)
	require.NoError(t, err)

	permIDs, exists, err = repo.GetResourcePermissionIDs(21, permission.ResourceTypeDatasource)
	require.NoError(t, err)
	assert.True(t, exists)
	assert.Empty(t, permIDs)
}

func TestGetResourcePermissionIDs(t *testing.T) {
	repo := setupResourcePermissionRepositoryTest(t)

	permIDs, exists, err := repo.GetResourcePermissionIDs(0, permission.ResourceTypeDataset)
	require.NoError(t, err)
	assert.False(t, exists)
	assert.Nil(t, permIDs)

	permIDs, exists, err = repo.GetResourcePermissionIDs(33, permission.ResourceTypeDataset)
	require.NoError(t, err)
	assert.False(t, exists)
	assert.Nil(t, permIDs)

	err = repo.RegisterResource(33, "dataset-33", permission.ResourceTypeDataset, nil)
	require.NoError(t, err)
	permIDs, exists, err = repo.GetResourcePermissionIDs(33, permission.ResourceTypeDataset)
	require.NoError(t, err)
	assert.True(t, exists)
	assert.Empty(t, permIDs)
}

func TestApplyGroupPermissions(t *testing.T) {
	t.Run("invalid input is no-op", func(t *testing.T) {
		repo := setupResourcePermissionRepositoryTest(t)
		err := repo.ApplyGroupPermissions(0, 1, permission.ResourceTypeDataset)
		require.NoError(t, err)

		permIDs, exists, err := repo.GetResourcePermissionIDs(1, permission.ResourceTypeDataset)
		require.NoError(t, err)
		assert.False(t, exists)
		assert.Nil(t, permIDs)
	})

	t.Run("missing parent governance is no-op", func(t *testing.T) {
		repo := setupResourcePermissionRepositoryTest(t)
		err := repo.ApplyGroupPermissions(41, 42, permission.ResourceTypeDataset)
		require.NoError(t, err)

		permIDs, exists, err := repo.GetResourcePermissionIDs(42, permission.ResourceTypeDataset)
		require.NoError(t, err)
		assert.False(t, exists)
		assert.Nil(t, permIDs)
	})

	t.Run("copies governed parent permissions to child", func(t *testing.T) {
		repo := setupResourcePermissionRepositoryTest(t)
		err := repo.ReplaceResourcePermissions(51, permission.ResourceTypeDataset, []int64{9, 8, 9})
		require.NoError(t, err)

		err = repo.ApplyGroupPermissions(51, 52, permission.ResourceTypeDataset)
		require.NoError(t, err)

		permIDs, exists, err := repo.GetResourcePermissionIDs(52, permission.ResourceTypeDataset)
		require.NoError(t, err)
		assert.True(t, exists)
		assert.ElementsMatch(t, []int64{9, 8}, permIDs)
	})
}

func TestPermissionCRUD(t *testing.T) {
	repo := setupResourcePermissionRepositoryTest(t)
	desc := "initial desc"
	perm := &permission.SysPerm{PermName: "View Dataset", PermKey: "dataset:view", PermType: permission.PermTypeData, PermDesc: &desc, Status: permission.StatusEnabled, DelFlag: permission.DelFlagNormal}

	err := repo.CreatePerm(perm)
	require.NoError(t, err)
	require.Positive(t, perm.PermID)

	byID, err := repo.GetPermByID(perm.PermID)
	require.NoError(t, err)
	assert.Equal(t, perm.PermKey, byID.PermKey)

	byKey, err := repo.GetPermByKey("dataset:view")
	require.NoError(t, err)
	assert.Equal(t, perm.PermID, byKey.PermID)

	updatedDesc := "updated desc"
	perm.PermName = "Manage Dataset"
	perm.PermDesc = &updatedDesc
	err = repo.UpdatePerm(perm)
	require.NoError(t, err)

	updated, err := repo.GetPermByID(perm.PermID)
	require.NoError(t, err)
	assert.Equal(t, "Manage Dataset", updated.PermName)
	require.NotNil(t, updated.PermDesc)
	assert.Equal(t, updatedDesc, *updated.PermDesc)

	perms, total, err := repo.ListPerms(permission.PermTypeData, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	require.Len(t, perms, 1)
	assert.Equal(t, perm.PermID, perms[0].PermID)

	err = repo.DeletePerm(perm.PermID)
	require.NoError(t, err)

	_, err = repo.GetPermByID(perm.PermID)
	require.Error(t, err)
	_, err = repo.GetPermByKey("dataset:view")
	require.Error(t, err)

	perms, total, err = repo.ListPerms(permission.PermTypeData, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Empty(t, perms)
}

func TestUserAndRolePermissionAssignments(t *testing.T) {
	repo := setupResourcePermissionRepositoryTest(t)

	err := repo.GrantPermToUser(101, 201, "tester")
	require.NoError(t, err)
	err = repo.GrantPermToUser(101, 202, "tester")
	require.NoError(t, err)
	err = repo.GrantPermToRole(301, 401)
	require.NoError(t, err)
	err = repo.GrantPermToRole(301, 402)
	require.NoError(t, err)
	err = repo.db.Create(&user.SysUserRole{UserID: 101, RoleID: 301, OrgID: 1}).Error
	require.NoError(t, err)

	userPermIDs, err := repo.GetUserPerms(101)
	require.NoError(t, err)
	assert.ElementsMatch(t, []int64{201, 202}, userPermIDs)

	hasUserPerm, err := repo.CheckUserPermission(101, 201)
	require.NoError(t, err)
	assert.True(t, hasUserPerm)
	hasUserPerm, err = repo.CheckUserPermission(101, 999)
	require.NoError(t, err)
	assert.False(t, hasUserPerm)

	rolePermIDs, err := repo.GetRolePerms(301)
	require.NoError(t, err)
	assert.ElementsMatch(t, []int64{401, 402}, rolePermIDs)

	hasRolePerm, err := repo.CheckRolePermission(301, 401)
	require.NoError(t, err)
	assert.True(t, hasRolePerm)
	hasRolePerm, err = repo.CheckRolePermission(301, 999)
	require.NoError(t, err)
	assert.False(t, hasRolePerm)

	roleIDs, err := repo.GetUserRoleIDs(101)
	require.NoError(t, err)
	assert.Equal(t, []int64{301}, roleIDs)

	err = repo.RevokePermFromUser(101, 201)
	require.NoError(t, err)
	userPermIDs, err = repo.GetUserPerms(101)
	require.NoError(t, err)
	assert.Equal(t, []int64{202}, userPermIDs)

	err = repo.RevokePermFromRole(301, 401)
	require.NoError(t, err)
	rolePermIDs, err = repo.GetRolePerms(301)
	require.NoError(t, err)
	assert.Equal(t, []int64{402}, rolePermIDs)
}

func TestCheckPermissionConsistency(t *testing.T) {
	repo := setupResourcePermissionRepositoryTest(t)

	result, err := repo.CheckPermissionConsistency()
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Consistent)
	assert.Zero(t, result.UserCount)
	assert.Zero(t, result.ResourceCount)
	assert.Empty(t, result.Inconsistencies)
}

func TestCheckPermissionConsistencyByOrg(t *testing.T) {
	repo := setupResourcePermissionRepositoryTest(t)
	datasetView := &permission.SysPerm{PermName: "Dataset View", PermKey: "dataset:view", PermType: permission.PermTypeData, Status: permission.StatusEnabled, DelFlag: permission.DelFlagNormal}
	missingPerm := &permission.SysPerm{PermName: "Dataset Edit", PermKey: "dataset:edit", PermType: permission.PermTypeData, Status: permission.StatusEnabled, DelFlag: permission.DelFlagNormal}
	require.NoError(t, repo.db.Create(datasetView).Error)
	require.NoError(t, repo.db.Create(missingPerm).Error)
	require.NoError(t, repo.db.Create(&user.SysUser{UserID: 801, Username: "org1-user", Status: user.StatusEnabled, DelFlag: user.DelFlagNormal}).Error)
	require.NoError(t, repo.db.Create(&user.SysUser{UserID: 802, Username: "org2-user", Status: user.StatusEnabled, DelFlag: user.DelFlagNormal}).Error)
	require.NoError(t, repo.db.Create(&permission.SysUserPerm{UserID: 801, OrgID: ptrInt64Value(1), PermID: datasetView.PermID, Status: 1, DelFlag: 0}).Error)
	require.NoError(t, repo.db.Create(&permission.SysUserPerm{UserID: 802, OrgID: ptrInt64Value(2), PermID: missingPerm.PermID, Status: 1, DelFlag: 0}).Error)
	require.NoError(t, repo.db.Create(&user.SysUserRole{UserID: 801, RoleID: 901, OrgID: 1}).Error)
	require.NoError(t, repo.db.Create(&user.SysUserRole{UserID: 802, RoleID: 902, OrgID: 2}).Error)
	require.NoError(t, repo.ReplaceResourcePermissions(101, permission.ResourceTypeDataset, []int64{datasetView.PermID}))

	orgScoped, err := repo.CheckPermissionConsistencyByOrg(1)
	require.NoError(t, err)
	assert.True(t, orgScoped.Consistent)
	assert.Equal(t, 1, orgScoped.UserCount)
	assert.Equal(t, 1, orgScoped.ResourceCount)
	assert.Empty(t, orgScoped.Inconsistencies)

	global, err := repo.CheckPermissionConsistency()
	require.NoError(t, err)
	assert.False(t, global.Consistent)
	assert.NotEmpty(t, global.Inconsistencies)
}

func ptrInt64Value(v int64) *int64 {
	return &v
}

func TestGetUserResources(t *testing.T) {
	t.Run("filters direct and role permissions by resource type prefix", func(t *testing.T) {
		repo := setupResourcePermissionRepositoryTest(t)
		datasetView := &permission.SysPerm{PermName: "Dataset View", PermKey: "dataset:view", PermType: permission.PermTypeData, Status: permission.StatusEnabled, DelFlag: permission.DelFlagNormal}
		datasetEdit := &permission.SysPerm{PermName: "Dataset Edit", PermKey: "dataset:edit", PermType: permission.PermTypeData, Status: permission.StatusEnabled, DelFlag: permission.DelFlagNormal}
		dashboardView := &permission.SysPerm{PermName: "Dashboard View", PermKey: "dashboard:view", PermType: permission.PermTypeData, Status: permission.StatusEnabled, DelFlag: permission.DelFlagNormal}
		require.NoError(t, repo.db.Create(datasetView).Error)
		require.NoError(t, repo.db.Create(datasetEdit).Error)
		require.NoError(t, repo.db.Create(dashboardView).Error)
		require.NoError(t, repo.db.Create(&permission.SysUserPerm{UserID: 501, PermID: datasetView.PermID, Status: 1, DelFlag: 0}).Error)
		require.NoError(t, repo.db.Create(&permission.SysUserPerm{UserID: 501, PermID: dashboardView.PermID, Status: 1, DelFlag: 0}).Error)
		require.NoError(t, repo.db.Create(&role.SysRole{RoleID: 601, RoleName: "Analyst", Status: role.StatusEnabled}).Error)
		require.NoError(t, repo.db.Create(&permission.SysRolePerm{RoleID: 601, PermID: datasetEdit.PermID}).Error)
		require.NoError(t, repo.db.Create(&user.SysUserRole{UserID: 501, RoleID: 601, OrgID: 1}).Error)

		results, err := repo.GetUserResources(501, permission.ResourceTypeDataset)
		require.NoError(t, err)
		require.Len(t, results, 2)
		assert.Equal(t, "dataset:view", results[0].PermKey)
		assert.Equal(t, "direct", results[0].SourceType)
		assert.Equal(t, permission.ResourceTypeDataset, results[0].ResourceType)
		assert.Equal(t, "dataset:edit", results[1].PermKey)
		assert.Equal(t, "role", results[1].SourceType)
		assert.Equal(t, int64(601), results[1].SourceID)
		assert.Equal(t, "Analyst", results[1].SourceName)
	})

	t.Run("custom resource type keeps fallback prefix", func(t *testing.T) {
		repo := setupResourcePermissionRepositoryTest(t)
		customPerm := &permission.SysPerm{PermName: "Custom View", PermKey: "custom:view", PermType: permission.PermTypeData, Status: permission.StatusEnabled, DelFlag: permission.DelFlagNormal}
		require.NoError(t, repo.db.Create(customPerm).Error)
		require.NoError(t, repo.db.Create(&permission.SysUserPerm{UserID: 777, PermID: customPerm.PermID, Status: 1, DelFlag: 0}).Error)

		results, err := repo.GetUserResources(777, "custom")
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, "custom:view", results[0].PermKey)
		assert.Equal(t, "direct", results[0].SourceType)
	})
}

func TestGetResourceUsers(t *testing.T) {
	t.Run("returns direct and role users when resource is governed", func(t *testing.T) {
		repo := setupResourcePermissionRepositoryTest(t)
		datasetView := &permission.SysPerm{PermName: "Dataset View", PermKey: "dataset:view", PermType: permission.PermTypeData, Status: permission.StatusEnabled, DelFlag: permission.DelFlagNormal}
		datasetEdit := &permission.SysPerm{PermName: "Dataset Edit", PermKey: "dataset:edit", PermType: permission.PermTypeData, Status: permission.StatusEnabled, DelFlag: permission.DelFlagNormal}
		require.NoError(t, repo.db.Create(datasetView).Error)
		require.NoError(t, repo.db.Create(datasetEdit).Error)
		nick := "Direct Nick"
		require.NoError(t, repo.db.Create(&user.SysUser{UserID: 701, Username: "direct-user", NickName: nick, Status: user.StatusEnabled, DelFlag: user.DelFlagNormal}).Error)
		require.NoError(t, repo.db.Create(&user.SysUser{UserID: 702, Username: "role-user", Status: user.StatusEnabled, DelFlag: user.DelFlagNormal}).Error)
		require.NoError(t, repo.db.Create(&role.SysRole{RoleID: 801, RoleName: "Viewer", Status: role.StatusEnabled}).Error)
		require.NoError(t, repo.db.Create(&permission.SysUserPerm{UserID: 701, PermID: datasetView.PermID, Status: 1, DelFlag: 0}).Error)
		require.NoError(t, repo.db.Create(&permission.SysRolePerm{RoleID: 801, PermID: datasetEdit.PermID}).Error)
		require.NoError(t, repo.db.Create(&user.SysUserRole{UserID: 702, RoleID: 801, OrgID: 1}).Error)
		require.NoError(t, repo.ReplaceResourcePermissions(91, permission.ResourceTypeDataset, []int64{datasetView.PermID, datasetEdit.PermID}))

		results, err := repo.GetResourceUsers(91, permission.ResourceTypeDataset)
		require.NoError(t, err)
		require.Len(t, results, 2)
		assert.Equal(t, int64(701), results[0].UserID)
		assert.Equal(t, "direct", results[0].SourceType)
		assert.Equal(t, "Direct Nick", results[0].NickName)
		assert.Equal(t, int64(702), results[1].UserID)
		assert.Equal(t, "role", results[1].SourceType)
		assert.Equal(t, int64(801), results[1].SourceID)
		assert.Equal(t, "Viewer", results[1].SourceName)
	})

	t.Run("governed resource with no permissions returns empty", func(t *testing.T) {
		repo := setupResourcePermissionRepositoryTest(t)
		require.NoError(t, repo.RegisterResource(92, "dataset-92", permission.ResourceTypeDataset, nil))

		results, err := repo.GetResourceUsers(92, permission.ResourceTypeDataset)
		require.NoError(t, err)
		assert.Empty(t, results)
	})

	t.Run("falls back to prefix lookup when resource is not governed", func(t *testing.T) {
		repo := setupResourcePermissionRepositoryTest(t)
		datasetView := &permission.SysPerm{PermName: "Dataset View", PermKey: "dataset:view", PermType: permission.PermTypeData, Status: permission.StatusEnabled, DelFlag: permission.DelFlagNormal}
		dashboardView := &permission.SysPerm{PermName: "Dashboard View", PermKey: "dashboard:view", PermType: permission.PermTypeData, Status: permission.StatusEnabled, DelFlag: permission.DelFlagNormal}
		require.NoError(t, repo.db.Create(datasetView).Error)
		require.NoError(t, repo.db.Create(dashboardView).Error)
		require.NoError(t, repo.db.Create(&user.SysUser{UserID: 703, Username: "prefix-user", Status: user.StatusEnabled, DelFlag: user.DelFlagNormal}).Error)
		require.NoError(t, repo.db.Create(&role.SysRole{RoleID: 802, RoleName: "Dataset Role", Status: role.StatusEnabled}).Error)
		require.NoError(t, repo.db.Create(&permission.SysUserPerm{UserID: 703, PermID: datasetView.PermID, Status: 1, DelFlag: 0}).Error)
		require.NoError(t, repo.db.Create(&permission.SysUserPerm{UserID: 703, PermID: dashboardView.PermID, Status: 1, DelFlag: 0}).Error)
		require.NoError(t, repo.db.Create(&permission.SysRolePerm{RoleID: 802, PermID: datasetView.PermID}).Error)
		require.NoError(t, repo.db.Create(&user.SysUserRole{UserID: 703, RoleID: 802, OrgID: 1}).Error)

		results, err := repo.GetResourceUsers(93, permission.ResourceTypeDataset)
		require.NoError(t, err)
		require.Len(t, results, 2)
		assert.Equal(t, "dataset:view", results[0].PermKey)
		assert.Equal(t, "direct", results[0].SourceType)
		assert.Equal(t, "dataset:view", results[1].PermKey)
		assert.Equal(t, "role", results[1].SourceType)
	})
}
