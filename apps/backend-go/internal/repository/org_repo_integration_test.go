//go:build integration
// +build integration

package repository

import (
	"fmt"
	"testing"
	"time"

	"dataease/backend/internal/domain/org"
)

func TestOrgRepository_CreateAndGetByID(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewOrgRepository(testDB)
	cleanupTables("sys_org")

	o := &org.SysOrg{
		OrgName:    "Test Org",
		Level:      1,
		DelFlag:    org.DelFlagNormal,
		CreateTime: time.Now(),
	}

	err := repo.Create(o)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if o.OrgID == 0 {
		t.Error("Expected OrgID to be set after creation")
	}

	found, err := repo.GetByID(o.OrgID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if found.OrgName != "Test Org" {
		t.Errorf("Expected OrgName 'Test Org', got '%s'", found.OrgName)
	}
}

func TestOrgRepository_GetByName(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewOrgRepository(testDB)
	cleanupTables("sys_org")

	o := &org.SysOrg{
		OrgName:    "Unique Org",
		Level:      1,
		DelFlag:    org.DelFlagNormal,
		CreateTime: time.Now(),
	}
	_ = repo.Create(o)

	found, err := repo.GetByName("Unique Org")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}

	if found.OrgID != o.OrgID {
		t.Errorf("Expected OrgID %d, got %d", o.OrgID, found.OrgID)
	}
}

func TestOrgRepository_Update(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewOrgRepository(testDB)
	cleanupTables("sys_org")

	o := &org.SysOrg{
		OrgName:    "Update Org",
		Level:      1,
		DelFlag:    org.DelFlagNormal,
		CreateTime: time.Now(),
	}
	_ = repo.Create(o)

	o.OrgName = "Updated Org Name"
	err := repo.Update(o)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	found, _ := repo.GetByID(o.OrgID)
	if found.OrgName != "Updated Org Name" {
		t.Errorf("Expected OrgName 'Updated Org Name', got '%s'", found.OrgName)
	}
}

func TestOrgRepository_Delete(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewOrgRepository(testDB)
	cleanupTables("sys_org")

	o := &org.SysOrg{
		OrgName:    "Delete Org",
		Level:      1,
		DelFlag:    org.DelFlagNormal,
		CreateTime: time.Now(),
	}
	_ = repo.Create(o)

	err := repo.Delete(o.OrgID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = repo.GetByID(o.OrgID)
	if err == nil {
		t.Error("Expected error when getting deleted org")
	}
}

func TestOrgRepository_List(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewOrgRepository(testDB)
	cleanupTables("sys_org")

	for i := 1; i <= 3; i++ {
		o := &org.SysOrg{
			OrgName:    fmt.Sprintf("List Org %d", i),
			Level:      i,
			DelFlag:    org.DelFlagNormal,
			CreateTime: time.Now(),
		}
		_ = repo.Create(o)
	}

	orgs, err := repo.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(orgs) != 3 {
		t.Errorf("Expected 3 orgs, got %d", len(orgs))
	}
}

func TestOrgRepository_ListByParentID(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewOrgRepository(testDB)
	cleanupTables("sys_org")

	parent := &org.SysOrg{
		OrgName:    "Parent Org",
		Level:      1,
		DelFlag:    org.DelFlagNormal,
		CreateTime: time.Now(),
	}
	_ = repo.Create(parent)

	for i := 1; i <= 2; i++ {
		child := &org.SysOrg{
			OrgName:    fmt.Sprintf("Child Org %d", i),
			ParentID:   parent.OrgID,
			Level:      2,
			DelFlag:    org.DelFlagNormal,
			CreateTime: time.Now(),
		}
		_ = repo.Create(child)
	}

	children, err := repo.ListByParentID(parent.OrgID)
	if err != nil {
		t.Fatalf("ListByParentID failed: %v", err)
	}

	if len(children) != 2 {
		t.Errorf("Expected 2 children, got %d", len(children))
	}
}

func TestOrgRepository_CheckNameExists(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewOrgRepository(testDB)
	cleanupTables("sys_org")

	o := &org.SysOrg{
		OrgName:    "Existing Org",
		Level:      1,
		DelFlag:    org.DelFlagNormal,
		CreateTime: time.Now(),
	}
	_ = repo.Create(o)

	count, err := repo.CheckNameExists("Existing Org", 0)
	if err != nil {
		t.Fatalf("CheckNameExists failed: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected count 1, got %d", count)
	}

	count, _ = repo.CheckNameExists("Not Existing", 0)
	if count != 0 {
		t.Errorf("Expected count 0, got %d", count)
	}
}

func TestOrgRepository_CountChildren(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewOrgRepository(testDB)
	cleanupTables("sys_org")

	parent := &org.SysOrg{
		OrgName:    "Parent",
		Level:      1,
		DelFlag:    org.DelFlagNormal,
		CreateTime: time.Now(),
	}
	_ = repo.Create(parent)

	for i := 1; i <= 3; i++ {
		child := &org.SysOrg{
			OrgName:    fmt.Sprintf("Child %d", i),
			ParentID:   parent.OrgID,
			Level:      2,
			DelFlag:    org.DelFlagNormal,
			CreateTime: time.Now(),
		}
		_ = repo.Create(child)
	}

	count, err := repo.CountChildren(parent.OrgID)
	if err != nil {
		t.Fatalf("CountChildren failed: %v", err)
	}

	if count != 3 {
		t.Errorf("Expected 3 children, got %d", count)
	}
}

func TestOrgRepository_GetByIDs(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewOrgRepository(testDB)
	cleanupTables("sys_org")

	var ids []int64
	for i := 1; i <= 3; i++ {
		o := &org.SysOrg{
			OrgName:    fmt.Sprintf("GetByIDs Org %d", i),
			Level:      1,
			DelFlag:    org.DelFlagNormal,
			CreateTime: time.Now(),
		}
		_ = repo.Create(o)
		ids = append(ids, o.OrgID)
	}

	orgs, err := repo.GetByIDs(ids)
	if err != nil {
		t.Fatalf("GetByIDs failed: %v", err)
	}

	if len(orgs) != 3 {
		t.Errorf("Expected 3 orgs, got %d", len(orgs))
	}
}
