//go:build integration
// +build integration

package repository

import (
	"testing"

	"dataease/backend/internal/domain/menu"
)

func TestRoleMenuRepository_SaveAndGetMenuIDs(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	roleMenuRepo := NewRoleMenuRepository(testDB)
	cleanupTables("core_menu", "sys_role_menu")

	m1 := &menu.CoreMenu{Pid: 0, Type: 0, Name: "Dashboard", Path: "/dashboard", MenuSort: 1}
	m2 := &menu.CoreMenu{Pid: 0, Type: 0, Name: "Datasource", Path: "/datasource", MenuSort: 2}
	m3 := &menu.CoreMenu{Pid: 0, Type: 0, Name: "Dataset", Path: "/dataset", MenuSort: 3}

	if err := testDB.Create(m1).Error; err != nil {
		t.Fatalf("Failed to create menu 1: %v", err)
	}
	if err := testDB.Create(m2).Error; err != nil {
		t.Fatalf("Failed to create menu 2: %v", err)
	}
	if err := testDB.Create(m3).Error; err != nil {
		t.Fatalf("Failed to create menu 3: %v", err)
	}

	roleID := int64(100)
	menuIDs := []int64{m1.ID, m2.ID}

	err := roleMenuRepo.SaveRoleMenus(roleID, menuIDs)
	if err != nil {
		t.Fatalf("SaveRoleMenus failed: %v", err)
	}

	retrievedIDs, err := roleMenuRepo.GetMenuIDsByRoleID(roleID)
	if err != nil {
		t.Fatalf("GetMenuIDsByRoleID failed: %v", err)
	}

	if len(retrievedIDs) != 2 {
		t.Errorf("Expected 2 menu IDs, got %d", len(retrievedIDs))
	}

	authorized, err := roleMenuRepo.IsMenuAuthorizedForRole(roleID, m1.ID)
	if err != nil {
		t.Fatalf("IsMenuAuthorizedForRole failed: %v", err)
	}
	if !authorized {
		t.Error("Expected menu 1 to be authorized")
	}

	authorized, err = roleMenuRepo.IsMenuAuthorizedForRole(roleID, m3.ID)
	if err != nil {
		t.Fatalf("IsMenuAuthorizedForRole failed: %v", err)
	}
	if authorized {
		t.Error("Expected menu 3 to NOT be authorized")
	}
}

func TestRoleMenuRepository_SaveRoleMenus_Idempotent(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	roleMenuRepo := NewRoleMenuRepository(testDB)
	cleanupTables("core_menu", "sys_role_menu")

	m1 := &menu.CoreMenu{Pid: 0, Type: 0, Name: "Dashboard", Path: "/dashboard", MenuSort: 1}
	m2 := &menu.CoreMenu{Pid: 0, Type: 0, Name: "Datasource", Path: "/datasource", MenuSort: 2}

	testDB.Create(m1)
	testDB.Create(m2)

	roleID := int64(200)

	err := roleMenuRepo.SaveRoleMenus(roleID, []int64{m1.ID, m2.ID})
	if err != nil {
		t.Fatalf("First SaveRoleMenus failed: %v", err)
	}

	err = roleMenuRepo.SaveRoleMenus(roleID, []int64{m1.ID})
	if err != nil {
		t.Fatalf("Second SaveRoleMenus failed: %v", err)
	}

	retrievedIDs, err := roleMenuRepo.GetMenuIDsByRoleID(roleID)
	if err != nil {
		t.Fatalf("GetMenuIDsByRoleID failed: %v", err)
	}

	if len(retrievedIDs) != 1 {
		t.Errorf("Expected 1 menu ID after update, got %d", len(retrievedIDs))
	}
}

func TestRoleMenuRepository_GetMenuIDsByRoleIDs(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	roleMenuRepo := NewRoleMenuRepository(testDB)
	cleanupTables("core_menu", "sys_role_menu")

	m1 := &menu.CoreMenu{Pid: 0, Type: 0, Name: "Dashboard", Path: "/dashboard", MenuSort: 1}
	m2 := &menu.CoreMenu{Pid: 0, Type: 0, Name: "Datasource", Path: "/datasource", MenuSort: 2}
	m3 := &menu.CoreMenu{Pid: 0, Type: 0, Name: "Dataset", Path: "/dataset", MenuSort: 3}

	testDB.Create(m1)
	testDB.Create(m2)
	testDB.Create(m3)

	role1ID := int64(301)
	role2ID := int64(302)

	roleMenuRepo.SaveRoleMenus(role1ID, []int64{m1.ID, m2.ID})
	roleMenuRepo.SaveRoleMenus(role2ID, []int64{m2.ID, m3.ID})

	combinedIDs, err := roleMenuRepo.GetMenuIDsByRoleIDs([]int64{role1ID, role2ID})
	if err != nil {
		t.Fatalf("GetMenuIDsByRoleIDs failed: %v", err)
	}

	if len(combinedIDs) != 3 {
		t.Errorf("Expected 3 unique menu IDs, got %d", len(combinedIDs))
	}
}

func TestRoleMenuRepository_DeleteByRoleID(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	roleMenuRepo := NewRoleMenuRepository(testDB)
	cleanupTables("core_menu", "sys_role_menu")

	m1 := &menu.CoreMenu{Pid: 0, Type: 0, Name: "Dashboard", Path: "/dashboard", MenuSort: 1}
	testDB.Create(m1)

	roleID := int64(400)
	roleMenuRepo.SaveRoleMenus(roleID, []int64{m1.ID})

	err := roleMenuRepo.DeleteByRoleID(roleID)
	if err != nil {
		t.Fatalf("DeleteByRoleID failed: %v", err)
	}

	retrievedIDs, err := roleMenuRepo.GetMenuIDsByRoleID(roleID)
	if err != nil {
		t.Fatalf("GetMenuIDsByRoleID failed: %v", err)
	}

	if len(retrievedIDs) != 0 {
		t.Errorf("Expected 0 menu IDs after delete, got %d", len(retrievedIDs))
	}
}

func TestRoleMenuRepository_DeleteByMenuID(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	roleMenuRepo := NewRoleMenuRepository(testDB)
	cleanupTables("core_menu", "sys_role_menu")

	m1 := &menu.CoreMenu{Pid: 0, Type: 0, Name: "Dashboard", Path: "/dashboard", MenuSort: 1}
	m2 := &menu.CoreMenu{Pid: 0, Type: 0, Name: "Datasource", Path: "/datasource", MenuSort: 2}
	testDB.Create(m1)
	testDB.Create(m2)

	roleID := int64(500)
	roleMenuRepo.SaveRoleMenus(roleID, []int64{m1.ID, m2.ID})

	err := roleMenuRepo.DeleteByMenuID(m1.ID)
	if err != nil {
		t.Fatalf("DeleteByMenuID failed: %v", err)
	}

	retrievedIDs, err := roleMenuRepo.GetMenuIDsByRoleID(roleID)
	if err != nil {
		t.Fatalf("GetMenuIDsByRoleID failed: %v", err)
	}

	if len(retrievedIDs) != 1 {
		t.Errorf("Expected 1 menu ID after menu delete, got %d", len(retrievedIDs))
	}
}

func TestRoleMenuRepository_IsMenuAuthorizedForRoles(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	roleMenuRepo := NewRoleMenuRepository(testDB)
	cleanupTables("core_menu", "sys_role_menu")

	m1 := &menu.CoreMenu{Pid: 0, Type: 0, Name: "Dashboard", Path: "/dashboard", MenuSort: 1}
	testDB.Create(m1)

	role1ID := int64(601)
	role2ID := int64(602)

	roleMenuRepo.SaveRoleMenus(role1ID, []int64{m1.ID})

	authorized, err := roleMenuRepo.IsMenuAuthorizedForRoles([]int64{role1ID, role2ID}, m1.ID)
	if err != nil {
		t.Fatalf("IsMenuAuthorizedForRoles failed: %v", err)
	}
	if !authorized {
		t.Error("Expected menu to be authorized via role1")
	}

	authorized, err = roleMenuRepo.IsMenuAuthorizedForRoles([]int64{role2ID}, m1.ID)
	if err != nil {
		t.Fatalf("IsMenuAuthorizedForRoles failed: %v", err)
	}
	if authorized {
		t.Error("Expected menu to NOT be authorized for role2 alone")
	}
}
