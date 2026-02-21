//go:build integration
// +build integration

package repository

import (
	"testing"
	"time"

	"dataease/backend/internal/domain/template"
)

func TestTemplateRepository_CreateAndGetByID(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewTemplateRepository(testDB)
	cleanupTables("core_visualization_template")

	tpl := &template.Template{
		Name:          "Test Template",
		Pid:           0,
		Level:         1,
		DvType:        "dashboard",
		NodeType:      "folder",
		CreateBy:      "admin",
		Snapshot:      "",
		TemplateType:  "system",
		TemplateStyle: "",
		TemplateData:  "{}",
		DynamicData:   "",
		AppData:       "",
		UseCount:      0,
		Version:       3,
	}

	err := repo.Create(tpl)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if tpl.ID == 0 {
		t.Error("Expected ID to be set after creation")
	}

	found, err := repo.GetByID(tpl.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if found.Name != "Test Template" {
		t.Errorf("Expected Name 'Test Template', got '%s'", found.Name)
	}
}

func TestTemplateRepository_List(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewTemplateRepository(testDB)
	cleanupTables("core_visualization_template")

	for i := 0; i < 3; i++ {
		tpl := &template.Template{
			Name:         "Dashboard Template",
			Pid:          1,
			Level:        2,
			DvType:       "dashboard",
			NodeType:     "template",
			CreateBy:     "admin",
			TemplateType: "system",
			Version:      3,
		}
		_ = repo.Create(tpl)
	}

	for i := 0; i < 2; i++ {
		tpl := &template.Template{
			Name:         "Screen Template",
			Pid:          1,
			Level:        2,
			DvType:       "screen",
			NodeType:     "template",
			CreateBy:     "admin",
			TemplateType: "system",
			Version:      3,
		}
		_ = repo.Create(tpl)
	}

	list, err := repo.List(1, "dashboard")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(list) != 3 {
		t.Errorf("Expected 3 dashboard templates, got %d", len(list))
	}

	list2, err := repo.List(1, "screen")
	if err != nil {
		t.Fatalf("List screen failed: %v", err)
	}

	if len(list2) != 2 {
		t.Errorf("Expected 2 screen templates, got %d", len(list2))
	}
}

func TestTemplateRepository_Update(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewTemplateRepository(testDB)
	cleanupTables("core_visualization_template")

	tpl := &template.Template{
		Name:         "Update Test",
		Pid:          0,
		Level:        1,
		DvType:       "dashboard",
		NodeType:     "template",
		CreateBy:     "admin",
		TemplateType: "system",
		Version:      3,
	}
	_ = repo.Create(tpl)

	tpl.Name = "Updated Name"
	tpl.TemplateData = `{"updated": true}`

	err := repo.Update(tpl)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	found, _ := repo.GetByID(tpl.ID)
	if found.Name != "Updated Name" {
		t.Errorf("Expected Name 'Updated Name', got '%s'", found.Name)
	}
}

func TestTemplateRepository_Delete(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewTemplateRepository(testDB)
	cleanupTables("core_visualization_template")

	tpl := &template.Template{
		Name:         "Delete Test",
		Pid:          0,
		Level:        1,
		DvType:       "dashboard",
		NodeType:     "template",
		CreateBy:     "admin",
		TemplateType: "system",
		Version:      3,
	}
	_ = repo.Create(tpl)

	err := repo.Delete(tpl.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = repo.GetByID(tpl.ID)
	if err == nil {
		t.Error("Expected error when getting deleted template")
	}
}

func TestTemplateRepository_Count(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewTemplateRepository(testDB)
	cleanupTables("core_visualization_template")

	for i := 0; i < 5; i++ {
		tpl := &template.Template{
			Name:         "Count Test",
			Pid:          10,
			Level:        2,
			DvType:       "dashboard",
			NodeType:     "template",
			CreateBy:     "admin",
			TemplateType: "system",
			Version:      3,
		}
		_ = repo.Create(tpl)
	}

	count, err := repo.Count(10, "dashboard")
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}

	if count != 5 {
		t.Errorf("Expected count 5, got %d", count)
	}
}

func TestTemplateRepository_IncrementUseCount(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewTemplateRepository(testDB)
	cleanupTables("core_visualization_template")

	tpl := &template.Template{
		Name:         "Use Count Test",
		Pid:          0,
		Level:        1,
		DvType:       "dashboard",
		NodeType:     "template",
		CreateBy:     "admin",
		TemplateType: "system",
		UseCount:     0,
		Version:      3,
	}
	_ = repo.Create(tpl)

	err := repo.IncrementUseCount(tpl.ID)
	if err != nil {
		t.Fatalf("IncrementUseCount failed: %v", err)
	}

	found, _ := repo.GetByID(tpl.ID)
	if found.UseCount != 1 {
		t.Errorf("Expected UseCount 1, got %d", found.UseCount)
	}

	time.Sleep(10 * time.Millisecond)
	_ = repo.IncrementUseCount(tpl.ID)
	found2, _ := repo.GetByID(tpl.ID)
	if found2.UseCount != 2 {
		t.Errorf("Expected UseCount 2, got %d", found2.UseCount)
	}
}
