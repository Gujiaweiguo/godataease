//go:build integration
// +build integration

package repository

import (
	"fmt"
	"testing"
	"time"

	"dataease/backend/internal/domain/datasource"
)

func TestDatasourceRepository_CreateAndGetByID(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewDatasourceRepository(testDB)
	cleanupTables("core_datasource")

	ds := &datasource.CoreDatasource{
		Name:       "Test MySQL",
		Type:       "mysql",
		Status:     strPtr("Success"),
		CreateTime: int64Ptr(time.Now().Unix()),
	}

	err := repo.Create(ds)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if ds.ID == 0 {
		t.Error("Expected ID to be set after creation")
	}

	found, err := repo.GetByID(ds.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if found.Name != "Test MySQL" {
		t.Errorf("Expected Name 'Test MySQL', got '%s'", found.Name)
	}
}

func TestDatasourceRepository_Query(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewDatasourceRepository(testDB)
	cleanupTables("core_datasource")

	for i := 1; i <= 3; i++ {
		ds := &datasource.CoreDatasource{
			Name:       fmt.Sprintf("Query DB %d", i),
			Type:       "mysql",
			Status:     strPtr("Success"),
			CreateTime: int64Ptr(time.Now().Unix()),
		}
		_ = repo.Create(ds)
	}

	keyword := "Query"
	req := &datasource.ListRequest{
		Keyword: &keyword,
		Current: 1,
		Size:    10,
	}

	list, total, err := repo.Query(req)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if total != 3 {
		t.Errorf("Expected total 3, got %d", total)
	}
	if len(list) != 3 {
		t.Errorf("Expected 3 items, got %d", len(list))
	}
}

func TestDatasourceRepository_Update(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewDatasourceRepository(testDB)
	cleanupTables("core_datasource")

	ds := &datasource.CoreDatasource{
		Name:       "Update DB",
		Type:       "mysql",
		Status:     strPtr("Success"),
		CreateTime: int64Ptr(time.Now().Unix()),
	}
	_ = repo.Create(ds)

	ds.Name = "Updated DB Name"
	err := repo.Update(ds)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	found, _ := repo.GetByID(ds.ID)
	if found.Name != "Updated DB Name" {
		t.Errorf("Expected Name 'Updated DB Name', got '%s'", found.Name)
	}
}

func TestDatasourceRepository_SoftDelete(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewDatasourceRepository(testDB)
	cleanupTables("core_datasource")

	ds := &datasource.CoreDatasource{
		Name:       "Delete DB",
		Type:       "mysql",
		Status:     strPtr("Success"),
		CreateTime: int64Ptr(time.Now().Unix()),
	}
	_ = repo.Create(ds)

	err := repo.SoftDelete(ds.ID)
	if err != nil {
		t.Fatalf("SoftDelete failed: %v", err)
	}

	_, err = repo.GetByID(ds.ID)
	if err == nil {
		t.Error("Expected error when getting deleted datasource")
	}
}

func TestDatasourceRepository_ListChildren(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewDatasourceRepository(testDB)
	cleanupTables("core_datasource")

	parent := &datasource.CoreDatasource{
		Name:       "Parent Folder",
		Type:       "folder",
		CreateTime: int64Ptr(time.Now().Unix()),
	}
	_ = repo.Create(parent)

	for i := 1; i <= 2; i++ {
		child := &datasource.CoreDatasource{
			Name:       fmt.Sprintf("Child %d", i),
			Type:       "mysql",
			PID:        int64Ptr(parent.ID),
			CreateTime: int64Ptr(time.Now().Unix()),
		}
		_ = repo.Create(child)
	}

	children, err := repo.ListChildren(parent.ID)
	if err != nil {
		t.Fatalf("ListChildren failed: %v", err)
	}

	if len(children) != 2 {
		t.Errorf("Expected 2 children, got %d", len(children))
	}
}

func TestDatasourceRepository_CountByNameAndPID(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewDatasourceRepository(testDB)
	cleanupTables("core_datasource")

	ds := &datasource.CoreDatasource{
		Name:       "Unique Name",
		Type:       "mysql",
		CreateTime: int64Ptr(time.Now().Unix()),
	}
	_ = repo.Create(ds)

	count, err := repo.CountByNameAndPID("Unique Name", 0, nil)
	if err != nil {
		t.Fatalf("CountByNameAndPID failed: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected count 1, got %d", count)
	}

	count, _ = repo.CountByNameAndPID("Not Exist", 0, nil)
	if count != 0 {
		t.Errorf("Expected count 0, got %d", count)
	}
}

func TestDatasourceRepository_ListAll(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewDatasourceRepository(testDB)
	cleanupTables("core_datasource")

	for i := 1; i <= 3; i++ {
		ds := &datasource.CoreDatasource{
			Name:       fmt.Sprintf("ListAll DB %d", i),
			Type:       "mysql",
			CreateTime: int64Ptr(time.Now().Unix()),
		}
		_ = repo.Create(ds)
	}

	list, err := repo.ListAll(nil)
	if err != nil {
		t.Fatalf("ListAll failed: %v", err)
	}

	if len(list) < 3 {
		t.Errorf("Expected at least 3 items, got %d", len(list))
	}
}

func TestDatasourceRepository_ListByType(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewDatasourceRepository(testDB)
	cleanupTables("core_datasource")

	for i := 1; i <= 2; i++ {
		ds := &datasource.CoreDatasource{
			Name:       fmt.Sprintf("MySQL DB %d", i),
			Type:       "mysql",
			CreateTime: int64Ptr(time.Now().Unix()),
		}
		_ = repo.Create(ds)
	}

	pg := &datasource.CoreDatasource{
		Name:       "PostgreSQL DB",
		Type:       "postgresql",
		CreateTime: int64Ptr(time.Now().Unix()),
	}
	_ = repo.Create(pg)

	list, err := repo.ListByType("mysql", nil)
	if err != nil {
		t.Fatalf("ListByType failed: %v", err)
	}

	if len(list) != 2 {
		t.Errorf("Expected 2 mysql datasources, got %d", len(list))
	}
}

func TestDatasourceRepository_ListSchemas(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewDatasourceRepository(testDB)

	schemas, err := repo.ListSchemas()
	if err != nil {
		t.Fatalf("ListSchemas failed: %v", err)
	}

	if len(schemas) == 0 {
		t.Error("Expected at least one schema")
	}
}

func int64Ptr(v int64) *int64 {
	return &v
}

func strPtr(v string) *string {
	return &v
}
