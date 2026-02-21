//go:build integration
// +build integration

package repository

import (
	"testing"
	"time"

	"dataease/backend/internal/domain/permission"
)

func TestPermRepository_CreateAndGetByID(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewPermRepository(testDB)
	cleanupTables("sys_perm")

	perm := &permission.SysPerm{
		PermName:   "Test Permission",
		PermKey:    "test:perm",
		PermType:   permission.PermTypeMenu,
		PermDesc:   strPtr("Test permission description"),
		Status:     permission.StatusEnabled,
		CreateBy:   strPtr("admin"),
		CreateTime: time.Now(),
		DelFlag:    permission.DelFlagNormal,
	}

	err := repo.Create(perm)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if perm.PermID == 0 {
		t.Error("Expected PermID to be set after creation")
	}

	found, err := repo.GetByID(perm.PermID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if found.PermName != "Test Permission" {
		t.Errorf("Expected PermName 'Test Permission', got '%s'", found.PermName)
	}
}

func TestPermRepository_GetByKey(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewPermRepository(testDB)
	cleanupTables("sys_perm")

	perm := &permission.SysPerm{
		PermName:   "Key Test Permission",
		PermKey:    "unique:key:test",
		PermType:   permission.PermTypeButton,
		Status:     permission.StatusEnabled,
		CreateTime: time.Now(),
		DelFlag:    permission.DelFlagNormal,
	}
	_ = repo.Create(perm)

	found, err := repo.GetByKey("unique:key:test")
	if err != nil {
		t.Fatalf("GetByKey failed: %v", err)
	}

	if found.PermKey != "unique:key:test" {
		t.Errorf("Expected PermKey 'unique:key:test', got '%s'", found.PermKey)
	}
}

func TestPermRepository_List(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewPermRepository(testDB)
	cleanupTables("sys_perm")

	for i := 0; i < 3; i++ {
		perm := &permission.SysPerm{
			PermName:   "List Permission",
			PermKey:    "list:perm",
			PermType:   permission.PermTypeMenu,
			Status:     permission.StatusEnabled,
			CreateTime: time.Now(),
			DelFlag:    permission.DelFlagNormal,
		}
		_ = repo.Create(perm)
	}

	perms, err := repo.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(perms) != 3 {
		t.Errorf("Expected 3 permissions, got %d", len(perms))
	}
}

func TestPermRepository_Update(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewPermRepository(testDB)
	cleanupTables("sys_perm")

	perm := &permission.SysPerm{
		PermName:   "Update Test",
		PermKey:    "update:test",
		PermType:   permission.PermTypeMenu,
		Status:     permission.StatusEnabled,
		CreateTime: time.Now(),
		DelFlag:    permission.DelFlagNormal,
	}
	_ = repo.Create(perm)

	perm.PermName = "Updated Name"
	perm.Status = permission.StatusDisabled

	err := repo.Update(perm)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	found, _ := repo.GetByID(perm.PermID)
	if found.PermName != "Updated Name" {
		t.Errorf("Expected PermName 'Updated Name', got '%s'", found.PermName)
	}
}

func TestPermRepository_Delete(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewPermRepository(testDB)
	cleanupTables("sys_perm")

	perm := &permission.SysPerm{
		PermName:   "Delete Test",
		PermKey:    "delete:test",
		PermType:   permission.PermTypeMenu,
		Status:     permission.StatusEnabled,
		CreateTime: time.Now(),
		DelFlag:    permission.DelFlagNormal,
	}
	_ = repo.Create(perm)

	err := repo.Delete(perm.PermID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = repo.GetByID(perm.PermID)
	if err == nil {
		t.Error("Expected error when getting deleted permission")
	}
}

func TestPermRepository_CheckKeyExists(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewPermRepository(testDB)
	cleanupTables("sys_perm")

	perm := &permission.SysPerm{
		PermName:   "Check Key Test",
		PermKey:    "check:key:test",
		PermType:   permission.PermTypeMenu,
		Status:     permission.StatusEnabled,
		CreateTime: time.Now(),
		DelFlag:    permission.DelFlagNormal,
	}
	_ = repo.Create(perm)

	count, err := repo.CheckKeyExists("check:key:test", 0)
	if err != nil {
		t.Fatalf("CheckKeyExists failed: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected count 1, got %d", count)
	}

	// Check excluding self
	count, err = repo.CheckKeyExists("check:key:test", perm.PermID)
	if err != nil {
		t.Fatalf("CheckKeyExists with exclude failed: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected count 0 when excluding self, got %d", count)
	}
}

func TestPermRepository_GetByType(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewPermRepository(testDB)
	cleanupTables("sys_perm")

	for i := 0; i < 2; i++ {
		perm := &permission.SysPerm{
			PermName:   "Menu Permission",
			PermKey:    "menu:perm",
			PermType:   permission.PermTypeMenu,
			Status:     permission.StatusEnabled,
			CreateTime: time.Now(),
			DelFlag:    permission.DelFlagNormal,
		}
		_ = repo.Create(perm)
	}

	for i := 0; i < 3; i++ {
		perm := &permission.SysPerm{
			PermName:   "Button Permission",
			PermKey:    "button:perm",
			PermType:   permission.PermTypeButton,
			Status:     permission.StatusEnabled,
			CreateTime: time.Now(),
			DelFlag:    permission.DelFlagNormal,
		}
		_ = repo.Create(perm)
	}

	menuPerms, err := repo.GetByType(permission.PermTypeMenu)
	if err != nil {
		t.Fatalf("GetByType failed: %v", err)
	}

	if len(menuPerms) != 2 {
		t.Errorf("Expected 2 menu permissions, got %d", len(menuPerms))
	}

	buttonPerms, err := repo.GetByType(permission.PermTypeButton)
	if err != nil {
		t.Fatalf("GetByType for button failed: %v", err)
	}

	if len(buttonPerms) != 3 {
		t.Errorf("Expected 3 button permissions, got %d", len(buttonPerms))
	}
}
