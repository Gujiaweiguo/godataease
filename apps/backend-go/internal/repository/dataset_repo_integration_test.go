//go:build integration
// +build integration

package repository

import (
	"fmt"
	"testing"

	"dataease/backend/internal/domain/dataset"
)

func TestDatasetRepository_CreateAndGetGroupByID(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewDatasetRepository(testDB)
	cleanupTables("core_dataset_group")

	group := &dataset.CoreDatasetGroup{
		Name:     "Test Dataset",
		NodeType: strPtr("dataset"),
	}

	err := repo.CreateGroup(group)
	if err != nil {
		t.Fatalf("CreateGroup failed: %v", err)
	}

	if group.ID == 0 {
		t.Error("Expected ID to be set after creation")
	}

	found, err := repo.GetGroupByID(group.ID)
	if err != nil {
		t.Fatalf("GetGroupByID failed: %v", err)
	}

	if found.Name != "Test Dataset" {
		t.Errorf("Expected Name 'Test Dataset', got '%s'", found.Name)
	}
}

func TestDatasetRepository_ListGroups(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewDatasetRepository(testDB)
	cleanupTables("core_dataset_group")

	for i := 1; i <= 3; i++ {
		group := &dataset.CoreDatasetGroup{
			Name:     fmt.Sprintf("List Dataset %d", i),
			NodeType: strPtr("dataset"),
		}
		_ = repo.CreateGroup(group)
	}

	keyword := "List"
	groups, err := repo.ListGroups(&keyword)
	if err != nil {
		t.Fatalf("ListGroups failed: %v", err)
	}

	if len(groups) != 3 {
		t.Errorf("Expected 3 groups, got %d", len(groups))
	}
}

func TestDatasetRepository_UpdateGroup(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewDatasetRepository(testDB)
	cleanupTables("core_dataset_group")

	group := &dataset.CoreDatasetGroup{
		Name:     "Update Dataset",
		NodeType: strPtr("dataset"),
	}
	_ = repo.CreateGroup(group)

	group.Name = "Updated Dataset Name"
	err := repo.UpdateGroup(group)
	if err != nil {
		t.Fatalf("UpdateGroup failed: %v", err)
	}

	found, _ := repo.GetGroupByID(group.ID)
	if found.Name != "Updated Dataset Name" {
		t.Errorf("Expected Name 'Updated Dataset Name', got '%s'", found.Name)
	}
}

func TestDatasetRepository_SoftDeleteGroup(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewDatasetRepository(testDB)
	cleanupTables("core_dataset_group")

	group := &dataset.CoreDatasetGroup{
		Name:     "Delete Dataset",
		NodeType: strPtr("dataset"),
	}
	_ = repo.CreateGroup(group)

	err := repo.SoftDeleteGroup(group.ID)
	if err != nil {
		t.Fatalf("SoftDeleteGroup failed: %v", err)
	}

	_, err = repo.GetGroupByID(group.ID)
	if err == nil {
		t.Error("Expected error when getting deleted group")
	}
}

func TestDatasetRepository_CountGroupByNameAndPID(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewDatasetRepository(testDB)
	cleanupTables("core_dataset_group")

	group := &dataset.CoreDatasetGroup{
		Name:     "Unique Dataset Name",
		NodeType: strPtr("dataset"),
	}
	_ = repo.CreateGroup(group)

	count, err := repo.CountGroupByNameAndPID("Unique Dataset Name", 0, nil)
	if err != nil {
		t.Fatalf("CountGroupByNameAndPID failed: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected count 1, got %d", count)
	}

	count, _ = repo.CountGroupByNameAndPID("Not Existing", 0, nil)
	if count != 0 {
		t.Errorf("Expected count 0, got %d", count)
	}
}

func TestDatasetRepository_ListGroupChildren(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewDatasetRepository(testDB)
	cleanupTables("core_dataset_group")

	parent := &dataset.CoreDatasetGroup{
		Name:     "Parent Folder",
		NodeType: strPtr("folder"),
	}
	_ = repo.CreateGroup(parent)

	for i := 1; i <= 2; i++ {
		child := &dataset.CoreDatasetGroup{
			Name:     fmt.Sprintf("Child %d", i),
			PID:      &parent.ID,
			NodeType: strPtr("dataset"),
		}
		_ = repo.CreateGroup(child)
	}

	children, err := repo.ListGroupChildren(parent.ID)
	if err != nil {
		t.Fatalf("ListGroupChildren failed: %v", err)
	}

	if len(children) != 2 {
		t.Errorf("Expected 2 children, got %d", len(children))
	}
}

func TestDatasetRepository_ListTablesByDatasetGroupID(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewDatasetRepository(testDB)
	cleanupTables("core_dataset_table")

	groupID := int64(100)
	for i := 1; i <= 3; i++ {
		table := &dataset.CoreDatasetTable{
			Name:           strPtr(fmt.Sprintf("Table %d", i)),
			DatasetGroupID: groupID,
		}
		if err := testDB.Create(table).Error; err != nil {
			t.Fatalf("Failed to create table: %v", err)
		}
	}

	tables, err := repo.ListTablesByDatasetGroupID(groupID)
	if err != nil {
		t.Fatalf("ListTablesByDatasetGroupID failed: %v", err)
	}

	if len(tables) != 3 {
		t.Errorf("Expected 3 tables, got %d", len(tables))
	}
}

func TestDatasetRepository_ListFields(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewDatasetRepository(testDB)
	cleanupTables("core_dataset_table_field")

	groupID := int64(200)
	for i := 1; i <= 3; i++ {
		field := &dataset.CoreDatasetTableField{
			DatasetGroupID: groupID,
			Name:           strPtr(fmt.Sprintf("Field %d", i)),
		}
		if err := testDB.Create(field).Error; err != nil {
			t.Fatalf("Failed to create field: %v", err)
		}
	}

	fields, err := repo.ListFields(groupID)
	if err != nil {
		t.Fatalf("ListFields failed: %v", err)
	}

	if len(fields) != 3 {
		t.Errorf("Expected 3 fields, got %d", len(fields))
	}
}

func TestDatasetRepository_GetFieldByID(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewDatasetRepository(testDB)
	cleanupTables("core_dataset_table_field")

	field := &dataset.CoreDatasetTableField{
		DatasetGroupID: 300,
		Name:           strPtr("Test Field"),
	}
	if err := testDB.Create(field).Error; err != nil {
		t.Fatalf("Failed to create field: %v", err)
	}

	found, err := repo.GetFieldByID(field.ID)
	if err != nil {
		t.Fatalf("GetFieldByID failed: %v", err)
	}

	if *found.Name != "Test Field" {
		t.Errorf("Expected Name 'Test Field', got '%s'", *found.Name)
	}
}

func TestDatasetRepository_GetTableByID(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewDatasetRepository(testDB)
	cleanupTables("core_dataset_table")

	table := &dataset.CoreDatasetTable{
		Name:           strPtr("Test Table"),
		DatasetGroupID: 400,
	}
	if err := testDB.Create(table).Error; err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	found, err := repo.GetTableByID(table.ID)
	if err != nil {
		t.Fatalf("GetTableByID failed: %v", err)
	}

	if *found.Name != "Test Table" {
		t.Errorf("Expected Name 'Test Table', got '%s'", *found.Name)
	}
}
