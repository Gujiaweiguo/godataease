//go:build integration
// +build integration

package repository

import (
	"fmt"
	"testing"
	"time"

	"dataease/backend/internal/domain/role"
)

func TestRoleRepository_CreateAndGetByID(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewRoleRepository(testDB)
	cleanupTables("sys_role")

	r := &role.SysRole{
		RoleName:   "Test Role",
		RoleCode:   "ROLE_TEST",
		Status:     role.StatusEnabled,
		CreateTime: ptrTime(time.Now()),
	}

	err := repo.Create(r)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if r.RoleID == 0 {
		t.Error("Expected RoleID to be set after creation")
	}

	found, err := repo.GetByID(r.RoleID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if found.RoleName != "Test Role" {
		t.Errorf("Expected RoleName 'Test Role', got '%s'", found.RoleName)
	}
}

func TestRoleRepository_Update(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewRoleRepository(testDB)
	cleanupTables("sys_role")

	r := &role.SysRole{
		RoleName:   "Update Role",
		RoleCode:   "ROLE_UPDATE",
		Status:     role.StatusEnabled,
		CreateTime: ptrTime(time.Now()),
	}
	_ = repo.Create(r)

	r.RoleName = "Updated Role Name"
	err := repo.Update(r)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	found, _ := repo.GetByID(r.RoleID)
	if found.RoleName != "Updated Role Name" {
		t.Errorf("Expected RoleName 'Updated Role Name', got '%s'", found.RoleName)
	}
}

func TestRoleRepository_Delete(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewRoleRepository(testDB)
	cleanupTables("sys_role")

	r := &role.SysRole{
		RoleName:   "Delete Role",
		RoleCode:   "ROLE_DELETE",
		Status:     role.StatusEnabled,
		CreateTime: ptrTime(time.Now()),
	}
	_ = repo.Create(r)

	err := repo.Delete(r.RoleID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = repo.GetByID(r.RoleID)
	if err == nil {
		t.Error("Expected error when getting deleted role")
	}
}

func TestRoleRepository_Query(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewRoleRepository(testDB)
	cleanupTables("sys_role")

	for i := 1; i <= 3; i++ {
		r := &role.SysRole{
			RoleName:   fmt.Sprintf("Query Role %d", i),
			RoleCode:   fmt.Sprintf("ROLE_QUERY_%d", i),
			Status:     role.StatusEnabled,
			CreateTime: ptrTime(time.Now()),
		}
		_ = repo.Create(r)
	}

	roles, err := repo.Query("Query")
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if len(roles) != 3 {
		t.Errorf("Expected 3 roles, got %d", len(roles))
	}

	roles, _ = repo.Query("")
	if len(roles) < 3 {
		t.Errorf("Expected at least 3 roles, got %d", len(roles))
	}
}

func TestRoleRepository_CountByRoleCode(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewRoleRepository(testDB)
	cleanupTables("sys_role")

	r := &role.SysRole{
		RoleName:   "Count Role",
		RoleCode:   "ROLE_COUNT",
		Status:     role.StatusEnabled,
		CreateTime: ptrTime(time.Now()),
	}
	_ = repo.Create(r)

	count, err := repo.CountByRoleCode("ROLE_COUNT")
	if err != nil {
		t.Fatalf("CountByRoleCode failed: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected count 1, got %d", count)
	}

	count, _ = repo.CountByRoleCode("ROLE_NOT_EXIST")
	if count != 0 {
		t.Errorf("Expected count 0, got %d", count)
	}
}

func ptrTime(t time.Time) *time.Time {
	return &t
}
