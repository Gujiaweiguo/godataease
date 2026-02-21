//go:build integration
// +build integration

package repository

import (
	"testing"
	"time"

	"dataease/backend/internal/domain/visualization"
)

func TestVisualizationRepository_CreateAndGetByID(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewVisualizationRepository(testDB)
	cleanupTables("data_visualization_info")

	now := time.Now().UnixMilli()
	creator := "testuser"
	dvType := "dashboard"
	v := &visualization.DataVisualizationInfo{
		Name:       "Test Dashboard",
		Type:       &dvType,
		CreateTime: &now,
		CreateBy:   &creator,
		UpdateTime: &now,
	}

	err := repo.Create(v)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if v.ID == 0 {
		t.Error("Expected ID to be set after creation")
	}

	found, err := repo.GetByID(v.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if found.Name != "Test Dashboard" {
		t.Errorf("Expected Name 'Test Dashboard', got '%s'", found.Name)
	}
}

func TestVisualizationRepository_Update(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewVisualizationRepository(testDB)
	cleanupTables("data_visualization_info")

	now := time.Now().UnixMilli()
	creator := "testuser"
	v := &visualization.DataVisualizationInfo{
		Name:       "Update Test",
		CreateTime: &now,
		CreateBy:   &creator,
		UpdateTime: &now,
	}
	_ = repo.Create(v)

	newName := "Updated Dashboard"
	v.Name = newName
	updatedTime := time.Now().UnixMilli()
	v.UpdateTime = &updatedTime

	err := repo.Update(v)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	found, _ := repo.GetByID(v.ID)
	if found.Name != newName {
		t.Errorf("Expected Name '%s', got '%s'", newName, found.Name)
	}
}

func TestVisualizationRepository_DeleteLogic(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewVisualizationRepository(testDB)
	cleanupTables("data_visualization_info")

	now := time.Now().UnixMilli()
	creator := "testuser"
	v := &visualization.DataVisualizationInfo{
		Name:       "Delete Test",
		CreateTime: &now,
		CreateBy:   &creator,
		UpdateTime: &now,
	}
	_ = repo.Create(v)

	err := repo.DeleteLogic(v.ID, "admin")
	if err != nil {
		t.Fatalf("DeleteLogic failed: %v", err)
	}

	_, err = repo.GetByID(v.ID)
	if err == nil {
		t.Error("Expected error when getting logically deleted visualization")
	}
}

func TestVisualizationRepository_Query(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewVisualizationRepository(testDB)
	cleanupTables("data_visualization_info")

	now := time.Now().UnixMilli()
	creator := "testuser"
	dashboardType := "dashboard"
	screenType := "screen"

	for i := 0; i < 3; i++ {
		v := &visualization.DataVisualizationInfo{
			Name:       "Dashboard Test",
			Type:       &dashboardType,
			CreateTime: &now,
			CreateBy:   &creator,
			UpdateTime: &now,
		}
		_ = repo.Create(v)
	}

	for i := 0; i < 2; i++ {
		v := &visualization.DataVisualizationInfo{
			Name:       "Screen Test",
			Type:       &screenType,
			CreateTime: &now,
			CreateBy:   &creator,
			UpdateTime: &now,
		}
		_ = repo.Create(v)
	}

	keyword := "Dashboard"
	req := &visualization.ListRequest{
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

	typeReq := &visualization.ListRequest{
		Type:    &screenType,
		Current: 1,
		Size:    10,
	}

	list2, total2, err := repo.Query(typeReq)
	if err != nil {
		t.Fatalf("Query by type failed: %v", err)
	}

	if total2 != 2 {
		t.Errorf("Expected total 2 for screen type, got %d", total2)
	}
}

func TestVisualizationRepository_QueryPagination(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewVisualizationRepository(testDB)
	cleanupTables("data_visualization_info")

	now := time.Now().UnixMilli()
	creator := "testuser"

	for i := 0; i < 15; i++ {
		v := &visualization.DataVisualizationInfo{
			Name:       "Pagination Test",
			CreateTime: &now,
			CreateBy:   &creator,
			UpdateTime: &now,
		}
		_ = repo.Create(v)
	}

	req := &visualization.ListRequest{
		Current: 1,
		Size:    5,
	}

	list, total, err := repo.Query(req)
	if err != nil {
		t.Fatalf("Query pagination failed: %v", err)
	}

	if total != 15 {
		t.Errorf("Expected total 15, got %d", total)
	}
	if len(list) != 5 {
		t.Errorf("Expected 5 items on page 1, got %d", len(list))
	}

	req.Current = 2
	list2, _, err := repo.Query(req)
	if err != nil {
		t.Fatalf("Query page 2 failed: %v", err)
	}
	if len(list2) != 5 {
		t.Errorf("Expected 5 items on page 2, got %d", len(list2))
	}
}
