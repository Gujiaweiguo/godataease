//go:build integration
// +build integration

package repository

import (
	"testing"
	"time"

	"dataease/backend/internal/domain/permission"
	"dataease/backend/internal/domain/user"
)

func TestCheckPermissionConsistency_ConsistentState_Integration(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewResourcePermissionRepository(testDB)
	cleanupPermissionConsistencyTables()
	t.Cleanup(func() {
		cleanupPermissionConsistencyTables()
	})

	userRecord := seedConsistencyUser(t, "consistency-user")
	permRecord := seedConsistencyPerm(t, "dashboard:view", "Dashboard View")
	seedConsistencyUserPerm(t, userRecord.UserID, permRecord.PermID)
	resourceRecord := seedConsistencyResource(t, permission.ResourceTypeDashboard, 101, "Integration Dashboard")
	seedConsistencyResourcePerm(t, resourceRecord.ResourceID, permRecord.PermID)

	result, err := repo.CheckPermissionConsistency()
	if err != nil {
		t.Fatalf("CheckPermissionConsistency failed: %v", err)
	}

	if !result.Consistent {
		t.Fatalf("expected consistent result, got inconsistencies: %+v", result.Inconsistencies)
	}
	if result.UserCount <= 0 {
		t.Fatalf("expected UserCount > 0, got %d", result.UserCount)
	}
	if result.ResourceCount <= 0 {
		t.Fatalf("expected ResourceCount > 0, got %d", result.ResourceCount)
	}
	if len(result.Inconsistencies) != 0 {
		t.Fatalf("expected no inconsistencies, got %d", len(result.Inconsistencies))
	}
}

func TestCheckPermissionConsistency_DivergentState_Integration(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewResourcePermissionRepository(testDB)
	cleanupPermissionConsistencyTables()
	t.Cleanup(func() {
		cleanupPermissionConsistencyTables()
	})

	userRecord := seedConsistencyUser(t, "divergent-user")
	permRecord := seedConsistencyPerm(t, "dashboard:view", "Dashboard View")
	seedConsistencyUserPerm(t, userRecord.UserID, permRecord.PermID)
	_ = seedConsistencyResource(t, permission.ResourceTypeDashboard, 202, "Unmapped Dashboard")

	result, err := repo.CheckPermissionConsistency()
	if err != nil {
		t.Fatalf("CheckPermissionConsistency failed: %v", err)
	}

	if result.Consistent {
		t.Fatal("expected inconsistent result")
	}
	if len(result.Inconsistencies) == 0 {
		t.Fatal("expected at least one inconsistency")
	}

	found := false
	for _, item := range result.Inconsistencies {
		if item.UserID == userRecord.UserID && item.UserView == "granted" && item.ResourceView == "missing" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected granted/missing inconsistency, got %+v", result.Inconsistencies)
	}
}

func TestCheckPermissionConsistency_EmptySystem_Integration(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewResourcePermissionRepository(testDB)
	cleanupPermissionConsistencyTables()
	t.Cleanup(func() {
		cleanupPermissionConsistencyTables()
	})

	result, err := repo.CheckPermissionConsistency()
	if err != nil {
		t.Fatalf("CheckPermissionConsistency failed: %v", err)
	}

	if !result.Consistent {
		t.Fatalf("expected empty system to be consistent, got %+v", result.Inconsistencies)
	}
	if result.UserCount != 0 {
		t.Fatalf("expected UserCount 0, got %d", result.UserCount)
	}
	if result.ResourceCount != 0 {
		t.Fatalf("expected ResourceCount 0, got %d", result.ResourceCount)
	}
	if len(result.Inconsistencies) != 0 {
		t.Fatalf("expected no inconsistencies, got %d", len(result.Inconsistencies))
	}
}

func cleanupPermissionConsistencyTables() {
	cleanupTables(
		"sys_resource_perm",
		"sys_resource",
		"sys_user_perm",
		"sys_user_role",
		"sys_role_perm",
		"sys_perm",
		"sys_user",
	)
}

func seedConsistencyUser(t *testing.T, username string) *user.SysUser {
	t.Helper()

	userRecord := &user.SysUser{
		Username:   username,
		Password:   "integration-test-password",
		NickName:   username,
		From:       user.FromLocal,
		Status:     user.StatusEnabled,
		DelFlag:    user.DelFlagNormal,
		CreateTime: time.Now(),
	}
	if err := testDB.Create(userRecord).Error; err != nil {
		t.Fatalf("create sys_user failed: %v", err)
	}
	return userRecord
}

func seedConsistencyPerm(t *testing.T, permKey, permName string) *permission.SysPerm {
	t.Helper()

	permRecord := &permission.SysPerm{
		PermName:   permName,
		PermKey:    permKey,
		PermType:   permission.PermTypeData,
		Status:     permission.StatusEnabled,
		CreateTime: time.Now(),
		DelFlag:    permission.DelFlagNormal,
	}
	if err := testDB.Create(permRecord).Error; err != nil {
		t.Fatalf("create sys_perm failed: %v", err)
	}
	return permRecord
}

func seedConsistencyUserPerm(t *testing.T, userID, permID int64) {
	t.Helper()

	userPerm := &user.SysUserPerm{
		UserID:  userID,
		OrgID:   1,
		PermID:  permID,
		Status:  permission.StatusEnabled,
		DelFlag: permission.DelFlagNormal,
	}
	if err := testDB.Create(userPerm).Error; err != nil {
		t.Fatalf("create sys_user_perm failed: %v", err)
	}
}

func seedConsistencyResource(t *testing.T, resourceType string, logicalResourceID int64, resourceName string) *permission.SysResource {
	t.Helper()

	now := time.Now()
	resourceRecord := &permission.SysResource{
		ResourceID:   scopedResourceID(resourceType, logicalResourceID),
		ResourceName: resourceName,
		ResourceType: resourceType,
		CreateTime:   &now,
	}
	if err := testDB.Create(resourceRecord).Error; err != nil {
		t.Fatalf("create sys_resource failed: %v", err)
	}
	return resourceRecord
}

func seedConsistencyResourcePerm(t *testing.T, resourceID, permID int64) {
	t.Helper()

	resourcePerm := &permission.SysResourcePerm{
		ResourceID: resourceID,
		PermID:     permID,
	}
	if err := testDB.Create(resourcePerm).Error; err != nil {
		t.Fatalf("create sys_resource_perm failed: %v", err)
	}
}
